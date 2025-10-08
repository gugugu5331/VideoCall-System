#!/bin/bash

################################################################################
# 微服务集成测试脚本
# 测试服务发现、服务注册和 Nginx 网关路由功能
################################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
NGINX_GATEWAY="${NGINX_GATEWAY:-http://localhost:80}"
ETCD_ENDPOINT="${ETCD_ENDPOINT:-localhost:2379}"
TEST_TIMEOUT="${TEST_TIMEOUT:-300}"

echo "================================================================================"
echo "Microservices Integration Test"
echo "================================================================================"
echo -e "${BLUE}Nginx Gateway:${NC} $NGINX_GATEWAY"
echo -e "${BLUE}Etcd Endpoint:${NC} $ETCD_ENDPOINT"
echo -e "${BLUE}Test Timeout:${NC} ${TEST_TIMEOUT}s"
echo "================================================================================"
echo ""

# 检查依赖
check_dependencies() {
    echo -e "${BLUE}[1/6]${NC} Checking dependencies..."
    
    local missing_deps=()
    
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        echo -e "${RED}✗ Missing dependencies: ${missing_deps[*]}${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ All dependencies installed${NC}"
}

# 检查 etcd 连接
check_etcd() {
    echo -e "\n${BLUE}[2/6]${NC} Checking etcd connection..."
    
    if command -v etcdctl &> /dev/null; then
        if etcdctl --endpoints="$ETCD_ENDPOINT" endpoint health &> /dev/null; then
            echo -e "${GREEN}✓ Etcd is healthy${NC}"
        else
            echo -e "${YELLOW}⚠ Etcd health check failed (may not be running)${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ etcdctl not installed, skipping etcd health check${NC}"
    fi
}

# 检查 Nginx 网关
check_nginx() {
    echo -e "\n${BLUE}[3/6]${NC} Checking Nginx gateway..."
    
    if curl -s -f "$NGINX_GATEWAY/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Nginx gateway is healthy${NC}"
    else
        echo -e "${RED}✗ Nginx gateway is not accessible${NC}"
        echo -e "${YELLOW}  Make sure Nginx is running and accessible at $NGINX_GATEWAY${NC}"
        exit 1
    fi
}

# 检查服务注册
check_service_registration() {
    echo -e "\n${BLUE}[4/6]${NC} Checking service registration in etcd..."
    
    if command -v etcdctl &> /dev/null; then
        local services=$(etcdctl --endpoints="$ETCD_ENDPOINT" get /services/ --prefix --keys-only 2>/dev/null | grep -o '/services/[^/]*' | sort -u | sed 's|/services/||' || echo "")
        
        if [ -n "$services" ]; then
            echo -e "${GREEN}✓ Registered services:${NC}"
            echo "$services" | while read -r svc; do
                local count=$(etcdctl --endpoints="$ETCD_ENDPOINT" get "/services/$svc/" --prefix --keys-only 2>/dev/null | wc -l)
                echo "  - $svc: $count instance(s)"
            done
        else
            echo -e "${YELLOW}⚠ No services registered in etcd${NC}"
        fi
    else
        echo -e "${YELLOW}⚠ etcdctl not installed, skipping service registration check${NC}"
    fi
}

# 测试服务路由
test_service_routing() {
    echo -e "\n${BLUE}[5/6]${NC} Testing service routing through Nginx..."
    
    local services=(
        "user-service:/api/v1/auth/health"
        "meeting-service:/api/v1/meetings"
        "media-service:/api/v1/media/health"
        "ai-service:/api/v1/ai/health"
    )
    
    local passed=0
    local failed=0
    
    for svc_endpoint in "${services[@]}"; do
        local svc_name="${svc_endpoint%%:*}"
        local endpoint="${svc_endpoint#*:}"
        local url="$NGINX_GATEWAY$endpoint"
        
        echo -n "  Testing $svc_name... "
        
        local status_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
        
        # 502/503/504 表示网关错误（服务不可达）
        # 其他状态码表示路由成功（即使是401/404等）
        if [ "$status_code" = "502" ] || [ "$status_code" = "503" ] || [ "$status_code" = "504" ] || [ "$status_code" = "000" ]; then
            echo -e "${RED}✗ FAIL (status: $status_code)${NC}"
            ((failed++))
        else
            echo -e "${GREEN}✓ PASS (status: $status_code)${NC}"
            ((passed++))
        fi
    done
    
    echo ""
    echo "  Routing test results: $passed passed, $failed failed"
    
    if [ $failed -gt 0 ]; then
        echo -e "${YELLOW}  ⚠ Some services are not accessible through Nginx${NC}"
    fi
}

# 运行 Go 集成测试
run_go_tests() {
    echo -e "\n${BLUE}[6/6]${NC} Running Go integration tests..."
    
    cd "$(dirname "$0")"
    
    export NGINX_GATEWAY
    export ETCD_ENDPOINT
    
    if go test -v -timeout "${TEST_TIMEOUT}s" -run "^Test" ./microservices_integration_test.go; then
        echo -e "\n${GREEN}✓ All Go tests passed${NC}"
        return 0
    else
        echo -e "\n${RED}✗ Some Go tests failed${NC}"
        return 1
    fi
}

# 生成测试报告
generate_report() {
    local exit_code=$1
    
    echo ""
    echo "================================================================================"
    echo "Test Execution Summary"
    echo "================================================================================"
    
    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}✅ ALL TESTS PASSED${NC}"
        echo ""
        echo "Service Discovery: ✓ Working"
        echo "Service Registration: ✓ Working"
        echo "Nginx Gateway Routing: ✓ Working"
        echo "Microservices Integration: ✓ Working"
    else
        echo -e "${RED}❌ SOME TESTS FAILED${NC}"
        echo ""
        echo "Please check the test output above for details."
    fi
    
    echo "================================================================================"
}

# 主函数
main() {
    local exit_code=0
    
    check_dependencies
    check_etcd
    check_nginx
    check_service_registration
    test_service_routing
    
    if run_go_tests; then
        exit_code=0
    else
        exit_code=1
    fi
    
    generate_report $exit_code
    
    exit $exit_code
}

# 执行主函数
main

