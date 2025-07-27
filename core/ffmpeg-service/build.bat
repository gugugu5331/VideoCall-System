@echo off
echo ========================================
echo FFmpeg 伪造检测服务编译脚本
echo ========================================

REM 设置环境变量
set FFMPEG_SERVICE_DIR=%~dp0
set BUILD_DIR=%FFMPEG_SERVICE_DIR%build
set INSTALL_DIR=%FFMPEG_SERVICE_DIR%install

REM 创建构建目录
if not exist "%BUILD_DIR%" (
    mkdir "%BUILD_DIR%"
)

REM 进入构建目录
cd "%BUILD_DIR%"

echo 配置CMake项目...

REM 配置项目
cmake .. -G "Visual Studio 16 2019" -A x64 ^
    -DCMAKE_INSTALL_PREFIX="%INSTALL_DIR%" ^
    -DCMAKE_BUILD_TYPE=Release ^
    -DFFMPEG_DIR="C:/ffmpeg" ^
    -DONNXRUNTIME_DIR="C:/onnxruntime"

if %ERRORLEVEL% neq 0 (
    echo CMake配置失败
    pause
    exit /b 1
)

echo 编译项目...

REM 编译项目
cmake --build . --config Release --target install

if %ERRORLEVEL% neq 0 (
    echo 编译失败
    pause
    exit /b 1
)

echo 编译完成！
echo 可执行文件位置: %INSTALL_DIR%\bin\ffmpeg_detection_service.exe

REM 复制配置文件
if not exist "%INSTALL_DIR%\config" (
    mkdir "%INSTALL_DIR%\config"
)
copy "%FFMPEG_SERVICE_DIR%config.json" "%INSTALL_DIR%\config\"

REM 创建模型目录
if not exist "%INSTALL_DIR%\models" (
    mkdir "%INSTALL_DIR%\models"
)

echo.
echo 请将ONNX模型文件放置在 %INSTALL_DIR%\models 目录下
echo 然后运行 start_service.bat 启动服务

pause 