#!/bin/bash

# =============================================================================
# E2E OAuth Flow Testing Script
# Tests OAuth flows for all 9 integration providers
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
# Test Functions
# =============================================================================

# Test: List all providers
test_list_providers() {
    print_header "TEST 1: LIST ALL PROVIDERS"

    print_test "GET /api/integrations/providers"

    RESPONSE=$(curl -s "${API_BASE_URL}/api/integrations/providers")

    # Check if response is valid JSON
    if echo "$RESPONSE" | jq empty 2>/dev/null; then
        # Check if response has new format {success, count, providers}
        if echo "$RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
            # New format: {success: true, count: N, providers: [...]}
            PROVIDER_COUNT=$(echo "$RESPONSE" | jq -r '.count')
            print_success "Found $PROVIDER_COUNT providers"

            # List providers from new format
            echo "$RESPONSE" | jq -r '.providers[] | "  - \(.id): \(.name)"'
        else
            # Old format: simple array [...]
            PROVIDER_COUNT=$(echo "$RESPONSE" | jq '. | length')

            if [ "$PROVIDER_COUNT" -eq 9 ]; then
                print_success "Found all 9 providers"
            else
                print_warning "Found $PROVIDER_COUNT providers"
            fi

            # List providers from old format
            echo "$RESPONSE" | jq -r '.[] | "  - \(.id): \(.name)"'
        fi

        print_success "Provider list endpoint works"
    else
        print_error "Invalid JSON response from providers endpoint"
        echo "Response: $RESPONSE"
    fi
}

# Test: OAuth initiation (without actual authorization)
test_oauth_initiation() {
    local PROVIDER=$1
    local PROVIDER_NAME=$2

    print_header "TEST: OAUTH INITIATION - $PROVIDER_NAME"

    print_test "GET /api/integrations/${PROVIDER}/connect"

    RESPONSE=$(curl -s -w "\n%{http_code}" "${API_BASE_URL}/api/integrations/${PROVIDER}/connect")
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "302" ] || [ "$HTTP_CODE" == "200" ]; then
        print_success "OAuth initiation endpoint responds correctly (HTTP $HTTP_CODE)"

        # Try to extract authorization URL
        AUTH_URL=$(echo "$BODY" | jq -r '.authorization_url // empty' 2>/dev/null || echo "")
        if [ -n "$AUTH_URL" ]; then
            print_info "Authorization URL: $AUTH_URL"
        fi
    else
        print_error "OAuth initiation failed (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: List user integrations (requires auth token)
test_list_user_integrations() {
    print_header "TEST: LIST USER INTEGRATIONS (AUTH REQUIRED)"

    if [ -z "$AUTH_TOKEN" ]; then
        print_warning "Skipping - AUTH_TOKEN not provided"
        print_info "Set AUTH_TOKEN environment variable to test authenticated endpoints"
        return
    fi

    print_test "GET /api/integrations/"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        "${API_BASE_URL}/api/integrations/")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ]; then
        print_success "User integrations endpoint works (HTTP 200)"

        # Check if response is valid JSON array
        if echo "$BODY" | jq empty 2>/dev/null; then
            INTEGRATION_COUNT=$(echo "$BODY" | jq '. | length')
            print_info "User has $INTEGRATION_COUNT connected integration(s)"

            if [ "$INTEGRATION_COUNT" -gt 0 ]; then
                echo "$BODY" | jq -r '.[] | "  - \(.provider): connected on \(.created_at)"'
            fi
        fi
    elif [ "$HTTP_CODE" == "401" ]; then
        print_error "Authentication failed (HTTP 401) - check AUTH_TOKEN"
    else
        print_error "User integrations endpoint failed (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Check specific integration status
test_integration_status() {
    local PROVIDER=$1
    local PROVIDER_NAME=$2

    print_header "TEST: INTEGRATION STATUS - $PROVIDER_NAME"

    if [ -z "$AUTH_TOKEN" ]; then
        print_warning "Skipping - AUTH_TOKEN not provided"
        return
    fi

    print_test "GET /api/integrations/${PROVIDER}"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        "${API_BASE_URL}/api/integrations/${PROVIDER}")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ]; then
        print_success "${PROVIDER_NAME} is connected"

        # Show integration details
        echo "$BODY" | jq -r '
            "  Provider: \(.provider)",
            "  Status: \(.status // "active")",
            "  Connected: \(.created_at)",
            "  Last Synced: \(.last_sync_at // "never")"
        ' 2>/dev/null || echo "  $BODY"

    elif [ "$HTTP_CODE" == "404" ]; then
        print_info "${PROVIDER_NAME} is not connected"
    else
        print_error "${PROVIDER_NAME} status check failed (HTTP $HTTP_CODE)"
    fi
}

