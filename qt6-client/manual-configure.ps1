# Manual CMake Configuration Script

Write-Host "=== Manual CMake Configuration ===" -ForegroundColor Cyan
Write-Host ""

# Qt6 paths to try
$qt6Paths = @(
    "D:\qt\6.10.0\msvc2022_64",
    "D:\qt\6.9.0\msvc2022_64",
    "D:\qt\6.10.0\msvc2019_64",
    "D:\qt\6.9.0\msvc2019_64",
    "C:\Qt\6.10.0\msvc2022_64",
    "C:\Qt\6.9.0\msvc2022_64",
    "C:\Qt\6.5.3\msvc2019_64"
)

$qt6Path = $null
foreach ($path in $qt6Paths) {
    if (Test-Path $path) {
        $qt6Path = $path
        Write-Host "Found Qt6: $qt6Path" -ForegroundColor Green
        break
    }
}

if (-not $qt6Path) {
    Write-Host "ERROR: Qt6 not found!" -ForegroundColor Red
    Write-Host "Please install Qt6 from https://www.qt.io/download" -ForegroundColor Yellow
    exit 1
}

# vcpkg toolchain
$vcpkgToolchain = "C:\vcpkg\scripts\buildsystems\vcpkg.cmake"
if (-not (Test-Path $vcpkgToolchain)) {
    Write-Host "ERROR: vcpkg toolchain not found!" -ForegroundColor Red
    exit 1
}

Write-Host "Using Qt6: $qt6Path" -ForegroundColor Cyan
Write-Host "Using vcpkg: $vcpkgToolchain" -ForegroundColor Cyan
Write-Host ""

# Create build directory
if (-not (Test-Path "build")) {
    New-Item -ItemType Directory -Path "build" | Out-Null
    Write-Host "Created build directory" -ForegroundColor Green
}

# Run CMake
Write-Host "Running CMake..." -ForegroundColor Cyan
Set-Location build

$cmakeCmd = "cmake -G `"Visual Studio 17 2022`" -A x64 -DCMAKE_PREFIX_PATH=`"$qt6Path`" -DCMAKE_TOOLCHAIN_FILE=`"$vcpkgToolchain`" .."

Write-Host "Command: $cmakeCmd" -ForegroundColor Gray
Write-Host ""

Invoke-Expression $cmakeCmd

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "=== Configuration Successful ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next step: Run build script" -ForegroundColor Cyan
    Write-Host "  .\build.ps1" -ForegroundColor Yellow
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "=== Configuration Failed ===" -ForegroundColor Red
    Write-Host ""
    Write-Host "Exit code: $LASTEXITCODE" -ForegroundColor Red
}

Set-Location ..

