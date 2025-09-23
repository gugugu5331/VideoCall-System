@echo off
chcp 65001 >nul
echo ==========================================
echo VideoCall Backend Status Check
echo ==========================================

echo.
echo Checking backend service...

REM Test health check
curl -s http://localhost:8000/health > temp_health.json
if %errorlevel% equ 0 (
    echo [OK] Backend is running
    type temp_health.json
) else (
    echo [ERROR] Backend is not running
    goto :end
)

echo.
echo Testing API endpoints...

REM Test root endpoint
curl -s http://localhost:8000/ > temp_root.json
if %errorlevel% equ 0 (
    echo [OK] Root endpoint working
) else (
    echo [ERROR] Root endpoint failed
)

REM Test user registration
curl -s -X POST http://localhost:8000/api/v1/auth/register ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"batch_test_user\",\"email\":\"batch@example.com\",\"password\":\"password123\",\"full_name\":\"Batch Test User\"}" > temp_register.json
if %errorlevel% equ 0 (
    echo [OK] User registration working
) else (
    echo [ERROR] User registration failed
)

REM Test user login
curl -s -X POST http://localhost:8000/api/v1/auth/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"testuser\",\"password\":\"password123\"}" > temp_login.json
if %errorlevel% equ 0 (
    echo [OK] User login working
    echo Token received successfully
) else (
    echo [ERROR] User login failed
)

echo.
echo ==========================================
echo Status check completed
echo ==========================================

:end
REM Clean up temporary files
del temp_*.json 2>nul
pause 