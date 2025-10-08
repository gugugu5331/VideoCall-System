# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置 Alpine 镜像源（使用阿里云镜像）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置工作目录
WORKDIR /app

# 安装必要的包（包括 ZMQ 依赖）
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev zeromq-dev

# 设置Go代理和工具链（使用多个镜像源）
ENV GOPROXY=https://goproxy.cn,https://mirrors.aliyun.com/goproxy/,https://goproxy.io,direct
ENV GOTOOLCHAIN=auto
ENV GO111MODULE=on

# 复制项目源代码
COPY . .

# 预下载依赖
RUN go mod download

# 构建应用（启用 CGO 以支持 ZMQ）
WORKDIR /app/signaling-service
RUN CGO_ENABLED=1 GOOS=linux go build -a -o signaling-service .

# 运行阶段
FROM alpine:latest

# 设置 Alpine 镜像源（使用阿里云镜像）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装运行时依赖（包括 ZMQ 运行时库）
RUN apk --no-cache add ca-certificates tzdata curl zeromq

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/signaling-service/signaling-service .

# 创建配置和日志目录
RUN mkdir -p config logs && \
    chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
CMD ["./signaling-service", "-config", "config/signaling-service.yaml"]
