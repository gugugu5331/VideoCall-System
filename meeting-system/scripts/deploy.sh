#!/bin/bash

# 智能视频会议平台部署脚本
# 用途：一键部署整个微服务架构

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

# 检查依赖
check_dependencies() {
    log_step "检查系统依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_warn "Go未安装，将跳过本地构建"
    fi
    
    log_info "依赖检查完成"
}

# 创建必要的目录
create_directories() {
    log_step "创建必要的目录..."
    
    mkdir -p logs
    mkdir -p data/postgres
    mkdir -p data/redis
    mkdir -p data/mongodb
    mkdir -p data/minio
    mkdir -p data/recordings
    mkdir -p data/hls
    mkdir -p ssl
    mkdir -p monitoring/grafana/dashboards
    mkdir -p monitoring/grafana/provisioning/dashboards
    mkdir -p monitoring/grafana/provisioning/datasources
    
    log_info "目录创建完成"
}

# 生成SSL证书
generate_ssl_certificates() {
    log_step "生成SSL证书..."
    
    if [ ! -f "ssl/cert.pem" ] || [ ! -f "ssl/key.pem" ]; then
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout ssl/key.pem \
            -out ssl/cert.pem \
            -subj "/C=CN/ST=Beijing/L=Beijing/O=Meeting System/CN=meeting.example.com"
        
        log_info "SSL证书生成完成"
    else
        log_info "SSL证书已存在，跳过生成"
    fi
}

# 构建服务镜像
build_services() {
    log_step "构建服务镜像..."
    
    # 构建用户服务
    log_info "构建用户服务..."
    docker build -t meeting-user-service:latest ./backend/user-service/
    
    # 构建信令服务
    log_info "构建信令服务..."
    docker build -t meeting-signaling-service:latest ./backend/signaling-service/
    
    # 构建会议服务
    log_info "构建会议服务..."
    docker build -t meeting-meeting-service:latest ./backend/meeting-service/
    
    # 构建媒体服务
    log_info "构建媒体服务..."
    docker build -t meeting-media-service:latest ./backend/media-service/
    
    # 构建AI服务
    log_info "构建AI服务..."
    docker build -t meeting-ai-service:latest ./backend/ai-service/
    
    # 构建通知服务
    log_info "构建通知服务..."
    docker build -t meeting-notification-service:latest ./backend/notification-service/
    
    log_info "服务镜像构建完成"
}

# 初始化数据库
init_database() {
    log_step "初始化数据库..."
    
    # 启动PostgreSQL
    docker-compose up -d postgres
    
    # 等待PostgreSQL启动
    log_info "等待PostgreSQL启动..."
    sleep 10
    
    # 检查PostgreSQL是否就绪
    until docker-compose exec postgres pg_isready -U postgres; do
        log_info "等待PostgreSQL就绪..."
        sleep 2
    done
    
    # 执行数据库初始化脚本
    if [ -f "backend/shared/database/schema.sql" ]; then
        docker-compose exec -T postgres psql -U postgres -d meeting_system < backend/shared/database/schema.sql
        log_info "数据库初始化完成"
    else
        log_warn "数据库初始化脚本不存在，跳过初始化"
    fi
}

# 配置Grafana
setup_grafana() {
    log_step "配置Grafana..."
    
    # 创建数据源配置
    cat > monitoring/grafana/provisioning/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
EOF

    # 创建仪表板配置
    cat > monitoring/grafana/provisioning/dashboards/dashboard.yml << EOF
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

    log_info "Grafana配置完成"
}

# 启动所有服务
start_services() {
    log_step "启动所有服务..."
    
    # 按依赖顺序启动服务
    log_info "启动基础设施服务..."
    docker-compose up -d postgres redis mongodb minio
    
    # 等待基础设施服务启动
    sleep 15
    
    log_info "启动应用服务..."
    docker-compose up -d user-service signaling-service meeting-service media-service ai-service notification-service
    
    # 等待应用服务启动
    sleep 10
    
    log_info "启动网关和监控服务..."
    docker-compose up -d nginx prometheus grafana
    
    log_info "所有服务启动完成"
}

