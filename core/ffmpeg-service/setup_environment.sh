#!/bin/bash

# FFmpeg服务环境准备脚本 (Linux/macOS)

set -e  # 遇到错误时退出

echo "========================================"
echo "FFmpeg服务环境准备脚本 (Linux/macOS)"
echo "========================================"

# 设置变量
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VCPKG_DIR="$PROJECT_DIR/../../frontend/vcpkg"
VCPKG_EXE="$VCPKG_DIR/vcpkg"

echo "检查vcpkg..."
if [ ! -f "$VCPKG_EXE" ]; then
    echo "❌ vcpkg未找到，请确保vcpkg已正确安装"
    echo "请运行: git clone https://github.com/Microsoft/vcpkg.git"
    echo "然后运行: $VCPKG_DIR/bootstrap-vcpkg.sh"
    exit 1
fi

echo "✅ vcpkg已找到: $VCPKG_EXE"

# 检测操作系统
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    TRIPLET="x64-linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    TRIPLET="x64-osx"
else
    echo "❌ 不支持的操作系统: $OSTYPE"
    exit 1
fi

echo "检测到操作系统: $OSTYPE, 使用triplet: $TRIPLET"

# 安装FFmpeg
echo ""
echo "正在安装FFmpeg..."
"$VCPKG_EXE" install "ffmpeg:$TRIPLET"
if [ $? -ne 0 ]; then
    echo "❌ FFmpeg安装失败"
    exit 1
fi
echo "✅ FFmpeg安装成功"

# 安装OpenCV
echo ""
echo "正在安装OpenCV..."
"$VCPKG_EXE" install "opencv4:$TRIPLET"
if [ $? -ne 0 ]; then
    echo "❌ OpenCV安装失败"
    exit 1
fi
echo "✅ OpenCV安装成功"

# 安装ONNX Runtime
echo ""
echo "正在安装ONNX Runtime..."
"$VCPKG_EXE" install "onnxruntime:$TRIPLET"
if [ $? -ne 0 ]; then
    echo "❌ ONNX Runtime安装失败"
    exit 1
fi
echo "✅ ONNX Runtime安装成功"

# 安装其他依赖
echo ""
echo "正在安装其他依赖..."
"$VCPKG_EXE" install "nlohmann-json:$TRIPLET"
"$VCPKG_EXE" install "spdlog:$TRIPLET"
"$VCPKG_EXE" install "fmt:$TRIPLET"

echo ""
echo "========================================"
echo "环境准备完成！"
echo "========================================"
echo ""
echo "已安装的包:"
echo "- FFmpeg (音视频处理)"
echo "- OpenCV (图像处理)"
echo "- ONNX Runtime (深度学习推理)"
echo "- nlohmann-json (JSON处理)"
echo "- spdlog (日志记录)"
echo "- fmt (格式化输出)"
echo ""
echo "下一步: 运行 ./build.sh 编译项目"
echo "" 