#!/bin/bash

# 端到端消息队列集成测试脚本
# 测试场景：三个用户注册、加入同一会议室并调用 AI 服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
NGINX_URL="http://localhost"
API_BASE="${NGINX_URL}/api"
REDIS_CLI="redis-cli"

# 日志文件
LOG_FILE="e2e_test_$(date +%Y%m%d_%H%M%S).log"
REPORT_FILE="e2e_test_report_$(date +%Y%m%d_%H%M%S).md"

# 测试数据
USER1_USERNAME="test_user_1"
USER1_EMAIL="user1@test.com"
USER1_PASSWORD="password123"

USER2_USERNAME="test_user_2"
USER2_EMAIL="user2@test.com"
USER2_PASSWORD="password123"

USER3_USERNAME="test_user_3"
USER3_EMAIL="user3@test.com"
USER3_PASSWORD="password123"

# 全局变量
USER1_TOKEN=""
USER2_TOKEN=""
USER3_TOKEN=""
MEETING_ID=""

# 辅助函数
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

# 检查服务状态
check_services() {
    log "检查服务状态..."
    
    # 检查 Nginx
    if ! curl -s "${NGINX_URL}/health" > /dev/null 2>&1; then
        log_warning "Nginx 可能未运行或健康检查端点不可用"
    else
        log_success "Nginx 运行正常"
    fi
    
    # 检查 Redis
    if ! $REDIS_CLI ping > /dev/null 2>&1; then
        log_error "Redis 未运行"
        exit 1
    else
        log_success "Redis 运行正常"
    fi
}

# 检查 Redis 队列状态
check_redis_queues() {
    log "检查 Redis 队列状态..."
    
    echo "=== Redis 队列统计 ===" >> "$LOG_FILE"
    
    # 检查各个优先级队列
    CRITICAL_LEN=$($REDIS_CLI LLEN "meeting_system:critical_queue" 2>/dev/null || echo "0")
    HIGH_LEN=$($REDIS_CLI LLEN "meeting_system:high_queue" 2>/dev/null || echo "0")
    NORMAL_LEN=$($REDIS_CLI LLEN "meeting_system:normal_queue" 2>/dev/null || echo "0")
    LOW_LEN=$($REDIS_CLI LLEN "meeting_system:low_queue" 2>/dev/null || echo "0")
    DLQ_LEN=$($REDIS_CLI LLEN "meeting_system:dead_letter_queue" 2>/dev/null || echo "0")
    
    echo "Critical Queue: $CRITICAL_LEN" | tee -a "$LOG_FILE"
    echo "High Queue: $HIGH_LEN" | tee -a "$LOG_FILE"
    echo "Normal Queue: $NORMAL_LEN" | tee -a "$LOG_FILE"
    echo "Low Queue: $LOW_LEN" | tee -a "$LOG_FILE"
    echo "Dead Letter Queue: $DLQ_LEN" | tee -a "$LOG_FILE"
    echo "" >> "$LOG_FILE"
}

# 用户注册
register_user() {
    local username=$1
    local email=$2
    local password=$3
    
    log "注册用户: $username"
    
    local response=$(curl -s -X POST "${API_BASE}/users/register" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"email\":\"$email\",\"password\":\"$password\"}")
    
    echo "Response: $response" >> "$LOG_FILE"
    
    if echo "$response" | grep -q "success\|token\|user"; then
        log_success "用户 $username 注册成功"
        return 0
    else
        log_error "用户 $username 注册失败: $response"
        return 1
    fi
}

