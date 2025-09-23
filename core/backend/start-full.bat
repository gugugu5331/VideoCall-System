@echo off
chcp 65001 >nul
echo ==========================================
echo 启动音视频通话系统完整后端服务
echo ==========================================

cd backend

echo 编译后端服务...
go build -o videocall-backend.exe .

if %errorlevel% neq 0 (
    echo Compilation failed!
    pause
    exit /b 1
)

echo Compilation successful!
echo.

echo 设置环境变量...
set DB_HOST=localhost
set DB_PORT=5432
set DB_NAME=videocall
set DB_USER=admin
set DB_PASSWORD=videocall123
set REDIS_HOST=localhost
set REDIS_PORT=6379
set JWT_SECRET=your-secret-key-here-change-in-production
set JWT_EXPIRE_HOURS=24
set PORT=8000
set GIN_MODE=debug

echo 启动完整后端服务...
.\videocall-backend.exe

pause 