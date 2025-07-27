@echo off
chcp 65001 >nul
title VideoCall System - GitHub Setup

echo ==========================================
echo VideoCall System - GitHub 仓库设置
echo ==========================================
echo.

echo [1/4] 检查Git配置...
git config --global user.name >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Git用户名未配置
    echo.
    set /p git_username="请输入您的Git用户名: "
    git config --global user.name "%git_username%"
    echo ✅ Git用户名已设置
) else (
    echo ✅ Git用户名已配置
    for /f "tokens=*" %%i in ('git config --global user.name') do echo   用户名: %%i
)

git config --global user.email >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Git邮箱未配置
    echo.
    set /p git_email="请输入您的Git邮箱: "
    git config --global user.email "%git_email%"
    echo ✅ Git邮箱已设置
) else (
    echo ✅ Git邮箱已配置
    for /f "tokens=*" %%i in ('git config --global user.email') do echo   邮箱: %%i
)

echo.
echo [2/4] 检查远程仓库...
git remote -v >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ 远程仓库已配置
    git remote -v
) else (
    echo ❌ 远程仓库未配置
    echo.
    echo 📋 请按以下步骤操作:
    echo.
    echo 1. 访问 https://github.com/new
    echo 2. 创建新仓库，建议名称: videocall-system
    echo 3. 不要初始化README、.gitignore或license
    echo 4. 复制仓库URL (例如: https://github.com/yourusername/videocall-system.git)
    echo.
    set /p repo_url="请输入GitHub仓库URL: "
    if not "%repo_url%"=="" (
        git remote add origin "%repo_url%"
        echo ✅ 远程仓库已添加
    )
)

echo.
echo [3/4] 检查GitHub CLI...
gh --version >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ GitHub CLI已安装
    echo.
    echo 是否要使用GitHub CLI创建仓库? (y/n)
    set /p use_gh="选择: "
    if /i "%use_gh%"=="y" (
        set /p repo_name="请输入仓库名称 (默认: videocall-system): "
        if "%repo_name%"=="" set repo_name=videocall-system
        echo 创建GitHub仓库: %repo_name%
        gh repo create %repo_name% --public --source=. --remote=origin --push
        if %errorlevel% equ 0 (
            echo ✅ GitHub仓库创建成功并已推送代码
            goto success
        ) else (
            echo ❌ GitHub CLI创建仓库失败
        )
    )
) else (
    echo ⚠️  GitHub CLI未安装 (可选)
    echo 下载地址: https://cli.github.com/
)

echo.
echo [4/4] 手动推送代码...
echo.
echo 📋 如果已创建GitHub仓库，请运行以下命令:
echo.
echo git remote add origin YOUR_REPOSITORY_URL
echo git branch -M main
echo git push -u origin main
echo.
echo 或者使用GitHub CLI:
echo gh repo create videocall-system --public --source=. --remote=origin --push
echo.

:success
echo.
echo ==========================================
echo GitHub设置完成！
echo ==========================================
echo.
echo 🎉 您的项目已准备好上传到GitHub
echo.
echo 📋 下一步操作:
echo 1. 如果使用GitHub CLI，运行: gh repo create videocall-system --public --source=. --remote=origin --push
echo 2. 如果手动操作，先创建仓库，然后运行推送命令
echo 3. 访问您的GitHub仓库查看代码
echo.
echo 💡 提示: 建议在GitHub仓库描述中添加项目介绍
echo.
pause 