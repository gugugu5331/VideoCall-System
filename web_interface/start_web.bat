@echo off
chcp 65001 >nul
title 音视频通话系统 - Web界面

echo.
echo ========================================
echo    音视频通话系统 - Web界面启动
echo ========================================
echo.

echo 检查Python环境...
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Python未安装或未添加到PATH
    echo 请先安装Python 3.7+
    pause
    exit /b 1
)
echo ✅ Python环境正常

echo.
echo 检查Web界面文件...
if not exist index.html (
    echo ❌ 找不到index.html文件
    echo 请确保在正确的目录中运行此脚本
    pause
    exit /b 1
)
echo ✅ Web界面文件存在

echo.
echo 检查后端服务...
curl -s http://localhost:8000/health >nul 2>&1
if %errorlevel% neq 0 (
    echo ⚠️  后端服务可能未运行
    echo 请确保后端服务在 http://localhost:8000 运行
    echo.
    echo 是否继续启动Web界面？ (Y/N)
    set /p continue=
    if /i not "%continue%"=="Y" (
        echo 已取消启动
        pause
        exit /b 1
    )
) else (
    echo ✅ 后端服务正常
)

echo.
echo 检查AI服务...
curl -s http://localhost:5001/health >nul 2>&1
if %errorlevel% neq 0 (
    echo ⚠️  AI服务可能未运行
    echo 请确保AI服务在 http://localhost:5001 运行
    echo.
    echo 是否继续启动Web界面？ (Y/N)
    set /p continue=
    if /i not "%continue%"=="Y" (
        echo 已取消启动
        pause
        exit /b 1
    )
) else (
    echo ✅ AI服务正常
)

echo.
echo 启动Web服务器...
echo.
echo 访问地址: http://localhost:8080
echo 按 Ctrl+C 停止服务器
echo.

python server.py

echo.
echo Web服务器已停止
pause 