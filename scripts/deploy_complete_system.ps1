# VideoCall System - 完整系统部署脚本
# Windows前端 + WSL后端完整部署方案

param(
    [switch]$SkipBackend,
    [switch]$SkipFrontend,
    [switch]$SkipChecks,
    [string]$WSLDistro = "Ubuntu"
)

# 颜色定义
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    Cyan = "Cyan"
    Magenta = "Magenta"
}

# 日志函数
function Write-Log {
    param([string]$Message, [string]$Level = "INFO", [string]$Color = "White")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $Color
}

function Write-Info { param([string]$Message) Write-Log $Message "INFO" $Colors.Green }
function Write-Warn { param([string]$Message) Write-Log $Message "WARN" $Colors.Yellow }
function Write-Error { param([string]$Message) Write-Log $Message "ERROR" $Colors.Red }
function Write-Step { param([string]$Message) Write-Log $Message "STEP" $Colors.Blue }
function Write-Success { param([string]$Message) Write-Log $Message "SUCCESS" $Colors.Green }

# 检查函数
function Test-WSL {
    Write-Step "检查WSL环境..."
    
    try {
        $wslList = wsl --list --running
        if ($LASTEXITCODE -ne 0) {
            Write-Error "WSL未安装或未运行"
            return $false
        }
        
        if ($wslList -notmatch $WSLDistro) {
            Write-Warn "WSL发行版 '$WSLDistro' 未运行，尝试启动..."
            wsl -d $WSLDistro echo "WSL启动测试"
            if ($LASTEXITCODE -ne 0) {
                Write-Error "无法启动WSL发行版 '$WSLDistro'"
                return $false
            }
        }
        
        Write-Success "WSL环境检查通过"
        return $true
    }
    catch {
        Write-Error "WSL检查失败: $($_.Exception.Message)"
        return $false
    }
}

function Test-Docker {
    Write-Step "检查Docker环境..."
    
    try {
        # 检查Docker Desktop是否运行
        $dockerProcess = Get-Process "Docker Desktop" -ErrorAction SilentlyContinue
        if (-not $dockerProcess) {
            Write-Warn "Docker Desktop未运行，尝试启动..."
            Start-Process "Docker Desktop" -WindowStyle Hidden
            Start-Sleep 30
        }
        
        # 在WSL中检查Docker
        $dockerCheck = wsl -d $WSLDistro docker info 2>&1
        if ($LASTEXITCODE -ne 0) {
            Write-Error "WSL中Docker未运行"
            return $false
        }
        
        Write-Success "Docker环境检查通过"
        return $true
    }
    catch {
        Write-Error "Docker检查失败: $($_.Exception.Message)"
        return $false
    }
}

function Test-Prerequisites {
    Write-Step "检查前置条件..."
    
    $checks = @()
    
    # 检查Git
    if (Get-Command git -ErrorAction SilentlyContinue) {
        Write-Info "✅ Git已安装"
        $checks += $true
    } else {
        Write-Error "❌ Git未安装"
        $checks += $false
    }
    
    # 检查CMake
    if (Get-Command cmake -ErrorAction SilentlyContinue) {
        Write-Info "✅ CMake已安装"
        $checks += $true
    } else {
        Write-Error "❌ CMake未安装"
        $checks += $false
    }
    
    # 检查Qt6
    $qtPath = Get-ChildItem -Path "C:\Qt" -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -like "6.*" }
    if ($qtPath) {
        Write-Info "✅ Qt6已安装: $($qtPath.FullName)"
        $checks += $true
    } else {
        Write-Error "❌ Qt6未安装"
        $checks += $false
    }
    
    # 检查Visual Studio Build Tools
    $vsBuildTools = Get-ChildItem -Path "C:\Program Files*\Microsoft Visual Studio\*\*\MSBuild\Current\Bin" -ErrorAction SilentlyContinue
    if ($vsBuildTools) {
        Write-Info "✅ Visual Studio Build Tools已安装"
        $checks += $true
    } else {
        Write-Error "❌ Visual Studio Build Tools未安装"
        $checks += $false
    }
    
    return ($checks -notcontains $false)
}

