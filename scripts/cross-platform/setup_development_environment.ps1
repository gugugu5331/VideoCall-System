# Windows前端开发环境设置脚本
# VideoCall System - Windows Frontend Development Setup

param(
    [switch]$InstallDependencies,
    [switch]$SetupQt,
    [switch]$SetupOpenCV,
    [switch]$SetupVSCode,
    [switch]$ConfigureGit,
    [switch]$All
)

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput "✅ $Message" "Green" }
function Write-Info { param([string]$Message) Write-ColorOutput "ℹ️  $Message" "Cyan" }
function Write-Warning { param([string]$Message) Write-ColorOutput "⚠️  $Message" "Yellow" }
function Write-Error { param([string]$Message) Write-ColorOutput "❌ $Message" "Red" }

# 检查管理员权限
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# 检查并安装Chocolatey
function Install-Chocolatey {
    Write-Info "检查Chocolatey包管理器..."
    
    if (!(Get-Command choco -ErrorAction SilentlyContinue)) {
        Write-Info "安装Chocolatey..."
        Set-ExecutionPolicy Bypass -Scope Process -Force
        [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
        iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Chocolatey安装成功"
        } else {
            Write-Error "Chocolatey安装失败"
            return $false
        }
    } else {
        Write-Success "Chocolatey已安装"
    }
    return $true
}

# 安装基础开发工具
function Install-BasicTools {
    Write-Info "安装基础开发工具..."
    
    $tools = @(
        "git",
        "cmake",
        "ninja",
        "python3",
        "nodejs",
        "7zip",
        "curl",
        "wget"
    )
    
    foreach ($tool in $tools) {
        Write-Info "安装 $tool..."
        choco install $tool -y
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "$tool 安装成功"
        } else {
            Write-Warning "$tool 安装可能失败，请手动检查"
        }
    }
}

# 安装Visual Studio Build Tools
function Install-VSBuildTools {
    Write-Info "安装Visual Studio Build Tools..."
    
    # 检查是否已安装
    $vsWhere = "${env:ProgramFiles(x86)}\Microsoft Visual Studio\Installer\vswhere.exe"
    if (Test-Path $vsWhere) {
        $installations = & $vsWhere -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64
        if ($installations) {
            Write-Success "Visual Studio Build Tools已安装"
            return $true
        }
    }
    
    # 下载并安装
    $installerUrl = "https://aka.ms/vs/17/release/vs_buildtools.exe"
    $installerPath = "$env:TEMP\vs_buildtools.exe"
    
    Write-Info "下载Visual Studio Build Tools..."
    Invoke-WebRequest -Uri $installerUrl -OutFile $installerPath
    
    Write-Info "安装Visual Studio Build Tools（这可能需要几分钟）..."
    Start-Process -FilePath $installerPath -ArgumentList @(
        "--quiet",
        "--wait",
        "--add", "Microsoft.VisualStudio.Workload.VCTools",
        "--add", "Microsoft.VisualStudio.Component.VC.Tools.x86.x64",
        "--add", "Microsoft.VisualStudio.Component.Windows10SDK.19041"
    ) -Wait
    
    Remove-Item $installerPath -Force
    Write-Success "Visual Studio Build Tools安装完成"
}

# 设置Qt6开发环境
function Setup-Qt6 {
    Write-Info "设置Qt6开发环境..."
    
    # 检查Qt6是否已安装
    $qtPath = Get-ChildItem -Path "C:\Qt" -Directory -ErrorAction SilentlyContinue | Where-Object { $_.Name -match "^6\." } | Select-Object -First 1
    
    if ($qtPath) {
        Write-Success "Qt6已安装在: $($qtPath.FullName)"
        $qtBinPath = Join-Path $qtPath.FullName "msvc2019_64\bin"
        
        # 添加到PATH
        $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        if ($currentPath -notlike "*$qtBinPath*") {
            [Environment]::SetEnvironmentVariable("PATH", "$currentPath;$qtBinPath", "User")
            Write-Success "Qt6已添加到PATH"
        }
        
        # 设置Qt6_DIR环境变量
        $qt6Dir = Join-Path $qtPath.FullName "msvc2019_64"
        [Environment]::SetEnvironmentVariable("Qt6_DIR", $qt6Dir, "User")
        Write-Success "Qt6_DIR环境变量已设置: $qt6Dir"
        
    } else {
        Write-Warning "未找到Qt6安装，请手动安装Qt6"
        Write-Info "下载地址: https://www.qt.io/download-qt-installer"
        Write-Info "推荐安装Qt 6.5+ with MSVC 2019 64-bit"
    }
}

