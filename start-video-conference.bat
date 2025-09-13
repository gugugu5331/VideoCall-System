@echo off
chcp 65001 >nul
echo.
echo ========================================
echo   🎥 视频会议系统完整版启动
echo ========================================
echo.

echo 🔍 检查环境...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go未安装或未配置环境变量
    echo 请先安装Go 1.21+: https://golang.org/dl/
    pause
    exit /b 1
)

echo ✅ Go环境检查通过

echo.
echo 📦 安装后端依赖...
cd demo
go mod tidy
if %errorlevel% neq 0 (
    echo ❌ 后端依赖安装失败
    pause
    exit /b 1
)

echo ✅ 后端依赖安装完成

echo.
echo 🚀 启动视频会议系统...
echo.
echo 系统组件:
echo   📡 后端API服务: http://localhost:8080
echo   💬 WebSocket信令: ws://localhost:8080/signaling
echo   🎥 视频会议界面: file:///%~dp0web-client\index.html
echo   🧪 API测试页面: file:///%~dp0demo\test.html
echo.
echo 功能特性:
echo   ✅ 多人视频会议 (WebRTC)
echo   ✅ 实时音视频传输
echo   ✅ 屏幕共享
echo   ✅ 文字聊天
echo   ✅ AI伪造检测
echo   ✅ 会议录制
echo.
echo 使用说明:
echo   1. 后端服务将在后台启动
echo   2. 浏览器将自动打开会议界面
echo   3. 输入用户名和会议ID加入会议
echo   4. 允许浏览器访问摄像头和麦克风
echo   5. 可以多开浏览器窗口模拟多用户
echo.
echo 按 Ctrl+C 停止服务
echo.

REM 启动后端服务
start /B go run main.go

REM 等待服务启动
timeout /t 3 >nul

REM 打开会议界面
start "" "file:///%~dp0web-client\index.html"

REM 等待一下再打开测试页面
timeout /t 2 >nul
start "" "file:///%~dp0demo\test.html"

REM 等待用户输入
echo 🎉 系统启动完成！
echo.
echo 💡 提示:
echo   - 可以在多个浏览器标签页中打开会议界面测试多用户功能
echo   - 使用不同的用户名加入同一个会议ID
echo   - 支持摄像头、麦克风、屏幕共享等功能
echo   - AI检测功能会实时分析音视频内容
echo.
pause
