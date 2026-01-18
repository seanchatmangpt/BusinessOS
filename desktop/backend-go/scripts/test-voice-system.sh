#!/bin/bash
# Voice System Integration Test Script
# Tests all voice agent endpoints to ensure they're working

echo "🧪 Voice System Integration Test"
echo "=================================="
echo ""

BASE_URL="http://localhost:8001"
FAILED=0
PASSED=0

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

test_endpoint() {
    local method=$1
    local endpoint=$2
    local description=$3
    local expected_code=$4

    echo -n "Testing: $description... "

    if [ "$method" = "GET" ]; then
        response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL$endpoint")
    else
        response=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d '{}')
    fi

    if [ "$response" = "$expected_code" ]; then
        echo -e "${GREEN}✅ PASS${NC} (HTTP $response)"
        ((PASSED++))
    else
        echo -e "${RED}❌ FAIL${NC} (Expected $expected_code, got $response)"
        ((FAILED++))
    fi
}

test_sse_endpoint() {
    echo -n "Testing: SSE Voice Events Stream... "

    # Test SSE endpoint - should get 'event: connected'
    response=$(curl -s -N -m 2 "$BASE_URL/api/voice/events" 2>&1 | head -1)

    if echo "$response" | grep -q "event: connected"; then
        echo -e "${GREEN}✅ PASS${NC} (SSE streaming)"
        ((PASSED++))
    else
        echo -e "${RED}❌ FAIL${NC} (No SSE connection event)"
        ((FAILED++))
    fi
}

echo "1. Voice Events (SSE)"
echo "--------------------"
test_endpoint "GET" "/api/voice/events/health" "Voice Events Health" "200"
test_sse_endpoint

echo ""
echo "2. Voice UI Commands"
echo "--------------------"
# These will return 400/500 without proper payload, but should not 404
test_endpoint "POST" "/api/ui/open-module" "Open Module Command" "400"
test_endpoint "POST" "/api/ui/open-app" "Open App Command" "400"
test_endpoint "POST" "/api/ui/switch-desktop" "Switch Desktop Command" "400"

echo ""
echo "3. Voice UI Discovery"
echo "--------------------"
# These need auth, so we expect 401, not 404
test_endpoint "GET" "/api/ui/state" "UI State Discovery" "401"
test_endpoint "GET" "/api/ui/modules" "Available Modules" "401"
test_endpoint "GET" "/api/ui/apps" "Installed Apps" "401"
test_endpoint "GET" "/api/ui/modules/search" "Module Search" "401"

echo ""
echo "4. Voice Node Queries"
echo "--------------------"
test_endpoint "GET" "/api/nodes/123/context" "Node Context" "401"
test_endpoint "GET" "/api/nodes/search" "Node Search" "401"
test_endpoint "GET" "/api/nodes/123/activity" "Node Activity" "401"
test_endpoint "GET" "/api/nodes/123/decisions" "Node Decisions" "401"
test_endpoint "GET" "/api/projects/123/tasks" "Project Tasks" "401"

echo ""
echo "=================================="
echo "Test Results:"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    echo ""
    echo "Voice system is ready to use!"
    echo "Next steps:"
    echo "  1. Restart frontend: cd frontend && npm run dev"
    echo "  2. Check browser console for SSE connection"
    echo "  3. Test voice commands via LiveKit Python agent"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    echo ""
    echo "Check if the Go backend is running on port 8001"
    exit 1
fi
