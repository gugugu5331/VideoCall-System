@echo off
chcp 65001 >nul
title 简化系统启动

echo ==========================================
echo 简化系统启动脚本
echo ==========================================
echo.

echo 检查Docker是否运行...
docker version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker未运行或未安装
    echo 请启动Docker Desktop
    pause
    exit /b 1
)

echo.
echo 启动数据库和Redis服务...
cd config
docker-compose up -d postgres redis
if %errorlevel% neq 0 (
    echo ERROR: 启动数据库服务失败
    pause
    exit /b 1
)

echo.
echo 等待数据库启动...
timeout /t 10 /nobreak >nul

echo.
echo 启动后端服务...
cd ..\core\backend
start "后端服务" cmd /k "go run main.go"

echo.
echo 等待后端服务启动...
timeout /t 5 /nobreak >nul

echo.
echo 启动AI服务...
cd ..\ai-service
start "AI服务" cmd /k "python main.py"

echo.
echo 系统启动完成！
echo 后端服务: http://localhost:8000
echo AI服务: http://localhost:5000
echo.
echo 按任意键打开测试页面...
pause >nul

start http://localhost:8000 