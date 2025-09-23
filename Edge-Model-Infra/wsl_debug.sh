#!/bin/bash

# WSL 系统调试脚本
# 用于诊断和修复 WSL 环境中的 systemctl 问题

echo "=========================================="
echo "WSL 系统调试脚本"
echo "=========================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 1. 环境检测
echo -e "\n${BLUE}=== 1. 环境检测 ===${NC}"

log_info "检查操作系统信息..."
cat /etc/os-release | head -5

log_info "检查内核版本..."
uname -a

log_info "检查 WSL 版本..."
if [ -f /proc/version ]; then
    grep -i microsoft /proc/version && log_success "检测到 WSL 环境" || log_warning "可能不在 WSL 环境中"
fi

# 检查 WSL 版本
if command -v wsl.exe >/dev/null 2>&1; then
    log_info "WSL 版本信息:"
    wsl.exe --version 2>/dev/null || wsl.exe -l -v 2>/dev/null || log_warning "无法获取 WSL 版本信息"
fi

# 2. systemd 状态检查
echo -e "\n${BLUE}=== 2. systemd 状态检查 ===${NC}"

log_info "检查 systemd 是否运行..."
if systemctl --version >/dev/null 2>&1; then
    log_success "systemctl 可用"
    systemctl --version | head -2
    
    log_info "检查 systemd 状态..."
    if systemctl is-system-running >/dev/null 2>&1; then
        log_success "systemd 正在运行: $(systemctl is-system-running)"
    else
        log_warning "systemd 状态异常: $(systemctl is-system-running 2>/dev/null || echo 'unknown')"
    fi
else
    log_error "systemctl 不可用"
fi

# 检查 init 系统
log_info "检查当前 init 系统..."
if [ -d /run/systemd/system ]; then
    log_success "使用 systemd"
elif [ -f /sbin/init ] && [ "$(readlink /sbin/init)" = "systemd" ]; then
    log_success "配置为使用 systemd"
else
    log_warning "可能使用传统 init 系统"
    ls -la /sbin/init 2>/dev/null || log_error "/sbin/init 不存在"
fi

# 3. WSL 配置检查
echo -e "\n${BLUE}=== 3. WSL 配置检查 ===${NC}"

log_info "检查 WSL 配置文件..."
if [ -f /etc/wsl.conf ]; then
    log_success "找到 /etc/wsl.conf"
    cat /etc/wsl.conf
else
    log_warning "/etc/wsl.conf 不存在"
fi

# 4. 服务状态检查
echo -e "\n${BLUE}=== 4. 服务状态检查 ===${NC}"

log_info "检查关键服务状态..."
services=("ssh" "openssh-server" "dbus")
for service in "${services[@]}"; do
    if systemctl is-active "$service" >/dev/null 2>&1; then
        log_success "$service: $(systemctl is-active $service)"
    elif service "$service" status >/dev/null 2>&1; then
        log_success "$service: 通过 service 命令检测到运行"
    else
        log_warning "$service: 未运行或未安装"
    fi
done

# 5. 包管理状态检查
echo -e "\n${BLUE}=== 5. 包管理状态检查 ===${NC}"

log_info "检查 dpkg 状态..."
if dpkg --audit 2>/dev/null | grep -q .; then
    log_error "发现包配置问题:"
    dpkg --audit
else
    log_success "包管理状态正常"
fi

log_info "检查未完成的包配置..."
dpkg -l | grep -E "^(iU|iF)" && log_warning "发现未完成配置的包" || log_success "所有包配置完整"

# 6. 权限和文件系统检查
echo -e "\n${BLUE}=== 6. 权限和文件系统检查 ===${NC}"

log_info "检查关键目录权限..."
dirs=("/etc/sudoers" "/usr/bin/systemctl" "/usr/bin/deb-systemd-invoke")
for dir in "${dirs[@]}"; do
    if [ -e "$dir" ]; then
        ls -la "$dir"
    else
        log_warning "$dir 不存在"
    fi
done

# 7. 网络和连接检查
echo -e "\n${BLUE}=== 7. 网络检查 ===${NC}"

log_info "检查网络连接..."
if ping -c 1 8.8.8.8 >/dev/null 2>&1; then
    log_success "网络连接正常"
else
    log_warning "网络连接可能有问题"
fi

log_info "检查 DNS 解析..."
if nslookup google.com >/dev/null 2>&1; then
    log_success "DNS 解析正常"
else
    log_warning "DNS 解析可能有问题"
fi

echo -e "\n${GREEN}=== 诊断完成 ===${NC}"
echo "请查看上述输出，识别需要修复的问题。"
echo "接下来将提供相应的修复方案。"
