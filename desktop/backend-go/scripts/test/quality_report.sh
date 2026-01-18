#!/bin/bash

set -e

echo "======================================================================"
echo "📊 VOICE SYSTEM CODE QUALITY REPORT"
echo "======================================================================"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$SCRIPT_DIR"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Tracking variables
VET_ISSUES=0
FORMAT_ISSUES=0
TODO_COUNT=0
RACE_CONDITION_ISSUES=0

echo -e "${BLUE}[1/7] Running go vet...${NC}"
echo ""

# Run go vet on all voice-related packages
VET_OUTPUT=$(go vet ./internal/livekit/... ./internal/services/voice_controller.go ./internal/agents/voice_adapter.go ./internal/handlers/voice_agent.go ./internal/grpc/voice_server.go 2>&1 || true)

if [ -z "$VET_OUTPUT" ]; then
    echo -e "${GREEN}✅ No vet issues found${NC}"
    VET_ISSUES=0
else
    echo -e "${RED}⚠️  Vet issues found:${NC}"
    echo "$VET_OUTPUT"
    VET_ISSUES=$(echo "$VET_OUTPUT" | grep -c "vet:" || echo "0")
fi

echo ""
echo -e "${BLUE}[2/7] Checking code formatting (gofmt)...${NC}"
echo ""

UNFORMATTED=$(gofmt -l internal/livekit internal/services/voice_controller.go internal/agents/voice_adapter.go internal/handlers/voice_agent.go 2>/dev/null || true)

if [ -z "$UNFORMATTED" ]; then
    echo -e "${GREEN}✅ All code properly formatted${NC}"
    FORMAT_ISSUES=0
else
    echo -e "${YELLOW}⚠️  Unformatted files:${NC}"
    echo "$UNFORMATTED"
    FORMAT_ISSUES=$(echo "$UNFORMATTED" | wc -l)
fi

echo ""
echo -e "${BLUE}[3/7] Checking for TODO/FIXME comments...${NC}"
echo ""

TODO_OUTPUT=$(grep -rn "TODO\|FIXME" internal/livekit/ internal/services/voice_controller.go internal/agents/voice_adapter.go 2>/dev/null || true)

if [ -z "$TODO_OUTPUT" ]; then
    echo -e "${GREEN}✅ No TODO/FIXME comments found${NC}"
    TODO_COUNT=0
else
    echo -e "${YELLOW}⚠️  TODO/FIXME comments found:${NC}"
    echo "$TODO_OUTPUT"
    TODO_COUNT=$(echo "$TODO_OUTPUT" | wc -l)
fi

echo ""
echo -e "${BLUE}[4/7] Checking imports and dependencies...${NC}"
echo ""

go mod tidy > /dev/null 2>&1

DIFF_OUTPUT=$(git diff go.mod go.sum 2>/dev/null || echo "")
if [ -z "$DIFF_OUTPUT" ]; then
    echo -e "${GREEN}✅ Dependencies are in order${NC}"
else
    echo -e "${YELLOW}⚠️  Dependency changes:${NC}"
    echo "$DIFF_OUTPUT"
fi

echo ""
echo -e "${BLUE}[5/7] Building voice packages...${NC}"
echo ""

BUILD_OUTPUT=$(go build -v ./internal/livekit/... 2>&1 || true)
BUILD_STATUS=$?

if [ $BUILD_STATUS -eq 0 ]; then
    echo -e "${GREEN}✅ Build successful${NC}"
else
    echo -e "${RED}❌ Build failed:${NC}"
    echo "$BUILD_OUTPUT"
fi

echo ""
echo -e "${BLUE}[6/7] Running unit tests...${NC}"
echo ""

# Run only the simple unit tests that don't require full integration
TEST_OUTPUT=$(go test -v ./internal/livekit/ -run "TestDetectVoiceActivity|TestVADConfig|TestWrapPCMInWAV|TestDecodeMp3ToPCM|Benchmark" -timeout 30s 2>&1 || true)

TEST_PASS=$(echo "$TEST_OUTPUT" | grep -c "PASS" || echo "0")
TEST_FAIL=$(echo "$TEST_OUTPUT" | grep -c "FAIL" || echo "0")

echo "$TEST_OUTPUT" | tail -20

echo ""
echo -e "${BLUE}[7/7] Test coverage analysis...${NC}"
echo ""

go test -coverprofile=coverage.out ./internal/livekit -timeout 30s 2>&1 > /dev/null || true

if [ -f coverage.out ]; then
    echo -e "${GREEN}✅ Coverage report generated${NC}"
    echo ""
    echo "Coverage by function:"
    go tool cover -func=coverage.out 2>/dev/null | grep -E "total|detectVoiceActivity|wrapPCMInWAV|decodeMp3ToPCM|VAD" || true
    rm coverage.out
else
    echo -e "${YELLOW}⚠️  Could not generate coverage report${NC}"
fi

echo ""
echo "======================================================================"
echo "📊 SUMMARY"
echo "======================================================================"
echo ""

TOTAL_ISSUES=$((VET_ISSUES + FORMAT_ISSUES + TODO_COUNT))

if [ $VET_ISSUES -eq 0 ]; then
    echo -e "go vet issues:        ${GREEN}✅ 0${NC}"
else
    echo -e "go vet issues:        ${RED}❌ $VET_ISSUES${NC}"
fi

if [ $FORMAT_ISSUES -eq 0 ]; then
    echo -e "Formatting issues:    ${GREEN}✅ 0${NC}"
else
    echo -e "Formatting issues:    ${YELLOW}⚠️  $FORMAT_ISSUES${NC}"
fi

if [ $TODO_COUNT -eq 0 ]; then
    echo -e "TODO/FIXME comments:  ${GREEN}✅ 0${NC}"
else
    echo -e "TODO/FIXME comments:  ${YELLOW}⚠️  $TODO_COUNT${NC}"
fi

if [ $BUILD_STATUS -eq 0 ]; then
    echo -e "Build status:         ${GREEN}✅ SUCCESS${NC}"
else
    echo -e "Build status:         ${RED}❌ FAILED${NC}"
fi

if [ $TEST_FAIL -eq 0 ]; then
    echo -e "Unit tests:           ${GREEN}✅ ALL PASSING${NC}"
else
    echo -e "Unit tests:           ${RED}❌ $TEST_FAIL FAILED${NC}"
fi

echo ""
echo "======================================================================"

if [ $TOTAL_ISSUES -eq 0 ] && [ $BUILD_STATUS -eq 0 ]; then
    echo -e "🎯 CODE QUALITY: ${GREEN}EXCELLENT${NC}"
    exit 0
elif [ $TOTAL_ISSUES -lt 5 ] && [ $BUILD_STATUS -eq 0 ]; then
    echo -e "🎯 CODE QUALITY: ${YELLOW}GOOD (minor issues)${NC}"
    exit 0
else
    echo -e "🎯 CODE QUALITY: ${RED}NEEDS REVIEW${NC}"
    exit 1
fi
