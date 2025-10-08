#!/bin/bash

###############################################################################
# 快速远程部署脚本
# 简化版本，直接执行关键步骤
###############################################################################

set -e

# 远程服务器配置
REMOTE_HOST="js1.blockelite.cn"
REMOTE_PORT="22124"
REMOTE_USER="root"
REMOTE_PASSWORD="beip3ius"
REMOTE_DIR="/root/meeting-system-server"

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
    sshpass -p "${REMOTE_PASSWORD}" ssh -p ${REMOTE_PORT} -o StrictHostKeyChecking=no ${REMOTE_USER}@${REMOTE_HOST} "$@"
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
ssh_exec "cd ${REMOTE_DIR}/meeting-system && docker-compose -f docker-compose.remote.yml down" || true

log_info "启动服务（这可能需要几分钟）..."
ssh_exec "cd ${REMOTE_DIR}/meeting-system && docker-compose -f docker-compose.remote.yml up -d" 2>&1 | tail -20

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
ssh_exec "docker logs meeting-edge-model-infra --tail 10 2>&1" || true
echo ""

# 7. 测试服务可访问性
log_info "7. 测试服务可访问性..."

log_info "测试 Nginx (22176)..."
if curl -f -s -o /dev/null -w "%{http_code}" "http://${REMOTE_HOST}:22176/health" | grep -q "200"; then
    log_success "Nginx 可访问"
else
    log_warning "Nginx 不可访问"
fi

log_info "测试 Jaeger (22177)..."
if curl -f -s -o /dev/null -w "%{http_code}" "http://${REMOTE_HOST}:22177/" | grep -q "200"; then
    log_success "Jaeger 可访问"
else
    log_warning "Jaeger 不可访问"
fi

log_info "测试 Prometheus (22178)..."
if curl -f -s -o /dev/null -w "%{http_code}" "http://${REMOTE_HOST}:22178/" | grep -q "200"; then
    log_success "Prometheus 可访问"
else
    log_warning "Prometheus 不可访问"
fi

log_info "测试 Grafana (22180)..."
if curl -f -s -o /dev/null -w "%{http_code}" "http://${REMOTE_HOST}:22180/" | grep -q "200\|302"; then
    log_success "Grafana 可访问"
else
    log_warning "Grafana 不可访问"
fi

echo ""

echo "=========================================="
echo "  部署完成"
echo "=========================================="
echo ""
echo "访问地址："
echo "  - Nginx:      http://${REMOTE_HOST}:22176"
echo "  - Jaeger:     http://${REMOTE_HOST}:22177"
echo "  - Prometheus: http://${REMOTE_HOST}:22178"
echo "  - Grafana:    http://${REMOTE_HOST}:22180"
echo ""
echo "下一步："
echo "  1. 运行集成测试: ./run-remote-integration-test.sh"
echo "  2. 验证 AI 服务: ./verify-ai-service-remote.sh"
echo "  3. 查看服务日志: ssh -p ${REMOTE_PORT} ${REMOTE_USER}@${REMOTE_HOST} 'docker logs meeting-[service-name]'"
echo ""

