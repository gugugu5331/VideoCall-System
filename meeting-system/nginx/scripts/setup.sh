#!/bin/bash

# ============================================================================
# Nginx API网关设置脚本
# ============================================================================

set -e

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

# 检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要root权限运行"
        exit 1
    fi
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    
    directories=(
        "/var/www/static"
        "/var/www/admin"
        "/var/www/hls"
        "/var/www/cdn"
        "/var/cache/nginx"
        "/var/log/nginx"
        "/etc/nginx/ssl"
        "./logs"
        "./ssl"
        "./www/static"
        "./www/admin"
        "./www/hls"
        "./www/cdn"
        "./certbot"
    )
    
    for dir in "${directories[@]}"; do
        mkdir -p "$dir"
        log_info "创建目录: $dir"
    done
    
    # 设置权限
    chown -R nginx:nginx /var/www/ 2>/dev/null || true
    chown -R nginx:nginx /var/cache/nginx/ 2>/dev/null || true
    chmod -R 755 /var/www/
    
    log_success "目录创建完成"
}

# 生成自签名SSL证书（用于测试）
generate_ssl_cert() {
    log_info "生成自签名SSL证书..."
    
    if [[ ! -f "./ssl/cert.pem" ]]; then
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout ./ssl/key.pem \
            -out ./ssl/cert.pem \
            -subj "/C=CN/ST=Beijing/L=Beijing/O=Meeting/OU=IT/CN=meeting.example.com"
        
        # 创建证书链文件
        cp ./ssl/cert.pem ./ssl/chain.pem
        
        log_success "SSL证书生成完成"
    else
        log_info "SSL证书已存在，跳过生成"
    fi
}

# 创建默认的静态文件
create_default_files() {
    log_info "创建默认静态文件..."
    
    # 创建默认的管理后台页面
    cat > ./www/admin/index.html << 'EOF'
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>会议系统管理后台</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; }
        .status { padding: 10px; margin: 10px 0; border-radius: 4px; }
        .success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
    </style>
</head>
<body>
    <div class="container">
        <h1>会议系统管理后台</h1>
        <div class="status success">
            <strong>系统状态：</strong> 正常运行
        </div>
        <p>欢迎使用智能视频会议平台管理后台。</p>
    </div>
</body>
</html>
EOF

    # 创建默认的静态资源
    echo "/* Meeting System Styles */" > ./www/static/style.css
    
    log_success "默认文件创建完成"
}

# 验证配置文件
validate_config() {
    log_info "验证Nginx配置文件..."
    
    if command -v nginx >/dev/null 2>&1; then
        if nginx -t -c "$(pwd)/nginx.conf"; then
            log_success "Nginx配置文件验证通过"
        else
            log_error "Nginx配置文件验证失败"
            exit 1
        fi
    else
        log_warning "未找到nginx命令，跳过配置验证"
    fi
}

# 设置防火墙规则
setup_firewall() {
    log_info "设置防火墙规则..."
    
    if command -v ufw >/dev/null 2>&1; then
        ufw allow 80/tcp
        ufw allow 443/tcp
        # ufw allow 1935/tcp  # RTMP端口（如果需要）
        log_success "防火墙规则设置完成"
    elif command -v firewall-cmd >/dev/null 2>&1; then
        firewall-cmd --permanent --add-port=80/tcp
        firewall-cmd --permanent --add-port=443/tcp
        # firewall-cmd --permanent --add-port=1935/tcp  # RTMP端口（如果需要）
        firewall-cmd --reload
        log_success "防火墙规则设置完成"
    else
        log_warning "未找到防火墙管理工具，请手动开放80和443端口"
    fi
}

# 创建systemd服务文件
create_systemd_service() {
    log_info "创建systemd服务文件..."
    
    cat > /etc/systemd/system/meeting-nginx.service << EOF
[Unit]
Description=Meeting System Nginx Gateway
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$(pwd)
ExecStart=/usr/bin/docker-compose up -d nginx-gateway
ExecStop=/usr/bin/docker-compose stop nginx-gateway
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    log_success "systemd服务文件创建完成"
}

# 创建监控脚本
create_monitoring_script() {
    log_info "创建监控脚本..."
    
    cat > ./scripts/monitor.sh << 'EOF'
#!/bin/bash

# Nginx网关监控脚本

CONTAINER_NAME="meeting-nginx-gateway"
LOG_FILE="./logs/monitor.log"

# 检查容器状态
check_container() {
    if docker ps | grep -q "$CONTAINER_NAME"; then
        echo "$(date): Container $CONTAINER_NAME is running" >> "$LOG_FILE"
        return 0
    else
        echo "$(date): Container $CONTAINER_NAME is not running" >> "$LOG_FILE"
        return 1
    fi
}

# 检查健康状态
check_health() {
    if curl -f -s http://localhost/health > /dev/null; then
        echo "$(date): Health check passed" >> "$LOG_FILE"
        return 0
    else
        echo "$(date): Health check failed" >> "$LOG_FILE"
        return 1
    fi
}

# 重启容器
restart_container() {
    echo "$(date): Restarting container $CONTAINER_NAME" >> "$LOG_FILE"
    docker-compose restart nginx-gateway
}

# 主监控逻辑
main() {
    if ! check_container || ! check_health; then
        restart_container
    fi
}

main
EOF

    chmod +x ./scripts/monitor.sh
    
    # 创建cron任务
    (crontab -l 2>/dev/null; echo "*/5 * * * * $(pwd)/scripts/monitor.sh") | crontab -
    
    log_success "监控脚本创建完成"
}

# 主函数
main() {
    log_info "开始设置Nginx API网关..."
    
    check_root
    create_directories
    generate_ssl_cert
    create_default_files
    validate_config
    setup_firewall
    create_systemd_service
    create_monitoring_script
    
    log_success "Nginx API网关设置完成！"
    log_info "使用以下命令启动服务："
    log_info "  docker-compose up -d"
    log_info "或者："
    log_info "  systemctl start meeting-nginx"
}

# 脚本参数处理
case "${1:-}" in
    "ssl")
        generate_ssl_cert
        ;;
    "dirs")
        create_directories
        ;;
    "validate")
        validate_config
        ;;
    "monitor")
        create_monitoring_script
        ;;
    *)
        main
        ;;
esac
