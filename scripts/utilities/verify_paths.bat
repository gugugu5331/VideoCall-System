@echo off
chcp 65001 >nul
title VideoCall System - Path Verification

echo ==========================================
echo VideoCall System - 路径验证
echo ==========================================
echo.

echo [1/4] 验证测试脚本路径...
if exist "%~dp0\..\testing\run_all_tests.py" (
    echo ✅ run_all_tests.py 路径正确
) else (
    echo ❌ run_all_tests.py 路径错误
)

if exist "%~dp0\..\testing\test_api.py" (
    echo ✅ test_api.py 路径正确
) else (
    echo ❌ test_api.py 路径错误
)

if exist "%~dp0\..\testing\check_database.py" (
    echo ✅ check_database.py 路径正确
) else (
    echo ❌ check_database.py 路径错误
)

if exist "%~dp0\..\testing\check_docker.py" (
    echo ✅ check_docker.py 路径正确
) else (
    echo ❌ check_docker.py 路径错误
)

echo.
echo [2/4] 验证后端脚本路径...
if exist "%~dp0\..\..\core\backend\start-basic.bat" (
    echo ✅ start-basic.bat 路径正确
) else (
    echo ❌ start-basic.bat 路径错误
)

if exist "%~dp0\..\..\core\backend\main-basic.go" (
    echo ✅ main-basic.go 路径正确
) else (
    echo ❌ main-basic.go 路径错误
)

echo.
echo [3/4] 验证AI服务脚本路径...
if exist "%~dp0\..\..\core\ai-service\start_ai_manual.bat" (
    echo ✅ start_ai_manual.bat 路径正确
) else (
    echo ❌ start_ai_manual.bat 路径错误
)

if exist "%~dp0\..\..\core\ai-service\main-simple.py" (
    echo ✅ main-simple.py 路径正确
) else (
    echo ❌ main-simple.py 路径错误
)

echo.
echo [4/4] 验证配置文件路径...
if exist "%~dp0\..\..\config\docker-compose.yml" (
    echo ✅ docker-compose.yml 路径正确
) else (
    echo ❌ docker-compose.yml 路径错误
)

echo.
echo ==========================================
echo 路径验证完成
echo ==========================================
echo.
echo 如果所有路径都正确，您可以:
echo 1. 运行 .\quick_start.bat 一键启动
echo 2. 运行 .\quick_manage.bat 使用管理菜单
echo 3. 运行 .\quick_test.bat 运行测试
echo.
pause 