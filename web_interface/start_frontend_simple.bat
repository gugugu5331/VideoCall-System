@echo off
chcp 65001 >nul
title 前端服务启动

echo ==========================================
echo 前端服务启动脚本
echo ==========================================
echo.

echo 检查端口8081是否被占用...
netstat -ano | findstr :8081 >nul
if %errorlevel% equ 0 (
    echo 端口8081已被占用，正在停止现有服务...
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8081') do (
        taskkill /f /pid %%a >nul 2>&1
    )
    timeout /t 2 /nobreak >nul
)

echo.
echo 启动前端服务...
echo 服务地址: http://localhost:8081
echo 测试页面: http://localhost:8081/test-simple.html
echo.
echo 按 Ctrl+C 停止服务
echo.

python -m http.server 8081

echo.
echo 前端服务已停止
pause 