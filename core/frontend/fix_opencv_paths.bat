@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    OpenCV 路径修复工具
echo ========================================
echo.

:: 检查当前OpenCV路径
echo 检查当前OpenCV安装...
set "opencv_found=0"
set "opencv_path="

:: 检查常见OpenCV安装路径
for %%p in (
    "C:\opencv\build\x64\vc15\bin"
    "C:\opencv\build\x64\vc16\bin"
    "C:\opencv\build\x64\mingw\bin"
    "C:\Program Files\opencv\build\x64\vc15\bin"
    "C:\Program Files\opencv\build\x64\vc16\bin"
    "C:\Program Files\opencv\build\x64\mingw\bin"
    "C:\vcpkg\installed\x64-windows\bin"
    "C:\vcpkg\installed\x64-windows\debug\bin"
) do (
    if exist "%%~p\opencv_world*.dll" (
        set "opencv_found=1"
        set "opencv_path=%%~p"
        echo ✓ 找到OpenCV: %%~p
        goto :found_opencv
    )
)

:found_opencv

if %opencv_found% equ 0 (
    echo ❌ 未找到OpenCV安装
    echo.
    echo 请先安装OpenCV，然后重新运行此脚本
    echo 安装方法请参考: INSTALLATION_GUIDE.md
    pause
    exit /b 1
)

echo.
echo 检测到的OpenCV路径: %opencv_path%
echo.

:: 检查项目文件中的OpenCV配置
echo 检查项目文件配置...
if not exist VideoCallApp.pro (
    echo ❌ VideoCallApp.pro 文件不存在
    pause
    exit /b 1
)

:: 备份原文件
echo 备份原项目文件...
copy VideoCallApp.pro VideoCallApp.pro.backup
echo ✓ 已备份到 VideoCallApp.pro.backup

:: 更新项目文件中的OpenCV路径
echo 更新OpenCV路径配置...

:: 创建临时文件
set "temp_file=%temp%\VideoCallApp_temp.pro"

:: 读取原文件并替换OpenCV路径
(
for /f "usebackq delims=" %%i in ("VideoCallApp.pro") do (
    set "line=%%i"
    setlocal enabledelayedexpansion
    set "line=!line!"
    
    :: 替换包含路径
    echo !line! | findstr /i "INCLUDEPATH.*opencv" >nul
    if !errorlevel! equ 0 (
        echo INCLUDEPATH += %opencv_path%\..\include
    ) else (
        :: 替换库路径
        echo !line! | findstr /i "LIBS.*opencv" >nul
        if !errorlevel! equ 0 (
            echo LIBS += -L%opencv_path% -lopencv_world
        ) else (
            echo !line!
        )
    )
    endlocal
)
) > "%temp_file%"

:: 检查是否成功更新
findstr /i "opencv_world" "%temp_file%" >nul
if %errorlevel% equ 0 (
    echo ✓ 已更新OpenCV库配置
) else (
    echo 添加OpenCV配置...
    echo. >> "%temp_file%"
    echo # OpenCV 配置 >> "%temp_file%"
    echo INCLUDEPATH += %opencv_path%\..\include >> "%temp_file%"
    echo LIBS += -L%opencv_path% -lopencv_world >> "%temp_file%"
    echo ✓ 已添加OpenCV配置
)

:: 替换原文件
move /y "%temp_file%" VideoCallApp.pro
echo ✓ 项目文件已更新

:: 检查PATH环境变量
echo.
echo 检查PATH环境变量...
echo %PATH% | findstr /i "%opencv_path%" >nul
if %errorlevel% equ 0 (
    echo ✓ OpenCV路径已在PATH中
) else (
    echo ⚠️  OpenCV路径未在PATH中
    echo 建议将以下路径添加到系统PATH:
    echo   %opencv_path%
    echo.
    echo 临时添加到当前会话PATH...
    set "PATH=%opencv_path%;%PATH%"
    echo ✓ 已临时添加到PATH
)

:: 验证配置
echo.
echo 验证OpenCV配置...
if exist "%opencv_path%\opencv_world*.dll" (
    echo ✓ OpenCV库文件存在
    dir "%opencv_path%\opencv_world*.dll" | findstr "opencv_world"
) else (
    echo ❌ OpenCV库文件不存在
)

if exist "%opencv_path%\..\include\opencv2\opencv.hpp" (
    echo ✓ OpenCV头文件存在
) else (
    echo ❌ OpenCV头文件不存在
)

echo.
echo ========================================
echo    OpenCV 路径修复完成
echo ========================================
echo.
echo 现在可以尝试构建项目:
echo   .\build_qt6.bat
echo.
echo 如果仍有问题，请检查:
echo 1. OpenCV版本是否与编译器兼容
echo 2. 库文件是否为正确的架构(x64/x86)
echo 3. 编译器设置是否正确
echo.

pause 