#!/bin/bash

# 运行完整的压力测试
echo "🔥 会议系统压力测试执行器"
echo "========================================"

# 检查依赖
check_dependencies() {
    echo "📋 检查依赖..."
    
    if ! command -v go &> /dev/null; then
        echo "❌ Go 未安装"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        echo "❌ curl 未安装"
        exit 1
    fi
    
    echo "✅ 依赖检查通过"
}

# 停止现有服务
stop_services() {
    echo "🛑 停止现有服务..."
    
    if [ -f "user-service.pid" ]; then
        USER_PID=$(cat user-service.pid)
        kill $USER_PID 2>/dev/null
        rm -f user-service.pid
        echo "停止用户服务 (PID: $USER_PID)"
    fi
    
    if [ -f "meeting-service.pid" ]; then
        MEETING_PID=$(cat meeting-service.pid)
        kill $MEETING_PID 2>/dev/null
        rm -f meeting-service.pid
        echo "停止会议服务 (PID: $MEETING_PID)"
    fi
    
    # 等待进程完全停止
    sleep 2
}

# 启动服务
start_services() {
    echo ""
    echo "🚀 启动服务..."
    echo "----------------------------------------"
    
    # 编译用户服务
    echo "🔨 编译用户服务..."
    cd user-service
    if ! go build -o user-service-stress main.go; then
        echo "❌ 用户服务编译失败"
        exit 1
    fi
    cd ..
    
    # 编译会议服务
    echo "🔨 编译会议服务..."
    cd meeting-service
    if ! go build -o meeting-service-stress main.go; then
        echo "❌ 会议服务编译失败"
        exit 1
    fi
    cd ..
    
    # 启动用户服务
    echo "🚀 启动用户服务..."
    cd user-service
    ./user-service-stress > ../user-service.log 2>&1 &
    USER_SERVICE_PID=$!
    echo $USER_SERVICE_PID > ../user-service.pid
    cd ..
    
    # 启动会议服务
    echo "🚀 启动会议服务..."
    cd meeting-service
    ./meeting-service-stress > ../meeting-service.log 2>&1 &
    MEETING_SERVICE_PID=$!
    echo $MEETING_SERVICE_PID > ../meeting-service.pid
    cd ..
    
    # 等待服务启动
    echo "⏳ 等待服务启动..."
    sleep 5
    
    # 检查服务状态
    echo "🔍 检查服务状态..."
    
    if ! curl -s http://localhost:8081/health > /dev/null; then
        echo "❌ 用户服务启动失败"
        cat user-service.log
        stop_services
        exit 1
    fi
    
    if ! curl -s http://localhost:8082/health > /dev/null; then
        echo "❌ 会议服务启动失败"
        cat meeting-service.log
        stop_services
        exit 1
    fi
    
    echo "✅ 所有服务启动成功"
}

# 编译压力测试工具
compile_stress_test() {
    echo ""
    echo "🔨 编译压力测试工具..."
    echo "----------------------------------------"
    
    cd stress-test
    
    # 初始化go模块（如果需要）
    if [ ! -f "go.mod" ]; then
        go mod init stress-test
    fi
    
    # 下载依赖
    go mod tidy
    
    # 编译
    if go build -o stress-test main.go; then
        echo "✅ 压力测试工具编译成功"
    else
        echo "❌ 压力测试工具编译失败"
        cd ..
        stop_services
        exit 1
    fi
    
    cd ..
}

# 运行压力测试
run_stress_test() {
    echo ""
    echo "🔥 开始压力测试..."
    echo "========================================"
    
    cd stress-test
    
    # 运行压力测试并保存结果
    ./stress-test | tee ../stress-test-results.log
    
    cd ..
    
    echo ""
    echo "📊 压力测试完成！"
    echo "结果已保存到: stress-test-results.log"
}

# 生成测试报告
generate_report() {
    echo ""
    echo "📋 生成详细测试报告..."
    echo "========================================"
    
    TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
    REPORT_FILE="stress-test-report-${TIMESTAMP}.md"
    
    cat > $REPORT_FILE << EOF
# 会议系统压力测试报告

**测试时间**: $(date)  
**测试环境**: $(uname -s) $(uname -r)  
**Go版本**: $(go version)

## 测试配置

- **用户服务**: http://localhost:8081
- **会议服务**: http://localhost:8082
- **并发级别**: 10, 50, 100, 200, 500 用户
- **请求超时**: 10秒
- **测试类型**: 用户注册、用户登录、会议创建

## 测试结果

\`\`\`
$(cat stress-test-results.log)
\`\`\`

## 服务日志

### 用户服务日志
\`\`\`
$(tail -50 user-service.log 2>/dev/null || echo "无日志文件")
\`\`\`

### 会议服务日志
\`\`\`
$(tail -50 meeting-service.log 2>/dev/null || echo "无日志文件")
\`\`\`

## 系统信息

- **CPU**: $(nproc) 核心
- **内存**: $(free -h 2>/dev/null | grep Mem | awk '{print $2}' || echo "未知")
- **磁盘**: $(df -h . | tail -1 | awk '{print $4}' || echo "未知") 可用空间

---
*报告生成时间: $(date)*
EOF

    echo "✅ 详细报告已生成: $REPORT_FILE"
}

# 清理函数
cleanup() {
    echo ""
    echo "🧹 清理资源..."
    stop_services
    
    # 清理编译文件
    rm -f user-service/user-service-stress
    rm -f meeting-service/meeting-service-stress
    rm -f stress-test/stress-test
    
    echo "✅ 清理完成"
}

# 主执行流程
main() {
    # 设置退出时清理
    trap cleanup EXIT
    
    echo "开始时间: $(date)"
    echo ""
    
    # 1. 检查依赖
    check_dependencies
    
    # 2. 停止现有服务
    stop_services
    
    # 3. 启动服务
    start_services
    
    # 4. 编译压力测试工具
    compile_stress_test
    
    # 5. 运行压力测试
    run_stress_test
    
    # 6. 生成报告
    generate_report
    
    echo ""
    echo "🎉 压力测试全部完成！"
    echo "========================================"
    echo "结束时间: $(date)"
}

# 执行主流程
main
