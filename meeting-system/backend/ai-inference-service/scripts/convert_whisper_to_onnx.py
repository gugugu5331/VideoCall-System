#!/usr/bin/env python3
"""
Whisper ASR 模型转换为 ONNX 格式

使用 OpenAI Whisper 模型进行语音识别
支持多语言（包括中文）
"""

import torch
import onnx
import numpy as np
from pathlib import Path
import sys

def convert_whisper_to_onnx(model_size="base", output_dir="/work/models"):
    """
    将 Whisper 模型转换为 ONNX 格式

    Args:
        model_size: 模型大小 (tiny, base, small, medium, large)
        output_dir: 输出目录
    """
    try:
        import whisper
    except ImportError:
        print(f"❌ 错误: 需要安装 openai-whisper")
        print(f"   pip install openai-whisper")
        return False

    print(f"=" * 80)
    print(f"🎯 开始转换 Whisper {model_size} 模型到 ONNX 格式")
    print(f"=" * 80)
    print()

    # 1. 下载 Whisper 模型
    print(f"📥 下载 Whisper {model_size} 模型...")
    try:
        model = whisper.load_model(model_size)
        print(f"✅ 模型下载成功")
    except Exception as e:
        print(f"❌ 模型下载失败: {e}")
        return False
    
    print()
    
    # 2. 准备导出
    print(f"🔧 准备导出编码器...")
    
    # Whisper 模型包含编码器和解码器
    # 为了简化，我们只导出编码器部分
    # 编码器将音频特征转换为隐藏状态
    encoder = model.encoder
    encoder.eval()
    
    # 3. 创建示例输入
    # Whisper 期望的输入是 mel-spectrogram
    # 形状: (batch_size, n_mels, n_frames)
    # n_mels = 80 (固定)
    # n_frames = 3000 (对应 30 秒音频)
    batch_size = 1
    n_mels = 80
    n_frames = 3000
    
    print(f"📊 创建示例输入: shape=({batch_size}, {n_mels}, {n_frames})")
    dummy_input = torch.randn(batch_size, n_mels, n_frames)
    
    # 4. 导出为 ONNX
    output_path = Path(output_dir) / "asr-model.onnx"
    output_path.parent.mkdir(parents=True, exist_ok=True)
    
    print(f"💾 导出 ONNX 模型到: {output_path}")
    
    try:
        torch.onnx.export(
            encoder,
            dummy_input,
            str(output_path),
            export_params=True,
            opset_version=14,
            do_constant_folding=True,
            input_names=['audio_input'],
            output_names=['transcription_output'],
            dynamic_axes={
                'audio_input': {0: 'batch_size', 2: 'n_frames'},
                'transcription_output': {0: 'batch_size', 1: 'sequence_length'}
            }
        )
        print(f"✅ ONNX 模型导出成功")
    except Exception as e:
        print(f"❌ ONNX 导出失败: {e}")
        return False
    
    print()
    
    # 5. 验证 ONNX 模型
    print(f"🔍 验证 ONNX 模型...")
    try:
        onnx_model = onnx.load(str(output_path))
        onnx.checker.check_model(onnx_model)
        print(f"✅ ONNX 模型验证通过")
    except Exception as e:
        print(f"❌ ONNX 模型验证失败: {e}")
        return False
    
    print()
    
    # 6. 显示模型信息
    print(f"📊 模型信息:")
    print(f"   文件大小: {output_path.stat().st_size / 1024 / 1024:.2f} MB")
    print(f"   输入: audio_input (mel-spectrogram)")
    print(f"   输出: transcription_output (hidden states)")
    print()
    
    # 7. 测试推理
    print(f"🧪 测试 ONNX 推理...")
    try:
        import onnxruntime as ort
        
        session = ort.InferenceSession(str(output_path))
        
        # 准备输入
        input_data = np.random.randn(1, 80, 3000).astype(np.float32)
        
        # 运行推理
        outputs = session.run(None, {'audio_input': input_data})
        
        print(f"✅ ONNX 推理成功")
        print(f"   输出形状: {outputs[0].shape}")
    except Exception as e:
        print(f"❌ ONNX 推理失败: {e}")
        return False
    
    print()
    print(f"=" * 80)
    print(f"🎉 Whisper 模型转换完成！")
    print(f"=" * 80)
    print()
    print(f"📝 注意事项:")
    print(f"   1. 此模型只包含编码器部分")
    print(f"   2. 需要在 C++ 代码中实现解码逻辑")
    print(f"   3. 或者使用简化的 CTC 解码")
    print()
    
    return True


