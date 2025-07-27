@echo off
chcp 65001 >nul
title VideoCall System - Go Environment Check

echo ==========================================
echo VideoCall System - Go环境检查
echo ==========================================
echo.

echo [1/3] 检查Go是否安装...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go未安装或不在PATH中
    echo.
    echo 📥 请下载并安装Go:
    echo    1. 访问 https://golang.org/dl/
    echo    2. 下载Windows版本 (推荐Go 1.21或更高版本)
    echo    3. 运行安装程序
    echo    4. 重启命令行窗口
    echo.
    echo 🔧 安装完成后，请重新运行此脚本
    pause
    exit /b 1
)

echo ✅ Go已安装
for /f "tokens=*" %%i in ('go version') do echo   版本: %%i

echo.
echo [2/3] 检查Go环境变量...
echo GOPATH: %GOPATH%
echo GOROOT: %GOROOT%
echo.

if "%GOPATH%"=="" (
    echo ⚠️  GOPATH未设置，使用默认路径
) else (
    echo ✅ GOPATH已设置
)

if "%GOROOT%"=="" (
    echo ⚠️  GOROOT未设置，使用默认路径
) else (
    echo ✅ GOROOT已设置
)

echo.
echo [3/3] 检查Go模块支持...
go env GOMOD >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go模块支持有问题
    echo.
    echo 🔧 请确保:
    echo    1. Go版本 >= 1.11
    echo    2. 在项目目录中运行
    echo    3. 存在go.mod文件
) else (
    echo ✅ Go模块支持正常
)

echo.
echo ==========================================
echo 检查完成
echo ==========================================
echo.
echo 如果所有检查都通过，您可以:
echo 1. 运行 .\core\backend\start-simple.bat 启动简化后端
echo 2. 运行 .\core\ai-service\start_ai_manual.bat 启动AI服务
echo 3. 运行 .\quick_start.bat 一键启动所有服务
echo.
pause 