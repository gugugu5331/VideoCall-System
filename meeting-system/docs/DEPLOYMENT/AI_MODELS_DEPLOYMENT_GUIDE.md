# AI模型部署指南

> 注意：本指南针对历史的 Python/ONNX 推理链路，已不再作为默认方案。当前 Triton/TensorRT 部署请参考 `docs/DEPLOYMENT/GPU_AI_NODES.md`。

## ✅ 已完成的工作

### 1. 删除所有模拟/降级逻辑
- ✅ 从 `ai_manager.go` 中删除 `getFallbackResponse()` 方法
- ✅ 恢复严格的真实推理要求
- ✅ 确保所有推理请求必须使用真实模型

### 2. 创建Python推理服务
- ✅ 创建 `Dockerfile.inference` - Python推理服务容器
- ✅ 创建 `inference_server.py` - 真实模型推理脚本
- ✅ 创建 `requirements.txt` - Python依赖列表
- ✅ 创建 `download_all_models.sh` - 自动下载所有7个模型

### 3. 更新AI服务
- ✅ 修改 `real_inference_service.go` 调用Python推理
- ✅ 实现 `callPythonInference()` 方法
- ✅ 更新所有推理方法（SpeechRecognition, EmotionDetection, Summarize等）

### 4. 更新Docker配置
- ✅ 在 `docker-compose.yml` 中添加 `python-inference` 服务
- ✅ 配置GPU支持（NVIDIA）
- ✅ 配置模型卷挂载

## 📋 需要手动执行的步骤

### 步骤1：构建Python推理服务

```bash
cd meeting-system

# 构建Python推理服务镜像（这将自动下载所有模型）
docker compose build python-inference

# 注意：这个过程可能需要30-60分钟，因为需要下载约1.3GB的模型文件
```

### 步骤2：启动Python推理服务

```bash
# 启动Python推理服务
docker compose up -d python-inference

# 检查服务状态
docker compose ps python-inference

# 查看日志
docker compose logs -f python-inference
```

### 步骤3：验证模型下载

```bash
# 进入容器检查模型
docker exec -it meeting-python-inference bash

# 在容器内执行
ls -lah /models/
du -sh /models/*

# 应该看到7个模型目录，每个都有模型文件
# speech_recognition/
# emotion_detection/
# text_summarization/
# audio_denoising/
# video_enhancement/
# audio_deepfake/
# face_deepfake/
```

### 步骤4：重新构建并启动AI服务

```bash
# 重新构建AI服务
docker compose build ai-service

# 重启AI服务
docker compose up -d ai-service

# 检查AI服务日志
docker compose logs -f ai-service
```

### 步骤5：运行E2E测试

```bash
cd meeting-system/backend/tests

# 运行完整的E2E测试
go test -v -run TestE2EIntegration

# 预期结果：所有AI模型测试应该通过，成功率100%
```

## 🔧 故障排除

### 问题1：模型下载失败

如果在Docker构建时模型下载失败，可以手动下载：

```bash
# 进入运行中的容器
docker exec -it meeting-python-inference bash

# 手动运行下载脚本
/app/download_models.sh

# 或者逐个下载
python3 << EOF
from huggingface_hub import snapshot_download
snapshot_download(
    repo_id="openai/whisper-tiny",
    local_dir="/models/speech_recognition",
    local_dir_use_symlinks=False
)
EOF
```

### 问题2：GPU不可用

如果没有GPU或GPU驱动问题：

```yaml
# 编辑 docker-compose.yml，注释掉GPU配置
python-inference:
  # deploy:
  #   resources:
  #     reservations:
  #       devices:
  #         - driver: nvidia
  #           count: 1
  #           capabilities: [gpu]
```

然后重新构建和启动。

### 问题3：内存不足

如果系统内存不足，可以：

1. 减少同时加载的模型数量
2. 使用更小的模型
3. 增加swap空间

```bash
# 增加swap空间
sudo fallocate -l 8G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

### 问题4：Python推理调用失败

检查AI服务是否能访问Python推理容器：

```bash
# 从AI服务容器测试
docker exec meeting-ai-service sh -c "docker exec meeting-python-inference echo 'test'"

