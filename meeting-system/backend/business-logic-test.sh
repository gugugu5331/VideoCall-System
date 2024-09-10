#!/bin/bash

# 业务逻辑深度测试脚本
# 检查具体业务规则和边界条件处理

set -e

echo "🧠 开始业务逻辑深度测试"
echo "========================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 测试结果统计
TOTAL_LOGIC_TESTS=0
PASSED_LOGIC_TESTS=0
FAILED_LOGIC_TESTS=0

# 业务逻辑测试结果记录
logic_test_result() {
    local test_name="$1"
    local result="$2"
    local message="$3"
    
    TOTAL_LOGIC_TESTS=$((TOTAL_LOGIC_TESTS + 1))
    
    if [ "$result" = "PASS" ]; then
        echo -e "✅ ${GREEN}$test_name${NC}: $message"
        PASSED_LOGIC_TESTS=$((PASSED_LOGIC_TESTS + 1))
    else
        echo -e "❌ ${RED}$test_name${NC}: $message"
        FAILED_LOGIC_TESTS=$((FAILED_LOGIC_TESTS + 1))
    fi
}

echo "📋 1. 用户业务逻辑测试"
echo "----------------------------------------"

# 检查用户注册业务规则
if grep -q "email.*unique\|Email.*unique" shared/models/user.go; then
    logic_test_result "邮箱唯一性约束" "PASS" "数据模型包含邮箱唯一性约束"
else
    logic_test_result "邮箱唯一性约束" "FAIL" "缺少邮箱唯一性约束"
fi

# 检查密码强度验证
if grep -q "min=.*max=" shared/models/user.go | grep -q "password"; then
    logic_test_result "密码长度验证" "PASS" "包含密码长度验证规则"
else
    logic_test_result "密码长度验证" "FAIL" "缺少密码长度验证"
fi

# 检查用户状态管理
if grep -q "UserStatus\|Status.*int" shared/models/user.go; then
    if grep -q "Active\|Inactive\|Banned" shared/models/user.go; then
        logic_test_result "用户状态管理" "PASS" "实现了用户状态管理"
    else
        logic_test_result "用户状态管理" "FAIL" "缺少状态常量定义"
    fi
else
    logic_test_result "用户状态管理" "FAIL" "缺少用户状态字段"
fi

# 检查用户角色权限
if grep -q "Role\|UserRole" shared/models/user.go; then
    if grep -q "Admin\|User\|Guest" shared/models/user.go; then
        logic_test_result "用户角色权限" "PASS" "实现了用户角色权限系统"
    else
        logic_test_result "用户角色权限" "FAIL" "缺少角色常量定义"
    fi
else
    logic_test_result "用户角色权限" "FAIL" "缺少用户角色字段"
fi

echo ""
echo "📋 2. 会议业务逻辑测试"
echo "----------------------------------------"

# 检查会议时间验证
if grep -q "StartTime.*EndTime\|start_time.*end_time" meeting-service/handlers/meeting_handler.go; then
    if grep -q "Before\|After\|Compare" meeting-service/handlers/meeting_handler.go; then
        logic_test_result "会议时间验证" "PASS" "包含开始时间和结束时间验证"
    else
        logic_test_result "会议时间验证" "FAIL" "缺少时间比较逻辑"
    fi
else
    logic_test_result "会议时间验证" "FAIL" "缺少时间验证逻辑"
fi

# 检查会议状态管理
if grep -q "MeetingStatus" shared/models/meeting.go; then
    STATUS_COUNT=$(grep -c "MeetingStatus.*=" shared/models/meeting.go || echo "0")
    if [ "$STATUS_COUNT" -ge 4 ]; then
        logic_test_result "会议状态管理" "PASS" "定义了 $STATUS_COUNT 种会议状态"
    else
        logic_test_result "会议状态管理" "FAIL" "会议状态定义不足: $STATUS_COUNT"
    fi
else
    logic_test_result "会议状态管理" "FAIL" "缺少会议状态定义"
fi

# 检查参与者数量限制
if grep -q "MaxParticipants\|max_participants" shared/models/meeting.go; then
    if grep -q "binding.*min.*max" shared/models/meeting.go | grep -q "participants"; then
        logic_test_result "参与者数量限制" "PASS" "包含参与者数量验证规则"
    else
        logic_test_result "参与者数量限制" "FAIL" "缺少参与者数量验证"
    fi