# 获取WSL IP地址
function Get-WSLIPAddress {
    Write-Step "获取WSL IP地址..."
    
    try {
        $wslIP = wsl -d $WSLDistro hostname -I | ForEach-Object { $_.Trim().Split(' ')[0] }
        if ($wslIP -and $wslIP -match '^\d+\.\d+\.\d+\.\d+$') {
            Write-Success "WSL IP地址: $wslIP"
            return $wslIP
        } else {
            Write-Warn "无法获取WSL IP，使用默认地址"
            return "172.20.0.1"
        }
    }
    catch {
        Write-Error "获取WSL IP失败: $($_.Exception.Message)"
        return "172.20.0.1"
    }
}

# 部署后端服务
function Deploy-Backend {
    Write-Step "部署WSL后端服务..."
    
    try {
        # 复制部署脚本到WSL
        $scriptPath = "scripts/deploy_wsl_backend.sh"
        wsl -d $WSLDistro chmod +x $scriptPath
        
        # 在WSL中执行部署脚本
        Write-Info "在WSL中执行后端部署..."
        wsl -d $WSLDistro bash $scriptPath
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "后端服务部署成功"
            return $true
        } else {
            Write-Error "后端服务部署失败"
            return $false
        }
    }
    catch {
        Write-Error "后端部署异常: $($_.Exception.Message)"
        return $false
    }
}

# 构建前端应用
function Build-Frontend {
    Write-Step "构建Windows前端应用..."
    
    try {
        $frontendPath = "src/frontend/qt-client-new"
        
        if (-not (Test-Path $frontendPath)) {
            Write-Error "前端目录不存在: $frontendPath"
            return $false
        }
        
        Push-Location $frontendPath
        
        # 创建构建目录
        if (Test-Path "build") {
            Remove-Item "build" -Recurse -Force
        }
        New-Item -ItemType Directory -Name "build" | Out-Null
        
        Push-Location "build"
        
        # CMake配置
        Write-Info "配置CMake..."
        cmake .. -G "Visual Studio 17 2022" -A x64
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "CMake配置失败"
            return $false
        }
        
        # 构建项目
        Write-Info "构建项目..."
        cmake --build . --config Release
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "项目构建失败"
            return $false
        }
        
        Pop-Location
        Pop-Location
        
        Write-Success "前端应用构建成功"
        return $true
    }
    catch {
        Write-Error "前端构建异常: $($_.Exception.Message)"
        return $false
    }
    finally {
        Pop-Location -ErrorAction SilentlyContinue
        Pop-Location -ErrorAction SilentlyContinue
    }
}

# 配置前端网络
function Configure-FrontendNetwork {
    param([string]$WSLIPAddress)
    
    Write-Step "配置前端网络连接..."
    
    try {
        $configPath = "src/frontend/qt-client-new/config/network_config.json"
        
        if (Test-Path $configPath) {
            $config = Get-Content $configPath | ConvertFrom-Json
            
            # 更新WSL IP地址
            $config.network.wsl_backend.base_url = "http://$WSLIPAddress:80"
            $config.network.wsl_backend.api_base_url = "http://$WSLIPAddress:80/api"
            $config.network.wsl_backend.websocket_url = "ws://$WSLIPAddress:80/ws"
            
            # 更新服务端点
            $config.network.services.ai_detection_service.unit_manager_endpoint = "http://$WSLIPAddress:10001"
            
            # 保存配置
            $config | ConvertTo-Json -Depth 10 | Set-Content $configPath
            
            Write-Success "前端网络配置已更新"
            return $true
        } else {
            Write-Error "网络配置文件不存在: $configPath"
            return $false
        }
    }
    catch {
        Write-Error "配置前端网络失败: $($_.Exception.Message)"
        return $false
    }
}

