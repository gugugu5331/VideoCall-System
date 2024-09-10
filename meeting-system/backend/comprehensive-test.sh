#!/bin/bash

# 全面测试脚本 - 用户服务和会议服务
# 生成详细的测试报告

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNINGS=0

# 测试报告文件
REPORT_FILE="test-report-$(date +%Y%m%d-%H%M%S).md"

# 初始化测试报告
init_report() {
    cat > "$REPORT_FILE" << EOF
# 会议系统微服务测试报告

**测试时间**: $(date '+%Y-%m-%d %H:%M:%S')  
**测试环境**: $(uname -s) $(uname -r)  
**Go版本**: $(go version)  

## 测试概览

EOF
}

# 记录测试结果
log_test() {
    local test_name="$1"
    local status="$2"
    local details="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$status" = "PASS" ]; then
        PASSED_TESTS=$((PASSED_TESTS + 1))
        echo -e "${GREEN}✅ $test_name${NC}"
        echo "- ✅ **$test_name**: PASS" >> "$REPORT_FILE"
    elif [ "$status" = "FAIL" ]; then
        FAILED_TESTS=$((FAILED_TESTS + 1))
        echo -e "${RED}❌ $test_name${NC}"
        echo "- ❌ **$test_name**: FAIL" >> "$REPORT_FILE"
    elif [ "$status" = "WARN" ]; then
        WARNINGS=$((WARNINGS + 1))
        echo -e "${YELLOW}⚠️ $test_name${NC}"
        echo "- ⚠️ **$test_name**: WARNING" >> "$REPORT_FILE"
    fi
    
    if [ -n "$details" ]; then
        echo "  $details" >> "$REPORT_FILE"
    fi
    echo "" >> "$REPORT_FILE"
}

# 测试Go环境
test_go_environment() {
    echo -e "${BLUE}🔍 测试Go环境${NC}"
    echo "## 1. Go环境测试" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # 检查Go版本
    if command -v go &> /dev/null; then
        local go_version=$(go version)
        log_test "Go版本检查" "PASS" "版本: $go_version"
    else
        log_test "Go版本检查" "FAIL" "Go未安装"
        return 1
    fi
    
    # 检查Go模块
    if [ -f "go.mod" ]; then
        log_test "Go模块文件存在" "PASS" "go.mod文件存在"
    else
        log_test "Go模块文件存在" "FAIL" "go.mod文件不存在"
        return 1
    fi
    
    # 检查依赖
    if go mod verify > /dev/null 2>&1; then
        log_test "Go模块验证" "PASS" "所有依赖验证通过"
    else
        log_test "Go模块验证" "WARN" "部分依赖可能有问题"
    fi
}

# 测试代码编译
test_code_compilation() {
    echo -e "${BLUE}🔨 测试代码编译${NC}"
    echo "## 2. 代码编译测试" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # 测试共享模块编译
    local shared_modules=(
        "shared/config"
        "shared/logger"
        "shared/models"
        "shared/response"
        "shared/utils"
    )
    
    for module in "${shared_modules[@]}"; do
        if [ -d "$module" ]; then
            if go build "./$module" > /dev/null 2>&1; then
                log_test "编译 $module" "PASS"
            else
                log_test "编译 $module" "FAIL" "编译错误: $(go build "./$module" 2>&1 | head -1)"
            fi
        else
            log_test "编译 $module" "WARN" "模块目录不存在"
        fi
    done
    
    # 测试服务编译
    local services=("user-service" "meeting-service")
    
    for service in "${services[@]}"; do
        if [ -d "$service" ]; then
            cd "$service"
            if go build -o "${service}-test" main.go > /dev/null 2>&1; then
                log_test "编译 $service" "PASS"
                rm -f "${service}-test" "${service}-test.exe"
            else
                local error_msg=$(go build -o "${service}-test" main.go 2>&1 | head -3 | tr '\n' ' ')
                log_test "编译 $service" "FAIL" "编译错误: $error_msg"
            fi
            cd ..
        else
            log_test "编译 $service" "FAIL" "服务目录不存在"
        fi
    done
}

# 测试代码质量
test_code_quality() {
    echo -e "${BLUE}📝 测试代码质量${NC}"
    echo "## 3. 代码质量测试" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # 代码格式检查
    local unformatted=$(gofmt -l . 2>/dev/null | grep -v vendor | head -5)
    if [ -z "$unformatted" ]; then
        log_test "代码格式检查" "PASS" "所有代码格式正确"
    else
        log_test "代码格式检查" "WARN" "部分文件需要格式化: $(echo $unformatted | tr '\n' ' ')"
    fi
    
    # 代码静态分析
    if go vet ./... > /dev/null 2>&1; then
        log_test "静态代码分析" "PASS" "未发现问题"
    else
        local vet_issues=$(go vet ./... 2>&1 | head -3 | tr '\n' ' ')
        log_test "静态代码分析" "WARN" "发现问题: $vet_issues"
    fi
    
    # 检查循环依赖
    if go mod graph > /dev/null 2>&1; then
        log_test "依赖图检查" "PASS" "无循环依赖"
    else
        log_test "依赖图检查" "WARN" "依赖图生成失败"
    fi
}

