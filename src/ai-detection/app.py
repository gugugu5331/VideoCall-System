#!/usr/bin/env python3
"""
AI检测服务
使用深度学习模型检测伪造音视频
"""

import os
import sys
import logging
import asyncio
from datetime import datetime
from typing import Dict, List, Optional, Tuple
import json
import uuid

import numpy as np
import cv2
import librosa
import tensorflow as tf
from flask import Flask, request, jsonify
from werkzeug.utils import secure_filename
import redis
import pika
from concurrent.futures import ThreadPoolExecutor

# 配置日志
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class FaceSwapDetector:
    """人脸交换检测器"""
    
    def __init__(self, model_path: str):
        self.model_path = model_path
        self.model = None
        self.face_cascade = cv2.CascadeClassifier(
            cv2.data.haarcascades + 'haarcascade_frontalface_default.xml'
        )
        self.load_model()
    
    def load_model(self):
        """加载预训练模型"""
        try:
            if os.path.exists(self.model_path):
                self.model = tf.keras.models.load_model(self.model_path)
                logger.info(f"Loaded face swap detection model from {self.model_path}")
            else:
                # 创建一个简单的CNN模型作为示例
                self.model = self._create_dummy_model()
                logger.warning("Using dummy model for face swap detection")
        except Exception as e:
            logger.error(f"Failed to load face swap model: {e}")
            self.model = self._create_dummy_model()
    
    def _create_dummy_model(self):
        """创建示例模型"""
        model = tf.keras.Sequential([
            tf.keras.layers.Conv2D(32, (3, 3), activation='relu', input_shape=(224, 224, 3)),
            tf.keras.layers.MaxPooling2D(2, 2),
            tf.keras.layers.Conv2D(64, (3, 3), activation='relu'),
            tf.keras.layers.MaxPooling2D(2, 2),
            tf.keras.layers.Conv2D(128, (3, 3), activation='relu'),
            tf.keras.layers.MaxPooling2D(2, 2),
            tf.keras.layers.Flatten(),
            tf.keras.layers.Dense(512, activation='relu'),
            tf.keras.layers.Dropout(0.5),
            tf.keras.layers.Dense(1, activation='sigmoid')
        ])
        model.compile(optimizer='adam', loss='binary_crossentropy', metrics=['accuracy'])
        return model
    
    def detect_faces(self, image: np.ndarray) -> List[Tuple[int, int, int, int]]:
        """检测人脸"""
        gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        faces = self.face_cascade.detectMultiScale(gray, 1.1, 4)
        return faces.tolist()
    
    def preprocess_face(self, image: np.ndarray, face_coords: Tuple[int, int, int, int]) -> np.ndarray:
        """预处理人脸图像"""
        x, y, w, h = face_coords
        face = image[y:y+h, x:x+w]
        face = cv2.resize(face, (224, 224))
        face = face.astype(np.float32) / 255.0
        return np.expand_dims(face, axis=0)
    
    def predict(self, image: np.ndarray) -> Dict:
        """检测图像中的人脸交换"""
        try:
            faces = self.detect_faces(image)
            if not faces:
                return {
                    'is_fake': False,
                    'confidence': 0.0,
                    'faces': [],
                    'details': 'No faces detected'
                }
            
            results = []
            max_fake_confidence = 0.0
            
            for face_coords in faces:
                face_image = self.preprocess_face(image, face_coords)
                prediction = self.model.predict(face_image, verbose=0)[0][0]
                
                is_fake = prediction > 0.5
                confidence = float(prediction if is_fake else 1 - prediction)
                
                results.append({
                    'bbox': face_coords,
                    'is_fake': is_fake,
                    'confidence': confidence
                })
                
                if is_fake and confidence > max_fake_confidence:
                    max_fake_confidence = confidence
            
            overall_is_fake = max_fake_confidence > 0.5
            
            return {
                'is_fake': overall_is_fake,
                'confidence': max_fake_confidence,
                'faces': results,
                'details': f'Detected {len(faces)} faces'
            }
            
        except Exception as e:
            logger.error(f"Face swap detection error: {e}")
            return {
                'is_fake': False,
                'confidence': 0.0,
                'faces': [],
                'details': f'Error: {str(e)}'
            }

