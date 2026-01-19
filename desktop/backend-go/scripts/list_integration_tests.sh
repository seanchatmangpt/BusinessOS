#!/bin/bash

# list_integration_tests.sh
# Lists all integration tests and their status

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}======================================${NC}"
echo -e "${BLUE}  Pure Go Voice Agent${NC}"
echo -e "${BLUE}  Integration Tests${NC}"
echo -e "${BLUE}======================================${NC}"
echo ""

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

TEST_FILE="$PROJECT_ROOT/internal/livekit/voice_agent_integration_test.go"

if [ ! -f "$TEST_FILE" ]; then
    echo -e "${RED}Error: Integration test file not found${NC}"
    exit 1
fi

# Count tests
TOTAL_TESTS=$(grep -c "^func TestIntegration_" "$TEST_FILE")
SKIPPED_TESTS=$(grep -c "t.Skip(" "$TEST_FILE" || true)
ACTIVE_TESTS=$((TOTAL_TESTS - SKIPPED_TESTS))

echo -e "${GREEN}Total Integration Tests: $TOTAL_TESTS${NC}"
echo -e "  ${GREEN}Active: $ACTIVE_TESTS${NC}"
echo -e "  ${YELLOW}Skipped: $SKIPPED_TESTS${NC}"
echo ""

echo "Tests by Category:"
echo ""

# Room Connection Tests
echo -e "${BLUE}Room Connection (5 tests):${NC}"
grep "^func TestIntegration_" "$TEST_FILE" | grep -i "room\|join\|leave\|disconnect" | sed 's/func /  - /' | sed 's/(t \*testing.T) {//'

echo ""

# Audio Track Tests
echo -e "${BLUE}Audio Tracks (3 tests):${NC}"
grep "^func TestIntegration_" "$TEST_FILE" | grep -i "audio.*track\|track.*audio\|packet" | sed 's/func /  - /' | sed 's/(t \*testing.T) {//'

echo ""

# E2E Voice Tests
echo -e "${BLUE}End-to-End Voice (4 tests):${NC}"
grep "^func TestIntegration_" "$TEST_FILE" | grep -i "stt\|tts\|conversation" | sed 's/func /  - /' | sed 's/(t \*testing.T) {//'

echo ""

# Performance Tests
echo -e "${BLUE}Performance (3 tests):${NC}"
grep "^func TestIntegration_" "$TEST_FILE" | grep -i "latency\|concurrent\|memory" | sed 's/func /  - /' | sed 's/(t \*testing.T) {//'

echo ""

# Error Handling Tests
echo -e "${BLUE}Error Handling (5 tests):${NC}"
grep "^func TestIntegration_" "$TEST_FILE" | grep -i "error\|failure\|notfound\|monitor" | sed 's/func /  - /' | sed 's/(t \*testing.T) {//'

echo ""
echo "Run tests with:"
echo -e "  ${GREEN}./scripts/run_integration_tests.sh${NC}"
echo ""
echo "Documentation:"
echo "  - docs/INTEGRATION_TESTING.md"
echo "  - internal/livekit/README_INTEGRATION_TESTS.md"
echo ""
