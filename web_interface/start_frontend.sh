#!/bin/bash

echo "启动智能视频通话系统 - Web前端"
echo "================================"

# 检查Python是否安装
if ! command -v python3 &> /dev/null; then
    echo "错误: 未找到Python3，请先安装Python3"
    echo "Ubuntu/Debian: sudo apt install python3"
    echo "CentOS/RHEL: sudo yum install python3"
    echo "macOS: brew install python3"
    exit 1
fi

echo "正在启动本地服务器..."
echo "前端地址: http://localhost:8081"
echo "按 Ctrl+C 停止服务器"
echo

# 启动Python HTTP服务器
python3 -m http.server 8081 