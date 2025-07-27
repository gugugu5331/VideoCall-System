@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    Qt6 安装助手
echo ========================================
echo.

:: 检查当前Qt版本
echo 检查当前Qt版本...
where qmake >nul 2>&1
if %errorlevel% equ 0 (
    echo 当前Qt版本:
    qmake -v
    echo.
    echo 是否已安装Qt6? (Y/N)
    set /p has_qt6="请输入选择: "
    if /i "%has_qt6%"=="Y" (
        echo 请确保Qt6的bin目录已添加到PATH
        echo 常见路径: C:\Qt\6.5.x\mingw_64\bin
        goto :check_qt6
    )
) else (
    echo 未检测到Qt安装
)

echo.
echo 请选择Qt6安装方式:
echo.
echo [1] 使用Qt在线安装器 (推荐)
echo [2] 使用Chocolatey包管理器
echo [3] 使用vcpkg包管理器
echo [4] 手动下载安装
echo [0] 退出
echo.
set /p choice="请输入选择 (0-4): "

if "%choice%"=="1" goto :online_installer
if "%choice%"=="2" goto :chocolatey_install
if "%choice%"=="3" goto :vcpkg_install
if "%choice%"=="4" goto :manual_install
if "%choice%"=="0" goto :end

echo 无效选择
goto :end

:online_installer
echo.
echo ========================================
echo    使用Qt在线安装器
echo ========================================
echo.
echo 正在打开Qt下载页面...
echo.
echo 请按以下步骤操作:
echo.
echo 1. 访问: https://www.qt.io/download
echo 2. 点击 "Download Qt"
echo 3. 下载 "Qt Online Installer"
echo 4. 运行安装器并创建Qt账户
echo 5. 选择以下组件:
echo    ✅ Qt 6.5.x (最新稳定版)
echo    ✅ MinGW 11.2.0 64-bit
echo    ✅ Qt Creator
echo    ✅ Qt Debug Information Files
echo    ✅ Qt WebEngine
echo    ✅ Qt Multimedia
echo 6. 安装到 C:\Qt
echo.
echo 安装完成后，请重新运行此脚本进行验证
echo.
pause
goto :end

:chocolatey_install
echo.
echo ========================================
echo    使用Chocolatey安装
echo ========================================
echo.
echo 检查Chocolatey...
where choco >nul 2>&1
if %errorlevel% neq 0 (
    echo 正在安装Chocolatey...
    echo 请以管理员权限运行此脚本
    powershell -Command "Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
    echo.
    echo Chocolatey安装完成，请重新运行此脚本
    pause
    goto :end
)

echo 正在安装Qt6...
choco install qt6 -y
echo.
echo Qt6安装完成，请重新运行此脚本进行验证
pause
goto :end

:vcpkg_install
echo.
echo ========================================
echo    使用vcpkg安装
echo ========================================
echo.
echo 检查vcpkg...
where vcpkg >nul 2>&1
if %errorlevel% neq 0 (
    echo 请先安装vcpkg:
    echo   git clone https://github.com/microsoft/vcpkg.git
    echo   cd vcpkg
    echo   .\bootstrap-vcpkg.bat
    echo.
    echo 安装完成后重新运行此脚本
    pause
    goto :end
)

echo 正在安装Qt6...
.\vcpkg install qt6-base qt6-multimedia qt6-webengine
echo.
echo Qt6安装完成，请重新运行此脚本进行验证
pause
goto :end

:manual_install
echo.
echo ========================================
echo    手动安装指南
echo ========================================
echo.
echo 手动安装步骤:
echo.
echo 1. 访问 Qt 官网: https://www.qt.io/download
echo 2. 下载 Qt 6.5.x 离线安装包
echo 3. 运行安装程序
echo 4. 选择组件并安装
echo 5. 配置环境变量
echo.
echo 安装完成后，请重新运行此脚本进行验证
pause
goto :end

:check_qt6
echo.
echo ========================================
echo    验证Qt6安装
echo ========================================
echo.
echo 检查Qt6安装...

:: 检查Qt6版本
qmake -v 2>&1 | findstr "Qt version 6" >nul
if %errorlevel% equ 0 (
    echo ✅ Qt6安装成功
    qmake -v
) else (
    echo ❌ 未检测到Qt6
    echo 当前Qt版本:
    qmake -v
    echo.
    echo 请确保Qt6已正确安装并添加到PATH
    echo 常见Qt6路径: C:\Qt\6.5.x\mingw_64\bin
    pause
    goto :end
)

:: 检查编译器
echo.
echo 检查编译器...
where g++ >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ MinGW编译器已找到
    g++ --version 2>&1 | findstr "g++"
) else (
    echo ❌ MinGW编译器未找到
    echo 请确保MinGW已安装并添加到PATH
    echo 常见MinGW路径: C:\Qt\Tools\mingw1120_64\bin
)

:: 检查Qt Creator
echo.
echo 检查Qt Creator...
where qtcreator >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Qt Creator已找到
) else (
    echo ⚠️  Qt Creator未找到或未添加到PATH
)

echo.
echo ========================================
echo    Qt6安装验证完成
echo ========================================
echo.
echo 现在可以构建项目:
echo   .\build_qt6.bat
echo.
echo 或运行简化版本:
echo   .\build_simple.bat
echo.

:end
echo.
echo ========================================
echo    操作完成
echo ========================================
echo.
pause 