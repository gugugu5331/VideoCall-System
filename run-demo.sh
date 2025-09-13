#!/bin/bash

# 设置颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo
echo "========================================"
echo "   视频会议系统演示版启动脚本"
echo "========================================"
echo

# 检查Go环境
echo -e "${BLUE}🔍 检查Go环境...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go未安装或未配置环境变量${NC}"
    echo "请先安装Go 1.21+: https://golang.org/dl/"
    exit 1
fi

echo -e "${GREEN}✅ Go环境检查通过${NC}"
echo "Go版本: $(go version)"

# 进入demo目录
cd demo

# 安装依赖
echo
echo -e "${BLUE}📦 安装依赖包...${NC}"
go mod tidy
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ 依赖包安装失败${NC}"
    exit 1
fi

echo -e "${GREEN}✅ 依赖包安装完成${NC}"

# 启动服务
echo
echo -e "${BLUE}🚀 启动视频会议系统演示版...${NC}"
echo
echo "服务将在以下地址启动:"
echo -e "  📍 主页: ${YELLOW}http://localhost:8080${NC}"
echo -e "  📖 API: ${YELLOW}http://localhost:8080/api/v1${NC}"
echo -e "  🔍 健康检查: ${YELLOW}http://localhost:8080/health${NC}"
echo -e "  💬 WebSocket: ${YELLOW}ws://localhost:8080/signaling${NC}"
echo -e "  🧪 测试页面: ${YELLOW}file://$(pwd)/test.html${NC}"
echo
echo -e "${YELLOW}按 Ctrl+C 停止服务${NC}"
echo

# 尝试打开浏览器
if command -v xdg-open &> /dev/null; then
    xdg-open "http://localhost:8080" &
    xdg-open "file://$(pwd)/test.html" &
elif command -v open &> /dev/null; then
    open "http://localhost:8080" &
    open "file://$(pwd)/test.html" &
fi

# 启动Go服务
go run main.go

echo
echo -e "${GREEN}服务已停止${NC}"
