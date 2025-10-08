#!/bin/bash

# AI Inference Service - 综合压力测试脚本 (Bash 版本)
#
# 测试场景：
# - 场景 A: 交替调用不同 AI 服务
# - 场景 B: 连续调用相同 AI 服务
# - 场景 C: 混合模式

set -e

# 配置
NGINX_BASE_URL="http://localhost:8800/api/v1/ai"
DIRECT_BASE_URL="http://localhost:8085/api/v1/ai"
TEST_VIDEO_DIR="/root/meeting-system-server/meeting-system/backend/media-service/test_video"
REQUEST_DELAY=0.2
TIMEOUT=30

# 颜色
GREEN='\033[92m'
RED='\033[91m'
YELLOW='\033[93m'
BLUE='\033[94m'
CYAN='\033[96m'
BOLD='\033[1m'
RESET='\033[0m'

# 统计变量
TOTAL_REQUESTS=0
SUCCESSFUL_REQUESTS=0
FAILED_REQUESTS=0
declare -a RESPONSE_TIMES=()
START_TIME=$(date +%s)

# 按服务统计
declare -A SERVICE_TOTAL
declare -A SERVICE_SUCCESS
declare -A SERVICE_FAILED

# 函数：打印标题
print_header() {
    echo -e "\n${BOLD}${CYAN}================================================================================${RESET}"
    echo -e "${BOLD}${CYAN}$1${RESET}"
    echo -e "${BOLD}${CYAN}================================================================================${RESET}\n"
}

# 函数：打印成功信息
print_success() {
    echo -e "${GREEN}✓ $1${RESET}"
}

# 函数：打印错误信息
print_error() {
    echo -e "${RED}✗ $1${RESET}"
}

# 函数：打印信息
print_info() {
    echo -e "${BLUE}ℹ $1${RESET}"
}

# 函数：打印警告
print_warning() {
    echo -e "${YELLOW}⚠ $1${RESET}"
}

