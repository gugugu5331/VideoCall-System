@echo off
chcp 65001 >nul
echo 正在启动前端应用并捕获错误信息...
echo.

cd /d "%~dp0"

echo 设置环境变量...
set PATH=C:\Qt\6.9.0\mingw_64\bin;%PATH%

echo 检查Qt DLL文件...
if exist "C:\Qt\6.9.0\mingw_64\bin\Qt6Core.dll" (
    echo ✓ Qt6Core.dll 存在
) else (
    echo ❌ Qt6Core.dll 不存在
)

if exist "C:\Qt\6.9.0\mingw_64\bin\Qt6Gui.dll" (
    echo ✓ Qt6Gui.dll 存在
) else (
    echo ❌ Qt6Gui.dll 不存在
)

if exist "C:\Qt\6.9.0\mingw_64\bin\Qt6Widgets.dll" (
    echo ✓ Qt6Widgets.dll 存在
) else (
    echo ❌ Qt6Widgets.dll 不存在
)

echo.
echo 复制Qt DLL到bin目录...
copy "C:\Qt\6.9.0\mingw_64\bin\Qt6Core.dll" bin\ >nul 2>&1
copy "C:\Qt\6.9.0\mingw_64\bin\Qt6Gui.dll" bin\ >nul 2>&1
copy "C:\Qt\6.9.0\mingw_64\bin\Qt6Widgets.dll" bin\ >nul 2>&1

echo.
echo 启动应用程序...
echo 如果程序闪退，请查看下面的错误信息：
echo.

bin\VideoCallFrontend.exe 2>&1

echo.
echo 程序退出，退出代码: %errorlevel%
echo.
pause 