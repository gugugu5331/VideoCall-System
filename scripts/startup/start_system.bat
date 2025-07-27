@echo off
chcp 65001 >nul
title VideoCall System - 一键启动

echo ==========================================
echo VideoCall System - 一键启动脚本
echo ==========================================
echo.

:: 检查Python环境
echo [1/6] Checking Python environment...
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Python not installed or not in PATH
    pause
    exit /b 1
)
echo OK: Python environment ready

:: 检查Docker环境
echo.
echo [2/6] Checking Docker environment...
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Docker not installed or not in PATH
    pause
    exit /b 1
)
echo OK: Docker environment ready

:: 启动数据库服务
echo.
echo [3/6] Starting database services...
docker-compose --project-name videocall-system -f config/docker-compose.yml up -d postgres redis
if %errorlevel% neq 0 (
    echo ERROR: Database services failed to start
    pause
    exit /b 1
)
echo OK: Database services started successfully

:: 等待数据库启动
echo.
echo [4/6] Waiting for database services to be ready...
timeout /t 5 /nobreak >nul
echo OK: Database services ready

:: 启动后端服务
echo.
echo [5/6] Starting backend service...
start "Backend Service" cmd /k "cd /d %~dp0\..\..\core\backend && start-full.bat"
if %errorlevel% neq 0 (
    echo ERROR: Backend service failed to start
    pause
    exit /b 1
)
echo OK: Backend service starting...

:: 等待后端启动
echo.
echo Waiting for backend service to be ready...
timeout /t 8 /nobreak >nul

:: 启动AI服务
echo.
echo [6/6] Starting AI service...
start "AI Service" cmd /k "cd /d %~dp0\..\..\core\ai-service && start_ai_manual.bat"
if %errorlevel% neq 0 (
    echo ERROR: AI service failed to start
    pause
    exit /b 1
)
echo OK: AI service starting...

:: 等待AI服务启动
echo.
echo Waiting for AI service to be ready...
timeout /t 5 /nobreak >nul

:: 运行系统测试
echo.
echo ==========================================
echo Running system tests...
echo ==========================================
python %~dp0\..\testing\run_all_tests.py

echo.
echo ==========================================
echo System startup completed!
echo ==========================================
echo.
echo Service status:
echo - Backend service: http://localhost:8000
echo - AI service: http://localhost:5001
echo - Database: PostgreSQL (5432), Redis (6379)
echo.
echo Management commands:
echo - Test system: python %~dp0\..\testing\run_all_tests.py
echo - Check database: python check_database.py
echo - Stop services: Close corresponding command windows
echo.
echo Press any key to exit...
pause >nul 