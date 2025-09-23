@echo off
chcp 65001 >nul
title VideoCall System - System Management

:menu
cls
echo ==========================================
echo VideoCall System - System Management
echo ==========================================
echo.
echo Available operations:
echo 1. Start all services
echo 2. Start database services
echo 3. Start backend service
echo 4. Start AI service
echo 5. Run full tests
echo 6. Run quick tests
echo 7. Check database status
echo 8. Check Docker status
echo 9. Stop all services
echo 10. Release all ports
echo 11. Release specific port
echo 0. Exit
echo.
set /p choice="Enter your choice (0-11): "

if "%choice%"=="1" goto start_all
if "%choice%"=="2" goto start_db
if "%choice%"=="3" goto start_backend
if "%choice%"=="4" goto start_ai
if "%choice%"=="5" goto run_full_test
if "%choice%"=="6" goto run_quick_test
if "%choice%"=="7" goto check_db
if "%choice%"=="8" goto check_docker
if "%choice%"=="9" goto stop_all
if "%choice%"=="10" goto release_all_ports
if "%choice%"=="11" goto release_specific_port
if "%choice%"=="0" goto exit
echo Invalid choice, please try again.
timeout /t 2 >nul
goto menu

:start_all
echo.
echo Starting all services...
docker-compose --project-name videocall-system -f config/docker-compose.yml up -d postgres redis
timeout /t 3 >nul
start "Backend Service" cmd /k "cd /d %~dp0\..\..\core\backend && start-basic.bat"
timeout /t 2 >nul
start "AI Service" cmd /k "cd /d %~dp0\..\..\core\ai-service && start_ai_manual.bat"
echo OK: All services starting...
pause
goto menu

:start_db
echo.
echo Starting database services...
docker-compose --project-name videocall-system -f config/docker-compose.yml up -d postgres redis
echo OK: Database services started
pause
goto menu

:start_backend
echo.
echo Starting backend service...
start "Backend Service" cmd /k "cd /d %~dp0\..\..\core\backend && start-basic.bat"
echo OK: Backend service starting...
pause
goto menu

:start_ai
echo.
echo Starting AI service...
start "AI Service" cmd /k "cd /d %~dp0\..\..\core\ai-service && start_ai_manual.bat"
echo OK: AI service starting...
pause
goto menu

:run_full_test
echo.
echo Running full tests...
python %~dp0\..\testing\run_all_tests.py
pause
goto menu

:run_quick_test
echo.
echo Running quick tests...
python %~dp0\..\testing\test_api.py
pause
goto menu

:check_db
echo.
echo Checking database status...
python %~dp0\..\testing\check_database.py
pause
goto menu

:check_docker
echo.
echo Checking Docker status...
python %~dp0\..\testing\check_docker.py
pause
goto menu

:stop_all
echo.
echo Stopping all services...
docker-compose --project-name videocall-system down
echo OK: All services stopped
pause
goto menu

:release_all_ports
echo.
echo Releasing all ports...
python %~dp0\release_ports.py
pause
goto menu

:release_specific_port
echo.
set /p port="Enter port number to release: "
if "%port%"=="" (
    echo Port number cannot be empty
    pause
    goto menu
)
echo Releasing port %port%...
python %~dp0\release_ports.py %port%
pause
goto menu

:exit
echo.
echo Exiting system management...
exit /b 0 