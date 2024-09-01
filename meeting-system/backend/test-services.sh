#!/bin/bash

# 测试微服务脚本
# 用于验证用户服务和会议服务是否正常运行

set -e

echo "🚀 开始测试微服务..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
USER_SERVICE_URL="http://localhost:8081"
MEETING_SERVICE_URL="http://localhost:8082"
TEST_USER_EMAIL="test@example.com"
TEST_USER_PASSWORD="password123"

# 等待服务启动
wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1

    echo -e "${YELLOW}等待 $service_name 启动...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s "$url/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✅ $service_name 已启动${NC}"
            return 0
        fi
        
        echo -e "${YELLOW}⏳ 等待 $service_name 启动 (尝试 $attempt/$max_attempts)${NC}"
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}❌ $service_name 启动超时${NC}"
    return 1
}

# 测试健康检查
test_health_check() {
    local url=$1
    local service_name=$2
    
    echo -e "${BLUE}🔍 测试 $service_name 健康检查...${NC}"
    
    response=$(curl -s "$url/health")
    if echo "$response" | grep -q '"status":"ok"'; then
        echo -e "${GREEN}✅ $service_name 健康检查通过${NC}"
        return 0
    else
        echo -e "${RED}❌ $service_name 健康检查失败${NC}"
        echo "响应: $response"
        return 1
    fi
}

# 测试用户注册
test_user_registration() {
    echo -e "${BLUE}🔍 测试用户注册...${NC}"
    
    response=$(curl -s -X POST "$USER_SERVICE_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"testuser\",
            \"email\": \"$TEST_USER_EMAIL\",
            \"password\": \"$TEST_USER_PASSWORD\",
            \"nickname\": \"Test User\"
        }")
    
    if echo "$response" | grep -q '"message":"User registered successfully"'; then
        echo -e "${GREEN}✅ 用户注册成功${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️ 用户可能已存在或注册失败${NC}"
        echo "响应: $response"
        return 0  # 不作为错误，可能用户已存在
    fi
}

# 测试用户登录
test_user_login() {
    echo -e "${BLUE}🔍 测试用户登录...${NC}"
    
    response=$(curl -s -X POST "$USER_SERVICE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"email\": \"$TEST_USER_EMAIL\",
            \"password\": \"$TEST_USER_PASSWORD\"
        }")
    
    if echo "$response" | grep -q '"access_token"'; then
        echo -e "${GREEN}✅ 用户登录成功${NC}"
        # 提取token
        ACCESS_TOKEN=$(echo "$response" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
        echo "Token: ${ACCESS_TOKEN:0:20}..."
        return 0
    else
        echo -e "${RED}❌ 用户登录失败${NC}"
        echo "响应: $response"
        return 1
    fi
}

# 测试获取用户信息
test_get_profile() {
    echo -e "${BLUE}🔍 测试获取用户信息...${NC}"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        echo -e "${RED}❌ 没有访问令牌${NC}"
        return 1
    fi
    
    response=$(curl -s -X GET "$USER_SERVICE_URL/api/v1/users/profile" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    if echo "$response" | grep -q '"email"'; then
        echo -e "${GREEN}✅ 获取用户信息成功${NC}"
        return 0
    else
        echo -e "${RED}❌ 获取用户信息失败${NC}"
        echo "响应: $response"
        return 1
    fi
}

# 测试创建会议
test_create_meeting() {
    echo -e "${BLUE}🔍 测试创建会议...${NC}"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        echo -e "${RED}❌ 没有访问令牌${NC}"
        return 1
    fi
    
    # 计算未来时间
    start_time=$(date -d "+1 hour" -Iseconds)
    end_time=$(date -d "+2 hours" -Iseconds)
    
    response=$(curl -s -X POST "$MEETING_SERVICE_URL/api/v1/meetings" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"title\": \"测试会议\",
            \"description\": \"这是一个测试会议\",
            \"start_time\": \"$start_time\",
            \"end_time\": \"$end_time\",
            \"max_participants\": 10,
            \"meeting_type\": \"video\"
        }")
    
    if echo "$response" | grep -q '"message":"Meeting created successfully"'; then
        echo -e "${GREEN}✅ 创建会议成功${NC}"
        # 提取会议ID
        MEETING_ID=$(echo "$response" | grep -o '"id":[0-9]*' | cut -d':' -f2)
        echo "会议ID: $MEETING_ID"
        return 0
    else
        echo -e "${RED}❌ 创建会议失败${NC}"
        echo "响应: $response"
        return 1
    fi
}

# 测试获取会议信息
test_get_meeting() {
    echo -e "${BLUE}🔍 测试获取会议信息...${NC}"
    
    if [ -z "$ACCESS_TOKEN" ] || [ -z "$MEETING_ID" ]; then
        echo -e "${RED}❌ 缺少访问令牌或会议ID${NC}"
        return 1
    fi
    
    response=$(curl -s -X GET "$MEETING_SERVICE_URL/api/v1/meetings/$MEETING_ID" \
        -H "Authorization: Bearer $ACCESS_TOKEN")
    
    if echo "$response" | grep -q '"title"'; then
        echo -e "${GREEN}✅ 获取会议信息成功${NC}"
        return 0
    else
        echo -e "${RED}❌ 获取会议信息失败${NC}"
        echo "响应: $response"
        return 1
    fi
}

# 测试加入会议
test_join_meeting() {
    echo -e "${BLUE}🔍 测试加入会议...${NC}"
    
    if [ -z "$ACCESS_TOKEN" ] || [ -z "$MEETING_ID" ]; then
        echo -e "${RED}❌ 缺少访问令牌或会议ID${NC}"
        return 1
    fi
    
    response=$(curl -s -X POST "$MEETING_SERVICE_URL/api/v1/meetings/$MEETING_ID/join" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{}")
    
    if echo "$response" | grep -q '"message":"Joined meeting successfully"'; then
        echo -e "${GREEN}✅ 加入会议成功${NC}"
        return 0
    else
        echo -e "${RED}❌ 加入会议失败${NC}"
        echo "响应: $response"
        return 1
    fi
}

# 主测试流程
main() {
    echo -e "${BLUE}🎯 开始微服务集成测试${NC}"
    echo "=================================="
    
    # 等待服务启动
    wait_for_service "$USER_SERVICE_URL" "用户服务" || exit 1
    wait_for_service "$MEETING_SERVICE_URL" "会议服务" || exit 1
    
    echo ""
    echo -e "${BLUE}📋 开始功能测试${NC}"
    echo "=================================="
    
    # 测试健康检查
    test_health_check "$USER_SERVICE_URL" "用户服务" || exit 1
    test_health_check "$MEETING_SERVICE_URL" "会议服务" || exit 1
    
    # 测试用户功能
    test_user_registration
    test_user_login || exit 1
    test_get_profile || exit 1
    
    # 测试会议功能
    test_create_meeting || exit 1
    test_get_meeting || exit 1
    test_join_meeting || exit 1
    
    echo ""
    echo -e "${GREEN}🎉 所有测试通过！${NC}"
    echo "=================================="
    echo -e "${GREEN}✅ 用户服务运行正常${NC}"
    echo -e "${GREEN}✅ 会议服务运行正常${NC}"
    echo -e "${GREEN}✅ 服务间集成正常${NC}"
}

# 运行主函数
main "$@"
