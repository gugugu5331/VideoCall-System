#!/usr/bin/env pwsh
# Use Qt's windeployqt tool to deploy all dependencies

Write-Host "=== Using windeployqt to Deploy Qt Dependencies ===" -ForegroundColor Cyan
Write-Host ""

$qtPath = "E:\Qt\6.5.3\msvc2019_64"
$buildPath = "build\bin\Release"
$exePath = Join-Path $buildPath "MeetingSystemClient.exe"

# Check if Qt path exists
if (-not (Test-Path $qtPath)) {
    Write-Host "[ERROR] Qt6 path not found: $qtPath" -ForegroundColor Red
    exit 1
}

# Check if executable exists
if (-not (Test-Path $exePath)) {
    Write-Host "[ERROR] Executable not found: $exePath" -ForegroundColor Red
    exit 1
}

$windeployqt = Join-Path $qtPath "bin\windeployqt.exe"

if (-not (Test-Path $windeployqt)) {
    Write-Host "[ERROR] windeployqt not found: $windeployqt" -ForegroundColor Red
    exit 1
}

Write-Host "Qt Path: $qtPath" -ForegroundColor Green
Write-Host "Executable: $exePath" -ForegroundColor Green
Write-Host "windeployqt: $windeployqt" -ForegroundColor Green
Write-Host ""

Write-Host "Running windeployqt..." -ForegroundColor Yellow
Write-Host ""

# Run windeployqt with QML support
& $windeployqt --qmldir qml --release $exePath

Write-Host ""
Write-Host "=== Done ===" -ForegroundColor Green
Write-Host ""

