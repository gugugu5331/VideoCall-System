#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}服务集成测试 - 快速检查${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查服务状态
echo -e "${YELLOW}检查服务状态...${NC}"
echo ""

services=(
    "用户服务:http://localhost:8080/health"
    "会议服务:http://localhost:8082/health"
    "信令服务:http://localhost:8083/health"
    "媒体服务:http://localhost:8084/health"
    "AI服务:http://localhost:8085/health"
)

healthy_count=0
total_count=${#services[@]}

for service in "${services[@]}"; do
    IFS=':' read -r name url <<< "$service"
    if curl -s -f "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $name"
        ((healthy_count++))
    else
        echo -e "${RED}✗${NC} $name"
    fi
done

echo ""
echo -e "服务状态: ${GREEN}$healthy_count${NC}/${total_count} 健康"
echo ""

# 检查是否至少有2个服务运行
if [ $healthy_count -lt 2 ]; then
    echo -e "${RED}错误: 至少需要2个服务运行才能进行集成测试${NC}"
    echo -e "${YELLOW}提示: 请先启动服务${NC}"
    echo ""
    echo "启动服务命令:"
    echo "  cd ../user-service && go run main.go &"
    echo "  cd ../meeting-service && go run main.go &"
    echo "  cd ../signaling-service && go run main.go &"
    echo ""
    exit 1
fi

echo -e "${GREEN}✓ 服务检查通过，可以运行测试${NC}"
echo ""

# 询问用户要运行哪个测试
echo -e "${BLUE}选择要运行的测试:${NC}"
echo "  1) 服务集成测试 (推荐)"
echo "  2) 端到端测试"
echo "  3) 所有测试"
echo "  4) 仅编译检查"
echo ""
read -p "请选择 (1-4): " choice

case $choice in
    1)
        echo ""
        echo -e "${BLUE}========================================${NC}"
        echo -e "${BLUE}运行服务集成测试${NC}"
        echo -e "${BLUE}========================================${NC}"
        echo ""
        go test -v -timeout 10m -run TestServiceIntegrationTestSuite
        ;;
    2)
        echo ""
        echo -e "${BLUE}========================================${NC}"
        echo -e "${BLUE}运行端到端测试${NC}"
        echo -e "${BLUE}========================================${NC}"
        echo ""
        go test -v -timeout 10m -run TestEndToEndTestSuite
        ;;
    3)
        echo ""
        echo -e "${BLUE}========================================${NC}"
        echo -e "${BLUE}运行所有测试${NC}"
        echo -e "${BLUE}========================================${NC}"
        echo ""
        go test -v -timeout 15m
        ;;
    4)
        echo ""
        echo -e "${BLUE}========================================${NC}"
        echo -e "${BLUE}编译检查${NC}"
        echo -e "${BLUE}========================================${NC}"
        echo ""
        if go test -c -o /tmp/test.bin . 2>&1; then
            echo -e "${GREEN}✓ 编译成功${NC}"
            rm -f /tmp/test.bin
        else
            echo -e "${RED}✗ 编译失败${NC}"
            exit 1
        fi
        ;;
    *)
        echo -e "${RED}无效的选择${NC}"
        exit 1
        ;;
esac

echo ""
echo -e "${GREEN}测试完成！${NC}"

