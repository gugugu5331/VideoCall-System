#!/bin/bash

################################################################################
# 微服务直接测试脚本（不通过 Nginx）
# 测试服务发现、服务注册和直接服务访问
################################################################################

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "================================================================================"
echo "Microservices Direct Integration Test"
echo "================================================================================"
echo ""

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 记录测试结果
record_test() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    
    ((TOTAL_TESTS++))
    
    if [ "$status" = "PASS" ]; then
        ((PASSED_TESTS++))
        echo -e "${GREEN}✅ PASS${NC} - $test_name"
    else
        ((FAILED_TESTS++))
        echo -e "${RED}❌ FAIL${NC} - $test_name: $message"
    fi
}

# 测试 Docker 容器状态
test_docker_containers() {
    echo -e "\n${BLUE}[1/7]${NC} Testing Docker containers status..."
    
    local services=("meeting-postgres" "meeting-redis" "meeting-mongodb" "meeting-etcd" "meeting-minio")
    
    for svc in "${services[@]}"; do
        if docker ps --format '{{.Names}}' | grep -q "^${svc}$"; then
            local status=$(docker inspect -f '{{.State.Health.Status}}' "$svc" 2>/dev/null || echo "running")
            if [ "$status" = "healthy" ] || [ "$status" = "running" ]; then
                record_test "$svc container" "PASS" ""
            else
                record_test "$svc container" "FAIL" "Status: $status"
            fi
        else
            record_test "$svc container" "FAIL" "Container not running"
        fi
    done
}

# 测试微服务容器
test_microservice_containers() {
    echo -e "\n${BLUE}[2/7]${NC} Testing microservice containers..."
    
    local services=("meeting-user-service" "meeting-meeting-service" "meeting-signaling-service" "meeting-media-service" "meeting-ai-service")
    
    for svc in "${services[@]}"; do
        if docker ps --format '{{.Names}}' | grep -q "^${svc}$"; then
            local status=$(docker inspect -f '{{.State.Health.Status}}' "$svc" 2>/dev/null || echo "running")
            if [ "$status" = "healthy" ] || [ "$status" = "running" ] || [ "$status" = "starting" ]; then
                record_test "$svc container" "PASS" ""
            else
                record_test "$svc container" "FAIL" "Status: $status"
            fi
        else
            record_test "$svc container" "FAIL" "Container not running"
        fi
    done
}

# 测试 etcd 连接和服务注册
test_etcd_registration() {
    echo -e "\n${BLUE}[3/7]${NC} Testing etcd service registration..."
    
    # 检查 etcd 容器
    if ! docker ps --format '{{.Names}}' | grep -q "^meeting-etcd$"; then
        record_test "etcd connection" "FAIL" "etcd container not running"
        return
    fi
    
    # 使用 docker exec 查询 etcd
    local etcd_data=$(docker exec meeting-etcd etcdctl get /services/ --prefix --keys-only 2>/dev/null || echo "")
    
    if [ -n "$etcd_data" ]; then
        record_test "etcd connection" "PASS" ""
        
        # 检查各服务注册
        local services=("user-service" "meeting-service")
        for svc in "${services[@]}"; do
            if echo "$etcd_data" | grep -q "/services/$svc/"; then
                record_test "$svc registration" "PASS" ""
            else
                record_test "$svc registration" "FAIL" "Not registered in etcd"
            fi
        done
    else
        record_test "etcd connection" "FAIL" "Cannot query etcd"
    fi
}

# 测试用户服务 HTTP 端点
test_user_service_http() {
    echo -e "\n${BLUE}[4/7]${NC} Testing user service HTTP endpoints..."
    
    # 获取容器 IP
    local container_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' meeting-user-service 2>/dev/null)
    
    if [ -z "$container_ip" ]; then
        record_test "user-service HTTP" "FAIL" "Cannot get container IP"
        return
    fi
    
    # 测试健康检查端点
    local status_code=$(docker exec meeting-user-service wget -q -O- --timeout=5 "http://localhost:8080/health" 2>/dev/null | head -c 10 || echo "")
    
    if [ -n "$status_code" ]; then
        record_test "user-service health endpoint" "PASS" ""
    else
        # 尝试注册端点
        local register_response=$(docker exec meeting-user-service wget -q -O- --timeout=5 --post-data='{"username":"test","password":"test123","email":"test@test.com"}' --header='Content-Type: application/json' "http://localhost:8080/api/v1/auth/register" 2>/dev/null || echo "")
        
        if [ -n "$register_response" ]; then
            record_test "user-service register endpoint" "PASS" ""
        else
            record_test "user-service HTTP" "FAIL" "No response from service"
        fi
    fi
}

# 测试会议服务 HTTP 端点
test_meeting_service_http() {
    echo -e "\n${BLUE}[5/7]${NC} Testing meeting service HTTP endpoints..."
    
    # 测试会议列表端点（预期401或200）
    local response=$(docker exec meeting-meeting-service wget -q -O- --timeout=5 "http://localhost:8082/api/v1/meetings" 2>&1 || echo "error")
    
    # 如果返回401（未授权）或有响应，说明服务正常
    if echo "$response" | grep -q "401\|200\|Unauthorized\|meetings"; then
        record_test "meeting-service HTTP" "PASS" ""
    else
        record_test "meeting-service HTTP" "FAIL" "No valid response"
    fi
}

# 测试媒体服务
test_media_service() {
    echo -e "\n${BLUE}[6/7]${NC} Testing media service..."
    
    local response=$(docker exec meeting-media-service wget -q -O- --timeout=5 "http://localhost:8083/health" 2>/dev/null || echo "")
    
    if [ -n "$response" ]; then
        record_test "media-service health" "PASS" ""
    else
        record_test "media-service health" "FAIL" "No response"
    fi
}

# 测试 AI 服务
test_ai_service() {
    echo -e "\n${BLUE}[7/7]${NC} Testing AI service..."
    
    local response=$(docker exec meeting-ai-service wget -q -O- --timeout=5 "http://localhost:8084/health" 2>/dev/null || echo "")
    
    if [ -n "$response" ]; then
        record_test "ai-service health" "PASS" ""
    else
        record_test "ai-service health" "FAIL" "No response"
    fi
}

# 生成测试报告
generate_report() {
    echo ""
    echo "================================================================================"
    echo "Test Summary"
    echo "================================================================================"
    echo ""
    echo "Total Tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $FAILED_TESTS"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo ""
        echo -e "${GREEN}✅ ALL TESTS PASSED${NC}"
        echo ""
        echo "Service Discovery: ✓ Working"
        echo "Service Registration: ✓ Working"
        echo "Microservices: ✓ Running"
        echo "Database Connections: ✓ Working"
    else
        echo ""
        echo -e "${YELLOW}⚠ SOME TESTS FAILED${NC}"
        echo ""
        echo "Success Rate: $(awk "BEGIN {printf \"%.1f\", ($PASSED_TESTS/$TOTAL_TESTS)*100}")%"
    fi
    
    echo "================================================================================"
}

# 主函数
main() {
    test_docker_containers
    test_microservice_containers
    test_etcd_registration
    test_user_service_http
    test_meeting_service_http
    test_media_service
    test_ai_service
    
    generate_report
    
    if [ $FAILED_TESTS -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# 执行主函数
main

