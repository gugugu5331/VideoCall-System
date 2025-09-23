#!/bin/bash

# FFmpeg服务 + ONNX检测器构建脚本

set -e  # 遇到错误时退出

echo "========================================"
echo "FFmpeg服务 + ONNX检测器构建脚本"
echo "========================================"

# 设置变量
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$PROJECT_DIR/build"
VCPKG_DIR="$PROJECT_DIR/../../core/frontend/vcpkg"
CMAKE_TOOLCHAIN_FILE="$VCPKG_DIR/scripts/buildsystems/vcpkg.cmake"

# 检查vcpkg是否存在
if [ ! -d "$VCPKG_DIR" ]; then
    echo "错误: 找不到vcpkg目录: $VCPKG_DIR"
    echo "请确保vcpkg已正确安装"
    exit 1
fi

# 检查CMake是否存在
if ! command -v cmake &> /dev/null; then
    echo "错误: 找不到CMake，请先安装CMake"
    exit 1
fi

# 检查编译器
if command -v gcc &> /dev/null; then
    echo "使用GCC编译器"
    export CC=gcc
    export CXX=g++
elif command -v clang &> /dev/null; then
    echo "使用Clang编译器"
    export CC=clang
    export CXX=clang++
else
    echo "错误: 找不到C++编译器"
    exit 1
fi

# 创建构建目录
if [ ! -d "$BUILD_DIR" ]; then
    echo "创建构建目录: $BUILD_DIR"
    mkdir -p "$BUILD_DIR"
fi

# 进入构建目录
cd "$BUILD_DIR"

# 配置CMake
echo ""
echo "配置CMake..."
cmake .. \
    -DCMAKE_TOOLCHAIN_FILE="$CMAKE_TOOLCHAIN_FILE" \
    -DCMAKE_BUILD_TYPE=Release \
    -DCMAKE_CXX_FLAGS="-O3 -march=native"

if [ $? -ne 0 ]; then
    echo "错误: CMake配置失败"
    exit 1
fi

# 获取CPU核心数
if command -v nproc &> /dev/null; then
    CORES=$(nproc)
else
    CORES=4
fi

# 编译项目
echo ""
echo "编译项目 (使用 $CORES 个核心)..."
cmake --build . --config Release --parallel $CORES

if [ $? -ne 0 ]; then
    echo "错误: 编译失败"
    exit 1
fi

# 运行测试（如果存在）
if [ -f "bin/ffmpeg_service_test" ]; then
    echo ""
    echo "运行测试..."
    ./bin/ffmpeg_service_test
    if [ $? -ne 0 ]; then
        echo "警告: 测试失败"
    else
        echo "测试通过"
    fi
fi

# 运行示例（如果存在）
if [ -f "bin/ffmpeg_service_example" ]; then
    echo ""
    echo "运行示例程序..."
    ./bin/ffmpeg_service_example
fi

echo ""
echo "========================================"
echo "构建完成！"
echo "========================================"
echo "构建目录: $BUILD_DIR"
echo "可执行文件位置: $BUILD_DIR/bin/"
echo "库文件位置: $BUILD_DIR/lib/"
echo "========================================"

# 显示构建结果
echo ""
echo "构建结果:"
ls -la bin/ 2>/dev/null || echo "没有找到可执行文件"
ls -la lib/ 2>/dev/null || echo "没有找到库文件" 