#!/bin/bash

# ============================================================================
# Nginx网关健康检查脚本
# ============================================================================

set -e

# 配置变量
NGINX_CONTAINER="meeting-nginx-gateway"
GATEWAY_URL="https://localhost"
HEALTH_ENDPOINT="/health"
DETAILED_HEALTH_ENDPOINT="/health/detailed"
LOG_FILE="./logs/health-check.log"
ALERT_THRESHOLD=3  # 连续失败次数阈值
FAILURE_COUNT_FILE="./logs/.failure_count"

# 后端服务列表
BACKEND_SERVICES=(
    "user-service:8080"
    "meeting-service:8082"
    "signaling-service:8081"
    "media-service:8083"
    "ai-service:8084"
    "notification-service:8085"
)

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [INFO] $1" >> "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [SUCCESS] $1" >> "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [WARNING] $1" >> "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [ERROR] $1" >> "$LOG_FILE"
}

# 创建日志目录
mkdir -p "$(dirname "$LOG_FILE")"

# 检查Docker容器状态
check_container_status() {
    log_info "检查Nginx容器状态..."
    
    if docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "$NGINX_CONTAINER.*Up"; then
        log_success "Nginx容器运行正常"
        return 0
    else
        log_error "Nginx容器未运行或状态异常"
        return 1
    fi
}

# 检查容器资源使用情况
check_container_resources() {
    log_info "检查容器资源使用情况..."
    
    local stats=$(docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep "$NGINX_CONTAINER")
    
    if [[ -n "$stats" ]]; then
        log_info "容器资源使用: $stats"
        
        # 提取CPU使用率
        local cpu_usage=$(echo "$stats" | awk '{print $2}' | sed 's/%//')
        if (( $(echo "$cpu_usage > 80" | bc -l) )); then
            log_warning "CPU使用率过高: ${cpu_usage}%"
        fi
        
        return 0
    else
        log_error "无法获取容器资源信息"
        return 1
    fi
}

# 检查网关健康状态
check_gateway_health() {
    log_info "检查网关健康状态..."
    
    local response=$(curl -s -k -w "%{http_code}" -o /tmp/health_response "$GATEWAY_URL$HEALTH_ENDPOINT" 2>/dev/null || echo "000")
    
    if [[ "$response" == "200" ]]; then
        local health_data=$(cat /tmp/health_response 2>/dev/null)
        log_success "网关健康检查通过: $health_data"
        return 0
    else
        log_error "网关健康检查失败，HTTP状态码: $response"
        return 1
    fi
}

