#!/bin/bash

# run_integration_tests.sh
# Runs integration tests for the Pure Go Voice Agent with LiveKit

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}  Pure Go Voice Agent${NC}"
echo -e "${GREEN}  Integration Tests${NC}"
echo -e "${GREEN}================================${NC}"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running${NC}"
    echo "Please start Docker and try again"
    exit 1
fi

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Parse command line arguments
SKIP_CLEANUP=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-cleanup)
            SKIP_CLEANUP=true
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --skip-cleanup    Don't stop Docker containers after tests"
            echo "  --verbose, -v     Show verbose test output"
            echo "  --help, -h        Show this help message"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Cleanup function
cleanup() {
    if [ "$SKIP_CLEANUP" = true ]; then
        echo -e "${YELLOW}Skipping cleanup (--skip-cleanup flag set)${NC}"
        echo "To stop containers manually, run:"
        echo "  docker-compose -f docker-compose.test.yml down"
    else
        echo ""
        echo -e "${YELLOW}Cleaning up...${NC}"
        docker-compose -f docker-compose.test.yml down -v
        echo -e "${GREEN}Cleanup complete${NC}"
    fi
}

# Register cleanup on exit
trap cleanup EXIT

# Start LiveKit server
echo -e "${YELLOW}Starting LiveKit server...${NC}"
docker-compose -f docker-compose.test.yml up -d livekit

# Wait for LiveKit to be ready
echo -e "${YELLOW}Waiting for LiveKit to be ready...${NC}"
MAX_ATTEMPTS=30
ATTEMPT=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    if docker-compose -f docker-compose.test.yml exec -T livekit wget --spider -q http://localhost:7881/ 2>/dev/null; then
        echo -e "${GREEN}LiveKit is ready${NC}"
        break
    fi

    ATTEMPT=$((ATTEMPT + 1))
    if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
        echo -e "${RED}Error: LiveKit failed to start${NC}"
        echo "Check logs with: docker-compose -f docker-compose.test.yml logs livekit"
        exit 1
    fi

    echo -n "."
    sleep 1
done

echo ""

# Export environment variables for tests
export INTEGRATION_TEST=true
export LIVEKIT_URL=ws://localhost:7880
export LIVEKIT_API_KEY=test-key
export LIVEKIT_API_SECRET=test-secret

# Show environment
echo -e "${YELLOW}Test environment:${NC}"
echo "  INTEGRATION_TEST: $INTEGRATION_TEST"
echo "  LIVEKIT_URL: $LIVEKIT_URL"
echo "  LIVEKIT_API_KEY: $LIVEKIT_API_KEY"
echo ""

# Run integration tests
echo -e "${YELLOW}Running integration tests...${NC}"
echo ""

if [ "$VERBOSE" = true ]; then
    # Verbose output
    go test -tags=integration -v -timeout=10m ./internal/livekit/... | tee test-results.log
else
    # Normal output
    go test -tags=integration -timeout=10m ./internal/livekit/...
fi

TEST_EXIT_CODE=$?

echo ""

# Report results
if [ $TEST_EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}  ✓ All tests passed${NC}"
    echo -e "${GREEN}================================${NC}"
else
    echo -e "${RED}================================${NC}"
    echo -e "${RED}  ✗ Some tests failed${NC}"
    echo -e "${RED}================================${NC}"
    exit $TEST_EXIT_CODE
fi
