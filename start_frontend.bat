@echo off
chcp 65001 >nul
title 智能视频通话系统 - 前端启动

echo ==========================================
echo 智能视频通话系统 - 前端启动
echo ==========================================
echo.

echo 检查服务状态...
echo.

echo [1/3] 检查后端服务...
curl -s http://localhost:8000/health >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ 后端服务运行正常
) else (
    echo ✗ 后端服务未运行，请先启动后端服务
    echo   运行: scripts/startup/start_system.bat
    pause
    exit /b 1
)

echo.
echo [2/3] 检查AI服务...
curl -s http://localhost:5001/health >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ AI服务运行正常
) else (
    echo ⚠ AI服务未运行，部分功能可能受限
)

echo.
echo [3/3] 启动前端服务...
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
echo ==========================================
echo 前端服务启动成功！
echo ==========================================
echo.
echo 访问地址:
echo - 主页面: http://localhost:8081
echo - 测试页面: http://localhost:8081/test-simple.html
echo.
echo 服务状态:
echo - 后端服务: http://localhost:8000/health
echo - AI服务: http://localhost:5001/health
echo - 前端服务: http://localhost:8081
echo.
echo 按 Ctrl+C 停止前端服务
echo.

cd web_interface
python -m http.server 8081

echo.
echo 前端服务已停止
pause 