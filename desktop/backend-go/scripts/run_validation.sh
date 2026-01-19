#!/bin/bash
# Script to run agent response validation

set -e

echo "=================================================="
echo "  Agent Response Quality Validation"
echo "=================================================="
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo "ERROR: .env file not found"
    echo "Please create a .env file with your configuration"
    exit 1
fi

# Default to standard test suite
TEST_MODE=${1:-"standard"}
OUTPUT_FILE=${2:-""}

case "$TEST_MODE" in
    "standard")
        echo "Running standard test suite..."
        if [ -n "$OUTPUT_FILE" ]; then
            go run ./cmd/validate_agents -output "$OUTPUT_FILE"
        else
            go run ./cmd/validate_agents
        fi
        ;;
    "custom")
        echo "Running custom test suite..."
        if [ -n "$OUTPUT_FILE" ]; then
            go run ./cmd/validate_agents -custom -output "$OUTPUT_FILE"
        else
            go run ./cmd/validate_agents -custom
        fi
        ;;
    "single")
        if [ -z "$OUTPUT_FILE" ]; then
            echo "ERROR: Please specify test case name"
            echo "Usage: $0 single <test_case_name>"
            exit 1
        fi
        echo "Running single test: $OUTPUT_FILE"
        go run ./cmd/validate_agents -test "$OUTPUT_FILE"
        ;;
    "verbose")
        echo "Running with verbose logging..."
        go run ./cmd/validate_agents -verbose
        ;;
    *)
        echo "Usage: $0 [mode] [output_file]"
        echo ""
        echo "Modes:"
        echo "  standard          - Run standard test suite (default)"
        echo "  custom            - Run custom test suite"
        echo "  single <name>     - Run single test case"
        echo "  verbose           - Run with verbose logging"
        echo ""
        echo "Examples:"
        echo "  $0                                  # Run standard tests"
        echo "  $0 standard report.json             # Save to file"
        echo "  $0 custom                           # Run custom tests"
        echo "  $0 single simple_greeting           # Run one test"
        echo "  $0 verbose                          # Verbose output"
        exit 1
        ;;
esac

echo ""
echo "=================================================="
echo "  Validation Complete"
echo "=================================================="