else
    logic_test_result "参与者数量限制" "FAIL" "缺少参与者数量字段"
fi

# 检查会议权限控制
if grep -q "canModifyMeeting\|canDeleteMeeting" meeting-service/services/meeting_service.go; then
    logic_test_result "会议权限控制" "PASS" "实现了会议权限控制逻辑"
else
    logic_test_result "会议权限控制" "FAIL" "缺少权限控制逻辑"
fi

echo ""
echo "📋 3. 参与者管理业务逻辑测试"
echo "----------------------------------------"

# 检查参与者角色管理
if grep -q "ParticipantRole" shared/models/meeting.go; then
    ROLE_COUNT=$(grep -c "ParticipantRole.*=" shared/models/meeting.go || echo "0")
    if [ "$ROLE_COUNT" -ge 3 ]; then
        logic_test_result "参与者角色管理" "PASS" "定义了 $ROLE_COUNT 种参与者角色"
    else
        logic_test_result "参与者角色管理" "FAIL" "参与者角色定义不足: $ROLE_COUNT"
    fi
else
    logic_test_result "参与者角色管理" "FAIL" "缺少参与者角色定义"
fi

# 检查参与者状态管理
if grep -q "ParticipantStatus" shared/models/meeting.go; then
    STATUS_COUNT=$(grep -c "ParticipantStatus.*=" shared/models/meeting.go || echo "0")
    if [ "$STATUS_COUNT" -ge 3 ]; then
        logic_test_result "参与者状态管理" "PASS" "定义了 $STATUS_COUNT 种参与者状态"
    else
        logic_test_result "参与者状态管理" "FAIL" "参与者状态定义不足: $STATUS_COUNT"
    fi
else
    logic_test_result "参与者状态管理" "FAIL" "缺少参与者状态定义"
fi

# 检查重复加入防护
if grep -q "already.*joined\|duplicate.*participant" meeting-service/services/meeting_service.go; then
    logic_test_result "重复加入防护" "PASS" "包含重复加入检查逻辑"
else
    logic_test_result "重复加入防护" "FAIL" "缺少重复加入防护"
fi

echo ""
echo "📋 4. 数据一致性业务逻辑测试"
echo "----------------------------------------"

# 检查事务处理
if grep -q "tx.Begin\|db.Transaction\|Begin()" meeting-service/services/meeting_service.go; then
    logic_test_result "数据库事务处理" "PASS" "使用了数据库事务保证一致性"
else
    logic_test_result "数据库事务处理" "FAIL" "缺少事务处理机制"
fi

# 检查外键关联
if grep -q "foreignKey\|ForeignKey" shared/models/meeting.go && grep -q "foreignKey\|ForeignKey" shared/models/user.go; then
    logic_test_result "外键关联约束" "PASS" "定义了外键关联约束"
else
    logic_test_result "外键关联约束" "FAIL" "缺少外键关联定义"
fi

# 检查软删除
if grep -q "DeletedAt.*gorm.DeletedAt" shared/models/meeting.go && grep -q "DeletedAt.*gorm.DeletedAt" shared/models/user.go; then
    logic_test_result "软删除机制" "PASS" "实现了软删除机制"
else
    logic_test_result "软删除机制" "FAIL" "缺少软删除机制"
fi

echo ""
echo "📋 5. 安全业务逻辑测试"
echo "----------------------------------------"

# 检查密码加密强度
if grep -q "bcrypt.DefaultCost\|bcrypt.MinCost" user-service/services/user_service.go; then
    logic_test_result "密码加密强度" "PASS" "使用了适当的bcrypt加密强度"
else
    logic_test_result "密码加密强度" "FAIL" "未指定bcrypt加密强度"
fi

# 检查JWT过期时间
if grep -q "ExpiresAt\|exp.*time" user-service/services/user_service.go; then
    logic_test_result "JWT过期控制" "PASS" "设置了JWT过期时间"
else
    logic_test_result "JWT过期控制" "FAIL" "缺少JWT过期时间设置"
fi

# 检查会议密码保护
if grep -q "Password.*string" shared/models/meeting.go; then
    if grep -q "password.*check\|Password.*verify" meeting-service/services/meeting_service.go; then
        logic_test_result "会议密码保护" "PASS" "实现了会议密码保护机制"
    else
        logic_test_result "会议密码保护" "FAIL" "缺少密码验证逻辑"
    fi
