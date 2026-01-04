#!/bin/bash

echo "========================================="
echo "AI服务验证脚本"
echo "========================================="
echo ""

# 1. 检查AI服务容器状态
echo "1. 检查AI服务容器状态..."
docker ps --filter "name=meeting-ai-service" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""

# 2. 检查AI服务健康状态
echo "2. 检查AI服务健康状态..."
curl -s http://localhost:8800/api/v1/health | python3 -m json.tool 2>/dev/null || echo "健康检查失败"
echo ""

# 3. 获取AI模型列表
echo "3. 获取AI模型列表..."
curl -s http://localhost:8800/api/v1/models | python3 -m json.tool 2>/dev/null || echo "获取模型列表失败"
echo ""

# 4. 测试音频降噪接口
echo "4. 测试音频降噪接口..."
echo '{"audio_data":"dGVzdCBhdWRpbyBkYXRh","audio_format":"wav"}' | \
curl -s -X POST http://localhost:8800/api/v1/audio/denoising \
  -H "Content-Type: application/json" \
  -d @- | python3 -m json.tool 2>/dev/null || echo "音频降噪测试失败"
echo ""

# 5. 检查模型目录
echo "5. 检查模型目录..."
if [ -d "/models" ]; then
    echo "模型目录存在："
    du -sh /models/* 2>/dev/null || echo "模型目录为空"
else
    echo "⚠ 模型目录不存在"
fi
echo ""

# 6. 查看AI服务日志（最后20行）
echo "6. AI服务日志（最后20行）..."
docker logs meeting-ai-service --tail 20 2>&1
echo ""

echo "========================================="
echo "验证完成"
echo "========================================="
