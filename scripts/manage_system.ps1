#!/usr/bin/env pwsh
# -*- coding: utf-8 -*-
# 音视频通话系统 - 一键管理脚本
# 功能：启动后端、前端、释放端口、测试服务

param(
    [Parameter(Position=0)]
    [ValidateSet("start", "stop", "restart", "test", "status", "clean", "help")]
    [string]$Action = "help"
)

# 颜色定义
$Colors = @{
    Success = "Green"
    Error = "Red"
    Warning = "Yellow"
    Info = "Cyan"
    Header = "Magenta"
}

# 服务配置
$Config = @{
    BackendPort = 8000
    FrontendPort = 8080
    AIServicePort = 5001
    DatabasePort = 5432
    RedisPort = 6379
    ProjectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
    BackendPath = "core/backend"
    FrontendPath = "web_interface"
    AIServicePath = "core/ai-service"
}

# 输出带颜色的消息
function Write-ColorMessage {
    param(
        [string]$Message,
        [string]$Color = "White",
        [switch]$NoNewline
    )
    
    if ($NoNewline) {
        Write-Host $Message -ForegroundColor $Color -NoNewline
    } else {
        Write-Host $Message -ForegroundColor $Color
    }
}

# 显示标题
function Show-Header {
    Write-ColorMessage "=" * 60 $Colors.Header
    Write-ColorMessage "🎥 音视频通话系统 - 一键管理脚本" $Colors.Header
    Write-ColorMessage "=" * 60 $Colors.Header
    Write-ColorMessage "📁 项目根目录: $($Config.ProjectRoot)" $Colors.Info
    Write-ColorMessage "🔧 支持的操作: start, stop, restart, test, status, clean, help" $Colors.Info
    Write-ColorMessage "=" * 60 $Colors.Header
}

# 检查端口是否被占用
function Test-Port {
    param([int]$Port)
    
    try {
        $connection = Test-NetConnection -ComputerName "localhost" -Port $Port -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
        return $connection.TcpTestSucceeded
    } catch {
        return $false
    }
}

# 释放端口
function Release-Port {
    param([int]$Port)
    
    Write-ColorMessage "🔓 正在释放端口 $Port..." $Colors.Warning
    
    try {
        # 查找占用端口的进程
        $processes = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue | 
                    Where-Object { $_.State -eq "Listen" } | 
                    Select-Object -ExpandProperty OwningProcess -Unique
        
        foreach ($pid in $processes) {
            try {
                $process = Get-Process -Id $pid -ErrorAction SilentlyContinue
                if ($process) {
                    Write-ColorMessage "   🛑 终止进程: $($process.ProcessName) (PID: $pid)" $Colors.Warning
                    Stop-Process -Id $pid -Force -ErrorAction SilentlyContinue
                }
            } catch {
                Write-ColorMessage "   ⚠️  无法终止进程 PID: $pid" $Colors.Warning
            }
        }
        
        Start-Sleep -Seconds 2
        
        if (Test-Port $Port) {
            Write-ColorMessage "   ❌ 端口 $Port 仍被占用" $Colors.Error
            return $false
        } else {
            Write-ColorMessage "   ✅ 端口 $Port 已释放" $Colors.Success
            return $true
        }
    } catch {
        Write-ColorMessage "   ❌ 释放端口 $Port 失败: $($_.Exception.Message)" $Colors.Error
        return $false
    }
}

# 释放所有相关端口
function Release-AllPorts {
    Write-ColorMessage "🔓 释放所有相关端口..." $Colors.Header
    
    $ports = @($Config.BackendPort, $Config.FrontendPort, $Config.AIServicePort, $Config.DatabasePort, $Config.RedisPort)
    $success = $true
    
    foreach ($port in $ports) {
        if (Test-Port $port) {
            if (-not (Release-Port $port)) {
                $success = $false
            }
        } else {
            Write-ColorMessage "   ✅ 端口 $port 未被占用" $Colors.Success
        }
    }
    
    return $success
}

