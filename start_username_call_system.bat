@echo off
chcp 65001 >nul
echo ========================================
echo 启动基于用户名的通话系统
echo ========================================

:: 检查Go环境
echo 检查Go环境...
go version >nul 2>&1
if errorlevel 1 (
    echo ❌ Go未安装或未配置到PATH
    echo 请先安装Go: https://golang.org/dl/
    pause
    exit /b 1
)

:: 检查Node.js环境
echo 检查Node.js环境...
node --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Node.js未安装或未配置到PATH
    echo 请先安装Node.js: https://nodejs.org/
    pause
    exit /b 1
)

:: 检查Python环境
echo 检查Python环境...
python --version >nul 2>&1
if errorlevel 1 (
    echo ❌ Python未安装或未配置到PATH
    echo 请先安装Python: https://python.org/
    pause
    exit /b 1
)

echo ✅ 环境检查通过

:: 启动后端服务
echo.
echo 🚀 启动后端服务...
cd core\backend
start "后端服务" cmd /k "go run main.go"
cd ..\..

:: 等待后端服务启动
echo 等待后端服务启动...
timeout /t 5 /nobreak >nul

:: 启动前端服务
echo.
echo 🌐 启动前端服务...
cd web_interface
start "前端服务" cmd /k "python -m http.server 3000"
cd ..

:: 等待前端服务启动
echo 等待前端服务启动...
timeout /t 3 /nobreak >nul

:: 运行测试
echo.
echo 🧪 运行功能测试...
python test_username_call.py

echo.
echo ========================================
echo 系统启动完成！
echo ========================================
echo.
echo 📱 前端界面: http://localhost:3000
echo 🔧 后端API: http://localhost:8000
echo 📚 API文档: http://localhost:8000/swagger/index.html
echo.
echo 💡 使用说明:
echo 1. 打开浏览器访问 http://localhost:3000
echo 2. 注册或登录用户账户
echo 3. 在通话页面搜索用户
echo 4. 点击"通话"按钮发起视频通话
echo.
echo 🧪 运行测试: python test_username_call.py
echo.
pause 