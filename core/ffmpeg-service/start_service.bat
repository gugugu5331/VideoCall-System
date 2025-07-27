@echo off
echo ========================================
echo FFmpeg 伪造检测服务启动脚本
echo ========================================

REM 设置环境变量
set FFMPEG_SERVICE_DIR=%~dp0
set PATH=%FFMPEG_SERVICE_DIR%bin;%PATH%

REM 检查可执行文件是否存在
if not exist "%FFMPEG_SERVICE_DIR%bin\ffmpeg_detection_service.exe" (
    echo 错误: 找不到可执行文件 ffmpeg_detection_service.exe
    echo 请先编译项目
    pause
    exit /b 1
)

REM 检查模型文件
if not exist "%FFMPEG_SERVICE_DIR%models\detection.onnx" (
    echo 警告: 找不到模型文件 models\detection.onnx
    echo 请将ONNX模型文件放置在models目录下
)

REM 创建日志目录
if not exist "%FFMPEG_SERVICE_DIR%logs" (
    mkdir "%FFMPEG_SERVICE_DIR%logs"
)

echo 启动FFmpeg检测服务...
echo 使用 Ctrl+C 停止服务

REM 启动服务
"%FFMPEG_SERVICE_DIR%bin\ffmpeg_detection_service.exe" ^
    -i "rtsp://localhost:8554/stream" ^
    -m "%FFMPEG_SERVICE_DIR%models\detection.onnx" ^
    -c "%FFMPEG_SERVICE_DIR%config.json" ^
    -o "%FFMPEG_SERVICE_DIR%logs\service.log" ^
    -v

pause 