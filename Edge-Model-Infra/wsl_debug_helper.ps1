# WSL 调试助手脚本 (PowerShell)
# 用于在 Windows 中自动化 WSL 调试过程

Write-Host "=========================================="
Write-Host "WSL 调试助手"
Write-Host "=========================================="

# 颜色函数
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Info($message) {
    Write-ColorOutput Blue "[INFO] $message"
}

function Write-Success($message) {
    Write-ColorOutput Green "[SUCCESS] $message"
}

function Write-Warning($message) {
    Write-ColorOutput Yellow "[WARNING] $message"
}

function Write-Error($message) {
    Write-ColorOutput Red "[ERROR] $message"
}

# 1. 检查 WSL 状态
Write-Info "检查 WSL 安装状态..."
try {
    $wslVersion = wsl --version 2>$null
    if ($wslVersion) {
        Write-Success "WSL 已安装"
        Write-Output $wslVersion
    }
} catch {
    Write-Error "WSL 未安装或版本过旧"
    Write-Info "请安装 WSL2: https://docs.microsoft.com/en-us/windows/wsl/install"
    exit 1
}

# 2. 检查 WSL 发行版
Write-Info "检查已安装的 WSL 发行版..."
try {
    $distributions = wsl --list --verbose
    Write-Output $distributions
    
    # 检查是否有 Ubuntu
    if ($distributions -match "Ubuntu") {
        Write-Success "找到 Ubuntu 发行版"
    } else {
        Write-Warning "未找到 Ubuntu 发行版"
        Write-Info "请安装 Ubuntu: wsl --install -d Ubuntu"
    }
} catch {
    Write-Error "无法获取 WSL 发行版信息"
}

# 3. 复制调试脚本到 WSL 可访问位置
Write-Info "准备调试脚本..."
$currentDir = Get-Location
$wslPath = "/mnt/c" + $currentDir.Path.Replace("C:", "").Replace("\", "/")

Write-Info "当前目录的 WSL 路径: $wslPath"

# 4. 创建 WSL 命令脚本
$wslCommands = @"
#!/bin/bash
echo "=========================================="
echo "在 WSL 中执行调试脚本"
echo "=========================================="

# 导航到项目目录
cd "$wslPath"
echo "当前目录: `$(pwd)"

# 检查调试脚本是否存在
if [ -f "wsl_debug.sh" ]; then
    echo "找到调试脚本，开始执行..."
    chmod +x wsl_debug.sh
    ./wsl_debug.sh
    
    echo ""
    echo "=========================================="
    echo "是否要运行修复脚本? (y/n)"
    read -p "请输入选择: " choice
    
    if [ "`$choice" = "y" ] || [ "`$choice" = "Y" ]; then
        if [ -f "wsl_fix.sh" ]; then
            echo "开始运行修复脚本..."
            chmod +x wsl_fix.sh
            sudo ./wsl_fix.sh
        else
            echo "错误: 找不到修复脚本 wsl_fix.sh"
        fi
    fi
else
    echo "错误: 找不到调试脚本 wsl_debug.sh"
    echo "请确保在正确的目录中运行此脚本"
fi

echo ""
echo "调试完成。请查看输出结果。"
echo "如需重启 WSL，请在 Windows PowerShell 中运行: wsl --shutdown"
"@

# 将命令写入临时文件
$tempScript = "$env:TEMP\wsl_debug_commands.sh"
$wslCommands | Out-File -FilePath $tempScript -Encoding UTF8

Write-Info "已创建临时脚本: $tempScript"

# 5. 提供执行选项
Write-Host ""
Write-ColorOutput Cyan "请选择执行方式:"
Write-Host "1. 自动执行 WSL 调试 (推荐)"
Write-Host "2. 手动进入 WSL 环境"
Write-Host "3. 仅显示命令，不执行"
Write-Host ""

$choice = Read-Host "请输入选择 (1-3)"

switch ($choice) {
    "1" {
        Write-Info "自动执行 WSL 调试..."
        try {
            # 复制脚本到 WSL 并执行
            wsl cp $tempScript /tmp/wsl_debug_commands.sh
            wsl chmod +x /tmp/wsl_debug_commands.sh
            wsl /tmp/wsl_debug_commands.sh
        } catch {
            Write-Error "自动执行失败，请尝试手动方式"
            Write-Info "手动命令: wsl"
        }
    }
    "2" {
        Write-Info "启动 WSL 环境..."
        Write-Warning "请在 WSL 中手动运行以下命令:"
        Write-Host "cd $wslPath"
        Write-Host "./wsl_debug.sh"
        Write-Host "sudo ./wsl_fix.sh"
        Write-Host ""
        wsl
    }
    "3" {
        Write-Info "WSL 调试命令:"
        Write-Host "wsl"
        Write-Host "cd $wslPath"
        Write-Host "chmod +x wsl_debug.sh wsl_fix.sh"
        Write-Host "./wsl_debug.sh"
        Write-Host "sudo ./wsl_fix.sh"
        Write-Host "exit"
        Write-Host "wsl --shutdown"
        Write-Host "wsl"
    }
    default {
        Write-Warning "无效选择，退出"
    }
}

# 清理临时文件
if (Test-Path $tempScript) {
    Remove-Item $tempScript -Force
}

Write-Host ""
Write-ColorOutput Green "WSL 调试助手完成"
Write-Info "如有问题，请查看 WSL_DEBUG_GUIDE.md 获取详细说明"
