@echo off
echo 编译简化视频处理测试程序...

REM 检查是否有g++编译器
where g++ >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到g++编译器
    echo 请安装MinGW-w64或MSYS2
    echo 或者使用Visual Studio Developer Command Prompt
    pause
    exit /b 1
)

REM 检查是否有OpenCV
echo 检查OpenCV安装...

REM 尝试编译一个简单的OpenCV测试
echo #include ^<opencv2/opencv.hpp^> > test_opencv.cpp
echo int main() { std::cout ^<^< CV_VERSION ^<^< std::endl; return 0; } >> test_opencv.cpp

g++ test_opencv.cpp -lopencv_core -lopencv_imgproc -lopencv_highgui -lopencv_imgcodecs -lopencv_videoio -lopencv_objdetect -o test_opencv.exe 2>nul

if %errorlevel% neq 0 (
    echo 错误: OpenCV库未正确安装或配置
    echo 请确保:
    echo 1. 已安装OpenCV
    echo 2. 库文件在系统PATH中
    echo 3. 头文件可以被找到
    del test_opencv.cpp 2>nul
    pause
    exit /b 1
)

del test_opencv.cpp test_opencv.exe 2>nul
echo OpenCV检查通过

REM 编译主程序
echo 编译主程序...
g++ -std=c++17 ^
    simple_test.cpp ^
    -o simple_video_test.exe ^
    -lopencv_core ^
    -lopencv_imgproc ^
    -lopencv_highgui ^
    -lopencv_imgcodecs ^
    -lopencv_videoio ^
    -lopencv_objdetect ^
    -O3 ^
    -Wall ^
    -Wextra

if %errorlevel% equ 0 (
    echo 编译成功!
    echo 运行程序: simple_video_test.exe
    echo.
    echo 控制说明:
    echo   ESC - 退出
    echo   SPACE - 截图
    echo   1-7 - 不同滤镜
    echo   0 - 移除滤镜
    echo   F - 切换人脸检测
    echo   +/- - 调整滤镜强度
    echo.
    pause
) else (
    echo 编译失败!
    pause
    exit /b 1
)
