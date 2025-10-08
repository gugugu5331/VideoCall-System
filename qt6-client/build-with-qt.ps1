# Build Script with Qt6 from E drive

Write-Host "=== Qt6 Project Build Script ===" -ForegroundColor Cyan
Write-Host ""

# Qt6 path
$qt6Path = "E:\Qt\6.5.3\msvc2019_64"
$vcpkgToolchain = "C:\vcpkg\scripts\buildsystems\vcpkg.cmake"

# Verify paths
Write-Host "Checking paths..." -ForegroundColor Cyan
if (-not (Test-Path $qt6Path)) {
    Write-Host "ERROR: Qt6 not found at $qt6Path" -ForegroundColor Red
    exit 1
}
Write-Host "  [OK] Qt6: $qt6Path" -ForegroundColor Green

if (-not (Test-Path $vcpkgToolchain)) {
    Write-Host "ERROR: vcpkg toolchain not found" -ForegroundColor Red
    exit 1
}
Write-Host "  [OK] vcpkg: $vcpkgToolchain" -ForegroundColor Green

Write-Host ""

# Create build directory
if (-not (Test-Path "build")) {
    New-Item -ItemType Directory -Path "build" | Out-Null
    Write-Host "Created build directory" -ForegroundColor Green
} else {
    Write-Host "Build directory exists" -ForegroundColor Yellow
}

Write-Host ""

# Step 1: Configure with CMake
Write-Host "=== Step 1: CMake Configuration ===" -ForegroundColor Cyan
Write-Host ""

Set-Location build

Write-Host "Running CMake..." -ForegroundColor Cyan
Write-Host "cmake -G `"Visual Studio 17 2022`" -A x64 -DCMAKE_PREFIX_PATH=`"$qt6Path`" -DCMAKE_TOOLCHAIN_FILE=`"$vcpkgToolchain`" .." -ForegroundColor Gray
Write-Host ""

& cmake -G "Visual Studio 17 2022" -A x64 -DCMAKE_PREFIX_PATH="$qt6Path" -DCMAKE_TOOLCHAIN_FILE="$vcpkgToolchain" ..

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "=== CMake Configuration Failed ===" -ForegroundColor Red
    Write-Host "Exit code: $LASTEXITCODE" -ForegroundColor Red
    Set-Location ..
    exit 1
}

Write-Host ""
Write-Host "=== CMake Configuration Successful ===" -ForegroundColor Green
Write-Host ""

# Step 2: Build with MSBuild
Write-Host "=== Step 2: Building Project ===" -ForegroundColor Cyan
Write-Host ""

Write-Host "Running build..." -ForegroundColor Cyan
Write-Host "cmake --build . --config Release --parallel" -ForegroundColor Gray
Write-Host ""

& cmake --build . --config Release --parallel

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "=== Build Failed ===" -ForegroundColor Red
    Write-Host "Exit code: $LASTEXITCODE" -ForegroundColor Red
    Set-Location ..
    exit 1
}

Write-Host ""
Write-Host "=== Build Successful ===" -ForegroundColor Green
Write-Host ""

Set-Location ..

# Check output
$exePath = "build\bin\Release\MeetingSystemClient.exe"
if (Test-Path $exePath) {
    $size = (Get-Item $exePath).Length / 1MB
    Write-Host "Executable created: $exePath" -ForegroundColor Green
    Write-Host "Size: $([math]::Round($size, 2)) MB" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "=== Build Complete ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "To run the application:" -ForegroundColor Cyan
    Write-Host "  cd build\bin\Release" -ForegroundColor Yellow
    Write-Host "  .\MeetingSystemClient.exe" -ForegroundColor Yellow
    Write-Host ""
} else {
    Write-Host "WARNING: Executable not found at expected location" -ForegroundColor Yellow
    Write-Host "Check build\bin\ directory" -ForegroundColor Yellow
}