# 检查后端服务连通性
check_backend_services() {
    log_info "检查后端服务连通性..."
    
    local failed_services=()
    
    for service in "${BACKEND_SERVICES[@]}"; do
        local service_name=$(echo "$service" | cut -d: -f1)
        local service_port=$(echo "$service" | cut -d: -f2)
        
        if timeout 5 bash -c "</dev/tcp/$service_name/$service_port" 2>/dev/null; then
            log_success "服务 $service 连通正常"
        else
            log_error "服务 $service 连通失败"
            failed_services+=("$service")
        fi
    done
    
    if [[ ${#failed_services[@]} -eq 0 ]]; then
        return 0
    else
        log_error "以下服务连通失败: ${failed_services[*]}"
        return 1
    fi
}

# 检查SSL证书状态
check_ssl_certificate() {
    log_info "检查SSL证书状态..."
    
    local cert_file="./ssl/cert.pem"
    
    if [[ -f "$cert_file" ]]; then
        # 检查证书是否在30天内过期
        if openssl x509 -checkend 2592000 -noout -in "$cert_file" >/dev/null 2>&1; then
            local expiry_date=$(openssl x509 -in "$cert_file" -enddate -noout | cut -d= -f2)
            log_success "SSL证书有效，过期时间: $expiry_date"
            return 0
        else
            log_warning "SSL证书将在30天内过期"
            return 1
        fi
    else
        log_error "SSL证书文件不存在"
        return 1
    fi
}

# 检查磁盘空间
check_disk_space() {
    log_info "检查磁盘空间..."
    
    local disk_usage=$(df -h . | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [[ $disk_usage -gt 90 ]]; then
        log_error "磁盘空间不足，使用率: ${disk_usage}%"
        return 1
    elif [[ $disk_usage -gt 80 ]]; then
        log_warning "磁盘空间紧张，使用率: ${disk_usage}%"
        return 0
    else
        log_success "磁盘空间充足，使用率: ${disk_usage}%"
        return 0
    fi
}

# 检查日志文件大小
check_log_files() {
    log_info "检查日志文件大小..."
    
    local log_dir="./logs"
    local max_size_mb=100
    
    if [[ -d "$log_dir" ]]; then
        while IFS= read -r -d '' file; do
            local size_mb=$(du -m "$file" | cut -f1)
            if [[ $size_mb -gt $max_size_mb ]]; then
                log_warning "日志文件过大: $file (${size_mb}MB)"
                # 可以在这里添加日志轮转逻辑
            fi
        done < <(find "$log_dir" -name "*.log" -print0)
    fi
    
    return 0
}

# 检查网络连接
check_network_connectivity() {
    log_info "检查网络连接..."
    
    # 检查DNS解析
    if nslookup google.com >/dev/null 2>&1; then
        log_success "DNS解析正常"
    else
        log_error "DNS解析失败"
        return 1
    fi
    
    # 检查外网连通性
    if curl -s --connect-timeout 5 http://www.google.com >/dev/null 2>&1; then
        log_success "外网连通正常"
    else
        log_warning "外网连通异常"
    fi
    
    return 0
}

# 性能测试
performance_test() {
    log_info "执行性能测试..."
    
    # 简单的响应时间测试
    local start_time=$(date +%s%N)
    curl -s -k "$GATEWAY_URL$HEALTH_ENDPOINT" >/dev/null 2>&1
    local end_time=$(date +%s%N)
    
    local response_time=$(( (end_time - start_time) / 1000000 ))  # 转换为毫秒
    
    if [[ $response_time -lt 100 ]]; then
        log_success "响应时间正常: ${response_time}ms"
    elif [[ $response_time -lt 500 ]]; then
        log_warning "响应时间较慢: ${response_time}ms"
    else
        log_error "响应时间过慢: ${response_time}ms"
        return 1
    fi
    
    return 0
}

# 获取失败计数
get_failure_count() {
    if [[ -f "$FAILURE_COUNT_FILE" ]]; then
        cat "$FAILURE_COUNT_FILE"
    else
        echo "0"
    fi
}

# 设置失败计数
set_failure_count() {
    echo "$1" > "$FAILURE_COUNT_FILE"
}

# 发送告警
send_alert() {
    local message="$1"
    local severity="$2"
    
    log_error "发送告警: $message"
    
    # 这里可以集成各种告警方式
    # 邮件告警
    # echo "$message" | mail -s "Nginx Gateway Alert [$severity]" admin@meeting.com
    
    # Webhook告警
    # curl -X POST -H "Content-Type: application/json" \
    #      -d "{\"text\":\"$message\",\"severity\":\"$severity\"}" \
    #      https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
    
    # 写入告警日志
    echo "$(date '+%Y-%m-%d %H:%M:%S') [ALERT] [$severity] $message" >> "./logs/alerts.log"
}

# 自动修复
auto_repair() {
    log_info "尝试自动修复..."
    
    # 重启Nginx容器
    if docker restart "$NGINX_CONTAINER" >/dev/null 2>&1; then
        log_success "Nginx容器重启成功"
        sleep 10  # 等待容器启动
        return 0
    else
        log_error "Nginx容器重启失败"
        return 1
    fi
}

# 生成健康报告
generate_health_report() {
    local report_file="./logs/health-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$report_file" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "gateway_status": "$(check_gateway_health && echo "healthy" || echo "unhealthy")",
    "container_status": "$(check_container_status && echo "running" || echo "stopped")",
    "ssl_status": "$(check_ssl_certificate && echo "valid" || echo "invalid")",
    "disk_usage": "$(df -h . | awk 'NR==2 {print $5}')",
    "backend_services": [
EOF

    local first=true
    for service in "${BACKEND_SERVICES[@]}"; do
        if [[ "$first" == "true" ]]; then
            first=false
        else
            echo "," >> "$report_file"
        fi
        
        local service_name=$(echo "$service" | cut -d: -f1)
        local service_port=$(echo "$service" | cut -d: -f2)
        local status="down"
        
        if timeout 5 bash -c "</dev/tcp/$service_name/$service_port" 2>/dev/null; then
            status="up"
        fi
        
        echo "        {\"name\": \"$service_name\", \"port\": $service_port, \"status\": \"$status\"}" >> "$report_file"
    done

    cat >> "$report_file" << EOF
    ]
}
EOF

    log_info "健康报告已生成: $report_file"
}

# 主健康检查函数
main_health_check() {
    log_info "开始健康检查..."
    
    local checks_passed=0
    local total_checks=0
    
    # 执行各项检查
    local checks=(
        "check_container_status"
        "check_gateway_health"
        "check_backend_services"
        "check_ssl_certificate"
        "check_disk_space"
        "check_network_connectivity"
        "performance_test"
    )
    
    for check in "${checks[@]}"; do
        total_checks=$((total_checks + 1))
        if $check; then
            checks_passed=$((checks_passed + 1))
        fi
    done
    
    # 计算健康分数
    local health_score=$((checks_passed * 100 / total_checks))
    log_info "健康检查完成，得分: $health_score% ($checks_passed/$total_checks)"
    
    # 判断整体健康状态
    if [[ $health_score -ge 80 ]]; then
        log_success "系统健康状态良好"
        set_failure_count 0
        return 0
    else
        log_error "系统健康状态异常"
        
        # 增加失败计数
        local failure_count=$(get_failure_count)
        failure_count=$((failure_count + 1))
        set_failure_count $failure_count
        
        # 检查是否需要告警和自动修复
        if [[ $failure_count -ge $ALERT_THRESHOLD ]]; then
            send_alert "Nginx网关连续$failure_count次健康检查失败，健康分数: $health_score%" "CRITICAL"
            
            # 尝试自动修复
            if auto_repair; then
                send_alert "自动修复成功，系统已恢复" "INFO"
                set_failure_count 0
            else
                send_alert "自动修复失败，需要人工干预" "CRITICAL"
            fi
        fi
        
        return 1
    fi
}

# 脚本参数处理
case "${1:-}" in
    "--container")
        check_container_status
        ;;
    "--gateway")
        check_gateway_health
        ;;
    "--backend")
        check_backend_services
        ;;
    "--ssl")
        check_ssl_certificate
        ;;
    "--disk")
        check_disk_space
        ;;
    "--network")
        check_network_connectivity
        ;;
    "--performance")
        performance_test
        ;;
    "--report")
        generate_health_report
        ;;
    "--repair")
        auto_repair
        ;;
    *)
        main_health_check
        ;;
esac