def create_simple_asr_model(output_dir="/work/models"):
    """
    创建一个简化的 ASR 模型用于演示
    
    这个模型会：
    1. 接收音频特征
    2. 通过简单的神经网络
    3. 输出字符概率
    """
    print(f"=" * 80)
    print(f"🎯 创建简化的 ASR 演示模型")
    print(f"=" * 80)
    print()
    
    class SimpleASR(torch.nn.Module):
        def __init__(self):
            super().__init__()
            # 简单的 LSTM + 线性层
            self.lstm = torch.nn.LSTM(
                input_size=80,  # mel-spectrogram features
                hidden_size=256,
                num_layers=2,
                batch_first=True,
                bidirectional=True
            )
            self.fc = torch.nn.Linear(512, 100)  # 100 个字符类别
        
        def forward(self, x):
            # x: (batch, n_mels, n_frames)
            x = x.transpose(1, 2)  # (batch, n_frames, n_mels)
            lstm_out, _ = self.lstm(x)  # (batch, n_frames, 512)
            output = self.fc(lstm_out)  # (batch, n_frames, 100)
            return output
    
    print(f"🔧 创建模型...")
    model = SimpleASR()
    model.eval()
    
    # 创建示例输入
    dummy_input = torch.randn(1, 80, 100)  # (batch, n_mels, n_frames)
    
    # 导出为 ONNX
    output_path = Path(output_dir) / "asr-model.onnx"
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
            output_names=['transcription_output'],
            dynamic_axes={
                'audio_input': {0: 'batch_size', 2: 'n_frames'},
                'transcription_output': {0: 'batch_size', 1: 'n_frames'}
            }
        )
        print(f"✅ ONNX 模型导出成功")
    except Exception as e:
        print(f"❌ ONNX 导出失败: {e}")
        return False
    
    print()
    
    # 验证模型
    print(f"🔍 验证 ONNX 模型...")
    try:
        onnx_model = onnx.load(str(output_path))
        onnx.checker.check_model(onnx_model)
        print(f"✅ ONNX 模型验证通过")
    except Exception as e:
        print(f"❌ ONNX 模型验证失败: {e}")
        return False
    
    print()
    
    # 显示模型信息
    print(f"📊 模型信息:")
    print(f"   文件大小: {output_path.stat().st_size / 1024 / 1024:.2f} MB")
    print(f"   输入形状: (batch, 80, n_frames)")
    print(f"   输出形状: (batch, n_frames, 100)")
    print()
    
    # 测试推理
    print(f"🧪 测试 ONNX 推理...")
    try:
        import onnxruntime as ort
        
        session = ort.InferenceSession(str(output_path))
        input_data = np.random.randn(1, 80, 100).astype(np.float32)
        outputs = session.run(None, {'audio_input': input_data})
        
        print(f"✅ ONNX 推理成功")
        print(f"   输出形状: {outputs[0].shape}")
    except Exception as e:
        print(f"❌ ONNX 推理失败: {e}")
        return False
    
    print()
    print(f"=" * 80)
    print(f"🎉 简化 ASR 模型创建完成！")
    print(f"=" * 80)
    
    return True


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="转换 Whisper 模型到 ONNX 格式")
    parser.add_argument(
        "--model-size",
        type=str,
        default="base",
        choices=["tiny", "base", "small", "medium", "large"],
        help="Whisper 模型大小"
    )
    parser.add_argument(
        "--output-dir",
        type=str,
        default="/work/models",
        help="输出目录"
    )
    parser.add_argument(
        "--simple",
        action="store_true",
        help="创建简化的演示模型（不下载 Whisper）"
    )
    
    args = parser.parse_args()
    
    if args.simple:
        success = create_simple_asr_model(args.output_dir)
    else:
        success = convert_whisper_to_onnx(args.model_size, args.output_dir)
    
    sys.exit(0 if success else 1)