# 健康检查
health_check() {
    log_step "执行健康检查..."
    
    services=("user-service:8080" "signaling-service:8081" "meeting-service:8082" "media-service:8083" "ai-service:8084" "notification-service:8085")
    
    for service in "${services[@]}"; do
        service_name=$(echo $service | cut -d':' -f1)
        port=$(echo $service | cut -d':' -f2)
        
        log_info "检查 $service_name..."
        
        max_attempts=30
        attempt=1
        
        while [ $attempt -le $max_attempts ]; do
            if curl -f -s "http://localhost:$port/health" > /dev/null; then
                log_info "$service_name 健康检查通过"
                break
            fi
            
            if [ $attempt -eq $max_attempts ]; then
                log_error "$service_name 健康检查失败"
                return 1
            fi
            
            sleep 2
            ((attempt++))
        done
    done
    
    log_info "所有服务健康检查通过"
}

# 显示服务状态
show_status() {
    log_step "显示服务状态..."
    
    echo ""
    echo "=== 服务状态 ==="
    docker-compose ps
    
    echo ""
    echo "=== 服务访问地址 ==="
    echo "Web界面: https://localhost"
    echo "用户服务: http://localhost:8080"
    echo "信令服务: http://localhost:8081"
    echo "会议服务: http://localhost:8082"
    echo "媒体服务: http://localhost:8083"
    echo "AI服务: http://localhost:8084"
    echo "通知服务: http://localhost:8085"
    echo "Grafana监控: http://localhost:3000 (admin/admin)"
    echo "Prometheus: http://localhost:9090"
    echo "MinIO控制台: http://localhost:9001 (minioadmin/minioadmin)"
    echo ""
}

# 清理函数
cleanup() {
    log_step "清理资源..."
    docker-compose down
    log_info "清理完成"
}

# 主函数
main() {
    echo "=========================================="
    echo "    智能视频会议平台部署脚本"
    echo "=========================================="
    echo ""
    
    case "${1:-deploy}" in
        "deploy")
            check_dependencies
            create_directories
            generate_ssl_certificates
            build_services
            setup_grafana
            init_database
            start_services
            health_check
            show_status
            ;;
        "start")
            log_info "启动现有服务..."
            docker-compose up -d
            health_check
            show_status
            ;;
        "stop")
            log_info "停止所有服务..."
            docker-compose down
            ;;
        "restart")
            log_info "重启所有服务..."
            docker-compose restart
            health_check
            show_status
            ;;
        "logs")
            service_name=${2:-}
            if [ -n "$service_name" ]; then
                docker-compose logs -f "$service_name"
            else
                docker-compose logs -f
            fi
            ;;
        "status")
            show_status
            ;;
        "clean")
            log_warn "这将删除所有容器和数据，确定要继续吗？(y/N)"
            read -r response
            if [[ "$response" =~ ^[Yy]$ ]]; then
                docker-compose down -v --remove-orphans
                docker system prune -f
                log_info "清理完成"
            else
                log_info "取消清理"
            fi
            ;;
        "help")
            echo "用法: $0 [命令]"
            echo ""
            echo "命令:"
            echo "  deploy   - 完整部署系统 (默认)"
            echo "  start    - 启动现有服务"
            echo "  stop     - 停止所有服务"
            echo "  restart  - 重启所有服务"
            echo "  logs     - 查看日志 (可指定服务名)"
            echo "  status   - 显示服务状态"
            echo "  clean    - 清理所有容器和数据"
            echo "  help     - 显示帮助信息"
            echo ""
            echo "示例:"
            echo "  $0 deploy          # 完整部署"
            echo "  $0 logs nginx      # 查看nginx日志"
            echo "  $0 restart         # 重启所有服务"
            ;;
        *)
            log_error "未知命令: $1"
            echo "使用 '$0 help' 查看帮助信息"
            exit 1
            ;;
    esac
}

# 捕获中断信号
trap cleanup INT TERM

# 执行主函数
main "$@"