# 测试连接
function Test-Connection {
    param([string]$WSLIPAddress)
    
    Write-Step "测试系统连接..."
    
    $endpoints = @(
        @{ Name = "网关服务"; URL = "http://$WSLIPAddress:80/health" },
        @{ Name = "API网关"; URL = "http://$WSLIPAddress:80/api/health" },
        @{ Name = "Edge AI"; URL = "http://$WSLIPAddress:10001/health" }
    )
    
    foreach ($endpoint in $endpoints) {
        try {
            Write-Info "测试 $($endpoint.Name)..."
            $response = Invoke-WebRequest -Uri $endpoint.URL -TimeoutSec 10 -ErrorAction Stop
            if ($response.StatusCode -eq 200) {
                Write-Success "✅ $($endpoint.Name) 连接正常"
            } else {
                Write-Warn "⚠️ $($endpoint.Name) 响应异常: $($response.StatusCode)"
            }
        }
        catch {
            Write-Warn "⚠️ $($endpoint.Name) 连接失败: $($_.Exception.Message)"
        }
    }
}

# 显示部署结果
function Show-DeploymentResult {
    param([string]$WSLIPAddress)
    
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "  VideoCall System 部署完成！" -ForegroundColor Cyan
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "🌐 系统访问信息:" -ForegroundColor Green
    Write-Host "  WSL IP地址: $WSLIPAddress"
    Write-Host "  后端API: http://$WSLIPAddress:80/api"
    Write-Host "  WebSocket: ws://$WSLIPAddress:80/ws"
    Write-Host "  Edge AI: http://$WSLIPAddress:10001"
    Write-Host ""
    Write-Host "🖥️ 前端应用:" -ForegroundColor Green
    Write-Host "  构建路径: src/frontend/qt-client-new/build/Release"
    Write-Host "  可执行文件: VideoCallClient.exe"
    Write-Host ""
    Write-Host "🔧 管理命令:" -ForegroundColor Green
    Write-Host "  查看后端日志: wsl -d $WSLDistro docker-compose -f deployment/docker-compose.wsl.yml logs -f"
    Write-Host "  停止后端服务: wsl -d $WSLDistro docker-compose -f deployment/docker-compose.wsl.yml down"
    Write-Host "  重启后端服务: wsl -d $WSLDistro docker-compose -f deployment/docker-compose.wsl.yml restart"
    Write-Host ""
    Write-Host "💡 使用提示:" -ForegroundColor Yellow
    Write-Host "  1. 启动前端应用前，请确保后端服务正在运行"
    Write-Host "  2. 前端应用会自动检测WSL IP地址并连接后端"
    Write-Host "  3. 如遇连接问题，请检查Windows防火墙设置"
    Write-Host ""
}

# 主函数
function Main {
    Write-Host "========================================" -ForegroundColor Magenta
    Write-Host "  VideoCall System 完整部署" -ForegroundColor Magenta
    Write-Host "  Windows前端 + WSL后端" -ForegroundColor Magenta
    Write-Host "========================================" -ForegroundColor Magenta
    Write-Host ""
    
    # 检查前置条件
    if (-not $SkipChecks) {
        if (-not (Test-Prerequisites)) {
            Write-Error "前置条件检查失败，请安装缺失的组件"
            exit 1
        }
        
        if (-not (Test-WSL)) {
            Write-Error "WSL环境检查失败"
            exit 1
        }
        
        if (-not (Test-Docker)) {
            Write-Error "Docker环境检查失败"
            exit 1
        }
    }
    
    # 获取WSL IP地址
    $wslIP = Get-WSLIPAddress
    
    # 部署后端服务
    if (-not $SkipBackend) {
        if (-not (Deploy-Backend)) {
            Write-Error "后端部署失败"
            exit 1
        }
    }
    
    # 构建前端应用
    if (-not $SkipFrontend) {
        if (-not (Build-Frontend)) {
            Write-Error "前端构建失败"
            exit 1
        }
        
        # 配置前端网络
        if (-not (Configure-FrontendNetwork -WSLIPAddress $wslIP)) {
            Write-Error "前端网络配置失败"
            exit 1
        }
    }
    
    # 测试连接
    Test-Connection -WSLIPAddress $wslIP
    
    # 显示部署结果
    Show-DeploymentResult -WSLIPAddress $wslIP
    
    Write-Success "完整系统部署成功！"
}

# 错误处理
$ErrorActionPreference = "Stop"
trap {
    Write-Error "部署过程中发生错误: $($_.Exception.Message)"
    exit 1
}

# 执行主函数
Main