# 检查服务状态
function Get-ServiceStatus {
    Write-ColorMessage "📊 检查服务状态..." $Colors.Header
    
    $services = @(
        @{ Name = "后端服务"; Port = $Config.BackendPort; URL = "http://localhost:$($Config.BackendPort)/health" },
        @{ Name = "前端服务"; Port = $Config.FrontendPort; URL = "http://localhost:$($Config.FrontendPort)" },
        @{ Name = "AI服务"; Port = $Config.AIServicePort; URL = "http://localhost:$($Config.AIServicePort)/health" },
        @{ Name = "数据库"; Port = $Config.DatabasePort; URL = $null },
        @{ Name = "Redis"; Port = $Config.RedisPort; URL = $null }
    )
    
    foreach ($service in $services) {
        $portStatus = if (Test-Port $service.Port) { "运行中" } else { "未运行" }
        Write-ColorMessage "   $($service.Name): $portStatus" $Colors.Info
        
        if ($service.URL) {
            try {
                $response = Invoke-WebRequest -Uri $service.URL -TimeoutSec 5 -ErrorAction SilentlyContinue
                if ($response.StatusCode -eq 200) {
                    Write-ColorMessage "      🌐 API响应: ✅ 正常" $Colors.Success
                } else {
                    Write-ColorMessage "      🌐 API响应: ⚠️  状态码 $($response.StatusCode)" $Colors.Warning
                }
            } catch {
                Write-ColorMessage "      🌐 API响应: ❌ 无法连接" $Colors.Error
            }
        }
    }
}

# 启动后端服务
function Start-BackendService {
    Write-ColorMessage "🚀 启动后端服务..." $Colors.Header
    
    $backendPath = Join-Path $Config.ProjectRoot $Config.BackendPath
    
    if (-not (Test-Path $backendPath)) {
        Write-ColorMessage "   ❌ 后端目录不存在: $backendPath" $Colors.Error
        return $false
    }
    
    # 检查是否有增强版后端
    $enhancedBackend = Join-Path $backendPath "enhanced-backend.go"
    if (Test-Path $enhancedBackend) {
        Write-ColorMessage "   📁 使用增强版后端: enhanced-backend.go" $Colors.Info
        
        # 切换到后端目录并启动服务
        Push-Location $backendPath
        try {
            Start-Process -FilePath "go" -ArgumentList "run", "enhanced-backend.go" -WindowStyle Hidden
            Start-Sleep -Seconds 3
            
            if (Test-Port $Config.BackendPort) {
                Write-ColorMessage "   ✅ 后端服务启动成功 (端口: $($Config.BackendPort))" $Colors.Success
                return $true
            } else {
                Write-ColorMessage "   ❌ 后端服务启动失败" $Colors.Error
                return $false
            }
        } finally {
            Pop-Location
        }
    } else {
        Write-ColorMessage "   ❌ 找不到后端服务文件" $Colors.Error
        return $false
    }
}

# 启动前端服务
function Start-FrontendService {
    Write-ColorMessage "🌐 启动前端服务..." $Colors.Header
    
    $frontendPath = Join-Path $Config.ProjectRoot $Config.FrontendPath
    
    if (-not (Test-Path $frontendPath)) {
        Write-ColorMessage "   ❌ 前端目录不存在: $frontendPath" $Colors.Error
        return $false
    }
    
    $serverFile = Join-Path $frontendPath "server.py"
    if (-not (Test-Path $serverFile)) {
        Write-ColorMessage "   ❌ 前端服务器文件不存在: $serverFile" $Colors.Error
        return $false
    }
    
    # 切换到前端目录并启动服务
    Push-Location $frontendPath
    try {
        Start-Process -FilePath "python" -ArgumentList "server.py" -WindowStyle Hidden
        Start-Sleep -Seconds 3
        
        if (Test-Port $Config.FrontendPort) {
            Write-ColorMessage "   ✅ 前端服务启动成功 (端口: $($Config.FrontendPort))" $Colors.Success
            return $true
        } else {
            Write-ColorMessage "   ❌ 前端服务启动失败" $Colors.Error
            return $false
        }
    } finally {
        Pop-Location
    }
}

