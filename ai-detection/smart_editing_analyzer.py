"""
AI智能剪辑 - 多模态内容分析引擎
实现语音活跃度检测、情绪识别、动作检测、人脸表情分析等功能
"""

import cv2
import numpy as np
import librosa
import tensorflow as tf
from transformers import pipeline, AutoTokenizer, AutoModelForSequenceClassification
import webrtcvad
import speech_recognition as sr
from sklearn.cluster import KMeans
from scipy import signal
import json
import logging
from typing import Dict, List, Tuple, Optional
from dataclasses import dataclass
from datetime import datetime
import torch
import torch.nn as nn
from facenet_pytorch import MTCNN, InceptionResnetV1
import mediapipe as mp

# 配置日志
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@dataclass
class AnalysisConfig:
    """分析配置"""
    video_sample_rate: float = 1.0  # 视频采样率（每秒分析帧数）
    audio_sample_rate: int = 16000  # 音频采样率
    emotion_threshold: float = 0.6  # 情绪检测阈值
    voice_activity_threshold: float = 0.5  # 语音活跃度阈值
    motion_threshold: float = 0.3  # 动作检测阈值
    face_confidence_threshold: float = 0.9  # 人脸检测置信度阈值

class VoiceActivityDetector:
    """语音活跃度检测器"""
    
    def __init__(self, sample_rate: int = 16000, frame_duration: int = 30):
        self.sample_rate = sample_rate
        self.frame_duration = frame_duration
        self.vad = webrtcvad.Vad(3)  # 最高敏感度
        self.frame_length = int(sample_rate * frame_duration / 1000)
        
    def detect_voice_activity(self, audio_data: np.ndarray) -> List[Dict]:
        """检测语音活跃度"""
        # 确保音频格式正确
        if audio_data.dtype != np.int16:
            audio_data = (audio_data * 32767).astype(np.int16)
            
        voice_segments = []
        frame_count = len(audio_data) // self.frame_length
        
        current_segment = None
        
        for i in range(frame_count):
            start_idx = i * self.frame_length
            end_idx = start_idx + self.frame_length
            frame = audio_data[start_idx:end_idx]
            
            # VAD检测
            is_speech = self.vad.is_speech(frame.tobytes(), self.sample_rate)
            timestamp = i * self.frame_duration / 1000.0
            
            if is_speech:
                if current_segment is None:
                    current_segment = {
                        'start_time': timestamp,
                        'end_time': timestamp,
                        'confidence': 1.0
                    }
                else:
                    current_segment['end_time'] = timestamp
            else:
                if current_segment is not None:
                    # 结束当前语音段
                    if current_segment['end_time'] - current_segment['start_time'] > 0.5:  # 最小0.5秒
                        voice_segments.append(current_segment)
                    current_segment = None
        
        # 处理最后一个段落
        if current_segment is not None:
            voice_segments.append(current_segment)
            
        return voice_segments

