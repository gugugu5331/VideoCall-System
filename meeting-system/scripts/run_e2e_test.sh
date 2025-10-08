#!/bin/bash

# 运行端到端测试的脚本

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

echo "========================================="
echo "运行端到端测试"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查服务是否运行
echo -e "${YELLOW}检查服务状态...${NC}"

SERVICES_OK=true

if ! pgrep -f "inference_server.py" > /dev/null; then
    echo -e "${RED}✗ Python推理服务未运行${NC}"
    SERVICES_OK=false
else
    echo -e "${GREEN}✓ Python推理服务运行中${NC}"
fi

if ! pgrep -f "ai-service" > /dev/null; then
    echo -e "${RED}✗ AI服务未运行${NC}"
    SERVICES_OK=false
else
    echo -e "${GREEN}✓ AI服务运行中${NC}"
fi

if [ "$SERVICES_OK" = false ]; then
    echo ""
    echo -e "${YELLOW}请先启动服务:${NC}"
    echo "  $SCRIPT_DIR/start_ai_services.sh"
    exit 1
fi

echo ""
echo -e "${YELLOW}运行测试...${NC}"
echo ""

# 进入媒体服务目录
cd "$PROJECT_ROOT/backend/media-service"

# 运行测试
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}开始端到端测试${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

go test -v -run TestE2EVideoProcessing -timeout 10m 2>&1 | tee test_output.log

# 检查测试结果
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo ""
    echo -e "${GREEN}=========================================${NC}"
    echo -e "${GREEN}✓ 测试通过！${NC}"
    echo -e "${GREEN}=========================================${NC}"
    
    # 显示结果统计
    echo ""
    echo -e "${YELLOW}查看AI推理结果:${NC}"
    echo "  cd $PROJECT_ROOT/ai-inference-service"
    echo "  python view_results.py --summary"
    echo "  python view_results.py --detail"
    
else
    echo ""
    echo -e "${RED}=========================================${NC}"
    echo -e "${RED}✗ 测试失败${NC}"
    echo -e "${RED}=========================================${NC}"
    
    echo ""
    echo "查看日志:"
    echo "  测试输出: $PROJECT_ROOT/backend/media-service/test_output.log"
    echo "  Python推理服务: $PROJECT_ROOT/ai-inference-service/logs/server.log"
    echo "  AI服务: $PROJECT_ROOT/backend/ai-service/logs/ai-service.log"
    
    exit 1
fi

echo ""