# 设置OpenCV
function Setup-OpenCV {
    Write-Info "设置OpenCV开发环境..."
    
    # 使用vcpkg安装OpenCV
    $vcpkgPath = "C:\vcpkg"
    
    if (!(Test-Path $vcpkgPath)) {
        Write-Info "安装vcpkg包管理器..."
        git clone https://github.com/Microsoft/vcpkg.git $vcpkgPath
        Set-Location $vcpkgPath
        .\bootstrap-vcpkg.bat
        .\vcpkg integrate install
    }
    
    Write-Info "安装OpenCV..."
    Set-Location $vcpkgPath
    .\vcpkg install opencv4[contrib,nonfree]:x64-windows
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "OpenCV安装成功"
        
        # 设置环境变量
        $opencvPath = "$vcpkgPath\installed\x64-windows"
        [Environment]::SetEnvironmentVariable("OpenCV_DIR", $opencvPath, "User")
        [Environment]::SetEnvironmentVariable("CMAKE_TOOLCHAIN_FILE", "$vcpkgPath\scripts\buildsystems\vcpkg.cmake", "User")
        
        Write-Success "OpenCV环境变量已设置"
    } else {
        Write-Error "OpenCV安装失败"
    }
}

# 配置VSCode
function Setup-VSCode {
    Write-Info "配置VSCode开发环境..."
    
    # 安装VSCode
    if (!(Get-Command code -ErrorAction SilentlyContinue)) {
        Write-Info "安装VSCode..."
        choco install vscode -y
    }
    
    # 安装必要的扩展
    $extensions = @(
        "ms-vscode.cpptools",
        "ms-vscode.cmake-tools",
        "twxs.cmake",
        "ms-python.python",
        "ms-vscode.vscode-json",
        "redhat.vscode-yaml",
        "ms-vscode.powershell"
    )
    
    foreach ($ext in $extensions) {
        Write-Info "安装VSCode扩展: $ext"
        code --install-extension $ext
    }
    
    # 创建VSCode配置
    $vscodeDir = ".vscode"
    if (!(Test-Path $vscodeDir)) {
        New-Item -ItemType Directory -Path $vscodeDir
    }
    
    # settings.json
    $settingsJson = @{
        "cmake.configureOnOpen" = $true
        "cmake.buildDirectory" = "`${workspaceFolder}/build"
        "cmake.generator" = "Ninja"
        "C_Cpp.default.configurationProvider" = "ms-vscode.cmake-tools"
        "files.associations" = @{
            "*.h" = "cpp"
            "*.hpp" = "cpp"
            "*.cpp" = "cpp"
            "*.qml" = "qml"
        }
    } | ConvertTo-Json -Depth 3
    
    $settingsJson | Out-File -FilePath "$vscodeDir\settings.json" -Encoding UTF8
    
    Write-Success "VSCode配置完成"
}

# 配置Git
function Configure-Git {
    Write-Info "配置Git开发环境..."
    
    # 设置Git配置
    Write-Info "请输入Git用户信息："
    $gitName = Read-Host "Git用户名"
    $gitEmail = Read-Host "Git邮箱"
    
    git config --global user.name $gitName
    git config --global user.email $gitEmail
    git config --global init.defaultBranch main
    git config --global core.autocrlf true
    git config --global core.editor "code --wait"
    
    Write-Success "Git配置完成"
}

