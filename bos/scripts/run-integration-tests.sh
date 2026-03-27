#!/bin/bash

#! BOS CLI Integration Test Runner
#! Comprehensive test suite for all bos commands
#! Tests: FIBO, Healthcare, SPARQL, Data Mesh, Error Handling
#!
#! Usage:
#!   ./scripts/run-integration-tests.sh [all|quick|benchmarks|scenario|fibo|healthcare]
#!
#! Examples:
#!   ./scripts/run-integration-tests.sh all        # Run all 23 tests
#!   ./scripts/run-integration-tests.sh quick      # Run tests only (skip benchmarks)
#!   ./scripts/run-integration-tests.sh benchmarks # Run 3 performance benchmarks
#!   ./scripts/run-integration-tests.sh fibo       # Run 4 FIBO workflow tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test directory
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESULTS_DIR="${TEST_DIR}/test-results"

# Create results directory
mkdir -p "${RESULTS_DIR}"

# Initialize counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0
START_TIME=$(date +%s)

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED_TESTS++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED_TESTS++))
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    ((SKIPPED_TESTS++))
}

header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# Test runner functions
run_test() {
    local test_name="$1"
    local description="$2"

    ((TOTAL_TESTS++))
    log_info "Running: $description"

    if cargo test --test comprehensive_integration_test "$test_name" -- --test-threads=1 --nocapture 2>&1 | tee -a "${RESULTS_DIR}/test.log"; then
        log_pass "$test_name"
    else
        log_fail "$test_name"
    fi
}

# Test categories
run_fibo_tests() {
    header "FIBO WORKFLOW TESTS (4 tests)"

    run_test "test_scenario_fibo_deal_workflow_complete" \
        "FIBO deal creation → compliance → reporting"

    run_test "test_scenario_deal_creation_with_compliance_check" \
        "Deal creation with compliance workflow"

    run_test "test_scenario_compliance_checking_with_audit_trail" \
        "Compliance checking with audit logging"

    run_test "test_scenario_contract_definition_and_validation" \
        "Contract definition and validation"
}

run_healthcare_tests() {
    header "HEALTHCARE TESTS (4 tests)"

    run_test "test_scenario_healthcare_phi_tracking_complete" \
        "Healthcare PHI tracking (HIPAA framework)"

    run_test "test_scenario_phi_lineage_tracking" \
        "PHI lineage and access tracking"

    run_test "test_scenario_healthcare_consent_enforcement" \
        "Healthcare consent enforcement"

    run_test "test_scenario_knowledge_base_indexing" \
        "Knowledge base indexing"
}

run_sparql_tests() {
    header "SPARQL & RDF ROUND-TRIP TESTS (2 tests)"

    run_test "test_scenario_sparql_round_trip_fibo_data" \
        "SQL → RDF → SPARQL SELECT → results"

    run_test "test_scenario_construct_query_generates_rdf" \
        "SPARQL CONSTRUCT query generation"
}

run_cross_command_tests() {
    header "CROSS-COMMAND WORKFLOW TESTS (4 tests)"

    run_test "test_scenario_domain_creation_to_discovery" \
        "Domain schema creation and discovery"

    run_test "test_scenario_workspace_initialization_and_validation" \
        "Workspace initialization and validation"

    run_test "test_scenario_process_discovery_from_event_log" \
        "Process discovery from XES event log"

    run_test "test_scenario_conformance_checking" \
        "Process conformance checking"
}

run_error_tests() {
    header "ERROR HANDLING TESTS (3 tests)"

    run_test "test_error_handling_missing_schema_file" \
        "Error handling for missing files"

    run_test "test_error_handling_invalid_sparql_query" \
        "Error handling for invalid SPARQL"

    run_test "test_edge_case_empty_dataset" \
        "Edge case: empty dataset validation"
}

run_benchmark_tests() {
    header "PERFORMANCE BENCHMARK TESTS (3 tests)"

    run_test "benchmark_schema_validation_speed" \
        "Schema validation performance (<5s)"

    run_test "benchmark_sparql_query_execution" \
        "SPARQL query execution (<1s)"

    run_test "benchmark_ontology_construct_generation" \
        "Ontology CONSTRUCT generation (<2s)"
}

run_additional_tests() {
    header "ADDITIONAL TESTS (3 tests)"

    run_test "test_edge_case_large_uuids" \
        "Edge case: large schema with 100+ UUIDs"

    run_test "test_scenario_decision_record_creation" \
        "Decision record management"
}

# Main execution
main() {
    local test_suite="${1:-all}"

    header "BOS CLI Integration Test Suite"
    log_info "Test suite: $test_suite"
    log_info "Results directory: $RESULTS_DIR"
    log_info "Timestamp: $(date)"
    echo ""

    # Clear previous results
    > "${RESULTS_DIR}/test.log"

    # Run requested tests
    case "$test_suite" in
        all)
            run_fibo_tests
            run_healthcare_tests
            run_sparql_tests
            run_cross_command_tests
            run_error_tests
            run_benchmark_tests
            run_additional_tests
            ;;
        quick)
            run_fibo_tests
            run_healthcare_tests
            run_sparql_tests
            run_cross_command_tests
            run_error_tests
            ;;
        benchmarks)
            run_benchmark_tests
            ;;
        fibo)
            run_fibo_tests
            ;;
        healthcare)
            run_healthcare_tests
            ;;
        sparql)
            run_sparql_tests
            ;;
        scenario)
            run_cross_command_tests
            ;;
        errors)
            run_error_tests
            ;;
        *)
            echo "Unknown test suite: $test_suite"
            echo "Valid options: all, quick, benchmarks, fibo, healthcare, sparql, scenario, errors"
            exit 1
            ;;
    esac

    # Calculate summary
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))

    # Print summary
    header "Test Execution Summary"
    echo -e "Total tests:  ${TOTAL_TESTS}"
    echo -e "${GREEN}Passed:      ${PASSED_TESTS}${NC}"
    echo -e "${RED}Failed:      ${FAILED_TESTS}${NC}"
    echo -e "${YELLOW}Skipped:     ${SKIPPED_TESTS}${NC}"
    echo -e "Duration:    ${DURATION}s"
    echo ""

    # Calculate pass rate
    if [ $TOTAL_TESTS -gt 0 ]; then
        PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
        echo -e "Pass rate:   ${PASS_RATE}%"
    fi

    echo ""
    echo "Log file: ${RESULTS_DIR}/test.log"
    echo ""

    # Exit with appropriate code
    if [ $FAILED_TESTS -eq 0 ] && [ $TOTAL_TESTS -gt 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}Some tests failed.${NC}"
        return 1
    fi
}

# Run with provided argument
main "$@"
