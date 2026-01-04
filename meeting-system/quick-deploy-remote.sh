#!/bin/bash

###############################################################################
# 快速远程部署脚本
# 简化版本，直接执行关键步骤
#
# 注意：为避免泄露敏感信息，本脚本不再内置远程主机/端口/密码。
# 请通过环境变量传入（推荐使用 SSH Key）。
###############################################################################

set -e

# ---------------------------------------------------------------------------
# Remote config (required via env)
# ---------------------------------------------------------------------------
# Required:
#   REMOTE_HOST         远程 SSH 主机
#   REMOTE_DIR          远程部署目录（包含 meeting-system 目录）
#
# Optional:
#   REMOTE_PORT         SSH 端口（默认 22）
#   REMOTE_USER         SSH 用户（默认 root）
#   REMOTE_SSH_KEY      SSH 私钥路径（推荐）
#   REMOTE_PASSWORD     SSH 密码（可选；需要 sshpass；不会写入命令行参数）
#   REMOTE_GATEWAY_URL  部署完成后的网关地址（用于健康检查），如 http://<host>:<port>
#   REMOTE_COMPOSE_FILE docker compose 文件（默认 docker-compose.remote.yml）
REMOTE_HOST="${REMOTE_HOST:?Please set REMOTE_HOST}"
REMOTE_DIR="${REMOTE_DIR:?Please set REMOTE_DIR (e.g. /root/meeting-system-server)}"

REMOTE_PORT="${REMOTE_PORT:-22}"
REMOTE_USER="${REMOTE_USER:-root}"
REMOTE_SSH_KEY="${REMOTE_SSH_KEY:-}"
REMOTE_PASSWORD="${REMOTE_PASSWORD:-}"
REMOTE_GATEWAY_URL="${REMOTE_GATEWAY_URL:-}"
REMOTE_COMPOSE_FILE="${REMOTE_COMPOSE_FILE:-docker-compose.remote.yml}"

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# SSH 命令封装
ssh_exec() {
    local ssh_opts=(-p "${REMOTE_PORT}" -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null)
    if [[ -n "${REMOTE_SSH_KEY}" ]]; then
        ssh_opts+=(-i "${REMOTE_SSH_KEY}")
    fi

    if [[ -n "${REMOTE_PASSWORD}" ]]; then
        if ! command -v sshpass >/dev/null 2>&1; then
            log_error "REMOTE_PASSWORD 已设置，但未找到 sshpass；请安装 sshpass 或改用 SSH key"
            exit 1
        fi
        SSHPASS="${REMOTE_PASSWORD}" sshpass -e ssh "${ssh_opts[@]}" "${REMOTE_USER}@${REMOTE_HOST}" "$@"
        return
    fi

    ssh "${ssh_opts[@]}" "${REMOTE_USER}@${REMOTE_HOST}" "$@"
}

echo "=========================================="
echo "  快速远程部署"
echo "=========================================="
echo ""

# 1. 检查远程服务器连接
log_info "1. 测试远程服务器连接..."
if ssh_exec "echo 'Connected'"; then
    log_success "SSH 连接成功"
else
    log_error "SSH 连接失败"
    exit 1
fi
echo ""

# 2. 检查 Docker 服务状态
log_info "2. 检查 Docker 服务状态..."
ssh_exec "docker ps --filter 'name=meeting-' --format 'table {{.Names}}\t{{.Status}}' | head -20"
echo ""

# 3. 启动/重启服务
log_info "3. 启动/重启 Docker 服务..."
log_info "停止现有服务..."
ssh_exec "cd ${REMOTE_DIR}/meeting-system && (docker compose -f ${REMOTE_COMPOSE_FILE} down || docker-compose -f ${REMOTE_COMPOSE_FILE} down)" || true

log_info "启动服务（这可能需要几分钟）..."
ssh_exec "cd ${REMOTE_DIR}/meeting-system && (docker compose -f ${REMOTE_COMPOSE_FILE} up -d || docker-compose -f ${REMOTE_COMPOSE_FILE} up -d)" 2>&1 | tail -20

log_success "服务启动命令已执行"
echo ""

# 4. 等待服务启动
log_info "4. 等待服务启动（60秒）..."
sleep 60
echo ""

# 5. 检查服务状态
log_info "5. 检查服务状态..."
ssh_exec "docker ps --filter 'name=meeting-' --format 'table {{.Names}}\t{{.Status}}'"
echo ""

# 6. 检查关键服务日志
log_info "6. 检查关键服务日志..."

log_info "Nginx 日志:"
ssh_exec "docker logs meeting-nginx --tail 10 2>&1" || true
echo ""

log_info "AI Service 日志:"
ssh_exec "docker logs meeting-ai-service --tail 10 2>&1" || true
echo ""

log_info "Edge Model Infra 日志:"
echo ""

# 7. 测试服务可访问性
log_info "7. 测试服务可访问性..."

if [[ -n "${REMOTE_GATEWAY_URL}" ]]; then
    log_info "测试网关健康检查: ${REMOTE_GATEWAY_URL}/health"
    if curl -f -s -o /dev/null -w "%{http_code}" "${REMOTE_GATEWAY_URL%/}/health" | grep -q "200"; then
        log_success "网关可访问"
    else
        log_warning "网关不可访问"
    fi
else
    log_warning "未设置 REMOTE_GATEWAY_URL，跳过外网可访问性检查"
fi

echo ""

echo "=========================================="
echo "  部署完成"
echo "=========================================="
echo ""
echo "访问地址："
if [[ -n "${REMOTE_GATEWAY_URL}" ]]; then
    echo "  - Gateway: ${REMOTE_GATEWAY_URL%/}"
else
    echo "  - Gateway: (unset) 请设置 REMOTE_GATEWAY_URL"
fi
echo ""
echo "下一步："
echo "  1. 运行集成测试: python3 backend/tests/complete_integration_test_remote.py"
echo "  2. 查看服务日志: ssh -p ${REMOTE_PORT} ${REMOTE_USER}@${REMOTE_HOST} 'docker logs meeting-[service-name]'"
echo ""
