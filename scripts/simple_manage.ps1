param(
    [Parameter(Position=0)]
    [ValidateSet("start", "stop", "restart", "test", "status", "clean", "help")]
    [string]$Action = "help"
)

# Configuration
$Config = @{
    BackendPort = 8000
    FrontendPort = 8080
    AIServicePort = 5001
    ProjectRoot = Split-Path -Parent $PSScriptRoot
    BackendPath = "core/backend"
    FrontendPath = "web_interface"
    AIServicePath = "core/ai-service"
}

# Test port
function Test-Port {
    param([int]$Port)
    try {
        $connection = Test-NetConnection -ComputerName "localhost" -Port $Port -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
        return $connection.TcpTestSucceeded
    } catch {
        return $false
    }
}

# Release all ports
function Release-AllPorts {
    Write-Host "Releasing all ports..." -ForegroundColor Magenta
    
    $ports = @($Config.BackendPort, $Config.FrontendPort, $Config.AIServicePort)
    $success = $true
    
    foreach ($port in $ports) {
        if (Test-Port $port) {
            if (-not (Release-Port $port)) {
                $success = $false
            }
        } else {
            Write-Host "  Port $port is not occupied" -ForegroundColor Green
        }
    }
    
    return $success
}

# Release port
function Release-Port {
    param([int]$Port)
    Write-Host "Releasing port $Port..." -ForegroundColor Yellow
    
    try {
        $processes = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue | 
                    Where-Object { $_.State -eq "Listen" } | 
                    Select-Object -ExpandProperty OwningProcess -Unique
        
        foreach ($pid in $processes) {
            try {
                $process = Get-Process -Id $pid -ErrorAction SilentlyContinue
                if ($process) {
                    Write-Host "  Stopping process: $($process.ProcessName) (PID: $pid)" -ForegroundColor Yellow
                    Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
                }
            } catch {
                Write-Host "  Cannot stop process PID: $pid" -ForegroundColor Yellow
            }
        }
        
        Start-Sleep -Seconds 2
        
        if (Test-Port $Port) {
            Write-Host "  FAILED: Port $Port still occupied" -ForegroundColor Red
            return $false
        } else {
            Write-Host "  SUCCESS: Port $Port released" -ForegroundColor Green
            return $true
        }
    } catch {
        Write-Host "  ERROR: Failed to release port $Port" -ForegroundColor Red
        return $false
    }
}

# Get service status
function Get-ServiceStatus {
    Write-Host "Checking service status..." -ForegroundColor Magenta
    
    $services = @(
        @{ Name = "Backend Service"; Port = $Config.BackendPort; URL = "http://localhost:$($Config.BackendPort)/health" },
        @{ Name = "Frontend Service"; Port = $Config.FrontendPort; URL = "http://localhost:$($Config.FrontendPort)" },
        @{ Name = "AI Service"; Port = $Config.AIServicePort; URL = "http://localhost:$($Config.AIServicePort)/health" }
    )
    
    foreach ($service in $services) {
        $portStatus = if (Test-Port $service.Port) { "RUNNING" } else { "STOPPED" }
        Write-Host "  $($service.Name): $portStatus" -ForegroundColor Cyan
        
        if ($service.URL) {
            try {
                $response = Invoke-WebRequest -Uri $service.URL -TimeoutSec 5 -ErrorAction SilentlyContinue
                if ($response.StatusCode -eq 200) {
                    Write-Host "    API Response: OK" -ForegroundColor Green
                } else {
                    Write-Host "    API Response: Status $($response.StatusCode)" -ForegroundColor Yellow
                }
            } catch {
                Write-Host "    API Response: FAILED" -ForegroundColor Red
            }
        }
    }
}

# Start backend service
function Start-BackendService {
    Write-Host "Starting backend service..." -ForegroundColor Magenta
    
    $backendPath = Join-Path $Config.ProjectRoot $Config.BackendPath
    $enhancedBackend = Join-Path $backendPath "enhanced-backend.go"
    
    if (-not (Test-Path $enhancedBackend)) {
        Write-Host "  ERROR: Backend file not found" -ForegroundColor Red
        return $false
    }
    
    Write-Host "  Using enhanced backend: enhanced-backend.go" -ForegroundColor Cyan
    
    Push-Location $backendPath
    try {
        Start-Process -FilePath "go" -ArgumentList "run", "enhanced-backend.go" -WindowStyle Hidden
        Start-Sleep -Seconds 3
        
        if (Test-Port $Config.BackendPort) {
            Write-Host "  SUCCESS: Backend service started (Port: $($Config.BackendPort))" -ForegroundColor Green
            return $true
        } else {
            Write-Host "  FAILED: Backend service failed to start" -ForegroundColor Red
            return $false
        }
    } finally {
        Pop-Location
    }
}

# Start frontend service
function Start-FrontendService {
    Write-Host "Starting frontend service..." -ForegroundColor Magenta
    
    $frontendPath = Join-Path $Config.ProjectRoot $Config.FrontendPath
    $serverFile = Join-Path $frontendPath "server.py"
    
    if (-not (Test-Path $serverFile)) {
        Write-Host "  ERROR: Frontend server file not found" -ForegroundColor Red
        return $false
    }
    
    Push-Location $frontendPath
    try {
        Start-Process -FilePath "python" -ArgumentList "server.py" -WindowStyle Hidden
        Start-Sleep -Seconds 3
        
        if (Test-Port $Config.FrontendPort) {
            Write-Host "  SUCCESS: Frontend service started (Port: $($Config.FrontendPort))" -ForegroundColor Green
            return $true
        } else {
            Write-Host "  FAILED: Frontend service failed to start" -ForegroundColor Red
            return $false
        }
    } finally {
        Pop-Location
    }
}

