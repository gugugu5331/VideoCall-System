@echo off
chcp 65001 >nul
echo 调试运行前端应用...
echo.

cd /d "%~dp0"

echo 设置环境变量...
set QTFRAMEWORK_BYPASS_LICENSE_CHECK=1
set PATH=C:\Qt\6.9.0\mingw_64\bin;C:\Qt\Tools\mingw1310_64\bin;%PATH%

echo 检查可执行文件...
if exist "bin\VideoCallFrontend.exe" (
    echo ✓ VideoCallFrontend.exe 存在
) else (
    echo ❌ VideoCallFrontend.exe 不存在
    pause
    exit /b 1
)

echo.
echo 复制Qt DLL文件...
copy "C:\Qt\6.9.0\mingw_64\bin\Qt6Core.dll" bin\ >nul 2>&1
copy "C:\Qt\6.9.0\mingw_64\bin\Qt6Gui.dll" bin\ >nul 2>&1
copy "C:\Qt\6.9.0\mingw_64\bin\Qt6Widgets.dll" bin\ >nul 2>&1
copy "C:\Qt\6.9.0\mingw_64\bin\Qt6Network.dll" bin\ >nul 2>&1

echo.
echo 启动应用程序...
echo 如果程序闪退，请查看错误信息：
echo.

cd bin
VideoCallFrontend.exe

echo.
echo 程序退出，退出代码: %errorlevel%
echo.
echo 如果程序闪退，可能的原因：
echo 1. Qt DLL文件缺失或版本不匹配
echo 2. 缺少必要的系统库
echo 3. 程序依赖的其他库缺失
echo 4. 程序内部错误
echo.
pause 