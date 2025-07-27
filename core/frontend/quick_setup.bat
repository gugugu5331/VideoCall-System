@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    音视频通话系统 - 快速设置
echo ========================================
echo.

:: 检查是否以管理员权限运行
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ⚠️  建议以管理员权限运行此脚本以获得最佳效果
    echo.
)

:: 显示菜单
:menu
echo 请选择要执行的操作:
echo.
echo [1] 检查环境状态
echo [2] 修复OpenCV路径
echo [3] 构建项目
echo [4] 运行程序
echo [5] 完整设置流程
echo [6] 查看安装指南
echo [0] 退出
echo.
set /p choice="请输入选择 (0-6): "

if "%choice%"=="1" goto :check_env
if "%choice%"=="2" goto :fix_opencv
if "%choice%"=="3" goto :build_project
if "%choice%"=="4" goto :run_app
if "%choice%"=="5" goto :full_setup
if "%choice%"=="6" goto :show_guide
if "%choice%"=="0" goto :end
echo 无效选择，请重新输入
goto :menu

:check_env
echo.
echo ========================================
echo    检查环境状态
echo ========================================
call check_environment.bat
goto :menu

:fix_opencv
echo.
echo ========================================
echo    修复OpenCV路径
echo ========================================
call fix_opencv_paths.bat
goto :menu

:build_project
echo.
echo ========================================
echo    构建项目
echo ========================================
call build_qt6.bat
goto :menu

:run_app
echo.
echo ========================================
echo    运行程序
echo ========================================
call run_qt_frontend.bat
goto :menu

:full_setup
echo.
echo ========================================
echo    完整设置流程
echo ========================================
echo.
echo 开始完整设置流程...
echo.

:: 步骤1: 检查环境
echo [步骤 1/5] 检查环境状态...
call check_environment.bat >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 环境检查失败，请先解决环境问题
    pause
    goto :menu
)
echo ✓ 环境检查通过

:: 步骤2: 修复OpenCV路径
echo.
echo [步骤 2/5] 修复OpenCV路径...
call fix_opencv_paths.bat >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ OpenCV路径修复失败
    pause
    goto :menu
)
echo ✓ OpenCV路径修复完成

:: 步骤3: 清理构建文件
echo.
echo [步骤 3/5] 清理构建文件...
if exist Makefile del Makefile
if exist Makefile.Debug del Makefile.Debug
if exist Makefile.Release del Makefile.Release
if exist debug rmdir /s /q debug 2>nul
if exist release rmdir /s /q release 2>nul
if exist .qmake.stash del .qmake.stash 2>nul
echo ✓ 构建文件清理完成

:: 步骤4: 构建项目
echo.
echo [步骤 4/5] 构建项目...
call build_qt6.bat >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ 项目构建失败
    pause
    goto :menu
)
echo ✓ 项目构建完成

:: 步骤5: 验证结果
echo.
echo [步骤 5/5] 验证构建结果...
if exist bin\VideoCallApp.exe (
    echo ✓ 发布版本构建成功
    set "app_path=bin\VideoCallApp.exe"
) else if exist bin\VideoCallApp_debug.exe (
    echo ✓ 调试版本构建成功
    set "app_path=bin\VideoCallApp_debug.exe"
) else (
    echo ❌ 未找到可执行文件
    pause
    goto :menu
)

echo.
echo ========================================
echo    完整设置流程完成！
echo ========================================
echo.
echo 可执行文件位置: %app_path%
echo.
echo 是否立即运行程序? (Y/N)
set /p run_now="请输入选择: "
if /i "%run_now%"=="Y" (
    echo.
    echo 启动程序...
    "%app_path%"
)

goto :menu

:show_guide
echo.
echo ========================================
echo    安装指南
echo ========================================
echo.
if exist INSTALLATION_GUIDE.md (
    echo 正在打开安装指南...
    start notepad INSTALLATION_GUIDE.md
) else (
    echo 安装指南文件不存在
    echo.
    echo 快速安装步骤:
    echo 1. 访问 https://www.qt.io/download 下载Qt6
    echo 2. 访问 https://opencv.org/releases/ 下载OpenCV
    echo 3. 将Qt和OpenCV的bin目录添加到系统PATH
    echo 4. 运行此脚本进行配置
)
echo.
pause
goto :menu

:end
echo.
echo 感谢使用音视频通话系统！
echo.
exit /b 0 