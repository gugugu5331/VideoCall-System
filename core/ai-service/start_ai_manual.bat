@echo off
chcp 65001 >nul
echo ==========================================
echo 手动启动AI服务
echo ==========================================

echo Checking Python environment...
python --version
if %errorlevel% neq 0 (
    echo ERROR: Python not installed or not in PATH
    pause
    exit /b 1
)

echo.
echo Checking AI service directory...
if not exist "main-simple.py" (
    echo ERROR: main-simple.py not found
    pause
    exit /b 1
)

echo.
echo Checking simplified AI service file...
if not exist "main-simple.py" (
    echo ERROR: main-simple.py not found
    pause
    exit /b 1
)

echo.
echo Starting simplified AI service...
echo Service will start at http://localhost:5001
echo Press Ctrl+C to stop service
echo.

python main-simple.py

echo.
echo AI service stopped
pause 