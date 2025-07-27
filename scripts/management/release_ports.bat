@echo off
chcp 65001 >nul
title VideoCall System - Release Ports

echo ==========================================
echo VideoCall System - Release Ports
echo ==========================================
echo.

:: 检查Python是否可用
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ERROR: Python not found
    echo Please install Python or add it to PATH
    pause
    exit /b 1
)

:: 检查psutil是否安装
python -c "import psutil" >nul 2>&1
if %errorlevel% neq 0 (
    echo Installing psutil...
    pip install psutil
    if %errorlevel% neq 0 (
        echo ERROR: Failed to install psutil
        echo Please run: pip install psutil
        pause
        exit /b 1
    )
)

:: 检查命令行参数
set FORCE_MODE=
set SPECIFIC_PORTS=

:parse_args
if "%1"=="" goto run_script
if "%1"=="--force" (
    set FORCE_MODE=--force
    shift
    goto parse_args
)
if "%1"=="-f" (
    set FORCE_MODE=-f
    shift
    goto parse_args
)
if "%1"=="--help" (
    echo Usage: release_ports.bat [options] [port1] [port2] ...
    echo.
    echo Options:
    echo   --force, -f    Force kill processes without asking
    echo   --help         Show this help message
    echo.
    echo Examples:
    echo   release_ports.bat              # Check all default ports
    echo   release_ports.bat --force      # Force kill all processes
    echo   release_ports.bat 8000 5001    # Check specific ports
    echo   release_ports.bat -f 8000      # Force kill process on port 8000
    echo.
    pause
    exit /b 0
)
if "%1"=="" goto run_script

:: 检查是否为数字（端口号）
echo %1| findstr /r "^[0-9]*$" >nul
if %errorlevel% equ 0 (
    set SPECIFIC_PORTS=%SPECIFIC_PORTS% %1
    shift
    goto parse_args
)

:run_script
echo Starting port release utility...
echo.

if defined FORCE_MODE (
    echo Force mode enabled
)
if defined SPECIFIC_PORTS (
    echo Checking specific ports:%SPECIFIC_PORTS%
)

:: 运行Python脚本
python release_ports.py %FORCE_MODE% %SPECIFIC_PORTS%

echo.
echo Port release completed!
pause 