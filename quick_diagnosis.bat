@echo off
chcp 65001 >nul
title 系统诊断

echo ==========================================
echo 系统诊断工具
echo ==========================================
echo.

echo 1. 检查端口占用情况...
echo.
echo 端口 8000 (后端服务):
netstat -an | findstr :8000
echo.

echo 端口 5432 (PostgreSQL):
netstat -an | findstr :5432
echo.

echo 端口 6379 (Redis):
netstat -an | findstr :6379
echo.

echo 端口 5000 (AI服务):
netstat -an | findstr :5000
echo.

echo.
echo 2. 检查Docker状态...
docker ps
echo.

echo 3. 检查Go环境...
go version
echo.

echo 4. 检查Python环境...
python --version
echo.

echo 5. 测试后端连接...
python test_simple_api.py
echo.

echo.
echo 诊断完成！
pause 