class EmotionAnalyzer:
    """情绪分析器"""
    
    def __init__(self):
        # 加载预训练的情绪识别模型
        self.text_emotion_pipeline = pipeline(
            "text-classification",
            model="j-hartmann/emotion-english-distilroberta-base",
            device=0 if torch.cuda.is_available() else -1
        )
        
        # 人脸情绪识别模型
        self.face_emotion_model = self._load_face_emotion_model()
        
        # MediaPipe人脸检测
        self.mp_face_detection = mp.solutions.face_detection
        self.mp_drawing = mp.solutions.drawing_utils
        self.face_detection = self.mp_face_detection.FaceDetection(
            model_selection=0, min_detection_confidence=0.5
        )
        
    def _load_face_emotion_model(self):
        """加载人脸情绪识别模型"""
        try:
            # 这里应该加载实际的人脸情绪识别模型
            # 示例使用简单的CNN模型
            model = tf.keras.Sequential([
                tf.keras.layers.Conv2D(32, (3, 3), activation='relu', input_shape=(48, 48, 1)),
                tf.keras.layers.MaxPooling2D(2, 2),
                tf.keras.layers.Conv2D(64, (3, 3), activation='relu'),
                tf.keras.layers.MaxPooling2D(2, 2),
                tf.keras.layers.Conv2D(128, (3, 3), activation='relu'),
                tf.keras.layers.MaxPooling2D(2, 2),
                tf.keras.layers.Flatten(),
                tf.keras.layers.Dense(512, activation='relu'),
                tf.keras.layers.Dropout(0.5),
                tf.keras.layers.Dense(7, activation='softmax')  # 7种情绪
            ])
            
            # 加载预训练权重（如果存在）
            try:
                model.load_weights('./models/face_emotion_weights.h5')
            except:
                logger.warning("未找到预训练的人脸情绪模型权重，使用随机初始化")
                
            return model
        except Exception as e:
            logger.error(f"加载人脸情绪模型失败: {e}")
            return None
    
    def analyze_text_emotion(self, text: str) -> Dict:
        """分析文本情绪"""
        try:
            results = self.text_emotion_pipeline(text)
            emotions = {}
            
            for result in results:
                emotions[result['label'].lower()] = result['score']
                
            return {
                'emotions': emotions,
                'dominant_emotion': max(emotions.items(), key=lambda x: x[1])[0],
                'confidence': max(emotions.values())
            }
        except Exception as e:
            logger.error(f"文本情绪分析失败: {e}")
            return {'emotions': {}, 'dominant_emotion': 'neutral', 'confidence': 0.0}
    
    def analyze_face_emotion(self, frame: np.ndarray) -> List[Dict]:
        """分析人脸情绪"""
        if self.face_emotion_model is None:
            return []
            
        try:
            # 转换为RGB
            rgb_frame = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
            
            # 检测人脸
            results = self.face_detection.process(rgb_frame)
            
            face_emotions = []
            
            if results.detections:
                for detection in results.detections:
                    # 获取人脸边界框
                    bbox = detection.location_data.relative_bounding_box
                    h, w, _ = frame.shape
                    
                    x = int(bbox.xmin * w)
                    y = int(bbox.ymin * h)
                    width = int(bbox.width * w)
                    height = int(bbox.height * h)
                    
                    # 提取人脸区域
                    face_roi = frame[y:y+height, x:x+width]
                    
                    if face_roi.size > 0:
                        # 预处理人脸图像
                        face_gray = cv2.cvtColor(face_roi, cv2.COLOR_BGR2GRAY)
                        face_resized = cv2.resize(face_gray, (48, 48))
                        face_normalized = face_resized.astype(np.float32) / 255.0
                        face_input = np.expand_dims(np.expand_dims(face_normalized, axis=0), axis=-1)
                        
                        # 情绪预测
                        emotion_probs = self.face_emotion_model.predict(face_input, verbose=0)[0]
                        
                        emotion_labels = ['angry', 'disgust', 'fear', 'happy', 'sad', 'surprise', 'neutral']
                        emotions = dict(zip(emotion_labels, emotion_probs))
                        
                        face_emotions.append({
                            'bounding_box': {'x': x, 'y': y, 'width': width, 'height': height},
                            'emotions': emotions,
                            'dominant_emotion': emotion_labels[np.argmax(emotion_probs)],
                            'confidence': float(np.max(emotion_probs))
                        })
            
            return face_emotions
            
        except Exception as e:
            logger.error(f"人脸情绪分析失败: {e}")
            return []

