@echo off
chcp 65001 >nul
echo ========================================
echo 启动真正的通话系统
echo ========================================
echo.

:: 检查Go环境
echo [1/5] 检查Go环境...
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go未安装或未配置PATH
    echo 请先安装Go: https://golang.org/dl/
    pause
    exit /b 1
)
echo ✅ Go环境正常

:: 检查Node.js环境
echo [2/5] 检查Node.js环境...
node --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Node.js未安装或未配置PATH
    echo 请先安装Node.js: https://nodejs.org/
    pause
    exit /b 1
)
echo ✅ Node.js环境正常

:: 检查Python环境
echo [3/5] 检查Python环境...
python --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Python未安装或未配置PATH
    echo 请先安装Python: https://python.org/
    pause
    exit /b 1
)
echo ✅ Python环境正常

:: 设置环境变量
echo [4/5] 设置环境变量...
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
echo ✅ 环境变量设置完成

:: 启动后端服务
echo [5/5] 启动后端服务...
echo.
echo 🚀 正在启动WebRTC信令服务器...
echo 📡 端口: 8000
echo 🔌 WebSocket: ws://localhost:8000/ws/call/
echo.

cd core\backend
start "WebRTC Backend" cmd /k "go run main.go"

:: 等待后端启动
echo ⏳ 等待后端服务启动...
timeout /t 5 /nobreak >nul

:: 启动前端服务
echo.
echo 🌐 正在启动前端界面...
echo 📱 地址: http://localhost:3000
echo.

cd ..\..\web_interface
start "Frontend" cmd /k "python -m http.server 3000"

:: 等待前端启动
echo ⏳ 等待前端服务启动...
timeout /t 3 /nobreak >nul

:: 打开浏览器
echo.
echo 🌍 正在打开浏览器...
start http://localhost:3000

echo.
echo ========================================
echo ✅ 真正的通话系统启动完成！
echo ========================================
echo.
echo 📞 功能特性:
echo   • 真正的WebRTC P2P连接
echo   • 实时音视频通话
echo   • 信令服务器
echo   • 通话房间管理
echo   • 安全检测
echo   • 通话历史记录
echo.
echo 🔧 测试步骤:
echo   1. 在浏览器中注册/登录用户
echo   2. 点击"开始通话"按钮
echo   3. 允许摄像头和麦克风权限
echo   4. 测试音视频通话功能
echo.
echo 📡 API端点:
echo   • 后端API: http://localhost:8000
echo   • WebSocket: ws://localhost:8000/ws/call/
echo   • 前端界面: http://localhost:3000
echo.
echo 🧪 运行测试:
echo   python test_real_call.py
echo.
echo 按任意键退出...
pause >nul 