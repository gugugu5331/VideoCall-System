@echo off
chcp 65001 >nul
title 启动前端服务

echo ================================================================
echo 🌐 启动前端服务
echo ================================================================
echo.

cd /d "%~dp0..\web_interface"

echo 📁 当前目录: %CD%
echo.

if exist "server.py" (
    echo ✅ 找到前端服务器: server.py
    echo 🌐 正在启动前端服务...
    echo.
    python server.py
) else (
    echo ❌ 找不到前端服务器文件
    echo 请确保 server.py 文件存在
    echo.
    pause
) 