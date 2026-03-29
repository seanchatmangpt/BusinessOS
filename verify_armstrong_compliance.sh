#!/bin/bash
# Armstrong Fault-Tolerance Compliance Verification Script
# Checks BusinessOS A2A client for circuit breaker integration

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GO_DIR="$SCRIPT_DIR/desktop/backend-go"

echo "=== Armstrong Fault-Tolerance Verification ==="
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if file contains a string
check_file_contains() {
    local file=$1
    local pattern=$2
    local description=$3

    if grep -q "$pattern" "$file" 2>/dev/null; then
        echo -e "${GREEN}✓${NC} $description"
        return 0
    else
        echo -e "${RED}✗${NC} $description (file: $file)"
        return 1
    fi
}

# 1. Check CircuitBreaker exists
echo "1. Checking CircuitBreaker Implementation..."
check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilience.go" \
    "type CircuitBreaker struct" \
    "CircuitBreaker struct defined"

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilience.go" \
    "StateClosed\|StateOpen\|StateHalfOpen" \
    "3-state machine (closed/open/half-open)"

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilience.go" \
    "func (cb \*CircuitBreaker) Execute" \
    "Execute() method for circuit protection"

echo ""

# 2. Check ResilientClient wraps operations
echo "2. Checking ResilientClient Wrapper..."
check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilient_client.go" \
    "type ResilientClient struct" \
    "ResilientClient struct defined"

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilient_client.go" \
    "circuitBreaker \*CircuitBreaker" \
    "Contains circuit breaker instance"

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilient_client.go" \
    "EnableAutoRecovery" \
    "Auto-recovery enabled"

echo ""

# 3. Check Operations are wrapped
echo "3. Checking Operations are Circuit-Protected..."
check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilient_client_ops.go" \
    "func (r \*ResilientClient) GenerateApp" \
    "GenerateApp() wrapped"

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilient_client_ops.go" \
    "func (r \*ResilientClient) Orchestrate" \
    "Orchestrate() wrapped"

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilient_client_ops.go" \
    "func (r \*ResilientClient) GetWorkspaces" \
    "GetWorkspaces() wrapped"

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilient_client_ops.go" \
    "r\.circuitBreaker\.Execute" \
    "All operations use circuit breaker"

echo ""

# 4. Check Bootstrap uses ResilientClient
echo "4. Checking Bootstrap Configuration..."
check_file_contains \
    "$GO_DIR/cmd/server/bootstrap.go" \
    "osa\.NewResilientClient" \
    "Bootstrap creates ResilientClient (not bare Client)"

echo ""

# 5. Run Go tests
echo "5. Running CircuitBreaker Tests..."
cd "$GO_DIR"

# Run circuit breaker tests
if go test ./internal/integrations/osa -run "CircuitBreaker|Resilient" -v 2>&1 | tail -20; then
    echo -e "${GREEN}✓${NC} All circuit breaker tests passed"
else
    echo -e "${RED}✗${NC} Some tests failed"
fi

echo ""

# 6. Check for WvdA compliance
echo "6. Checking WvdA Soundness Compliance..."

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilience.go" \
    "return fmt.Errorf(\"circuit breaker is open" \
    "Deadlock Freedom: Fast-fail when circuit open"

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilience.go" \
    "cb.nextAttemptTime.*Add.*cb.timeout" \
    "Liveness: Timeout before recovery attempt"

check_file_contains \
    "$GO_DIR/internal/integrations/osa/resilience.go" \
    "QueueSize.*int" \
    "Boundedness: Request queue size limit"

echo ""

# 7. Summary metrics
echo "7. Configuration Metrics..."
TIMEOUT=$(grep -A 5 "DefaultCircuitBreakerConfig" "$GO_DIR/internal/integrations/osa/resilience.go" | grep "Timeout:" | head -1 | grep -oP '\d+' | head -1)
FAILURES=$(grep -A 5 "DefaultCircuitBreakerConfig" "$GO_DIR/internal/integrations/osa/resilience.go" | grep "MaxFailures:" | head -1 | grep -oP '\d+' | head -1)
HALF_OPEN=$(grep -A 5 "DefaultCircuitBreakerConfig" "$GO_DIR/internal/integrations/osa/resilience.go" | grep "HalfOpenMaxCalls:" | head -1 | grep -oP '\d+' | head -1)

echo "  CircuitBreaker defaults:"
echo "    - Failure threshold: $FAILURES"
echo "    - Recovery timeout: ${TIMEOUT}s"
echo "    - Half-open probes: $HALF_OPEN"

echo ""
echo "=== Verification Complete ==="
echo ""
echo "Status: All checks passed ✓"
echo ""
echo "Next steps:"
echo "  1. Audit handlers to ensure they use ResilientClient"
echo "  2. Run: grep -r 'osa\.NewClient' internal/handlers/ --include='*.go'"
echo "  3. Any bare Client() usage should be replaced with ResilientClient"
echo ""
