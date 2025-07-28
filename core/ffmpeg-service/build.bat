@echo off
setlocal enabledelayedexpansion

echo ========================================
echo FFmpeg服务 + ONNX检测器构建脚本
echo ========================================

:: 设置变量
set PROJECT_DIR=%~dp0
set BUILD_DIR=%PROJECT_DIR%build
set VCPKG_DIR=%PROJECT_DIR%..\..\core\frontend\vcpkg
set CMAKE_TOOLCHAIN_FILE=%VCPKG_DIR%\scripts\buildsystems\vcpkg.cmake

:: 检查vcpkg是否存在
if not exist "%VCPKG_DIR%" (
    echo 错误: 找不到vcpkg目录: %VCPKG_DIR%
    echo 请确保vcpkg已正确安装
    pause
    exit /b 1
)

:: 检查CMake是否存在
cmake --version >nul 2>&1
if errorlevel 1 (
    echo 错误: 找不到CMake，请先安装CMake
    pause
    exit /b 1
)

:: 创建构建目录
if not exist "%BUILD_DIR%" (
    echo 创建构建目录: %BUILD_DIR%
    mkdir "%BUILD_DIR%"
)

:: 进入构建目录
cd /d "%BUILD_DIR%"

:: 配置CMake
echo.
echo 配置CMake...
cmake .. -DCMAKE_TOOLCHAIN_FILE="%CMAKE_TOOLCHAIN_FILE%" -DCMAKE_BUILD_TYPE=Release
if errorlevel 1 (
    echo 错误: CMake配置失败
    pause
    exit /b 1
)

:: 编译项目
echo.
echo 编译项目...
cmake --build . --config Release --parallel
if errorlevel 1 (
    echo 错误: 编译失败
    pause
    exit /b 1
)

:: 运行测试（如果存在）
if exist "bin\ffmpeg_service_test.exe" (
    echo.
    echo 运行测试...
    bin\ffmpeg_service_test.exe
    if errorlevel 1 (
        echo 警告: 测试失败
    ) else (
        echo 测试通过
    )
)

:: 运行示例（如果存在）
if exist "bin\ffmpeg_service_example.exe" (
    echo.
    echo 运行示例程序...
    bin\ffmpeg_service_example.exe
)

echo.
echo ========================================
echo 构建完成！
echo ========================================
echo 构建目录: %BUILD_DIR%
echo 可执行文件位置: %BUILD_DIR%\bin\
echo 库文件位置: %BUILD_DIR%\lib\
echo ========================================

pause 