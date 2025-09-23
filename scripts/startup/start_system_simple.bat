@echo off
chcp 65001 >nul
title VideoCall System - 快速启动

echo ==========================================
echo VideoCall System - 快速启动脚本
echo ==========================================
echo.

:: 启动数据库服务
echo [1/3] Starting database services...
docker-compose --project-name videocall-system -f deployment/docker/docker-compose.yml up -d postgres redis
if %errorlevel% neq 0 (
    echo ERROR: Database services failed to start
    pause
    exit /b 1
)
echo OK: Database services started successfully

:: 等待数据库启动
echo.
echo Waiting for database services to be ready...
timeout /t 3 /nobreak >nul

:: 启动后端服务
echo.
echo [2/3] Starting backend service...
start "Backend Service" cmd /k "cd /d %~dp0\..\..\core\backend && start-basic.bat"
echo OK: Backend service starting...

:: 启动AI服务
echo.
echo [3/3] Starting AI service...
start "AI Service" cmd /k "cd /d %~dp0\..\..\core\ai-service && start_ai_manual.bat"
echo OK: AI service starting...

echo.
echo ==========================================
echo System startup completed!
echo ==========================================
echo.
echo Service addresses:
echo - Backend service: http://localhost:8000
echo - AI service: http://localhost:5001
echo.
echo Test commands:
echo - Full test: python %~dp0\..\..\tests\api\run_all_tests.py
echo - Quick test: python %~dp0\..\..\tests\api\test_api.py
echo.
echo Press any key to exit...
pause >nul 