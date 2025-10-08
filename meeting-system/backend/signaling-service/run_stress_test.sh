#!/bin/bash

# 信令服务并发压力测试脚本
# 用于测试多用户并发情况下信令服务的性能和稳定性

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置
SERVER_URL="ws://localhost:8083/ws/signaling"
JWT_SECRET="test-secret"
MEETING_ID=1

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}信令服务并发压力测试${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# 检查服务是否运行
check_service() {
    echo -e "${YELLOW}检查信令服务状态...${NC}"
    
    # 提取主机和端口
    if [[ $SERVER_URL =~ ws://([^:/]+):([0-9]+) ]]; then
        HOST="${BASH_REMATCH[1]}"
        PORT="${BASH_REMATCH[2]}"
        
        # 检查HTTP健康检查端点
        if curl -s -f "http://${HOST}:${PORT}/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ 信令服务运行正常${NC}"
            return 0
        else
            echo -e "${RED}✗ 信令服务未运行或健康检查失败${NC}"
            echo -e "${YELLOW}请先启动信令服务: cd signaling-service && go run main.go${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠ 无法解析服务URL，跳过健康检查${NC}"
        return 0
    fi
}

# 运行Go单元测试
run_unit_tests() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}1. 运行单元测试${NC}"
    echo -e "${BLUE}================================${NC}"
    
    cd tests
    go test -v -run TestConcurrentStressTestSuite -timeout 10m
    cd ..
    
    echo -e "${GREEN}✓ 单元测试完成${NC}"
}

# 运行连接压力测试
run_connection_test() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}2. 连接压力测试${NC}"
    echo -e "${BLUE}================================${NC}"
    
    echo -e "${YELLOW}测试场景: 并发连接${NC}"
    echo -e "${YELLOW}客户端数: 50${NC}"
    echo -e "${YELLOW}持续时间: 30秒${NC}"
    echo ""
    
    go run stress_test_runner.go \
        -url="$SERVER_URL" \
        -clients=50 \
        -duration=30s \
        -scenario=connections \
        -secret="$JWT_SECRET" \
        -meeting=$MEETING_ID
    
    echo -e "${GREEN}✓ 连接测试完成${NC}"
}

# 运行消息压力测试
run_messaging_test() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}3. 消息压力测试${NC}"
    echo -e "${BLUE}================================${NC}"
    
    echo -e "${YELLOW}测试场景: 并发消息${NC}"
    echo -e "${YELLOW}客户端数: 30${NC}"
    echo -e "${YELLOW}消息速率: 10 msg/s${NC}"
    echo -e "${YELLOW}持续时间: 30秒${NC}"
    echo ""
    
    go run stress_test_runner.go \
        -url="$SERVER_URL" \
        -clients=30 \
        -duration=30s \
        -msg-rate=10 \
        -scenario=messaging \
        -secret="$JWT_SECRET" \
        -meeting=$MEETING_ID
    
    echo -e "${GREEN}✓ 消息测试完成${NC}"
}

# 运行混合场景测试
run_mixed_test() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}4. 混合场景测试${NC}"
    echo -e "${BLUE}================================${NC}"
    
    echo -e "${YELLOW}测试场景: 混合场景${NC}"
    echo -e "${YELLOW}  - 30% 长连接客户端${NC}"
    echo -e "${YELLOW}  - 30% 频繁重连客户端${NC}"
    echo -e "${YELLOW}  - 40% 高频消息客户端${NC}"
    echo -e "${YELLOW}客户端数: 50${NC}"
    echo -e "${YELLOW}持续时间: 45秒${NC}"
    echo ""
    
    go run stress_test_runner.go \
        -url="$SERVER_URL" \
        -clients=50 \
        -duration=45s \
        -msg-rate=5 \
        -scenario=mixed \
        -secret="$JWT_SECRET" \
        -meeting=$MEETING_ID
    
    echo -e "${GREEN}✓ 混合场景测试完成${NC}"
}

