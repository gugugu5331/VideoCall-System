# Complete Setup and Build Script
# This script installs Qt6, configures, and builds the project

Write-Host "=== Qt6 Video Meeting Client - Complete Setup ===" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check if Qt6 is installed
Write-Host "Step 1: Checking Qt6 installation..." -ForegroundColor Cyan

$qt6Paths = @(
    "C:\Qt\6.5.3\msvc2019_64",
    "C:\Qt\6.5.0\msvc2019_64",
    "C:\Qt\6.5.1\msvc2019_64",
    "C:\Qt\6.5.2\msvc2019_64",
    "C:\Qt\6.6.0\msvc2019_64",
    "C:\Qt\6.7.0\msvc2019_64",
    "C:\Qt\6.8.0\msvc2019_64",
    "C:\Qt\6.9.0\msvc2019_64",
    "C:\Qt\6.10.0\msvc2019_64"
)

$qt6Path = $null
foreach ($path in $qt6Paths) {
    if (Test-Path "$path\bin\qmake.exe") {
        $qt6Path = $path
        Write-Host "Found Qt6 at: $qt6Path" -ForegroundColor Green
        break
    }
}

if ($null -eq $qt6Path) {
    Write-Host "Qt6 not found!" -ForegroundColor Yellow
    Write-Host ""
    $response = Read-Host "Do you want to install Qt6 now? (Y/n)"
    
    if ($response -eq "n" -or $response -eq "N") {
        Write-Host "Installation cancelled." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Please install Qt6 manually and run this script again." -ForegroundColor Yellow
        Write-Host "Or run: .\install-qt6.ps1" -ForegroundColor Cyan
        exit 1
    }
    
    Write-Host ""
    Write-Host "Installing Qt6..." -ForegroundColor Cyan
    
    # Run install-qt6.ps1
    if (Test-Path "install-qt6.ps1") {
        & ".\install-qt6.ps1"
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Qt6 installation failed!" -ForegroundColor Red
            exit 1
        }
        
        # Check again after installation
        $qt6Path = "C:\Qt\6.5.3\msvc2019_64"
        if (-not (Test-Path "$qt6Path\bin\qmake.exe")) {
            Write-Host "Qt6 installation verification failed!" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "install-qt6.ps1 not found!" -ForegroundColor Red
        Write-Host "Please run: .\install-qt6.ps1" -ForegroundColor Yellow
        exit 1
    }
}

Write-Host ""
Write-Host "Qt6 Path: $qt6Path" -ForegroundColor Green

# Step 2: Configure project
Write-Host ""
Write-Host "Step 2: Configuring project..." -ForegroundColor Cyan

# Set environment variable
$env:CMAKE_PREFIX_PATH = $qt6Path

# Create build directory
if (Test-Path "build") {
    Write-Host "Removing existing build directory..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force "build"
}

Write-Host "Creating build directory..." -ForegroundColor Cyan
New-Item -ItemType Directory -Path "build" | Out-Null

# Run CMake
Write-Host "Running CMake..." -ForegroundColor Cyan
Set-Location "build"

$cmakeCmd = "cmake -DCMAKE_PREFIX_PATH=`"$qt6Path`" .."
Write-Host "Executing: $cmakeCmd" -ForegroundColor Gray
Invoke-Expression $cmakeCmd

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "CMake configuration failed!" -ForegroundColor Red
    Set-Location ..
    exit 1
}

Write-Host ""
Write-Host "CMake configuration successful!" -ForegroundColor Green

# Step 3: Build project
Write-Host ""
Write-Host "Step 3: Building project..." -ForegroundColor Cyan
Write-Host "This may take 5-10 minutes..." -ForegroundColor Yellow
Write-Host ""

$buildCmd = "cmake --build . --config Release"
Write-Host "Executing: $buildCmd" -ForegroundColor Gray
Invoke-Expression $buildCmd

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "Build failed!" -ForegroundColor Red
    Set-Location ..
    exit 1
}

Set-Location ..

# Step 4: Verify build
Write-Host ""
Write-Host "Step 4: Verifying build..." -ForegroundColor Cyan

$exePath = "build\bin\Release\MeetingSystemClient.exe"
if (Test-Path $exePath) {
    Write-Host ""
    Write-Host "=== Build Successful! ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "Executable: $exePath" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "To run the application:" -ForegroundColor Cyan
    Write-Host "  cd build\bin\Release" -ForegroundColor Yellow
    Write-Host "  .\MeetingSystemClient.exe" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Or run:" -ForegroundColor Cyan
    Write-Host "  .\run.ps1" -ForegroundColor Yellow
    Write-Host ""
    
    # Ask if user wants to run the application
    $response = Read-Host "Do you want to run the application now? (Y/n)"
    if ($response -ne "n" -and $response -ne "N") {
        Write-Host ""
        Write-Host "Starting application..." -ForegroundColor Cyan
        Start-Process $exePath
    }
} else {
    Write-Host ""
    Write-Host "Build verification failed!" -ForegroundColor Red
    Write-Host "Executable not found: $exePath" -ForegroundColor Yellow
    exit 1
}

