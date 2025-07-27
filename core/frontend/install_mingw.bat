@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    MinGW 安装助手
echo ========================================
echo.

:: 检查当前MinGW安装
echo 检查当前MinGW安装...
where g++ >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ g++ 已找到
    g++ --version 2>&1 | findstr "g++"
) else (
    echo ❌ g++ 未找到
)

where mingw32-make >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ mingw32-make 已找到
    mingw32-make --version 2>&1 | findstr "GNU Make"
) else (
    echo ❌ mingw32-make 未找到
)

echo.

:: 检查MSYS2
echo 检查MSYS2安装...
if exist "C:\msys64\mingw64\bin\g++.exe" (
    echo ✓ MSYS2 MinGW 已安装
    echo 路径: C:\msys64\mingw64\bin
) else (
    echo ❌ MSYS2 MinGW 未安装
)

echo.

:: 检查Chocolatey
echo 检查Chocolatey...
where choco >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ Chocolatey 已安装
) else (
    echo ❌ Chocolatey 未安装
)

echo.
echo 请选择MinGW安装方式:
echo.
echo [1] 使用MSYS2安装 (推荐)
echo [2] 使用Chocolatey安装
echo [3] 手动下载安装
echo [4] 检查PATH配置
echo [0] 退出
echo.
set /p choice="请输入选择 (0-4): "

if "%choice%"=="1" goto :install_msys2
if "%choice%"=="2" goto :install_chocolatey
if "%choice%"=="3" goto :manual_install
if "%choice%"=="4" goto :check_path
if "%choice%"=="0" goto :end

echo 无效选择
goto :end

:install_msys2
echo.
echo ========================================
echo    使用MSYS2安装MinGW
echo ========================================
echo.
echo 请按以下步骤操作:
echo.
echo 1. 访问: https://www.msys2.org/
echo 2. 下载Windows版本 (64位)
echo 3. 安装到默认路径: C:\msys64
echo 4. 打开MSYS2 MinGW 64-bit终端
echo 5. 运行以下命令:
echo    pacman -S mingw-w64-x86_64-gcc
echo    pacman -S mingw-w64-x86_64-make
echo    pacman -S mingw-w64-x86_64-binutils
echo 6. 将 C:\msys64\mingw64\bin 添加到系统PATH
echo.
echo 安装完成后，请重新运行此脚本进行验证
pause
goto :end

:install_chocolatey
echo.
echo ========================================
echo    使用Chocolatey安装MinGW
echo ========================================
echo.
if exist "C:\ProgramData\chocolatey\bin\choco.exe" (
    echo 正在安装MinGW...
    choco install mingw -y
    echo.
    echo MinGW安装完成，请重新运行此脚本进行验证
) else (
    echo 请先安装Chocolatey:
    echo 1. 以管理员权限打开PowerShell
    echo 2. 运行以下命令:
    echo    Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
    echo 3. 安装完成后重新运行此脚本
)
pause
goto :end

:manual_install
echo.
echo ========================================
echo    手动安装MinGW
echo ========================================
echo.
echo 手动安装步骤:
echo.
echo 1. 访问: https://www.mingw-w64.org/downloads/
echo 2. 下载Windows x86_64版本
echo 3. 解压到: C:\mingw64
echo 4. 将 C:\mingw64\bin 添加到系统PATH
echo.
echo 安装完成后，请重新运行此脚本进行验证
pause
goto :end

:check_path
echo.
echo ========================================
echo    检查PATH配置
echo ========================================
echo.
echo 当前PATH中的MinGW相关路径:
for %%i in (%PATH%) do (
    echo %%i | findstr /i "mingw" >nul && echo   ✓ %%i
    echo %%i | findstr /i "msys" >nul && echo   ✓ %%i
)

echo.
echo 建议的PATH配置:
echo   C:\msys64\mingw64\bin
echo   或
echo   C:\mingw64\bin
echo.
echo 请确保MinGW的bin目录在PATH中
pause
goto :end

:end
echo.
echo ========================================
echo    操作完成
echo ========================================
echo.
pause 