# GPU AI 推理节点部署（多机）

本指南用于在 **多台 GPU 服务器** 上部署 AI 推理能力（Triton Inference Server + TensorRT），并通过上层网关做负载均衡。

> 安全提示：不要把真实服务器地址/账号/密码写入仓库；请在本地通过环境变量或私有配置文件管理。

## 架构说明

- 每台 GPU 服务器独立运行一套 AI 推理节点：
  - `triton`（Triton Inference Server，TensorRT GPU 推理）
  - `ai-inference-service`（Go 服务，提供 HTTP `/api/v1/ai/*` 和 gRPC 流）
- 模型文件放在每台 GPU 节点的 `/models`（Triton model repository）。
- `ai-inference-service` 通过 `ai.runtime.triton.endpoint` 连接 Triton（`ai.http.endpoint` 留空表示使用本地 Triton，不再依赖 ZMQ/Edge-LLM）。
- 多机扩展建议：每台机器各自部署一套，然后在主站（Nginx/网关）做 **多上游负载均衡**。

## 单台 GPU 服务器启动

在 GPU 服务器上（示例端口可按你的 NAT 映射调整）：

```bash
cd <repo-root>/meeting-system

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
curl -s "http://localhost:${AI_HTTP_PORT}/health"
curl -s "http://localhost:${AI_HTTP_PORT}/api/v1/ai/info"
curl -s "http://localhost:${AI_HTTP_PORT}/api/v1/ai/health"
```

## 多台 GPU 服务器一键部署（SSH 公钥）

管理机上运行（需要目标服务器已配置 `authorized_keys`）：

```bash
cd <repo-root>/meeting-system/deployment/gpu-ai

export GPU_AI_NODES="<gpu1-host>:<ssh-port> <gpu2-host>:<ssh-port>"
export REMOTE_DIR="/root/VideoCall-System"
export MODEL_DIR="/models"
export AI_HTTP_PORT=8800
export AI_GRPC_PORT=9800
export TRITON_HTTP_PORT=8000
export TRITON_GRPC_PORT=8001
export TRITON_METRICS_PORT=8002

./deploy_gpu_ai_nodes.sh
```

## 主站网关接入（示例）

在主站 Nginx 将 `/api/v1/ai/*` 转发到多台 GPU 节点（端口填公网可达地址）。

本项目网关已在 `meeting-system/nginx/nginx.conf` 中把 `/api/v1/ai/*` 代理到 upstream `ai_inference_service`，并通过 `include /etc/nginx/conf.d/ai_inference_service.servers*.conf` 读取上游列表。

推荐做法：生成一个本地私有文件（已加入 `.gitignore`）：

```bash
AI_INFERENCE_UPSTREAMS="<gpu-node-1-host>:<public-port> <gpu-node-2-host>:<public-port>" \
bash meeting-system/nginx/scripts/gen_ai_inference_service_servers_conf.sh
```

生成后会写入：`meeting-system/nginx/conf.d/ai_inference_service.servers.local.conf`，然后重启/重载网关 Nginx 即可生效。

> gRPC 流式建议走 L4 负载均衡或 gRPC-aware 代理；HTTP 入口可继续走 Nginx。

## 常见问题排查

### 1) Triton 不可达

现象：`ai-inference-service` 启动后模型加载失败，日志提示连接 Triton 失败。

处理：
- 检查 `ai.runtime.triton.endpoint` 是否指向正确的 `triton` 地址。
- 直接访问健康检查：`curl http://<triton-host>:8000/v2/health/ready`。

### 2) GPU 不可见

检查 Docker 是否启用 GPU：

```bash
nvidia-smi
docker run --rm --gpus all nvidia/cuda:12.2.2-cudnn8-runtime-ubuntu22.04 nvidia-smi
```

### 3) 模型加载失败

- 检查 `/models` 是否挂载正确（容器内路径）。
- 检查 Triton 模型仓库结构与 `config.pbtxt` 是否完整。
- 检查 `backend/ai-inference-service/config/ai-inference-service.yaml` 中模型名/输入输出名是否匹配。
