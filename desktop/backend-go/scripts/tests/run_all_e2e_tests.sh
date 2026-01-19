#!/bin/bash

# =============================================================================
# Master E2E Test Runner
# Runs all E2E test suites and generates comprehensive report
# =============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="${API_BASE_URL:-http://localhost:8001}"
AUTH_TOKEN="${AUTH_TOKEN:-}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Test suite results (bash 3.2 compatible - using parallel arrays)
SUITE_NAMES=()
SUITE_STATUSES=()
TOTAL_SUITES=0
PASSED_SUITES=0
FAILED_SUITES=0
SKIPPED_SUITES=0

# =============================================================================
# Helper Functions
# =============================================================================

# Add test result (bash 3.2 compatible)
add_suite_result() {
    local SUITE_NAME=$1
    local STATUS=$2
    SUITE_NAMES+=("$SUITE_NAME")
    SUITE_STATUSES+=("$STATUS")
}

print_banner() {
    echo -e "${CYAN}"
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════════════╗
║                                                                      ║
║   ██████╗ ██╗   ██╗███████╗██╗███╗   ██╗███████╗███████╗███████╗   ║
║   ██╔══██╗██║   ██║██╔════╝██║████╗  ██║██╔════╝██╔════╝██╔════╝   ║
║   ██████╔╝██║   ██║███████╗██║██╔██╗ ██║█████╗  ███████╗███████╗   ║
║   ██╔══██╗██║   ██║╚════██║██║██║╚██╗██║██╔══╝  ╚════██║╚════██║   ║
║   ██████╔╝╚██████╔╝███████║██║██║ ╚████║███████╗███████║███████║   ║
║   ╚═════╝  ╚═════╝ ╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝╚══════╝   ║
║                                                                      ║
║              E2E INTEGRATION TESTING SUITE                           ║
║                                                                      ║
╚══════════════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}\n"
}

print_header() {
    echo -e "\n${BOLD}${BLUE}═══════════════════════════════════════════════════════════════════${NC}"
    echo -e "${BOLD}${BLUE}  $1${NC}"
    echo -e "${BOLD}${BLUE}═══════════════════════════════════════════════════════════════════${NC}\n"
}

