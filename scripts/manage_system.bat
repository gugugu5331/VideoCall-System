@echo off
chcp 65001 >nul
title VideoCall System - Management Script

:menu
cls
echo ================================================================
echo VideoCall System - Management Script
echo ================================================================
echo.
echo Project Root: %~dp0..
echo Please select an option:
echo.
echo [1] Start All Services
echo [2] Stop All Services
echo [3] Restart All Services
echo [4] Test All Services
echo [5] Show Service Status
echo [6] Clean Port Usage
echo [7] Open Frontend
echo [8] Show Help
echo [0] Exit
echo.
echo ================================================================
echo.

set /p choice=Enter option (0-8): 

if "%choice%"=="1" goto start
if "%choice%"=="2" goto stop
if "%choice%"=="3" goto restart
if "%choice%"=="4" goto test
if "%choice%"=="5" goto status
if "%choice%"=="6" goto clean
if "%choice%"=="7" goto open_frontend
if "%choice%"=="8" goto help
if "%choice%"=="0" goto exit
goto menu

:start
echo.
echo Starting all services...
powershell -ExecutionPolicy Bypass -File "%~dp0simple_manage.ps1" start
echo.
pause
goto menu

:stop
echo.
echo Stopping all services...
powershell -ExecutionPolicy Bypass -File "%~dp0simple_manage.ps1" stop
echo.
pause
goto menu

:restart
echo.
echo Restarting all services...
powershell -ExecutionPolicy Bypass -File "%~dp0simple_manage.ps1" restart
echo.
pause
goto menu

:test
echo.
echo Testing all services...
powershell -ExecutionPolicy Bypass -File "%~dp0simple_manage.ps1" test
echo.
pause
goto menu

:status
echo.
echo Showing service status...
powershell -ExecutionPolicy Bypass -File "%~dp0simple_manage.ps1" status
echo.
pause
goto menu

:clean
echo.
echo Cleaning port usage...
powershell -ExecutionPolicy Bypass -File "%~dp0simple_manage.ps1" clean
echo.
pause
goto menu

:open_frontend
echo.
echo Opening frontend...
start http://localhost:8080
echo Frontend opened in browser
echo.
pause
goto menu

:help
echo.
echo Help Information:
echo.
echo Service Ports:
echo   - Backend Service: 8000
echo   - Frontend Service: 8080
echo   - AI Service: 5001
echo   - Database: 5432
echo   - Redis: 6379
echo.
echo Access URLs:
echo   - Frontend: http://localhost:8080
echo   - Backend API: http://localhost:8000
echo   - Health Check: http://localhost:8000/health
echo.
echo Usage Instructions:
echo   - First time: Select [6] to clean ports
echo   - Then select [1] to start all services
echo   - Use [7] to open frontend interface
echo   - Use [4] to test if services are running
echo.
pause
goto menu

:exit
echo.
echo Thank you for using VideoCall System!
echo.
exit /b 0 