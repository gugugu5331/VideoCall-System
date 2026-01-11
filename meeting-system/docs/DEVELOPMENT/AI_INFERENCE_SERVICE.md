# AI Inference Service

Go 实现的 AI 推理微服务，通过 Triton Inference Server（GPU）提供 ASR、情绪分析与音频伪造检测，暴露 HTTP/gRPC API，并与网关 `/api/v1/ai/*` 对接。

> 基础 compose 默认为了轻量化注释了 AI 相关容器，请使用远程/单机 GPU 方案或自行启用。

## 能力

- **ASR**：`POST /api/v1/ai/asr`，输入音频（base64，`format`=`wav`，`sample_rate`=16000），返回 `text/ confidence`
- **情绪**：`POST /api/v1/ai/emotion`，音频或文本（取决于模型配置），返回主要情绪及分数
- **深度伪造检测**：`POST /api/v1/ai/synthesis`，音频伪造/合成检测
- **批量/预热**：`/api/v1/ai/{batch,setup}`；信息与健康：`/api/v1/ai/{info,health}`、根路径 `/health`

响应格式统一为 `{"code":200,"message":"success","data":{...}}`。

## 架构要点

- **Triton Runtime**：`ai.runtime.triton.endpoint`（默认 `http://triton:8000`）；超时、线程等在配置中调整。
- **模型映射**：`ai.models.*` 指定模型名、输入输出、采样率及 tokenizer/labels 路径，需与 Triton `config.pbtxt` 一致。
- **可观测性**：`/metrics` 暴露 Prometheus 指标；Jaeger tracing、结构化日志输出到 Loki。
- **服务发现**：可注册到 etcd（可选）；Redis 用于缓存与速率限制。

## 配置与文件

- 主配置：`backend/ai-inference-service/config/ai-inference-service.yaml`
- 关键字段：
  - `ai.runtime.triton.endpoint`：Triton HTTP 地址
  - `ai.models.asr|emotion|synthesis`：模型名、输入输出、样本率、路径
  - `ai.http.endpoint`：如需将本服务作为网关转发到远端 AI，可填写此字段（留空表示直连 Triton）
  - 监控、安全、缓存、限流等开关可在对应段落调整
- 环境变量可覆盖上述配置（端口、Redis/Triton 地址等）；生产建议用 `.env` 或 Secret 管理敏感信息

## 部署路径

- **单机 GPU**：`docker compose -f deployment/gpu-ai/docker-compose.gpu-ai.yml up -d --build`（需准备 `/models`）
- **远程/生产**：使用 `docker-compose.remote.yml` 或独立 GPU 节点 + 网关 upstream（见 `docs/DEPLOYMENT/GPU_AI_NODES.md`）
- **本地启用**：在 `docker-compose.yml` 中取消 `triton` 与 `ai-inference-service` 注释，并提供模型仓库

## 验证

```bash
curl http://<host>:8085/health
curl http://<host>:8085/api/v1/ai/info
curl http://<host>:8085/api/v1/ai/health
```

脚本：
- `backend/ai-inference-service/test_ai_service.py`
- `backend/ai-inference-service/scripts/e2e_stream_pcm.sh <host> <grpc_port>`（需暴露 gRPC 端口）

## 开发扩展

1. 新增任务：在 `services/ai_inference_service.go` 添加方法，`handlers/ai_handler.go` 注册路由
2. 配置新模型：在 `ai.models.*` 增加条目并准备对应 Triton 模型与资源文件
3. 补充前/后处理：在 `runtime/` 与 `services/` 中实现转换与结果封装
4. 更新文档与前端调用（如适用）

## 常见问题

- **模型不匹配/报错**：确认模型名、输入输出与 `config.pbtxt` 对齐；检查路径（tokenizer/labels）是否挂载到容器。
- **超时**：调整 `ai.runtime.triton.timeout_ms` 或 `ai.request.timeout`，必要时增加 `setup` 预热。
- **无 GPU**：可使用 CPU 镜像测试，但性能有限；推荐在 GPU 环境部署生产流量。