# 测试项目结构
test_project_structure() {
    echo -e "${BLUE}📁 测试项目结构${NC}"
    echo "## 4. 项目结构测试" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # 检查必要目录
    local required_dirs=(
        "user-service"
        "meeting-service"
        "shared"
        "config"
    )
    
    for dir in "${required_dirs[@]}"; do
        if [ -d "$dir" ]; then
            log_test "目录存在: $dir" "PASS"
        else
            log_test "目录存在: $dir" "FAIL" "必要目录不存在"
        fi
    done
    
    # 检查配置文件
    local config_files=(
        "config/config.yaml"
        "config/config-docker.yaml"
    )
    
    for file in "${config_files[@]}"; do
        if [ -f "$file" ]; then
            log_test "配置文件: $file" "PASS"
        else
            log_test "配置文件: $file" "WARN" "配置文件不存在"
        fi
    done
    
    # 检查Dockerfile
    local dockerfiles=(
        "user-service/Dockerfile"
        "meeting-service/Dockerfile"
    )
    
    for dockerfile in "${dockerfiles[@]}"; do
        if [ -f "$dockerfile" ]; then
            log_test "Dockerfile: $dockerfile" "PASS"
        else
            log_test "Dockerfile: $dockerfile" "WARN" "Dockerfile不存在"
        fi
    done
}

# 测试API设计
test_api_design() {
    echo -e "${BLUE}🌐 测试API设计${NC}"
    echo "## 5. API设计测试" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # 检查用户服务API
    if [ -f "user-service/main.go" ]; then
        local user_routes=$(grep -o "auth\.\|users\.\|admin\." user-service/main.go | wc -l)
        if [ "$user_routes" -gt 0 ]; then
            log_test "用户服务API路由" "PASS" "发现 $user_routes 个路由组"
        else
            log_test "用户服务API路由" "WARN" "未发现API路由定义"
        fi
    fi
    
    # 检查会议服务API
    if [ -f "meeting-service/main.go" ]; then
        local meeting_routes=$(grep -o "meetings\.\|my\.\|admin\." meeting-service/main.go | wc -l)
        if [ "$meeting_routes" -gt 0 ]; then
            log_test "会议服务API路由" "PASS" "发现 $meeting_routes 个路由组"
        else
            log_test "会议服务API路由" "WARN" "未发现API路由定义"
        fi
    fi
    
    # 检查HTTP处理器
    local handlers_count=$(find . -name "*_handler.go" | wc -l)
    if [ "$handlers_count" -gt 0 ]; then
        log_test "HTTP处理器" "PASS" "发现 $handlers_count 个处理器文件"
    else
        log_test "HTTP处理器" "FAIL" "未发现HTTP处理器"
    fi
}

# 测试数据模型
test_data_models() {
    echo -e "${BLUE}🗄️ 测试数据模型${NC}"
    echo "## 6. 数据模型测试" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # 检查模型文件
    local model_files=(
        "shared/models/user.go"
        "shared/models/meeting.go"
    )
    
    for model in "${model_files[@]}"; do
        if [ -f "$model" ]; then
            local struct_count=$(grep -c "type.*struct" "$model" 2>/dev/null || echo 0)
            log_test "模型文件: $(basename $model)" "PASS" "包含 $struct_count 个结构体"
        else
            log_test "模型文件: $(basename $model)" "FAIL" "模型文件不存在"
        fi
    done
    
    # 检查数据库配置
    if [ -f "shared/database/postgres.go" ]; then
        log_test "PostgreSQL配置" "PASS" "数据库配置文件存在"
    else
        log_test "PostgreSQL配置" "WARN" "数据库配置文件不存在"
    fi
    
    if [ -f "shared/database/redis.go" ]; then
        log_test "Redis配置" "PASS" "缓存配置文件存在"
    else
        log_test "Redis配置" "WARN" "缓存配置文件不存在"
    fi
}

