#!/bin/bash

# 视频会议系统集成测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
API_BASE_URL="http://localhost:8080"
SIGNALING_URL="ws://localhost:8083"
AI_SERVICE_URL="http://localhost:8501"

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    ((TOTAL_TESTS++))
    log_info "运行测试: $test_name"
    
    if eval "$test_command"; then
        log_success "$test_name"
        return 0
    else
        log_error "$test_name"
        return 1
    fi
}

# API测试函数
test_api() {
    local endpoint="$1"
    local method="${2:-GET}"
    local data="$3"
    local expected_status="${4:-200}"
    
    local curl_cmd="curl -s -w '%{http_code}' -X $method"
    
    if [ -n "$data" ]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi
    
    curl_cmd="$curl_cmd $API_BASE_URL$endpoint"
    
    local response=$(eval "$curl_cmd")
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [ "$status_code" = "$expected_status" ]; then
        return 0
    else
        echo "Expected status $expected_status, got $status_code"
        echo "Response: $body"
        return 1
    fi
}

# 健康检查测试
test_health_checks() {
    log_info "=== 健康检查测试 ==="
    
    local services=("user-service:8081" "meeting-service:8082" "signaling-service:8083" 
                   "media-service:8084" "detection-service:8085" "record-service:8086")
    
    for service in "${services[@]}"; do
        local service_name=$(echo $service | cut -d':' -f1)
        local port=$(echo $service | cut -d':' -f2)
        
        run_test "$service_name 健康检查" \
            "curl -f -s http://localhost:$port/health > /dev/null"
    done
    
    # AI服务健康检查
    run_test "AI检测服务健康检查" \
        "curl -f -s $AI_SERVICE_URL/health > /dev/null"
}

# 用户服务测试
test_user_service() {
    log_info "=== 用户服务测试 ==="
    
    # 用户注册测试
    local register_data='{
        "username": "testuser",
        "email": "test@example.com",
        "password": "password123",
        "full_name": "Test User"
    }'
    
    run_test "用户注册" \
        "test_api '/api/v1/users/register' 'POST' '$register_data' '201'"
    
    # 用户登录测试
    local login_data='{
        "email": "test@example.com",
        "password": "password123"
    }'
    
    if run_test "用户登录" \
        "test_api '/api/v1/users/login' 'POST' '$login_data' '200'"; then
        
        # 提取JWT令牌
        local login_response=$(curl -s -X POST \
            -H 'Content-Type: application/json' \
            -d "$login_data" \
            "$API_BASE_URL/api/v1/users/login")
        
        JWT_TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        
        if [ -n "$JWT_TOKEN" ]; then
            log_info "JWT令牌获取成功"
            
            # 获取用户资料测试
            run_test "获取用户资料" \
                "curl -f -s -H 'Authorization: Bearer $JWT_TOKEN' $API_BASE_URL/api/v1/users/profile > /dev/null"
        else
            log_error "JWT令牌获取失败"
        fi
    fi
}

# 会议服务测试
test_meeting_service() {
    log_info "=== 会议服务测试 ==="
    
    if [ -z "$JWT_TOKEN" ]; then
        log_error "需要JWT令牌进行会议服务测试"
        return 1
    fi
    
    # 创建会议测试
    local meeting_data='{
        "title": "测试会议",
        "description": "这是一个测试会议",
        "start_time": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
        "duration": 60,
        "max_participants": 10,
        "is_public": false,
        "recording_enabled": true,
        "detection_enabled": true
    }'
    
    if run_test "创建会议" \
        "curl -f -s -X POST -H 'Authorization: Bearer $JWT_TOKEN' -H 'Content-Type: application/json' -d '$meeting_data' $API_BASE_URL/api/v1/meetings > /dev/null"; then
        
        # 获取会议列表
        run_test "获取会议列表" \
            "curl -f -s -H 'Authorization: Bearer $JWT_TOKEN' $API_BASE_URL/api/v1/meetings > /dev/null"
    fi
}

# AI检测服务测试
test_ai_detection() {
    log_info "=== AI检测服务测试 ==="
    
    # 创建测试图片文件
    local test_image="/tmp/test_image.jpg"
    
    # 创建一个简单的测试图片（1x1像素）
    echo -e '\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x01\x00H\x00H\x00\x00\xff\xdb\x00C\x00\x08\x06\x06\x07\x06\x05\x08\x07\x07\x07\t\t\x08\n\x0c\x14\r\x0c\x0b\x0b\x0c\x19\x12\x13\x0f\x14\x1d\x1a\x1f\x1e\x1d\x1a\x1c\x1c $.\' ",#\x1c\x1c(7),01444\x1f\'9=82<.342\xff\xc0\x00\x11\x08\x00\x01\x00\x01\x01\x01\x11\x00\x02\x11\x01\x03\x11\x01\xff\xc4\x00\x14\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x08\xff\xc4\x00\x14\x10\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\xda\x00\x0c\x03\x01\x00\x02\x11\x03\x11\x00\x3f\x00\xaa\xff\xd9' > "$test_image"
    
    if [ -f "$test_image" ]; then
        # 提交检测任务
        local detection_response=$(curl -s -X POST \
            -F "file=@$test_image" \
            -F "type=image" \
            "$AI_SERVICE_URL/detect")
        
        local task_id=$(echo "$detection_response" | grep -o '"task_id":"[^"]*"' | cut -d'"' -f4)
        
        if [ -n "$task_id" ]; then
            log_success "AI检测任务提交成功"
            
            # 等待检测完成并获取结果
            local max_attempts=10
            local attempt=1
            
            while [ $attempt -le $max_attempts ]; do
                local result_response=$(curl -s "$AI_SERVICE_URL/result/$task_id")
                local status=$(echo "$result_response" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)
                
                if [ "$status" = "completed" ]; then
                    log_success "AI检测任务完成"
                    break
                elif [ "$status" = "processing" ]; then
                    log_info "等待AI检测完成... (尝试 $attempt/$max_attempts)"
                    sleep 2
                    ((attempt++))
                else
                    log_error "AI检测任务状态异常: $status"
                    break
                fi
            done
        else
            log_error "AI检测任务提交失败"
        fi
        
        # 清理测试文件
        rm -f "$test_image"
    else
        log_error "无法创建测试图片文件"
    fi
}

