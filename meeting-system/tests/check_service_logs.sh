#!/bin/bash

# 服务日志检查脚本
# 用于验证消息队列系统在各个服务中的运行状态

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 服务列表
SERVICES=("user-service" "meeting-service" "media-service" "signaling-service" "ai-service")

# 日志目录
LOG_DIR="../backend"

# 输出文件
OUTPUT_FILE="service_logs_check_$(date +%Y%m%d_%H%M%S).md"

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 检查服务日志
check_service_log() {
    local service=$1
    local log_file="${LOG_DIR}/${service}/logs/service.log"
    
    log "检查 ${service} 日志..."
    
    if [ ! -f "$log_file" ]; then
        log_warning "日志文件不存在: $log_file"
        echo "## ${service}" >> "$OUTPUT_FILE"
        echo "❌ 日志文件不存在" >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
        return
    fi
    
    echo "## ${service}" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    # 检查队列系统初始化
    if grep -q "Initializing message queue system" "$log_file" 2>/dev/null; then
        log_success "  ✅ 队列系统初始化"
        echo "✅ 队列系统初始化成功" >> "$OUTPUT_FILE"
    else
        log_warning "  ❌ 未找到队列系统初始化日志"
        echo "❌ 未找到队列系统初始化日志" >> "$OUTPUT_FILE"
    fi
    
    # 检查任务处理器注册
    if grep -q "Registering.*task handlers" "$log_file" 2>/dev/null; then
        log_success "  ✅ 任务处理器注册"
        echo "✅ 任务处理器注册成功" >> "$OUTPUT_FILE"
    else
        log_warning "  ❌ 未找到任务处理器注册日志"
        echo "❌ 未找到任务处理器注册日志" >> "$OUTPUT_FILE"
    fi
    
    # 检查 Redis 消息队列
    if grep -q "Redis message queue handlers registered" "$log_file" 2>/dev/null; then
        log_success "  ✅ Redis 消息队列处理器"
        echo "✅ Redis 消息队列处理器注册成功" >> "$OUTPUT_FILE"
    else
        log_warning "  ❌ 未找到 Redis 消息队列处理器日志"
        echo "❌ 未找到 Redis 消息队列处理器日志" >> "$OUTPUT_FILE"
    fi
    
    # 检查 PubSub 订阅
    if grep -q "PubSub handlers registered" "$log_file" 2>/dev/null; then
        log_success "  ✅ PubSub 处理器"
        echo "✅ PubSub 处理器注册成功" >> "$OUTPUT_FILE"
    else
        log_warning "  ❌ 未找到 PubSub 处理器日志"
        echo "❌ 未找到 PubSub 处理器日志" >> "$OUTPUT_FILE"
    fi
    
    # 检查本地事件总线
    if grep -q "Local event bus handlers registered" "$log_file" 2>/dev/null; then
        log_success "  ✅ 本地事件总线"
        echo "✅ 本地事件总线处理器注册成功" >> "$OUTPUT_FILE"
    else
        log_warning "  ❌ 未找到本地事件总线日志"
        echo "❌ 未找到本地事件总线日志" >> "$OUTPUT_FILE"
    fi
    
    # 检查任务处理
    local task_count=$(grep -c "Processing.*task" "$log_file" 2>/dev/null || echo "0")
    log "  📊 处理任务数: $task_count"
    echo "📊 处理任务数: $task_count" >> "$OUTPUT_FILE"
    
    # 检查事件接收
    local event_count=$(grep -c "Received.*event" "$log_file" 2>/dev/null || echo "0")
    log "  📊 接收事件数: $event_count"
    echo "📊 接收事件数: $event_count" >> "$OUTPUT_FILE"
    
    # 检查错误
    local error_count=$(grep -c "ERROR\|Failed" "$log_file" 2>/dev/null || echo "0")
    if [ "$error_count" -gt 0 ]; then
        log_warning "  ⚠️  错误数: $error_count"
        echo "⚠️ 错误数: $error_count" >> "$OUTPUT_FILE"
        
        # 显示最近的错误
        echo "" >> "$OUTPUT_FILE"
        echo "### 最近的错误" >> "$OUTPUT_FILE"
        echo "\`\`\`" >> "$OUTPUT_FILE"
        grep "ERROR\|Failed" "$log_file" 2>/dev/null | tail -5 >> "$OUTPUT_FILE" || true
        echo "\`\`\`" >> "$OUTPUT_FILE"
    else
        log_success "  ✅ 无错误"
        echo "✅ 无错误" >> "$OUTPUT_FILE"
    fi
    
    echo "" >> "$OUTPUT_FILE"
}

