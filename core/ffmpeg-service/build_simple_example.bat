@echo off
setlocal enabledelayedexpansion

echo ========================================
echo FFmpeg服务简单示例构建脚本
echo ========================================

:: 设置变量
set PROJECT_DIR=%~dp0
set BUILD_DIR=%PROJECT_DIR%build_simple
set VCPKG_DIR=%PROJECT_DIR%..\..\core\frontend\vcpkg
set CMAKE_TOOLCHAIN_FILE=%VCPKG_DIR%\scripts\buildsystems\vcpkg.cmake

:: 检查vcpkg是否存在
if not exist "%VCPKG_DIR%" (
    echo 错误: 找不到vcpkg目录: %VCPKG_DIR%
    echo 请确保vcpkg已正确安装
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

:: 尝试使用Visual Studio的CMake
set CMAKE_PATH="C:\Program Files\Microsoft Visual Studio\2022\Community\Common7\IDE\CommonExtensions\Microsoft\CMake\CMake\bin\cmake.exe"

if exist %CMAKE_PATH% (
    echo 使用Visual Studio CMake: %CMAKE_PATH%
    set CMAKE_CMD=%CMAKE_PATH%
) else (
    echo 错误: 找不到CMake，请安装CMake或Visual Studio
    echo 下载地址: https://cmake.org/download/
    pause
    exit /b 1
)

:: 复制简化的CMakeLists.txt
copy ..\CMakeLists_simple_example.txt CMakeLists.txt

:: 配置CMake
echo.
echo 配置CMake...
%CMAKE_CMD% . -DCMAKE_TOOLCHAIN_FILE="%CMAKE_TOOLCHAIN_FILE%" -DCMAKE_BUILD_TYPE=Release
if errorlevel 1 (
    echo 错误: CMake配置失败
    pause
    exit /b 1
)

:: 编译项目
echo.
echo 编译项目...
%CMAKE_CMD% --build . --config Release --parallel
if errorlevel 1 (
    echo 错误: 编译失败
    pause
    exit /b 1
)

:: 运行示例
if exist "bin\ffmpeg_simple_example.exe" (
    echo.
    echo 运行示例程序...
    bin\ffmpeg_simple_example.exe
)

echo.
echo ========================================
echo 构建完成！
echo ========================================
echo 构建目录: %BUILD_DIR%
echo 可执行文件位置: %BUILD_DIR%\bin\
echo ========================================

pause 