#!/bin/bash

# 修复并启动所有服务

set -e

echo "=========================================="
echo "清理并重新启动会议系统服务"
echo "=========================================="

cd /root/meeting-system-server/meeting-system

# 停止并删除所有服务容器
echo "停止并删除旧容器..."
docker-compose down

# 修复配置文件权限
echo "修复配置文件权限..."
chmod -R 777 backend/config
chmod -R 777 backend/logs

# 启动基础设施服务
echo "启动基础设施服务..."
docker-compose up -d postgres redis etcd mongodb jaeger minio

# 等待基础设施服务就绪
echo "等待基础设施服务就绪..."
sleep 15

# 启动应用服务
echo "启动应用服务..."
docker-compose up -d user-service meeting-service signaling-service media-service ai-service

# 等待应用服务启动
echo "等待应用服务启动..."
sleep 10

# 检查服务状态
echo ""
echo "=========================================="
echo "服务状态:"
echo "=========================================="
docker-compose ps

echo ""
echo "=========================================="
echo "检查服务日志:"
echo "=========================================="
echo "用户服务日志:"
docker-compose logs --tail=15 user-service

echo ""
echo "会议服务日志:"
docker-compose logs --tail=15 meeting-service

echo ""
echo "信令服务日志:"
docker-compose logs --tail=15 signaling-service

echo ""
echo "媒体服务日志:"
docker-compose logs --tail=15 media-service

echo ""
echo "AI服务日志:"
docker-compose logs --tail=15 ai-service

echo ""
echo "=========================================="
echo "✅ 服务启动完成！"
echo "=========================================="