class VoiceSynthesisDetector:
    """语音合成检测器"""
    
    def __init__(self, model_path: str):
        self.model_path = model_path
        self.model = None
        self.load_model()
    
    def load_model(self):
        """加载预训练模型"""
        try:
            if os.path.exists(self.model_path):
                self.model = tf.keras.models.load_model(self.model_path)
                logger.info(f"Loaded voice synthesis detection model from {self.model_path}")
            else:
                # 创建一个简单的RNN模型作为示例
                self.model = self._create_dummy_model()
                logger.warning("Using dummy model for voice synthesis detection")
        except Exception as e:
            logger.error(f"Failed to load voice synthesis model: {e}")
            self.model = self._create_dummy_model()
    
    def _create_dummy_model(self):
        """创建示例模型"""
        model = tf.keras.Sequential([
            tf.keras.layers.LSTM(128, return_sequences=True, input_shape=(None, 13)),
            tf.keras.layers.LSTM(64),
            tf.keras.layers.Dense(32, activation='relu'),
            tf.keras.layers.Dropout(0.5),
            tf.keras.layers.Dense(1, activation='sigmoid')
        ])
        model.compile(optimizer='adam', loss='binary_crossentropy', metrics=['accuracy'])
        return model
    
    def extract_features(self, audio_path: str) -> np.ndarray:
        """提取音频特征"""
        try:
            # 加载音频文件
            y, sr = librosa.load(audio_path, sr=16000)
            
            # 提取MFCC特征
            mfcc = librosa.feature.mfcc(y=y, sr=sr, n_mfcc=13)
            
            # 转置并标准化
            mfcc = mfcc.T
            mfcc = (mfcc - np.mean(mfcc, axis=0)) / (np.std(mfcc, axis=0) + 1e-8)
            
            return np.expand_dims(mfcc, axis=0)
            
        except Exception as e:
            logger.error(f"Feature extraction error: {e}")
            return np.zeros((1, 100, 13))  # 返回零特征
    
    def predict(self, audio_path: str) -> Dict:
        """检测音频中的语音合成"""
        try:
            features = self.extract_features(audio_path)
            prediction = self.model.predict(features, verbose=0)[0][0]
            
            is_fake = prediction > 0.5
            confidence = float(prediction if is_fake else 1 - prediction)
            
            return {
                'is_fake': is_fake,
                'confidence': confidence,
                'details': f'Audio analysis completed'
            }
            
        except Exception as e:
            logger.error(f"Voice synthesis detection error: {e}")
            return {
                'is_fake': False,
                'confidence': 0.0,
                'details': f'Error: {str(e)}'
            }

