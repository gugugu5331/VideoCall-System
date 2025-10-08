#!/bin/bash

# 完整集成测试运行脚本

set -e

echo "=========================================="
echo "会议系统完整集成测试"
echo "=========================================="

# 检查Python环境
if ! command -v python3 &> /dev/null; then
    echo "错误: 未找到 python3"
    exit 1
fi

# 检查必要的Python包
echo "检查Python依赖..."
python3 -c "import requests" 2>/dev/null || {
    echo "安装 requests 包..."
    pip3 install requests
}

# 创建日志目录
mkdir -p logs

# 检查Docker服务
echo "检查Docker服务状态..."
if ! docker ps &> /dev/null; then
    echo "错误: Docker服务未运行"
    exit 1
fi

# 检查必要的容器
echo "检查必要的容器..."
required_containers=(
    "meeting-system-postgres"
    "meeting-system-user-service"
    "meeting-system-meeting-service"
    "meeting-system-signaling-service"
    "meeting-system-media-service"
    "meeting-system-ai-service"
    "meeting-system-nginx"
)

for container in "${required_containers[@]}"; do
    if ! docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
        echo "警告: 容器 ${container} 未运行"
    fi
done

# 检查测试视频目录
TEST_VIDEO_DIR="/root/meeting-system-server/meeting-system/backend/media-service/test_video"
if [ ! -d "$TEST_VIDEO_DIR" ]; then
    echo "错误: 测试视频目录不存在: $TEST_VIDEO_DIR"
    exit 1
fi

echo "测试视频文件:"
ls -lh "$TEST_VIDEO_DIR"

# 运行测试
echo ""
echo "=========================================="
echo "开始运行集成测试..."
echo "=========================================="
echo ""

cd "$(dirname "$0")"
python3 complete_integration_test.py

# 检查测试结果
if [ $? -eq 0 ]; then
    echo ""
    echo "=========================================="
    echo "✅ 测试成功完成！"
    echo "=========================================="
    exit 0
else
    echo ""
    echo "=========================================="
    echo "❌ 测试失败！"
    echo "=========================================="
    echo "请查看日志文件: logs/integration_test.log"
    exit 1
fi

