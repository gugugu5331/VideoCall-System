# GPU 服务器部署（AI 推理节点）

本目录用于在 **单台 GPU 服务器** 上部署一套 AI 推理节点（Triton Inference Server + TensorRT）：

- `triton`（Triton Inference Server，TensorRT GPU 推理）
- `ai-inference-service`（Go 服务，提供 `/api/v1/ai/*` 与 gRPC 流）

---

## 端口约定（建议）

建议在 GPU 服务器上预留一组内网端口（例如 `8800~8805`），再由你的 NAT/安全组映射到公网端口。

本部署默认使用：

- `AI_HTTP_PORT`（默认 `8800`）：映射到容器 `ai-inference-service:8085`
- `AI_GRPC_PORT`（默认 `9800`）：映射到容器 `ai-inference-service:9085`
- `TRITON_HTTP_PORT`（默认 `8000`）：映射到容器 `triton:8000`
- `TRITON_GRPC_PORT`（默认 `8001`）：映射到容器 `triton:8001`
- `TRITON_METRICS_PORT`（默认 `8002`）：映射到容器 `triton:8002`

---

## 在单台 GPU 服务器上启动

前置条件：

- 已安装 `docker` + `docker compose` 插件
- 服务器上已有模型目录（默认挂载 `MODEL_DIR=/models` 到容器 `/models`）

启动：

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
```

---

## 多台 GPU 服务器协作方式（推荐）

1. 在 GPU-1、GPU-2 各自运行一套本目录的 `docker-compose.gpu-ai.yml`
2. 在主站 Nginx（或任何网关）里将 `/api/v1/ai/*` 代理到两个节点的 `AI_HTTP_PORT`
3. gRPC 流式建议走 L4 负载均衡或 gRPC-aware 代理

本项目网关已在 `meeting-system/nginx/nginx.conf` 中把 `/api/v1/ai/*` 代理到 upstream `ai_inference_service`，并通过 `include /etc/nginx/conf.d/ai_inference_service.servers*.conf` 读取上游列表。

推荐做法：在主站生成一个本地私有文件（已加入 `.gitignore`）：

```bash
AI_INFERENCE_UPSTREAMS="<gpu-node-1-host>:<public-port> <gpu-node-2-host>:<public-port>" \
bash meeting-system/nginx/scripts/gen_ai_inference_service_servers_conf.sh
```

生成后会写入：`meeting-system/nginx/conf.d/ai_inference_service.servers.local.conf`，然后重启/重载网关 Nginx 即可生效。

---

## SSH 自动化说明

为避免在命令/日志中泄露敏感信息，本目录的自动化脚本默认 **仅支持 SSH 公钥登录**（`authorized_keys`），不在仓库内保存任何远程主机/账号/密码。

已提供基于公钥登录的一键部署脚本：`meeting-system/deployment/gpu-ai/deploy_gpu_ai_nodes.sh`。

示例：

```bash
cd /root/VideoCall-System/meeting-system/deployment/gpu-ai
GPU_AI_NODES_FILE=./nodes.example.txt \
MODEL_DIR=/models AI_HTTP_PORT=8800 AI_GRPC_PORT=9800 TRITON_HTTP_PORT=8000 TRITON_GRPC_PORT=8001 TRITON_METRICS_PORT=8002 \
./deploy_gpu_ai_nodes.sh
```
