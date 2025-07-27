@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    编译器环境检测工具
echo ========================================
echo.

:: 检查MSVC编译器
echo [1/4] 检查MSVC编译器...
where cl >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ MSVC编译器已找到
    cl 2>&1 | findstr "Microsoft"
) else (
    echo ❌ MSVC编译器未找到
    echo   建议: 安装Visual Studio并选择"使用C++的桌面开发"
)
echo.

:: 检查MinGW编译器
echo [2/4] 检查MinGW编译器...
where g++ >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ MinGW编译器已找到
    g++ --version 2>&1 | findstr "g++"
) else (
    echo ❌ MinGW编译器未找到
    echo   建议: 安装MSYS2或MinGW-w64
)
echo.

:: 检查Visual Studio工具
echo [3/4] 检查Visual Studio工具...
where devenv >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ Visual Studio已安装
    where devenv
) else (
    echo ❌ Visual Studio未安装或未添加到PATH
)
echo.

where msbuild >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ MSBuild已找到
    where msbuild
) else (
    echo ❌ MSBuild未找到
)
echo.

:: 检查vcpkg
echo [4/4] 检查vcpkg...
where vcpkg >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ vcpkg已安装并添加到PATH
    where vcpkg
) else (
    echo ❌ vcpkg未添加到PATH
    echo   请将vcpkg目录添加到系统PATH
)
echo.

:: 生成建议
echo ========================================
echo    环境诊断结果
echo ========================================
echo.

set "msvc_ok=0"
set "mingw_ok=0"
set "vs_ok=0"

where cl >nul 2>&1 && set "msvc_ok=1"
where g++ >nul 2>&1 && set "mingw_ok=1"
where devenv >nul 2>&1 && set "vs_ok=1"

if %msvc_ok% equ 1 (
    echo ✅ MSVC环境正常，可以使用Visual Studio开发者命令提示符
    echo.
    echo 建议操作:
    echo 1. 打开"x64 本机工具命令提示符 for VS 2022"
    echo 2. 运行: .\vcpkg install opencv4[contrib]:x64-windows
) else if %mingw_ok% equ 1 (
    echo ✅ MinGW环境正常，可以使用MinGW编译OpenCV
    echo.
    echo 建议操作:
    echo 1. 在普通cmd中运行: .\vcpkg install opencv4[contrib]:x64-mingw-dynamic
) else (
    echo ❌ 未找到任何可用的C++编译器
    echo.
    echo 解决方案:
    echo 1. 安装Visual Studio Community（推荐）
    echo    - 下载: https://visualstudio.microsoft.com/zh-hans/vs/community/
    echo    - 选择"使用C++的桌面开发"工作负载
    echo.
    echo 2. 或安装MSYS2 + MinGW-w64
    echo    - 下载: https://www.msys2.org/
    echo    - 安装后运行: pacman -S mingw-w64-x86_64-gcc
    echo.
    echo 3. 或使用预编译的OpenCV
    echo    - 下载: https://opencv.org/releases/
    echo    - 解压到C:\opencv并添加到PATH
)

echo.
echo ========================================
echo    检测完成
echo ========================================
echo.

pause 