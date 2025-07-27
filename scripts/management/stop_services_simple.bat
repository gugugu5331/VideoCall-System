@echo off
chcp 65001 >nul
title VideoCall System - Stop Services Simple

echo ==========================================
echo VideoCall System - Stop Services Simple
echo ==========================================
echo.

:: 1. 停止Docker容器
echo [1/3] Stopping Docker containers...
docker-compose --project-name videocall-system down
if %errorlevel% equ 0 (
    echo OK: Docker containers stopped
) else (
    echo WARNING: Failed to stop Docker containers
)
echo.

:: 2. 停止Python进程（AI服务）
echo [2/3] Stopping Python processes...
taskkill /f /im python.exe >nul 2>&1
if %errorlevel% equ 0 (
    echo OK: Python processes stopped
) else (
    echo INFO: No Python processes found or already stopped
)
echo.

:: 3. 停止Go进程（后端服务）
echo [3/3] Stopping Go processes...
taskkill /f /im *.exe /fi "WINDOWTITLE eq *Backend*" >nul 2>&1
taskkill /f /im *.exe /fi "WINDOWTITLE eq *AI*" >nul 2>&1
echo OK: Service processes stopped
echo.

:: 等待进程完全停止
echo Waiting for processes to fully stop...
timeout /t 2 /nobreak >nul

:: 检查端口状态
echo ==========================================
echo Checking port status...
echo ==========================================

echo Checking port 8000 (Backend)...
netstat -ano | findstr :8000 >nul
if %errorlevel% equ 0 (
    echo WARNING: Port 8000 is still in use
) else (
    echo OK: Port 8000 is free
)

echo Checking port 5001 (AI Service)...
netstat -ano | findstr :5001 >nul
if %errorlevel% equ 0 (
    echo WARNING: Port 5001 is still in use
) else (
    echo OK: Port 5001 is free
)

echo Checking port 5432 (PostgreSQL)...
netstat -ano | findstr :5432 >nul
if %errorlevel% equ 0 (
    echo WARNING: Port 5432 is still in use
) else (
    echo OK: Port 5432 is free
)

echo Checking port 6379 (Redis)...
netstat -ano | findstr :6379 >nul
if %errorlevel% equ 0 (
    echo WARNING: Port 6379 is still in use
) else (
    echo OK: Port 6379 is free
)

echo.
echo ==========================================
echo Service stop completed!
echo ==========================================
echo.
echo If any ports are still in use, run:
echo   python release_ports.py
echo.
pause 