# 数据库连接测试
test_database_connections() {
    log_info "=== 数据库连接测试 ==="
    
    # PostgreSQL连接测试
    run_test "PostgreSQL连接" \
        "docker-compose exec -T postgres pg_isready -U admin -d video_conference"
    
    # MongoDB连接测试
    run_test "MongoDB连接" \
        "docker-compose exec -T mongodb mongosh --quiet --eval 'db.adminCommand(\"ping\").ok' | grep -q 1"
    
    # Redis连接测试
    run_test "Redis连接" \
        "docker-compose exec -T redis redis-cli ping | grep -q PONG"
}

# 消息队列测试
test_message_queue() {
    log_info "=== 消息队列测试 ==="
    
    # RabbitMQ连接测试
    run_test "RabbitMQ连接" \
        "curl -f -s -u admin:password123 http://localhost:15672/api/overview > /dev/null"
    
    # 检查队列状态
    run_test "检查消息队列" \
        "curl -f -s -u admin:password123 http://localhost:15672/api/queues | grep -q '\"name\"'"
}

# 性能测试
test_performance() {
    log_info "=== 性能测试 ==="
    
    # API响应时间测试
    local response_time=$(curl -w '%{time_total}' -s -o /dev/null "$API_BASE_URL/api/v1/users/login" \
        -X POST -H 'Content-Type: application/json' \
        -d '{"email":"test@example.com","password":"password123"}')
    
    if (( $(echo "$response_time < 2.0" | bc -l) )); then
        log_success "API响应时间测试 (${response_time}s < 2.0s)"
        ((PASSED_TESTS++))
    else
        log_error "API响应时间测试 (${response_time}s >= 2.0s)"
        ((FAILED_TESTS++))
    fi
    ((TOTAL_TESTS++))
    
    # 内存使用测试
    local memory_usage=$(docker stats --no-stream --format "table {{.Container}}\t{{.MemUsage}}" | grep -E "(user-service|meeting-service)" | head -1 | awk '{print $2}' | cut -d'/' -f1)
    
    if [ -n "$memory_usage" ]; then
        log_info "服务内存使用: $memory_usage"
    fi
}

# 安全测试
test_security() {
    log_info "=== 安全测试 ==="
    
    # 未授权访问测试
    run_test "未授权访问保护" \
        "! curl -f -s $API_BASE_URL/api/v1/users/profile > /dev/null 2>&1"
    
    # SQL注入测试
    local malicious_data='{"email":"test@example.com'\''OR 1=1--","password":"anything"}'
    run_test "SQL注入防护" \
        "! test_api '/api/v1/users/login' 'POST' '$malicious_data' '200'"
    
    # XSS测试
    local xss_data='{"username":"<script>alert(1)</script>","email":"xss@test.com","password":"password123","full_name":"XSS Test"}'
    run_test "XSS防护" \
        "! test_api '/api/v1/users/register' 'POST' '$xss_data' '201'"
}

# 清理测试数据
cleanup_test_data() {
    log_info "清理测试数据..."
    
    # 删除测试用户（如果存在）
    if [ -n "$JWT_TOKEN" ]; then
        curl -s -X DELETE \
            -H "Authorization: Bearer $JWT_TOKEN" \
            "$API_BASE_URL/api/v1/users/profile" > /dev/null 2>&1 || true
    fi
    
    log_info "测试数据清理完成"
}

# 生成测试报告
generate_report() {
    echo
    echo "=== 测试报告 ==="
    echo "总测试数: $TOTAL_TESTS"
    echo "通过测试: $PASSED_TESTS"
    echo "失败测试: $FAILED_TESTS"
    echo "成功率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    echo
    
    if [ $FAILED_TESTS -eq 0 ]; then
        log_success "所有测试通过！"
        return 0
    else
        log_error "有 $FAILED_TESTS 个测试失败"
        return 1
    fi
}

# 主函数
main() {
    local test_type="${1:-all}"
    
    log_info "开始运行视频会议系统集成测试..."
    log_info "测试类型: $test_type"
    echo
    
    case "$test_type" in
        "all")
            test_health_checks
            test_database_connections
            test_message_queue
            test_user_service
            test_meeting_service
            test_ai_detection
            test_performance
            test_security
            ;;
        "health")
            test_health_checks
            ;;
        "database")
            test_database_connections
            ;;
        "api")
            test_user_service
            test_meeting_service
            ;;
        "ai")
            test_ai_detection
            ;;
        "performance")
            test_performance
            ;;
        "security")
            test_security
            ;;
        *)
            echo "用法: $0 [test_type]"
            echo "测试类型: all, health, database, api, ai, performance, security"
            exit 1
            ;;
    esac
    
    cleanup_test_data
    generate_report
}

# 执行主函数
main "$@"
