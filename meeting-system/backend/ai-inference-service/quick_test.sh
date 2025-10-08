#!/bin/bash

# Quick test script for AI Inference Service

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

HOST="${1:-localhost}"
PORT="${2:-8085}"
BASE_URL="http://$HOST:$PORT"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}AI Inference Service Quick Test${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "Base URL: $BASE_URL"
echo ""

# Test 1: Health Check
echo -e "${YELLOW}[1/6] Testing Health Check...${NC}"
if curl -s -f "$BASE_URL/health" > /dev/null; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
    exit 1
fi

# Test 2: Service Info
echo -e "${YELLOW}[2/6] Testing Service Info...${NC}"
if curl -s -f "$BASE_URL/api/v1/ai/info" > /dev/null; then
    echo -e "${GREEN}✓ Service info passed${NC}"
else
    echo -e "${RED}✗ Service info failed${NC}"
    exit 1
fi

# Test 3: AI Health Check
echo -e "${YELLOW}[3/6] Testing AI Health Check...${NC}"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/v1/ai/health")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ AI health check passed${NC}"
    echo "$BODY" | jq '.' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ AI health check failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
    exit 1
fi

# Test 4: ASR
echo -e "${YELLOW}[4/6] Testing ASR (Speech Recognition)...${NC}"
ASR_DATA='{"audio_data":"c2FtcGxlIGF1ZGlvIGRhdGE=","format":"wav","sample_rate":16000}'
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/ai/asr" \
    -H "Content-Type: application/json" \
    -d "$ASR_DATA")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ ASR test passed${NC}"
    echo "$BODY" | jq '.data.text' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ ASR test failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi

# Test 5: Emotion Detection
echo -e "${YELLOW}[5/6] Testing Emotion Detection...${NC}"
EMOTION_DATA='{"text":"I am very happy today!"}'
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/ai/emotion" \
    -H "Content-Type: application/json" \
    -d "$EMOTION_DATA")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Emotion detection test passed${NC}"
    echo "$BODY" | jq '.data.emotion' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ Emotion detection test failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi

# Test 6: Synthesis Detection
echo -e "${YELLOW}[6/6] Testing Synthesis Detection...${NC}"
SYNTHESIS_DATA='{"audio_data":"c2FtcGxlIGF1ZGlvIGRhdGE=","format":"wav","sample_rate":16000}'
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/ai/synthesis" \
    -H "Content-Type: application/json" \
    -d "$SYNTHESIS_DATA")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✓ Synthesis detection test passed${NC}"
    echo "$BODY" | jq '.data.is_synthetic' 2>/dev/null || echo "$BODY"
else
    echo -e "${RED}✗ Synthesis detection test failed (HTTP $HTTP_CODE)${NC}"
    echo "$BODY"
fi

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}All tests completed!${NC}"
echo -e "${GREEN}========================================${NC}"

