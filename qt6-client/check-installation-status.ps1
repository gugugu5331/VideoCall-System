# OpenCV Installation Status Checker

Write-Host "=== OpenCV Installation Status Checker ===" -ForegroundColor Cyan
Write-Host ""

$VCPKG_ROOT = "C:\vcpkg"

# Check if vcpkg exists
if (-not (Test-Path $VCPKG_ROOT)) {
    Write-Host "✗ vcpkg not found at $VCPKG_ROOT" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install vcpkg first:" -ForegroundColor Yellow
    Write-Host "  git clone https://github.com/Microsoft/vcpkg.git C:\vcpkg" -ForegroundColor Cyan
    Write-Host "  cd C:\vcpkg" -ForegroundColor Cyan
    Write-Host "  .\bootstrap-vcpkg.bat" -ForegroundColor Cyan
    Write-Host ""
    exit 1
}

Write-Host "✓ vcpkg found at $VCPKG_ROOT" -ForegroundColor Green
Write-Host ""

# Check OpenCV installation
Write-Host "Checking OpenCV installation..." -ForegroundColor Cyan
Set-Location $VCPKG_ROOT

$installed = .\vcpkg list 2>$null | Select-String "opencv4:x64-windows"

if ($installed) {
    Write-Host "✓ OpenCV is installed!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Details:" -ForegroundColor Cyan
    Write-Host $installed -ForegroundColor Gray
    Write-Host ""
    
    # Check installation directory
    $installDir = "$VCPKG_ROOT\installed\x64-windows"
    if (Test-Path "$installDir\include\opencv2") {
        Write-Host "✓ OpenCV headers found" -ForegroundColor Green
    }
    
    if (Test-Path "$installDir\lib\opencv_core4.lib") {
        Write-Host "✓ OpenCV libraries found" -ForegroundColor Green
    }
    
    if (Test-Path "$installDir\bin\opencv_world*.dll") {
        Write-Host "✓ OpenCV DLLs found" -ForegroundColor Green
    }
    
    Write-Host ""
    Write-Host "=== Installation Complete ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "1. cd d:\qt6-client" -ForegroundColor Yellow
    Write-Host "2. .\configure.ps1" -ForegroundColor Yellow
    Write-Host "3. .\build.ps1" -ForegroundColor Yellow
    Write-Host ""
    
} else {
    Write-Host "✗ OpenCV is not installed" -ForegroundColor Red
    Write-Host ""
    
    # Check if installation is in progress
    $buildDir = "$VCPKG_ROOT\buildtrees\opencv4"
    if (Test-Path $buildDir) {
        Write-Host "⏳ Installation may be in progress..." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Build directory found: $buildDir" -ForegroundColor Gray
        Write-Host ""
        
        # Check for log files
        $logFiles = Get-ChildItem -Path $buildDir -Filter "*.log" -ErrorAction SilentlyContinue | Sort-Object LastWriteTime -Descending
        
        if ($logFiles) {
            $latestLog = $logFiles[0]
            Write-Host "Latest log file: $($latestLog.Name)" -ForegroundColor Cyan
            Write-Host "Last modified: $($latestLog.LastWriteTime)" -ForegroundColor Gray
            Write-Host ""
            
            # Check if log was recently updated (within last 5 minutes)
            $timeDiff = (Get-Date) - $latestLog.LastWriteTime
            if ($timeDiff.TotalMinutes -lt 5) {
                Write-Host "✓ Installation is actively running" -ForegroundColor Green
                Write-Host "  (Log file updated $([math]::Round($timeDiff.TotalMinutes, 1)) minutes ago)" -ForegroundColor Gray
            } else {
                Write-Host "⚠ Installation may have stalled" -ForegroundColor Yellow
                Write-Host "  (Log file last updated $([math]::Round($timeDiff.TotalMinutes, 1)) minutes ago)" -ForegroundColor Gray
            }
            
            Write-Host ""
            Write-Host "To view the log:" -ForegroundColor Cyan
            Write-Host "  Get-Content '$($latestLog.FullName)' -Tail 50" -ForegroundColor Yellow
        }
        
        # Check for running processes
        Write-Host ""
        Write-Host "Checking for vcpkg/cmake processes..." -ForegroundColor Cyan
        $processes = Get-Process | Where-Object { $_.ProcessName -match "vcpkg|cmake|cl|link" } | Select-Object ProcessName, CPU, WorkingSet
        
        if ($processes) {
            Write-Host "✓ Found active build processes:" -ForegroundColor Green
            $processes | Format-Table -AutoSize
        } else {
            Write-Host "✗ No active build processes found" -ForegroundColor Red
        }
        
    } else {
        Write-Host "Installation has not been started yet." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "To install OpenCV, run:" -ForegroundColor Cyan
        Write-Host "  cd C:\vcpkg" -ForegroundColor Yellow
        Write-Host "  .\vcpkg install opencv4:x64-windows" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Or use the installation script:" -ForegroundColor Cyan
        Write-Host "  cd d:\qt6-client" -ForegroundColor Yellow
        Write-Host "  .\install-opencv.bat" -ForegroundColor Yellow
    }
    
    Write-Host ""
}

Set-Location $PSScriptRoot

# Summary
Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Cyan
Write-Host ""

$summary = @{
    "vcpkg" = if (Test-Path $VCPKG_ROOT) { "✓ Installed" } else { "✗ Not found" }
    "OpenCV" = if ($installed) { "✓ Installed" } else { "✗ Not installed" }
    "Build in progress" = if (Test-Path "$VCPKG_ROOT\buildtrees\opencv4") { "⏳ Possibly" } else { "✗ No" }
}

$summary.GetEnumerator() | ForEach-Object {
    $value = $_.Value
    $color = "Gray"
    if ($value -like "*Installed*") {
        $color = "Green"
    } elseif ($value -like "*Possibly*") {
        $color = "Yellow"
    } else {
        $color = "Red"
    }
    Write-Host "$($_.Key): " -NoNewline
    Write-Host $value -ForegroundColor $color
}

Write-Host ""