# 函数：调用 AI 服务
call_ai_service() {
    local service=$1
    local data=$2
    local url="${NGINX_BASE_URL}/${service}"
    
    local start_time=$(date +%s.%N)
    
    local response=$(curl -s -w "\n%{http_code}" -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$data" \
        --max-time $TIMEOUT 2>&1)
    
    local end_time=$(date +%s.%N)
    local response_time=$(echo "$end_time - $start_time" | bc)
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    TOTAL_REQUESTS=$((TOTAL_REQUESTS + 1))
    SERVICE_TOTAL[$service]=$((${SERVICE_TOTAL[$service]:-0} + 1))
    
    if [ "$http_code" = "200" ]; then
        local code=$(echo "$body" | jq -r '.code' 2>/dev/null || echo "0")
        if [ "$code" = "200" ]; then
            SUCCESSFUL_REQUESTS=$((SUCCESSFUL_REQUESTS + 1))
            SERVICE_SUCCESS[$service]=$((${SERVICE_SUCCESS[$service]:-0} + 1))
            RESPONSE_TIMES+=($response_time)
            echo "success|$response_time"
            return 0
        fi
    fi
    
    FAILED_REQUESTS=$((FAILED_REQUESTS + 1))
    SERVICE_FAILED[$service]=$((${SERVICE_FAILED[$service]:-0} + 1))
    local error_msg=$(echo "$body" | jq -r '.message' 2>/dev/null || echo "Unknown error")
    echo "failed|$response_time|$error_msg"
    return 1
}

# 函数：准备测试数据
prepare_test_data() {
    print_info "Preparing test data..."
    
    # 查找音频文件
    local audio_file=$(find "$TEST_VIDEO_DIR" -name "*.mp3" | head -n1)
    
    if [ -z "$audio_file" ]; then
        print_warning "No MP3 files found, using sample base64 data"
        AUDIO_BASE64="c2FtcGxlIGF1ZGlvIGRhdGE="
    else:
        print_info "Using audio file: $(basename $audio_file)"
        AUDIO_BASE64=$(base64 -w 0 "$audio_file" 2>/dev/null || base64 "$audio_file")
    fi
    
    # ASR 测试数据
    ASR_DATA="{\"audio_data\":\"$AUDIO_BASE64\",\"format\":\"mp3\",\"sample_rate\":16000}"
    
    # Emotion 测试数据
    EMOTION_DATA='{"text":"I am very happy today! This is a wonderful day and everything is going great."}'
    
    # Synthesis 测试数据
    SYNTHESIS_DATA="{\"audio_data\":\"$AUDIO_BASE64\",\"format\":\"mp3\",\"sample_rate\":16000}"
    
    print_success "Test data prepared"
}

# 场景 A: 交替调用不同 AI 服务
scenario_a_alternating_calls() {
    local rounds=${1:-5}
    
    print_header "Scenario A: Alternating Calls to Different AI Services"
    print_info "Calling each service $rounds times in rotation..."
    
    local services=("asr" "emotion" "synthesis")
    local total_calls=$((rounds * 3))
    local current_call=0
    
    for ((round=1; round<=rounds; round++)); do
        echo -e "\n${BOLD}Round $round/$rounds${RESET}"
        
        for service in "${services[@]}"; do
            current_call=$((current_call + 1))
            echo -n "  [$current_call/$total_calls] Calling $service... "
            
            local data_var="${service^^}_DATA"
            local result=$(call_ai_service "$service" "${!data_var}")
            
            local status=$(echo "$result" | cut -d'|' -f1)
            local response_time=$(echo "$result" | cut -d'|' -f2)
            
            if [ "$status" = "success" ]; then
                print_success "OK (${response_time}s)"
            else
                local error=$(echo "$result" | cut -d'|' -f3)
                print_error "FAILED ($error)"
            fi
            
            sleep $REQUEST_DELAY
        done
    done
    
    print_success "\nScenario A completed: $total_calls requests"
}

# 场景 B: 连续调用相同 AI 服务
scenario_b_consecutive_calls() {
    local calls_per_service=${1:-10}
    
    print_header "Scenario B: Consecutive Calls to Same AI Service"
    print_info "Calling each service $calls_per_service times consecutively..."
    
    local services=("asr" "emotion" "synthesis")
    
    for service in "${services[@]}"; do
        echo -e "\n${BOLD}Testing ${service^^} ($calls_per_service consecutive calls)${RESET}"
        
        local data_var="${service^^}_DATA"
        
        for ((call=1; call<=calls_per_service; call++)); do
            echo -n "  [$call/$calls_per_service] Calling $service... "
            
            local result=$(call_ai_service "$service" "${!data_var}")
            
            local status=$(echo "$result" | cut -d'|' -f1)
            local response_time=$(echo "$result" | cut -d'|' -f2)
            
            if [ "$status" = "success" ]; then
                print_success "OK (${response_time}s)"
            else
                local error=$(echo "$result" | cut -d'|' -f3)
                print_error "FAILED ($error)"
            fi
            
            sleep $REQUEST_DELAY
        done
    done
    
    local total_calls=$((calls_per_service * 3))
    print_success "\nScenario B completed: $total_calls requests"
}

# 场景 C: 混合模式
scenario_c_mixed_mode() {
    local rounds=${1:-3}
    
    print_header "Scenario C: Mixed Mode (Alternating Scenarios A and B)"
    print_info "Executing $rounds rounds of mixed scenarios..."
    
    for ((round=1; round<=rounds; round++)); do
        echo -e "\n${BOLD}Mixed Round $round/$rounds${RESET}"
        
        echo -e "\n${CYAN}  → Running Scenario A (simplified)${RESET}"
        scenario_a_alternating_calls 2
        
        sleep 0.5
        
        echo -e "\n${CYAN}  → Running Scenario B (simplified)${RESET}"
        scenario_b_consecutive_calls 3
        
        sleep 0.5
    done
    
    print_success "\nScenario C completed: $rounds mixed rounds"
}

# 函数：计算统计信息
calculate_statistics() {
    print_header "Test Statistics Summary"
    
    local end_time=$(date +%s)
    local duration=$((end_time - START_TIME))
    
    # 总体统计
    echo -e "${BOLD}Overall Statistics:${RESET}"
    echo "  Total Requests:      $TOTAL_REQUESTS"
    echo -e "  Successful Requests: ${GREEN}$SUCCESSFUL_REQUESTS${RESET}"
    echo -e "  Failed Requests:     ${RED}$FAILED_REQUESTS${RESET}"
    
    local success_rate=0
    if [ $TOTAL_REQUESTS -gt 0 ]; then
        success_rate=$(echo "scale=2; $SUCCESSFUL_REQUESTS * 100 / $TOTAL_REQUESTS" | bc)
    fi
    
    if (( $(echo "$success_rate >= 95" | bc -l) )); then
        echo -e "  Success Rate:        ${GREEN}${success_rate}%${RESET}"
    else
        echo -e "  Success Rate:        ${RED}${success_rate}%${RESET}"
    fi
    
    local error_rate=$(echo "scale=2; $FAILED_REQUESTS * 100 / $TOTAL_REQUESTS" | bc)
    echo "  Error Rate:          ${error_rate}%"
    echo "  Duration:            ${duration}s"
    
    local qps=0
    if [ $duration -gt 0 ]; then
        qps=$(echo "scale=2; $TOTAL_REQUESTS / $duration" | bc)
    fi
    echo "  Throughput (QPS):    $qps"
    
    # 响应时间统计
    if [ ${#RESPONSE_TIMES[@]} -gt 0 ]; then
        echo -e "\n${BOLD}Response Time Statistics:${RESET}"
        
        # 排序响应时间
        IFS=$'\n' sorted_times=($(sort -n <<<"${RESPONSE_TIMES[*]}"))
        unset IFS
        
        local min_time=${sorted_times[0]}
        local max_time=${sorted_times[-1]}
        
        # 计算平均值
        local sum=0
        for time in "${RESPONSE_TIMES[@]}"; do
            sum=$(echo "$sum + $time" | bc)
        done
        local avg_time=$(echo "scale=3; $sum / ${#RESPONSE_TIMES[@]}" | bc)
        
        # 计算中位数
        local median_index=$((${#sorted_times[@]} / 2))
        local median_time=${sorted_times[$median_index]}
        
        # 计算 P95 和 P99
        local p95_index=$(echo "${#sorted_times[@]} * 0.95 / 1" | bc)
        local p99_index=$(echo "${#sorted_times[@]} * 0.99 / 1" | bc)
        local p95_time=${sorted_times[$p95_index]}
        local p99_time=${sorted_times[$p99_index]}
        
        echo "  Min:     ${min_time}s"
        echo "  Max:     ${max_time}s"
        
        if (( $(echo "$avg_time < 3" | bc -l) )); then
            echo -e "  Average: ${GREEN}${avg_time}s${RESET}"
        else
            echo -e "  Average: ${RED}${avg_time}s${RESET}"
        fi
        
        echo "  Median:  ${median_time}s"
        echo "  P95:     ${p95_time}s"
        echo "  P99:     ${p99_time}s"
    fi
    
    # 按服务统计
    echo -e "\n${BOLD}Statistics by Service:${RESET}"
    for service in "asr" "emotion" "synthesis"; do
        local total=${SERVICE_TOTAL[$service]:-0}
        local success=${SERVICE_SUCCESS[$service]:-0}
        local failed=${SERVICE_FAILED[$service]:-0}
        
        if [ $total -gt 0 ]; then
            local rate=$(echo "scale=2; $success * 100 / $total" | bc)
            
            echo "  ${service^^}:"
            echo "    Total:   $total"
            echo -e "    Success: ${GREEN}$success${RESET}"
            echo -e "    Failed:  ${RED}$failed${RESET}"
            
            if (( $(echo "$rate >= 95" | bc -l) )); then
                echo -e "    Rate:    ${GREEN}${rate}%${RESET}"
            else
                echo -e "    Rate:    ${RED}${rate}%${RESET}"
            fi
        fi
    done
    
    # 成功标准检查
    print_header "Success Criteria Check"
    
    local success_rate_ok=0
    local avg_time_ok=0
    local no_critical_errors=0
    
    if (( $(echo "$success_rate >= 95" | bc -l) )); then
        success_rate_ok=1
        echo -e "  Success Rate ≥ 95%:        ${GREEN}✓ PASS${RESET} (${success_rate}%)"
    else
        echo -e "  Success Rate ≥ 95%:        ${RED}✗ FAIL${RESET} (${success_rate}%)"
    fi
    
    if (( $(echo "$avg_time < 3" | bc -l) )); then
        avg_time_ok=1
        echo -e "  Avg Response Time < 3s:    ${GREEN}✓ PASS${RESET} (${avg_time}s)"
    else
        echo -e "  Avg Response Time < 3s:    ${RED}✗ FAIL${RESET} (${avg_time}s)"
    fi
    
    local error_threshold=$(echo "scale=0; $TOTAL_REQUESTS * 0.05" | bc)
    if [ $FAILED_REQUESTS -lt $error_threshold ]; then
        no_critical_errors=1
        echo -e "  No Critical Errors:        ${GREEN}✓ PASS${RESET}"
    else
        echo -e "  No Critical Errors:        ${RED}✗ FAIL${RESET}"
    fi
    
    if [ $success_rate_ok -eq 1 ] && [ $avg_time_ok -eq 1 ] && [ $no_critical_errors -eq 1 ]; then
        echo -e "\n${GREEN}${BOLD}✓ ALL TESTS PASSED${RESET}"
        return 0
    else
        echo -e "\n${RED}${BOLD}✗ SOME TESTS FAILED${RESET}"
        return 1
    fi
}

# 主函数
main() {
    print_header "AI Inference Service - Comprehensive Stress Test"
    echo "Start Time: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "Nginx Gateway: $NGINX_BASE_URL"
    echo "Request Delay: ${REQUEST_DELAY}s"
    echo "Timeout: ${TIMEOUT}s"
    
    # 准备测试数据
    prepare_test_data
    
    # 执行场景 A
    scenario_a_alternating_calls 5
    sleep 1
    
    # 执行场景 B
    scenario_b_consecutive_calls 10
    sleep 1
    
    # 执行场景 C
    scenario_c_mixed_mode 3
    
    # 计算并打印统计信息
    calculate_statistics
    
    return $?
}

# 执行主函数
main
exit $?

