@echo off
chcp 65001 >nul
title 前端服务启动

echo ==========================================
echo 前端服务启动脚本
echo ==========================================
echo.

echo 检查服务状态...
python check_system_status.py

echo.
echo 启动前端服务...
echo 前端地址: http://localhost:3000
echo.

cd web_interface

echo 使用Python启动简单的HTTP服务器...
python -m http.server 3000

echo.
echo 前端服务已停止
pause 