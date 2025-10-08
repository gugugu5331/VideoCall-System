#!/bin/bash

################################################################################
# 微服务架构完整集成测试
# 测试服务发现、服务注册和 Nginx 网关路由功能
################################################################################

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo "================================================================================"
echo -e "${CYAN}Microservices Architecture Integration Test Suite${NC}"
echo "================================================================================"
echo "Test Date: $(date '+%Y-%m-%d %H:%M:%S')"
echo "Test Scope:"
echo "  - Service Discovery & Registration (etcd)"
echo "  - Nginx Gateway Routing"
echo "  - Microservices Health & Functionality"
echo "  - Real Implementation Verification"
echo "================================================================================"
echo ""

TOTAL_PASSED=0
TOTAL_FAILED=0

# Test 1: Quick Integration Test
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Test Suite 1: Service Discovery & Registration${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════════════${NC}"
echo ""

if ./quick_integration_test.sh; then
    echo -e "\n${GREEN}✅ Service Discovery & Registration: PASSED${NC}"
    ((TOTAL_PASSED++))
else
    echo -e "\n${RED}❌ Service Discovery & Registration: FAILED${NC}"
    ((TOTAL_FAILED++))
fi

# Test 2: Nginx Gateway Test
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}Test Suite 2: Nginx Gateway Routing${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════════════${NC}"
echo ""

if ./test_nginx_gateway.sh; then
    echo -e "\n${GREEN}✅ Nginx Gateway Routing: PASSED${NC}"
    ((TOTAL_PASSED++))
else
    echo -e "\n${RED}❌ Nginx Gateway Routing: FAILED${NC}"
    ((TOTAL_FAILED++))
fi

# Final Summary
echo ""
echo "================================================================================"
echo -e "${CYAN}Final Test Summary${NC}"
echo "================================================================================"
echo ""
echo "Test Suites Executed: $((TOTAL_PASSED + TOTAL_FAILED))"
echo "Passed: $TOTAL_PASSED"
echo "Failed: $TOTAL_FAILED"
echo ""

if [ $TOTAL_FAILED -eq 0 ]; then
    echo -e "${GREEN}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                                           ║${NC}"
    echo -e "${GREEN}║                    ✅ ALL TESTS PASSED (100%)                            ║${NC}"
    echo -e "${GREEN}║                                                                           ║${NC}"
    echo -e "${GREEN}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "✓ Service Discovery: Working"
    echo "✓ Service Registration: Working"
    echo "✓ Nginx Gateway: Working"
    echo "✓ Microservices: All Running"
    echo "✓ Real Implementation: Verified"
    echo ""
    echo -e "${GREEN}The microservices architecture is production-ready!${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}╔═══════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${RED}║                                                                           ║${NC}"
    echo -e "${RED}║                    ❌ SOME TESTS FAILED                                   ║${NC}"
    echo -e "${RED}║                                                                           ║${NC}"
    echo -e "${RED}╚═══════════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Please review the test output above for details."
    echo ""
    exit 1
fi

