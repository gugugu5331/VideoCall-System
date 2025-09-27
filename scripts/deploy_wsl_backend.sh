#!/bin/bash

# VideoCall System - WSL后端部署脚本
# 在WSL中部署后端服务，支持Windows前端通信

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# 检查函数
check_wsl() {
    log_step "检查WSL环境..."
    
    if ! grep -q Microsoft /proc/version 2>/dev/null && ! grep -q WSL /proc/version 2>/dev/null; then
        log_error "当前不在WSL环境中"
        exit 1
    fi
    
    log_success "WSL环境检查通过"
}

check_docker() {
    log_step "检查Docker..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        log_error "Docker未运行，请启动Docker Desktop"
        exit 1
    fi
    
    log_success "Docker检查通过"
}

check_docker_compose() {
    log_step "检查Docker Compose..."
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装"
        exit 1
    fi
    
    log_success "Docker Compose检查通过"
}

# 环境准备
prepare_environment() {
    log_step "准备部署环境..."
    
    # 创建必要目录
    mkdir -p /tmp/llm
    mkdir -p /tmp/detection_uploads
    mkdir -p storage/detection
    mkdir -p storage/media
    mkdir -p logs
    
    # 设置权限
    chmod 755 /tmp/llm
    chmod 755 /tmp/detection_uploads
    chmod 755 storage/detection
    chmod 755 storage/media
    
    log_success "环境准备完成"
}

# 构建Edge-Model-Infra
build_edge_infra() {
    log_step "构建Edge-Model-Infra..."
    
    if [ ! -d "Edge-Model-Infra" ]; then
        log_error "Edge-Model-Infra目录不存在"
        exit 1
    fi
    
    cd Edge-Model-Infra
    
    # 构建AI检测节点
    if [ -d "node/ai-detection" ]; then
        log_info "构建AI检测节点..."
        cd node/ai-detection
        
        if [ ! -d "build" ]; then
            mkdir build
        fi
        
        cd build
        cmake .. -DCMAKE_BUILD_TYPE=Release
        make -j$(nproc)
        cd ../../..
        
        log_success "AI检测节点构建完成"
    fi
    
    # 构建Unit Manager
    if [ -d "unit-manager" ]; then
        log_info "构建Unit Manager..."
        cd unit-manager
        
        if [ ! -d "build" ]; then
            mkdir build
        fi
        
        cd build
        cmake .. -DCMAKE_BUILD_TYPE=Release
        make -j$(nproc)
        cd ../..
        
        log_success "Unit Manager构建完成"
    fi
    
    cd ..
}

# 构建后端服务
build_backend_services() {
    log_step "构建后端服务..."
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        log_error "Go未安装"
        exit 1
    fi
    
    # 构建各个服务
    services=("user" "meeting" "signaling" "media" "notification" "record" "smart-editing" "gateway")
    
    for service in "${services[@]}"; do
        service_path="src/backend/services/$service"
        if [ -d "$service_path" ]; then
            log_info "构建 $service 服务..."
            cd "$service_path"
            go mod tidy
            go build -o "../../../build-linux/${service}-service" .
            cd - > /dev/null
            log_success "$service 服务构建完成"
        else
            log_warn "$service 服务目录不存在，跳过"
        fi
    done
}

# 部署服务
deploy_services() {
    log_step "部署服务..."
    
    # 停止现有服务
    log_info "停止现有服务..."
    docker-compose -f deployment/docker-compose.wsl.yml down
    
    # 清理旧容器和镜像
    log_info "清理旧容器..."
    docker system prune -f
    
    # 启动服务
    log_info "启动服务..."
    docker-compose -f deployment/docker-compose.wsl.yml up --build -d
    
    log_success "服务部署完成"
}

# 等待服务启动
wait_for_services() {
    log_step "等待服务启动..."
    
    # 等待基础服务
    log_info "等待数据库服务..."
    sleep 30
    
    # 检查服务状态
    log_info "检查服务状态..."
    docker-compose -f deployment/docker-compose.wsl.yml ps
    
    # 等待应用服务
    log_info "等待应用服务..."
    sleep 60
    
    log_success "服务启动完成"
}

# 健康检查
health_check() {
    log_step "执行健康检查..."
    
    # 检查网关服务
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_success "网关服务健康"
    else
        log_warn "网关服务可能未就绪"
    fi
    
    # 检查Edge-Model-Infra
    if curl -f http://localhost:10001/health &> /dev/null; then
        log_success "Edge-Model-Infra Unit Manager健康"
    else
        log_warn "Edge-Model-Infra Unit Manager可能未就绪"
    fi
    
    # 检查Nginx
    if curl -f http://localhost:80/health &> /dev/null; then
        log_success "Nginx代理健康"
    else
        log_warn "Nginx代理可能未就绪"
    fi
}

# 显示部署信息
show_deployment_info() {
    log_step "部署信息"
    
    # 获取WSL IP
    WSL_IP=$(hostname -I | awk '{print $1}')
    
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  VideoCall System 部署完成！${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    echo -e "${GREEN}🌐 服务访问地址:${NC}"
    echo -e "  主网关: http://localhost:80"
    echo -e "  API网关: http://localhost:80/api"
    echo -e "  WebSocket: ws://localhost:80/ws"
    echo -e "  Edge AI: http://localhost:10001"
    echo ""
    echo -e "${GREEN}🖥️ Windows客户端配置:${NC}"
    echo -e "  WSL IP: $WSL_IP"
    echo -e "  后端URL: http://$WSL_IP:80"
    echo -e "  API URL: http://$WSL_IP:80/api"
    echo -e "  WebSocket: ws://$WSL_IP:80/ws"
    echo ""
    echo -e "${GREEN}🔧 管理命令:${NC}"
    echo -e "  查看日志: docker-compose -f deployment/docker-compose.wsl.yml logs -f"
    echo -e "  停止服务: docker-compose -f deployment/docker-compose.wsl.yml down"
    echo -e "  重启服务: docker-compose -f deployment/docker-compose.wsl.yml restart"
    echo ""
    echo -e "${GREEN}📊 服务状态:${NC}"
    docker-compose -f deployment/docker-compose.wsl.yml ps
    echo ""
    echo -e "${YELLOW}💡 提示: 请在Windows Qt客户端中配置WSL IP地址: $WSL_IP${NC}"
}

# 主函数
main() {
    echo -e "${PURPLE}========================================${NC}"
    echo -e "${PURPLE}  VideoCall System WSL后端部署${NC}"
    echo -e "${PURPLE}========================================${NC}"
    echo ""
    
    # 检查环境
    check_wsl
    check_docker
    check_docker_compose
    
    # 准备环境
    prepare_environment
    
    # 构建组件
    build_edge_infra
    build_backend_services
    
    # 部署服务
    deploy_services
    
    # 等待服务启动
    wait_for_services
    
    # 健康检查
    health_check
    
    # 显示部署信息
    show_deployment_info
    
    log_success "WSL后端部署完成！"
}

# 错误处理
trap 'log_error "部署过程中发生错误，请检查日志"; exit 1' ERR

# 执行主函数
main "$@"
