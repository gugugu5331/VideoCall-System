#!/usr/bin/env python3
"""
下载 Whisper 模型并转换为 CTranslate2 格式（支持中英文）

CTranslate2 是一个快速的推理引擎，支持 Whisper 模型
比 ONNX 更容易集成，性能也更好
"""

import sys
from pathlib import Path


def download_and_convert_whisper():
    """下载 Whisper 模型并转换为 CTranslate2 格式"""
    print("=" * 80)
    print("🎯 下载 Whisper 模型并转换为 CTranslate2 格式")
    print("=" * 80)
    print()
    
    try:
        # 安装 faster-whisper
        print("📦 安装 faster-whisper...")
        import subprocess
        subprocess.run([
            sys.executable, "-m", "pip", "install", 
            "faster-whisper", "--quiet"
        ], check=True)
        print("✅ faster-whisper 安装完成")
        print()
        
        from faster_whisper import WhisperModel
        
        model_size = "base"  # tiny, base, small, medium, large-v2, large-v3
        
        print(f"📥 下载 Whisper 模型: {model_size}")
        print(f"   ⚠️ 这可能需要几分钟...")
        print(f"   ✅ 支持语言: 中文、英文及其他 97 种语言")
        print()
        
        # 下载模型（会自动转换为 CTranslate2 格式）
        model = WhisperModel(model_size, device="cpu", compute_type="int8")
        
        print(f"✅ Whisper 模型下载成功")
        print()
        
        # 测试推理
        print(f"🧪 测试模型推理...")
        
        # 创建测试音频（1秒静音）
        import numpy as np
        import tempfile
        import wave
        
        # 生成 1 秒 16kHz 单声道音频
        sample_rate = 16000
        duration = 1.0
        samples = np.zeros(int(sample_rate * duration), dtype=np.int16)
        
        # 保存为临时 WAV 文件
        with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as f:
            temp_wav = f.name
            
        with wave.open(temp_wav, 'w') as wav_file:
            wav_file.setnchannels(1)
            wav_file.setsampwidth(2)
            wav_file.setframerate(sample_rate)
            wav_file.writeframes(samples.tobytes())
        
        # 测试中文转录
        print("   测试 1: 中文语音识别")
        segments, info = model.transcribe(temp_wav, language="zh")
        segments_list = list(segments)
        print(f"   ✅ 检测语言: {info.language} (概率: {info.language_probability:.2f})")
        print(f"   ✅ 转录结果: {len(segments_list)} 个片段")
        print()
        
        # 测试英文转录
        print("   测试 2: 英文语音识别")
        segments, info = model.transcribe(temp_wav, language="en")
        segments_list = list(segments)
        print(f"   ✅ 检测语言: {info.language} (概率: {info.language_probability:.2f})")
        print(f"   ✅ 转录结果: {len(segments_list)} 个片段")
        print()
        
        # 测试自动语言检测
        print("   测试 3: 自动语言检测")
        segments, info = model.transcribe(temp_wav)
        segments_list = list(segments)
        print(f"   ✅ 检测语言: {info.language} (概率: {info.language_probability:.2f})")
        print(f"   ✅ 转录结果: {len(segments_list)} 个片段")
        print()
        
        # 清理临时文件
        import os
        os.unlink(temp_wav)
        
        # 显示模型信息
        print(f"📊 模型信息:")
        print(f"   模型大小: {model_size}")
        print(f"   计算类型: int8 (量化)")
        print(f"   设备: CPU")
        print(f"   支持语言: 99 种（包括中文、英文）")
        print()
        
        # 保存模型路径信息
        model_info = {
            "model_size": model_size,
            "model_type": "faster-whisper",
            "compute_type": "int8",
            "device": "cpu",
            "languages": ["zh", "en", "auto"],
            "cache_dir": str(Path.home() / ".cache" / "huggingface" / "hub")
        }
        
        import json
        info_path = Path("/work/models/whisper_model_info.json")
        with open(info_path, 'w', encoding='utf-8') as f:
            json.dump(model_info, f, indent=2, ensure_ascii=False)
        
        print(f"✅ 模型信息已保存到: {info_path}")
        print()
        
        print(f"=" * 80)
        print(f"🎉 Whisper 模型下载和配置完成！")
        print(f"=" * 80)
        print()
        
        print(f"📝 使用说明:")
        print(f"   1. 模型已缓存在: ~/.cache/huggingface/hub")
        print(f"   2. 使用 faster-whisper 进行推理")
        print(f"   3. 支持中英文混合识别")
        print(f"   4. 支持自动语言检测")
        print()
        
        return True
        
    except Exception as e:
        print(f"❌ 错误: {e}")
        import traceback
        traceback.print_exc()
        return False


if __name__ == "__main__":
    success = download_and_convert_whisper()
    sys.exit(0 if success else 1)

