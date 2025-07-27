@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    音视频通话系统 - Qt5 前端构建
echo ========================================
echo.

:: 检查Qt安装
echo 检查Qt安装...
where qmake >nul 2>&1
if %errorlevel% neq 0 (
    echo 错误: 未找到 qmake，请确保Qt5已正确安装并添加到PATH
    echo 请访问 https://www.qt.io/download 下载并安装Qt5
    pause
    exit /b 1
)

:: 检查编译器
echo 检查编译器...
where g++ >nul 2>&1
if %errorlevel% neq 0 (
    echo 警告: 未找到g++编译器，尝试使用MSVC...
    where cl >nul 2>&1
    if %errorlevel% neq 0 (
        echo 错误: 未找到编译器，请安装MinGW或Visual Studio
        pause
        exit /b 1
    )
)

:: 显示Qt版本
echo 检测到的Qt版本:
qmake -v
echo.

:: 清理之前的构建
echo 清理之前的构建文件...
if exist Makefile del Makefile
if exist Makefile.Debug del Makefile.Debug
if exist Makefile.Release del Makefile.Release
if exist debug rmdir /s /q debug
if exist release rmdir /s /q release
if exist .qmake.stash del .qmake.stash

:: 创建必要的目录
echo 创建构建目录...
if not exist debug mkdir debug
if not exist release mkdir release
if not exist bin mkdir bin

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

:: 生成Makefile
echo 生成Makefile...
qmake VideoCallApp.pro -spec win32-g++ CONFIG+=debug
if %errorlevel% neq 0 (
    echo 错误: qmake失败
    pause
    exit /b 1
)

:: 编译项目
echo.
echo 开始编译项目...
echo 这可能需要几分钟时间，请耐心等待...
echo.

%make_cmd% -j%NUMBER_OF_PROCESSORS%
if %errorlevel% neq 0 (
    echo.
    echo 错误: 编译失败
    echo 请检查以下可能的问题:
    echo 1. Qt5是否正确安装
    echo 2. 编译器是否正确配置
    echo 3. 依赖库是否缺失
    echo.
    pause
    exit /b 1
)

:: 复制可执行文件
echo.
echo 复制可执行文件...
if exist debug\VideoCallApp.exe (
    copy debug\VideoCallApp.exe bin\VideoCallApp_debug.exe
    echo 调试版本构建完成: bin\VideoCallApp_debug.exe
) else (
    echo ❌ 构建失败，未找到可执行文件
    pause
    exit /b 1
)

:: 构建发布版本
echo.
echo 构建发布版本...
qmake VideoCallApp.pro -spec win32-g++ CONFIG+=release
if %errorlevel% equ 0 (
    %make_cmd% -j%NUMBER_OF_PROCESSORS%
    if %errorlevel% equ 0 (
        if exist release\VideoCallApp.exe (
            copy release\VideoCallApp.exe bin\VideoCallApp.exe
            echo 发布版本构建完成: bin\VideoCallApp.exe
        )
    )
)

:: 检查构建结果
echo.
echo 检查构建结果...
if exist bin\VideoCallApp.exe (
    echo ✅ 发布版本构建成功
) else if exist bin\VideoCallApp_debug.exe (
    echo ✅ 调试版本构建成功
) else (
    echo ❌ 构建失败，未找到可执行文件
    pause
    exit /b 1
)

:: 显示文件信息
echo.
echo 构建文件信息:
if exist bin\VideoCallApp.exe (
    echo 发布版本: bin\VideoCallApp.exe
    dir bin\VideoCallApp.exe
)
if exist bin\VideoCallApp_debug.exe (
    echo 调试版本: bin\VideoCallApp_debug.exe
    dir bin\VideoCallApp_debug.exe
)

echo.
echo ========================================
echo    构建完成！
echo ========================================
echo.
echo 运行应用程序:
echo   bin\VideoCallApp.exe
echo.
echo 如需调试，请运行:
echo   bin\VideoCallApp_debug.exe
echo.

pause 