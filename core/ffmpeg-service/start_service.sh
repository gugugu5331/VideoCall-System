#!/bin/bash

echo "========================================"
echo "FFmpeg 伪造检测服务启动脚本"
echo "========================================"

# 设置环境变量
FFMPEG_SERVICE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export PATH="$FFMPEG_SERVICE_DIR/bin:$PATH"
export LD_LIBRARY_PATH="$FFMPEG_SERVICE_DIR/lib:/usr/local/lib:$LD_LIBRARY_PATH"

# 检查可执行文件是否存在
if [ ! -f "$FFMPEG_SERVICE_DIR/bin/ffmpeg_detection_service" ]; then
    echo "错误: 找不到可执行文件 ffmpeg_detection_service"
    echo "请先编译项目"
    exit 1
fi

# 检查模型文件
if [ ! -f "$FFMPEG_SERVICE_DIR/models/detection.onnx" ]; then
    echo "警告: 找不到模型文件 models/detection.onnx"
    echo "请将ONNX模型文件放置在models目录下"
fi

# 创建日志目录
if [ ! -d "$FFMPEG_SERVICE_DIR/logs" ]; then
    mkdir -p "$FFMPEG_SERVICE_DIR/logs"
fi

# 检查依赖库
echo "检查依赖库..."
ldd "$FFMPEG_SERVICE_DIR/bin/ffmpeg_detection_service" | grep -E "(not found|=>)" | grep -v "linux-vdso.so.1"

echo "启动FFmpeg检测服务..."
echo "使用 Ctrl+C 停止服务"

# 启动服务
"$FFMPEG_SERVICE_DIR/bin/ffmpeg_detection_service" \
    -i "rtsp://localhost:8554/stream" \
    -m "$FFMPEG_SERVICE_DIR/models/detection.onnx" \
    -c "$FFMPEG_SERVICE_DIR/config.json" \
    -o "$FFMPEG_SERVICE_DIR/logs/service.log" \
    -v 