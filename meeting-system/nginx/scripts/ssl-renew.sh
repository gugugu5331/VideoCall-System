#!/bin/bash

# ============================================================================
# SSL证书自动续期脚本
# ============================================================================

set -e

# 配置变量
DOMAINS="meeting.example.com,api.meeting.com,admin.meeting.com"
EMAIL="admin@meeting.com"
WEBROOT_PATH="/var/www/certbot"
CERT_PATH="/etc/letsencrypt/live/meeting.example.com"
NGINX_CONTAINER="meeting-nginx-gateway"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# 检查证书是否需要续期
check_cert_expiry() {
    log_info "检查证书过期时间..."
    
    if [[ ! -f "$CERT_PATH/cert.pem" ]]; then
        log_warning "证书文件不存在，需要首次申请"
        return 1
    fi
    
    # 检查证书是否在30天内过期
    if openssl x509 -checkend 2592000 -noout -in "$CERT_PATH/cert.pem" >/dev/null 2>&1; then
        log_info "证书仍然有效，无需续期"
        return 0
    else
        log_warning "证书将在30天内过期，需要续期"
        return 1
    fi
}

# 申请或续期证书
renew_certificate() {
    log_info "开始申请/续期SSL证书..."
    
    # 确保webroot目录存在
    mkdir -p "$WEBROOT_PATH"
    
    # 使用certbot申请证书
    if docker run --rm \
        -v "$(pwd)/ssl:/etc/letsencrypt" \
        -v "$WEBROOT_PATH:/var/www/certbot" \
        certbot/certbot:latest \
        certonly \
        --webroot \
        --webroot-path=/var/www/certbot \
        --email "$EMAIL" \
        --agree-tos \
        --no-eff-email \
        --force-renewal \
        -d "${DOMAINS//,/ -d }"; then
        
        log_success "证书申请/续期成功"
        return 0
    else
        log_error "证书申请/续期失败"
        return 1
    fi
}

# 复制证书到nginx目录
copy_certificates() {
    log_info "复制证书文件..."
    
    if [[ -f "$CERT_PATH/fullchain.pem" && -f "$CERT_PATH/privkey.pem" ]]; then
        cp "$CERT_PATH/fullchain.pem" "$(pwd)/ssl/cert.pem"
        cp "$CERT_PATH/privkey.pem" "$(pwd)/ssl/key.pem"
        cp "$CERT_PATH/chain.pem" "$(pwd)/ssl/chain.pem" 2>/dev/null || \
        cp "$CERT_PATH/fullchain.pem" "$(pwd)/ssl/chain.pem"
        
        # 设置正确的权限
        chmod 644 "$(pwd)/ssl/cert.pem"
        chmod 600 "$(pwd)/ssl/key.pem"
        chmod 644 "$(pwd)/ssl/chain.pem"
        
        log_success "证书文件复制完成"
        return 0
    else
        log_error "证书文件不存在"
        return 1
    fi
}

# 重新加载nginx配置
reload_nginx() {
    log_info "重新加载Nginx配置..."
    
    if docker exec "$NGINX_CONTAINER" nginx -t; then
        if docker exec "$NGINX_CONTAINER" nginx -s reload; then
            log_success "Nginx配置重新加载成功"
            return 0
        else
            log_error "Nginx配置重新加载失败"
            return 1
        fi
    else
        log_error "Nginx配置验证失败"
        return 1
    fi
}

# 发送通知
send_notification() {
    local status=$1
    local message=$2
    
    # 这里可以集成邮件、Slack、钉钉等通知方式
    log_info "发送通知: $message"
    
    # 示例：发送邮件通知（需要配置邮件服务）
    # echo "$message" | mail -s "SSL Certificate $status" "$EMAIL"
    
    # 示例：发送到日志文件
    echo "$(date): SSL Certificate $status - $message" >> "$(pwd)/logs/ssl-renewal.log"
}

# 清理过期的证书备份
cleanup_old_backups() {
    log_info "清理过期的证书备份..."
    
    # 删除7天前的备份文件
    find "$(pwd)/ssl/backup" -name "*.pem" -mtime +7 -delete 2>/dev/null || true
    
    log_info "清理完成"
}

# 备份当前证书
backup_current_cert() {
    log_info "备份当前证书..."
    
    local backup_dir="$(pwd)/ssl/backup/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    if [[ -f "$(pwd)/ssl/cert.pem" ]]; then
        cp "$(pwd)/ssl/cert.pem" "$backup_dir/"
        cp "$(pwd)/ssl/key.pem" "$backup_dir/"
        cp "$(pwd)/ssl/chain.pem" "$backup_dir/" 2>/dev/null || true
        log_success "证书备份完成: $backup_dir"
    fi
}

# 验证证书
verify_certificate() {
    log_info "验证新证书..."
    
    local cert_file="$(pwd)/ssl/cert.pem"
    
    if [[ -f "$cert_file" ]]; then
        # 检查证书有效性
        if openssl x509 -in "$cert_file" -text -noout >/dev/null 2>&1; then
            # 检查证书域名
            local cert_domains=$(openssl x509 -in "$cert_file" -text -noout | grep -A1 "Subject Alternative Name" | tail -1 | sed 's/DNS://g' | sed 's/,//g')
            log_info "证书包含的域名: $cert_domains"
            
            # 检查证书过期时间
            local expiry_date=$(openssl x509 -in "$cert_file" -enddate -noout | cut -d= -f2)
            log_info "证书过期时间: $expiry_date"
            
            log_success "证书验证通过"
            return 0
        else
            log_error "证书格式无效"
            return 1
        fi
    else
        log_error "证书文件不存在"
        return 1
    fi
}

# 主函数
main() {
    log_info "开始SSL证书续期流程..."
    
    # 检查Docker是否运行
    if ! docker ps >/dev/null 2>&1; then
        log_error "Docker未运行或无权限访问"
        exit 1
    fi
    
    # 检查Nginx容器是否运行
    if ! docker ps | grep -q "$NGINX_CONTAINER"; then
        log_error "Nginx容器未运行"
        exit 1
    fi
    
    # 备份当前证书
    backup_current_cert
    
    # 检查是否需要续期
    if check_cert_expiry && [[ "${1:-}" != "--force" ]]; then
        log_info "证书仍然有效，无需续期"
        cleanup_old_backups
        exit 0
    fi
    
    # 续期证书
    if renew_certificate; then
        if copy_certificates && verify_certificate; then
            if reload_nginx; then
                send_notification "SUCCESS" "SSL证书续期成功"
                log_success "SSL证书续期流程完成"
            else
                send_notification "WARNING" "SSL证书续期成功，但Nginx重载失败"
                log_warning "证书续期成功，但Nginx重载失败"
                exit 1
            fi
        else
            send_notification "ERROR" "SSL证书续期失败：证书文件处理错误"
            log_error "证书文件处理失败"
            exit 1
        fi
    else
        send_notification "ERROR" "SSL证书续期失败：证书申请失败"
        log_error "证书申请失败"
        exit 1
    fi
    
    # 清理旧备份
    cleanup_old_backups
}

# 脚本参数处理
case "${1:-}" in
    "--check")
        check_cert_expiry
        ;;
    "--force")
        main --force
        ;;
    "--verify")
        verify_certificate
        ;;
    "--backup")
        backup_current_cert
        ;;
    "--cleanup")
        cleanup_old_backups
        ;;
    *)
        main
        ;;
esac
