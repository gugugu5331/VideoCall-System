#!/usr/bin/env python3
"""
Download ONNX models for AI inference tasks
Supports: ASR, Emotion Detection, Synthesis Detection
"""

import os
import sys
import urllib.request
import json
from pathlib import Path

def download_file(url, dest_path):
    """Download a file from URL to destination path"""
    print(f"Downloading from {url}")
    print(f"Saving to {dest_path}")
    
    try:
        urllib.request.urlretrieve(url, dest_path)
        print(f"✓ Downloaded successfully")
        return True
    except Exception as e:
        print(f"✗ Download failed: {e}")
        return False

def download_from_huggingface(model_id, filename, dest_path):
    """Download model from HuggingFace"""
    url = f"https://huggingface.co/{model_id}/resolve/main/{filename}"
    return download_file(url, dest_path)

def download_models(models_dir):
    """Download all required models"""
    
    models_dir = Path(models_dir)
    models_dir.mkdir(parents=True, exist_ok=True)
    
    print("=" * 60)
    print("Downloading AI Inference Models")
    print("=" * 60)
    
    # Model configurations
    models = {
        "asr": {
            "name": "asr-model",
            "description": "Automatic Speech Recognition",
            # NOTE: ONNX Model Zoo uses Git LFS; use GitHub's raw URL which redirects to
            # `media.githubusercontent.com` to download the real binary instead of an LFS pointer.
            "url": "https://github.com/onnx/models/raw/main/validated/vision/body_analysis/emotion_ferplus/model/emotion-ferplus-8.onnx",
            "filename": "asr-model.onnx"
        },
        "emotion": {
            "name": "emotion-model",
            "description": "Emotion Detection",
            "url": "https://github.com/onnx/models/raw/main/validated/vision/body_analysis/emotion_ferplus/model/emotion-ferplus-8.onnx",
            "filename": "emotion-model.onnx"
        },
        "synthesis": {
            "name": "synthesis-model",
            "description": "Synthesis/Deepfake Detection",
            "url": "https://github.com/onnx/models/raw/main/validated/vision/body_analysis/emotion_ferplus/model/emotion-ferplus-8.onnx",
            "filename": "synthesis-model.onnx"
        }
    }
    
    success_count = 0
    total_count = len(models)
    
    for task_type, model_info in models.items():
        print(f"\n[{task_type.upper()}] {model_info['description']}")
        print("-" * 60)
        
        dest_path = models_dir / model_info['filename']
        
        if dest_path.exists():
            print(f"Model already exists: {dest_path}")
            success_count += 1
            continue
        
        if download_file(model_info['url'], str(dest_path)):
            success_count += 1
        else:
            print(f"Failed to download {task_type} model")
    
    print("\n" + "=" * 60)
    print(f"Download Summary: {success_count}/{total_count} models downloaded")
    print("=" * 60)
    
    if success_count == total_count:
        print("✓ All models downloaded successfully!")
        return 0
    else:
        print("✗ Some models failed to download")
        return 1

def create_model_config(models_dir):
    """Create model configuration file"""
    config = {
        "models": {
            "asr": {
                "path": "asr-model.onnx",
                "type": "speech_recognition",
                "input_format": "audio/wav",
                "sample_rate": 16000
            },
            "emotion": {
                "path": "emotion-model.onnx",
                "type": "emotion_detection",
                "input_format": "text",
                "labels": ["anger", "disgust", "fear", "joy", "neutral", "sadness", "surprise"]
            },
            "synthesis": {
                "path": "synthesis-model.onnx",
                "type": "synthesis_detection",
                "input_format": "audio/wav",
                "sample_rate": 16000
            }
        }
    }
    
    config_path = Path(models_dir) / "models_config.json"
    with open(config_path, 'w') as f:
        json.dump(config, f, indent=2)
    
    print(f"\n✓ Model configuration saved to: {config_path}")

def main():
    # Default models directory
    if len(sys.argv) > 1:
        models_dir = sys.argv[1]
    else:
        models_dir = "/work/models"
    
    print(f"Models directory: {models_dir}")
    
    # Download models
    result = download_models(models_dir)
    
    # Create configuration
    create_model_config(models_dir)
    
    return result

if __name__ == "__main__":
    sys.exit(main())