# 检查 Docker 容器日志
check_docker_logs() {
    log "检查 Docker 容器日志..."
    
    echo "# Docker 容器日志检查" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    for service in "${SERVICES[@]}"; do
        local container_name="${service}"
        
        if docker ps --format '{{.Names}}' | grep -q "$container_name" 2>/dev/null; then
            log "  检查容器: $container_name"
            echo "## ${container_name}" >> "$OUTPUT_FILE"
            
            # 检查队列相关日志
            local queue_logs=$(docker logs "$container_name" 2>&1 | grep -i "queue\|task\|event" | tail -10 || echo "无队列相关日志")
            
            echo "\`\`\`" >> "$OUTPUT_FILE"
            echo "$queue_logs" >> "$OUTPUT_FILE"
            echo "\`\`\`" >> "$OUTPUT_FILE"
            echo "" >> "$OUTPUT_FILE"
        else
            log_warning "  容器未运行: $container_name"
        fi
    done
}

# 检查 Redis 统计
check_redis_stats() {
    log "检查 Redis 统计..."
    
    echo "# Redis 队列统计" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    # 检查队列长度
    echo "## 队列长度" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    echo "| 队列名称 | 长度 |" >> "$OUTPUT_FILE"
    echo "|---------|------|" >> "$OUTPUT_FILE"
    
    for queue in "critical_queue" "high_queue" "normal_queue" "low_queue" "dead_letter_queue"; do
        local length=$(redis-cli LLEN "meeting_system:${queue}" 2>/dev/null || echo "N/A")
        echo "| ${queue} | ${length} |" >> "$OUTPUT_FILE"
        log "  ${queue}: ${length}"
    done
    
    echo "" >> "$OUTPUT_FILE"
    
    # 检查 Pub/Sub 频道
    echo "## Pub/Sub 频道" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    local channels=$(redis-cli PUBSUB CHANNELS "meeting_system:*" 2>/dev/null || echo "无")
    echo "\`\`\`" >> "$OUTPUT_FILE"
    echo "$channels" >> "$OUTPUT_FILE"
    echo "\`\`\`" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
}

# 生成总结
generate_summary() {
    log "生成总结..."
    
    # 在文件开头插入总结
    local temp_file=$(mktemp)
    
    cat > "$temp_file" << EOF
# 服务日志检查报告

**检查时间**: $(date +'%Y-%m-%d %H:%M:%S')

## 总结

EOF
    
    # 统计各服务状态
    for service in "${SERVICES[@]}"; do
        local status="✅"
        local log_file="${LOG_DIR}/${service}/logs/service.log"
        
        if [ ! -f "$log_file" ]; then
            status="❌"
        elif ! grep -q "Registering.*task handlers" "$log_file" 2>/dev/null; then
            status="⚠️"
        fi
        
        echo "- ${status} ${service}" >> "$temp_file"
    done
    
    echo "" >> "$temp_file"
    echo "---" >> "$temp_file"
    echo "" >> "$temp_file"
    
    # 合并原有内容
    cat "$OUTPUT_FILE" >> "$temp_file"
    mv "$temp_file" "$OUTPUT_FILE"
}

# 主函数
main() {
    log "========================================="
    log "开始检查服务日志"
    log "========================================="
    
    # 初始化输出文件
    echo "" > "$OUTPUT_FILE"
    
    # 检查各服务日志
    echo "# 服务日志详细检查" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    
    for service in "${SERVICES[@]}"; do
        check_service_log "$service"
    done
    
    # 检查 Redis 统计
    check_redis_stats
    
    # 检查 Docker 日志（如果使用 Docker）
    if command -v docker &> /dev/null; then
        check_docker_logs
    fi
    
    # 生成总结
    generate_summary
    
    log "========================================="
    log "检查完成！"
    log "报告文件: $OUTPUT_FILE"
    log "========================================="
    
    # 显示报告
    cat "$OUTPUT_FILE"
}

# 运行
main

