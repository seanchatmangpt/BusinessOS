#!/bin/bash

# =============================================================================
# E2E Webhook Testing Script
# Simulates webhook events from all 9 integration providers
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

# Compute HMAC-SHA256 signature (requires openssl)
compute_hmac() {
    local SECRET=$1
    local DATA=$2
    echo -n "$DATA" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //'
}

# =============================================================================
# Webhook Test Functions
# =============================================================================

# Test: Slack Webhook
test_slack_webhook() {
    print_header "TEST: SLACK WEBHOOK"

    local TIMESTAMP=$(date +%s)
    local PAYLOAD='{"type":"event_callback","team_id":"T123","event":{"type":"message","channel":"C123","user":"U123","text":"Test message","ts":"1234567890.123456"}}'

    # Calculate signature (v0={timestamp}:{body})
    local SIG_BASE="v0:${TIMESTAMP}:${PAYLOAD}"
    local SECRET="${SLACK_WEBHOOK_SECRET:-test-slack-secret}"
    local SIGNATURE="v0=$(compute_hmac "$SECRET" "$SIG_BASE")"

    print_test "Sending Slack webhook event"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-Slack-Request-Timestamp: $TIMESTAMP" \
        -H "X-Slack-Signature: $SIGNATURE" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}/api/webhooks/slack/events")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "Slack webhook accepted (HTTP $HTTP_CODE)"
    else
        print_error "Slack webhook rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Linear Webhook
test_linear_webhook() {
    print_header "TEST: LINEAR WEBHOOK"

    local PAYLOAD='{"action":"create","type":"Issue","organizationId":"test-org","data":{"id":"issue-123","identifier":"LIN-123","title":"Test Issue","state":{"name":"Todo","type":"todo"},"team":{"id":"team-1","name":"Engineering"}}}'

    # Calculate signature
    local SECRET="${LINEAR_WEBHOOK_SECRET:-test-linear-secret}"
    local SIGNATURE=$(compute_hmac "$SECRET" "$PAYLOAD")

    print_test "Sending Linear webhook event"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "Linear-Signature: $SIGNATURE" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}/api/webhooks/linear")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "Linear webhook accepted (HTTP $HTTP_CODE)"
    else
        print_error "Linear webhook rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Google Calendar Webhook
test_google_calendar_webhook() {
    print_header "TEST: GOOGLE CALENDAR WEBHOOK"

    print_test "Sending Google Calendar webhook notification"

    # Google Calendar webhooks are notification pings, not event data
    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "X-Goog-Channel-ID: channel-123" \
        -H "X-Goog-Resource-ID: resource-123" \
        -H "X-Goog-Resource-State: exists" \
        -H "X-Goog-Channel-Token: user-uuid-here" \
        "${API_BASE_URL}/api/webhooks/google/calendar")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "Google Calendar webhook accepted (HTTP $HTTP_CODE)"
    else
        print_error "Google Calendar webhook rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: HubSpot Webhook
test_hubspot_webhook() {
    print_header "TEST: HUBSPOT WEBHOOK"

    local TIMESTAMP=$(date +%s%3N)  # milliseconds
    local METHOD="POST"
    local URI="/api/webhooks/hubspot"
    local PAYLOAD='[{"objectId":123,"subscriptionType":"contact.creation","propertyName":"email","propertyValue":"test@example.com"}]'

    # Calculate signature (method + uri + body + timestamp)
    local SOURCE="${METHOD}${URI}${PAYLOAD}${TIMESTAMP}"
    local SECRET="${HUBSPOT_WEBHOOK_SECRET:-test-hubspot-secret}"
    local SIGNATURE=$(compute_hmac "$SECRET" "$SOURCE")

    print_test "Sending HubSpot webhook event"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-HubSpot-Signature-v3: $SIGNATURE" \
        -H "X-HubSpot-Request-Timestamp: $TIMESTAMP" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}${URI}")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "HubSpot webhook accepted (HTTP $HTTP_CODE)"
    else
        print_error "HubSpot webhook rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Notion Webhook
test_notion_webhook() {
    print_header "TEST: NOTION WEBHOOK"

    local PAYLOAD='{"event":"page.updated","page":{"id":"page-123","title":"Test Page"}}'

    # Calculate signature
    local SECRET="${NOTION_WEBHOOK_SECRET:-test-notion-secret}"
    local SIGNATURE=$(compute_hmac "$SECRET" "$PAYLOAD")

    print_test "Sending Notion webhook event"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "Notion-Signature: $SIGNATURE" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}/api/webhooks/notion")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "Notion webhook accepted (HTTP $HTTP_CODE)"
    else
        print_error "Notion webhook rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Airtable Webhook
test_airtable_webhook() {
    print_header "TEST: AIRTABLE WEBHOOK"

    local TIMESTAMP=$(date +%s)
    local PAYLOAD='{"records":[{"id":"rec123","fields":{"Name":"Test Record"}}]}'

    # Calculate signature (timestamp.body)
    local SOURCE="${TIMESTAMP}.${PAYLOAD}"
    local SECRET="${AIRTABLE_WEBHOOK_SECRET:-test-airtable-secret}"
    local SIGNATURE=$(compute_hmac "$SECRET" "$SOURCE")

    print_test "Sending Airtable webhook event"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-Airtable-Content-MAC: $SIGNATURE" \
        -H "X-Airtable-Timestamp: $TIMESTAMP" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}/api/webhooks/airtable")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "Airtable webhook accepted (HTTP $HTTP_CODE)"
    else
        print_error "Airtable webhook rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Fathom Webhook
