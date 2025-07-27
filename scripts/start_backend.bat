@echo off
chcp 65001 >nul
title 启动后端服务

echo ================================================================
echo 🚀 启动后端服务
echo ================================================================
echo.

cd /d "%~dp0..\core\backend"

echo 📁 当前目录: %CD%
echo.

if exist "enhanced-backend.go" (
    echo ✅ 找到增强版后端: enhanced-backend.go
    echo 🚀 正在启动后端服务...
    echo.
    go run enhanced-backend.go
) else (
    echo ❌ 找不到后端服务文件
    echo 请确保 enhanced-backend.go 文件存在
    echo.
    pause
) 