print_suite_header() {
    echo -e "\n${BOLD}${CYAN}┌─────────────────────────────────────────────────────────────────┐${NC}"
    echo -e "${BOLD}${CYAN}│  SUITE $1: $2${NC}"
    echo -e "${BOLD}${CYAN}└─────────────────────────────────────────────────────────────────┘${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# Run a test suite
run_test_suite() {
    local SUITE_NAME=$1
    local SUITE_DESCRIPTION=$2
    local SCRIPT_PATH=$3
    local REQUIRED_AUTH=${4:-false}

    TOTAL_SUITES=$((TOTAL_SUITES + 1))

    print_suite_header "$TOTAL_SUITES" "$SUITE_DESCRIPTION"

    # Check if script exists
    if [ ! -f "$SCRIPT_PATH" ]; then
        print_error "Test script not found: $SCRIPT_PATH"
        add_suite_result "$SUITE_NAME" "FAILED"
        FAILED_SUITES=$((FAILED_SUITES + 1))
        return 1
    fi

    # Check if auth token is required but not provided
    if [ "$REQUIRED_AUTH" == "true" ] && [ -z "$AUTH_TOKEN" ]; then
        print_warning "Skipping - AUTH_TOKEN required but not provided"
        add_suite_result "$SUITE_NAME" "SKIPPED"
        SKIPPED_SUITES=$((SKIPPED_SUITES + 1))
        return 0
    fi

    # Run the test suite
    print_info "Running: $SCRIPT_PATH"
    echo ""

    if bash "$SCRIPT_PATH"; then
        add_suite_result "$SUITE_NAME" "PASSED"
        PASSED_SUITES=$((PASSED_SUITES + 1))
        print_success "Suite completed successfully"
    else
        EXIT_CODE=$?
        if [ $EXIT_CODE -eq 0 ]; then
            add_suite_result "$SUITE_NAME" "PASSED"
            PASSED_SUITES=$((PASSED_SUITES + 1))
        else
            add_suite_result "$SUITE_NAME" "FAILED"
            FAILED_SUITES=$((FAILED_SUITES + 1))
            print_error "Suite failed with exit code $EXIT_CODE"
        fi
    fi
}

# Check prerequisites
check_prerequisites() {
    print_header "CHECKING PREREQUISITES"

    local ALL_GOOD=true

    # Check if server is running
    print_info "Checking server availability at $API_BASE_URL..."
    if curl -s -f "${API_BASE_URL}/health" > /dev/null 2>&1; then
        print_success "Server is running"
    else
        print_error "Server is not running at ${API_BASE_URL}"
        ALL_GOOD=false
    fi

    # Check if jq is installed
    print_info "Checking for jq..."
    if command -v jq &> /dev/null; then
        print_success "jq is installed"
    else
        print_error "jq is not installed (required for JSON parsing)"
        echo "Install with: brew install jq (macOS) or apt-get install jq (Linux)"
        ALL_GOOD=false
    fi

    # Check if openssl is available
    print_info "Checking for openssl..."
    if command -v openssl &> /dev/null; then
        print_success "openssl is available"
    else
        print_warning "openssl not found (webhook signature tests may fail)"
    fi

    # Check auth token
    if [ -z "$AUTH_TOKEN" ]; then
        print_warning "AUTH_TOKEN not provided (authenticated tests will be skipped)"
        print_info "Set AUTH_TOKEN environment variable to run all tests"
    else
        print_success "AUTH_TOKEN provided"
    fi

    echo ""

    if [ "$ALL_GOOD" = false ]; then
        print_error "Prerequisites check failed"
        exit 1
    fi

    print_success "All prerequisites satisfied"
}

# Print final summary
print_summary() {
    print_header "FINAL TEST SUMMARY"

    echo -e "${BOLD}Test Suite Results:${NC}\n"

    # Print individual suite results (bash 3.2 compatible)
    for i in "${!SUITE_NAMES[@]}"; do
        SUITE="${SUITE_NAMES[$i]}"
        RESULT="${SUITE_STATUSES[$i]}"
        case "$RESULT" in
            "PASSED")
                echo -e "  ${GREEN}✓ PASSED${NC}  $SUITE"
                ;;
            "FAILED")
                echo -e "  ${RED}✗ FAILED${NC}  $SUITE"
                ;;
            "SKIPPED")
                echo -e "  ${YELLOW}⊘ SKIPPED${NC} $SUITE"
                ;;
        esac
    done

    echo ""
    echo -e "${BOLD}Overall Statistics:${NC}"
    echo -e "  Total Suites:   $TOTAL_SUITES"
    echo -e "  ${GREEN}Passed:         $PASSED_SUITES${NC}"
    echo -e "  ${RED}Failed:         $FAILED_SUITES${NC}"
    echo -e "  ${YELLOW}Skipped:        $SKIPPED_SUITES${NC}"

    # Calculate percentage
    if [ $TOTAL_SUITES -gt 0 ]; then
        local RUN_SUITES=$((TOTAL_SUITES - SKIPPED_SUITES))
        if [ $RUN_SUITES -gt 0 ]; then
            local PASS_PERCENT=$((PASSED_SUITES * 100 / RUN_SUITES))
            echo -e "  ${BOLD}Pass Rate:      ${PASS_PERCENT}%${NC}"
        fi
    fi

    echo ""

    # Final verdict
    if [ $FAILED_SUITES -eq 0 ]; then
        echo -e "${GREEN}${BOLD}╔═══════════════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}${BOLD}║                                                   ║${NC}"
        echo -e "${GREEN}${BOLD}║   ✓ ALL TESTS PASSED!                            ║${NC}"
        echo -e "${GREEN}${BOLD}║                                                   ║${NC}"
        echo -e "${GREEN}${BOLD}╚═══════════════════════════════════════════════════╝${NC}"
        echo ""
        exit 0
    else
        echo -e "${RED}${BOLD}╔═══════════════════════════════════════════════════╗${NC}"
        echo -e "${RED}${BOLD}║                                                   ║${NC}"
        echo -e "${RED}${BOLD}║   ✗ SOME TESTS FAILED                            ║${NC}"
        echo -e "${RED}${BOLD}║                                                   ║${NC}"
        echo -e "${RED}${BOLD}╚═══════════════════════════════════════════════════╝${NC}"
        echo ""
        exit 1
    fi
}

