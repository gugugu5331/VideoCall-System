#!/bin/bash

################################################################################
# 快速微服务集成测试
################################################################################

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PASSED=0
FAILED=0

test_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ PASS${NC} - $2"
        ((PASSED++))
    else
        echo -e "${RED}❌ FAIL${NC} - $2: $3"
        ((FAILED++))
    fi
}

echo "================================================================================"
echo "Microservices Integration Test"
echo "================================================================================"

# Test 1: Docker containers
echo -e "\n${BLUE}[1/5]${NC} Testing infrastructure containers..."
for svc in meeting-postgres meeting-redis meeting-mongodb meeting-etcd meeting-minio; do
    if docker ps | grep -q "$svc"; then
        test_result 0 "$svc"
    else
        test_result 1 "$svc" "not running"
    fi
done

# Test 2: Microservices containers
echo -e "\n${BLUE}[2/5]${NC} Testing microservice containers..."
for svc in meeting-user-service meeting-meeting-service meeting-signaling-service meeting-media-service meeting-ai-service; do
    if docker ps | grep -q "$svc"; then
        test_result 0 "$svc"
    else
        test_result 1 "$svc" "not running"
    fi
done

# Test 3: Service registration in etcd
echo -e "\n${BLUE}[3/5]${NC} Testing service registration in etcd..."
ETCD_DATA=$(docker exec meeting-etcd etcdctl get /services/ --prefix --keys-only 2>/dev/null)

if echo "$ETCD_DATA" | grep -q "/services/user-service/"; then
    test_result 0 "user-service registration"
else
    test_result 1 "user-service registration" "not found in etcd"
fi

if echo "$ETCD_DATA" | grep -q "/services/meeting-service/"; then
    test_result 0 "meeting-service registration"
else
    test_result 1 "meeting-service registration" "not found in etcd"
fi

# Test 4: Service HTTP endpoints
echo -e "\n${BLUE}[4/5]${NC} Testing service HTTP endpoints..."

# User service
USER_RESP=$(docker exec meeting-user-service wget -q -O- --timeout=3 "http://localhost:8080/health" 2>/dev/null || echo "")
if [ -n "$USER_RESP" ]; then
    test_result 0 "user-service HTTP endpoint"
else
    test_result 1 "user-service HTTP endpoint" "no response"
fi

# Meeting service
MEETING_RESP=$(docker exec meeting-meeting-service wget -q -S -O- --timeout=3 "http://localhost:8082/api/v1/meetings" 2>&1 || echo "")
if echo "$MEETING_RESP" | grep -q "HTTP/1.1"; then
    test_result 0 "meeting-service HTTP endpoint"
else
    test_result 1 "meeting-service HTTP endpoint" "no response"
fi

# Media service
MEDIA_RESP=$(docker exec meeting-media-service wget -q -O- --timeout=3 "http://localhost:8083/health" 2>/dev/null || echo "")
if [ -n "$MEDIA_RESP" ]; then
    test_result 0 "media-service HTTP endpoint"
else
    test_result 1 "media-service HTTP endpoint" "no response"
fi

# AI service
AI_RESP=$(docker exec meeting-ai-service wget -q -O- --timeout=3 "http://localhost:8084/health" 2>/dev/null || echo "")
if [ -n "$AI_RESP" ]; then
    test_result 0 "ai-service HTTP endpoint"
else
    test_result 1 "ai-service HTTP endpoint" "no response"
fi

# Test 5: Service discovery
echo -e "\n${BLUE}[5/5]${NC} Testing service discovery..."

# Count registered services
USER_COUNT=$(echo "$ETCD_DATA" | grep -c "/services/user-service/" || echo "0")
MEETING_COUNT=$(echo "$ETCD_DATA" | grep -c "/services/meeting-service/" || echo "0")

echo "  user-service instances: $USER_COUNT"
echo "  meeting-service instances: $MEETING_COUNT"

if [ "$USER_COUNT" -gt 0 ] && [ "$MEETING_COUNT" -gt 0 ]; then
    test_result 0 "service discovery"
else
    test_result 1 "service discovery" "insufficient instances"
fi

# Summary
echo ""
echo "================================================================================"
echo "Test Summary"
echo "================================================================================"
echo "Total: $((PASSED + FAILED))"
echo "Passed: $PASSED"
echo "Failed: $FAILED"
echo "Success Rate: $(awk "BEGIN {printf \"%.1f\", ($PASSED/($PASSED+$FAILED))*100}")%"
echo "================================================================================"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✅ ALL TESTS PASSED${NC}\n"
    exit 0
else
    echo -e "\n${YELLOW}⚠ SOME TESTS FAILED${NC}\n"
    exit 1
fi