test_fathom_webhook() {
    print_header "TEST: FATHOM WEBHOOK"

    local PAYLOAD='{"type":"meeting.completed","meeting":{"id":"meeting-123","title":"Test Meeting"}}'

    # Calculate signature
    local SECRET="${FATHOM_WEBHOOK_SECRET:-test-fathom-secret}"
    local SIGNATURE=$(compute_hmac "$SECRET" "$PAYLOAD")

    print_test "Sending Fathom webhook event"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-Fathom-Signature: $SIGNATURE" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}/api/webhooks/fathom")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "Fathom webhook accepted (HTTP $HTTP_CODE)"
    else
        print_error "Fathom webhook rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: Microsoft Webhook
test_microsoft_webhook() {
    print_header "TEST: MICROSOFT WEBHOOK"

    print_test "Sending Microsoft webhook notification"

    local CLIENT_STATE="${MICROSOFT_CLIENT_STATE:-expected-client-state}"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "ClientState: $CLIENT_STATE" \
        -d '{"value":[{"subscriptionId":"sub-123","resource":"users/me/events"}]}' \
        "${API_BASE_URL}/api/webhooks/microsoft/calendar")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "Microsoft webhook accepted (HTTP $HTTP_CODE)"
    else
        print_error "Microsoft webhook rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# Test: ClickUp Webhook
test_clickup_webhook() {
    print_header "TEST: CLICKUP WEBHOOK"

    local PAYLOAD='{"event":"taskCreated","task_id":"task-123","history_items":[{"field":"status","after":"open"}]}'

    # Calculate signature
    local SECRET="${CLICKUP_WEBHOOK_SECRET:-test-clickup-secret}"
    local SIGNATURE=$(compute_hmac "$SECRET" "$PAYLOAD")

    print_test "Sending ClickUp webhook event"

    RESPONSE=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "X-Signature: $SIGNATURE" \
        -d "$PAYLOAD" \
        "${API_BASE_URL}/api/webhooks/clickup")

    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" == "200" ] || [ "$HTTP_CODE" == "202" ]; then
        print_success "ClickUp webhook accepted (HTTP $HTTP_CODE)"
    else
        print_error "ClickUp webhook rejected (HTTP $HTTP_CODE)"
        echo "Response: $BODY"
    fi
}

# =============================================================================
# Main Test Execution
# =============================================================================

main() {
    print_header "E2E WEBHOOK TESTING"
    print_info "Simulating webhook events from all 9 providers"
    print_info "API Base URL: ${API_BASE_URL}"

    print_warning "Using default test secrets for signature verification"
    print_info "Set PROVIDER_WEBHOOK_SECRET environment variables to use custom secrets"

    echo ""

    # Test all webhooks
    test_slack_webhook
    test_linear_webhook
    test_google_calendar_webhook
    test_hubspot_webhook
    test_notion_webhook
    test_airtable_webhook
    test_fathom_webhook
    test_microsoft_webhook
    test_clickup_webhook

    # Print summary
    print_header "TEST SUMMARY"
    echo -e "Total Tests: ${TESTS_RUN}"
    echo -e "${GREEN}Passed: ${TESTS_PASSED}${NC}"
    echo -e "${RED}Failed: ${TESTS_FAILED}${NC}"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}✓ All webhook tests passed!${NC}\n"
        print_info "Check database to verify events were processed:"
        echo "  SELECT * FROM webhook_subscriptions ORDER BY last_event_at DESC LIMIT 10;"
        exit 0
    else
        echo -e "\n${RED}✗ Some webhook tests failed${NC}\n"
        exit 1
    fi
}

# =============================================================================
# Script Entry Point
# =============================================================================

# Check if help is requested
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    cat << EOF
E2E Webhook Testing Script

Simulates webhook events from all 9 integration providers.

Usage:
  $0 [OPTIONS]

Options:
  -h, --help     Show this help message

Environment Variables:
  API_BASE_URL              Base URL of the API server (default: http://localhost:8001)
  SLACK_WEBHOOK_SECRET      Slack webhook secret (default: test-slack-secret)
  LINEAR_WEBHOOK_SECRET     Linear webhook secret (default: test-linear-secret)
  HUBSPOT_WEBHOOK_SECRET    HubSpot webhook secret (default: test-hubspot-secret)
  NOTION_WEBHOOK_SECRET     Notion webhook secret (default: test-notion-secret)
  AIRTABLE_WEBHOOK_SECRET   Airtable webhook secret (default: test-airtable-secret)
  FATHOM_WEBHOOK_SECRET     Fathom webhook secret (default: test-fathom-secret)
  MICROSOFT_CLIENT_STATE    Microsoft client state (default: expected-client-state)
  CLICKUP_WEBHOOK_SECRET    ClickUp webhook secret (default: test-clickup-secret)

Examples:
  # Basic test with default secrets
  $0

  # Test with custom secrets (production)
  export SLACK_WEBHOOK_SECRET="your-real-slack-secret"
  export LINEAR_WEBHOOK_SECRET="your-real-linear-secret"
  $0

  # Test against different server
  export API_BASE_URL="https://your-server.com"
  $0

Note: This script sends simulated webhook events. Check your database
to verify events were processed correctly.

EOF
    exit 0
fi

# Run main function
main
