"""
AI智能剪辑 - 智能评分与片段提取算法
自动识别会议中的重要时刻、精彩讨论、关键决策等片段
"""

import numpy as np
import pandas as pd
from typing import Dict, List, Tuple, Optional
from dataclasses import dataclass
import json
import logging
from sklearn.preprocessing import MinMaxScaler
from sklearn.cluster import DBSCAN
from scipy.signal import find_peaks, savgol_filter
from scipy.stats import zscore
import re
from collections import Counter
import nltk
from nltk.corpus import stopwords
from nltk.tokenize import word_tokenize
from textstat import flesch_reading_ease
import jieba
import jieba.analyse

# 确保NLTK数据下载
try:
    nltk.data.find('tokenizers/punkt')
    nltk.data.find('corpora/stopwords')
except LookupError:
    nltk.download('punkt')
    nltk.download('stopwords')

logger = logging.getLogger(__name__)

@dataclass
class HighlightConfig:
    """高光片段提取配置"""
    min_segment_duration: float = 5.0  # 最小片段时长（秒）
    max_segment_duration: float = 60.0  # 最大片段时长（秒）
    overlap_threshold: float = 0.3  # 重叠阈值
    score_threshold: float = 0.6  # 评分阈值
    max_highlights: int = 10  # 最大高光片段数量
    
    # 权重配置
    audio_weight: float = 0.3
    visual_weight: float = 0.25
    text_weight: float = 0.25
    interaction_weight: float = 0.2

