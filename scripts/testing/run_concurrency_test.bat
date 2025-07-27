@echo off
chcp 65001 >nul
title VideoCall System - Concurrency Test

echo ==========================================
echo VideoCall System - 并发性能测试
echo ==========================================
echo.

:: 检查Python环境
echo [1/4] Checking Python environment...
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Python not installed or not in PATH
    pause
    exit /b 1
)
echo OK: Python environment ready

:: 检查依赖
echo.
echo [2/4] Checking dependencies...
python -c "import aiohttp, asyncio, statistics" >nul 2>&1
if %errorlevel% neq 0 (
    echo Installing required dependencies...
    pip install aiohttp asyncio-throttle
    if %errorlevel% neq 0 (
        echo ERROR: Failed to install dependencies
        pause
        exit /b 1
    )
)
echo OK: Dependencies ready

:: 检查服务状态
echo.
echo [3/4] Checking service status...
python -c "import requests; requests.get('http://localhost:8000/health', timeout=5)" >nul 2>&1
if %errorlevel% neq 0 (
    echo WARNING: Backend service may not be running
    echo Please ensure backend service is started on port 8000
)

python -c "import requests; requests.get('http://localhost:5001/health', timeout=5)" >nul 2>&1
if %errorlevel% neq 0 (
    echo WARNING: AI service may not be running
    echo Please ensure AI service is started on port 5001
)

echo OK: Service check completed

:: 运行并发测试
echo.
echo [4/4] Running concurrency tests...
echo.

:: 设置测试参数
set /p TEST_TYPE="选择测试类型 (health/detection/batch): "
if "%TEST_TYPE%"=="" set TEST_TYPE=health

set /p NUM_REQUESTS="并发请求数量 (默认100): "
if "%NUM_REQUESTS%"=="" set NUM_REQUESTS=100

echo.
echo Starting concurrency test...
echo Type: %TEST_TYPE%
echo Requests: %NUM_REQUESTS%
echo.

:: 运行测试
python test_concurrency.py --type %TEST_TYPE% --requests %NUM_REQUESTS%

if %errorlevel% equ 0 (
    echo.
    echo ==========================================
    echo Concurrency test completed successfully!
    echo ==========================================
) else (
    echo.
    echo ==========================================
    echo Concurrency test failed!
    echo ==========================================
)

echo.
echo Press any key to continue...
pause >nul 