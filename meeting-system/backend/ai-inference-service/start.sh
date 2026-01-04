#!/bin/bash

# AI Inference Service 启动脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}AI Inference Service Startup Script${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Go version: $(go version)${NC}"

# 检查配置文件
CONFIG_FILE="config/ai-inference-service.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}❌ Configuration file not found: $CONFIG_FILE${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Configuration file found${NC}"

# 创建日志目录
mkdir -p logs
echo -e "${GREEN}✓ Log directory created${NC}"

# 提示 Triton 依赖
echo -e "${YELLOW}ℹ 请确保 Triton Inference Server 可访问（ai.runtime.triton.endpoint）${NC}"

# 下载依赖
echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod download
echo -e "${GREEN}✓ Dependencies downloaded${NC}"

# 编译服务
echo -e "${YELLOW}Building service...${NC}"
go build -o ai-inference-service .
if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Build failed${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Service built successfully${NC}"

# 启动服务
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Starting AI Inference Service...${NC}"
echo -e "${GREEN}========================================${NC}"

./ai-inference-service --config "$CONFIG_FILE"