# 如果失败，可能需要使用网络调用而不是docker exec
```

## 📊 模型列表

| # | 模型类型 | 模型ID | 大小 | 用途 |
|---|---------|--------|------|------|
| 1 | 语音识别 | openai/whisper-tiny | 39MB | 会议语音转文字 |
| 2 | 情绪检测 | j-hartmann/emotion-english-distilroberta-base | 82MB | 检测说话者情绪 |
| 3 | 文本摘要 | sshleifer/distilbart-cnn-6-6 | 306MB | 会议记录摘要 |
| 4 | 音频降噪 | speechbrain/sepformer-wham | ~100MB | 实时音频降噪 |
| 5 | 视频增强 | caidas/swin2SR-classical-sr-x2-64 | ~50MB | 视频质量提升 |
| 6 | 音频伪造检测 | microsoft/wavlm-base-plus | 378MB | 检测AI生成音频 |
| 7 | 人脸伪造检测 | google/vit-base-patch16-224 | 346MB | 检测AI生成人脸 |

**总大小**: ~1.3GB

## 🚀 架构说明

### 当前架构

```
E2E Test
    ↓
AI Service (Go)
    ↓
docker exec → Python Inference Container
                    ↓
              Transformers + PyTorch
                    ↓
              真实AI模型推理
```

### 推理流程

1. E2E测试发送请求到AI服务（通过Nginx网关）
2. AI服务接收请求，调用 `RealInferenceService`
3. `RealInferenceService` 通过 `docker exec` 调用Python推理容器
4. Python容器加载对应的模型并执行推理
5. 推理结果返回给AI服务
6. AI服务返回结果给E2E测试

## ✅ 验证清单

完成部署后，验证以下内容：

- [ ] Python推理容器成功启动
- [ ] 所有7个模型成功下载到 `/models/` 目录
- [ ] AI服务能够成功调用Python推理
- [ ] E2E测试中所有AI模型测试通过
- [ ] 推理响应时间 < 5秒
- [ ] 推理结果格式正确且非空
- [ ] 系统内存占用合理（< 8GB）

## 📝 性能优化建议

### 1. 模型预加载
在Python推理服务启动时预加载所有模型到内存：

```python
# 在 inference_server.py 中添加
if __name__ == "__main__":
    # 预加载所有模型
    for model_type in MODEL_PATHS.keys():
        try:
            load_model(model_type)
            logger.info(f"Preloaded: {model_type}")
        except Exception as e:
            logger.error(f"Failed to preload {model_type}: {e}")
```

### 2. 使用HTTP服务
将Python推理改为HTTP服务，避免每次都启动新进程：

```python
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/inference/<task_type>', methods=['POST'])
def inference(task_type):
    data = request.json
    result = process_inference(task_type, data)
    return jsonify(result)

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8085)
```

### 3. 批处理
支持批量推理以提高吞吐量。

### 4. 模型量化
使用FP16或INT8量化减少内存占用和提高速度。

## 🔗 相关文件

- Python推理服务：`backend/ai-service/Dockerfile.inference`
- 推理脚本：`backend/ai-service/scripts/inference_server.py`
- 模型下载脚本：`backend/ai-service/scripts/download_all_models.sh`
- AI服务更新：`backend/ai-service/services/real_inference_service.go`
- Docker配置：`docker-compose.yml`
- E2E测试：`backend/tests/e2e_integration_test.go`

## ⚠️ 重要提示

1. **不允许任何模拟**：所有推理必须使用真实模型，不允许降级到模拟响应
2. **模型必须下载**：在运行E2E测试前，确保所有模型已成功下载
3. **GPU推荐**：虽然可以使用CPU，但GPU会显著提高推理速度
4. **内存要求**：建议至少8GB RAM + 4GB Swap
5. **磁盘空间**：需要至少2GB空间存储模型

## 📞 下一步

执行上述步骤后，运行E2E测试验证所有功能：

```bash
cd meeting-system/backend/tests
go test -v -run TestE2EIntegration 2>&1 | tee /tmp/e2e_with_real_models.log
```

预期所有AI模型测试成功率达到100%！