# 启动AI服务
function Start-AIService {
    Write-ColorMessage "🤖 启动AI服务..." $Colors.Header
    
    $aiServicePath = Join-Path $Config.ProjectRoot $Config.AIServicePath
    
    if (-not (Test-Path $aiServicePath)) {
        Write-ColorMessage "   ❌ AI服务目录不存在: $aiServicePath" $Colors.Error
        return $false
    }
    
    $mainFile = Join-Path $aiServicePath "main.py"
    if (-not (Test-Path $mainFile)) {
        Write-ColorMessage "   ❌ AI服务主文件不存在: $mainFile" $Colors.Error
        return $false
    }
    
    # 切换到AI服务目录并启动服务
    Push-Location $aiServicePath
    try {
        Start-Process -FilePath "python" -ArgumentList "main.py" -WindowStyle Hidden
        Start-Sleep -Seconds 3
        
        if (Test-Port $Config.AIServicePort) {
            Write-ColorMessage "   ✅ AI服务启动成功 (端口: $($Config.AIServicePort))" $Colors.Success
            return $true
        } else {
            Write-ColorMessage "   ❌ AI服务启动失败" $Colors.Error
            return $false
        }
    } finally {
        Pop-Location
    }
}

# 停止所有服务
function Stop-AllServices {
    Write-ColorMessage "🛑 停止所有服务..." $Colors.Header
    
    $ports = @($Config.BackendPort, $Config.FrontendPort, $Config.AIServicePort, $Config.DatabasePort, $Config.RedisPort)
    
    foreach ($port in $ports) {
        if (Test-Port $port) {
            Release-Port $port | Out-Null
        }
    }
    
    # 终止相关进程
    $processes = @("go", "python", "node")
    foreach ($processName in $processes) {
        $runningProcesses = Get-Process -Name $processName -ErrorAction SilentlyContinue
        foreach ($process in $runningProcesses) {
            try {
                Write-ColorMessage "   🛑 终止进程: $($process.ProcessName) (PID: $($process.Id))" $Colors.Warning
                Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
            } catch {
                Write-ColorMessage "   ⚠️  无法终止进程: $($process.ProcessName)" $Colors.Warning
            }
        }
    }
    
    Write-ColorMessage "   ✅ 所有服务已停止" $Colors.Success
}

# 测试所有服务
function Test-AllServices {
    Write-ColorMessage "🧪 测试所有服务..." $Colors.Header
    
    # 等待服务启动
    Write-ColorMessage "   ⏳ 等待服务启动..." $Colors.Info
    Start-Sleep -Seconds 5
    
    # 测试后端服务
    Write-ColorMessage "   🔍 测试后端服务..." $Colors.Info
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:$($Config.BackendPort)/health" -Method GET -TimeoutSec 10
        Write-ColorMessage "      ✅ 后端服务正常: $($response.message)" $Colors.Success
    } catch {
        Write-ColorMessage "      ❌ 后端服务异常: $($_.Exception.Message)" $Colors.Error
    }
    
    # 测试前端服务
    Write-ColorMessage "   🔍 测试前端服务..." $Colors.Info
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:$($Config.FrontendPort)" -Method GET -TimeoutSec 10
        if ($response.StatusCode -eq 200) {
            Write-ColorMessage "      ✅ 前端服务正常" $Colors.Success
        } else {
            Write-ColorMessage "      ⚠️  前端服务状态码: $($response.StatusCode)" $Colors.Warning
        }
    } catch {
        Write-ColorMessage "      ❌ 前端服务异常: $($_.Exception.Message)" $Colors.Error
    }
    
    # 测试API功能
    Write-ColorMessage "   🔍 测试API功能..." $Colors.Info
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:$($Config.BackendPort)/api/v1/status" -Method GET -TimeoutSec 10
        Write-ColorMessage "      ✅ API状态正常: $($response.status)" $Colors.Success
    } catch {
        Write-ColorMessage "      ❌ API状态异常: $($_.Exception.Message)" $Colors.Error
    }
    
    Write-ColorMessage "   🎉 测试完成！" $Colors.Success
}

