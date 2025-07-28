@echo off
setlocal enabledelayedexpansion

echo ========================================
echo FFmpeg服务快速开始脚本 (Windows)
echo ========================================

:: 设置变量
set PROJECT_DIR=%~dp0
set VCPKG_DIR=%PROJECT_DIR%..\..\core\frontend\vcpkg

echo 步骤 1: 检查环境...
if not exist "%VCPKG_DIR%\vcpkg.exe" (
    echo ❌ vcpkg未找到，请先安装vcpkg
    echo 请运行: git clone https://github.com/Microsoft/vcpkg.git %VCPKG_DIR%
    echo 然后运行: %VCPKG_DIR%\bootstrap-vcpkg.bat
    pause
    exit /b 1
)

echo ✅ vcpkg已找到

echo.
echo 步骤 2: 准备环境...
call setup_environment.bat
if %ERRORLEVEL% neq 0 (
    echo ❌ 环境准备失败
    pause
    exit /b 1
)

echo.
echo 步骤 3: 编译项目...
call build.bat
if %ERRORLEVEL% neq 0 (
    echo ❌ 编译失败
    pause
    exit /b 1
)

echo.
echo 步骤 4: 运行测试...
python test_basic_functionality.py
if %ERRORLEVEL% neq 0 (
    echo ❌ 测试失败
    pause
    exit /b 1
)

echo.
echo 步骤 5: 集成到项目...
python integrate_with_project.py
if %ERRORLEVEL% neq 0 (
    echo ❌ 集成失败
    pause
    exit /b 1
)

echo.
echo ========================================
echo 🎉 FFmpeg服务快速开始完成！
echo ========================================
echo.
echo 已完成的步骤:
echo ✅ 环境准备 - 安装FFmpeg、OpenCV、ONNX Runtime等依赖
echo ✅ 项目编译 - 构建C++库和示例程序
echo ✅ 功能测试 - 验证基本功能正常
echo ✅ 项目集成 - 集成到Python AI服务、Go后端、WebRTC前端
echo.
echo 下一步:
echo 1. 查看 README.md 了解详细使用方法
echo 2. 运行示例程序: build\bin\ffmpeg_service_example.exe
echo 3. 在您的项目中使用集成接口
echo.
pause 