# =============================================================================
# Main Test Execution
# =============================================================================

main() {
    print_banner

    print_info "BusinessOS Integration System E2E Tests"
    print_info "API Base URL: ${API_BASE_URL}"
    print_info "Test Scripts Directory: ${SCRIPT_DIR}"
    echo ""

    # Step 1: Check prerequisites
    check_prerequisites

    # Step 2: Run unit tests (signature verification)
    print_header "PHASE 1: UNIT TESTS"
    print_info "Running Go unit tests for webhook signature verification..."
    echo ""

    cd "$(dirname "$SCRIPT_DIR")/../.."  # Go to project root

    if go test -v ./internal/webhooks; then
        print_success "Unit tests passed"
        add_suite_result "unit_tests" "PASSED"
        PASSED_SUITES=$((PASSED_SUITES + 1))
    else
        print_error "Unit tests failed"
        add_suite_result "unit_tests" "FAILED"
        FAILED_SUITES=$((FAILED_SUITES + 1))
    fi

    TOTAL_SUITES=$((TOTAL_SUITES + 1))

    cd "$SCRIPT_DIR"

    # Step 3: Run E2E test suites
    print_header "PHASE 2: E2E INTEGRATION TESTS"

    run_test_suite \
        "oauth_flows" \
        "OAuth Flow Testing" \
        "${SCRIPT_DIR}/e2e_oauth_test.sh" \
        false

    run_test_suite \
        "mcp_tools" \
        "MCP Tools Testing" \
        "${SCRIPT_DIR}/e2e_mcp_tools_test.sh" \
        true

    run_test_suite \
        "webhooks" \
        "Webhook Event Simulation" \
        "${SCRIPT_DIR}/e2e_webhook_test.sh" \
        false

    run_test_suite \
        "error_handling" \
        "Error Handling & Edge Cases" \
        "${SCRIPT_DIR}/e2e_error_handling_test.sh" \
        false

    # Step 4: Print summary
    print_summary
}

# =============================================================================
# Script Entry Point
# =============================================================================

# Check if help is requested
if [ "$1" == "--help" ] || [ "$1" == "-h" ]; then
    cat << EOF
BusinessOS E2E Test Suite Runner

Runs all E2E test suites for the integration system.

Usage:
  $0 [OPTIONS]

Options:
  -h, --help     Show this help message

Environment Variables:
  API_BASE_URL   Base URL of the API server (default: http://localhost:8001)
  AUTH_TOKEN     Bearer token for authenticated requests (optional)

Test Suites:
  1. Unit Tests              - Webhook signature verification (Go tests)
  2. OAuth Flow Testing      - OAuth initiation and callbacks
  3. MCP Tools Testing       - MCP tool execution (requires auth)
  4. Webhook Testing         - Simulated webhook events
  5. Error Handling Testing  - Edge cases and error scenarios

Examples:
  # Run all tests (unauthenticated only)
  $0

  # Run all tests including authenticated ones
  export AUTH_TOKEN="your-jwt-token-here"
  $0

  # Test against different server
  export API_BASE_URL="https://your-server.com"
  export AUTH_TOKEN="your-jwt-token-here"
  $0

Output:
  - Colored terminal output with test results
  - Exit code 0 if all tests pass
  - Exit code 1 if any tests fail

EOF
    exit 0
fi

# Trap errors and cleanup
trap 'echo -e "\n${RED}Test suite interrupted${NC}\n"; exit 130' INT TERM

# Run main function
main