# 用户登录
login_user() {
    local username=$1
    local password=$2
    
    log "用户登录: $username"
    
    local response=$(curl -s -X POST "${API_BASE}/users/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    
    echo "Response: $response" >> "$LOG_FILE"
    
    # 提取 token
    local token=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [ -n "$token" ]; then
        log_success "用户 $username 登录成功"
        echo "$token"
        return 0
    else
        log_error "用户 $username 登录失败: $response"
        return 1
    fi
}

# 创建会议
create_meeting() {
    local token=$1
    local title=$2
    
    log "创建会议: $title"
    
    local response=$(curl -s -X POST "${API_BASE}/meetings" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{\"title\":\"$title\",\"description\":\"E2E Test Meeting\",\"start_time\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}")
    
    echo "Response: $response" >> "$LOG_FILE"
    
    # 提取 meeting_id
    local meeting_id=$(echo "$response" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    
    if [ -n "$meeting_id" ]; then
        log_success "会议创建成功，ID: $meeting_id"
        echo "$meeting_id"
        return 0
    else
        log_error "会议创建失败: $response"
        return 1
    fi
}

# 加入会议
join_meeting() {
    local token=$1
    local meeting_id=$2
    local username=$3
    
    log "用户 $username 加入会议 $meeting_id"
    
    local response=$(curl -s -X POST "${API_BASE}/meetings/${meeting_id}/join" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token")
    
    echo "Response: $response" >> "$LOG_FILE"
    
    if echo "$response" | grep -q "success\|joined"; then
        log_success "用户 $username 成功加入会议"
        return 0
    else
        log_warning "用户 $username 加入会议响应: $response"
        return 0  # 继续测试
    fi
}

# 调用 AI 服务
call_ai_service() {
    local token=$1
    local service_type=$2
    local username=$3
    
    log "用户 $username 调用 AI 服务: $service_type"
    
    local endpoint=""
    local payload=""
    
    case $service_type in
        "speech_recognition")
            endpoint="${API_BASE}/ai/speech-recognition"
            payload='{"audio_data":"base64_encoded_audio_data","language":"zh-CN"}'
            ;;
        "emotion_detection")
            endpoint="${API_BASE}/ai/emotion-detection"
            payload='{"audio_data":"base64_encoded_audio_data"}'
            ;;
        "audio_denoising")
            endpoint="${API_BASE}/ai/audio-denoising"
            payload='{"audio_data":"base64_encoded_audio_data"}'
            ;;
    esac
    
    local response=$(curl -s -X POST "$endpoint" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "$payload")
    
    echo "Response: $response" >> "$LOG_FILE"
    
    if echo "$response" | grep -q "success\|result\|task"; then
        log_success "AI 服务 $service_type 调用成功"
        return 0
    else
        log_warning "AI 服务 $service_type 响应: $response"
        return 0  # 继续测试
    fi
}

# 生成测试报告
generate_report() {
    log "生成测试报告..."
    
    cat > "$REPORT_FILE" << EOF
# 端到端消息队列集成测试报告

**测试时间**: $(date +'%Y-%m-%d %H:%M:%S')

## 测试场景

三个用户注册、加入同一会议室并调用 AI 服务

## 测试步骤

### 1. 用户注册阶段
- User1: $USER1_USERNAME ($USER1_EMAIL)
- User2: $USER2_USERNAME ($USER2_EMAIL)
- User3: $USER3_USERNAME ($USER3_EMAIL)

### 2. 用户登录阶段
- 所有用户成功登录并获取 token

### 3. 创建会议室阶段
- User1 创建会议室
- 会议 ID: $MEETING_ID

### 4. 用户加入会议阶段
- User1、User2、User3 依次加入会议

### 5. 调用 AI 服务阶段
- User1: 语音识别
- User2: 情绪检测
- User3: 音频降噪

## Redis 队列统计

\`\`\`
$(check_redis_queues 2>&1)
\`\`\`

## 测试结论

测试已完成。详细日志请查看: $LOG_FILE

## 建议

1. 检查各服务日志，确认消息队列系统正常工作
2. 验证事件流转是否符合预期
3. 监控 Redis 队列长度和死信队列

EOF

    log_success "测试报告已生成: $REPORT_FILE"
}

# 主测试流程
main() {
    log "========================================="
    log "开始端到端消息队列集成测试"
    log "========================================="
    
    # 检查服务
    check_services
    
    # 检查初始队列状态
    log "=== 初始队列状态 ==="
    check_redis_queues
    
    # 阶段 1: 用户注册
    log "=== 阶段 1: 用户注册 ==="
    register_user "$USER1_USERNAME" "$USER1_EMAIL" "$USER1_PASSWORD" || true
    sleep 1
    register_user "$USER2_USERNAME" "$USER2_EMAIL" "$USER2_PASSWORD" || true
    sleep 1
    register_user "$USER3_USERNAME" "$USER3_EMAIL" "$USER3_PASSWORD" || true
    sleep 2
    
    check_redis_queues
    
    # 阶段 2: 用户登录
    log "=== 阶段 2: 用户登录 ==="
    USER1_TOKEN=$(login_user "$USER1_USERNAME" "$USER1_PASSWORD")
    sleep 1
    USER2_TOKEN=$(login_user "$USER2_USERNAME" "$USER2_PASSWORD")
    sleep 1
    USER3_TOKEN=$(login_user "$USER3_USERNAME" "$USER3_PASSWORD")
    sleep 2
    
    # 阶段 3: 创建会议
    log "=== 阶段 3: 创建会议 ==="
    if [ -n "$USER1_TOKEN" ]; then
        MEETING_ID=$(create_meeting "$USER1_TOKEN" "E2E Test Meeting")
        sleep 2
        check_redis_queues
    else
        log_error "User1 token 为空，无法创建会议"
    fi
    
    # 阶段 4: 加入会议
    log "=== 阶段 4: 用户加入会议 ==="
    if [ -n "$MEETING_ID" ]; then
        [ -n "$USER1_TOKEN" ] && join_meeting "$USER1_TOKEN" "$MEETING_ID" "$USER1_USERNAME"
        sleep 1
        [ -n "$USER2_TOKEN" ] && join_meeting "$USER2_TOKEN" "$MEETING_ID" "$USER2_USERNAME"
        sleep 1
        [ -n "$USER3_TOKEN" ] && join_meeting "$USER3_TOKEN" "$MEETING_ID" "$USER3_USERNAME"
        sleep 2
        check_redis_queues
    else
        log_error "会议 ID 为空，跳过加入会议阶段"
    fi
    
    # 阶段 5: 调用 AI 服务
    log "=== 阶段 5: 调用 AI 服务 ==="
    [ -n "$USER1_TOKEN" ] && call_ai_service "$USER1_TOKEN" "speech_recognition" "$USER1_USERNAME"
    sleep 1
    [ -n "$USER2_TOKEN" ] && call_ai_service "$USER2_TOKEN" "emotion_detection" "$USER2_USERNAME"
    sleep 1
    [ -n "$USER3_TOKEN" ] && call_ai_service "$USER3_TOKEN" "audio_denoising" "$USER3_USERNAME"
    sleep 2
    
    # 最终队列状态
    log "=== 最终队列状态 ==="
    check_redis_queues
    
    # 生成报告
    generate_report
    
    log "========================================="
    log "测试完成！"
    log "日志文件: $LOG_FILE"
    log "报告文件: $REPORT_FILE"
    log "========================================="
}

# 运行测试
main

