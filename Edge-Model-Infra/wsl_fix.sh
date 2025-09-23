#!/bin/bash

# WSL 系统修复脚本
# 用于修复 WSL 环境中的 systemctl 和服务管理问题

echo "=========================================="
echo "WSL 系统修复脚本"
echo "=========================================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "此脚本需要 root 权限运行"
        echo "请使用: sudo $0"
        exit 1
    fi
}

# 备份重要文件
backup_files() {
    log_info "备份重要配置文件..."
    backup_dir="/tmp/wsl_backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    files_to_backup=("/etc/wsl.conf" "/etc/sudoers")
    for file in "${files_to_backup[@]}"; do
        if [ -f "$file" ]; then
            cp "$file" "$backup_dir/"
            log_success "已备份: $file"
        fi
    done
    
    log_info "备份目录: $backup_dir"
}

# 修复 1: 配置 WSL 以支持 systemd
fix_wsl_systemd() {
    echo -e "\n${BLUE}=== 修复 1: 配置 WSL systemd 支持 ===${NC}"
    
    log_info "配置 /etc/wsl.conf..."
    cat > /etc/wsl.conf << 'EOF'
[boot]
systemd=true

[user]
default=root

[network]
generateHosts=true
generateResolvConf=true

[interop]
enabled=true
appendWindowsPath=true
EOF
    
    log_success "已配置 /etc/wsl.conf"
    log_warning "需要重启 WSL 才能生效。请在 Windows PowerShell 中运行: wsl --shutdown"
}

# 修复 2: 修复 openssh-server 安装问题
fix_openssh_server() {
    echo -e "\n${BLUE}=== 修复 2: 修复 openssh-server 安装 ===${NC}"
    
    log_info "清理损坏的包状态..."
    dpkg --configure -a --force-depends
    
    log_info "修复 openssh-server 配置..."
    # 创建一个临时的 systemctl 替代脚本
    if [ ! -f /usr/bin/systemctl.bak ]; then
        cp /usr/bin/systemctl /usr/bin/systemctl.bak 2>/dev/null || true
    fi
    
    # 创建临时的 systemctl 脚本，跳过服务启动
    cat > /tmp/systemctl_temp << 'EOF'
#!/bin/bash
# 临时 systemctl 脚本，用于跳过服务启动
case "$1" in
    "enable"|"start"|"restart"|"reload")
        echo "Skipping systemctl $1 $2 (WSL compatibility)"
        exit 0
        ;;
    *)
        exec /usr/bin/systemctl.bak "$@"
        ;;
esac
EOF
    
    chmod +x /tmp/systemctl_temp
    mv /usr/bin/systemctl /usr/bin/systemctl.real
    cp /tmp/systemctl_temp /usr/bin/systemctl
    
    log_info "重新配置 openssh-server..."
    dpkg-reconfigure -fnoninteractive openssh-server
    
    # 恢复原始 systemctl
    mv /usr/bin/systemctl.real /usr/bin/systemctl
    rm -f /tmp/systemctl_temp
    
    log_success "openssh-server 配置完成"
}

# 修复 3: 修复 sudoers 权限
fix_sudoers() {
    echo -e "\n${BLUE}=== 修复 3: 修复 sudoers 权限 ===${NC}"
    
    log_info "检查 sudoers 文件权限..."
    if [ -f /etc/sudoers ]; then
        current_perm=$(stat -c "%a" /etc/sudoers)
        if [ "$current_perm" != "440" ]; then
            log_warning "sudoers 权限不正确: $current_perm，正在修复..."
            chmod 440 /etc/sudoers
            log_success "sudoers 权限已修复为 440"
        else
            log_success "sudoers 权限正确"
        fi
    fi
}

# 修复 4: 配置服务管理
fix_service_management() {
    echo -e "\n${BLUE}=== 修复 4: 配置服务管理 ===${NC}"
    
    log_info "创建服务管理脚本..."
    cat > /usr/local/bin/wsl-service << 'EOF'
#!/bin/bash
# WSL 服务管理脚本

case "$1" in
    "ssh"|"openssh-server")
        case "$2" in
            "start")
                /usr/sbin/sshd -D &
                echo "SSH 服务已启动"
                ;;
            "stop")
                pkill sshd
                echo "SSH 服务已停止"
                ;;
            "status")
                if pgrep sshd > /dev/null; then
                    echo "SSH 服务正在运行"
                else
                    echo "SSH 服务未运行"
                fi
                ;;
            *)
                echo "用法: $0 ssh {start|stop|status}"
                ;;
        esac
        ;;
    *)
        echo "支持的服务: ssh"
        echo "用法: $0 <service> {start|stop|status}"
        ;;
esac
EOF
    
    chmod +x /usr/local/bin/wsl-service
    log_success "服务管理脚本已创建: /usr/local/bin/wsl-service"
}

# 修复 5: 更新包管理器
fix_package_manager() {
    echo -e "\n${BLUE}=== 修复 5: 更新包管理器 ===${NC}"
    
    log_info "更新包列表..."
    apt update
    
    log_info "修复损坏的依赖..."
    apt --fix-broken install -y
    
    log_info "清理包缓存..."
    apt autoclean
    apt autoremove -y
    
    log_success "包管理器已更新和清理"
}

# 验证修复结果
verify_fixes() {
    echo -e "\n${BLUE}=== 验证修复结果 ===${NC}"
    
    log_info "检查 WSL 配置..."
    if [ -f /etc/wsl.conf ]; then
        log_success "/etc/wsl.conf 已配置"
    fi
    
    log_info "检查 openssh-server 状态..."
    if dpkg -l | grep -q "^ii.*openssh-server"; then
        log_success "openssh-server 已正确安装"
    else
        log_warning "openssh-server 可能仍有问题"
    fi
    
    log_info "检查服务管理脚本..."
    if [ -x /usr/local/bin/wsl-service ]; then
        log_success "服务管理脚本可用"
    fi
    
    log_info "检查包管理状态..."
    if ! dpkg --audit 2>/dev/null | grep -q .; then
        log_success "包管理状态正常"
    else
        log_warning "仍有包配置问题"
    fi
}

# 主函数
main() {
    check_root
    backup_files
    
    echo -e "\n${YELLOW}开始修复 WSL 系统问题...${NC}"
    
    fix_wsl_systemd
    fix_openssh_server
    fix_sudoers
    fix_service_management
    fix_package_manager
    
    verify_fixes
    
    echo -e "\n${GREEN}=== 修复完成 ===${NC}"
    echo -e "${YELLOW}重要提醒:${NC}"
    echo "1. 请在 Windows PowerShell 中运行 'wsl --shutdown' 重启 WSL"
    echo "2. 重启后，systemd 将可用"
    echo "3. 使用 'wsl-service ssh start' 启动 SSH 服务"
    echo "4. 使用 'systemctl status <service>' 检查服务状态"
}

# 运行主函数
main "$@"
