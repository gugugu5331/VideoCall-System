# Qt6 Build Script for Windows
# This script builds the project after configuration

Write-Host "=== Qt6 Project Build ===" -ForegroundColor Cyan

# Check if build directory exists
if (-not (Test-Path "build")) {
    Write-Host "Build directory not found!" -ForegroundColor Red
    Write-Host "Please run configure.ps1 first:" -ForegroundColor Yellow
    Write-Host "  .\configure.ps1" -ForegroundColor Yellow
    exit 1
}

# Check if CMakeCache.txt exists
if (-not (Test-Path "build\CMakeCache.txt")) {
    Write-Host "CMake not configured!" -ForegroundColor Red
    Write-Host "Please run configure.ps1 first:" -ForegroundColor Yellow
    Write-Host "  .\configure.ps1" -ForegroundColor Yellow
    exit 1
}

# Build the project
Write-Host "`n=== Building Project ===" -ForegroundColor Cyan
Set-Location "build"

$buildCmd = "cmake --build . --config Release"
Write-Host "Executing: $buildCmd" -ForegroundColor Gray
Invoke-Expression $buildCmd

if ($LASTEXITCODE -ne 0) {
    Write-Host "`nBuild failed!" -ForegroundColor Red
    Set-Location ..
    exit 1
}

Write-Host "`n=== Build Successful ===" -ForegroundColor Green
Write-Host "`nExecutable location:" -ForegroundColor Cyan
Write-Host "  build\bin\Release\MeetingSystemClient.exe" -ForegroundColor Yellow

Write-Host "`nTo run the application:" -ForegroundColor Cyan
Write-Host "  cd build\bin\Release" -ForegroundColor Yellow
Write-Host "  .\MeetingSystemClient.exe" -ForegroundColor Yellow

Set-Location ..