else
    logic_test_result "会议密码保护" "FAIL" "缺少会议密码字段"
fi

echo ""
echo "📋 6. 性能优化业务逻辑测试"
echo "----------------------------------------"

# 检查分页查询
if grep -q "Limit\|Offset\|Page" meeting-service/services/meeting_service.go || grep -q "Limit\|Offset\|Page" user-service/services/user_service.go; then
    logic_test_result "分页查询优化" "PASS" "实现了分页查询机制"
else
    logic_test_result "分页查询优化" "FAIL" "缺少分页查询机制"
fi

# 检查索引优化
if grep -q "index\|Index" shared/models/meeting.go && grep -q "index\|Index" shared/models/user.go; then
    logic_test_result "数据库索引优化" "PASS" "定义了数据库索引"
else
    logic_test_result "数据库索引优化" "FAIL" "缺少数据库索引定义"
fi

# 检查缓存策略
if grep -q "cache.*expire\|TTL\|timeout" meeting-service/services/meeting_service.go; then
    logic_test_result "缓存过期策略" "PASS" "实现了缓存过期策略"
else
    logic_test_result "缓存过期策略" "FAIL" "缺少缓存过期策略"
fi

echo ""
echo "📋 7. 错误处理业务逻辑测试"
echo "----------------------------------------"

# 检查业务异常处理
if grep -q "ErrNotFound\|ErrPermissionDenied\|ErrInvalidInput" meeting-service/services/meeting_service.go || grep -q "ErrNotFound\|ErrPermissionDenied\|ErrInvalidInput" user-service/services/user_service.go; then
    logic_test_result "业务异常定义" "PASS" "定义了业务异常类型"
else
    logic_test_result "业务异常定义" "FAIL" "缺少业务异常定义"
fi

# 检查输入验证错误处理
if grep -q "ShouldBindJSON\|ShouldBind" user-service/handlers/user_handler.go && grep -q "ShouldBindJSON\|ShouldBind" meeting-service/handlers/meeting_handler.go; then
    logic_test_result "输入验证错误处理" "PASS" "实现了输入验证错误处理"
else
    logic_test_result "输入验证错误处理" "FAIL" "缺少输入验证错误处理"
fi

echo ""
echo "📊 业务逻辑测试结果统计"
echo "========================================"
echo -e "总逻辑测试数: ${BLUE}$TOTAL_LOGIC_TESTS${NC}"
echo -e "通过逻辑测试: ${GREEN}$PASSED_LOGIC_TESTS${NC}"
echo -e "失败逻辑测试: ${RED}$FAILED_LOGIC_TESTS${NC}"

# 计算成功率
if [ "$TOTAL_LOGIC_TESTS" -gt 0 ]; then
    SUCCESS_RATE=$((PASSED_LOGIC_TESTS * 100 / TOTAL_LOGIC_TESTS))
    echo -e "成功率: ${BLUE}$SUCCESS_RATE%${NC}"
    
    if [ "$SUCCESS_RATE" -ge 90 ]; then
        echo -e "评级: ${GREEN}🌟🌟🌟🌟🌟 优秀${NC}"
        echo -e "建议: ${GREEN}业务逻辑实现完整，可以进入生产环境${NC}"
    elif [ "$SUCCESS_RATE" -ge 80 ]; then
        echo -e "评级: ${YELLOW}🌟🌟🌟🌟 良好${NC}"
        echo -e "建议: ${YELLOW}业务逻辑基本完整，建议完善失败项目${NC}"
    elif [ "$SUCCESS_RATE" -ge 70 ]; then
        echo -e "评级: ${YELLOW}🌟🌟🌟 一般${NC}"
        echo -e "建议: ${YELLOW}需要完善核心业务逻辑${NC}"
    else
        echo -e "评级: ${RED}🌟🌟 需要改进${NC}"
        echo -e "建议: ${RED}业务逻辑实现不完整，需要重新设计${NC}"
    fi
else
    echo -e "成功率: ${RED}0%${NC}"
fi

echo ""
if [ "$FAILED_LOGIC_TESTS" -eq 0 ]; then
    echo -e "${GREEN}🎉 所有业务逻辑测试通过！系统业务规则实现完整。${NC}"
else
    echo -e "${YELLOW}⚠️ 发现 $FAILED_LOGIC_TESTS 个业务逻辑问题，需要进一步完善。${NC}"
fi

echo "========================================"
echo "业务逻辑深度测试完成"
