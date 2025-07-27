#!/bin/bash

echo "=========================================="
echo "音视频通话系统 - 本地开发模式启动"
echo "=========================================="

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "错误: Docker未运行，请启动Docker Desktop"
    exit 1
fi

echo "正在启动数据库服务..."

# 只启动数据库和Redis
docker-compose --project-name videocall-system up -d postgres redis

echo "等待数据库启动..."
sleep 10

echo "检查数据库状态..."
docker-compose --project-name videocall-system ps

echo ""
echo "=========================================="
echo "数据库服务启动完成！"
echo "=========================================="
echo ""
echo "现在您可以："
echo "1. 手动启动后端服务"
echo "2. 手动启动AI服务"
echo "3. 或者运行完整启动: docker-compose up -d"
echo ""
echo "数据库连接信息:"
echo "- PostgreSQL: localhost:5432"
echo "- Redis: localhost:6379"
echo ""
echo "查看日志: docker-compose logs -f"
echo "停止服务: docker-compose down"
echo "==========================================" 