#!/bin/bash

# Qt6客户端构建测试脚本

set -e

echo "================================"
echo "Qt6客户端构建测试"
echo "================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查依赖
echo -e "${YELLOW}检查依赖...${NC}"

# 检查Qt6
if ! command -v qmake6 &> /dev/null && ! command -v qmake &> /dev/null; then
    echo -e "${RED}错误: 未找到Qt6${NC}"
    echo "请安装Qt6: https://www.qt.io/download"
    exit 1
fi
echo -e "${GREEN}✓ Qt6已安装${NC}"

# 检查CMake
if ! command -v cmake &> /dev/null; then
    echo -e "${RED}错误: 未找到CMake${NC}"
    echo "请安装CMake 3.16+"
    exit 1
fi
CMAKE_VERSION=$(cmake --version | head -n1 | cut -d' ' -f3)
echo -e "${GREEN}✓ CMake $CMAKE_VERSION已安装${NC}"

# 检查编译器
if ! command -v g++ &> /dev/null && ! command -v clang++ &> /dev/null; then
    echo -e "${RED}错误: 未找到C++编译器${NC}"
    exit 1
fi
echo -e "${GREEN}✓ C++编译器已安装${NC}"

# 创建构建目录
BUILD_DIR="build_test"
if [ -d "$BUILD_DIR" ]; then
    echo -e "${YELLOW}清理旧的构建目录...${NC}"
    rm -rf "$BUILD_DIR"
fi

echo -e "${YELLOW}创建构建目录...${NC}"
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# 配置
echo -e "${YELLOW}配置项目...${NC}"
if cmake .. -DCMAKE_BUILD_TYPE=Debug; then
    echo -e "${GREEN}✓ 配置成功${NC}"
else
    echo -e "${RED}✗ 配置失败${NC}"
    exit 1
fi

# 构建
echo -e "${YELLOW}构建项目...${NC}"
if cmake --build . --config Debug -j$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 4); then
    echo -e "${GREEN}✓ 构建成功${NC}"
else
    echo -e "${RED}✗ 构建失败${NC}"
    exit 1
fi

# 检查可执行文件
if [ -f "bin/MeetingSystemClient" ] || [ -f "bin/MeetingSystemClient.exe" ]; then
    echo -e "${GREEN}✓ 可执行文件生成成功${NC}"
else
    echo -e "${RED}✗ 未找到可执行文件${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}构建测试通过！${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo "可执行文件位置: $BUILD_DIR/bin/MeetingSystemClient"
echo ""
echo "运行应用:"
echo "  cd $BUILD_DIR/bin"
echo "  ./MeetingSystemClient"

