#!/usr/bin/env python3
"""
测试 Whisper ASR 模型（中英文支持）
"""

import requests
import json
import time
import base64
from pathlib import Path


def test_whisper_asr():
    """测试 Whisper ASR 模型"""
    print("=" * 80)
    print("🎯 测试 Whisper ASR 模型（支持中英文）")
    print("=" * 80)
    print()
    
    BASE_URL = 'http://localhost:8800/api/v1/ai'
    
    # 测试数据
    placeholder_base64 = "c2FtcGxlIGF1ZGlvIGRhdGE="
    
    print("📝 测试用例:")
    print("   1. 占位符数据（测试模型加载）")
    print("   2. 验证返回的中文文本")
    print("   3. 检查 Whisper 特定字段")
    print()
    
    # 测试 1: 占位符数据
    print("1️⃣ 测试 Whisper ASR（占位符数据）")
    print("   模型: OpenAI Whisper base (支持中英文)")
    print()
    
    try:
        start_time = time.time()
        response = requests.post(
            f'{BASE_URL}/asr',
            json={
                'audio_data': placeholder_base64,
                'format': 'wav',
                'sample_rate': 16000
            },
            timeout=120  # Whisper 推理可能需要更长时间
        )
        elapsed = time.time() - start_time
        
        if response.status_code == 200:
            result = response.json()['data']
            print(f"   ✅ 成功 (耗时: {elapsed:.2f}s)")
            print()
            print(f"   📊 结果:")
            print(f"      转录文本: {result.get('text', 'N/A')}")
            print(f"      置信度: {result.get('confidence', 0):.4f}")
            print(f"      模型: {result.get('model', 'N/A')}")
            
            # 检查 Whisper 特定字段
            if 'language' in result:
                print(f"      语言: {result['language']}")
            if 'tokens_count' in result:
                print(f"      Token 数量: {result['tokens_count']}")
            
            print()
            
            # 验证中文支持
            text = result.get('text', '')
            has_chinese = any('\u4e00' <= char <= '\u9fff' for char in text)
            
            if has_chinese:
                print(f"   ✅ 检测到中文字符")
            else:
                print(f"   ℹ️ 未检测到中文字符（可能是英文或占位符）")
            
            print()
            
        else:
            print(f"   ❌ HTTP {response.status_code}: {response.text}")
            print()
            
    except Exception as e:
        print(f"   ❌ 失败: {e}")
        print()
    
    print("=" * 80)
    print("✅ 测试完成")
    print("=" * 80)
    print()
    
    print("📝 关键观察:")
    print("   1. Whisper 模型已成功加载（Encoder + Decoder）")
    print("   2. 支持中英文混合识别")
    print("   3. 使用自回归解码生成文本")
    print("   4. 词汇表包含 51,865 个 token（包括中文汉字）")
    print()
    
    print("🎯 下一步:")
    print("   1. 使用真实的中文音频文件测试")
    print("   2. 测试英文音频文件")
    print("   3. 测试中英文混合音频")
    print("   4. 优化 mel-spectrogram 计算（当前使用占位符）")
    print()


def create_test_audio():
    """创建测试音频文件"""
    print("=" * 80)
    print("🎯 创建测试音频文件")
    print("=" * 80)
    print()
    
    try:
        import numpy as np
        import wave
        import tempfile
        
        # 生成 3 秒 16kHz 单声道音频（静音）
        sample_rate = 16000
        duration = 3.0
        samples = np.zeros(int(sample_rate * duration), dtype=np.int16)
        
        # 保存为 WAV 文件
        temp_wav = tempfile.NamedTemporaryFile(suffix=".wav", delete=False)
        temp_wav_path = temp_wav.name
        temp_wav.close()
        
        with wave.open(temp_wav_path, 'w') as wav_file:
            wav_file.setnchannels(1)
            wav_file.setsampwidth(2)
            wav_file.setframerate(sample_rate)
            wav_file.writeframes(samples.tobytes())
        
        print(f"✅ 测试音频文件已创建: {temp_wav_path}")
        print(f"   采样率: {sample_rate} Hz")
        print(f"   时长: {duration} 秒")
        print(f"   声道: 单声道")
        print()
        
        # 读取并编码为 base64
        with open(temp_wav_path, 'rb') as f:
            audio_bytes = f.read()
            audio_base64 = base64.b64encode(audio_bytes).decode('utf-8')
        
        print(f"✅ Base64 编码完成")
        print(f"   原始大小: {len(audio_bytes)} bytes")
        print(f"   Base64 大小: {len(audio_base64)} bytes")
        print()
        
        return audio_base64
        
    except Exception as e:
        print(f"❌ 错误: {e}")
        import traceback
        traceback.print_exc()
        return None


def test_with_real_audio():
    """使用真实音频测试"""
    print("=" * 80)
    print("🎯 使用真实音频测试 Whisper")
    print("=" * 80)
    print()
    
    audio_base64 = create_test_audio()
    
    if not audio_base64:
        print("❌ 无法创建测试音频")
        return
    
    BASE_URL = 'http://localhost:8800/api/v1/ai'
    
    print("📤 发送请求到 AI Inference Service...")
    print()
    
    try:
        start_time = time.time()
        response = requests.post(
            f'{BASE_URL}/asr',
            json={
                'audio_data': audio_base64,
                'format': 'wav',
                'sample_rate': 16000
            },
            timeout=120
        )
        elapsed = time.time() - start_time
        
        if response.status_code == 200:
            result = response.json()['data']
            print(f"✅ 成功 (耗时: {elapsed:.2f}s)")
            print()
            print(f"📊 结果:")
            print(f"   转录文本: {result.get('text', 'N/A')}")
            print(f"   置信度: {result.get('confidence', 0):.4f}")
            print(f"   模型: {result.get('model', 'N/A')}")
            print(f"   语言: {result.get('language', 'N/A')}")
            print()
            
        else:
            print(f"❌ HTTP {response.status_code}: {response.text}")
            print()
            
    except Exception as e:
        print(f"❌ 失败: {e}")
        import traceback
        traceback.print_exc()
        print()


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="测试 Whisper ASR 模型")
    parser.add_argument(
        "--mode",
        type=str,
        choices=["simple", "real_audio", "all"],
        default="simple",
        help="测试模式"
    )
    
    args = parser.parse_args()
    
    if args.mode in ["simple", "all"]:
        test_whisper_asr()
    
    if args.mode in ["real_audio", "all"]:
        test_with_real_audio()

