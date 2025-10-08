#!/usr/bin/env python3
"""
Synthesis Detection 模型转换为 ONNX 格式

用于检测音频是否为 AI 生成（深度伪造检测）
输出: 真实/合成的概率
"""

import torch
import torch.nn as nn
import onnx
import numpy as np
from pathlib import Path
import sys


class SimpleSynthesisDetector(nn.Module):
    """
    简化的深度伪造检测模型
    
    架构:
    - 输入: 音频特征 (mel-spectrogram)
    - CNN 层提取局部特征
    - LSTM 层提取时序特征
    - 全连接层二分类
    - 输出: 合成概率
    """
    
    def __init__(self, n_mels=80, hidden_size=128):
        super().__init__()
        
        # CNN 层
        self.conv1 = nn.Conv1d(n_mels, 128, kernel_size=3, padding=1)
        self.conv2 = nn.Conv1d(128, 256, kernel_size=3, padding=1)
        self.pool = nn.MaxPool1d(2)
        self.dropout1 = nn.Dropout(0.3)
        
        # LSTM 层
        self.lstm = nn.LSTM(
            input_size=256,
            hidden_size=hidden_size,
            num_layers=2,
            batch_first=True,
            bidirectional=True,
            dropout=0.3
        )
        
        # 注意力层
        self.attention = nn.Linear(hidden_size * 2, 1)
        
        # 分类层
        self.fc1 = nn.Linear(hidden_size * 2, 64)
        self.dropout2 = nn.Dropout(0.3)
        self.fc2 = nn.Linear(64, 1)  # 二分类：真实 vs 合成
    
    def forward(self, x):
        # x: (batch, n_mels, n_frames)
        
        # CNN
        x = torch.relu(self.conv1(x))
        x = self.pool(x)
        x = torch.relu(self.conv2(x))
        x = self.pool(x)
        x = self.dropout1(x)
        
        # 转换为 LSTM 输入格式
        x = x.transpose(1, 2)  # (batch, n_frames, 256)
        
        # LSTM
        lstm_out, _ = self.lstm(x)  # (batch, n_frames, hidden_size*2)
        
        # Attention pooling
        attention_weights = torch.softmax(self.attention(lstm_out), dim=1)
        attended = torch.sum(lstm_out * attention_weights, dim=1)
        
        # Classification
        x = torch.relu(self.fc1(attended))
        x = self.dropout2(x)
        output = self.fc2(x)  # (batch, 1)
        
        return output


def create_synthesis_model(output_dir="/work/models"):
    """
    创建深度伪造检测模型并转换为 ONNX 格式
    
    Args:
        output_dir: 输出目录
    """
    print(f"=" * 80)
    print(f"🎯 创建 Synthesis Detection 模型")
    print(f"=" * 80)
    print()
    
    print(f"📊 模型配置:")
    print(f"   输入: mel-spectrogram (80 x n_frames)")
    print(f"   输出: 合成概率 (0-1)")
    print(f"   任务: 二分类（真实 vs 合成）")
    print()
    
    # 1. 创建模型
    print(f"🔧 创建模型...")
    model = SimpleSynthesisDetector(n_mels=80, hidden_size=128)
    model.eval()
    
    # 初始化权重
    print(f"⚙️ 初始化模型权重...")
    for name, param in model.named_parameters():
        if 'weight' in name:
            if 'conv' in name or 'fc' in name:
                nn.init.xavier_uniform_(param)
            else:
                nn.init.orthogonal_(param)
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
    output_path = Path(output_dir) / "synthesis-model.onnx"
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
            output_names=['synthesis_output'],
            dynamic_axes={
                'audio_input': {0: 'batch_size', 2: 'n_frames'},
                'synthesis_output': {0: 'batch_size'}
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
    print(f"   输出形状: (batch, 1)")
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
        print(f"   原始输出: {outputs[0][0][0]:.4f}")
        
        # 应用 sigmoid
        raw_score = outputs[0][0][0]
        probability = 1.0 / (1.0 + np.exp(-raw_score))
        
        print()
        print(f"📊 检测结果:")
        print(f"   合成概率: {probability:.4f}")
        print(f"   真实概率: {1 - probability:.4f}")
        
        if probability > 0.5:
            print(f"   🎯 判断: 合成音频 (置信度: {probability:.4f})")
        else:
            print(f"   🎯 判断: 真实音频 (置信度: {1 - probability:.4f})")
        
    except Exception as e:
        print(f"❌ ONNX 推理失败: {e}")
        return False
    
    print()
    print(f"=" * 80)
    print(f"🎉 Synthesis Detection 模型创建完成！")
    print(f"=" * 80)
    print()
    print(f"📝 注意事项:")
    print(f"   1. 这是一个随机初始化的模型")
    print(f"   2. 实际使用需要在真实/合成音频数据集上训练")
    print(f"   3. 推荐数据集: ASVspoof 2019/2021")
    print(f"   4. 或者使用预训练的 RawNet2/AASIST 模型")
    print()
    
    return True


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="创建 Synthesis Detection 模型")
    parser.add_argument(
        "--output-dir",
        type=str,
        default="/work/models",
        help="输出目录"
    )
    
    args = parser.parse_args()
    
    success = create_synthesis_model(args.output_dir)
    
    sys.exit(0 if success else 1)

