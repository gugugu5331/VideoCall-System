#!/usr/bin/env python3
import os
import sys

# 配置使用国内镜像
os.environ['HF_ENDPOINT'] = 'https://hf-mirror.com'

try:
    from modelscope import snapshot_download
    use_modelscope = True
    print("使用ModelScope下载模型")
except ImportError:
    from huggingface_hub import snapshot_download
    use_modelscope = False
    print("使用HuggingFace镜像下载模型")

# 模型映射
MODELS = [
    {
        "name": "语音识别 - Whisper Tiny",
        "hf_id": "openai/whisper-tiny",
        "ms_id": "AI-ModelScope/whisper-tiny",
        "path": "/models/speech_recognition",
        "size": "39MB"
    },
    {
        "name": "情绪检测 - DistilRoBERTa",
        "hf_id": "j-hartmann/emotion-english-distilroberta-base",
        "ms_id": "AI-ModelScope/emotion-english-distilroberta-base",
        "path": "/models/emotion_detection",
        "size": "82MB"
    },
    {
        "name": "文本摘要 - DistilBART",
        "hf_id": "sshleifer/distilbart-cnn-6-6",
        "ms_id": "AI-ModelScope/distilbart-cnn-6-6",
        "path": "/models/text_summarization",
        "size": "306MB"
    },
    {
        "name": "音频降噪 - SepFormer",
        "hf_id": "speechbrain/sepformer-wham",
        "ms_id": "speechbrain/sepformer-wham",
        "path": "/models/audio_denoising",
        "size": "100MB"
    },
    {
        "name": "视频增强 - Swin2SR",
        "hf_id": "caidas/swin2SR-classical-sr-x2-64",
        "ms_id": "AI-ModelScope/swin2SR-classical-sr-x2-64",
        "path": "/models/video_enhancement",
        "size": "50MB"
    },
    {
        "name": "音频伪造检测 - WavLM",
        "hf_id": "microsoft/wavlm-base-plus",
        "ms_id": "AI-ModelScope/wavlm-base-plus",
        "path": "/models/audio_deepfake",
        "size": "378MB"
    },
    {
        "name": "人脸伪造检测 - ViT",
        "hf_id": "google/vit-base-patch16-224",
        "ms_id": "AI-ModelScope/vit-base-patch16-224",
        "path": "/models/face_deepfake",
        "size": "346MB"
    }
]

def download_model(model_info):
    name = model_info["name"]
    model_id = model_info["ms_id"] if use_modelscope else model_info["hf_id"]
    path = model_info["path"]
    size = model_info["size"]
    
    print(f"\n{'='*60}")
    print(f"下载: {name}")
    print(f"模型ID: {model_id}")
    print(f"大小: {size}")
    print(f"路径: {path}")
    print(f"{'='*60}")
    
    # 检查是否已存在
    if os.path.exists(path) and os.listdir(path):
        print(f"✓ 模型已存在，跳过")
        return True
    
    # 创建目录
    os.makedirs(path, exist_ok=True)
    
    try:
        if use_modelscope:
            snapshot_download(
                model_id=model_id,
                cache_dir=path,
                revision='master'
            )
        else:
            snapshot_download(
                repo_id=model_id,
                local_dir=path,
                local_dir_use_symlinks=False,
                resume_download=True
            )
        print(f"✓ {name} 下载完成")
        return True
    except Exception as e:
        print(f"✗ {name} 下载失败: {e}")
        return False

def main():
    print("="*60)
    print("开始下载所有AI模型")
    print("="*60)
    
    # 创建模型根目录
    os.makedirs("/models", exist_ok=True)
    
    success_count = 0
    fail_count = 0
    
    for model in MODELS:
        if download_model(model):
            success_count += 1
        else:
            fail_count += 1
    
    print(f"\n{'='*60}")
    print("下载完成")
    print(f"{'='*60}")
    print(f"成功: {success_count}/{len(MODELS)}")
    print(f"失败: {fail_count}/{len(MODELS)}")
    
    if fail_count > 0:
        sys.exit(1)

if __name__ == "__main__":
    main()

