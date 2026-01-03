# GPU AI 推理节点部署（多机）

本指南用于在 **多台 GPU 服务器** 上部署 AI 推理能力，并通过上层网关做负载均衡。

> 安全提示：不要把真实服务器地址/账号/密码写入仓库；请在本地通过环境变量或私有配置文件管理。

## 架构说明

- 每台 GPU 服务器独立运行一套 AI 推理节点：
  - `ai-inference-service`（Go HTTP API，提供 `/api/v1/ai/*`）
  - `edge-unit-manager`（Edge-LLM-Infra unit-manager）
  - `edge-llm-node`（Edge-LLM-Infra 推理节点 `llm`）
- 由于 Edge-LLM-Infra 内部依赖 IPC（`/tmp/rpc.*` + `/tmp/llm/*`），`unit-manager` 与 `llm` **必须在同一台机器**。
- 多机扩展建议：每台机器各自部署一套，然后在主站（Nginx/网关）做 **多上游负载均衡**。

## 单台 GPU 服务器启动

在 GPU 服务器上（示例端口可按你的 NAT 映射调整）：

```bash
cd <repo-root>/meeting-system

export MODEL_DIR=/models
export AI_HTTP_PORT=8800
export UNIT_MANAGER_PORT=8801  # 可选：不建议暴露公网

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
export UNIT_MANAGER_PORT=8801

./deploy_gpu_ai_nodes.sh
```

## 主站网关接入（示例）

在主站 Nginx 将 `/api/v1/ai/*` 转发到两台 GPU 节点（端口填你的公网可达地址）。

本项目网关已在 `meeting-system/nginx/nginx.conf` 中把 `/api/v1/ai/*` 代理到 upstream `ai_inference_service`，并通过 `include /etc/nginx/conf.d/ai_inference_service.servers*.conf` 读取上游列表。

推荐做法：生成一个本地私有文件（已加入 `.gitignore`）：

```bash
AI_INFERENCE_UPSTREAMS="<gpu-node-1-host>:<public-port> <gpu-node-2-host>:<public-port>" \
bash meeting-system/nginx/scripts/gen_ai_inference_service_servers_conf.sh
```

生成后会写入：`meeting-system/nginx/conf.d/ai_inference_service.servers.local.conf`，然后重启/重载网关 Nginx 即可生效。

## 常见问题排查

### 1) 大音频请求报 `json format error`

现象：
- `/api/v1/ai/asr`、`/api/v1/ai/synthesis` 在较大音频（例如几百 KB 以上）时，偶发或必现返回 `edge-llm error (code -2): json format error`

原因：
- Edge-LLM-Infra 的 `unit-manager` TCP 接入侧如果未做 **按行（`\n`）切分**，会出现粘包/拆包；当一次 `recv()` 得到半条或多条 JSON 时，解析会直接失败。

处理：
- **推荐（根治）**：升级 GPU 节点上的 `unit-manager`（本仓库已修复按行切分逻辑，见 `meeting-system/Edge-LLM-Infra-master/unit-manager/src/zmq_bus.cpp`），然后重新构建并重启 GPU 节点容器。
- **临时（规避）**：在 GPU 节点 `ai-inference-service` 配置中降低单次发送量并加入 chunk 间隔，例如：
  - `ai.request.max_single_delta_len: 2048`
  - `ai.request.audio_stream_chunk_size: 512`
  - `ai.request.audio_stream_chunk_delay_ms: 10`

> 如果 `unit-manager` 已升级支持按行切分，可把 `audio_stream_chunk_delay_ms` 设为 `0` 并适当增大 `audio_stream_chunk_size` 以减少延迟。