class MotionDetector:
    """动作检测器"""
    
    def __init__(self):
        self.prev_frame = None
        self.motion_history = []
        
    def detect_motion(self, frame: np.ndarray) -> Dict:
        """检测帧间动作"""
        gray = cv2.cvtColor(frame, cv2.COLOR_BGR2GRAY)
        
        if self.prev_frame is None:
            self.prev_frame = gray
            return {'motion_intensity': 0.0, 'motion_areas': []}
        
        # 计算帧差
        frame_diff = cv2.absdiff(self.prev_frame, gray)
        
        # 阈值处理
        _, thresh = cv2.threshold(frame_diff, 30, 255, cv2.THRESH_BINARY)
        
        # 形态学操作
        kernel = np.ones((5, 5), np.uint8)
        thresh = cv2.morphologyEx(thresh, cv2.MORPH_OPEN, kernel)
        thresh = cv2.morphologyEx(thresh, cv2.MORPH_CLOSE, kernel)
        
        # 查找轮廓
        contours, _ = cv2.findContours(thresh, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
        
        motion_areas = []
        total_motion = 0
        
        for contour in contours:
            area = cv2.contourArea(contour)
            if area > 500:  # 过滤小的噪声
                x, y, w, h = cv2.boundingRect(contour)
                motion_areas.append({
                    'x': int(x), 'y': int(y), 
                    'width': int(w), 'height': int(h),
                    'area': float(area)
                })
                total_motion += area
        
        # 计算运动强度
        frame_area = frame.shape[0] * frame.shape[1]
        motion_intensity = min(total_motion / frame_area, 1.0)
        
        self.prev_frame = gray
        self.motion_history.append(motion_intensity)
        
        # 保持历史记录长度
        if len(self.motion_history) > 30:  # 保持30帧历史
            self.motion_history.pop(0)
        
        return {
            'motion_intensity': motion_intensity,
            'motion_areas': motion_areas,
            'average_motion': np.mean(self.motion_history) if self.motion_history else 0.0
        }

class SpeechToTextProcessor:
    """语音转文字处理器"""
    
    def __init__(self):
        self.recognizer = sr.Recognizer()
        
    def transcribe_audio(self, audio_file: str) -> List[Dict]:
        """转录音频为文字"""
        try:
            # 加载音频
            audio_data, sr_rate = librosa.load(audio_file, sr=16000)
            
            # 分段处理（每30秒一段）
            segment_length = 30 * sr_rate
            segments = []
            
            for i in range(0, len(audio_data), segment_length):
                segment = audio_data[i:i + segment_length]
                
                if len(segment) < sr_rate:  # 跳过太短的段落
                    continue
                
                try:
                    # 转换为AudioData格式
                    audio_segment = sr.AudioData(
                        (segment * 32767).astype(np.int16).tobytes(),
                        sr_rate, 2
                    )
                    
                    # 语音识别
                    text = self.recognizer.recognize_google(audio_segment, language='zh-CN')
                    
                    segments.append({
                        'start_time': i / sr_rate,
                        'end_time': min((i + segment_length) / sr_rate, len(audio_data) / sr_rate),
                        'text': text,
                        'confidence': 0.8  # Google API不提供置信度，使用默认值
                    })
                    
                except sr.UnknownValueError:
                    # 无法识别的音频段落
                    continue
                except sr.RequestError as e:
                    logger.error(f"语音识别服务错误: {e}")
                    continue
            
            return segments
            
        except Exception as e:
            logger.error(f"语音转文字失败: {e}")
            return []

class ContentAnalyzer:
    """综合内容分析器"""
    
    def __init__(self, config: AnalysisConfig = None):
        self.config = config or AnalysisConfig()
        
        # 初始化各个分析器
        self.voice_detector = VoiceActivityDetector(self.config.audio_sample_rate)
        self.emotion_analyzer = EmotionAnalyzer()
        self.motion_detector = MotionDetector()
        self.speech_processor = SpeechToTextProcessor()
        
    def analyze_video(self, video_path: str) -> Dict:
        """分析视频内容"""
        logger.info(f"开始分析视频: {video_path}")
        
        # 打开视频
        cap = cv2.VideoCapture(video_path)
        if not cap.isOpened():
            raise ValueError(f"无法打开视频文件: {video_path}")
        
        # 获取视频信息
        fps = cap.get(cv2.CAP_PROP_FPS)
        total_frames = int(cap.get(cv2.CAP_PROP_FRAME_COUNT))
        duration = total_frames / fps
        
        logger.info(f"视频信息: {total_frames}帧, {fps}FPS, {duration:.2f}秒")
        
        # 分析结果存储
        analysis_results = {
            'duration': duration,
            'fps': fps,
            'total_frames': total_frames,
            'voice_activity': [],
            'emotions': [],
            'motion_data': [],
            'scene_changes': [],
            'highlights': []
        }
        
        frame_count = 0
        sample_interval = int(fps / self.config.video_sample_rate)
        
        while True:
            ret, frame = cap.read()
            if not ret:
                break
            
            # 按采样率处理帧
            if frame_count % sample_interval == 0:
                timestamp = frame_count / fps
                
                # 动作检测
                motion_data = self.motion_detector.detect_motion(frame)
                motion_data['timestamp'] = timestamp
                analysis_results['motion_data'].append(motion_data)
                
                # 人脸情绪分析
                face_emotions = self.emotion_analyzer.analyze_face_emotion(frame)
                if face_emotions:
                    analysis_results['emotions'].append({
                        'timestamp': timestamp,
                        'faces': face_emotions
                    })
                
                # 场景变化检测（简单的直方图比较）
                if frame_count > 0:
                    scene_change_score = self._detect_scene_change(frame)
                    if scene_change_score > 0.5:
                        analysis_results['scene_changes'].append(timestamp)
            
            frame_count += 1
            
            # 进度日志
            if frame_count % (total_frames // 10) == 0:
                progress = (frame_count / total_frames) * 100
                logger.info(f"视频分析进度: {progress:.1f}%")
        
        cap.release()
        
        # 音频分析
        logger.info("开始音频分析...")
        audio_analysis = self._analyze_audio(video_path)
        analysis_results.update(audio_analysis)
        
        # 生成综合分析报告
        analysis_results['summary'] = self._generate_summary(analysis_results)
        
        logger.info("视频分析完成")
        return analysis_results
    
    def _detect_scene_change(self, frame: np.ndarray) -> float:
        """检测场景变化"""
        # 简单的直方图比较方法
        if not hasattr(self, '_prev_hist'):
            self._prev_hist = cv2.calcHist([frame], [0, 1, 2], None, [50, 50, 50], [0, 256, 0, 256, 0, 256])
            return 0.0
        
        current_hist = cv2.calcHist([frame], [0, 1, 2], None, [50, 50, 50], [0, 256, 0, 256, 0, 256])
        correlation = cv2.compareHist(self._prev_hist, current_hist, cv2.HISTCMP_CORREL)
        
        self._prev_hist = current_hist
        return 1.0 - correlation  # 相关性越低，场景变化越大
    
    def _analyze_audio(self, video_path: str) -> Dict:
        """分析音频内容"""
        try:
            # 提取音频
            audio_data, sr = librosa.load(video_path, sr=self.config.audio_sample_rate)
            
            # 语音活跃度检测
            voice_segments = self.voice_detector.detect_voice_activity(
                (audio_data * 32767).astype(np.int16)
            )
            
            # 语音转文字
            transcript = self.speech_processor.transcribe_audio(video_path)
            
            # 文本情绪分析
            text_emotions = []
            for segment in transcript:
                emotion_result = self.emotion_analyzer.analyze_text_emotion(segment['text'])
                text_emotions.append({
                    'start_time': segment['start_time'],
                    'end_time': segment['end_time'],
                    'text': segment['text'],
                    'emotions': emotion_result['emotions'],
                    'dominant_emotion': emotion_result['dominant_emotion']
                })
            
            return {
                'voice_activity': voice_segments,
                'transcript': transcript,
                'text_emotions': text_emotions,
                'audio_duration': len(audio_data) / sr
            }
            
        except Exception as e:
            logger.error(f"音频分析失败: {e}")
            return {
                'voice_activity': [],
                'transcript': [],
                'text_emotions': [],
                'audio_duration': 0.0
            }
    
    def _generate_summary(self, analysis_results: Dict) -> Dict:
        """生成分析摘要"""
        summary = {
            'total_speaking_time': sum(
                segment['end_time'] - segment['start_time'] 
                for segment in analysis_results.get('voice_activity', [])
            ),
            'average_motion_intensity': np.mean([
                data['motion_intensity'] 
                for data in analysis_results.get('motion_data', [])
            ]) if analysis_results.get('motion_data') else 0.0,
            'scene_changes_count': len(analysis_results.get('scene_changes', [])),
            'dominant_emotions': self._get_dominant_emotions(analysis_results),
            'engagement_score': self._calculate_engagement_score(analysis_results)
        }
        
        return summary
    
    def _get_dominant_emotions(self, analysis_results: Dict) -> Dict:
        """获取主要情绪分布"""
        all_emotions = {}
        
        # 统计文本情绪
        for emotion_data in analysis_results.get('text_emotions', []):
            for emotion, score in emotion_data.get('emotions', {}).items():
                all_emotions[emotion] = all_emotions.get(emotion, 0) + score
        
        # 归一化
        if all_emotions:
            total = sum(all_emotions.values())
            all_emotions = {k: v/total for k, v in all_emotions.items()}
        
        return all_emotions
    
    def _calculate_engagement_score(self, analysis_results: Dict) -> float:
        """计算参与度评分"""
        # 基于多个因素计算参与度
        factors = []
        
        # 语音活跃度因子
        speaking_ratio = analysis_results['summary']['total_speaking_time'] / analysis_results['duration']
        factors.append(min(speaking_ratio * 2, 1.0))  # 说话时间占比
        
        # 动作活跃度因子
        motion_score = analysis_results['summary']['average_motion_intensity']
        factors.append(motion_score)
        
        # 场景变化因子
        scene_changes_per_minute = analysis_results['summary']['scene_changes_count'] / (analysis_results['duration'] / 60)
        factors.append(min(scene_changes_per_minute / 5, 1.0))  # 每分钟场景变化次数
        
        # 情绪多样性因子
        emotions = analysis_results['summary']['dominant_emotions']
        emotion_diversity = len([e for e in emotions.values() if e > 0.1])  # 显著情绪种类
        factors.append(min(emotion_diversity / 5, 1.0))
        
        return np.mean(factors) if factors else 0.0

# 使用示例
if __name__ == "__main__":
    config = AnalysisConfig()
    analyzer = ContentAnalyzer(config)
    
    # 分析视频
    video_path = "sample_meeting.mp4"
    results = analyzer.analyze_video(video_path)
    
    # 保存结果
    with open("analysis_results.json", "w", encoding="utf-8") as f:
        json.dump(results, f, ensure_ascii=False, indent=2)
    
    print("分析完成，结果已保存到 analysis_results.json")
