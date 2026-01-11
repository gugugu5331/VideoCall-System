# 远程部署指南

适用于通过 SSH 将 `meeting-system` 部署到远程主机（含启用 AI 的 `docker-compose.remote.yml`）。不在仓库记录任何真实主机/密码，请使用环境变量或私有配置。

## 安全提示

- 禁止在仓库/命令中写入真实 IP、端口映射、账号或密码
- 优先使用 SSH 公钥登录；如需密码，使用环境变量传递（`sshpass -e`），避免出现在历史记录
- 生产环境务必更换数据库/MinIO/Kafka/Redis 凭据，设置强 `JWT_SECRET`

## 一键部署（quick-deploy-remote.sh）

脚本路径：`meeting-system/quick-deploy-remote.sh`，使用 `docker-compose.remote.yml`。

必填环境变量：
- `REMOTE_HOST`：SSH 主机
- `REMOTE_DIR`：远程代码目录（包含 `meeting-system/`）

常用可选项：
- `REMOTE_PORT`（默认 22）、`REMOTE_USER`（默认 root）
- `REMOTE_SSH_KEY`（推荐）或 `REMOTE_PASSWORD`（需本地安装 sshpass）
- `REMOTE_GATEWAY_URL`：部署完成后用于健康检查的网关地址，例如 `http://<public-host>:<public-port>`
- `REMOTE_COMPOSE_FILE`：默认 `docker-compose.remote.yml`

示例（公钥登录）：

```bash
export REMOTE_HOST="example.com"
export REMOTE_PORT="22"
export REMOTE_USER="root"
export REMOTE_SSH_KEY="$HOME/.ssh/id_ed25519"
export REMOTE_DIR="/root/VideoCall-System"
export REMOTE_GATEWAY_URL="http://example.com:8800"
./meeting-system/quick-deploy-remote.sh
```

## 远程测试

- **HTTP 入口**：`curl $REMOTE_GATEWAY_URL/health`
- **完整集成测试**：`REMOTE_BASE_URL=$REMOTE_GATEWAY_URL python3 meeting-system/backend/tests/complete_integration_test_remote.py`
- 可选 SSH 清库参数：`REMOTE_SSH_HOST/PORT/USER` + `REMOTE_SSH_KEY` 或 `REMOTE_SSH_PASSWORD`
  
如测试失败，先在远程执行 `docker compose -f docker-compose.remote.yml ps` 与 `logs -f nginx`，再检查防火墙/NAT 映射与证书配置。

## 端口映射建议

内网保持 `8800~8805`，按需在云防火墙或 NAT 暴露：

| 内网端口 | 建议用途 |
| --- | --- |
| 8800 | 网关 |
| 8801~8805 | Prometheus / Alertmanager / Jaeger / Grafana / Loki（如需开放） |
| 9000/9001 | MinIO API/Console（如需开放） |
| 8000/8001 | Triton HTTP/gRPC（可选，仅 AI 节点） |

## 常用远程命令

```bash
cd <REMOTE_DIR>/meeting-system
docker compose -f docker-compose.remote.yml up -d          # 启动
docker compose -f docker-compose.remote.yml down           # 停止
docker compose -f docker-compose.remote.yml ps             # 状态
docker compose -f docker-compose.remote.yml logs -f nginx  # 网关日志
```

## 故障排查

1. 网关不可达：检查远程防火墙/端口映射，确认 Nginx 容器健康
2. AI 报错：确认 Triton 模型仓库挂载正确、`ai-inference-service` 可访问 Triton
3. 存储/上传异常：核对 MinIO 凭据、桶名称与 `media-service` 配置

更多场景（多 GPU、K8s）请参考 `GPU_AI_NODES.md` 与 `deployment/k8s/README.md`。
