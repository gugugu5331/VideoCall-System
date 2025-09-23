#!/bin/bash

# FFmpeg服务快速开始脚本 (Linux/macOS)

set -e  # 遇到错误时退出

echo "========================================"
echo "FFmpeg服务快速开始脚本 (Linux/macOS)"
echo "========================================"

# 设置变量
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VCPKG_DIR="$PROJECT_DIR/../../frontend/vcpkg"

echo "步骤 1: 检查环境..."
if [ ! -f "$VCPKG_DIR/vcpkg" ]; then
    echo "❌ vcpkg未找到，请先安装vcpkg"
    echo "请运行: git clone https://github.com/Microsoft/vcpkg.git $VCPKG_DIR"
    echo "然后运行: $VCPKG_DIR/bootstrap-vcpkg.sh"
    exit 1
fi

echo "✅ vcpkg已找到"

echo ""
echo "步骤 2: 准备环境..."
./setup_environment.sh
if [ $? -ne 0 ]; then
    echo "❌ 环境准备失败"
    exit 1
fi

echo ""
echo "步骤 3: 编译项目..."
./build.sh
if [ $? -ne 0 ]; then
    echo "❌ 编译失败"
    exit 1
fi

echo ""
echo "步骤 4: 运行测试..."
python3 test_basic_functionality.py
if [ $? -ne 0 ]; then
    echo "❌ 测试失败"
    exit 1
fi

echo ""
echo "步骤 5: 集成到项目..."
python3 integrate_with_project.py
if [ $? -ne 0 ]; then
    echo "❌ 集成失败"
    exit 1
fi

echo ""
echo "========================================"
echo "🎉 FFmpeg服务快速开始完成！"
echo "========================================"
echo ""
echo "已完成的步骤:"
echo "✅ 环境准备 - 安装FFmpeg、OpenCV、ONNX Runtime等依赖"
echo "✅ 项目编译 - 构建C++库和示例程序"
echo "✅ 功能测试 - 验证基本功能正常"
echo "✅ 项目集成 - 集成到Python AI服务、Go后端、WebRTC前端"
echo ""
echo "下一步:"
echo "1. 查看 README.md 了解详细使用方法"
echo "2. 运行示例程序: ./build/bin/ffmpeg_service_example"
echo "3. 在您的项目中使用集成接口"
echo "" 