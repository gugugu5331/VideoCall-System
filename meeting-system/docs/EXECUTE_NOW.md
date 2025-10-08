# 立即执行 - AI模型部署

## ✅ 已完成的工作

1. ✅ 删除所有模拟/降级逻辑
2. ✅ 创建Python推理服务 (`Dockerfile.inference`, `inference_server.py`)
3. ✅ 创建模型下载脚本 (`download_all_models.sh`)
4. ✅ 更新AI服务调用Python推理
5. ✅ 更新docker-compose.yml配置

## 🚀 立即执行以下命令

### 步骤1：下载所有AI模型（约1.3GB）

```bash
cd /root/meeting-system-server/meeting-system

# 创建模型目录
mkdir -p /models/{speech_recognition,emotion_detection,text_summarization,audio_denoising,video_enhancement,audio_deepfake,face_deepfake}

# 安装Python依赖
pip3 install huggingface_hub transformers torch

# 下载模型（这将需要20-30分钟）
python3 << 'EOF'
from huggingface_hub import snapshot_download

models = [
    ("openai/whisper-tiny", "/models/speech_recognition"),
    ("j-hartmann/emotion-english-distilroberta-base", "/models/emotion_detection"),
    ("sshleifer/distilbart-cnn-6-6", "/models/text_summarization"),
    ("speechbrain/sepformer-wham", "/models/audio_denoising"),
    ("caidas/swin2SR-classical-sr-x2-64", "/models/video_enhancement"),
    ("microsoft/wavlm-base-plus", "/models/audio_deepfake"),
    ("google/vit-base-patch16-224", "/models/face_deepfake"),
]

for model_id, path in models:
    print(f"\n下载: {model_id}")
    try:
        snapshot_download(repo_id=model_id, local_dir=path, local_dir_use_symlinks=False)
        print(f"✓ {model_id} 完成")
    except Exception as e:
        print(f"✗ {model_id} 失败: {e}")

print("\n所有模型下载完成！")
EOF

# 验证下载
du -sh /models/*
```

### 步骤2：构建并启动Python推理服务

```bash
cd /root/meeting-system-server/meeting-system

# 构建Python推理服务
docker-compose build python-inference

# 启动服务
docker-compose up -d python-inference

# 检查状态
docker-compose ps python-inference
docker-compose logs python-inference
```

### 步骤3：重新构建并启动AI服务

```bash
# 重新构建AI服务
docker-compose build ai-service

# 重启AI服务
docker-compose up -d ai-service

# 检查状态
docker-compose ps ai-service
docker-compose logs --tail=50 ai-service
```

### 步骤4：运行E2E测试

```bash
cd /root/meeting-system-server/meeting-system/backend/tests

# 运行完整测试
go test -v -run TestE2EIntegration

# 预期：所有AI模型测试通过，成功率100%
```

## 🔍 验证命令

```bash
# 检查模型是否下载
ls -lah /models/
du -sh /models/*

# 检查Python推理容器
docker exec meeting-python-inference ls -lah /models/

# 测试AI服务
curl http://localhost:8800/api/v1/models | python3 -m json.tool

# 查看服务日志
docker-compose logs -f ai-service
docker-compose logs -f python-inference
```

## ⚠️ 重要提示

- 模型下载需要20-30分钟，请耐心等待
- 确保有足够的磁盘空间（至少2GB）
- 如果GPU不可用，编辑docker-compose.yml删除`runtime: nvidia`行
- 所有推理必须使用真实模型，不允许任何模拟

## 📊 预期结果

E2E测试输出应该显示：

```
=== 步骤7: AI服务完整测试 ===
✓ 找到 5 个AI模型
[1/5] 测试模型: Audio Denoising Model - ✓ 模型测试成功
[2/5] 测试模型: Video Enhancement Model - ✓ 模型测试成功
[3/5] 测试模型: Speech Recognition Model - ✓ 模型测试成功
[4/5] 测试模型: Text Summarization Model - ✓ 模型测试成功
[5/5] 测试模型: Emotion Detection Model - ✓ 模型测试成功

总模型数: 5
测试成功: 5
测试失败: 0
成功率: 100.0%

🎉 所有测试通过！系统运行正常！
```

