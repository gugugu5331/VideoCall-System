#!/bin/bash

# Jaeger 依赖安装脚本

set -e

echo "🔍 Installing Jaeger tracing dependencies..."

# 进入 shared 目录
cd "$(dirname "$0")/../shared"

echo "📦 Installing OpenTracing and Jaeger client..."
go get github.com/opentracing/opentracing-go@v1.2.0
go get github.com/uber/jaeger-client-go@v2.30.0+incompatible
go get github.com/uber/jaeger-lib@v2.4.1+incompatible

echo "🔄 Tidying go.mod..."
go mod tidy

echo "✅ Jaeger dependencies installed successfully!"
echo ""
echo "📝 Next steps:"
echo "1. Start Jaeger: docker-compose up -d jaeger"
echo "2. Access Jaeger UI: http://localhost:16686"
echo "3. Start your services with Jaeger enabled"

