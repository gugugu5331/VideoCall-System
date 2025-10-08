# 构建阶段
FROM golang:1.24-alpine AS builder

# 安装必要的工具
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /build

# 复制整个 backend 目录（包含 shared 和 ai-inference-service）
COPY . .

# 进入 ai-inference-service 目录
WORKDIR /build/ai-inference-service

# 下载依赖
RUN go mod download

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ai-inference-service .

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates（用于 HTTPS 请求）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件和配置文件
COPY --from=builder /build/ai-inference-service/ai-inference-service .
COPY --from=builder /build/ai-inference-service/config ./config

# 创建日志目录
RUN mkdir -p logs && chown -R appuser:appuser /app

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 8085

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8085/health || exit 1

# 启动应用
CMD ["./ai-inference-service"]

