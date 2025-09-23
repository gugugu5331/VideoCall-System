@echo off
chcp 65001 >nul
echo ==========================================
echo VideoCall Database Management
echo ==========================================

:menu
echo.
echo Select an option:
echo 1. Start database services
echo 2. Stop database services
echo 3. Check database status
echo 4. View database logs
echo 5. Test database connection
echo 6. Exit
echo.
set /p choice="Enter your choice (1-6): "

if "%choice%"=="1" goto start_db
if "%choice%"=="2" goto stop_db
if "%choice%"=="3" goto check_status
if "%choice%"=="4" goto view_logs
if "%choice%"=="5" goto test_connection
if "%choice%"=="6" goto exit
echo Invalid choice. Please try again.
goto menu

:start_db
echo.
echo Starting database services...
docker-compose --project-name videocall-system up -d postgres redis
echo Database services started.
pause
goto menu

:stop_db
echo.
echo Stopping database services...
docker-compose --project-name videocall-system down
echo Database services stopped.
pause
goto menu

:check_status
echo.
echo Checking database status...
docker-compose --project-name videocall-system ps
echo.
echo Testing backend connection...
python test_backend.py
pause
goto menu

:view_logs
echo.
echo Select logs to view:
echo 1. PostgreSQL logs
echo 2. Redis logs
echo 3. Back to main menu
set /p log_choice="Enter choice (1-3): "

if "%log_choice%"=="1" (
    echo PostgreSQL logs:
    docker-compose --project-name videocall-system logs postgres
    pause
    goto view_logs
)
if "%log_choice%"=="2" (
    echo Redis logs:
    docker-compose --project-name videocall-system logs redis
    pause
    goto view_logs
)
if "%log_choice%"=="3" goto menu
echo Invalid choice.
pause
goto view_logs

:test_connection
echo.
echo Testing database connections...
python check_database.py
pause
goto menu

:exit
echo.
echo Exiting database management...
exit /b 0 