# 会议系统远程部署指南

## 安全说明（重要）

为避免泄露敏感信息，本仓库不保存任何以下内容：
- 真实远程主机名/IP、SSH 端口、账号/密码
- NAT/端口映射的具体公网入口

请在本地通过环境变量（或本地 `.env` 文件）提供部署参数；生产环境推荐使用 **SSH Key**。

---

## 需要的变量

### 1) 远程部署（`meeting-system/quick-deploy-remote.sh`）

必填：
- `REMOTE_HOST`：远程 SSH 主机
- `REMOTE_DIR`：远程部署目录（包含 `meeting-system/`）

可选：
- `REMOTE_PORT`：SSH 端口（默认 `22`）
- `REMOTE_USER`：SSH 用户（默认 `root`）
- `REMOTE_SSH_KEY`：SSH 私钥路径（推荐）
- `REMOTE_PASSWORD`：SSH 密码（可选；需要 `sshpass`，脚本使用 `sshpass -e` 避免把密码写在命令行）
- `REMOTE_GATEWAY_URL`：部署完成后的网关地址（用于健康检查），如 `http://<host>:<port>`
- `REMOTE_COMPOSE_FILE`：compose 文件名（默认 `docker-compose.remote.yml`）

示例（推荐 Key）：
```bash
export REMOTE_HOST="<remote-host>"
export REMOTE_PORT="22"
export REMOTE_USER="root"
export REMOTE_SSH_KEY="$HOME/.ssh/id_ed25519"
export REMOTE_DIR="/root/meeting-system-server"
export REMOTE_GATEWAY_URL="http://<public-host>:<public-port>"
./meeting-system/quick-deploy-remote.sh
```

---

### 2) 远程集成测试（`meeting-system/backend/tests/complete_integration_test_remote.py`）

必填：
- `REMOTE_BASE_URL`：网关基础地址，例如 `http://<public-host>:<public-port>`

可选（如果需要通过 SSH 清库）：
- `REMOTE_SSH_HOST` / `REMOTE_SSH_PORT` / `REMOTE_SSH_USER`
- `REMOTE_SSH_KEY`（推荐）或 `REMOTE_SSH_PASSWORD`（可选，依赖 `sshpass`）

示例：
```bash
export REMOTE_BASE_URL="http://<public-host>:<public-port>"
python3 meeting-system/backend/tests/complete_integration_test_remote.py
```

---

## 端口映射建议（示例）

如果你的环境提供 NAT（外网端口 → 内网端口），建议在内网统一使用 `8800~8805`：

| 内网端口 | 建议用途 |
|---------|----------|
| 8800 | Nginx 网关 / Web 入口 |
| 8801 | 监控/追踪 UI（可选） |
| 8802 | 监控（Prometheus 等，可选） |
| 8803 | 告警（可选） |
| 8804 | 仪表盘（可选） |
| 8805 | 日志聚合（可选） |

具体映射由你的云厂商/运维平台配置；不要将真实映射写入仓库。

---

## 运维常用命令（远程执行）

```bash
cd <REMOTE_DIR>/meeting-system

# 启动
docker compose -f docker-compose.remote.yml up -d

# 停止
docker compose -f docker-compose.remote.yml down

# 查看状态
docker compose -f docker-compose.remote.yml ps

# 查看日志
docker compose -f docker-compose.remote.yml logs -f nginx
```

---

## 故障排查（通用）

1. **网关不可访问**：先确认远程 Nginx/网关容器健康，再检查端口映射与防火墙。
2. **AI 接口报错**：确认 AI 推理节点与 unit-manager 进程/容器正常运行，并检查 `/tmp/llm`（IPC 模式）或相关端口（TCP 模式）。

更多问题排查建议请结合你的实际部署拓扑（主机名/端口映射/网关）检查容器日志与网络连通性；不要将真实信息写入仓库。
