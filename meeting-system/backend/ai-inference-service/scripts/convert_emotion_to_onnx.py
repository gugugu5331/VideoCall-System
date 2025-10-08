#!/usr/bin/env python3
"""
Emotion Detection 模型转换为 ONNX 格式

使用预训练的音频情感分类模型
输出 7 种情感：neutral, happy, sad, angry, fearful, disgusted, surprised
"""

import torch
import torch.nn as nn
import onnx
import numpy as np
from pathlib import Path
import sys


class SimpleEmotionModel(nn.Module):
    """
    简化的情感检测模型
    
    架构:
    - 输入: 音频特征 (mel-spectrogram)
    - LSTM 层提取时序特征
    - 全连接层分类
    - 输出: 7 种情感的概率
    """
    
    def __init__(self, n_mels=80, hidden_size=256, n_emotions=7):
        super().__init__()
        
        self.lstm = nn.LSTM(
            input_size=n_mels,
            hidden_size=hidden_size,
            num_layers=2,
            batch_first=True,
            bidirectional=True,
            dropout=0.3
        )
        
        self.attention = nn.Linear(hidden_size * 2, 1)
        self.fc = nn.Linear(hidden_size * 2, n_emotions)
    
    def forward(self, x):
        # x: (batch, n_mels, n_frames)
        x = x.transpose(1, 2)  # (batch, n_frames, n_mels)
        
        # LSTM
        lstm_out, _ = self.lstm(x)  # (batch, n_frames, hidden_size*2)
        
        # Attention pooling
        attention_weights = torch.softmax(self.attention(lstm_out), dim=1)  # (batch, n_frames, 1)
        attended = torch.sum(lstm_out * attention_weights, dim=1)  # (batch, hidden_size*2)
        
        # Classification
        output = self.fc(attended)  # (batch, n_emotions)
        
        return output


def create_emotion_model(output_dir="/work/models"):
    """
    创建情感检测模型并转换为 ONNX 格式
    
    Args:
        output_dir: 输出目录
    """
    print(f"=" * 80)
    print(f"🎯 创建 Emotion Detection 模型")
    print(f"=" * 80)
    print()
    
    # 情感标签
    emotion_labels = ["neutral", "happy", "sad", "angry", "fearful", "disgusted", "surprised"]
    
    print(f"📊 模型配置:")
    print(f"   输入: mel-spectrogram (80 x n_frames)")
    print(f"   输出: 7 种情感概率")
    print(f"   情感: {', '.join(emotion_labels)}")
    print()
    
    # 1. 创建模型
    print(f"🔧 创建模型...")
    model = SimpleEmotionModel(n_mels=80, hidden_size=256, n_emotions=7)
    model.eval()
    
    # 初始化权重（使用预训练权重会更好，但这里用随机初始化演示）
    print(f"⚙️ 初始化模型权重...")
    for name, param in model.named_parameters():
        if 'weight' in name:
            nn.init.xavier_uniform_(param)
        elif 'bias' in name:
            nn.init.zeros_(param)
    
    print(f"✅ 模型创建成功")
    print()
    
    # 2. 创建示例输入
    batch_size = 1
    n_mels = 80
    n_frames = 100
    
    print(f"📊 创建示例输入: shape=({batch_size}, {n_mels}, {n_frames})")
    dummy_input = torch.randn(batch_size, n_mels, n_frames)
    
    # 3. 导出为 ONNX
    output_path = Path(output_dir) / "emotion-model.onnx"
    output_path.parent.mkdir(parents=True, exist_ok=True)
    
    print(f"💾 导出 ONNX 模型到: {output_path}")
    
    try:
        torch.onnx.export(
            model,
            dummy_input,
            str(output_path),
            export_params=True,
            opset_version=14,
            do_constant_folding=True,
            input_names=['audio_input'],
            output_names=['emotion_output'],
            dynamic_axes={
                'audio_input': {0: 'batch_size', 2: 'n_frames'},
                'emotion_output': {0: 'batch_size'}
            }
        )
        print(f"✅ ONNX 模型导出成功")
    except Exception as e:
        print(f"❌ ONNX 导出失败: {e}")
        return False
    
    print()
    
    # 4. 验证 ONNX 模型
    print(f"🔍 验证 ONNX 模型...")
    try:
        onnx_model = onnx.load(str(output_path))
        onnx.checker.check_model(onnx_model)
        print(f"✅ ONNX 模型验证通过")
    except Exception as e:
        print(f"❌ ONNX 模型验证失败: {e}")
        return False
    
    print()
    
    # 5. 显示模型信息
    print(f"📊 模型信息:")
    print(f"   文件大小: {output_path.stat().st_size / 1024 / 1024:.2f} MB")
    print(f"   输入形状: (batch, 80, n_frames)")
    print(f"   输出形状: (batch, 7)")
    print(f"   参数数量: {sum(p.numel() for p in model.parameters()):,}")
    print()
    
    # 6. 测试推理
    print(f"🧪 测试 ONNX 推理...")
    try:
        import onnxruntime as ort
        
        session = ort.InferenceSession(str(output_path))
        
        # 准备输入
        input_data = np.random.randn(1, 80, 100).astype(np.float32)
        
        # 运行推理
        outputs = session.run(None, {'audio_input': input_data})
        
        print(f"✅ ONNX 推理成功")
        print(f"   输出形状: {outputs[0].shape}")
        print(f"   输出值: {outputs[0][0]}")
        
        # 应用 softmax
        logits = outputs[0][0]
        probs = np.exp(logits) / np.sum(np.exp(logits))
        
        print()
        print(f"📊 情感概率分布:")
        for i, (label, prob) in enumerate(zip(emotion_labels, probs)):
            print(f"   {label:12s}: {prob:.4f}")
        
        predicted_emotion = emotion_labels[np.argmax(probs)]
        print()
        print(f"🎯 预测情感: {predicted_emotion} (置信度: {np.max(probs):.4f})")
        
    except Exception as e:
        print(f"❌ ONNX 推理失败: {e}")
        return False
    
    print()
    print(f"=" * 80)
    print(f"🎉 Emotion Detection 模型创建完成！")
    print(f"=" * 80)
    print()
    print(f"📝 注意事项:")
    print(f"   1. 这是一个随机初始化的模型")
    print(f"   2. 实际使用需要在情感数据集上训练")
    print(f"   3. 或者使用预训练的 Wav2Vec2 + 情感分类头")
    print()
    
    return True