class HighlightExtractor:
    """高光片段提取器"""
    
    def __init__(self, config: HighlightConfig = None):
        self.config = config or HighlightConfig()
        self.scaler = MinMaxScaler()
        
        # 中文停用词
        self.chinese_stopwords = set([
            '的', '了', '在', '是', '我', '有', '和', '就', '不', '人', '都', '一', '一个', '上', '也', '很', '到', '说', '要', '去', '你', '会', '着', '没有', '看', '好', '自己', '这'
        ])
        
        # 关键词权重
        self.keyword_weights = {
            # 决策相关
            '决定': 2.0, '决策': 2.0, '确定': 1.8, '同意': 1.5, '反对': 1.5,
            '投票': 2.0, '表决': 2.0, '通过': 1.8, '否决': 1.8,
            
            # 讨论相关
            '讨论': 1.5, '分析': 1.3, '建议': 1.3, '提议': 1.5, '方案': 1.5,
            '问题': 1.3, '解决': 1.5, '方法': 1.3,
            
            # 情绪相关
            '重要': 1.8, '关键': 1.8, '紧急': 2.0, '严重': 1.8,
            '成功': 1.5, '失败': 1.5, '困难': 1.3, '挑战': 1.3,
            
            # 时间相关
            '截止': 1.8, '期限': 1.8, '紧急': 2.0, '立即': 1.8,
            '马上': 1.5, '尽快': 1.5,
            
            # 否定词
            '不行': 1.5, '不可以': 1.5, '不同意': 1.8, '拒绝': 1.8
        }
    
    def extract_highlights(self, analysis_results: Dict) -> List[Dict]:
        """提取高光片段"""
        logger.info("开始提取高光片段...")
        
        # 1. 计算各维度评分
        audio_scores = self._calculate_audio_scores(analysis_results)
        visual_scores = self._calculate_visual_scores(analysis_results)
        text_scores = self._calculate_text_scores(analysis_results)
        interaction_scores = self._calculate_interaction_scores(analysis_results)
        
        # 2. 综合评分
        combined_scores = self._combine_scores(
            audio_scores, visual_scores, text_scores, interaction_scores
        )
        
        # 3. 识别峰值片段
        peak_segments = self._identify_peak_segments(combined_scores, analysis_results['duration'])
        
        # 4. 片段分类和优化
        classified_segments = self._classify_segments(peak_segments, analysis_results)
        
        # 5. 去重和排序
        final_highlights = self._deduplicate_and_rank(classified_segments)
        
        logger.info(f"提取到 {len(final_highlights)} 个高光片段")
        return final_highlights
    
    def _calculate_audio_scores(self, analysis_results: Dict) -> np.ndarray:
        """计算音频维度评分"""
        duration = analysis_results['duration']
        time_points = np.arange(0, duration, 1.0)  # 每秒一个评分点
        scores = np.zeros(len(time_points))
        
        # 语音活跃度评分
        voice_segments = analysis_results.get('voice_activity', [])
        for segment in voice_segments:
            start_idx = int(segment['start_time'])
            end_idx = min(int(segment['end_time']), len(scores) - 1)
            if start_idx < len(scores):
                scores[start_idx:end_idx + 1] += segment.get('confidence', 1.0)
        
        # 音量变化评分（基于语音段的密度）
        for i, point in enumerate(time_points):
            # 计算该时间点周围的语音密度
            window_start = max(0, point - 5)  # 前后5秒窗口
            window_end = min(duration, point + 5)
            
            voice_time_in_window = sum(
                min(segment['end_time'], window_end) - max(segment['start_time'], window_start)
                for segment in voice_segments
                if segment['start_time'] < window_end and segment['end_time'] > window_start
            )
            
            voice_density = voice_time_in_window / (window_end - window_start)
            scores[i] += voice_density * 0.5
        
        return self._smooth_scores(scores)
    
    def _calculate_visual_scores(self, analysis_results: Dict) -> np.ndarray:
        """计算视觉维度评分"""
        duration = analysis_results['duration']
        time_points = np.arange(0, duration, 1.0)
        scores = np.zeros(len(time_points))
        
        # 动作强度评分
        motion_data = analysis_results.get('motion_data', [])
        for data in motion_data:
            timestamp = data['timestamp']
            idx = int(timestamp)
            if idx < len(scores):
                scores[idx] += data['motion_intensity'] * 2.0
        
        # 情绪变化评分
        emotions_data = analysis_results.get('emotions', [])
        for emotion_frame in emotions_data:
            timestamp = emotion_frame['timestamp']
            idx = int(timestamp)
            if idx < len(scores):
                # 计算情绪强度（非中性情绪的总和）
                emotion_intensity = 0
                for face in emotion_frame.get('faces', []):
                    emotions = face.get('emotions', {})
                    # 排除中性情绪，计算其他情绪强度
                    non_neutral = sum(v for k, v in emotions.items() if k != 'neutral')
                    emotion_intensity += non_neutral
                
                scores[idx] += min(emotion_intensity, 2.0)  # 限制最大值
        
        # 场景变化评分
        scene_changes = analysis_results.get('scene_changes', [])
        for change_time in scene_changes:
            idx = int(change_time)
            if idx < len(scores):
                scores[idx] += 1.5
        
        return self._smooth_scores(scores)
    
    def _calculate_text_scores(self, analysis_results: Dict) -> np.ndarray:
        """计算文本维度评分"""
        duration = analysis_results['duration']
        time_points = np.arange(0, duration, 1.0)
        scores = np.zeros(len(time_points))
        
        # 文本情绪评分
        text_emotions = analysis_results.get('text_emotions', [])
        for emotion_segment in text_emotions:
            start_time = emotion_segment['start_time']
            end_time = emotion_segment['end_time']
            text = emotion_segment['text']
            emotions = emotion_segment.get('emotions', {})
            
            # 计算时间范围内的索引
            start_idx = int(start_time)
            end_idx = min(int(end_time), len(scores) - 1)
            
            if start_idx < len(scores):
                # 情绪强度评分
                emotion_score = sum(v for k, v in emotions.items() if k in ['anger', 'fear', 'joy', 'surprise'])
                
                # 关键词评分
                keyword_score = self._calculate_keyword_score(text)
                
                # 文本复杂度评分
                complexity_score = self._calculate_text_complexity(text)
                
                total_score = emotion_score + keyword_score + complexity_score
                scores[start_idx:end_idx + 1] += total_score
        
        return self._smooth_scores(scores)
    
    def _calculate_interaction_scores(self, analysis_results: Dict) -> np.ndarray:
        """计算交互维度评分"""
        duration = analysis_results['duration']
        time_points = np.arange(0, duration, 1.0)
        scores = np.zeros(len(time_points))
        
        # 说话人切换评分
        transcript = analysis_results.get('transcript', [])
        prev_speaker = None
        
        for segment in transcript:
            current_speaker = segment.get('speaker_id', 'unknown')
            start_time = segment['start_time']
            idx = int(start_time)
            
            if idx < len(scores):
                # 说话人切换加分
                if prev_speaker and prev_speaker != current_speaker:
                    scores[idx] += 1.0
                
                # 文本长度加分（表示详细讨论）
                text_length = len(segment.get('text', ''))
                length_score = min(text_length / 100, 1.0)  # 标准化到0-1
                scores[idx] += length_score * 0.5
            
            prev_speaker = current_speaker
        
        # 多人同时说话检测（基于语音活跃度重叠）
        voice_segments = analysis_results.get('voice_activity', [])
        for i, segment1 in enumerate(voice_segments):
            for segment2 in voice_segments[i+1:]:
                # 检测重叠
                overlap_start = max(segment1['start_time'], segment2['start_time'])
                overlap_end = min(segment1['end_time'], segment2['end_time'])
                
                if overlap_start < overlap_end:  # 有重叠
                    start_idx = int(overlap_start)
                    end_idx = min(int(overlap_end), len(scores) - 1)
                    if start_idx < len(scores):
                        scores[start_idx:end_idx + 1] += 0.8  # 多人讨论加分
        
        return self._smooth_scores(scores)
    
    def _calculate_keyword_score(self, text: str) -> float:
        """计算关键词评分"""
        if not text:
            return 0.0
        
        # 中文分词
        words = jieba.lcut(text.lower())
        
        score = 0.0
        for word in words:
            if word in self.keyword_weights:
                score += self.keyword_weights[word]
        
        # 标准化
        return min(score / len(words), 2.0) if words else 0.0
    
    def _calculate_text_complexity(self, text: str) -> float:
        """计算文本复杂度评分"""
        if not text or len(text) < 10:
            return 0.0
        
        # 句子长度
        sentences = re.split(r'[。！？.!?]', text)
        avg_sentence_length = np.mean([len(s) for s in sentences if s.strip()])
        
        # 词汇多样性
        words = jieba.lcut(text)
        unique_words = set(words) - self.chinese_stopwords
        vocabulary_diversity = len(unique_words) / len(words) if words else 0
        
        # 综合复杂度评分
        complexity = (avg_sentence_length / 20 + vocabulary_diversity) / 2
        return min(complexity, 1.0)
    
    def _combine_scores(self, audio_scores: np.ndarray, visual_scores: np.ndarray, 
                       text_scores: np.ndarray, interaction_scores: np.ndarray) -> np.ndarray:
        """综合各维度评分"""
        # 确保所有评分数组长度一致
        min_length = min(len(audio_scores), len(visual_scores), len(text_scores), len(interaction_scores))
        
        audio_scores = audio_scores[:min_length]
        visual_scores = visual_scores[:min_length]
        text_scores = text_scores[:min_length]
        interaction_scores = interaction_scores[:min_length]
        
        # 标准化各维度评分
        audio_norm = self._normalize_scores(audio_scores)
        visual_norm = self._normalize_scores(visual_scores)
        text_norm = self._normalize_scores(text_scores)
        interaction_norm = self._normalize_scores(interaction_scores)
        
        # 加权组合
        combined = (
            audio_norm * self.config.audio_weight +
            visual_norm * self.config.visual_weight +
            text_norm * self.config.text_weight +
            interaction_norm * self.config.interaction_weight
        )
        
        return combined
    
    def _normalize_scores(self, scores: np.ndarray) -> np.ndarray:
        """标准化评分"""
        if len(scores) == 0 or np.max(scores) == 0:
            return scores
        
        # 使用Z-score标准化，然后映射到0-1范围
        z_scores = zscore(scores)
        # 将Z-score映射到0-1范围（3个标准差内）
        normalized = np.clip((z_scores + 3) / 6, 0, 1)
        
        return normalized
    
    def _smooth_scores(self, scores: np.ndarray, window_length: int = 5) -> np.ndarray:
        """平滑评分曲线"""
        if len(scores) < window_length:
            return scores
        
        # 使用Savitzky-Golay滤波器平滑
        try:
            smoothed = savgol_filter(scores, window_length, 2)
            return np.maximum(smoothed, 0)  # 确保非负
        except:
            return scores
    
    def _identify_peak_segments(self, scores: np.ndarray, duration: float) -> List[Dict]:
        """识别峰值片段"""
        # 找到峰值点
        peaks, properties = find_peaks(
            scores, 
            height=self.config.score_threshold,
            distance=int(self.config.min_segment_duration),  # 最小间隔
            prominence=0.1
        )
        
        segments = []
        
        for peak_idx in peaks:
            peak_time = float(peak_idx)
            peak_score = scores[peak_idx]
            
            # 确定片段边界
            start_time = max(0, peak_time - self.config.max_segment_duration / 2)
            end_time = min(duration, peak_time + self.config.max_segment_duration / 2)
            
            # 优化边界（寻找局部最小值）
            start_time, end_time = self._optimize_segment_boundaries(
                scores, int(start_time), int(end_time), peak_idx
            )
            
            # 确保最小时长
            if end_time - start_time >= self.config.min_segment_duration:
                segments.append({
                    'start_time': start_time,
                    'end_time': end_time,
                    'peak_time': peak_time,
                    'score': float(peak_score),
                    'duration': end_time - start_time
                })
        
        return segments
    
    def _optimize_segment_boundaries(self, scores: np.ndarray, start_idx: int, 
                                   end_idx: int, peak_idx: int) -> Tuple[float, float]:
        """优化片段边界"""
        # 向前寻找局部最小值作为开始点
        search_start = max(0, peak_idx - int(self.config.max_segment_duration / 2))
        for i in range(peak_idx, search_start, -1):
            if i > 0 and scores[i] < scores[i-1] and scores[i] < scores[i+1]:
                start_idx = i
                break
        
        # 向后寻找局部最小值作为结束点
        search_end = min(len(scores) - 1, peak_idx + int(self.config.max_segment_duration / 2))
        for i in range(peak_idx, search_end):
            if i < len(scores) - 1 and scores[i] < scores[i-1] and scores[i] < scores[i+1]:
                end_idx = i
                break
        
        return float(start_idx), float(end_idx)
    
    def _classify_segments(self, segments: List[Dict], analysis_results: Dict) -> List[Dict]:
        """对片段进行分类"""
        classified_segments = []
        
        for segment in segments:
            start_time = segment['start_time']
            end_time = segment['end_time']
            
            # 分析片段内容特征
            segment_features = self._analyze_segment_features(
                start_time, end_time, analysis_results
            )
            
            # 确定片段类型
            segment_type = self._determine_segment_type(segment_features)
            
            # 生成片段摘要
            summary = self._generate_segment_summary(
                start_time, end_time, analysis_results, segment_features
            )
            
            classified_segment = {
                **segment,
                'type': segment_type,
                'features': segment_features,
                'summary': summary,
                'participants': segment_features.get('participants', []),
                'keywords': segment_features.get('keywords', [])
            }
            
            classified_segments.append(classified_segment)
        
        return classified_segments
    
    def _analyze_segment_features(self, start_time: float, end_time: float, 
                                 analysis_results: Dict) -> Dict:
        """分析片段特征"""
        features = {
            'participants': set(),
            'keywords': [],
            'emotions': {},
            'motion_intensity': 0.0,
            'voice_activity': 0.0,
            'text_content': []
        }
        
        # 分析转录文本
        transcript = analysis_results.get('transcript', [])
        for segment in transcript:
            if (segment['start_time'] >= start_time and segment['end_time'] <= end_time):
                features['participants'].add(segment.get('speaker_id', 'unknown'))
                features['text_content'].append(segment['text'])
                
                # 提取关键词
                words = jieba.analyse.extract_tags(segment['text'], topK=5)
                features['keywords'].extend(words)
        
        # 分析情绪
        text_emotions = analysis_results.get('text_emotions', [])
        emotion_counts = {}
        for emotion_segment in text_emotions:
            if (emotion_segment['start_time'] >= start_time and 
                emotion_segment['end_time'] <= end_time):
                dominant_emotion = emotion_segment['dominant_emotion']
                emotion_counts[dominant_emotion] = emotion_counts.get(dominant_emotion, 0) + 1
        
        features['emotions'] = emotion_counts
        features['participants'] = list(features['participants'])
        features['keywords'] = list(set(features['keywords']))  # 去重
        
        return features
    
    def _determine_segment_type(self, features: Dict) -> str:
        """确定片段类型"""
        keywords = features.get('keywords', [])
        emotions = features.get('emotions', {})
        participants = features.get('participants', [])
        
        # 决策类型
        decision_keywords = ['决定', '决策', '确定', '同意', '反对', '投票', '表决']
        if any(keyword in keywords for keyword in decision_keywords):
            return 'decision'
        
        # 讨论类型
        discussion_keywords = ['讨论', '分析', '建议', '提议', '方案']
        if any(keyword in keywords for keyword in discussion_keywords):
            return 'discussion'
        
        # 演示类型
        presentation_keywords = ['展示', '演示', '介绍', '说明', '报告']
        if any(keyword in keywords for keyword in presentation_keywords):
            return 'presentation'
        
        # 反应类型（基于情绪）
        if emotions and max(emotions.values()) > 2:  # 强烈情绪反应
            return 'reaction'
        
        # 多人互动
        if len(participants) > 2:
            return 'interaction'
        
        return 'general'
    
    def _generate_segment_summary(self, start_time: float, end_time: float, 
                                 analysis_results: Dict, features: Dict) -> str:
        """生成片段摘要"""
        text_content = features.get('text_content', [])
        participants = features.get('participants', [])
        keywords = features.get('keywords', [])
        
        if not text_content:
            return f"时间段 {start_time:.1f}s-{end_time:.1f}s 的重要片段"
        
        # 合并文本内容
        full_text = ' '.join(text_content)
        
        # 生成简单摘要（取前50个字符）
        summary = full_text[:50] + "..." if len(full_text) > 50 else full_text
        
        # 添加参与者信息
        if participants:
            participant_info = f"参与者: {', '.join(participants[:3])}"
            if len(participants) > 3:
                participant_info += f" 等{len(participants)}人"
        else:
            participant_info = "参与者: 未知"
        
        # 添加关键词
        keyword_info = f"关键词: {', '.join(keywords[:5])}" if keywords else ""
        
        return f"{summary} | {participant_info} | {keyword_info}"
    
    def _deduplicate_and_rank(self, segments: List[Dict]) -> List[Dict]:
        """去重和排序"""
        # 按评分排序
        segments.sort(key=lambda x: x['score'], reverse=True)
        
        # 去除重叠片段
        final_segments = []
        
        for segment in segments:
            is_overlapping = False
            
            for existing in final_segments:
                overlap = self._calculate_overlap(segment, existing)
                if overlap > self.config.overlap_threshold:
                    is_overlapping = True
                    break
            
            if not is_overlapping:
                final_segments.append(segment)
                
                # 限制最大数量
                if len(final_segments) >= self.config.max_highlights:
                    break
        
        # 按时间排序
        final_segments.sort(key=lambda x: x['start_time'])
        
        return final_segments
    
    def _calculate_overlap(self, segment1: Dict, segment2: Dict) -> float:
        """计算两个片段的重叠度"""
        start1, end1 = segment1['start_time'], segment1['end_time']
        start2, end2 = segment2['start_time'], segment2['end_time']
        
        overlap_start = max(start1, start2)
        overlap_end = min(end1, end2)
        
        if overlap_start >= overlap_end:
            return 0.0
        
        overlap_duration = overlap_end - overlap_start
        total_duration = max(end1, end2) - min(start1, start2)
        
        return overlap_duration / total_duration if total_duration > 0 else 0.0