# 显示帮助信息
function Show-Help {
    Write-ColorMessage "📖 使用说明:" $Colors.Header
    Write-ColorMessage ""
    Write-ColorMessage "   .\manage_system.ps1 start     - 启动所有服务" $Colors.Info
    Write-ColorMessage "   .\manage_system.ps1 stop      - 停止所有服务" $Colors.Info
    Write-ColorMessage "   .\manage_system.ps1 restart   - 重启所有服务" $Colors.Info
    Write-ColorMessage "   .\manage_system.ps1 test      - 测试所有服务" $Colors.Info
    Write-ColorMessage "   .\manage_system.ps1 status    - 查看服务状态" $Colors.Info
    Write-ColorMessage "   .\manage_system.ps1 clean     - 清理端口占用" $Colors.Info
    Write-ColorMessage "   .\manage_system.ps1 help      - 显示此帮助" $Colors.Info
    Write-ColorMessage ""
    Write-ColorMessage "📋 服务端口:" $Colors.Header
    Write-ColorMessage "   - 后端服务: $($Config.BackendPort)" $Colors.Info
    Write-ColorMessage "   - 前端服务: $($Config.FrontendPort)" $Colors.Info
    Write-ColorMessage "   - AI服务: $($Config.AIServicePort)" $Colors.Info
    Write-ColorMessage "   - 数据库: $($Config.DatabasePort)" $Colors.Info
    Write-ColorMessage "   - Redis: $($Config.RedisPort)" $Colors.Info
    Write-ColorMessage ""
    Write-ColorMessage "🌐 访问地址:" $Colors.Header
    Write-ColorMessage "   - 前端界面: http://localhost:$($Config.FrontendPort)" $Colors.Info
    Write-ColorMessage "   - 后端API: http://localhost:$($Config.BackendPort)" $Colors.Info
    Write-ColorMessage "   - 健康检查: http://localhost:$($Config.BackendPort)/health" $Colors.Info
}

# 主函数
function Main {
    Show-Header
    
    switch ($Action.ToLower()) {
        "start" {
            Write-ColorMessage "🚀 启动所有服务..." $Colors.Header
            
            # 释放端口
            if (-not (Release-AllPorts)) {
                Write-ColorMessage "⚠️  部分端口释放失败，继续启动服务..." $Colors.Warning
            }
            
            # 启动服务
            $backendSuccess = Start-BackendService
            $frontendSuccess = Start-FrontendService
            $aiSuccess = Start-AIService
            
            if ($backendSuccess -and $frontendSuccess) {
                Write-ColorMessage "✅ 核心服务启动成功！" $Colors.Success
                Write-ColorMessage "🌐 访问前端界面: http://localhost:$($Config.FrontendPort)" $Colors.Info
                Write-ColorMessage "🔗 后端API地址: http://localhost:$($Config.BackendPort)" $Colors.Info
            } else {
                Write-ColorMessage "❌ 部分服务启动失败" $Colors.Error
            }
        }
        
        "stop" {
            Stop-AllServices
        }
        
        "restart" {
            Write-ColorMessage "🔄 重启所有服务..." $Colors.Header
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
            Write-ColorMessage "❌ 未知操作: $Action" $Colors.Error
            Show-Help
        }
    }
}

# 执行主函数
Main 