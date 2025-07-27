@echo off
chcp 65001 >nul
echo ==========================================
echo 启动AI服务
echo ==========================================

echo 检查Python环境...
python --version
if %errorlevel% neq 0 (
    echo ❌ Python未安装或不在PATH中
    pause
    exit /b 1
)

echo.
echo 检查AI服务目录...
if not exist "ai-service" (
    echo ❌ AI服务目录不存在
    pause
    exit /b 1
)

echo.
echo 切换到AI服务目录...
cd ai-service

echo.
echo 检查依赖包...
python -c "import fastapi, uvicorn, redis, numpy" 2>nul
if %errorlevel% neq 0 (
    echo 安装AI服务依赖包...
    pip install -r requirements.txt
    if %errorlevel% neq 0 (
        echo ❌ 依赖包安装失败
        pause
        exit /b 1
    )
)

echo.
echo 启动AI服务...
echo 服务将在 http://localhost:5000 启动
echo 按 Ctrl+C 停止服务
echo.

python main.py

echo.
echo AI服务已停止
pause 