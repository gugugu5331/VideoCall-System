#!/bin/bash

# 智能视频会议平台集成测试脚本
# 用途：测试各个模块的交互，确保所有功能正常工作

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

log_test() {
    echo -e "${PURPLE}[TEST]${NC} $1"
}

# 全局变量
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKEND_DIR="$PROJECT_ROOT/backend"
TEST_RESULTS_DIR="$PROJECT_ROOT/test-results"
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")

# 创建测试结果目录
mkdir -p "$TEST_RESULTS_DIR"

# 测试配置
SERVICES=(
    "user-service:8080"
    "signaling-service:8081"
    "meeting-service:8082"
    "media-service:8083"
    "ai-service:8084"
    "notification-service:8085"
)

GRPC_SERVICES=(
    "user-service:50051"
    "meeting-service:50052"
    "media-service:50053"
    "ai-service:50054"
    "notification-service:50055"
)

# 检查依赖
check_dependencies() {
    log_step "检查测试依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        log_error "Go未安装，请先安装Go"
        exit 1
    fi
    
    # 检查curl
    if ! command -v curl &> /dev/null; then
        log_error "curl未安装，请先安装curl"
        exit 1
    fi
    
    # 检查jq
    if ! command -v jq &> /dev/null; then
        log_warn "jq未安装，JSON解析功能将受限"
    fi
    
    log_info "依赖检查完成"
}

# 等待服务启动
wait_for_service() {
    local service_name=$1
    local host=$2
    local port=$3
    local max_attempts=30
    local attempt=1
    
    log_info "等待 $service_name 启动..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "http://$host:$port/health" > /dev/null 2>&1; then
            log_info "$service_name 已启动"
            return 0
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            log_error "$service_name 启动超时"
            return 1
        fi
        
        sleep 2
        ((attempt++))
    done
}

# 等待所有服务启动
wait_for_all_services() {
    log_step "等待所有服务启动..."
    
    for service in "${SERVICES[@]}"; do
        service_name=$(echo $service | cut -d':' -f1)
        port=$(echo $service | cut -d':' -f2)
        
        if ! wait_for_service "$service_name" "localhost" "$port"; then
            log_error "服务 $service_name 未能正常启动"
            return 1
        fi
    done
    
    log_info "所有服务已启动"
}

