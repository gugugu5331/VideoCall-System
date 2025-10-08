#!/bin/bash

# 启动AI服务的脚本

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

echo "========================================="
echo "启动AI服务"
echo "========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查Python推理服务
echo -e "${YELLOW}[1/3] 检查Python推理服务...${NC}"
INFERENCE_DIR="$PROJECT_ROOT/ai-inference-service"

if [ ! -d "$INFERENCE_DIR/venv" ]; then
    echo -e "${RED}错误: 虚拟环境不存在${NC}"
    echo "请先运行: cd $INFERENCE_DIR && ./setup.sh"
    exit 1
fi

# 启动Python推理服务
echo -e "${GREEN}启动Python推理服务...${NC}"
cd "$INFERENCE_DIR"
source venv/bin/activate

# 检查是否已经在运行
if pgrep -f "inference_server.py" > /dev/null; then
    echo -e "${YELLOW}Python推理服务已在运行${NC}"
else
    nohup python inference_server.py > logs/server.log 2>&1 &
    INFERENCE_PID=$!
    echo "Python推理服务已启动 (PID: $INFERENCE_PID)"
    echo "日志: $INFERENCE_DIR/logs/server.log"
    
    # 等待服务启动
    echo "等待服务启动..."
    sleep 5
fi

# 启动AI服务 (Go)
echo ""
echo -e "${YELLOW}[2/3] 启动AI服务 (Go)...${NC}"
AI_SERVICE_DIR="$PROJECT_ROOT/backend/ai-service"

cd "$AI_SERVICE_DIR"

# 检查是否已经在运行
if pgrep -f "ai-service" > /dev/null; then
    echo -e "${YELLOW}AI服务已在运行${NC}"
else
    # 编译
    if [ ! -f "ai-service" ]; then
        echo "编译AI服务..."
        go build -o ai-service main.go
    fi
    
    nohup ./ai-service > logs/ai-service.log 2>&1 &
    AI_SERVICE_PID=$!
    echo "AI服务已启动 (PID: $AI_SERVICE_PID)"
    echo "日志: $AI_SERVICE_DIR/logs/ai-service.log"
    
    # 等待服务启动
    sleep 3
fi

# 检查服务状态
echo ""
echo -e "${YELLOW}[3/3] 检查服务状态...${NC}"

# 检查Python推理服务
if pgrep -f "inference_server.py" > /dev/null; then
    echo -e "${GREEN}✓ Python推理服务运行中${NC}"
    
    # 检查ZMQ端口
    if netstat -tuln 2>/dev/null | grep -q ":5555"; then
        echo -e "${GREEN}✓ ZMQ端口 5555 已监听${NC}"
    else
        echo -e "${RED}✗ ZMQ端口 5555 未监听${NC}"
    fi
else
    echo -e "${RED}✗ Python推理服务未运行${NC}"
fi

# 检查AI服务
if pgrep -f "ai-service" > /dev/null; then
    echo -e "${GREEN}✓ AI服务运行中${NC}"
    
    # 检查HTTP端口
    if netstat -tuln 2>/dev/null | grep -q ":8085"; then
        echo -e "${GREEN}✓ HTTP端口 8085 已监听${NC}"
    else
        echo -e "${RED}✗ HTTP端口 8085 未监听${NC}"
    fi
else
    echo -e "${RED}✗ AI服务未运行${NC}"
fi

echo ""
echo "========================================="
echo "服务启动完成"
echo "========================================="
echo ""
echo "查看日志:"
echo "  Python推理服务: tail -f $INFERENCE_DIR/logs/server.log"
echo "  AI服务: tail -f $AI_SERVICE_DIR/logs/ai-service.log"
echo ""
echo "查看结果:"
echo "  cd $INFERENCE_DIR"
echo "  python view_results.py --summary"
echo ""
echo "停止服务:"
echo "  pkill -f inference_server.py"
echo "  pkill -f ai-service"
echo ""
