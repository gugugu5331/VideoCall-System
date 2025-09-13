#!/bin/bash

# 智能视频会议系统 - GitHub上传脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo
echo "========================================"
echo "   智能视频会议系统 - GitHub上传脚本"
echo "========================================"
echo

echo -e "${BLUE}📋 当前项目状态:${NC}"
git log --oneline -1
echo

echo -e "${BLUE}🔍 检查Git状态...${NC}"
if ! git status --porcelain > /dev/null 2>&1; then
    echo -e "${RED}❌ Git状态检查失败${NC}"
    exit 1
fi

echo
echo -e "${YELLOW}📤 准备上传到GitHub...${NC}"
echo
echo "请按照以下步骤操作:"
echo
echo "1️⃣ 首先在GitHub上创建新仓库:"
echo "   - 访问 https://github.com"
echo "   - 点击右上角 '+' → 'New repository'"
echo "   - 仓库名建议: VideoCall-System"
echo "   - 描述: 智能视频会议系统 - 带AI伪造音视频检测功能"
echo "   - 选择 Public 或 Private"
echo "   - 不要勾选任何初始化选项"
echo "   - 点击 'Create repository'"
echo

read -p "2️⃣ 请输入您的GitHub仓库URL (例如: https://github.com/username/VideoCall-System.git): " repo_url

if [ -z "$repo_url" ]; then
    echo -e "${RED}❌ 仓库URL不能为空${NC}"
    exit 1
fi

echo
echo -e "${BLUE}🔗 添加远程仓库...${NC}"
if ! git remote add origin "$repo_url" 2>/dev/null; then
    echo -e "${YELLOW}⚠️ 远程仓库可能已存在，尝试更新...${NC}"
    git remote set-url origin "$repo_url"
fi

echo
echo -e "${BLUE}📡 验证远程仓库连接...${NC}"
git remote -v

echo
echo -e "${BLUE}🚀 推送代码到GitHub...${NC}"
git branch -M main

if git push -u origin main; then
    echo
    echo -e "${GREEN}✅ 成功上传到GitHub!${NC}"
    echo
    echo -e "${GREEN}🎉 您的项目现在可以在以下地址访问:${NC}"
    echo "${repo_url%.git}"
    echo
    echo -e "${BLUE}📋 建议的后续步骤:${NC}"
    echo "- 在GitHub上设置仓库描述和标签"
    echo "- 创建第一个Release版本"
    echo "- 设置Issues和Projects"
    echo "- 邀请协作者参与开发"
    echo
else
    echo
    echo -e "${RED}❌ 上传失败，可能的原因:${NC}"
    echo "- 网络连接问题"
    echo "- GitHub认证失败 (需要设置Personal Access Token)"
    echo "- 仓库URL错误"
    echo "- 权限不足"
    echo
    echo -e "${YELLOW}💡 解决方案:${NC}"
    echo "1. 检查网络连接"
    echo "2. 设置GitHub Personal Access Token"
    echo "3. 确认仓库URL正确"
    echo "4. 查看详细错误信息"
    echo
fi

echo
read -p "按Enter键继续..."
