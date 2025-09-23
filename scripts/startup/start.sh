#!/bin/bash

# 音视频通话系统启动脚本

echo "=========================================="
echo "音视频通话系统 - 第一阶段基础架构"
echo "=========================================="

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "错误: Docker未安装，请先安装Docker"
    exit 1
fi

# 检查Docker Compose是否安装
if ! command -v docker-compose &> /dev/null; then
    echo "错误: Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

echo "正在启动服务..."

# 构建并启动所有服务
docker-compose up --build -d

echo "等待服务启动..."

# 等待服务启动
sleep 10

# 检查服务状态
echo "检查服务状态..."
docker-compose ps

echo ""
echo "=========================================="
echo "服务启动完成！"
echo "=========================================="
echo ""
echo "访问地址:"
echo "- 后端API: http://localhost:8000"
echo "- AI服务: http://localhost:5000"
echo "- API文档: http://localhost:8000/swagger/index.html"
echo "- 健康检查: http://localhost:8000/health"
echo ""
echo "数据库:"
echo "- PostgreSQL: localhost:5432"
echo "- Redis: localhost:6379"
echo ""
echo "查看日志: docker-compose logs -f"
echo "停止服务: docker-compose down"
echo "==========================================" 