# Test: Database verification
test_database_verification() {
    print_header "TEST: DATABASE VERIFICATION"

    print_info "This test requires direct database access"
    print_info "Run these queries manually to verify:"

    cat << 'EOF'

-- Check OAuth tokens
SELECT provider, user_id, expires_at, created_at
FROM oauth_tokens
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- Check webhook subscriptions
SELECT provider, resource_type, status, event_count, last_event_at
FROM webhook_subscriptions
WHERE status = 'active'
ORDER BY last_event_at DESC;

-- Check synced data counts
SELECT 'Calendar Events' as entity, COUNT(*) as count FROM synced_calendar_events
UNION ALL
SELECT 'Tasks' as entity, COUNT(*) as count FROM synced_tasks
UNION ALL
SELECT 'Messages' as entity, COUNT(*) as count FROM synced_messages
UNION ALL
SELECT 'Contacts' as entity, COUNT(*) as count FROM synced_contacts
UNION ALL
SELECT 'Meetings' as entity, COUNT(*) as count FROM synced_meetings;

-- Check sync tokens
SELECT provider, resource_type, last_sync_at
FROM sync_tokens
ORDER BY last_sync_at DESC;

EOF
}

# =============================================================================
# Main Test Execution
# =============================================================================

main() {
    print_header "E2E OAUTH FLOW TESTING"
    print_info "Testing BusinessOS Integration System"
    print_info "API Base URL: ${API_BASE_URL}"

    if [ -n "$AUTH_TOKEN" ]; then
        print_info "Auth Token: Provided ✓"
    else
        print_warning "Auth Token: Not provided (some tests will be skipped)"
    fi

    echo ""

    # Prerequisites
    check_prerequisites

    # Test 1: List providers (unauthenticated)
    test_list_providers

    # Test 2: OAuth initiation for each provider
    test_oauth_initiation "google" "Google"
    test_oauth_initiation "slack" "Slack"
    test_oauth_initiation "linear" "Linear"
    test_oauth_initiation "hubspot" "HubSpot"
    test_oauth_initiation "notion" "Notion"
    test_oauth_initiation "clickup" "ClickUp"
    test_oauth_initiation "airtable" "Airtable"
    test_oauth_initiation "fathom" "Fathom"
    test_oauth_initiation "microsoft" "Microsoft"

    # Test 3: List user integrations (authenticated)
    test_list_user_integrations

    # Test 4: Check specific integration status (authenticated)
    if [ -n "$AUTH_TOKEN" ]; then
        test_integration_status "google" "Google"
        test_integration_status "slack" "Slack"
        test_integration_status "linear" "Linear"
    fi

    # Test 5: Database verification instructions
    test_database_verification

    # Print summary
    print_header "TEST SUMMARY"
    echo -e "Total Tests: ${TESTS_RUN}"
    echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
    echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}✓ All tests passed!${NC}\n"
        exit 0
    else
        echo -e "\n${RED}✗ Some tests failed${NC}\n"
        exit 1
    fi
}

# =============================================================================
# Script Entry Point
# =============================================================================

# Check if help is requested
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    cat << EOF
E2E OAuth Flow Testing Script

Usage:
  $0 [OPTIONS]

Options:
  -h, --help     Show this help message

Environment Variables:
  API_BASE_URL   Base URL of the API server (default: http://localhost:8001)
  AUTH_TOKEN     Bearer token for authenticated requests (optional)

Examples:
  # Basic test (unauthenticated endpoints only)
  $0

  # Test with authentication
  export AUTH_TOKEN="your-jwt-token-here"
  $0

  # Test against different server
  export API_BASE_URL="https://your-server.com"
  $0

EOF
    exit 0
fi

# Run main function
main
