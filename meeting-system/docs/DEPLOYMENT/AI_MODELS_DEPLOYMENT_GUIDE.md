# AI 模型与推理节点部署指南

本指南聚焦于为 `ai-inference-service` 准备 Triton 模型仓库并启动推理节点。基础版 `docker-compose.yml` 默认不启用 AI；请使用远程/单机 GPU 方案或手动启用。

## 目标形态

- **Triton Inference Server**：加载模型仓库（挂载到容器 `/models`），开放 HTTP/gRPC（默认 8000/8001）
- **ai-inference-service**：HTTP/gRPC `/api/v1/ai/*`，直连本地 Triton（端口 8085/9085）
- **Nginx 网关**：将 `/api/v1/ai/*` 代理到上述服务，可按需配置多上游

## 模型准备

1. 按 `backend/ai-inference-service/config/ai-inference-service.yaml` 中的模型名/输入输出准备 Triton 模型仓库，常用目录示例：
   ```
   /models
   ├── whisper-encoder/
   │   ├── 1/model.plan           # 编码器（示例）
   │   └── config.pbtxt
   ├── whisper-decoder/
   │   ├── 1/model.plan
   │   └── config.pbtxt
   ├── emotion/
   │   ├── 1/model.plan
   │   └── config.pbtxt
   └── synthesis/
       ├── 1/model.plan
       └── config.pbtxt
   ```
   实际格式（ONNX/Plan/PyTorch）依赖你的训练/转换结果；确保输入输出名称与配置一致。

2. 需要的资源路径（tokenizer、labels 等）在配置文件中指定，如：
   - Whisper：`tokenizer_path`、`special_tokens_path`、`config_path`
   - Emotion：`labels_path`
   根据模型放置到相同卷内（例如 `/models/whisper/whisper_vocab.json`）。

> 仓库中的 `download_models.py` 等脚本用于拉取原始 HuggingFace 模型，**不会自动生成 Triton 配置**；请按自身模型完成转换与 `config.pbtxt` 配置。

## 启动方式

- **单机 GPU（推荐）**  
  ```bash
  cd meeting-system
  MODEL_DIR=/models docker compose -f deployment/gpu-ai/docker-compose.gpu-ai.yml up -d --build
  ```
- **远程/生产**  
  使用 `docker-compose.remote.yml` 或在 GPU 节点独立运行 Triton + `ai-inference-service`，再在主站配置 Nginx 上游。
- **本地测试（轻量）**  
  如需在基础 compose 启用，可自行取消注释 `docker-compose.yml` 中的 `triton` 与 `ai-inference-service`，但镜像较大且需可用模型仓库。

调优与配置同步：
- 调整 `backend/ai-inference-service/config/ai-inference-service.yaml` 中的 `model_name`、输入输出、超时与并发，确保与实际模型一致。
- 如果改用远端 HTTP AI 网关，设置 `ai.http.endpoint`，并在 Nginx upstream 添加对应上游。

## 验证

```bash
curl http://<ai-host>:8085/health
curl http://<ai-host>:8085/api/v1/ai/info
curl http://<triton-host>:8000/v2/health/ready
```

可选：`backend/ai-inference-service/test_ai_service.py`、`scripts/e2e_stream_pcm.sh` 进行端到端校验。

## 常见问题

- **模型名称/输入输出不匹配**：对照 `ai-inference-service.yaml` 中的 `model_name`、`input_name`、`output_names` 修改 Triton `config.pbtxt`。
- **性能/超时**：调整 `ai.runtime.triton.timeout_ms` 或 `ai.request.timeout`，必要时预热（调用 `/api/v1/ai/setup`）。
- **GPU 不可用**：确认宿主机 GPU 与驱动正常，`docker run --rm --gpus all nvidia/cuda:12.2.2-cudnn8-runtime-ubuntu22.04 nvidia-smi` 通过。
- **资源不足**：按模型大小预估显存/内存，必要时减少并发或切换精度（FP16/INT8），并监控 Triton 指标。
