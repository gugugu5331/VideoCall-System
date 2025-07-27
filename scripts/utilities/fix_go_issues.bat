@echo off
chcp 65001 >nul
title VideoCall System - Go Issues Fix

echo ==========================================
echo VideoCall System - Go问题快速修复
echo ==========================================
echo.

echo [1/4] 检查Go环境...
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go未安装或不在PATH中
    echo.
    echo 📥 请下载并安装Go:
    echo    1. 访问 https://golang.org/dl/
    echo    2. 下载Windows版本 (推荐Go 1.21或更高版本)
    echo    3. 运行安装程序
    echo    4. 重启命令行窗口
    echo.
    pause
    exit /b 1
)

echo ✅ Go已安装
for /f "tokens=*" %%i in ('go version') do echo   版本: %%i

echo.
echo [2/4] 清理Go模块缓存...
cd core\backend
go clean -modcache
go mod tidy
go mod download

if %errorlevel% neq 0 (
    echo ❌ Go模块清理失败
    pause
    exit /b 1
)

echo ✅ Go模块清理完成

echo.
echo [3/4] 测试基础版本编译...
go build -o test-compile.exe main-basic.go

if %errorlevel% neq 0 (
    echo ❌ 基础版本编译失败
    echo 请检查Go版本和依赖
    pause
    exit /b 1
)

echo ✅ 基础版本编译成功
del test-compile.exe

echo.
echo [4/4] 修复完成！
echo.
echo 🎉 所有Go问题已修复
echo.
echo 📋 现在您可以:
echo 1. 运行 .\core\backend\start-basic.bat 启动基础后端
echo 2. 运行 .\quick_start.bat 一键启动所有服务
echo 3. 运行 .\quick_manage.bat 使用管理菜单
echo.
echo 💡 如果仍有问题，请参考 docs/guides/TROUBLESHOOTING.md
echo.
pause 