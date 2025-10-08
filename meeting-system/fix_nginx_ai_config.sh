#!/bin/bash

echo "=========================================="
echo "修复 Nginx AI 服务配置"
echo "=========================================="
echo ""

# 备份当前配置
echo "1. 备份当前 Nginx 配置..."
docker exec meeting-nginx cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup
echo "✓ 备份完成"
echo ""

# 在 /api/v1/speech location 中添加 client_max_body_size
echo "2. 更新 Nginx 配置..."
docker exec meeting-nginx bash -c 'sed -i "/location \/api\/v1\/speech {/a\            client_max_body_size 100M;" /etc/nginx/nginx.conf'
echo "✓ 配置已更新"
echo ""

# 验证配置
echo "3. 验证 Nginx 配置..."
if docker exec meeting-nginx nginx -t 2>&1 | grep -q "successful"; then
    echo "✓ 配置验证成功"
else
    echo "✗ 配置验证失败，恢复备份..."
    docker exec meeting-nginx cp /etc/nginx/nginx.conf.backup /etc/nginx/nginx.conf
    exit 1
fi
echo ""

# 重新加载 Nginx
echo "4. 重新加载 Nginx..."
docker exec meeting-nginx nginx -s reload
echo "✓ Nginx 已重新加载"
echo ""

# 验证更新
echo "5. 验证更新后的配置..."
docker exec meeting-nginx cat /etc/nginx/nginx.conf | grep -A 5 "location /api/v1/speech" | grep "client_max_body_size"
echo ""

echo "=========================================="
echo "✓ Nginx 配置修复完成！"
echo "=========================================="

