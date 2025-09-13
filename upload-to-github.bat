@echo off
chcp 65001 >nul
echo.
echo ========================================
echo   智能视频会议系统 - GitHub上传脚本
echo ========================================
echo.

echo 📋 当前项目状态:
git log --oneline -1
echo.

echo 🔍 检查Git状态...
git status --porcelain
if %errorlevel% neq 0 (
    echo ❌ Git状态检查失败
    pause
    exit /b 1
)

echo.
echo 📤 准备上传到GitHub...
echo.
echo 请按照以下步骤操作:
echo.
echo 1️⃣ 首先在GitHub上创建新仓库:
echo    - 访问 https://github.com
echo    - 点击右上角 "+" → "New repository"
echo    - 仓库名建议: VideoCall-System
echo    - 描述: 智能视频会议系统 - 带AI伪造音视频检测功能
echo    - 选择 Public 或 Private
echo    - 不要勾选任何初始化选项
echo    - 点击 "Create repository"
echo.

set /p repo_url="2️⃣ 请输入您的GitHub仓库URL (例如: https://github.com/username/VideoCall-System.git): "

if "%repo_url%"=="" (
    echo ❌ 仓库URL不能为空
    pause
    exit /b 1
)

echo.
echo 🔗 添加远程仓库...
git remote add origin %repo_url%
if %errorlevel% neq 0 (
    echo ⚠️ 远程仓库可能已存在，尝试更新...
    git remote set-url origin %repo_url%
)

echo.
echo 📡 验证远程仓库连接...
git remote -v

echo.
echo 🚀 推送代码到GitHub...
git branch -M main
git push -u origin main

if %errorlevel% equ 0 (
    echo.
    echo ✅ 成功上传到GitHub!
    echo.
    echo 🎉 您的项目现在可以在以下地址访问:
    echo %repo_url:~0,-4%
    echo.
    echo 📋 建议的后续步骤:
    echo - 在GitHub上设置仓库描述和标签
    echo - 创建第一个Release版本
    echo - 设置Issues和Projects
    echo - 邀请协作者参与开发
    echo.
) else (
    echo.
    echo ❌ 上传失败，可能的原因:
    echo - 网络连接问题
    echo - GitHub认证失败 (需要设置Personal Access Token)
    echo - 仓库URL错误
    echo - 权限不足
    echo.
    echo 💡 解决方案:
    echo 1. 检查网络连接
    echo 2. 设置GitHub Personal Access Token
    echo 3. 确认仓库URL正确
    echo 4. 查看详细错误信息
)

echo.
pause
