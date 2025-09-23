@echo off
echo 启动智能视频通话系统 - Web前端
echo =================================

REM 切换到web_interface目录
cd /d "%~dp0"

echo 当前目录: %CD%
echo 正在启动HTTP服务器...
echo 前端地址: http://localhost:8081
echo 按 Ctrl+C 停止服务器
echo.

REM 启动Python HTTP服务器
python -m http.server 8081

pause 