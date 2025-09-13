@echo off
chcp 65001 >nul
echo.
echo ========================================
echo   视频会议系统演示版启动脚本
echo ========================================
echo.

echo 🔍 检查Go环境...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go未安装或未配置环境变量
    echo 请先安装Go 1.21+: https://golang.org/dl/
    pause
    exit /b 1
)

echo ✅ Go环境检查通过

echo.
echo 📦 安装依赖包...
cd demo
go mod tidy
if %errorlevel% neq 0 (
    echo ❌ 依赖包安装失败
    pause
    exit /b 1
)

echo ✅ 依赖包安装完成

echo.
echo 🚀 启动视频会议系统演示版...
echo.
echo 服务将在以下地址启动:
echo   📍 主页: http://localhost:8080
echo   📖 API: http://localhost:8080/api/v1
echo   🔍 健康检查: http://localhost:8080/health
echo   💬 WebSocket: ws://localhost:8080/signaling
echo   🧪 测试页面: file:///%~dp0demo\test.html
echo.
echo 按 Ctrl+C 停止服务
echo.

start "" "http://localhost:8080"
timeout /t 2 >nul
start "" "file:///%~dp0demo\test.html"

go run main.go

echo.
echo 服务已停止
pause