# 使用示例
if __name__ == "__main__":
    # 示例分析结果
    sample_analysis = {
        'duration': 1800.0,  # 30分钟
        'voice_activity': [
            {'start_time': 10.0, 'end_time': 25.0, 'confidence': 0.9},
            {'start_time': 30.0, 'end_time': 45.0, 'confidence': 0.8},
        ],
        'motion_data': [
            {'timestamp': 15.0, 'motion_intensity': 0.7},
            {'timestamp': 35.0, 'motion_intensity': 0.9},
        ],
        'emotions': [
            {'timestamp': 20.0, 'faces': [{'emotions': {'happy': 0.8, 'neutral': 0.2}}]},
        ],
        'text_emotions': [
            {
                'start_time': 10.0, 'end_time': 25.0, 'text': '我们需要做出重要决定',
                'emotions': {'joy': 0.3, 'anger': 0.1}, 'dominant_emotion': 'joy'
            }
        ],
        'transcript': [
            {'start_time': 10.0, 'end_time': 25.0, 'speaker_id': 'speaker1', 'text': '我们需要做出重要决定'},
        ],
        'scene_changes': [18.0, 40.0]
    }
    
    config = HighlightConfig()
    extractor = HighlightExtractor(config)
    
    highlights = extractor.extract_highlights(sample_analysis)
    
    print(f"提取到 {len(highlights)} 个高光片段:")
    for i, highlight in enumerate(highlights, 1):
        print(f"{i}. {highlight['start_time']:.1f}s-{highlight['end_time']:.1f}s "
              f"(评分: {highlight['score']:.2f}, 类型: {highlight['type']})")
        print(f"   摘要: {highlight['summary']}")
        print()
