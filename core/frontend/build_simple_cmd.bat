@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    音视频通话系统 - 简化构建 (CMD版本)
echo ========================================
echo.

:: 检查Qt安装
echo 检查Qt安装...
where qmake >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 qmake，请确保Qt已正确安装并添加到PATH
    pause
    exit /b 1
)

:: 显示Qt版本
echo 检测到的Qt版本:
qmake -v
echo.

:: 检查项目文件
if not exist VideoCallApp_simple.pro (
    echo 错误: 未找到 VideoCallApp_simple.pro 项目文件
    pause
    exit /b 1
)

:: 清理之前的构建
echo 清理之前的构建文件...
if exist Makefile del Makefile
if exist Makefile.Debug del Makefile.Debug
if exist Makefile.Release del Makefile.Release
if exist debug rmdir /s /q debug 2>nul
if exist release rmdir /s /q release 2>nul
if exist .qmake.stash del .qmake.stash 2>nul

:: 创建必要的目录
echo 创建构建目录...
if not exist debug mkdir debug
if not exist release mkdir release
if not exist bin mkdir bin

:: 生成Makefile
echo 生成Makefile...
qmake VideoCallApp_simple.pro -spec win32-g++ CONFIG+=debug
if %errorlevel% neq 0 (
    echo 错误: qmake失败
    echo 请检查项目文件配置
    pause
    exit /b 1
)

:: 检查make工具
echo 检查make工具...
where mingw32-make >nul 2>&1
if %errorlevel% equ 0 (
    set "make_cmd=mingw32-make"
    echo 使用 mingw32-make
) else (
    where make >nul 2>&1
    if %errorlevel% equ 0 (
        set "make_cmd=make"
        echo 使用 make
    ) else (
        echo 错误: 未找到make工具
        echo 请安装MinGW或确保make在PATH中
        pause
        exit /b 1
    )
)

:: 编译项目
echo.
echo 开始编译项目（简化模式）...
echo 注意: 此版本不包含视频通话和安全检测功能
echo 这可能需要几分钟时间，请耐心等待...
echo.

%make_cmd% -j%NUMBER_OF_PROCESSORS%
if %errorlevel% neq 0 (
    echo.
    echo 错误: 编译失败
    echo.
    echo 可能的解决方案:
    echo 1. 确保Qt正确安装
    echo 2. 检查编译器配置
    echo 3. 检查项目文件语法
    echo 4. 尝试安装标准Qt而不是Anaconda Qt
    echo.
    pause
    exit /b 1
)

:: 复制可执行文件
echo.
echo 复制可执行文件...
if exist debug\VideoCallApp.exe (
    copy debug\VideoCallApp.exe bin\VideoCallApp_simple.exe
    echo 简化版本构建完成: bin\VideoCallApp_simple.exe
) else (
    echo 构建失败，未找到可执行文件
    pause
    exit /b 1
)

:: 显示文件信息
echo.
echo 构建文件信息:
dir bin\VideoCallApp_simple.exe

echo.
echo ========================================
echo    简化构建完成！
echo ========================================
echo.
echo 简化版本功能:
echo - 基本界面框架
echo - 登录界面
echo - 主窗口
echo.
echo 注意: 此版本不包含以下功能:
echo - 视频通话
echo - 音频处理
echo - 安全检测
echo - 网络通信
echo.
echo 如需完整功能，请:
echo 1. 安装标准Qt (推荐)
echo 2. 安装OpenCV库
echo 3. 运行完整构建脚本
echo.
echo 运行简化版本:
echo   bin\VideoCallApp_simple.exe
echo.

pause 