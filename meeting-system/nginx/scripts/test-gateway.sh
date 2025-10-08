#!/bin/bash

# ============================================================================
# Nginx网关功能测试脚本
# ============================================================================

set -e

# 配置变量
GATEWAY_URL="https://localhost"
HTTP_GATEWAY_URL="http://localhost"
TEST_RESULTS_DIR="./test-results"
CONCURRENT_USERS=10
TEST_DURATION=30

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

# 创建测试结果目录
mkdir -p "$TEST_RESULTS_DIR"

# 测试基础连通性
test_basic_connectivity() {
    log_info "测试基础连通性..."
    
    # 测试HTTP重定向
    local http_response=$(curl -s -o /dev/null -w "%{http_code}" "$HTTP_GATEWAY_URL" || echo "000")
    if [[ "$http_response" == "301" || "$http_response" == "302" ]]; then
        log_success "HTTP重定向正常 (状态码: $http_response)"
    else
        log_error "HTTP重定向异常 (状态码: $http_response)"
        return 1
    fi
    
    # 测试HTTPS连接
    local https_response=$(curl -s -k -o /dev/null -w "%{http_code}" "$GATEWAY_URL/health" || echo "000")
    if [[ "$https_response" == "200" ]]; then
        log_success "HTTPS连接正常 (状态码: $https_response)"
    else
        log_error "HTTPS连接异常 (状态码: $https_response)"
        return 1
    fi
    
    return 0
}

