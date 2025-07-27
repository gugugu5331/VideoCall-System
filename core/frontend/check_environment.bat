@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    环境检查工具 - Qt6 + OpenCV
echo ========================================
echo.

:: 检查系统信息
echo [1/8] 检查系统信息...
echo 操作系统: %OS%
echo 处理器架构: %PROCESSOR_ARCHITECTURE%
echo 处理器数量: %NUMBER_OF_PROCESSORS%
echo.

:: 检查PATH环境变量
echo [2/8] 检查PATH环境变量...
echo 当前PATH包含以下Qt相关路径:
for %%i in (%PATH%) do (
    echo %%i | findstr /i "qt" >nul && echo   ✓ %%i
)
echo.

:: 检查Qt安装
echo [3/8] 检查Qt安装...
where qmake >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ qmake 已找到
    echo Qt版本信息:
    qmake -v 2>&1 | findstr /i "version"
) else (
    echo ❌ qmake 未找到
    echo 请安装Qt6并添加到PATH
)
echo.

:: 检查编译器
echo [4/8] 检查编译器...
where g++ >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ g++ 编译器已找到
    g++ --version 2>&1 | findstr /i "g++"
) else (
    echo ❌ g++ 编译器未找到
    where cl >nul 2>&1
    if %errorlevel% equ 0 (
        echo ✓ MSVC 编译器已找到
        cl 2>&1 | findstr /i "version"
    ) else (
        echo ❌ 未找到任何C++编译器
    )
)
echo.

:: 检查OpenCV
echo [5/8] 检查OpenCV...
where opencv_version >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ OpenCV 已找到
    opencv_version
) else (
    echo ❌ OpenCV 未找到或未添加到PATH
    echo 检查常见OpenCV路径:
    if exist "C:\opencv\build\x64\vc15\bin\opencv_world*.dll" (
        echo   ✓ C:\opencv\build\x64\vc15\bin
    )
    if exist "C:\Program Files\opencv\build\x64\vc15\bin\opencv_world*.dll" (
        echo   ✓ C:\Program Files\opencv\build\x64\vc15\bin
    )
)
echo.

:: 检查项目文件
echo [6/8] 检查项目文件...
if exist VideoCallApp.pro (
    echo ✓ VideoCallApp.pro 项目文件存在
) else (
    echo ❌ VideoCallApp.pro 项目文件不存在
)
if exist main.cpp (
    echo ✓ main.cpp 源文件存在
) else (
    echo ❌ main.cpp 源文件不存在
)
echo.

:: 检查构建目录
echo [7/8] 检查构建目录...
if exist debug (
    echo ✓ debug 目录存在
    if exist debug\VideoCallApp.exe (
        echo   ✓ 调试版本可执行文件存在
    ) else (
        echo   ❌ 调试版本可执行文件不存在
    )
) else (
    echo ❌ debug 目录不存在
)

if exist release (
    echo ✓ release 目录存在
    if exist release\VideoCallApp.exe (
        echo   ✓ 发布版本可执行文件存在
    ) else (
        echo   ❌ 发布版本可执行文件不存在
    )
) else (
    echo ❌ release 目录不存在
)
echo.

:: 检查依赖库
echo [8/8] 检查依赖库...
echo 检查Qt库文件:
for %%i in (Qt6Core Qt6Gui Qt6Widgets Qt6Multimedia Qt6Network) do (
    where %%i.dll >nul 2>&1 && echo   ✓ %%i.dll || echo   ❌ %%i.dll
)

echo 检查OpenCV库文件:
for %%i in (opencv_core opencv_imgproc opencv_videoio opencv_face opencv_dnn) do (
    where %%i*.dll >nul 2>&1 && echo   ✓ %%i*.dll || echo   ❌ %%i*.dll
)
echo.

:: 生成诊断报告
echo ========================================
echo    诊断报告
echo ========================================
echo.

:: 检查关键组件
set "qt_ok=0"
set "opencv_ok=0"
set "compiler_ok=0"

where qmake >nul 2>&1 && set "qt_ok=1"
where opencv_version >nul 2>&1 && set "opencv_ok=1"
where g++ >nul 2>&1 && set "compiler_ok=1"
where cl >nul 2>&1 && set "compiler_ok=1"

if %qt_ok% equ 1 (
    echo ✓ Qt6 环境正常
) else (
    echo ❌ Qt6 环境异常
    echo   建议: 安装Qt6并添加到PATH
)

if %opencv_ok% equ 1 (
    echo ✓ OpenCV 环境正常
) else (
    echo ❌ OpenCV 环境异常
    echo   建议: 安装OpenCV并添加到PATH
)

if %compiler_ok% equ 1 (
    echo ✓ 编译器环境正常
) else (
    echo ❌ 编译器环境异常
    echo   建议: 安装MinGW或Visual Studio
)

echo.
if %qt_ok% equ 1 if %opencv_ok% equ 1 if %compiler_ok% equ 1 (
    echo 🎉 所有环境检查通过！可以开始构建项目。
    echo.
    echo 运行以下命令开始构建:
    echo   .\build_qt6.bat
    echo   或
    echo   .\run_qt_frontend.bat
) else (
    echo ⚠️  环境检查发现问题，请先解决上述问题再构建项目。
    echo.
    echo 参考安装指南: INSTALLATION_GUIDE.md
)

echo.
echo ========================================
echo    环境检查完成
echo ========================================
echo.

pause 