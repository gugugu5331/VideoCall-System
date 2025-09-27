#!/bin/bash

# 简单的OpenCV编译脚本
echo "编译简化视频处理测试程序..."

# 检查是否安装了OpenCV
if ! pkg-config --exists opencv4; then
    if ! pkg-config --exists opencv; then
        echo "错误: 未找到OpenCV库"
        echo "请安装OpenCV开发包:"
        echo "  Ubuntu/Debian: sudo apt-get install libopencv-dev"
        echo "  CentOS/RHEL: sudo yum install opencv-devel"
        echo "  macOS: brew install opencv"
        exit 1
    else
        OPENCV_PKG="opencv"
    fi
else
    OPENCV_PKG="opencv4"
fi

echo "找到OpenCV包: $OPENCV_PKG"

# 编译命令
g++ -std=c++17 \
    simple_test.cpp \
    -o simple_video_test \
    $(pkg-config --cflags --libs $OPENCV_PKG) \
    -O3 \
    -Wall \
    -Wextra

if [ $? -eq 0 ]; then
    echo "编译成功!"
    echo "运行程序: ./simple_video_test"
    
    # 检查是否有摄像头设备
    if ls /dev/video* 1> /dev/null 2>&1; then
        echo "检测到摄像头设备"
    else
        echo "警告: 未检测到摄像头设备"
    fi
    
else
    echo "编译失败!"
    exit 1
fi
