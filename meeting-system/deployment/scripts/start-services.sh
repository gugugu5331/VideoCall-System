#!/bin/bash

# 启动微服务脚本
# 用于启动用户服务和会议服务进行测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
DOCKER_DIR="$PROJECT_ROOT/deployment/docker"

echo -e "${BLUE}🚀 启动会议系统微服务${NC}"
echo "项目根目录: $PROJECT_ROOT"
echo "=================================="

# 检查Docker是否运行
check_docker() {
    echo -e "${YELLOW}🔍 检查Docker状态...${NC}"
    if ! docker info > /dev/null 2>&1; then
        echo -e "${RED}❌ Docker未运行，请先启动Docker${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Docker运行正常${NC}"
}

# 构建Go服务
build_services() {
    echo -e "${YELLOW}🔨 构建Go微服务...${NC}"
    
    cd "$BACKEND_DIR"
    
    # 检查go.mod文件
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}❌ 未找到go.mod文件${NC}"
        exit 1
    fi
    
    # 下载依赖
    echo -e "${YELLOW}📦 下载Go依赖...${NC}"
    go mod download
    go mod tidy
    
    # 构建用户服务
    echo -e "${YELLOW}🔨 构建用户服务...${NC}"
    cd user-service
    go build -o user-service main.go
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 用户服务构建成功${NC}"
    else
        echo -e "${RED}❌ 用户服务构建失败${NC}"
        exit 1
    fi
    
    # 构建会议服务
    echo -e "${YELLOW}🔨 构建会议服务...${NC}"
    cd ../meeting-service
    go build -o meeting-service main.go
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 会议服务构建成功${NC}"
    else
        echo -e "${RED}❌ 会议服务构建失败${NC}"
        exit 1
    fi
    
    cd "$PROJECT_ROOT"
}

# 启动数据库服务
start_databases() {
    echo -e "${YELLOW}🗄️ 启动数据库服务...${NC}"
    
    cd "$DOCKER_DIR"
    
    # 启动数据库相关服务
    docker-compose up -d postgres redis mongodb minio
    
    echo -e "${YELLOW}⏳ 等待数据库服务启动...${NC}"
    sleep 10
    
    # 检查服务状态
    check_service_health "postgres" "PostgreSQL"
    check_service_health "redis" "Redis"
    check_service_health "mongodb" "MongoDB"
    check_service_health "minio" "MinIO"
}

# 检查服务健康状态
check_service_health() {
    local service_name=$1
    local display_name=$2
    local max_attempts=30
    local attempt=1
    
    echo -e "${YELLOW}🔍 检查 $display_name 健康状态...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if docker-compose exec -T $service_name echo "healthy" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $display_name 健康检查通过${NC}"
            return 0
        fi
        
        echo -e "${YELLOW}⏳ 等待 $display_name 启动 (尝试 $attempt/$max_attempts)${NC}"
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}❌ $display_name 健康检查超时${NC}"
    return 1
}

# 初始化数据库
init_database() {
    echo -e "${YELLOW}🗄️ 初始化数据库...${NC}"
    
    cd "$DOCKER_DIR"
    
    # 等待PostgreSQL完全启动
    sleep 5
    
    # 检查数据库是否已初始化
    if docker-compose exec -T postgres psql -U postgres -d meeting_system -c "SELECT 1;" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 数据库已存在${NC}"
    else
        echo -e "${YELLOW}📝 创建数据库...${NC}"
        docker-compose exec -T postgres createdb -U postgres meeting_system || true
    fi
    
    # 运行数据库迁移脚本
    if [ -f "../../backend/shared/database/schema.sql" ]; then
        echo -e "${YELLOW}📝 执行数据库迁移...${NC}"
        docker-compose exec -T postgres psql -U postgres -d meeting_system -f /docker-entrypoint-initdb.d/01-schema.sql || true
    fi
}

# 启动微服务
start_microservices() {
    echo -e "${YELLOW}🚀 启动微服务...${NC}"
    
    cd "$DOCKER_DIR"
    
    # 启动用户服务和会议服务
    docker-compose up -d user-service meeting-service
    
    echo -e "${YELLOW}⏳ 等待微服务启动...${NC}"
    sleep 15
    
    # 检查服务状态
    check_microservice_health "user-service" "用户服务" "8081"
    check_microservice_health "meeting-service" "会议服务" "8082"
}

# 检查微服务健康状态
check_microservice_health() {
    local service_name=$1
    local display_name=$2
    local port=$3
    local max_attempts=20
    local attempt=1
    
    echo -e "${YELLOW}🔍 检查 $display_name 健康状态...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $display_name 健康检查通过${NC}"
            return 0
        fi
        
        echo -e "${YELLOW}⏳ 等待 $display_name 启动 (尝试 $attempt/$max_attempts)${NC}"
        sleep 3
        ((attempt++))
    done
    
    echo -e "${RED}❌ $display_name 健康检查超时${NC}"
    echo -e "${YELLOW}📋 查看 $display_name 日志:${NC}"
    docker-compose logs --tail=20 $service_name
    return 1
}

# 显示服务状态
show_service_status() {
    echo ""
    echo -e "${BLUE}📊 服务状态总览${NC}"
    echo "=================================="
    
    cd "$DOCKER_DIR"
    docker-compose ps
    
    echo ""
    echo -e "${BLUE}🌐 服务访问地址${NC}"
    echo "=================================="
    echo -e "${GREEN}用户服务:${NC} http://localhost:8081"
    echo -e "${GREEN}会议服务:${NC} http://localhost:8082"
    echo -e "${GREEN}PostgreSQL:${NC} localhost:5432"
    echo -e "${GREEN}Redis:${NC} localhost:6379"
    echo -e "${GREEN}MongoDB:${NC} localhost:27017"
    echo -e "${GREEN}MinIO:${NC} http://localhost:9000 (admin/minioadmin)"
    
    echo ""
    echo -e "${BLUE}🔧 管理命令${NC}"
    echo "=================================="
    echo "查看日志: docker-compose logs -f [service-name]"
    echo "停止服务: docker-compose down"
    echo "重启服务: docker-compose restart [service-name]"
    echo "运行测试: ./test-services.sh"
}

# 主函数
main() {
    check_docker
    build_services
    start_databases
    init_database
    start_microservices
    show_service_status
    
    echo ""
    echo -e "${GREEN}🎉 微服务启动完成！${NC}"
    echo -e "${YELLOW}💡 运行测试: cd $BACKEND_DIR && ./test-services.sh${NC}"
}

# 错误处理
trap 'echo -e "${RED}❌ 启动过程中发生错误${NC}"; exit 1' ERR

# 运行主函数
main "$@"
