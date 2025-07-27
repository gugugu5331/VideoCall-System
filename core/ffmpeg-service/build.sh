#!/bin/bash

echo "========================================"
echo "FFmpeg 伪造检测服务编译脚本"
echo "========================================"

# 设置环境变量
FFMPEG_SERVICE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$FFMPEG_SERVICE_DIR/build"
INSTALL_DIR="$FFMPEG_SERVICE_DIR/install"

# 检查依赖
echo "检查依赖..."

# 检查CMake
if ! command -v cmake &> /dev/null; then
    echo "错误: 找不到CMake，请先安装CMake"
    exit 1
fi

# 检查FFmpeg
if ! pkg-config --exists libavcodec libavformat libavutil libswscale libswresample; then
    echo "错误: 找不到FFmpeg库，请先安装FFmpeg开发包"
    echo "Ubuntu/Debian: sudo apt install ffmpeg libavcodec-dev libavformat-dev libavutil-dev libswscale-dev libswresample-dev"
    exit 1
fi

# 检查ONNX Runtime
if [ ! -f "/usr/local/include/onnxruntime_cxx_api.h" ] && [ ! -f "/usr/include/onnxruntime/onnxruntime_cxx_api.h" ]; then
    echo "警告: 找不到ONNX Runtime头文件"
    echo "请确保ONNX Runtime已正确安装"
fi

# 创建构建目录
if [ ! -d "$BUILD_DIR" ]; then
    mkdir -p "$BUILD_DIR"
fi

# 进入构建目录
cd "$BUILD_DIR"

echo "配置CMake项目..."

# 配置项目
cmake .. \
    -DCMAKE_INSTALL_PREFIX="$INSTALL_DIR" \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_CXX_FLAGS="-O3 -march=native" \
    -DCMAKE_EXE_LINKER_FLAGS="-Wl,-rpath,/usr/local/lib"

if [ $? -ne 0 ]; then
    echo "CMake配置失败"
    exit 1
fi

echo "编译项目..."

# 编译项目
make -j$(nproc)

if [ $? -ne 0 ]; then
    echo "编译失败"
    exit 1
fi

echo "安装项目..."

# 安装项目
make install

if [ $? -ne 0 ]; then
    echo "安装失败"
    exit 1
fi

echo "编译完成！"
echo "可执行文件位置: $INSTALL_DIR/bin/ffmpeg_detection_service"

# 复制配置文件
if [ ! -d "$INSTALL_DIR/config" ]; then
    mkdir -p "$INSTALL_DIR/config"
fi
cp "$FFMPEG_SERVICE_DIR/config.json" "$INSTALL_DIR/config/"

# 创建模型目录
if [ ! -d "$INSTALL_DIR/models" ]; then
    mkdir -p "$INSTALL_DIR/models"
fi

# 设置可执行权限
chmod +x "$INSTALL_DIR/bin/ffmpeg_detection_service"
chmod +x "$FFMPEG_SERVICE_DIR/start_service.sh"

echo ""
echo "请将ONNX模型文件放置在 $INSTALL_DIR/models 目录下"
echo "然后运行 ./start_service.sh 启动服务"

# 显示依赖库信息
echo ""
echo "依赖库信息:"
ldd "$INSTALL_DIR/bin/ffmpeg_detection_service" | grep -E "(libav|libsw|libonnx)" || echo "未找到相关依赖库" 