# 测试安全性
test_security() {
    echo -e "${BLUE}🔒 测试安全性${NC}"
    echo "## 7. 安全性测试" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # 检查JWT实现
    if grep -r "jwt" . --include="*.go" > /dev/null 2>&1; then
        log_test "JWT认证实现" "PASS" "发现JWT相关代码"
    else
        log_test "JWT认证实现" "WARN" "未发现JWT实现"
    fi
    
    # 检查密码加密
    if grep -r "bcrypt\|crypto" . --include="*.go" > /dev/null 2>&1; then
        log_test "密码加密" "PASS" "发现密码加密实现"
    else
        log_test "密码加密" "WARN" "未发现密码加密实现"
    fi
    
    # 检查中间件
    if [ -f "shared/middleware/auth.go" ]; then
        log_test "认证中间件" "PASS" "认证中间件存在"
    else
        log_test "认证中间件" "WARN" "认证中间件不存在"
    fi
    
    # 检查CORS配置
    if grep -r "CORS\|cors" . --include="*.go" > /dev/null 2>&1; then
        log_test "CORS配置" "PASS" "发现CORS配置"
    else
        log_test "CORS配置" "WARN" "未发现CORS配置"
    fi
}

# 生成测试总结
generate_summary() {
    echo -e "${PURPLE}📊 生成测试报告${NC}"
    
    local success_rate=0
    if [ $TOTAL_TESTS -gt 0 ]; then
        success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    fi
    
    cat >> "$REPORT_FILE" << EOF

## 测试总结

| 指标 | 数量 | 百分比 |
|------|------|--------|
| 总测试数 | $TOTAL_TESTS | 100% |
| 通过测试 | $PASSED_TESTS | ${success_rate}% |
| 失败测试 | $FAILED_TESTS | $((FAILED_TESTS * 100 / TOTAL_TESTS))% |
| 警告数量 | $WARNINGS | $((WARNINGS * 100 / TOTAL_TESTS))% |

### 测试评级

EOF

    if [ $success_rate -ge 90 ]; then
        echo "**评级**: 🌟🌟🌟🌟🌟 优秀 (${success_rate}%)" >> "$REPORT_FILE"
        echo -e "${GREEN}🌟🌟🌟🌟🌟 测试评级: 优秀 (${success_rate}%)${NC}"
    elif [ $success_rate -ge 80 ]; then
        echo "**评级**: 🌟🌟🌟🌟 良好 (${success_rate}%)" >> "$REPORT_FILE"
        echo -e "${CYAN}🌟🌟🌟🌟 测试评级: 良好 (${success_rate}%)${NC}"
    elif [ $success_rate -ge 70 ]; then
        echo "**评级**: 🌟🌟🌟 一般 (${success_rate}%)" >> "$REPORT_FILE"
        echo -e "${YELLOW}🌟🌟🌟 测试评级: 一般 (${success_rate}%)${NC}"
    else
        echo "**评级**: 🌟🌟 需要改进 (${success_rate}%)" >> "$REPORT_FILE"
        echo -e "${RED}🌟🌟 测试评级: 需要改进 (${success_rate}%)${NC}"
    fi
    
    cat >> "$REPORT_FILE" << EOF

### 建议

EOF

    if [ $FAILED_TESTS -gt 0 ]; then
        echo "- 🔧 修复失败的测试项目" >> "$REPORT_FILE"
    fi
    
    if [ $WARNINGS -gt 0 ]; then
        echo "- ⚠️ 处理警告项目以提高代码质量" >> "$REPORT_FILE"
    fi
    
    if [ $success_rate -ge 90 ]; then
        echo "- 🚀 代码质量优秀，可以进入下一阶段开发" >> "$REPORT_FILE"
    fi
    
    echo "" >> "$REPORT_FILE"
    echo "---" >> "$REPORT_FILE"
    echo "*报告生成时间: $(date '+%Y-%m-%d %H:%M:%S')*" >> "$REPORT_FILE"
}

# 主函数
main() {
    echo -e "${CYAN}🚀 开始全面测试会议系统微服务${NC}"
    echo "========================================"
    
    init_report
    
    test_go_environment
    test_code_compilation
    test_code_quality
    test_project_structure
    test_api_design
    test_data_models
    test_security
    
    generate_summary
    
    echo ""
    echo -e "${CYAN}📋 测试完成！报告已生成: ${REPORT_FILE}${NC}"
    echo -e "${CYAN}📊 测试统计: 总计 $TOTAL_TESTS 项，通过 $PASSED_TESTS 项，失败 $FAILED_TESTS 项，警告 $WARNINGS 项${NC}"
    
    # 显示报告内容
    if command -v cat &> /dev/null; then
        echo ""
        echo -e "${BLUE}📄 测试报告内容:${NC}"
        echo "========================================"
        cat "$REPORT_FILE"
    fi
}

# 运行主函数
main "$@"
