#!/bin/bash

echo "=========================================="
echo "AI Service 路由发现测试"
echo "=========================================="
echo ""

# 测试各种可能的路由
routes=(
    "/health"
    "/api/health"
    "/api/v1/health"
    "/api/v1/ai/health"
    "/api/v1/speech/recognize"
    "/api/v1/speech/emotion"
    "/api/v1/speech/synthesis-detection"
    "/api/v1/ai/speech/recognize"
    "/api/v1/ai/emotion"
    "/api/v1/ai/synthesis"
    "/speech/recognize"
    "/emotion/detect"
    "/synthesis/detect"
    "/ai/speech"
    "/ai/emotion"
    "/ai/synthesis"
)

echo "测试 AI Service (ai-service:8084) 的路由..."
echo ""

for route in "${routes[@]}"; do
    echo -n "Testing $route ... "
    response=$(docker exec meeting-nginx curl -s -o /dev/null -w "%{http_code}" http://ai-service:8084$route 2>&1)
    if [ "$response" = "200" ] || [ "$response" = "405" ] || [ "$response" = "400" ]; then
        echo "✓ Found! (HTTP $response)"
    elif [ "$response" = "404" ]; then
        echo "✗ Not Found"
    else
        echo "? Unknown ($response)"
    fi
done

echo ""
echo "=========================================="
echo "测试 POST 请求"
echo "=========================================="
echo ""

post_routes=(
    "/api/v1/speech/recognize"
    "/api/v1/speech/emotion"
    "/api/v1/ai/speech/recognize"
    "/speech/recognize"
    "/ai/speech"
)

for route in "${post_routes[@]}"; do
    echo -n "Testing POST $route ... "
    response=$(docker exec meeting-nginx curl -s -o /dev/null -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d '{"test":"data"}' \
        http://ai-service:8084$route 2>&1)
    if [ "$response" = "200" ] || [ "$response" = "400" ] || [ "$response" = "422" ]; then
        echo "✓ Found! (HTTP $response)"
    elif [ "$response" = "404" ]; then
        echo "✗ Not Found"
    else
        echo "? Unknown ($response)"
    fi
done

echo ""
echo "完成！"

