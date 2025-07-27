@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    音视频通话系统 - 智能启动
echo ========================================
echo.

:: 检查是否有可执行文件
set "app_found=0"
set "app_path="

if exist bin\VideoCallApp.exe (
    set "app_found=1"
    set "app_path=bin\VideoCallApp.exe"
    set "app_type=完整版本"
) else if exist bin\VideoCallApp_debug.exe (
    set "app_found=1"
    set "app_path=bin\VideoCallApp_debug.exe"
    set "app_type=调试版本"
) else if exist bin\VideoCallApp_simple.exe (
    set "app_found=1"
    set "app_path=bin\VideoCallApp_simple.exe"
    set "app_type=简化版本"
)

if %app_found% equ 1 (
    echo ✓ 找到可执行文件: %app_path%
    echo   类型: %app_type%
    echo.
    echo 正在启动程序...
    echo.
    "%app_path%"
    goto :end
)

:: 没有找到可执行文件，检查环境
echo ❌ 未找到可执行文件
echo.
echo 正在检查环境...

:: 检查Qt
where qmake >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Qt5未安装或未添加到PATH
    echo.
    echo 请先安装Qt5:
    echo 1. 访问 https://www.qt.io/download
    echo 2. 下载并安装Qt5.15+
    echo 3. 将Qt的bin目录添加到系统PATH
    echo.
    pause
    exit /b 1
)

echo ✓ Qt6环境正常

:: 检查OpenCV
where opencv_version >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ OpenCV环境正常
    set "opencv_available=1"
) else (
    echo ⚠️  OpenCV未安装或未添加到PATH
    set "opencv_available=0"
)

echo.
echo 请选择构建方式:
echo.
if %opencv_available% equ 1 (
    echo [1] 构建完整版本 (推荐)
    echo [2] 构建简化版本
    echo [3] 运行环境检查
    echo [0] 退出
    echo.
    set /p choice="请输入选择 (0-3): "
    
    if "%choice%"=="1" goto :build_full
    if "%choice%"=="2" goto :build_simple
    if "%choice%"=="3" goto :check_env
    if "%choice%"=="0" goto :end
) else (
    echo [1] 构建简化版本 (推荐)
    echo [2] 安装OpenCV后构建完整版本
    echo [3] 运行环境检查
    echo [0] 退出
    echo.
    set /p choice="请输入选择 (0-3): "
    
    if "%choice%"=="1" goto :build_simple
    if "%choice%"=="2" goto :install_opencv
    if "%choice%"=="3" goto :check_env
    if "%choice%"=="0" goto :end
)

echo 无效选择
goto :end

:build_full
echo.
echo ========================================
echo    构建完整版本
echo ========================================
call build_qt5.bat
if %errorlevel% equ 0 (
    echo.
    echo 构建成功！正在启动程序...
    echo.
    if exist bin\VideoCallApp.exe (
        bin\VideoCallApp.exe
    ) else if exist bin\VideoCallApp_debug.exe (
        bin\VideoCallApp_debug.exe
    )
)
goto :end

:build_simple
echo.
echo ========================================
echo    构建简化版本
echo ========================================
call build_simple.bat
if %errorlevel% equ 0 (
    echo.
    echo 构建成功！正在启动程序...
    echo.
    if exist bin\VideoCallApp_simple.exe (
        bin\VideoCallApp_simple.exe
    )
)
goto :end

:install_opencv
echo.
echo ========================================
echo    安装OpenCV
echo ========================================
echo.
echo 请选择OpenCV安装方式:
echo.
echo [1] 使用vcpkg安装 (推荐)
echo [2] 下载预编译包
echo [3] 查看详细安装指南
echo [0] 返回
echo.
set /p opencv_choice="请输入选择 (0-3): "

if "%opencv_choice%"=="1" (
    echo.
    echo 使用vcpkg安装OpenCV...
    echo 请确保已安装vcpkg，然后运行:
    echo   .\vcpkg install opencv4[contrib]
    echo.
    echo 安装完成后重新运行此脚本
) else if "%opencv_choice%"=="2" (
    echo.
    echo 下载预编译包:
    echo 1. 访问 https://opencv.org/releases/
    echo 2. 下载Windows版本
    echo 3. 解压到 C:\opencv
    echo 4. 将 C:\opencv\build\x64\vc15\bin 添加到PATH
    echo.
    echo 安装完成后重新运行此脚本
) else if "%opencv_choice%"=="3" (
    if exist INSTALLATION_GUIDE.md (
        start notepad INSTALLATION_GUIDE.md
    ) else (
        echo 安装指南文件不存在
    )
) else if "%opencv_choice%"=="0" (
    goto :end
)
echo.
pause
goto :end

:check_env
echo.
echo ========================================
echo    环境检查
echo ========================================
call check_environment.bat
goto :end

:end
echo.
echo 程序已退出
echo.
pause 