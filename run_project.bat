@echo off
chcp 65001 >nul
echo ========================================
echo 基于用户名的通话系统
echo ========================================

echo.
echo 🚀 启动后端服务...
cd core\backend
start "后端服务" cmd /k "go run main.go"
cd ..\..

echo.
echo 🌐 启动前端服务...
cd web_interface
start "前端服务" cmd /k "python -m http.server 3000"
cd ..

echo.
echo ⏳ 等待服务启动...
timeout /t 5 /nobreak >nul

echo.
echo ========================================
echo ✅ 系统启动完成！
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
echo 🧪 测试功能:
echo python test_username_call.py
echo.
pause 