# 创建项目构建脚本
function Create-BuildScripts {
    Write-Info "创建项目构建脚本..."
    
    # Windows构建脚本
    $buildScript = @'
@echo off
setlocal enabledelayedexpansion

echo 🚀 VideoCall System - Windows Build Script

:: 设置构建类型
set BUILD_TYPE=%1
if "%BUILD_TYPE%"=="" set BUILD_TYPE=Release

:: 设置构建目录
set BUILD_DIR=build-windows
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
mkdir %BUILD_DIR%
cd %BUILD_DIR%

:: 配置CMake
echo 📋 配置CMake...
cmake -G "Visual Studio 17 2022" -A x64 ^
    -DCMAKE_BUILD_TYPE=%BUILD_TYPE% ^
    -DCMAKE_TOOLCHAIN_FILE=C:/vcpkg/scripts/buildsystems/vcpkg.cmake ^
    -DQt6_DIR=%Qt6_DIR% ^
    -DOpenCV_DIR=%OpenCV_DIR% ^
    ..

if errorlevel 1 (
    echo ❌ CMake配置失败
    exit /b 1
)

:: 构建项目
echo 🔨 构建项目...
cmake --build . --config %BUILD_TYPE% --parallel

if errorlevel 1 (
    echo ❌ 构建失败
    exit /b 1
)

echo ✅ 构建完成！
echo 📁 输出目录: %CD%\%BUILD_TYPE%
'@
    
    $buildScript | Out-File -FilePath "build-windows.bat" -Encoding ASCII
    
    # PowerShell构建脚本
    $psBuildScript = @'
param(
    [string]$BuildType = "Release",
    [switch]$Clean,
    [switch]$Test
)

Write-Host "🚀 VideoCall System - Windows Build Script" -ForegroundColor Green

$BuildDir = "build-windows"

if ($Clean -and (Test-Path $BuildDir)) {
    Write-Host "🧹 清理构建目录..." -ForegroundColor Yellow
    Remove-Item $BuildDir -Recurse -Force
}

if (!(Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir
}

Set-Location $BuildDir

Write-Host "📋 配置CMake..." -ForegroundColor Cyan
cmake -G "Visual Studio 17 2022" -A x64 `
    -DCMAKE_BUILD_TYPE=$BuildType `
    -DCMAKE_TOOLCHAIN_FILE="C:/vcpkg/scripts/buildsystems/vcpkg.cmake" `
    -DQt6_DIR=$env:Qt6_DIR `
    -DOpenCV_DIR=$env:OpenCV_DIR `
    ..

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ CMake配置失败" -ForegroundColor Red
    exit 1
}

Write-Host "🔨 构建项目..." -ForegroundColor Cyan
cmake --build . --config $BuildType --parallel

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ 构建失败" -ForegroundColor Red
    exit 1
}

if ($Test) {
    Write-Host "🧪 运行测试..." -ForegroundColor Cyan
    ctest -C $BuildType --output-on-failure
}

Write-Host "✅ 构建完成！" -ForegroundColor Green
Write-Host "📁 输出目录: $(Get-Location)\$BuildType" -ForegroundColor Yellow
'@
    
    $psBuildScript | Out-File -FilePath "build-windows.ps1" -Encoding UTF8
    
    Write-Success "构建脚本创建完成"
}

# 主函数
function Main {
    Write-Host "🎯 VideoCall System - Windows前端开发环境设置" -ForegroundColor Green
    Write-Host "================================================" -ForegroundColor Green
    
    if (!(Test-Administrator)) {
        Write-Warning "建议以管理员身份运行此脚本以获得最佳体验"
    }
    
    if ($All -or $InstallDependencies) {
        if (Install-Chocolatey) {
            Install-BasicTools
            Install-VSBuildTools
        }
    }
    
    if ($All -or $SetupQt) {
        Setup-Qt6
    }
    
    if ($All -or $SetupOpenCV) {
        Setup-OpenCV
    }
    
    if ($All -or $SetupVSCode) {
        Setup-VSCode
    }
    
    if ($All -or $ConfigureGit) {
        Configure-Git
    }
    
    Create-BuildScripts
    
    Write-Host ""
    Write-Success "Windows前端开发环境设置完成！"
    Write-Info "请重启PowerShell以使环境变量生效"
    Write-Info "然后运行: .\build-windows.ps1 来构建项目"
}

# 显示帮助
if ($args.Count -eq 0 -and !$All) {
    Write-Host "用法: .\setup_development_environment.ps1 [选项]"
    Write-Host ""
    Write-Host "选项:"
    Write-Host "  -InstallDependencies  安装基础依赖"
    Write-Host "  -SetupQt             设置Qt6环境"
    Write-Host "  -SetupOpenCV         设置OpenCV环境"
    Write-Host "  -SetupVSCode         配置VSCode"
    Write-Host "  -ConfigureGit        配置Git"
    Write-Host "  -All                 执行所有设置"
    Write-Host ""
    Write-Host "示例:"
    Write-Host "  .\setup_development_environment.ps1 -All"
    Write-Host "  .\setup_development_environment.ps1 -SetupQt -SetupOpenCV"
    exit 0
}

Main
