# GPU AI 推理节点部署（多机）

说明如何在多台 GPU 服务器上部署 Triton + `ai-inference-service`，并通过主站 Nginx 做负载均衡。默认仓库的基础 compose 未启用 AI，本指南专用于需要 GPU 推理的场景。

## 架构

- 每台 GPU 服务器运行：
  - `triton`（Triton Inference Server，GPU 推理）
  - `ai-inference-service`（HTTP/gRPC `/api/v1/ai/*`，连接本地 Triton）
- 模型仓库挂载到容器 `/models`
- 主站 Nginx 将 `/api/v1/ai/*` 代理到各 GPU 节点上游

## 单台 GPU 节点

在 GPU 服务器执行：

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

验证：

```bash
curl -s http://localhost:${AI_HTTP_PORT}/health
curl -s http://localhost:${AI_HTTP_PORT}/api/v1/ai/info
curl -s http://localhost:${AI_HTTP_PORT}/api/v1/ai/health
curl -s http://localhost:${TRITON_HTTP_PORT}/v2/health/ready
```

## 多节点上游

在主站生成本地 upstream（文件已在 `.gitignore` 中忽略）：

```bash
AI_INFERENCE_UPSTREAMS="<gpu-1-host>:<public-port> <gpu-2-host>:<public-port>" \
bash meeting-system/nginx/scripts/gen_ai_inference_service_servers_conf.sh
```

生成的 `nginx/conf.d/ai_inference_service.servers.local.conf` 会被网关加载，重启/重载 Nginx 生效。

> gRPC 流式建议走 L4 或支持 gRPC 的代理；HTTP 路径沿用 Nginx 即可。

## SSH 自动化

`deployment/gpu-ai/deploy_gpu_ai_nodes.sh` 支持基于公钥的一键部署，需先在各 GPU 服务器配置 `authorized_keys`。

```bash
cd meeting-system/deployment/gpu-ai
GPU_AI_NODES_FILE=./nodes.example.txt \
MODEL_DIR=/models AI_HTTP_PORT=8800 AI_GRPC_PORT=9800 \
TRITON_HTTP_PORT=8000 TRITON_GRPC_PORT=8001 TRITON_METRICS_PORT=8002 \
./deploy_gpu_ai_nodes.sh
```

## 常见问题

1) **Triton 不可达**：检查 `ai.runtime.triton.endpoint` 是否指向正确地址；`curl http://<triton-host>:8000/v2/health/ready`。

2) **GPU 不可见**：确认宿主机 `nvidia-smi` 正常；`docker run --rm --gpus all nvidia/cuda:12.2.2-cudnn8-runtime-ubuntu22.04 nvidia-smi` 测试容器访问。

3) **模型加载失败**：确保 `/models` 挂载到容器并包含与 `ai-inference-service` 配置匹配的模型名/输入输出。
 
4) **主站未均衡到节点**：检查生成的 `nginx/conf.d/ai_inference_service.servers.local.conf` 是否包含正确公网/内网地址，重载 Nginx 后再调用 `/api/v1/ai/health`。
