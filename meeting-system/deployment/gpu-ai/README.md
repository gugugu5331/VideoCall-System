# GPU 服务器部署（AI 推理节点）

在单台 GPU 服务器上启动 Triton + `ai-inference-service`。适用于为主站提供 AI 上游；基础 compose 未启用 AI，需要本方案或远程部署。

## 端口（默认）

- `AI_HTTP_PORT` 8800 → `ai-inference-service:8085`
- `AI_GRPC_PORT` 9800 → `ai-inference-service:9085`
- `TRITON_HTTP_PORT` 8000 → `triton:8000`
- `TRITON_GRPC_PORT` 8001 → `triton:8001`
- `TRITON_METRICS_PORT` 8002 → `triton:8002`

## 启动步骤

前置：服务器已安装 Docker + Compose，模型仓库挂载目录 `MODEL_DIR` 已准备好（与 `ai-inference-service` 配置匹配）。

```bash
cd /root/VideoCall-System/meeting-system

export MODEL_DIR=/models
export AI_HTTP_PORT=8800
export AI_GRPC_PORT=9800
export TRITON_HTTP_PORT=8000
export TRITON_GRPC_PORT=8001
export TRITON_METRICS_PORT=8002

docker compose -f deployment/gpu-ai/docker-compose.gpu-ai.yml up -d --build
```

自检：
```bash
curl -s http://localhost:${AI_HTTP_PORT}/health
curl -s http://localhost:${AI_HTTP_PORT}/api/v1/ai/info
curl -s http://localhost:${AI_HTTP_PORT}/api/v1/ai/health
docker compose -f deployment/gpu-ai/docker-compose.gpu-ai.yml logs -f triton ai-inference-service
```

## 接入主站

在主站 Nginx 配置 `/api/v1/ai/*` 上游：

```bash
AI_INFERENCE_UPSTREAMS="<gpu-host>:<public-port>" \
bash meeting-system/nginx/scripts/gen_ai_inference_service_servers_conf.sh
```

生成的 `nginx/conf.d/ai_inference_service.servers.local.conf` 会被自动包含，重启/重载 Nginx 生效。

## SSH 自动化

`deploy_gpu_ai_nodes.sh` 支持基于公钥的批量部署（需提前配置 `authorized_keys`）：

```bash
cd meeting-system/deployment/gpu-ai
GPU_AI_NODES_FILE=./nodes.example.txt \
MODEL_DIR=/models AI_HTTP_PORT=8800 AI_GRPC_PORT=9800 \
TRITON_HTTP_PORT=8000 TRITON_GRPC_PORT=8001 TRITON_METRICS_PORT=8002 \
./deploy_gpu_ai_nodes.sh
```

## 常见问题

- **Triton 无法加载模型**：确认 `/models` 挂载正确且 `config.pbtxt` 与 `ai-inference-service` 配置一致。
- **GPU 不可见**：`nvidia-smi` 与 `docker run --rm --gpus all nvidia/cuda:12.2.2-cudnn8-runtime-ubuntu22.04 nvidia-smi` 检查。
- **上游不可达**：确认 AI 节点公网/NAT 端口映射，或在主站使用内网地址。