def download_pretrained_emotion_model(output_dir="/work/models"):
    """
    下载预训练的情感检测模型（如果可用）
    
    注意: 这需要 transformers 库和 HuggingFace 模型
    """
    print(f"=" * 80)
    print(f"🎯 下载预训练的 Emotion Detection 模型")
    print(f"=" * 80)
    print()
    
    try:
        from transformers import Wav2Vec2ForSequenceClassification, Wav2Vec2Processor
        
        model_name = "ehcalabres/wav2vec2-lg-xlsr-en-speech-emotion-recognition"
        
        print(f"📥 下载模型: {model_name}")
        print(f"   (这可能需要几分钟...)")
        
        processor = Wav2Vec2Processor.from_pretrained(model_name)
        model = Wav2Vec2ForSequenceClassification.from_pretrained(model_name)
        model.eval()
        
        print(f"✅ 模型下载成功")
        print()
        
        # 导出为 ONNX
        # 注意: Wav2Vec2 模型较大，导出可能需要一些时间
        print(f"💾 导出 ONNX 模型...")
        print(f"   ⚠️ 这可能需要几分钟...")
        
        # 创建示例输入
        dummy_input = torch.randn(1, 16000)  # 1 秒音频 @ 16kHz
        
        output_path = Path(output_dir) / "emotion-model.onnx"
        output_path.parent.mkdir(parents=True, exist_ok=True)
        
        torch.onnx.export(
            model,
            dummy_input,
            str(output_path),
            export_params=True,
            opset_version=14,
            input_names=['audio_input'],
            output_names=['emotion_output'],
            dynamic_axes={
                'audio_input': {0: 'batch_size', 1: 'sequence_length'}
            }
        )
        
        print(f"✅ ONNX 模型导出成功")
        print(f"   文件大小: {output_path.stat().st_size / 1024 / 1024:.2f} MB")
        
        return True
        
    except ImportError:
        print(f"❌ 需要安装 transformers 库:")
        print(f"   pip install transformers")
        return False
    except Exception as e:
        print(f"❌ 下载失败: {e}")
        return False


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="创建 Emotion Detection 模型")
    parser.add_argument(
        "--output-dir",
        type=str,
        default="/work/models",
        help="输出目录"
    )
    parser.add_argument(
        "--pretrained",
        action="store_true",
        help="下载预训练模型（需要 transformers 库）"
    )
    
    args = parser.parse_args()
    
    if args.pretrained:
        success = download_pretrained_emotion_model(args.output_dir)
    else:
        success = create_emotion_model(args.output_dir)
    
    sys.exit(0 if success else 1)