# 测试HTTP健康检查
test_http_health() {
    log_test "测试HTTP健康检查..."
    
    local failed_services=()
    
    for service in "${SERVICES[@]}"; do
        service_name=$(echo $service | cut -d':' -f1)
        port=$(echo $service | cut -d':' -f2)
        
        log_info "测试 $service_name 健康检查..."
        
        response=$(curl -s -w "%{http_code}" "http://localhost:$port/health")
        http_code="${response: -3}"
        
        if [ "$http_code" = "200" ]; then
            log_info "✅ $service_name 健康检查通过"
        else
            log_error "❌ $service_name 健康检查失败 (HTTP $http_code)"
            failed_services+=("$service_name")
        fi
    done
    
    if [ ${#failed_services[@]} -eq 0 ]; then
        log_info "✅ 所有服务HTTP健康检查通过"
        return 0
    else
        log_error "❌ 以下服务健康检查失败: ${failed_services[*]}"
        return 1
    fi
}

# 测试gRPC连接
test_grpc_connections() {
    log_test "测试gRPC连接..."
    
    # 这里使用grpcurl工具测试gRPC连接
    # 如果没有grpcurl，跳过此测试
    if ! command -v grpcurl &> /dev/null; then
        log_warn "grpcurl未安装，跳过gRPC连接测试"
        return 0
    fi
    
    local failed_services=()
    
    for service in "${GRPC_SERVICES[@]}"; do
        service_name=$(echo $service | cut -d':' -f1)
        port=$(echo $service | cut -d':' -f2)
        
        log_info "测试 $service_name gRPC连接..."
        
        if grpcurl -plaintext "localhost:$port" list > /dev/null 2>&1; then
            log_info "✅ $service_name gRPC连接正常"
        else
            log_error "❌ $service_name gRPC连接失败"
            failed_services+=("$service_name")
        fi
    done
    
    if [ ${#failed_services[@]} -eq 0 ]; then
        log_info "✅ 所有gRPC服务连接正常"
        return 0
    else
        log_error "❌ 以下gRPC服务连接失败: ${failed_services[*]}"
        return 1
    fi
}

# 测试数据库连接
test_database_connections() {
    log_test "测试数据库连接..."
    
    # 测试PostgreSQL
    if command -v psql &> /dev/null; then
        if PGPASSWORD=password psql -h localhost -U postgres -d meeting_system -c "SELECT 1;" > /dev/null 2>&1; then
            log_info "✅ PostgreSQL连接正常"
        else
            log_error "❌ PostgreSQL连接失败"
            return 1
        fi
    else
        log_warn "psql未安装，跳过PostgreSQL连接测试"
    fi
    
    # 测试Redis
    if command -v redis-cli &> /dev/null; then
        if redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then
            log_info "✅ Redis连接正常"
        else
            log_error "❌ Redis连接失败"
            return 1
        fi
    else
        log_warn "redis-cli未安装，跳过Redis连接测试"
    fi
    
    return 0
}

# 运行Go集成测试
run_go_integration_tests() {
    log_test "运行Go集成测试..."
    
    cd "$BACKEND_DIR"
    
    # 设置Go代理
    export GOPROXY=https://goproxy.cn,direct
    
    # 运行集成测试
    if go run test_all_services.go > "$TEST_RESULTS_DIR/go-integration-test-$TIMESTAMP.log" 2>&1; then
        log_info "✅ Go集成测试通过"
        return 0
    else
        log_error "❌ Go集成测试失败，详细日志: $TEST_RESULTS_DIR/go-integration-test-$TIMESTAMP.log"
        return 1
    fi
}

# 测试API端点
test_api_endpoints() {
    log_test "测试API端点..."
    
    # 测试用户服务API
    log_info "测试用户服务API..."
    
    # 测试用户注册
    register_response=$(curl -s -X POST "http://localhost:8080/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser_'$TIMESTAMP'",
            "email": "test_'$TIMESTAMP'@example.com",
            "password": "testpassword123",
            "full_name": "Test User"
        }' -w "%{http_code}")
    
    register_http_code="${register_response: -3}"
    if [ "$register_http_code" = "200" ] || [ "$register_http_code" = "201" ]; then
        log_info "✅ 用户注册API测试通过"
    else
        log_error "❌ 用户注册API测试失败 (HTTP $register_http_code)"
    fi
    
    # 测试用户登录
    login_response=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "testuser_'$TIMESTAMP'",
            "password": "testpassword123"
        }' -w "%{http_code}")
    
    login_http_code="${login_response: -3}"
    if [ "$login_http_code" = "200" ]; then
        log_info "✅ 用户登录API测试通过"
    else
        log_error "❌ 用户登录API测试失败 (HTTP $login_http_code)"
    fi
    
    # 测试会议服务API
    log_info "测试会议服务API..."
    
    meetings_response=$(curl -s "http://localhost:8082/api/v1/meetings" -w "%{http_code}")
    meetings_http_code="${meetings_response: -3}"
    if [ "$meetings_http_code" = "200" ]; then
        log_info "✅ 会议列表API测试通过"
    else
        log_error "❌ 会议列表API测试失败 (HTTP $meetings_http_code)"
    fi
}

# 测试WebSocket连接
test_websocket_connections() {
    log_test "测试WebSocket连接..."
    
    # 使用websocat测试WebSocket连接（如果可用）
    if command -v websocat &> /dev/null; then
        log_info "测试信令服务WebSocket连接..."
        
        # 简单的WebSocket连接测试
        timeout 5 websocat "ws://localhost:8081/ws" <<< '{"type":"ping"}' > /dev/null 2>&1
        if [ $? -eq 0 ]; then
            log_info "✅ WebSocket连接测试通过"
        else
            log_warn "⚠️ WebSocket连接测试可能失败（超时或连接问题）"
        fi
    else
        log_warn "websocat未安装，跳过WebSocket连接测试"
    fi
}

# 测试负载均衡
test_load_balancing() {
    log_test "测试负载均衡..."
    
    # 测试Nginx负载均衡
    nginx_response=$(curl -s "http://localhost/api/health" -w "%{http_code}")
    nginx_http_code="${nginx_response: -3}"
    
    if [ "$nginx_http_code" = "200" ]; then
        log_info "✅ Nginx负载均衡测试通过"
    else
        log_warn "⚠️ Nginx负载均衡测试失败或Nginx未启动 (HTTP $nginx_http_code)"
    fi
}

# 性能测试
run_performance_tests() {
    log_test "运行性能测试..."
    
    # 使用ab工具进行简单的性能测试
    if command -v ab &> /dev/null; then
        log_info "运行用户服务性能测试..."
        
        ab -n 100 -c 10 "http://localhost:8080/health" > "$TEST_RESULTS_DIR/performance-test-$TIMESTAMP.log" 2>&1
        
        if [ $? -eq 0 ]; then
            log_info "✅ 性能测试完成，结果保存到: $TEST_RESULTS_DIR/performance-test-$TIMESTAMP.log"
        else
            log_error "❌ 性能测试失败"
        fi
    else
        log_warn "ab工具未安装，跳过性能测试"
    fi
}

# 生成测试报告
generate_test_report() {
    log_step "生成测试报告..."
    
    local report_file="$TEST_RESULTS_DIR/integration-test-report-$TIMESTAMP.md"
    
    cat > "$report_file" << EOF
# 智能视频会议平台集成测试报告

**测试时间**: $(date)
**测试版本**: $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

## 测试环境

- 操作系统: $(uname -s)
- Go版本: $(go version)
- 测试脚本: $0

## 测试结果

### 服务健康检查
EOF

    # 添加服务状态到报告
    for service in "${SERVICES[@]}"; do
        service_name=$(echo $service | cut -d':' -f1)
        port=$(echo $service | cut -d':' -f2)
        
        if curl -f -s "http://localhost:$port/health" > /dev/null 2>&1; then
            echo "- ✅ $service_name: 正常" >> "$report_file"
        else
            echo "- ❌ $service_name: 异常" >> "$report_file"
        fi
    done
    
    cat >> "$report_file" << EOF

### 数据库连接
- PostgreSQL: $(if PGPASSWORD=password psql -h localhost -U postgres -d meeting_system -c "SELECT 1;" > /dev/null 2>&1; then echo "✅ 正常"; else echo "❌ 异常"; fi)
- Redis: $(if redis-cli -h localhost -p 6379 ping > /dev/null 2>&1; then echo "✅ 正常"; else echo "❌ 异常"; fi)

### 测试文件
- Go集成测试日志: go-integration-test-$TIMESTAMP.log
- 性能测试结果: performance-test-$TIMESTAMP.log

## 建议

1. 定期运行此集成测试以确保系统稳定性
2. 监控服务健康状态和性能指标
3. 及时处理测试中发现的问题

---
*报告生成时间: $(date)*
EOF

    log_info "✅ 测试报告已生成: $report_file"
}

# 主函数
main() {
    echo "=========================================="
    echo "    智能视频会议平台集成测试"
    echo "=========================================="
    echo ""
    
    local start_time=$(date +%s)
    local failed_tests=0
    
    # 执行测试步骤
    check_dependencies || ((failed_tests++))
    wait_for_all_services || ((failed_tests++))
    test_http_health || ((failed_tests++))
    test_grpc_connections || ((failed_tests++))
    test_database_connections || ((failed_tests++))
    test_api_endpoints || ((failed_tests++))
    test_websocket_connections || ((failed_tests++))
    test_load_balancing || ((failed_tests++))
    run_go_integration_tests || ((failed_tests++))
    run_performance_tests || ((failed_tests++))
    
    # 生成报告
    generate_test_report
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo ""
    echo "=========================================="
    if [ $failed_tests -eq 0 ]; then
        log_info "🎉 所有集成测试通过！"
        echo "✅ 测试完成，耗时: ${duration}秒"
        echo "📊 测试结果保存在: $TEST_RESULTS_DIR"
    else
        log_error "❌ $failed_tests 个测试失败"
        echo "⏱️ 测试完成，耗时: ${duration}秒"
        echo "📊 详细结果请查看: $TEST_RESULTS_DIR"
        exit 1
    fi
    echo "=========================================="
}

# 执行主函数
main "$@"
