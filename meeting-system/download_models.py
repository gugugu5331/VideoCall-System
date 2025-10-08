#!/usr/bin/env python3
"""
下载所有AI模型到本地
"""
import os
import sys
from huggingface_hub import snapshot_download

# 模型配置 - 使用确认可公开访问的轻量级模型
MODELS = {
    "speech_recognition": {
        "source": "huggingface",
        "model_id": "openai/whisper-tiny",
        "path": "/models/speech_recognition",
        "description": "Whisper Tiny - 语音识别模型 (39MB)"
    },
    "emotion_detection": {
        "source": "huggingface",
        "model_id": "j-hartmann/emotion-english-distilroberta-base",
        "path": "/models/emotion_detection",
        "description": "DistilRoBERTa - 情绪检测模型 (82MB)"
    },
    "text_summarization": {
        "source": "huggingface",
        "model_id": "sshleifer/distilbart-cnn-6-6",
        "path": "/models/text_summarization",
        "description": "DistilBART - 文本摘要模型 (306MB)"
    },
    "audio_denoising": {
        "source": "huggingface",
        "model_id": "speechbrain/sepformer-wham",
        "path": "/models/audio_denoising",
        "description": "SepFormer - 音频分离/降噪模型"
    },
    "video_enhancement": {
        "source": "huggingface",
        "model_id": "caidas/swin2SR-classical-sr-x2-64",
        "path": "/models/video_enhancement",
        "description": "Swin2SR - 图像/视频超分辨率模型"
    },
    "audio_deepfake": {
        "source": "huggingface",
        "model_id": "microsoft/wavlm-base-plus",
        "path": "/models/audio_deepfake",
        "description": "WavLM - 音频表征模型(可用于伪造检测)"
    },
    "face_deepfake": {
        "source": "huggingface",
        "model_id": "google/vit-base-patch16-224",
        "path": "/models/face_deepfake",
        "description": "ViT - 视觉Transformer(可用于人脸检测)"
    }
}

def download_model(name, config):
    """下载单个模型"""
    print(f"\n{'='*80}")
    print(f"开始下载: {config['description']}")
    print(f"模型ID: {config['model_id']}")
    print(f"保存路径: {config['path']}")
    print(f"{'='*80}\n")
    
    try:
        snapshot_download(
            repo_id=config['model_id'],
            local_dir=config['path'],
            local_dir_use_symlinks=False,
            resume_download=True
        )

        print(f"\n✅ {name} 下载完成!")
        return True
    except Exception as e:
        print(f"\n❌ {name} 下载失败: {e}")
        import traceback
        traceback.print_exc()
        return False

def main():
    print("开始下载所有AI模型...")
    print(f"总共需要下载 {len(MODELS)} 个模型\n")
    
    results = {}
    for name, config in MODELS.items():
        results[name] = download_model(name, config)
    
    # 打印总结
    print("\n" + "="*80)
    print("下载总结:")
    print("="*80)
    
    success_count = sum(1 for v in results.values() if v)
    fail_count = len(results) - success_count
    
    for name, success in results.items():
        status = "✅ 成功" if success else "❌ 失败"
        print(f"{status} - {name}: {MODELS[name]['description']}")
    
    print(f"\n总计: {success_count}/{len(MODELS)} 成功, {fail_count} 失败")
    
    return 0 if fail_count == 0 else 1

if __name__ == "__main__":
    sys.exit(main())

