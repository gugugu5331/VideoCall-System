#!/bin/bash

# 视频会议系统部署脚本
# 支持Windows 11环境下的Docker部署

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查系统依赖..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker Desktop"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    # 检查Git
    if ! command -v git &> /dev/null; then
        log_warning "Git未安装，某些功能可能受限"
    fi
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_warning "Go未安装，无法本地编译后端服务"
    fi
    
    # 检查Node.js (如果需要前端构建)
    if ! command -v node &> /dev/null; then
        log_warning "Node.js未安装，无法构建Web前端"
    fi
    
    log_success "依赖检查完成"
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    
    mkdir -p logs
    mkdir -p data/postgres
    mkdir -p data/mongodb
    mkdir -p data/redis
    mkdir -p data/rabbitmq
    mkdir -p storage/media
    mkdir -p storage/detection
    mkdir -p storage/uploads
    mkdir -p ai-detection/models
    mkdir -p backend/deploy/nginx
    mkdir -p backend/deploy/ssl
    
    log_success "目录创建完成"
}

# 生成配置文件
generate_configs() {
    log_info "生成配置文件..."
    
    # 生成环境变量文件
    cat > .env << EOF
# 数据库配置
POSTGRES_DB=video_conference
POSTGRES_USER=admin
POSTGRES_PASSWORD=password123

MONGODB_USER=admin
MONGODB_PASSWORD=password123

REDIS_PASSWORD=

RABBITMQ_USER=admin
RABBITMQ_PASSWORD=password123

# 服务配置
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
API_BASE_URL=http://localhost:8080

# AI服务配置
AI_SERVICE_URL=http://ai-detection:8501

# 存储配置
STORAGE_PATH=./storage

# 日志配置
LOG_LEVEL=info
EOF
    
    # 生成Nginx配置
    cat > backend/deploy/nginx/nginx.conf << 'EOF'
events {
    worker_connections 1024;
}

http {
    upstream api_gateway {
        server gateway:8080;
    }
    
    upstream signaling_service {
        server signaling-service:8080;
    }
    
    server {
        listen 80;
        server_name localhost;
        
        # API网关
        location /api/ {
            proxy_pass http://api_gateway;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
        # WebSocket信令
        location /signaling/ {
            proxy_pass http://signaling_service;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
        
        # 静态文件
        location / {
            root /usr/share/nginx/html;
            index index.html index.htm;
            try_files $uri $uri/ /index.html;
        }
        
        # 健康检查
        location /health {
            access_log off;
            return 200 "healthy\n";
            add_header Content-Type text/plain;
        }
    }
}
EOF
    
    log_success "配置文件生成完成"
}

# 构建Docker镜像
build_images() {
    log_info "构建Docker镜像..."
    
    # 构建后端服务镜像
    log_info "构建用户服务镜像..."
    docker build -t video-conference/user-service:latest -f backend/services/user/Dockerfile backend/
    
    log_info "构建会议服务镜像..."
    docker build -t video-conference/meeting-service:latest -f backend/services/meeting/Dockerfile backend/
    
    log_info "构建信令服务镜像..."
    docker build -t video-conference/signaling-service:latest -f backend/services/signaling/Dockerfile backend/
    
    log_info "构建媒体服务镜像..."
    docker build -t video-conference/media-service:latest -f backend/services/media/Dockerfile backend/
    
    log_info "构建检测服务镜像..."
    docker build -t video-conference/detection-service:latest -f backend/services/detection/Dockerfile backend/
    
    log_info "构建记录服务镜像..."
    docker build -t video-conference/record-service:latest -f backend/services/record/Dockerfile backend/
    
    log_info "构建通知服务镜像..."
    docker build -t video-conference/notification-service:latest -f backend/services/notification/Dockerfile backend/
    
    log_info "构建网关服务镜像..."
    docker build -t video-conference/gateway-service:latest -f backend/services/gateway/Dockerfile backend/
    
    # 构建AI检测服务镜像
    log_info "构建AI检测服务镜像..."
    docker build -t video-conference/ai-detection:latest ai-detection/
    
    log_success "Docker镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    # 首先启动基础设施服务
    log_info "启动基础设施服务..."
    docker-compose up -d postgres mongodb redis rabbitmq consul
    
    # 等待数据库启动
    log_info "等待数据库启动..."
    sleep 30
    
    # 启动应用服务
    log_info "启动应用服务..."
    docker-compose up -d
    
    log_success "服务启动完成"
}

# 等待服务就绪
wait_for_services() {
    log_info "等待服务就绪..."
    
    # 检查服务健康状态
    services=("user-service" "meeting-service" "signaling-service" "media-service" "detection-service" "record-service" "ai-detection")
    
    for service in "${services[@]}"; do
        log_info "检查 $service 服务状态..."
        
        max_attempts=30
        attempt=1
        
        while [ $attempt -le $max_attempts ]; do
            if docker-compose ps $service | grep -q "Up"; then
                log_success "$service 服务已启动"
                break
            fi
            
            if [ $attempt -eq $max_attempts ]; then
                log_error "$service 服务启动失败"
                docker-compose logs $service
                exit 1
            fi
            
            log_info "等待 $service 服务启动... (尝试 $attempt/$max_attempts)"
            sleep 10
            ((attempt++))
        done
    done
    
    log_success "所有服务已就绪"
}

# 运行测试
run_tests() {
    log_info "运行系统测试..."
    
    # API健康检查测试
    log_info "测试API健康检查..."
    
    services=("user-service:8080" "meeting-service:8080" "signaling-service:8080" "media-service:8080" "detection-service:8080" "record-service:8080")
    
    for service in "${services[@]}"; do
        service_name=$(echo $service | cut -d':' -f1)
        port=$(echo $service | cut -d':' -f2)
        
        if curl -f -s "http://localhost:$port/health" > /dev/null; then
            log_success "$service_name 健康检查通过"
        else
            log_error "$service_name 健康检查失败"
        fi
    done
    
    # 数据库连接测试
    log_info "测试数据库连接..."
    
    # PostgreSQL
    if docker-compose exec -T postgres pg_isready -U admin -d video_conference > /dev/null 2>&1; then
        log_success "PostgreSQL 连接正常"
    else
        log_error "PostgreSQL 连接失败"
    fi
    
    # MongoDB
    if docker-compose exec -T mongodb mongosh --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
        log_success "MongoDB 连接正常"
    else
        log_error "MongoDB 连接失败"
    fi
    
    # Redis
    if docker-compose exec -T redis redis-cli ping | grep -q "PONG"; then
        log_success "Redis 连接正常"
    else
        log_error "Redis 连接失败"
    fi
    
    log_success "系统测试完成"
}

# 显示部署信息
show_deployment_info() {
    log_success "部署完成！"
    echo
    echo "=== 服务访问信息 ==="
    echo "API网关: http://localhost:80"
    echo "用户服务: http://localhost:8081"
    echo "会议服务: http://localhost:8082"
    echo "信令服务: http://localhost:8083"
    echo "媒体服务: http://localhost:8084"
    echo "检测服务: http://localhost:8085"
    echo "记录服务: http://localhost:8086"
    echo "AI检测服务: http://localhost:8501"
    echo
    echo "=== 管理界面 ==="
    echo "RabbitMQ管理: http://localhost:15672 (admin/password123)"
    echo "Consul管理: http://localhost:8500"
    echo
    echo "=== 数据库连接 ==="
    echo "PostgreSQL: localhost:5432 (admin/password123)"
    echo "MongoDB: localhost:27017 (admin/password123)"
    echo "Redis: localhost:6379"
    echo
    echo "=== 常用命令 ==="
    echo "查看服务状态: docker-compose ps"
    echo "查看服务日志: docker-compose logs [service-name]"
    echo "停止所有服务: docker-compose down"
    echo "重启服务: docker-compose restart [service-name]"
    echo
}

# 清理函数
cleanup() {
    log_info "清理部署环境..."
    docker-compose down -v
    docker system prune -f
    log_success "清理完成"
}

# 主函数
main() {
    case "${1:-deploy}" in
        "deploy")
            log_info "开始部署视频会议系统..."
            check_dependencies
            create_directories
            generate_configs
            build_images
            start_services
            wait_for_services
            run_tests
            show_deployment_info
            ;;
        "start")
            log_info "启动服务..."
            start_services
            wait_for_services
            show_deployment_info
            ;;
        "stop")
            log_info "停止服务..."
            docker-compose down
            log_success "服务已停止"
            ;;
        "restart")
            log_info "重启服务..."
            docker-compose restart
            log_success "服务已重启"
            ;;
        "logs")
            docker-compose logs -f "${2:-}"
            ;;
        "status")
            docker-compose ps
            ;;
        "test")
            run_tests
            ;;
        "cleanup")
            cleanup
            ;;
        "help")
            echo "用法: $0 [command]"
            echo
            echo "命令:"
            echo "  deploy   - 完整部署系统 (默认)"
            echo "  start    - 启动服务"
            echo "  stop     - 停止服务"
            echo "  restart  - 重启服务"
            echo "  logs     - 查看日志"
            echo "  status   - 查看服务状态"
            echo "  test     - 运行测试"
            echo "  cleanup  - 清理环境"
            echo "  help     - 显示帮助"
            ;;
        *)
            log_error "未知命令: $1"
            echo "使用 '$0 help' 查看可用命令"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