class DetectionService:
    """检测服务主类"""
    
    def __init__(self):
        self.face_detector = FaceSwapDetector('./models/face_swap_detector.h5')
        self.voice_detector = VoiceSynthesisDetector('./models/voice_synthesis_detector.h5')
        self.redis_client = None
        self.rabbitmq_connection = None
        self.executor = ThreadPoolExecutor(max_workers=4)
        
        # 初始化Redis连接
        try:
            self.redis_client = redis.Redis(
                host=os.getenv('REDIS_HOST', 'localhost'),
                port=int(os.getenv('REDIS_PORT', 6379)),
                db=int(os.getenv('REDIS_DB', 0)),
                decode_responses=True
            )
            self.redis_client.ping()
            logger.info("Connected to Redis")
        except Exception as e:
            logger.error(f"Failed to connect to Redis: {e}")
        
        # 初始化RabbitMQ连接
        try:
            rabbitmq_url = os.getenv('RABBITMQ_URL', 'amqp://admin:password123@localhost:5672/')
            self.rabbitmq_connection = pika.BlockingConnection(pika.URLParameters(rabbitmq_url))
            logger.info("Connected to RabbitMQ")
        except Exception as e:
            logger.error(f"Failed to connect to RabbitMQ: {e}")
    
    def detect_image(self, image_path: str) -> Dict:
        """检测图像"""
        try:
            image = cv2.imread(image_path)
            if image is None:
                raise ValueError("Failed to load image")
            
            result = self.face_detector.predict(image)
            result['file_path'] = image_path
            result['detection_type'] = 'face_swap'
            result['timestamp'] = datetime.now().isoformat()
            
            return result
            
        except Exception as e:
            logger.error(f"Image detection error: {e}")
            return {
                'is_fake': False,
                'confidence': 0.0,
                'detection_type': 'face_swap',
                'error': str(e)
            }
    
    def detect_audio(self, audio_path: str) -> Dict:
        """检测音频"""
        try:
            result = self.voice_detector.predict(audio_path)
            result['file_path'] = audio_path
            result['detection_type'] = 'voice_synthesis'
            result['timestamp'] = datetime.now().isoformat()
            
            return result
            
        except Exception as e:
            logger.error(f"Audio detection error: {e}")
            return {
                'is_fake': False,
                'confidence': 0.0,
                'detection_type': 'voice_synthesis',
                'error': str(e)
            }
    
    def detect_video(self, video_path: str) -> Dict:
        """检测视频"""
        try:
            cap = cv2.VideoCapture(video_path)
            if not cap.isOpened():
                raise ValueError("Failed to open video")
            
            frame_results = []
            frame_count = 0
            fake_frames = 0
            total_confidence = 0.0
            
            # 每隔30帧检测一次
            while True:
                ret, frame = cap.read()
                if not ret:
                    break
                
                if frame_count % 30 == 0:
                    result = self.face_detector.predict(frame)
                    frame_results.append({
                        'frame': frame_count,
                        'is_fake': result['is_fake'],
                        'confidence': result['confidence']
                    })
                    
                    if result['is_fake']:
                        fake_frames += 1
                    total_confidence += result['confidence']
                
                frame_count += 1
            
            cap.release()
            
            # 计算整体结果
            if frame_results:
                avg_confidence = total_confidence / len(frame_results)
                fake_ratio = fake_frames / len(frame_results)
                is_fake = fake_ratio > 0.3  # 如果超过30%的帧被检测为伪造
            else:
                avg_confidence = 0.0
                fake_ratio = 0.0
                is_fake = False
            
            return {
                'is_fake': is_fake,
                'confidence': avg_confidence,
                'detection_type': 'face_swap',
                'file_path': video_path,
                'timestamp': datetime.now().isoformat(),
                'details': {
                    'total_frames': frame_count,
                    'analyzed_frames': len(frame_results),
                    'fake_frames': fake_frames,
                    'fake_ratio': fake_ratio,
                    'frame_results': frame_results[:10]  # 只返回前10个结果
                }
            }
            
        except Exception as e:
            logger.error(f"Video detection error: {e}")
            return {
                'is_fake': False,
                'confidence': 0.0,
                'detection_type': 'face_swap',
                'error': str(e)
            }
    
    def cache_result(self, task_id: str, result: Dict):
        """缓存检测结果"""
        if self.redis_client:
            try:
                self.redis_client.setex(
                    f"detection_result:{task_id}",
                    3600,  # 1小时过期
                    json.dumps(result)
                )
            except Exception as e:
                logger.error(f"Failed to cache result: {e}")
    
    def get_cached_result(self, task_id: str) -> Optional[Dict]:
        """获取缓存的检测结果"""
        if self.redis_client:
            try:
                cached = self.redis_client.get(f"detection_result:{task_id}")
                if cached:
                    return json.loads(cached)
            except Exception as e:
                logger.error(f"Failed to get cached result: {e}")
        return None

# 创建Flask应用
app = Flask(__name__)
app.config['MAX_CONTENT_LENGTH'] = 100 * 1024 * 1024  # 100MB

# 创建检测服务实例
detection_service = DetectionService()

