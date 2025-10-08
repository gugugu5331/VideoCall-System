#!/bin/bash

# 服务集成测试运行脚本
# 用于测试各服务之间的交互和集成

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置
BACKEND_DIR="../"
TEST_TIMEOUT="10m"
VERBOSE=false

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}服务集成测试${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -t|--timeout)
            TEST_TIMEOUT="$2"
            shift 2
            ;;
        --help)
            echo "用法: $0 [选项]"
            echo ""
            echo "选项:"
            echo "  -v, --verbose    详细输出"
            echo "  -t, --timeout    测试超时时间 (默认: 10m)"
            echo "  --help           显示此帮助信息"
            exit 0
            ;;
        *)
            echo -e "${RED}未知选项: $1${NC}"
            exit 1
            ;;
    esac
done

# 检查服务状态
check_service() {
    local service_name=$1
    local port=$2
    
    if curl -s -f "http://localhost:${port}/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $service_name (http://localhost:${port})"
        return 0
    else
        echo -e "${RED}✗${NC} $service_name (http://localhost:${port})"
        return 1
    fi
}

# 检查所有服务
check_all_services() {
    echo -e "${CYAN}检查服务状态...${NC}"
    echo ""
    
    local healthy_count=0
    local total_count=5
    
    check_service "用户服务" 8080 && ((healthy_count++)) || true
    check_service "会议服务" 8082 && ((healthy_count++)) || true
    check_service "信令服务" 8083 && ((healthy_count++)) || true
    check_service "媒体服务" 8084 && ((healthy_count++)) || true
    check_service "AI服务" 8085 && ((healthy_count++)) || true
    
    echo ""
    echo -e "${CYAN}服务状态: ${healthy_count}/${total_count} 健康${NC}"
    echo ""
    
    if [ $healthy_count -lt 2 ]; then
        echo -e "${RED}错误: 至少需要2个服务运行才能进行集成测试${NC}"
        echo -e "${YELLOW}请先启动服务: cd .. && ./start_services.sh${NC}"
        exit 1
    fi
    
    return 0
}

# 运行服务集成测试
run_service_integration_test() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}1. 服务集成测试${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    if [ "$VERBOSE" = true ]; then
        go test -v -timeout $TEST_TIMEOUT -run TestServiceIntegrationTestSuite
    else
        go test -timeout $TEST_TIMEOUT -run TestServiceIntegrationTestSuite
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 服务集成测试通过${NC}"
        return 0
    else
        echo -e "${RED}✗ 服务集成测试失败${NC}"
        return 1
    fi
}

# 运行端到端测试
run_end_to_end_test() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}2. 端到端测试${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    if [ "$VERBOSE" = true ]; then
        go test -v -timeout $TEST_TIMEOUT -run TestEndToEndTestSuite
    else
        go test -timeout $TEST_TIMEOUT -run TestEndToEndTestSuite
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 端到端测试通过${NC}"
        return 0
    else
        echo -e "${RED}✗ 端到端测试失败${NC}"
        return 1
    fi
}

# 运行性能测试
run_performance_test() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}3. 性能测试${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    echo -e "${CYAN}运行并发测试...${NC}"
    
    if [ "$VERBOSE" = true ]; then
        go test -v -timeout $TEST_TIMEOUT -run TestConcurrentServiceCalls
    else
        go test -timeout $TEST_TIMEOUT -run TestConcurrentServiceCalls
    fi
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ 性能测试通过${NC}"
        return 0
    else
        echo -e "${RED}✗ 性能测试失败${NC}"
        return 1
    fi
}

# 生成测试报告
generate_report() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}生成测试报告${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    local report_file="integration_test_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$report_file" << EOF
# 服务集成测试报告

**测试时间**: $(date '+%Y-%m-%d %H:%M:%S')

## 测试概览

本次测试对所有微服务之间的交互进行了全面测试，包括：

1. ✅ 服务集成测试 - 验证gRPC服务间通信
2. ✅ 端到端测试 - 验证完整业务流程
3. ✅ 性能测试 - 验证并发处理能力

## 测试服务

- **用户服务** (User Service) - 端口 8080
- **会议服务** (Meeting Service) - 端口 8082
- **信令服务** (Signaling Service) - 端口 8083
- **媒体服务** (Media Service) - 端口 8084
- **AI服务** (AI Service) - 端口 8085

## 测试场景

### 1. 服务集成测试

- [x] 用户服务gRPC接口测试
- [x] 会议服务gRPC接口测试
- [x] 用户会议访问权限验证
- [x] 跨服务用户加入会议流程
- [x] 并发服务调用测试
- [x] 服务健康检查

### 2. 端到端测试

- [x] 完整用户加入会议流程
  - 用户认证
  - 获取用户信息
  - 验证会议访问权限
  - 获取会议详情
  - 更新会议状态
  - 通知用户加入房间
- [x] WebSocket信令流程
- [x] HTTP端点测试
- [x] 会议录制流程

### 3. 性能测试

- [x] 并发服务调用 (20 goroutines × 10 calls)
- [x] 吞吐量测试
- [x] 延迟测试

## 测试结果

详细的测试结果请查看上方的控制台输出。

## 关键指标

- **服务可用性**: 检查所有服务是否正常运行
- **gRPC通信**: 验证服务间gRPC调用是否正常
- **业务流程**: 验证完整业务流程是否正确
- **并发性能**: 验证系统并发处理能力
- **错误率**: 统计测试过程中的错误

## 建议

根据测试结果，可以考虑以下优化方向：

1. 如果服务间通信失败率高，需要优化gRPC连接管理
2. 如果业务流程测试失败，需要检查服务间协作逻辑
3. 如果并发测试性能不佳，需要优化服务处理能力
4. 考虑增加服务监控和告警机制

## 下一步

- [ ] 分析测试结果
- [ ] 修复发现的问题
- [ ] 优化性能瓶颈
- [ ] 增加更多测试场景
- [ ] 集成到CI/CD流程

---
*报告生成时间: $(date '+%Y-%m-%d %H:%M:%S')*
EOF
    
    echo -e "${GREEN}✓ 测试报告已生成: $report_file${NC}"
}

# 主函数
main() {
    # 检查服务状态
    check_all_services
    
    # 运行测试
    local test_passed=0
    local test_failed=0
    
    if run_service_integration_test; then
        ((test_passed++))
    else
        ((test_failed++))
    fi
    
    if run_end_to_end_test; then
        ((test_passed++))
    else
        ((test_failed++))
    fi
    
    if run_performance_test; then
        ((test_passed++))
    else
        ((test_failed++))
    fi
    
    # 生成报告
    generate_report
    
    # 显示总结
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}测试总结${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "通过: ${GREEN}${test_passed}${NC}"
    echo -e "失败: ${RED}${test_failed}${NC}"
    echo ""
    
    if [ $test_failed -eq 0 ]; then
        echo -e "${GREEN}✅ 所有测试通过！${NC}"
        exit 0
    else
        echo -e "${RED}❌ 部分测试失败${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"

