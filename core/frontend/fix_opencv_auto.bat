@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    OpenCV 自动安装修复工具
echo ========================================
echo.

:: 检查当前OpenCV状态
echo 检查当前OpenCV安装状态...
set "opencv_found=0"
set "opencv_path="

:: 检查常见路径
for %%p in (
    "C:\opencv\build\x64\vc15\bin"
    "C:\opencv\build\x64\vc16\bin"
    "C:\opencv\build\x64\mingw\bin"
    "C:\Program Files\opencv\build\x64\vc15\bin"
    "C:\Program Files\opencv\build\x64\vc16\bin"
    "C:\vcpkg\installed\x64-windows\bin"
    "C:\vcpkg\installed\x64-mingw-dynamic\bin"
) do (
    if exist "%%~p\opencv_world*.dll" (
        set "opencv_found=1"
        set "opencv_path=%%~p"
        echo ✓ 找到OpenCV: %%~p
        goto :found_opencv
    )
)

:found_opencv

if %opencv_found% equ 1 (
    echo.
    echo OpenCV已安装，路径: %opencv_path%
    echo.
    echo 是否要配置项目文件以使用此OpenCV? (Y/N)
    set /p configure="请输入选择: "
    if /i "%configure%"=="Y" (
        call fix_opencv_paths.bat
        goto :end
    )
) else (
    echo ❌ 未找到OpenCV安装
    echo.
    echo 请选择安装方式:
    echo.
    echo [1] 下载预编译OpenCV (推荐)
    echo [2] 尝试修复vcpkg安装
    echo [3] 使用简化版本 (无需OpenCV)
    echo [0] 退出
    echo.
    set /p choice="请输入选择 (0-3): "
    
    if "%choice%"=="1" goto :download_opencv
    if "%choice%"=="2" goto :fix_vcpkg
    if "%choice%"=="3" goto :use_simple
    if "%choice%"=="0" goto :end
)

goto :end

:download_opencv
echo.
echo ========================================
echo    下载预编译OpenCV
echo ========================================
echo.
echo 正在打开OpenCV下载页面...
echo 请按以下步骤操作:
echo.
echo 1. 访问: https://opencv.org/releases/
echo 2. 下载最新版本的Windows包
echo 3. 双击下载的文件，解压到 C:\opencv
echo 4. 完成后重新运行此脚本
echo.
echo 是否已下载并解压完成? (Y/N)
set /p downloaded="请输入选择: "
if /i "%downloaded%"=="Y" (
    echo.
    echo 正在配置环境变量...
    echo 请手动将以下路径添加到系统PATH:
    echo   C:\opencv\build\x64\vc16\bin
    echo.
    echo 添加完成后，运行: .\fix_opencv_paths.bat
) else (
    echo 请先下载并解压OpenCV，然后重新运行此脚本
)
goto :end

:fix_vcpkg
echo.
echo ========================================
echo    修复vcpkg安装
echo ========================================
echo.
echo 正在检查vcpkg环境...

:: 检查vcpkg
where vcpkg >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ vcpkg未找到
    echo 请先安装vcpkg:
    echo   git clone https://github.com/microsoft/vcpkg.git
    echo   cd vcpkg
    echo   .\bootstrap-vcpkg.bat
    goto :end
)

echo ✓ vcpkg已找到

:: 检查编译器
echo 检查编译器...
where cl >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ MSVC编译器已找到
    echo 尝试使用MSVC安装OpenCV...
    echo.
    echo 请运行以下命令:
    echo   .\vcpkg install opencv4[contrib]:x64-windows
    echo.
    echo 如果仍有问题，请:
    echo 1. 使用"x64 本机工具命令提示符 for VS 2022"
    echo 2. 或安装Visual Studio Community
    goto :end
)

where g++ >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ MinGW编译器已找到
    echo 尝试使用MinGW安装OpenCV...
    echo.
    echo 请运行以下命令:
    echo   .\vcpkg install opencv4[contrib]:x64-mingw-dynamic
    goto :end
)

echo ❌ 未找到编译器
echo.
echo 建议安装Visual Studio Community:
echo 1. 下载: https://visualstudio.microsoft.com/zh-hans/vs/community/
echo 2. 选择"使用C++的桌面开发"工作负载
echo 3. 安装完成后重新运行此脚本
goto :end

:use_simple
echo.
echo ========================================
echo    使用简化版本
echo ========================================
echo.
echo 简化版本不需要OpenCV，可以直接运行:
echo.
echo 构建简化版本:
echo   .\build_simple.bat
echo.
echo 运行简化版本:
echo   .\bin\VideoCallApp_simple.exe
echo.
echo 是否现在构建简化版本? (Y/N)
set /p build_now="请输入选择: "
if /i "%build_now%"=="Y" (
    echo.
    echo 正在构建简化版本...
    call build_simple.bat
    if %errorlevel% equ 0 (
        echo.
        echo 构建成功！是否运行程序? (Y/N)
        set /p run_now="请输入选择: "
        if /i "%run_now%"=="Y" (
            if exist bin\VideoCallApp_simple.exe (
                echo 启动程序...
                bin\VideoCallApp_simple.exe
            )
        )
    )
)
goto :end

:end
echo.
echo ========================================
echo    操作完成
echo ========================================
echo.
pause 