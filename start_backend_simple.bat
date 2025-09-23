@echo off
chcp 65001 >nul
title 后端服务启动

echo ==========================================
echo 后端服务启动脚本
echo ==========================================
echo.

echo 检查Go环境...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Go环境未安装或未配置
    echo 请安装Go语言环境: https://golang.org/dl/
    pause
    exit /b 1
)

echo.
echo 检查后端目录...
if not exist "core\backend\main.go" (
    echo ERROR: 后端主文件不存在
    pause
    exit /b 1
)

echo.
echo 启动后端服务...
echo 服务地址: http://localhost:8000
echo 按 Ctrl+C 停止服务
echo.

cd core\backend
go run main.go

echo.
echo 后端服务已停止
pause 