# Qt6 Configuration Script for Windows
# This script helps configure the build environment for Qt6, OpenCV, and OpenGL

Write-Host "=== Qt6 Build Configuration ===" -ForegroundColor Cyan
Write-Host ""

# ============================================================================
# Configuration
# ============================================================================

$VCPKG_ROOT = "C:\vcpkg"
$USE_VCPKG = Test-Path $VCPKG_ROOT

# Common Qt6 installation paths
$qt6Paths = @(
    "C:\Qt\6.5.0\msvc2019_64",
    "C:\Qt\6.5.1\msvc2019_64",
    "C:\Qt\6.5.2\msvc2019_64",
    "C:\Qt\6.5.3\msvc2019_64",
    "C:\Qt\6.6.0\msvc2019_64",
    "C:\Qt\6.6.1\msvc2019_64",
    "C:\Qt\6.7.0\msvc2019_64",
    "C:\Qt\6.7.1\msvc2019_64",
    "C:\Qt\6.7.2\msvc2019_64",
    "C:\Qt\6.8.0\msvc2019_64",
    "C:\Qt\6.9.0\msvc2019_64",
    "C:\Qt\6.10.0\msvc2019_64",
    "C:\Qt\6.5.0\mingw_64",
    "C:\Qt\6.6.0\mingw_64",
    "C:\Qt\6.7.0\mingw_64",
    "C:\Qt\6.8.0\mingw_64"
)

# Try to find Qt6
$qt6Path = $null
foreach ($path in $qt6Paths) {
    if (Test-Path $path) {
        $qt6Path = $path
        Write-Host "Found Qt6 at: $qt6Path" -ForegroundColor Green
        break
    }
}

if ($null -eq $qt6Path) {
    Write-Host "Qt6 not found in common locations!" -ForegroundColor Red
    Write-Host "Please install Qt6 or specify the path manually:" -ForegroundColor Yellow
    Write-Host "  1. Download Qt6 from https://www.qt.io/download" -ForegroundColor Yellow
    Write-Host "  2. Or set CMAKE_PREFIX_PATH manually:" -ForegroundColor Yellow
    Write-Host "     cmake -DCMAKE_PREFIX_PATH=C:\Path\To\Qt6 .." -ForegroundColor Yellow
    exit 1
}

# Set environment variable
$env:CMAKE_PREFIX_PATH = $qt6Path
Write-Host "Set CMAKE_PREFIX_PATH=$qt6Path" -ForegroundColor Green

# Create build directory
if (Test-Path "build") {
    Write-Host "Removing existing build directory..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force "build"
}

Write-Host "Creating build directory..." -ForegroundColor Cyan
New-Item -ItemType Directory -Path "build" | Out-Null

# Check for OpenCV
Write-Host "`n=== Checking Dependencies ===" -ForegroundColor Cyan

$opencvFound = $false
$opencvPath = $null

# Check vcpkg
if (Test-Path $VCPKG_ROOT) {
    Write-Host "Checking vcpkg for OpenCV..." -ForegroundColor Cyan
    Set-Location $VCPKG_ROOT
    $vcpkgList = .\vcpkg list 2>$null | Select-String "opencv4:x64-windows"
    Set-Location $PSScriptRoot

    if ($vcpkgList) {
        Write-Host "✓ OpenCV found in vcpkg" -ForegroundColor Green
        $opencvFound = $true
        $USE_VCPKG = $true
    }
}

# Check prebuilt OpenCV
if (-not $opencvFound) {
    $opencvPaths = @("C:\opencv\build", "C:\OpenCV\build")
    foreach ($path in $opencvPaths) {
        if (Test-Path $path) {
            Write-Host "✓ OpenCV found at: $path" -ForegroundColor Green
            $opencvPath = $path
            $opencvFound = $true
            $USE_VCPKG = $false
            break
        }
    }
}

if (-not $opencvFound) {
    Write-Host "✗ OpenCV not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install OpenCV first:" -ForegroundColor Yellow
    Write-Host "  .\install-opencv-opengl.ps1" -ForegroundColor Cyan
    Write-Host ""
    exit 1
}

# Run CMake
Write-Host "`n=== Running CMake ===" -ForegroundColor Cyan
Set-Location "build"

if ($USE_VCPKG) {
    # Use vcpkg toolchain
    $toolchainFile = "$VCPKG_ROOT\scripts\buildsystems\vcpkg.cmake"
    $cmakeCmd = "cmake -DCMAKE_PREFIX_PATH=`"$qt6Path`" -DCMAKE_TOOLCHAIN_FILE=`"$toolchainFile`" .."
    Write-Host "Using vcpkg toolchain" -ForegroundColor Cyan
} else {
    # Use OpenCV_DIR
    $cmakeCmd = "cmake -DCMAKE_PREFIX_PATH=`"$qt6Path`" -DOpenCV_DIR=`"$opencvPath`" .."
    Write-Host "Using OpenCV at: $opencvPath" -ForegroundColor Cyan
}

Write-Host "Executing: $cmakeCmd" -ForegroundColor Gray
Invoke-Expression $cmakeCmd

if ($LASTEXITCODE -ne 0) {
    Write-Host "`nCMake configuration failed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Troubleshooting:" -ForegroundColor Yellow
    Write-Host "1. Make sure Qt6 is installed" -ForegroundColor Yellow
    Write-Host "2. Make sure OpenCV is installed" -ForegroundColor Yellow
    Write-Host "3. Check CMake output for errors" -ForegroundColor Yellow
    Write-Host ""
    Set-Location ..
    exit 1
}

Write-Host "`n=== CMake Configuration Successful ===" -ForegroundColor Green
Write-Host ""
Write-Host "Configuration summary:" -ForegroundColor Cyan
Write-Host "  Qt6: $qt6Path" -ForegroundColor Gray
if ($USE_VCPKG) {
    Write-Host "  OpenCV: vcpkg ($VCPKG_ROOT)" -ForegroundColor Gray
} else {
    Write-Host "  OpenCV: $opencvPath" -ForegroundColor Gray
}
Write-Host ""
Write-Host "To build the project, run:" -ForegroundColor Cyan
Write-Host "  cd build" -ForegroundColor Yellow
Write-Host "  cmake --build . --config Release" -ForegroundColor Yellow
Write-Host "`nOr run:" -ForegroundColor Cyan
Write-Host "  .\build.ps1" -ForegroundColor Yellow
Write-Host ""

Set-Location ..

