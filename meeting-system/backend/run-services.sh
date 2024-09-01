#!/bin/bash

# 运行微服务脚本
# 用于直接运行Go源码而不是构建二进制文件

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 启动会议系统微服务 (开发模式)${NC}"
echo "=================================="

# 检查Go环境
check_go() {
    echo -e "${YELLOW}🔍 检查Go环境...${NC}"
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go未安装${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Go环境正常${NC}"
    go version
}

# 启动数据库服务
start_databases() {
    echo -e "${YELLOW}🗄️ 启动数据库服务...${NC}"
    
    cd ../deployment/docker
    
    # 启动数据库相关服务
    docker-compose up -d postgres redis mongodb minio
    
    echo -e "${YELLOW}⏳ 等待数据库服务启动...${NC}"
    sleep 15
    
    cd ../../backend
}

# 启动用户服务
start_user_service() {
    echo -e "${YELLOW}🚀 启动用户服务...${NC}"
    
    # 设置环境变量
    export CONFIG_PATH="config/config-docker.yaml"
    export GIN_MODE="debug"
    
    # 启动用户服务
    cd user-service
    go run main.go -config ../config/config-docker.yaml &
    USER_SERVICE_PID=$!
    echo "用户服务 PID: $USER_SERVICE_PID"
    
    cd ..
    
    # 等待服务启动
    sleep 5
    
    # 检查服务是否启动
    if curl -s http://localhost:8081/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 用户服务启动成功${NC}"
    else
        echo -e "${RED}❌ 用户服务启动失败${NC}"
        kill $USER_SERVICE_PID 2>/dev/null || true
        exit 1
    fi
}

# 启动会议服务
start_meeting_service() {
    echo -e "${YELLOW}🚀 启动会议服务...${NC}"
    
    # 设置环境变量
    export CONFIG_PATH="config/config-docker.yaml"
    export GIN_MODE="debug"
    
    # 启动会议服务
    cd meeting-service
    go run main.go -config ../config/config-docker.yaml &
    MEETING_SERVICE_PID=$!
    echo "会议服务 PID: $MEETING_SERVICE_PID"
    
    cd ..
    
    # 等待服务启动
    sleep 5
    
    # 检查服务是否启动
    if curl -s http://localhost:8082/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 会议服务启动成功${NC}"
    else
        echo -e "${RED}❌ 会议服务启动失败${NC}"
        kill $MEETING_SERVICE_PID 2>/dev/null || true
        kill $USER_SERVICE_PID 2>/dev/null || true
        exit 1
    fi
}

# 显示服务状态
show_status() {
    echo ""
    echo -e "${BLUE}📊 服务状态${NC}"
    echo "=================================="
    echo -e "${GREEN}用户服务:${NC} http://localhost:8081/health"
    echo -e "${GREEN}会议服务:${NC} http://localhost:8082/health"
    
    echo ""
    echo -e "${BLUE}🔧 测试命令${NC}"
    echo "=================================="
    echo "运行测试: ./test-services.sh"
    echo "停止服务: kill $USER_SERVICE_PID $MEETING_SERVICE_PID"
}

# 清理函数
cleanup() {
    echo -e "${YELLOW}🛑 停止服务...${NC}"
    kill $USER_SERVICE_PID 2>/dev/null || true
    kill $MEETING_SERVICE_PID 2>/dev/null || true
    echo -e "${GREEN}✅ 服务已停止${NC}"
}

# 设置信号处理
trap cleanup EXIT INT TERM

# 主函数
main() {
    check_go
    start_databases
    start_user_service
    start_meeting_service
    show_status
    
    echo ""
    echo -e "${GREEN}🎉 所有服务启动完成！${NC}"
    echo -e "${YELLOW}💡 按 Ctrl+C 停止所有服务${NC}"
    
    # 保持脚本运行
    wait
}

# 运行主函数
main "$@"
