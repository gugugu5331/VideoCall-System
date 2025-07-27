@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    音视频通话系统 - Qt前端运行
echo ========================================
echo.

:: 检查Qt安装
echo 检查Qt安装...
where qmake >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 qmake，请确保Qt6已正确安装并添加到PATH
    echo.
    echo 安装步骤:
    echo 1. 访问 https://www.qt.io/download
    echo 2. 下载Qt在线安装器
    echo 3. 安装Qt6.5+ 和 MinGW编译器
    echo 4. 将Qt的bin目录添加到系统PATH
    echo.
    pause
    exit /b 1
)

:: 显示Qt版本
echo 检测到的Qt版本:
qmake -v
echo.

:: 检查项目文件
if not exist VideoCallApp.pro (
    echo 错误: 未找到 VideoCallApp.pro 项目文件
    echo 请确保在正确的目录中运行此脚本
    pause
    exit /b 1
)

:: 检查是否已构建
if exist bin\VideoCallApp.exe (
    echo 找到已构建的可执行文件，直接运行...
    echo.
    echo 启动音视频通话系统...
    echo.
    bin\VideoCallApp.exe
    goto :end
)

if exist bin\VideoCallApp_debug.exe (
    echo 找到调试版本的可执行文件，直接运行...
    echo.
    echo 启动音视频通话系统 (调试模式)...
    echo.
    bin\VideoCallApp_debug.exe
    goto :end
)

:: 尝试构建项目
echo 未找到可执行文件，尝试构建项目...
echo.

:: 清理之前的构建
echo 清理之前的构建文件...
if exist Makefile del Makefile
if exist Makefile.Debug del Makefile.Debug
if exist Makefile.Release del Makefile.Release
if exist debug rmdir /s /q debug 2>nul
if exist release rmdir /s /q release 2>nul
if exist .qmake.stash del .qmake.stash 2>nul

:: 创建必要的目录
if not exist debug mkdir debug
if not exist release mkdir release
if not exist bin mkdir bin

:: 生成Makefile
echo 生成Makefile...
qmake VideoCallApp.pro -spec win32-g++ CONFIG+=debug
if %errorlevel% neq 0 (
    echo 错误: qmake失败
    echo 请检查项目文件配置
    pause
    exit /b 1
)

:: 编译项目
echo.
echo 开始编译项目...
echo 这可能需要几分钟时间，请耐心等待...
echo.

make -j%NUMBER_OF_PROCESSORS%
if %errorlevel% neq 0 (
    echo.
    echo 错误: 编译失败
    echo.
    echo 可能的解决方案:
    echo 1. 确保Qt6正确安装
    echo 2. 检查编译器配置
    echo 3. 安装必要的依赖库
    echo 4. 检查项目文件语法
    echo.
    pause
    exit /b 1
)

:: 检查构建结果
echo.
echo 检查构建结果...
if exist debug\VideoCallApp.exe (
    copy debug\VideoCallApp.exe bin\VideoCallApp_debug.exe
    echo ✅ 调试版本构建成功
    echo.
    echo 启动音视频通话系统...
    echo.
    bin\VideoCallApp_debug.exe
) else (
    echo ❌ 构建失败，未找到可执行文件
    echo.
    echo 请检查编译输出中的错误信息
    pause
    exit /b 1
)

:end
echo.
echo 程序已退出
pause 