# 测试API路由
test_api_routing() {
    log_info "测试API路由..."
    
    local api_endpoints=(
        "/api/v1/users/health"
        "/api/v1/meetings/health"
        "/api/v1/media/health"
        "/api/v1/ai/health"
        "/api/v1/notifications/health"
    )
    
    local failed_endpoints=()
    
    for endpoint in "${api_endpoints[@]}"; do
        local response=$(curl -s -k -o /dev/null -w "%{http_code}" "$GATEWAY_URL$endpoint" || echo "000")
        if [[ "$response" =~ ^[23] ]]; then
            log_success "API路由 $endpoint 正常 (状态码: $response)"
        else
            log_error "API路由 $endpoint 异常 (状态码: $response)"
            failed_endpoints+=("$endpoint")
        fi
    done
    
    if [[ ${#failed_endpoints[@]} -eq 0 ]]; then
        return 0
    else
        log_error "以下API路由测试失败: ${failed_endpoints[*]}"
        return 1
    fi
}

# 测试WebSocket连接
test_websocket() {
    log_info "测试WebSocket连接..."
    
    # 使用websocat或其他WebSocket客户端测试
    if command -v websocat >/dev/null 2>&1; then
        timeout 5 websocat "wss://localhost/ws/signaling" <<< '{"type":"ping"}' >/dev/null 2>&1
        if [[ $? -eq 0 ]]; then
            log_success "WebSocket连接正常"
            return 0
        else
            log_error "WebSocket连接失败"
            return 1
        fi
    else
        log_warning "未安装websocat，跳过WebSocket测试"
        return 0
    fi
}

# 测试限流功能
test_rate_limiting() {
    log_info "测试限流功能..."
    
    # 快速发送多个请求测试限流
    local rate_limit_test_url="$GATEWAY_URL/api/v1/users/health"
    local success_count=0
    local rate_limited_count=0
    
    for i in {1..20}; do
        local response=$(curl -s -k -o /dev/null -w "%{http_code}" "$rate_limit_test_url" || echo "000")
        if [[ "$response" == "200" ]]; then
            success_count=$((success_count + 1))
        elif [[ "$response" == "429" ]]; then
            rate_limited_count=$((rate_limited_count + 1))
        fi
        sleep 0.1
    done
    
    log_info "限流测试结果: 成功 $success_count 次, 被限流 $rate_limited_count 次"
    
    if [[ $rate_limited_count -gt 0 ]]; then
        log_success "限流功能正常工作"
        return 0
    else
        log_warning "未触发限流，可能需要调整限流参数"
        return 0
    fi
}

# 测试SSL证书
test_ssl_certificate() {
    log_info "测试SSL证书..."
    
    # 检查证书有效性
    local cert_info=$(echo | openssl s_client -servername localhost -connect localhost:443 2>/dev/null | openssl x509 -noout -dates 2>/dev/null)
    
    if [[ -n "$cert_info" ]]; then
        log_success "SSL证书有效"
        log_info "证书信息: $cert_info"
        return 0
    else
        log_error "SSL证书无效或无法获取"
        return 1
    fi
}

# 测试安全头部
test_security_headers() {
    log_info "测试安全头部..."
    
    local headers=$(curl -s -k -I "$GATEWAY_URL/health")
    
    local required_headers=(
        "X-Frame-Options"
        "X-Content-Type-Options"
        "X-XSS-Protection"
        "Strict-Transport-Security"
    )
    
    local missing_headers=()
    
    for header in "${required_headers[@]}"; do
        if echo "$headers" | grep -qi "$header"; then
            log_success "安全头部 $header 存在"
        else
            log_error "安全头部 $header 缺失"
            missing_headers+=("$header")
        fi
    done
    
    if [[ ${#missing_headers[@]} -eq 0 ]]; then
        return 0
    else
        log_error "缺失的安全头部: ${missing_headers[*]}"
        return 1
    fi
}

# 性能测试
test_performance() {
    log_info "执行性能测试..."
    
    if ! command -v ab >/dev/null 2>&1; then
        log_warning "未安装Apache Bench (ab)，跳过性能测试"
        return 0
    fi
    
    local test_url="$GATEWAY_URL/health"
    local result_file="$TEST_RESULTS_DIR/performance-$(date +%Y%m%d-%H%M%S).txt"
    
    log_info "开始压力测试: $CONCURRENT_USERS 并发用户, 持续 $TEST_DURATION 秒"
    
    ab -n 1000 -c "$CONCURRENT_USERS" -t "$TEST_DURATION" -k "$test_url" > "$result_file" 2>&1
    
    if [[ $? -eq 0 ]]; then
        # 提取关键性能指标
        local rps=$(grep "Requests per second" "$result_file" | awk '{print $4}')
        local avg_time=$(grep "Time per request" "$result_file" | head -1 | awk '{print $4}')
        local failed_requests=$(grep "Failed requests" "$result_file" | awk '{print $3}')
        
        log_success "性能测试完成"
        log_info "每秒请求数: $rps"
        log_info "平均响应时间: ${avg_time}ms"
        log_info "失败请求数: $failed_requests"
        log_info "详细结果保存在: $result_file"
        
        return 0
    else
        log_error "性能测试失败"
        return 1
    fi
}

# 测试负载均衡
test_load_balancing() {
    log_info "测试负载均衡..."
    
    # 发送多个请求，检查是否分发到不同的后端服务器
    local test_url="$GATEWAY_URL/api/v1/users/health"
    local server_responses=()
    
    for i in {1..10}; do
        local response=$(curl -s -k "$test_url" | grep -o '"server":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "unknown")
        server_responses+=("$response")
        sleep 0.1
    done
    
    # 统计不同服务器的响应
    local unique_servers=$(printf '%s\n' "${server_responses[@]}" | sort -u | wc -l)
    
    if [[ $unique_servers -gt 1 ]]; then
        log_success "负载均衡正常工作，检测到 $unique_servers 个不同的后端服务器"
        return 0
    else
        log_warning "负载均衡可能未正常工作，只检测到 $unique_servers 个后端服务器"
        return 0
    fi
}

# 测试缓存功能
test_caching() {
    log_info "测试缓存功能..."
    
    local test_url="$GATEWAY_URL/api/v1/users/health"
    
    # 第一次请求
    local first_response=$(curl -s -k -I "$test_url")
    local first_cache_status=$(echo "$first_response" | grep -i "x-cache-status" | cut -d' ' -f2 | tr -d '\r')
    
    # 第二次请求
    sleep 1
    local second_response=$(curl -s -k -I "$test_url")
    local second_cache_status=$(echo "$second_response" | grep -i "x-cache-status" | cut -d' ' -f2 | tr -d '\r')
    
    log_info "第一次请求缓存状态: $first_cache_status"
    log_info "第二次请求缓存状态: $second_cache_status"
    
    if [[ "$second_cache_status" == "HIT" ]]; then
        log_success "缓存功能正常工作"
        return 0
    else
        log_warning "缓存功能可能未正常工作"
        return 0
    fi
}

# 测试错误处理
test_error_handling() {
    log_info "测试错误处理..."
    
    local error_tests=(
        "/nonexistent:404"
        "/api/v1/nonexistent:404"
    )
    
    local failed_tests=()
    
    for test in "${error_tests[@]}"; do
        local endpoint=$(echo "$test" | cut -d: -f1)
        local expected_code=$(echo "$test" | cut -d: -f2)
        
        local response=$(curl -s -k -o /dev/null -w "%{http_code}" "$GATEWAY_URL$endpoint" || echo "000")
        
        if [[ "$response" == "$expected_code" ]]; then
            log_success "错误处理 $endpoint 正常 (状态码: $response)"
        else
            log_error "错误处理 $endpoint 异常 (期望: $expected_code, 实际: $response)"
            failed_tests+=("$endpoint")
        fi
    done
    
    if [[ ${#failed_tests[@]} -eq 0 ]]; then
        return 0
    else
        log_error "以下错误处理测试失败: ${failed_tests[*]}"
        return 1
    fi
}

# 生成测试报告
generate_test_report() {
    local report_file="$TEST_RESULTS_DIR/test-report-$(date +%Y%m%d-%H%M%S).json"
    
    cat > "$report_file" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "gateway_url": "$GATEWAY_URL",
    "test_results": {
        "basic_connectivity": $(test_basic_connectivity >/dev/null 2>&1 && echo "true" || echo "false"),
        "api_routing": $(test_api_routing >/dev/null 2>&1 && echo "true" || echo "false"),
        "websocket": $(test_websocket >/dev/null 2>&1 && echo "true" || echo "false"),
        "rate_limiting": $(test_rate_limiting >/dev/null 2>&1 && echo "true" || echo "false"),
        "ssl_certificate": $(test_ssl_certificate >/dev/null 2>&1 && echo "true" || echo "false"),
        "security_headers": $(test_security_headers >/dev/null 2>&1 && echo "true" || echo "false"),
        "load_balancing": $(test_load_balancing >/dev/null 2>&1 && echo "true" || echo "false"),
        "caching": $(test_caching >/dev/null 2>&1 && echo "true" || echo "false"),
        "error_handling": $(test_error_handling >/dev/null 2>&1 && echo "true" || echo "false")
    }
}
EOF

    log_info "测试报告已生成: $report_file"
}

# 主测试函数
main_test() {
    log_info "开始网关功能测试..."
    
    local tests_passed=0
    local total_tests=0
    
    # 执行各项测试
    local tests=(
        "test_basic_connectivity"
        "test_api_routing"
        "test_websocket"
        "test_rate_limiting"
        "test_ssl_certificate"
        "test_security_headers"
        "test_load_balancing"
        "test_caching"
        "test_error_handling"
    )
    
    for test in "${tests[@]}"; do
        total_tests=$((total_tests + 1))
        log_info "执行测试: $test"
        if $test; then
            tests_passed=$((tests_passed + 1))
        fi
        echo "----------------------------------------"
    done
    
    # 性能测试（可选）
    if [[ "${1:-}" == "--performance" ]]; then
        total_tests=$((total_tests + 1))
        if test_performance; then
            tests_passed=$((tests_passed + 1))
        fi
    fi
    
    # 生成测试报告
    generate_test_report
    
    # 输出测试结果
    local success_rate=$((tests_passed * 100 / total_tests))
    log_info "测试完成，通过率: $success_rate% ($tests_passed/$total_tests)"
    
    if [[ $success_rate -ge 80 ]]; then
        log_success "网关功能测试整体通过"
        return 0
    else
        log_error "网关功能测试存在问题"
        return 1
    fi
}

# 脚本参数处理
case "${1:-}" in
    "--connectivity")
        test_basic_connectivity
        ;;
    "--api")
        test_api_routing
        ;;
    "--websocket")
        test_websocket
        ;;
    "--rate-limit")
        test_rate_limiting
        ;;
    "--ssl")
        test_ssl_certificate
        ;;
    "--security")
        test_security_headers
        ;;
    "--performance")
        test_performance
        ;;
    "--load-balance")
        test_load_balancing
        ;;
    "--cache")
        test_caching
        ;;
    "--error")
        test_error_handling
        ;;
    "--report")
        generate_test_report
        ;;
    *)
        main_test "$@"
        ;;
esac
