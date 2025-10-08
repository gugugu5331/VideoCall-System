# Simple OpenCV Installation Script using vcpkg

Write-Host "=== OpenCV Installation (Simple Method) ===" -ForegroundColor Cyan
Write-Host ""

$VCPKG_ROOT = "C:\vcpkg"

# Check if vcpkg exists
if (-not (Test-Path $VCPKG_ROOT)) {
    Write-Host "vcpkg not found at $VCPKG_ROOT" -ForegroundColor Red
    Write-Host ""
    Write-Host "Installing vcpkg..." -ForegroundColor Cyan
    
    # Clone vcpkg
    git clone https://github.com/Microsoft/vcpkg.git $VCPKG_ROOT
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to clone vcpkg!" -ForegroundColor Red
        exit 1
    }
    
    # Bootstrap
    Set-Location $VCPKG_ROOT
    .\bootstrap-vcpkg.bat
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Failed to bootstrap vcpkg!" -ForegroundColor Red
        exit 1
    }
    
    # Integrate
    .\vcpkg integrate install
    
    Set-Location $PSScriptRoot
    
    Write-Host "vcpkg installed successfully!" -ForegroundColor Green
    Write-Host ""
}

# Check if OpenCV is already installed
Write-Host "Checking for existing OpenCV installation..." -ForegroundColor Cyan
Set-Location $VCPKG_ROOT

$installed = .\vcpkg list | Select-String "opencv4:x64-windows"

if ($installed) {
    Write-Host "OpenCV is already installed!" -ForegroundColor Green
    Write-Host $installed -ForegroundColor Gray
    Set-Location $PSScriptRoot
    
    Write-Host ""
    Write-Host "=== Installation Complete ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "1. Run: .\configure.ps1" -ForegroundColor Yellow
    Write-Host "2. Run: .\build.ps1" -ForegroundColor Yellow
    Write-Host ""
    exit 0
}

# Install OpenCV
Write-Host ""
Write-Host "Installing OpenCV via vcpkg..." -ForegroundColor Cyan
Write-Host "This will take 30-60 minutes. Please be patient..." -ForegroundColor Yellow
Write-Host ""

.\vcpkg install opencv4:x64-windows

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "OpenCV installation failed!" -ForegroundColor Red
    Set-Location $PSScriptRoot
    exit 1
}

Set-Location $PSScriptRoot

Write-Host ""
Write-Host "=== Installation Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "OpenCV installed successfully!" -ForegroundColor Green
Write-Host "Location: $VCPKG_ROOT\installed\x64-windows" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Run: .\configure.ps1" -ForegroundColor Yellow
Write-Host "2. Run: .\build.ps1" -ForegroundColor Yellow
Write-Host ""