# Stop all services
function Stop-AllServices {
    Write-Host "Stopping all services..." -ForegroundColor Magenta
    
    $ports = @($Config.BackendPort, $Config.FrontendPort, $Config.AIServicePort)
    
    foreach ($port in $ports) {
        if (Test-Port $port) {
            Release-Port $port | Out-Null
        }
    }
    
    $processes = @("go", "python")
    foreach ($processName in $processes) {
        $runningProcesses = Get-Process -Name $processName -ErrorAction SilentlyContinue
        foreach ($process in $runningProcesses) {
            try {
                Write-Host "  Stopping process: $($process.ProcessName) (PID: $($process.Id))" -ForegroundColor Yellow
                Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
            } catch {
                Write-Host "  Cannot stop process: $($process.ProcessName)" -ForegroundColor Yellow
            }
        }
    }
    
    Write-Host "  All services stopped" -ForegroundColor Green
}

# Test all services
function Test-AllServices {
    Write-Host "Testing all services..." -ForegroundColor Magenta
    
    Start-Sleep -Seconds 5
    
    Write-Host "  Testing backend service..." -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:$($Config.BackendPort)/health" -Method GET -TimeoutSec 10
        Write-Host "    SUCCESS: Backend service OK - $($response.message)" -ForegroundColor Green
    } catch {
        Write-Host "    FAILED: Backend service error - $($_.Exception.Message)" -ForegroundColor Red
    }
    
    Write-Host "  Testing frontend service..." -ForegroundColor Cyan
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:$($Config.FrontendPort)" -Method GET -TimeoutSec 10
        if ($response.StatusCode -eq 200) {
            Write-Host "    SUCCESS: Frontend service OK" -ForegroundColor Green
        } else {
            Write-Host "    WARNING: Frontend service status $($response.StatusCode)" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "    FAILED: Frontend service error - $($_.Exception.Message)" -ForegroundColor Red
    }
    
    Write-Host "  Testing API functionality..." -ForegroundColor Cyan
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:$($Config.BackendPort)/api/v1/status" -Method GET -TimeoutSec 10
        Write-Host "    SUCCESS: API status OK - $($response.status)" -ForegroundColor Green
    } catch {
        Write-Host "    FAILED: API status error - $($_.Exception.Message)" -ForegroundColor Red
    }
    
    Write-Host "  Testing completed!" -ForegroundColor Green
}

# Show help
function Show-Help {
    Write-Host "Usage:" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "  .\simple_manage.ps1 start     - Start all services" -ForegroundColor Cyan
    Write-Host "  .\simple_manage.ps1 stop      - Stop all services" -ForegroundColor Cyan
    Write-Host "  .\simple_manage.ps1 restart   - Restart all services" -ForegroundColor Cyan
    Write-Host "  .\simple_manage.ps1 test      - Test all services" -ForegroundColor Cyan
    Write-Host "  .\simple_manage.ps1 status    - Check service status" -ForegroundColor Cyan
    Write-Host "  .\simple_manage.ps1 clean     - Clean port usage" -ForegroundColor Cyan
    Write-Host "  .\simple_manage.ps1 help      - Show this help" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Service Ports:" -ForegroundColor Magenta
    Write-Host "  - Backend Service: $($Config.BackendPort)" -ForegroundColor Cyan
    Write-Host "  - Frontend Service: $($Config.FrontendPort)" -ForegroundColor Cyan
    Write-Host "  - AI Service: $($Config.AIServicePort)" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Access URLs:" -ForegroundColor Magenta
    Write-Host "  - Frontend: http://localhost:$($Config.FrontendPort)" -ForegroundColor Cyan
    Write-Host "  - Backend API: http://localhost:$($Config.BackendPort)" -ForegroundColor Cyan
    Write-Host "  - Health Check: http://localhost:$($Config.BackendPort)/health" -ForegroundColor Cyan
}

# Main function
function Main {
    Write-Host "================================================================" -ForegroundColor Magenta
    Write-Host "VideoCall System - Management Script" -ForegroundColor Magenta
    Write-Host "================================================================" -ForegroundColor Magenta
    Write-Host "Project Root: $($Config.ProjectRoot)" -ForegroundColor Cyan
    Write-Host "================================================================" -ForegroundColor Magenta
    
    switch ($Action.ToLower()) {
        "start" {
            Write-Host "Starting all services..." -ForegroundColor Magenta
            
            if (-not (Release-AllPorts)) {
                Write-Host "Warning: Some ports failed to release, continuing..." -ForegroundColor Yellow
            }
            
            $backendSuccess = Start-BackendService
            $frontendSuccess = Start-FrontendService
            
            if ($backendSuccess -and $frontendSuccess) {
                Write-Host "SUCCESS: Core services started!" -ForegroundColor Green
                Write-Host "Frontend: http://localhost:$($Config.FrontendPort)" -ForegroundColor Cyan
                Write-Host "Backend API: http://localhost:$($Config.BackendPort)" -ForegroundColor Cyan
            } else {
                Write-Host "FAILED: Some services failed to start" -ForegroundColor Red
            }
        }
        
        "stop" {
            Stop-AllServices
        }
        
        "restart" {
            Write-Host "Restarting all services..." -ForegroundColor Magenta
            Stop-AllServices
            Start-Sleep -Seconds 2
            & $PSCommandPath "start"
        }
        
        "test" {
            Test-AllServices
        }
        
        "status" {
            Get-ServiceStatus
        }
        
        "clean" {
            Release-AllPorts
        }
        
        "help" {
            Show-Help
        }
        
        default {
            Write-Host "ERROR: Unknown action: $Action" -ForegroundColor Red
            Show-Help
        }
    }
}

# Execute main function
Main 