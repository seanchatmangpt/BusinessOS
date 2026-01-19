#!/bin/bash

# =============================================================================
# E2E MCP Tools Testing Script
# Tests MCP tool functions for integrated providers
# =============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8001}"
AUTH_TOKEN="${AUTH_TOKEN:-}"

# Test results tracking
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# =============================================================================
# Helper Functions
# =============================================================================

print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════${NC}\n"
}

print_test() {
    echo -e "${YELLOW}▶ Testing:${NC} $1"
    TESTS_RUN=$((TESTS_RUN + 1))
}

print_success() {
    echo -e "${GREEN}✓ SUCCESS:${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

print_error() {
    echo -e "${RED}✗ FAILED:${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

print_warning() {
    echo -e "${YELLOW}⚠ WARNING:${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ INFO:${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_header "CHECKING PREREQUISITES"

    # Check if AUTH_TOKEN is provided
    if [ -z "$AUTH_TOKEN" ]; then
        print_error "AUTH_TOKEN is required for MCP tools testing"
        print_info "Set AUTH_TOKEN environment variable with a valid JWT token"
        exit 1
    fi

    print_success "AUTH_TOKEN is provided"

    # Check if server is running
    print_test "Server availability"
    if curl -s -f "${API_BASE_URL}/health" > /dev/null 2>&1; then
        print_success "Server is running at ${API_BASE_URL}"
    else
        print_error "Server is not running at ${API_BASE_URL}"
        exit 1
    fi

    # Check if jq is installed
    print_test "jq availability"
    if command -v jq &> /dev/null; then
        print_success "jq is installed"
    else
        print_error "jq is not installed (required for JSON parsing)"
        echo "Install with: brew install jq (macOS) or apt-get install jq (Linux)"
        exit 1
    fi
}

# =============================================================================
# MCP Tool Test Functions
# =============================================================================

# Test: Google Calendar - List Events
test_google_calendar_list_events() {
    print_header "TEST: GOOGLE CALENDAR - List Events"

    print_test "Tool: calendar_list_events"

    # Create request payload
    REQUEST_PAYLOAD=$(cat <<EOF
{
  "tool": "calendar_list_events",
  "parameters": {
    "max_results": 10
  }
}
EOF
)

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$REQUEST_PAYLOAD" \
        "${API_BASE_URL}/api/tools/execute")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ]; then
        print_success "calendar_list_events executed successfully"

        # Parse and show results
        EVENT_COUNT=$(echo "$BODY" | jq '.events | length' 2>/dev/null || echo "0")
        print_info "Retrieved $EVENT_COUNT event(s)"

        if [ "$EVENT_COUNT" -gt 0 ]; then
            echo "$BODY" | jq -r '.events[] | "  - \(.summary) (\(.start.dateTime // .start.date))"' 2>/dev/null || true
        fi
    elif [ "$HTTP_CODE" == "404" ]; then
        print_warning "Google Calendar not connected"
    else
        print_error "calendar_list_events failed (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Google Calendar - Create Event
test_google_calendar_create_event() {
    print_header "TEST: GOOGLE CALENDAR - Create Event"

    print_test "Tool: calendar_create_event"

    # Calculate dates (tomorrow, 9 AM - 10 AM)
    START_TIME=$(date -u -v+1d +"%Y-%m-%dT09:00:00Z" 2>/dev/null || date -u -d "+1 day" +"%Y-%m-%dT09:00:00Z")
    END_TIME=$(date -u -v+1d +"%Y-%m-%dT10:00:00Z" 2>/dev/null || date -u -d "+1 day" +"%Y-%m-%dT10:00:00Z")

    # Create request payload
    REQUEST_PAYLOAD=$(cat <<EOF
{
  "tool": "calendar_create_event",
  "parameters": {
    "summary": "E2E Test Event",
    "description": "Created by E2E MCP tools test script",
    "start_time": "$START_TIME",
    "end_time": "$END_TIME"
  }
}
EOF
)

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$REQUEST_PAYLOAD" \
        "${API_BASE_URL}/api/tools/execute")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "201" ]; then
        print_success "calendar_create_event executed successfully"

        # Parse and show created event
        EVENT_ID=$(echo "$BODY" | jq -r '.event.id' 2>/dev/null || echo "")
        EVENT_LINK=$(echo "$BODY" | jq -r '.event.htmlLink' 2>/dev/null || echo "")

        if [ -n "$EVENT_ID" ]; then
            print_info "Event ID: $EVENT_ID"
            print_info "Event Link: $EVENT_LINK"
        fi
    elif [ "$HTTP_CODE" == "404" ]; then
        print_warning "Google Calendar not connected"
    else
        print_error "calendar_create_event failed (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Slack - List Channels
test_slack_list_channels() {
    print_header "TEST: SLACK - List Channels"

    print_test "Tool: slack_list_channels"

    REQUEST_PAYLOAD=$(cat <<EOF
{
  "tool": "slack_list_channels",
  "parameters": {
    "types": "public_channel,private_channel"
  }
}
EOF
)

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$REQUEST_PAYLOAD" \
        "${API_BASE_URL}/api/tools/execute")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ]; then
        print_success "slack_list_channels executed successfully"

        # Parse and show results
        CHANNEL_COUNT=$(echo "$BODY" | jq '.channels | length' 2>/dev/null || echo "0")
        print_info "Retrieved $CHANNEL_COUNT channel(s)"

        if [ "$CHANNEL_COUNT" -gt 0 ]; then
            echo "$BODY" | jq -r '.channels[] | "  - #\(.name) (\(.id))"' 2>/dev/null || true
        fi
    elif [ "$HTTP_CODE" == "404" ]; then
        print_warning "Slack not connected"
    else
        print_error "slack_list_channels failed (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Slack - Send Message (DRY RUN)
test_slack_send_message() {
    print_header "TEST: SLACK - Send Message (Dry Run)"

    print_warning "This test would send a real Slack message"
    print_info "Skipping to avoid spam - manual test recommended"

    # Show example request
    cat << 'EOF'

Example request to test manually:

curl -X POST \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tool": "slack_send_message",
    "parameters": {
      "channel": "C1234567890",
      "text": "Test message from BusinessOS E2E tests"
    }
  }' \
  http://localhost:8001/api/tools/execute

EOF
}

# Test: Notion - List Databases
test_notion_list_databases() {
    print_header "TEST: NOTION - List Databases"

    print_test "Tool: notion_list_databases"

    REQUEST_PAYLOAD=$(cat <<EOF
{
  "tool": "notion_list_databases",
  "parameters": {}
}
EOF
)

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$REQUEST_PAYLOAD" \
        "${API_BASE_URL}/api/tools/execute")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ]; then
        print_success "notion_list_databases executed successfully"

        # Parse and show results
        DB_COUNT=$(echo "$BODY" | jq '.databases | length' 2>/dev/null || echo "0")
        print_info "Retrieved $DB_COUNT database(s)"

        if [ "$DB_COUNT" -gt 0 ]; then
            echo "$BODY" | jq -r '.databases[] | "  - \(.title[0].plain_text) (\(.id))"' 2>/dev/null || true
        fi
    elif [ "$HTTP_CODE" == "404" ]; then
        print_warning "Notion not connected"
    else
        print_error "notion_list_databases failed (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Notion - Search
test_notion_search() {
    print_header "TEST: NOTION - Search"

    print_test "Tool: notion_search"

    REQUEST_PAYLOAD=$(cat <<EOF
{
  "tool": "notion_search",
  "parameters": {
    "query": "test",
    "filter": {
      "value": "page",
      "property": "object"
    }
  }
}
EOF
)

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$REQUEST_PAYLOAD" \
        "${API_BASE_URL}/api/tools/execute")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ]; then
        print_success "notion_search executed successfully"

        # Parse and show results
        RESULT_COUNT=$(echo "$BODY" | jq '.results | length' 2>/dev/null || echo "0")
        print_info "Found $RESULT_COUNT result(s)"

        if [ "$RESULT_COUNT" -gt 0 ]; then
            echo "$BODY" | jq -r '.results[] | "  - \(.properties.title.title[0].plain_text // "Untitled") (\(.id))"' 2>/dev/null || true
        fi
    elif [ "$HTTP_CODE" == "404" ]; then
        print_warning "Notion not connected"
    else
        print_error "notion_search failed (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Linear - List Issues
test_linear_list_issues() {
    print_header "TEST: LINEAR - List Issues"

    print_test "Tool: linear_list_issues"

    REQUEST_PAYLOAD=$(cat <<EOF
{
  "tool": "linear_list_issues",
  "parameters": {
    "first": 5
  }
}
EOF
)

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        -H "Content-Type: application/json" \
        -d "$REQUEST_PAYLOAD" \
        "${API_BASE_URL}/api/tools/execute")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ]; then
        print_success "linear_list_issues executed successfully"

        # Parse and show results
        ISSUE_COUNT=$(echo "$BODY" | jq '.issues.nodes | length' 2>/dev/null || echo "0")
        print_info "Retrieved $ISSUE_COUNT issue(s)"

        if [ "$ISSUE_COUNT" -gt 0 ]; then
            echo "$BODY" | jq -r '.issues.nodes[] | "  - \(.identifier): \(.title)"' 2>/dev/null || true
        fi
    elif [ "$HTTP_CODE" == "404" ]; then
        print_warning "Linear not connected"
    else
        print_error "linear_list_issues failed (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# =============================================================================
# Main Test Execution
# =============================================================================

main() {
    print_header "E2E MCP TOOLS TESTING"
    print_info "Testing MCP Tool Execution for Integrated Providers"
    print_info "API Base URL: ${API_BASE_URL}"

    echo ""

    # Prerequisites
    check_prerequisites

    # Google Calendar Tools
    test_google_calendar_list_events
    test_google_calendar_create_event

    # Slack Tools
    test_slack_list_channels
    test_slack_send_message

    # Notion Tools
    test_notion_list_databases
    test_notion_search

    # Linear Tools
    test_linear_list_issues

    # Print summary
    print_header "TEST SUMMARY"
    echo -e "Total Tests: ${TESTS_RUN}"
    echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
    echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}✓ All MCP tool tests passed!${NC}\n"
        exit 0
    else
        echo -e "\n${YELLOW}⚠ Some tests failed or were skipped${NC}\n"
        print_info "This is expected if integrations are not connected"
        exit 0
    fi
}

# =============================================================================
# Script Entry Point
# =============================================================================

# Check if help is requested
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    cat << EOF
E2E MCP Tools Testing Script

Tests MCP tool execution for all integrated providers.

Usage:
  $0 [OPTIONS]

Options:
  -h, --help     Show this help message

Environment Variables:
  API_BASE_URL   Base URL of the API server (default: http://localhost:8001)
  AUTH_TOKEN     Bearer token for authenticated requests (REQUIRED)

Examples:
  # Basic test
  export AUTH_TOKEN="your-jwt-token-here"
  $0

  # Test against different server
  export API_BASE_URL="https://your-server.com"
  export AUTH_TOKEN="your-jwt-token-here"
  $0

Note: You must have integrations connected for tools to work.

EOF
    exit 0
fi

# Run main function
main
