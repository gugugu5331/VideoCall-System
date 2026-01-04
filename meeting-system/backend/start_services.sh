#!/bin/bash

# 启动所有微服务
# 用途：快速启动用户服务和会议服务进行测试

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  启动智能视频会议平台微服务${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 清理函数
cleanup() {
    echo -e "\n${YELLOW}正在停止所有服务...${NC}"
    pkill -f "go run.*user-service" || true
    pkill -f "go run.*meeting-service" || true
    pkill -f "go run.*signaling-service" || true
    pkill -f "go run.*media-service" || true
    pkill -f "go run.*ai-inference-service" || true
    echo -e "${GREEN}所有服务已停止${NC}"
    exit 0
}

# 设置信号处理
trap cleanup EXIT INT TERM

# 检查配置文件
if [ ! -f "config/config.yaml" ]; then
    echo -e "${RED}错误: 配置文件 config/config.yaml 不存在${NC}"
    exit 1
fi

echo -e "${GREEN}[1/6] 启动用户服务 (HTTP:8080, gRPC:50051)...${NC}"
cd user-service
nohup go run main.go grpc_server.go -config=../config/user-service.yaml > logs/service.log 2>&1 &
USER_PID=$!
echo -e "${GREEN}    用户服务已启动 (PID: $USER_PID)${NC}"
cd ..
sleep 3

echo -e "${GREEN}[2/6] 启动会议服务 (HTTP:8082, gRPC:50052)...${NC}"
cd meeting-service
nohup go run main.go grpc_server.go -config=../config/meeting-service.yaml > logs/service.log 2>&1 &
MEETING_PID=$!
echo -e "${GREEN}    会议服务已启动 (PID: $MEETING_PID)${NC}"
cd ..
sleep 35  # 等待ZMQ连接超时

echo -e "${GREEN}[3/6] 启动信令服务 (HTTP:8081)...${NC}"
cd signaling-service
nohup go run main.go -config=../config/signaling-service.yaml > logs/service.log 2>&1 &
SIGNALING_PID=$!
echo -e "${GREEN}    信令服务已启动 (PID: $SIGNALING_PID)${NC}"
cd ..
sleep 35  # 等待ZMQ连接超时

echo -e "${GREEN}[4/6] 启动媒体服务 (HTTP:8083)...${NC}"
cd media-service
nohup go run main.go -config=config/media-service.yaml > logs/service.log 2>&1 &
MEDIA_PID=$!
echo -e "${GREEN}    媒体服务已启动 (PID: $MEDIA_PID)${NC}"
cd ..
sleep 3

echo -e "${GREEN}[5/6] 启动AI推理服务 (HTTP:8085, gRPC:9085)...${NC}"
cd ai-inference-service
nohup go run . -config=config/ai-inference-service-local.yaml > logs/service.log 2>&1 &
AI_PID=$!
echo -e "${GREEN}    AI推理服务已启动 (PID: $AI_PID)${NC}"
cd ..
sleep 35  # 等待ZMQ连接超时

echo -e "${GREEN}[6/6] 检查服务状态...${NC}"
echo ""

# 检查用户服务
if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 用户服务 HTTP (8080) - 运行正常${NC}"
else
    echo -e "${RED}❌ 用户服务 HTTP (8080) - 未响应${NC}"
fi

# 检查会议服务
if curl -f -s http://localhost:8082/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 会议服务 HTTP (8082) - 运行正常${NC}"
else
    echo -e "${RED}❌ 会议服务 HTTP (8082) - 未响应${NC}"
fi

# 检查信令服务
if curl -f -s http://localhost:8081/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 信令服务 HTTP (8081) - 运行正常${NC}"
else
    echo -e "${RED}❌ 信令服务 HTTP (8081) - 未响应${NC}"
fi

# 检查媒体服务
if curl -f -s http://localhost:8083/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 媒体服务 HTTP (8083) - 运行正常${NC}"
else
    echo -e "${RED}❌ 媒体服务 HTTP (8083) - 未响应${NC}"
fi

# 检查AI服务
if curl -f -s http://localhost:8085/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ AI推理服务 HTTP (8085) - 运行正常${NC}"
else
    echo -e "${RED}❌ AI推理服务 HTTP (8085) - 未响应${NC}"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  所有服务已启动完成${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "可用的服务端点："
echo "  - 用户服务 HTTP:  http://localhost:8080"
echo "  - 用户服务 gRPC:  localhost:50051"
echo "  - 会议服务 HTTP:  http://localhost:8082"
echo "  - 会议服务 gRPC:  localhost:50052"
echo "  - 信令服务 HTTP:  http://localhost:8081"
echo "  - 媒体服务 HTTP:  http://localhost:8083"
echo "  - AI推理服务 HTTP: http://localhost:8085"
echo ""
echo "健康检查: /health"
echo "Prometheus指标: /metrics"
echo ""
echo -e "${YELLOW}按 Ctrl+C 停止所有服务${NC}"
echo ""

# 保持脚本运行
while true; do
    sleep 1
done
