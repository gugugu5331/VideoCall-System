@echo off
setlocal enabledelayedexpansion

echo ========================================
echo FFmpeg服务环境准备脚本 (Windows)
echo ========================================

:: 设置变量
set PROJECT_DIR=%~dp0
set VCPKG_DIR=%PROJECT_DIR%..\..\core\frontend\vcpkg
set VCPKG_EXE=%VCPKG_DIR%\vcpkg.exe

echo 检查vcpkg...
if not exist "%VCPKG_EXE%" (
    echo ❌ vcpkg未找到，请确保vcpkg已正确安装
    echo 请运行: git clone https://github.com/Microsoft/vcpkg.git
    echo 然后运行: %VCPKG_DIR%\bootstrap-vcpkg.bat
    pause
    exit /b 1
)

echo ✅ vcpkg已找到: %VCPKG_EXE%

:: 安装FFmpeg
echo.
echo 正在安装FFmpeg...
%VCPKG_EXE% install ffmpeg:x64-windows
if %ERRORLEVEL% neq 0 (
    echo ❌ FFmpeg安装失败
    pause
    exit /b 1
)
echo ✅ FFmpeg安装成功

:: 安装OpenCV
echo.
echo 正在安装OpenCV...
%VCPKG_EXE% install opencv4:x64-windows
if %ERRORLEVEL% neq 0 (
    echo ❌ OpenCV安装失败
    pause
    exit /b 1
)
echo ✅ OpenCV安装成功

:: 安装ONNX Runtime
echo.
echo 正在安装ONNX Runtime...
%VCPKG_EXE% install onnxruntime-gpu:x64-windows
if %ERRORLEVEL% neq 0 (
    echo ❌ ONNX Runtime安装失败
    pause
    exit /b 1
)
echo ✅ ONNX Runtime安装成功

:: 安装其他依赖
echo.
echo 正在安装其他依赖...
%VCPKG_EXE% install nlohmann-json:x64-windows
%VCPKG_EXE% install spdlog:x64-windows
%VCPKG_EXE% install fmt:x64-windows

echo.
echo ========================================
echo 环境准备完成！
echo ========================================
echo.
echo 已安装的包:
echo - FFmpeg (音视频处理)
echo - OpenCV (图像处理)
echo - ONNX Runtime (深度学习推理)
echo - nlohmann-json (JSON处理)
echo - spdlog (日志记录)
echo - fmt (格式化输出)
echo.
echo 下一步: 运行 build.bat 编译项目
echo.
pause 