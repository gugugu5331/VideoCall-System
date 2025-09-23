@echo off
echo 启动智能视频通话系统 - Web前端
echo =================================

REM 检查Python是否安装
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到Python，请先安装Python
    echo 下载地址: https://www.python.org/downloads/
    pause
    exit /b 1
)

echo 正在启动本地服务器...
echo 前端地址: http://localhost:8081
echo 按 Ctrl+C 停止服务器
echo.

REM 启动Python HTTP服务器
python -m http.server 8081

pause 