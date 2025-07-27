@echo off
chcp 65001 >nul
title VideoCall System - Stop All Services

echo ==========================================
echo VideoCall System - Stop All Services
echo ==========================================
echo.

:: 检查管理员权限
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo WARNING: This script may need administrator privileges to kill processes
    echo Some processes may not be stopped properly
    echo.
)

:: 1. 停止Docker容器
echo [1/5] Stopping Docker containers...
docker-compose --project-name videocall-system down
if %errorlevel% equ 0 (
    echo OK: Docker containers stopped
) else (
    echo WARNING: Failed to stop Docker containers
)
echo.

:: 2. 查找并停止后端服务进程
echo [2/5] Stopping backend service (port 8000)...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :8000 ^| findstr LISTENING') do (
    echo Found process %%a on port 8000
    taskkill /f /pid %%a >nul 2>&1
    if errorlevel equ 0 (
        echo OK: Backend service stopped (PID: %%a)
    ) else (
        echo WARNING: Failed to stop process %%a
    )
)
echo.

:: 3. 查找并停止AI服务进程
echo [3/5] Stopping AI service (port 5001)...
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :5001 ^| findstr LISTENING') do (
    echo Found process %%a on port 5001
    taskkill /f /pid %%a >nul 2>&1
    if errorlevel equ 0 (
        echo OK: AI service stopped (PID: %%a)
    ) else (
        echo WARNING: Failed to stop process %%a
    )
)
echo.

:: 4. 查找并停止Python进程（AI服务相关）
echo [4/5] Stopping Python processes...
tasklist /fi "imagename eq python.exe" /fo csv | findstr /i "python" >nul
if %errorlevel% equ 0 (
    echo Found Python processes, stopping them...
    taskkill /f /im python.exe >nul 2>&1
    if errorlevel equ 0 (
        echo OK: Python processes stopped
    ) else (
        echo WARNING: Failed to stop some Python processes
    )
) else (
    echo OK: No Python processes found
)
echo.

:: 5. 查找并停止Go进程（后端服务相关）
echo [5/5] Stopping Go processes...
tasklist /fi "imagename eq *.exe" /fo csv | findstr /i "videocall" >nul
if %errorlevel% equ 0 (
    echo Found Go processes, stopping them...
    taskkill /f /im *.exe /fi "WINDOWTITLE eq *videocall*" >nul 2>&1
    if errorlevel equ 0 (
        echo OK: Go processes stopped
    ) else (
        echo WARNING: Failed to stop some Go processes
    )
) else (
    echo OK: No Go processes found
)
echo.

:: 等待进程完全停止
echo Waiting for processes to fully stop...
timeout /t 3 /nobreak >nul

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
echo If any ports are still in use, you may need to:
echo 1. Run this script as administrator
echo 2. Manually close the applications using those ports
echo 3. Restart your computer if necessary
echo.
pause 