@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo ========================================
echo    Anaconda Qt 库路径修复工具
echo ========================================
echo.

:: 检查Anaconda Qt库文件
echo 检查Anaconda Qt库文件...
set "qt_lib_path=C:\ProgramData\anaconda3\Library\lib"
set "qt_bin_path=C:\ProgramData\anaconda3\Library\bin"

if exist "%qt_lib_path%\libQt5Widgets_conda.a" (
    echo ✓ 找到 libQt5Widgets_conda.a
) else (
    echo ❌ 未找到 libQt5Widgets_conda.a
    echo 检查其他可能的库文件...
    
    if exist "%qt_lib_path%\Qt5Widgets.lib" (
        echo ✓ 找到 Qt5Widgets.lib (MSVC版本)
        set "qt_lib_type=msvc"
    ) else (
        if exist "%qt_lib_path%\libQt5Widgets.a" (
            echo ✓ 找到 libQt5Widgets.a (MinGW版本)
            set "qt_lib_type=mingw"
        ) else (
            echo ❌ 未找到Qt5库文件
            echo 请检查Anaconda Qt安装
            pause
            exit /b 1
        )
    )
)

:: 检查Qt DLL文件
echo.
echo 检查Qt DLL文件...
if exist "%qt_bin_path%\Qt5Widgets.dll" (
    echo ✓ 找到 Qt5Widgets.dll
) else (
    echo ❌ 未找到 Qt5Widgets.dll
)

:: 创建修复后的项目文件
echo.
echo 创建修复后的项目文件...

:: 备份原文件
if exist VideoCallApp_anaconda.pro (
    copy VideoCallApp_anaconda.pro VideoCallApp_anaconda.pro.backup
    echo ✓ 已备份原项目文件
)

:: 创建修复后的项目文件
(
echo QT += core gui widgets
echo.
echo greaterThan(QT_MAJOR_VERSION, 4^): QT += widgets
echo.
echo CONFIG += c++14
echo.
echo # Anaconda Qt修复版本
echo SOURCES += \
echo     main.cpp \
echo     mainwindow.cpp \
echo     loginwidget.cpp
echo.
echo HEADERS += \
echo     mainwindow.h \
echo     loginwidget.h
echo.
echo FORMS += \
echo     mainwindow.ui \
echo     loginwidget.ui
echo.
echo # 资源文件
echo RESOURCES += \
echo     resources.qrc
echo.
echo # 编译配置
echo CONFIG(debug, debug^|release^) {
echo     DESTDIR = debug
echo } else {
echo     DESTDIR = release
echo }
echo.
echo # Anaconda Qt库路径
echo INCLUDEPATH += "%qt_lib_path%\..\include"
echo INCLUDEPATH += "%qt_lib_path%\..\include\QtWidgets"
echo INCLUDEPATH += "%qt_lib_path%\..\include\QtGui"
echo INCLUDEPATH += "%qt_lib_path%\..\include\QtCore"
echo.
echo # 库文件路径
echo LIBS += -L"%qt_lib_path%"
echo.
echo # Windows特定配置
echo win32 {
echo     LIBS += -lws2_32 -liphlpapi
echo }
echo.
echo # 定义
echo DEFINES += \
echo     QT_DEPRECATED_WARNINGS \
echo     VIDEO_CALL_APP_VERSION=\"1.0.0\" \
echo     SIMPLE_MODE
) > VideoCallApp_anaconda_fixed.pro

echo ✓ 已创建修复后的项目文件: VideoCallApp_anaconda_fixed.pro

:: 清理之前的构建
echo.
echo 清理之前的构建文件...
if exist Makefile del Makefile
if exist Makefile.Debug del Makefile.Debug
if exist Makefile.Release del Makefile.Release
if exist debug rmdir /s /q debug 2>nul
if exist release rmdir /s /q release 2>nul
if exist .qmake.stash del .qmake.stash 2>nul

:: 创建必要的目录
echo 创建构建目录...
if not exist debug mkdir debug
if not exist release mkdir release
if not exist bin mkdir bin

:: 生成Makefile
echo 生成Makefile...
qmake VideoCallApp_anaconda_fixed.pro -spec win32-g++ CONFIG+=debug
if %errorlevel% neq 0 (
    echo 错误: qmake失败
    pause
    exit /b 1
)

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
        pause
        exit /b 1
    )
)

:: 编译项目
echo.
echo 开始编译项目（修复版本）...
echo 这可能需要几分钟时间，请耐心等待...
echo.

%make_cmd% -j%NUMBER_OF_PROCESSORS%
if %errorlevel% neq 0 (
    echo.
    echo 错误: 编译失败
    echo.
    echo 建议尝试以下解决方案:
    echo 1. 使用cmd而不是PowerShell
    echo 2. 安装标准Qt而不是Anaconda Qt
    echo 3. 使用conda重新安装Qt
    echo.
    pause
    exit /b 1
)

:: 复制可执行文件
echo.
echo 复制可执行文件...
if exist debug\VideoCallApp.exe (
    copy debug\VideoCallApp.exe bin\VideoCallApp_fixed.exe
    echo 修复版本构建完成: bin\VideoCallApp_fixed.exe
) else (
    echo 构建失败，未找到可执行文件
    pause
    exit /b 1
)

echo.
echo ========================================
echo    Anaconda Qt修复完成！
echo ========================================
echo.
echo 修复版本功能:
echo - 基本界面框架
echo - 登录界面
echo - 主窗口
echo.
echo 运行修复版本:
echo   bin\VideoCallApp_fixed.exe
echo.

pause 