# 允许的文件扩展名
ALLOWED_EXTENSIONS = {
    'image': {'png', 'jpg', 'jpeg', 'gif', 'bmp'},
    'audio': {'wav', 'mp3', 'flac', 'ogg', 'm4a'},
    'video': {'mp4', 'avi', 'mov', 'mkv', 'webm'}
}

def allowed_file(filename: str, file_type: str) -> bool:
    """检查文件扩展名是否允许"""
    return '.' in filename and \
           filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS.get(file_type, set())

@app.route('/health', methods=['GET'])
def health_check():
    """健康检查"""
    return jsonify({
        'status': 'ok',
        'service': 'ai-detection',
        'timestamp': datetime.now().isoformat()
    })

@app.route('/detect', methods=['POST'])
def detect():
    """检测接口"""
    try:
        # 检查文件
        if 'file' not in request.files:
            return jsonify({'error': 'No file provided'}), 400
        
        file = request.files['file']
        if file.filename == '':
            return jsonify({'error': 'No file selected'}), 400
        
        file_type = request.form.get('type', 'image')
        if file_type not in ['image', 'audio', 'video']:
            return jsonify({'error': 'Invalid file type'}), 400
        
        if not allowed_file(file.filename, file_type):
            return jsonify({'error': 'File type not allowed'}), 400
        
        # 生成任务ID
        task_id = str(uuid.uuid4())
        
        # 检查缓存
        cached_result = detection_service.get_cached_result(task_id)
        if cached_result:
            return jsonify({
                'task_id': task_id,
                'status': 'completed',
                'result': cached_result
            })
        
        # 保存文件
        filename = secure_filename(file.filename)
        upload_dir = f'./uploads/{file_type}'
        os.makedirs(upload_dir, exist_ok=True)
        file_path = os.path.join(upload_dir, f"{task_id}_{filename}")
        file.save(file_path)
        
        # 异步处理检测
        future = detection_service.executor.submit(
            process_detection, task_id, file_path, file_type
        )
        
        return jsonify({
            'task_id': task_id,
            'status': 'processing',
            'message': 'Detection started'
        }), 202
        
    except Exception as e:
        logger.error(f"Detection API error: {e}")
        return jsonify({'error': str(e)}), 500

@app.route('/result/<task_id>', methods=['GET'])
def get_result(task_id: str):
    """获取检测结果"""
    try:
        result = detection_service.get_cached_result(task_id)
        if result:
            return jsonify({
                'task_id': task_id,
                'status': 'completed',
                'result': result
            })
        else:
            return jsonify({
                'task_id': task_id,
                'status': 'processing',
                'message': 'Detection in progress'
            }), 202
            
    except Exception as e:
        logger.error(f"Get result API error: {e}")
        return jsonify({'error': str(e)}), 500

def process_detection(task_id: str, file_path: str, file_type: str):
    """处理检测任务"""
    try:
        if file_type == 'image':
            result = detection_service.detect_image(file_path)
        elif file_type == 'audio':
            result = detection_service.detect_audio(file_path)
        elif file_type == 'video':
            result = detection_service.detect_video(file_path)
        else:
            result = {'error': 'Unknown file type'}
        
        # 缓存结果
        detection_service.cache_result(task_id, result)
        
        # 清理临时文件
        try:
            os.remove(file_path)
        except:
            pass
        
        logger.info(f"Detection completed for task {task_id}")
        
    except Exception as e:
        logger.error(f"Detection processing error: {e}")
        error_result = {
            'is_fake': False,
            'confidence': 0.0,
            'error': str(e)
        }
        detection_service.cache_result(task_id, error_result)

if __name__ == '__main__':
    # 创建必要的目录
    os.makedirs('./models', exist_ok=True)
    os.makedirs('./uploads', exist_ok=True)
    
    # 启动Flask应用
    app.run(
        host='0.0.0.0',
        port=int(os.getenv('PORT', 8501)),
        debug=os.getenv('DEBUG', 'false').lower() == 'true'
    )
