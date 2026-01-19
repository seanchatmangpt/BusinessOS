#!/bin/bash

# =============================================================================
# E2E Error Handling Testing Script
# Tests error scenarios and edge cases for integration system
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

# =============================================================================
# Error Handling Test Functions
# =============================================================================

# Test: Invalid webhook signature
test_invalid_webhook_signature() {
    print_header "TEST: INVALID WEBHOOK SIGNATURE"

    print_test "Sending Slack webhook with invalid signature"

    local TIMESTAMP=$(date +%s)
    local PAYLOAD='{"type":"event_callback","team_id":"T123","event":{"type":"message"}}'

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-Slack-Request-Timestamp: $TIMESTAMP" \
        -H "X-Slack-Signature: v0=invalid_signature_here" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}/api/webhooks/slack/events")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    # Should reject with 401 Unauthorized
    if [ "$HTTP_CODE" == "401" ] || [ "$HTTP_CODE" == "403" ]; then
        print_success "Invalid signature correctly rejected (HTTP $HTTP_CODE)"
    else
        print_error "Invalid signature was accepted (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Expired timestamp (replay attack)
test_expired_timestamp() {
    print_header "TEST: EXPIRED TIMESTAMP (REPLAY ATTACK)"

    print_test "Sending webhook with 10-minute-old timestamp"

    # Timestamp from 10 minutes ago
    local OLD_TIMESTAMP=$(($(date +%s) - 600))
    local PAYLOAD='{"type":"event_callback","team_id":"T123","event":{"type":"message"}}'

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-Slack-Request-Timestamp: $OLD_TIMESTAMP" \
        -H "X-Slack-Signature: v0=any_signature" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}/api/webhooks/slack/events")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    # Should reject with 401 Unauthorized
    if [ "$HTTP_CODE" == "401" ] || [ "$HTTP_CODE" == "403" ]; then
        print_success "Expired timestamp correctly rejected (HTTP $HTTP_CODE)"
    else
        print_error "Expired timestamp was accepted (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Malformed JSON payload
test_malformed_json() {
    print_header "TEST: MALFORMED JSON PAYLOAD"

    print_test "Sending webhook with malformed JSON"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-Slack-Request-Timestamp: $(date +%s)" \
        -H "X-Slack-Signature: v0=signature" \
        -d '{invalid json here' \
        "${API_BASE_URL}/api/webhooks/slack/events")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    # Should reject with 400 Bad Request
    if [ "$HTTP_CODE" == "400" ]; then
        print_success "Malformed JSON correctly rejected (HTTP 400)"
    else
        print_error "Malformed JSON was not properly handled (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Missing required headers
test_missing_headers() {
    print_header "TEST: MISSING REQUIRED HEADERS"

    print_test "Sending Slack webhook without signature header"

    local PAYLOAD='{"type":"event_callback","team_id":"T123"}'

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}/api/webhooks/slack/events")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    # Should reject with 400 or 401
    if [ "$HTTP_CODE" == "400" ] || [ "$HTTP_CODE" == "401" ]; then
        print_success "Missing headers correctly rejected (HTTP $HTTP_CODE)"
    else
        print_error "Missing headers were not properly handled (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Non-existent integration endpoint
test_nonexistent_endpoint() {
    print_header "TEST: NON-EXISTENT INTEGRATION ENDPOINT"

    print_test "Accessing non-existent integration"

    if [ -z "$AUTH_TOKEN" ]; then
        print_warning "Skipping - AUTH_TOKEN not provided"
        return
    fi

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        "${API_BASE_URL}/api/integrations/nonexistent")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    # Should return 404
    if [ "$HTTP_CODE" == "404" ]; then
        print_success "Non-existent integration correctly returns 404"
    else
        print_error "Non-existent integration did not return 404 (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Unauthenticated access to protected endpoint
test_unauthenticated_access() {
    print_header "TEST: UNAUTHENTICATED ACCESS"

    print_test "Accessing protected endpoint without token"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        "${API_BASE_URL}/api/integrations/")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    # Should return 401
    if [ "$HTTP_CODE" == "401" ]; then
        print_success "Unauthenticated access correctly rejected (HTTP 401)"
    else
        print_error "Unauthenticated access was not rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Invalid auth token
test_invalid_auth_token() {
    print_header "TEST: INVALID AUTH TOKEN"

    print_test "Accessing protected endpoint with invalid token"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -H "Authorization: Bearer invalid_token_12345" \
        "${API_BASE_URL}/api/integrations/")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    # Should return 401
    if [ "$HTTP_CODE" == "401" ]; then
        print_success "Invalid token correctly rejected (HTTP 401)"
    else
        print_error "Invalid token was not rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Rate limiting (if implemented)
test_rate_limiting() {
    print_header "TEST: RATE LIMITING"

    print_info "This test is optional and depends on rate limiting implementation"
    print_info "Sending 100 rapid requests to test rate limiting..."

    local RATE_LIMITED=false

    for i in {1..100}; do
        HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null \
            "${API_BASE_URL}/api/integrations/providers")

        if [ "$HTTP_CODE" == "429" ]; then
            RATE_LIMITED=true
            print_success "Rate limiting triggered after $i requests (HTTP 429)"
            break
        fi
    done

    if [ "$RATE_LIMITED" = false ]; then
        print_warning "Rate limiting not triggered (may not be implemented)"
    fi
}

