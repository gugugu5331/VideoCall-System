#!/bin/bash

# ============================================================================
# AI Inference Service - Nginx Gateway Integration Test
# ============================================================================
# 
# 这个脚本测试 AI Inference Service 通过 Nginx 网关的外部访问能力
#
# 使用方法:
#   ./test_nginx_gateway_integration.sh [nginx_host] [nginx_port] [ai_service_host] [ai_service_port]
#
# 示例:
#   ./test_nginx_gateway_integration.sh localhost 8800 localhost 8085
# ============================================================================

# 默认配置
NGINX_HOST="${1:-localhost}"
NGINX_PORT="${2:-8800}"
AI_SERVICE_HOST="${3:-localhost}"
AI_SERVICE_PORT="${4:-8085}"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 统计变量
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试结果数组
declare -a TEST_RESULTS

# 打印标题
print_header() {
    echo -e "${BLUE}============================================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}============================================================================${NC}"
}

# 打印子标题
print_subheader() {
    echo -e "\n${YELLOW}--- $1 ---${NC}"
}

# 打印成功消息
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# 打印失败消息
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# 打印信息
print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# 执行测试
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_pattern="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "\n${BLUE}Test $TOTAL_TESTS: $test_name${NC}"
    
    # 执行命令并捕获输出
    local output=$(eval "$test_command" 2>&1)
    local exit_code=$?
    
    # 检查结果
    if [ $exit_code -eq 0 ] && echo "$output" | grep -q "$expected_pattern"; then
        print_success "PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        TEST_RESULTS+=("✓ $test_name")
    else
        print_error "FAILED"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        TEST_RESULTS+=("✗ $test_name")
        echo "Expected pattern: $expected_pattern"
        echo "Output: $output"
    fi
}

# 开始测试
print_header "AI Inference Service - Nginx Gateway Integration Test"

echo "Configuration:"
echo "  Nginx Gateway: http://$NGINX_HOST:$NGINX_PORT"
echo "  AI Service:    http://$AI_SERVICE_HOST:$AI_SERVICE_PORT"
echo ""

# ============================================================================
# 1. 基础设施检查
# ============================================================================
print_subheader "1. Infrastructure Services Check"

print_info "Checking Docker containers..."
docker ps | grep -E "postgres|redis|etcd|jaeger|nginx|ai-inference" || print_error "Some containers are not running"

print_info "Checking Edge-LLM-Infra..."
netstat -tlnp 2>/dev/null | grep 19001 && print_success "unit-manager is running" || print_error "unit-manager is not running"

# ============================================================================
# 2. 直接访问测试（绕过 Nginx）
# ============================================================================
print_subheader "2. Direct Access Tests (Bypass Nginx)"

run_test "Direct Health Check" \
    "curl -s http://$AI_SERVICE_HOST:$AI_SERVICE_PORT/health" \
    "ok"

run_test "Direct AI Service Info" \
    "curl -s http://$AI_SERVICE_HOST:$AI_SERVICE_PORT/api/v1/ai/info" \
    "ai-inference-service"

run_test "Direct AI Health Check" \
    "curl -s http://$AI_SERVICE_HOST:$AI_SERVICE_PORT/api/v1/ai/health" \
    "code"

# ============================================================================
# 3. Nginx 网关访问测试
# ============================================================================
print_subheader "3. Nginx Gateway Access Tests"

run_test "Nginx Gateway Health Check" \
    "curl -s http://$NGINX_HOST:$NGINX_PORT/api/v1/ai/health" \
    "code"

run_test "Nginx Gateway Service Info" \
    "curl -s http://$NGINX_HOST:$NGINX_PORT/api/v1/ai/info" \
    "ai-inference-service"

# ============================================================================
# 4. AI 功能测试（通过 Nginx）
# ============================================================================
print_subheader "4. AI Functionality Tests (Through Nginx)"

run_test "ASR (Speech Recognition) via Nginx" \
    "curl -s -X POST http://$NGINX_HOST:$NGINX_PORT/api/v1/ai/asr -H 'Content-Type: application/json' -d '{\"audio_data\":\"c2FtcGxlIGF1ZGlvIGRhdGE=\",\"format\":\"wav\",\"sample_rate\":16000}'" \
    "code"

run_test "Emotion Detection via Nginx" \
    "curl -s -X POST http://$NGINX_HOST:$NGINX_PORT/api/v1/ai/emotion -H 'Content-Type: application/json' -d '{\"text\":\"I am very happy today!\"}'" \
    "code"

run_test "Synthesis Detection via Nginx" \
    "curl -s -X POST http://$NGINX_HOST:$NGINX_PORT/api/v1/ai/synthesis -H 'Content-Type: application/json' -d '{\"audio_data\":\"c2FtcGxlIGF1ZGlvIGRhdGE=\",\"format\":\"wav\",\"sample_rate\":16000}'" \
    "code"

# ============================================================================
# 5. 服务注册验证
# ============================================================================
print_subheader "5. Service Registration Verification"

print_info "Checking Etcd service registration..."
docker exec meeting-etcd etcdctl get --prefix "/services/ai-inference-service" 2>/dev/null && \
    print_success "Service registered in Etcd" || \
    print_error "Service not found in Etcd"

# ============================================================================
# 6. 日志检查
# ============================================================================
print_subheader "6. Log Verification"

print_info "Checking AI service logs..."
docker logs meeting-ai-inference-service 2>&1 | tail -10

print_info "Checking Nginx access logs..."
docker exec meeting-nginx tail -5 /var/log/nginx/access.log 2>/dev/null || echo "No access logs"

# ============================================================================
# 测试总结
# ============================================================================
print_header "Test Summary"

echo ""
echo "Total Tests:  $TOTAL_TESTS"
echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
echo -e "${RED}Failed:       $FAILED_TESTS${NC}"
echo ""

echo "Test Results:"
for result in "${TEST_RESULTS[@]}"; do
    echo "  $result"
done

echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    print_success "ALL TESTS PASSED! ✓"
    exit 0
else
    print_error "SOME TESTS FAILED! ✗"
    exit 1
fi

