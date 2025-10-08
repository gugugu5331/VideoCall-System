#!/bin/bash

# 启动所有微服务的脚本
# 用途：启动智能视频会议平台的所有后端服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 全局变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$PROJECT_ROOT/backend"
PIDS_FILE="$PROJECT_ROOT/service_pids.txt"

# 清理函数
cleanup() {
    log_step "正在停止所有服务..."
    
    if [ -f "$PIDS_FILE" ]; then
        while read -r line; do
            if [ -n "$line" ]; then
                service_name=$(echo "$line" | cut -d':' -f1)
                pid=$(echo "$line" | cut -d':' -f2)
                
                if kill -0 "$pid" 2>/dev/null; then
                    log_info "停止 $service_name (PID: $pid)"
                    kill "$pid"
                    sleep 1
                    
                    # 如果进程仍在运行，强制杀死
                    if kill -0 "$pid" 2>/dev/null; then
                        log_warn "强制停止 $service_name (PID: $pid)"
                        kill -9 "$pid"
                    fi
                fi
            fi
        done < "$PIDS_FILE"
        
        rm -f "$PIDS_FILE"
    fi
    
    log_info "所有服务已停止"
}

# 设置信号处理
trap cleanup EXIT INT TERM

# 启动单个服务
start_service() {
    local service_name=$1
    local service_dir=$2
    local port=$3
    local grpc_port=$4
    
    log_info "启动 $service_name..."
    
    cd "$BACKEND_DIR/$service_dir"
    
    # 设置Go代理
    export GOPROXY=https://goproxy.cn,direct
    
    # 启动服务
    nohup go run main.go > "logs/${service_name}.log" 2>&1 &
    local pid=$!
    
    # 记录PID
    echo "$service_name:$pid" >> "$PIDS_FILE"
    
    log_info "$service_name 已启动 (PID: $pid, HTTP: $port, gRPC: $grpc_port)"
    
    # 等待服务启动
    sleep 3
    
    # 检查服务是否正常启动
    if ! kill -0 "$pid" 2>/dev/null; then
        log_error "$service_name 启动失败"
        return 1
    fi
    
    # 检查HTTP端口
    if [ -n "$port" ]; then
        local max_attempts=10
        local attempt=1
        
        while [ $attempt -le $max_attempts ]; do
            if curl -f -s "http://localhost:$port/health" > /dev/null 2>&1; then
                log_info "$service_name HTTP服务就绪 (端口: $port)"
                break
            fi
            
            if [ $attempt -eq $max_attempts ]; then
                log_warn "$service_name HTTP服务可能未就绪 (端口: $port)"
            fi
            
            sleep 2
            ((attempt++))
        done
    fi
    
    return 0
}

# 检查依赖
check_dependencies() {
    log_step "检查依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go未安装，请先安装Go"
        exit 1
    fi
    
    # 检查PostgreSQL
    if ! command -v psql &> /dev/null; then
        log_warn "psql未安装，无法直接测试PostgreSQL连接"
    else
        if ! PGPASSWORD=password psql -h localhost -U postgres -d meeting_system -c "SELECT 1;" > /dev/null 2>&1; then
            log_error "PostgreSQL连接失败，请确保数据库服务正在运行"
            exit 1
        fi
        log_info "PostgreSQL连接正常"
    fi
    
    # 检查Redis
    if ! command -v redis-cli &> /dev/null; then
        log_warn "redis-cli未安装，无法直接测试Redis连接"
    else
        if ! redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then
            log_error "Redis连接失败，请确保Redis服务正在运行"
            exit 1
        fi
        log_info "Redis连接正常"
    fi
    
    log_info "依赖检查完成"
}

# 创建日志目录
create_log_dirs() {
    log_step "创建日志目录..."
    
    mkdir -p "$BACKEND_DIR/user-service/logs"
    mkdir -p "$BACKEND_DIR/meeting-service/logs"
    mkdir -p "$BACKEND_DIR/signaling-service/logs"
    mkdir -p "$BACKEND_DIR/media-service/logs"
    mkdir -p "$BACKEND_DIR/ai-service/logs"
    
    log_info "日志目录创建完成"
}

# 编译所有服务
build_services() {
    log_step "编译所有服务..."
    
    cd "$BACKEND_DIR"
    
    export GOPROXY=https://goproxy.cn,direct
    
    if ! go build ./...; then
        log_error "服务编译失败"
        exit 1
    fi
    
    log_info "所有服务编译成功"
}

# 启动所有服务
start_all_services() {
    log_step "启动所有微服务..."
    
    # 清空PID文件
    > "$PIDS_FILE"
    
    # 启动用户服务
    start_service "user-service" "user-service" "8080" "50051"
    
    # 启动会议服务
    start_service "meeting-service" "meeting-service" "8082" "50052"
    
    # 启动信令服务
    start_service "signaling-service" "signaling-service" "8081" ""
    
    # 启动媒体服务
    start_service "media-service" "media-service" "8083" "50053"
    
    # 启动AI服务
    start_service "ai-service" "ai-service" "8084" "50054"
    
    log_info "所有服务启动完成"
}

# 显示服务状态
show_service_status() {
    log_step "服务状态检查..."
    
    echo ""
    echo "服务状态："
    echo "----------------------------------------"
    
    local services=(
        "user-service:8080"
        "meeting-service:8082"
        "signaling-service:8081"
        "media-service:8083"
        "ai-service:8084"
    )
    
    for service in "${services[@]}"; do
        service_name=$(echo $service | cut -d':' -f1)
        port=$(echo $service | cut -d':' -f2)
        
        if curl -f -s "http://localhost:$port/health" > /dev/null 2>&1; then
            echo "✅ $service_name (http://localhost:$port)"
        else
            echo "❌ $service_name (http://localhost:$port)"
        fi
    done
    
    echo "----------------------------------------"
    echo ""
}

# 等待用户输入
wait_for_user() {
    echo "所有服务已启动。按 Ctrl+C 停止所有服务。"
    echo ""
    echo "可用的服务端点："
    echo "- 用户服务: http://localhost:8080"
    echo "- 会议服务: http://localhost:8082"
    echo "- 信令服务: http://localhost:8081"
    echo "- 媒体服务: http://localhost:8083"
    echo "- AI服务: http://localhost:8084"
    echo ""
    echo "健康检查端点: /health"
    echo "Prometheus指标: /metrics"
    echo ""
    
    # 等待中断信号
    while true; do
        sleep 1
    done
}

# 主函数
main() {
    echo "=========================================="
    echo "    智能视频会议平台服务启动器"
    echo "=========================================="
    echo ""
    
    check_dependencies
    create_log_dirs
    build_services
    start_all_services
    show_service_status
    wait_for_user
}

# 执行主函数
main "$@"