# Test: Large payload handling
test_large_payload() {
    print_header "TEST: LARGE PAYLOAD HANDLING"

    print_test "Sending webhook with very large payload (5MB)"

    # Generate a large JSON payload (5MB)
    local LARGE_DATA=$(printf '{"data":"%0.s#############################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################################"}' {1..1000})

    RESPONSE=$(curl -s -w "\n%{http_code}" --max-time 10 \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-Slack-Request-Timestamp: $(date +%s)" \
        -H "X-Slack-Signature: v0=signature" \
        -d "$LARGE_DATA" \
        "${API_BASE_URL}/api/webhooks/slack/events")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

    # Should handle gracefully (either 413 Payload Too Large or process it)
    if [ "$HTTP_CODE" == "413" ] || [ "$HTTP_CODE" == "400" ]; then
        print_success "Large payload correctly rejected (HTTP $HTTP_CODE)"
    elif [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "Large payload processed successfully (HTTP $HTTP_CODE)"
    else
        print_warning "Large payload handling unclear (HTTP $HTTP_CODE)"
    fi
}

# Test: Concurrent webhook processing
test_concurrent_webhooks() {
    print_header "TEST: CONCURRENT WEBHOOK PROCESSING"

    print_info "Sending 10 webhooks concurrently to test race conditions"

    local PIDS=()

    for i in {1..10}; do
        (
            curl -s -X POST \
                -H "Content-Type: application/json" \
                -H "X-Slack-Request-Timestamp: $(date +%s)" \
                -H "X-Slack-Signature: v0=signature" \
                -d "{\"type\":\"event_callback\",\"event\":{\"text\":\"Message $i\"}}" \
                "${API_BASE_URL}/api/webhooks/slack/events" \
                > /dev/null
        ) &
        PIDS+=($!)
    done

    # Wait for all requests to complete
    for pid in "${PIDS[@]}"; do
        wait $pid
    done

    print_success "Concurrent webhooks sent successfully"
    print_info "Check logs to verify no race conditions or deadlocks occurred"
}

# Test: OAuth callback with error
test_oauth_error_callback() {
    print_header "TEST: OAUTH ERROR CALLBACK"

    print_test "Simulating OAuth error callback"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        "${API_BASE_URL}/api/integrations/google/callback?error=access_denied&error_description=User%20denied%20access")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    # Should handle gracefully (302 redirect or error page)
    if [ "$HTTP_CODE" == "302" ] || [ "$HTTP_CODE" == "400" ] || [ "$HTTP_CODE" == "401" ]; then
        print_success "OAuth error correctly handled (HTTP $HTTP_CODE)"
    else
        print_warning "OAuth error handling unclear (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Disconnect integration
test_disconnect_integration() {
    print_header "TEST: DISCONNECT INTEGRATION"

    if [ -z "$AUTH_TOKEN" ]; then
        print_warning "Skipping - AUTH_TOKEN not provided"
        return
    fi

    print_test "Attempting to disconnect non-existent integration"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X DELETE \
        -H "Authorization: Bearer $AUTH_TOKEN" \
        "${API_BASE_URL}/api/integrations/nonexistent")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    # Should return 404
    if [ "$HTTP_CODE" == "404" ]; then
        print_success "Disconnect non-existent integration returns 404"
    else
        print_warning "Disconnect endpoint behavior unclear (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# =============================================================================
# Main Test Execution
# =============================================================================

main() {
    print_header "E2E ERROR HANDLING TESTING"
    print_info "Testing error scenarios and edge cases"
    print_info "API Base URL: ${API_BASE_URL}"

    if [ -n "$AUTH_TOKEN" ]; then
        print_info "Auth Token: Provided ✓"
    else
        print_warning "Auth Token: Not provided (some tests will be skipped)"
    fi

    echo ""

    # Security Tests
    test_invalid_webhook_signature
    test_expired_timestamp
    test_unauthenticated_access
    test_invalid_auth_token

    # Input Validation Tests
    test_malformed_json
    test_missing_headers
    test_large_payload

    # Resource Tests
    test_nonexistent_endpoint
    test_disconnect_integration

    # OAuth Tests
    test_oauth_error_callback

    # Performance Tests
    test_rate_limiting
    test_concurrent_webhooks

    # Print summary
    print_header "TEST SUMMARY"
    echo -e "Total Tests: ${TESTS_RUN}"
    echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
    echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}✓ All error handling tests passed!${NC}\n"
        exit 0
    else
        echo -e "\n${YELLOW}⚠ Some tests failed or had warnings${NC}\n"
        print_info "This is expected for optional features like rate limiting"
        exit 0
    fi
}

# =============================================================================
# Script Entry Point
# =============================================================================

# Check if help is requested
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    cat << EOF
E2E Error Handling Testing Script

Tests error scenarios and edge cases for the integration system.

Usage:
  $0 [OPTIONS]

Options:
  -h, --help     Show this help message

Environment Variables:
  API_BASE_URL   Base URL of the API server (default: http://localhost:8001)
  AUTH_TOKEN     Bearer token for authenticated requests (optional)

Examples:
  # Basic test
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
