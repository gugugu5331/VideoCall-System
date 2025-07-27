@echo off
chcp 65001 >nul
echo ========================================
echo 快速测试基于用户名的通话功能
echo ========================================

echo.
echo 🔍 检查服务状态...

echo 检查后端服务...
curl -s http://localhost:8000/health >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ 后端服务正常运行
) else (
    echo ❌ 后端服务未运行，请先启动系统
    pause
    exit /b 1
)

echo 检查前端服务...
curl -s http://localhost:3000 >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ 前端服务正常运行
) else (
    echo ❌ 前端服务未运行，请先启动系统
    pause
    exit /b 1
)

echo.
echo 🧪 运行功能测试...
python test_username_call.py

echo.
echo 🎯 运行功能演示...
python simple_username_demo.py

echo.
echo ========================================
echo ✅ 测试完成！
echo ========================================
echo.
echo 📱 访问系统: http://localhost:3000
echo.
pause