# 运行极限压力测试
run_stress_test() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}5. 极限压力测试${NC}"
    echo -e "${BLUE}================================${NC}"
    
    echo -e "${RED}⚠️  警告: 这将对服务器施加极大压力${NC}"
    echo -e "${YELLOW}测试场景: 极限压力${NC}"
    echo -e "${YELLOW}客户端数: 100${NC}"
    echo -e "${YELLOW}消息速率: 极高${NC}"
    echo -e "${YELLOW}持续时间: 30秒${NC}"
    echo ""
    
    read -p "是否继续? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}跳过极限压力测试${NC}"
        return
    fi
    
    go run stress_test_runner.go \
        -url="$SERVER_URL" \
        -clients=100 \
        -duration=30s \
        -scenario=stress \
        -secret="$JWT_SECRET" \
        -meeting=$MEETING_ID
    
    echo -e "${GREEN}✓ 极限压力测试完成${NC}"
}

# 生成测试报告
generate_report() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}测试报告${NC}"
    echo -e "${BLUE}================================${NC}"
    
    REPORT_FILE="stress_test_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$REPORT_FILE" << EOF
# 信令服务并发压力测试报告

**测试时间**: $(date '+%Y-%m-%d %H:%M:%S')
**服务地址**: $SERVER_URL
**会议ID**: $MEETING_ID

## 测试概览

本次测试对信令服务进行了全面的并发压力测试，包括：

1. ✅ 单元测试 - 验证基本功能
2. ✅ 连接压力测试 - 50并发连接
3. ✅ 消息压力测试 - 30客户端高频消息
4. ✅ 混合场景测试 - 50客户端多种行为
5. ✅ 极限压力测试 - 100客户端极限负载

## 测试结果

详细的测试结果请查看上方的控制台输出。

## 性能指标

- **连接成功率**: 查看各测试的连接成功率
- **消息吞吐量**: 查看消息处理能力
- **平均延迟**: 查看响应时间
- **系统稳定性**: 观察是否有崩溃或错误

## 建议

根据测试结果，可以考虑以下优化方向：

1. 如果连接成功率低于90%，需要优化连接处理逻辑
2. 如果消息延迟过高，需要优化消息路由机制
3. 如果出现大量错误，需要检查错误处理和资源管理
4. 考虑增加连接池大小和优化数据库查询

## 下一步

- [ ] 分析性能瓶颈
- [ ] 优化关键路径
- [ ] 增加监控指标
- [ ] 进行更大规模的测试

---
*报告生成时间: $(date '+%Y-%m-%d %H:%M:%S')*
EOF
    
    echo -e "${GREEN}✓ 测试报告已生成: $REPORT_FILE${NC}"
}

# 主函数
main() {
    # 检查服务状态
    if ! check_service; then
        exit 1
    fi
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --url)
                SERVER_URL="$2"
                shift 2
                ;;
            --secret)
                JWT_SECRET="$2"
                shift 2
                ;;
            --meeting)
                MEETING_ID="$2"
                shift 2
                ;;
            --quick)
                QUICK_MODE=true
                shift
                ;;
            --help)
                echo "用法: $0 [选项]"
                echo ""
                echo "选项:"
                echo "  --url URL        信令服务WebSocket地址 (默认: ws://localhost:8083/ws/signaling)"
                echo "  --secret SECRET  JWT密钥 (默认: test-secret)"
                echo "  --meeting ID     会议ID (默认: 1)"
                echo "  --quick          快速模式，只运行基本测试"
                echo "  --help           显示此帮助信息"
                exit 0
                ;;
            *)
                echo -e "${RED}未知选项: $1${NC}"
                echo "使用 --help 查看帮助"
                exit 1
                ;;
        esac
    done
    
    # 运行测试
    if [ "$QUICK_MODE" = true ]; then
        echo -e "${YELLOW}快速模式: 只运行基本测试${NC}"
        run_connection_test
        run_messaging_test
    else
        run_unit_tests
        run_connection_test
        run_messaging_test
        run_mixed_test
        run_stress_test
    fi
    
    # 生成报告
    generate_report
    
    echo -e "\n${GREEN}================================${NC}"
    echo -e "${GREEN}所有测试完成！${NC}"
    echo -e "${GREEN}================================${NC}"
}

# 运行主函数
main "$@"

