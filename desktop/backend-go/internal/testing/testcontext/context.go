package testcontext

import (
	"fmt"
)

// Context provides debugging information for test failures.
type Context struct {
	Category   string
	Details    map[string]interface{}
	ErrorMsg   string
	DebugSteps []string
}

// APIFailure creates context for API/network failures.
func APIFailure(endpoint string, expectedStatus, actualStatus int, reason string) *Context {
	errorMsg := fmt.Sprintf(`
API Request Failed

Endpoint: %s
Expected Status: %d
Actual Status: %d
Reason: %s

Debugging Steps:
  1. Is localhost:8001 running? Run: curl http://localhost:8001/api/health
  2. Check endpoint exists: curl -v http://localhost:8001%s
  3. Check BusinessOS logs: docker logs businessos-backend
  4. Verify response format: curl http://localhost:8001%s | jq .
`, endpoint, expectedStatus, actualStatus, reason, endpoint, endpoint)

	return &Context{
		Category: "api_failure",
		Details: map[string]interface{}{
			"endpoint":        endpoint,
			"expected_status": expectedStatus,
			"actual_status":   actualStatus,
			"reason":          reason,
		},
		ErrorMsg: errorMsg,
		DebugSteps: []string{
			"Verify service is running",
			"Check endpoint URL",
			"Check response status",
			"Check response format",
			"Check logs",
		},
	}
}

// TimingFailure creates context for timing/race condition failures.
func TimingFailure(operation string, timeoutMs int, expectedValue interface{}) *Context {
	errorMsg := fmt.Sprintf(`
Timing/Race Condition Detected

Operation: %s
Timeout: %dms
Expected: %+v

Debugging Steps:
  1. Run test in isolation: go test ./... -run TestName -v
  2. Run serially (one test at a time): go test -p 1 ./...
  3. Add t.Logf at each step to trace execution
  4. Check for missing synchronization (channels, WaitGroup)
  5. Verify all goroutines complete before assertion

If test passes in isolation but fails in parallel:
  → It's a timing issue, not logic
  → Check for race conditions: go test -race ./...
  → Add proper synchronization (channels, mutex, WaitGroup)
`, operation, timeoutMs, expectedValue)

	return &Context{
		Category: "timing_failure",
		Details: map[string]interface{}{
			"operation":      operation,
			"timeout_ms":     timeoutMs,
			"expected_value": expectedValue,
		},
		ErrorMsg: errorMsg,
		DebugSteps: []string{
			"Run test in isolation",
			"Run tests serially",
			"Check for race conditions",
			"Add explicit synchronization",
		},
	}
}

// LogicFailure creates context for logic/assertion failures.
func LogicFailure(operation string, expected, actual interface{}) *Context {
	errorMsg := fmt.Sprintf(`
Logic/Assertion Failure

Operation: %s
Expected: %+v
Actual: %+v

Debugging Steps:
  1. Add t.Logf at each step to trace execution
  2. Check for off-by-one errors
  3. Verify data types match (string vs int, interface{} vs concrete type)
  4. Check for nil values where non-nil expected
  5. Use debugger: dlv test ./path/to/package -test.run TestName

Try this:
  # In terminal:
  dlv test ./internal/package
  (dlv) break TestName
  (dlv) continue
  (dlv) next
  # Step through execution
`, operation, expected, actual)

	return &Context{
		Category: "logic_failure",
		Details: map[string]interface{}{
			"operation": operation,
			"expected":  expected,
			"actual":    actual,
		},
		ErrorMsg: errorMsg,
		DebugSteps: []string{
			"Add debug output at each step",
			"Check data types",
			"Verify expected vs actual",
			"Trace execution path",
		},
	}
}

// ResourceFailure creates context for resource exhaustion failures.
func ResourceFailure(resourceType string, limit, actual interface{}) *Context {
	errorMsg := fmt.Sprintf(`
Resource Exhaustion Detected

Resource: %s
Limit: %+v
Actual: %+v

Debugging Steps:
  1. Check goroutine count: dlv debug ./cmd/server
     (dlv) runtime.NumGoroutine()
  2. Monitor during test: watch -n 1 'ps aux | grep server'
  3. Check for unclosed resources (connections, files)
  4. Verify resource cleanup in defer statements
  5. Use pprof for memory/CPU analysis:
     go test -memprofile=mem.prof -cpuprofile=cpu.prof ./...
     go tool pprof -http=:8080 mem.prof

Common causes:
  • Goroutines not cleaned up (missing close/defer)
  • Connections not closed (file descriptor leak)
  • Memory not released (large allocations not freed)
  • Channels not drained (goroutines blocked)
  • Unbounded slice growth (no max size)
`, resourceType, limit, actual)

	return &Context{
		Category: "resource_failure",
		Details: map[string]interface{}{
			"resource_type": resourceType,
			"limit":         limit,
			"actual":        actual,
		},
		ErrorMsg: errorMsg,
		DebugSteps: []string{
			"Check goroutine count",
			"Monitor resource usage",
			"Verify cleanup code",
			"Profile memory/CPU",
		},
	}
}

// FlakyTest creates context for flaky test failures.
func FlakyTest(testName string, passRate string, failurePattern string) *Context {
	errorMsg := fmt.Sprintf(`
Flaky Test Detected

Test: %s
Pass Rate: %s
Failure Pattern: %s

Debugging Steps:
  1. Run test multiple times:
     for i in {1..20}; do go test ./... -run %s || break; done
  2. Check for race conditions: go test -race ./...
  3. Run serially: go test -p 1 ./...
  4. Check for shared state between tests
  5. Look for timing assumptions

If fails only in parallel:
  → Missing synchronization (channels, WaitGroup, mutex)
  → Or check for test isolation issues

If fails unpredictably:
  → Likely timing or external service issue
  → Mock external services with fixed responses
  → Use fake time/clock instead of real time
`, testName, passRate, failurePattern, testName)

	return &Context{
		Category: "flaky_test",
		Details: map[string]interface{}{
			"test_name":       testName,
			"pass_rate":       passRate,
			"failure_pattern": failurePattern,
		},
		ErrorMsg: errorMsg,
		DebugSteps: []string{
			"Run test multiple times",
			"Check for race conditions",
			"Verify synchronization",
			"Mock external dependencies",
		},
	}
}

// ErrorMessage returns the formatted error message.
func (c *Context) ErrorMessage() string {
	return c.ErrorMsg
}

// DebugStepsStr returns debug steps as a formatted string.
func (c *Context) DebugStepsStr() string {
	steps := ""
	for i, step := range c.DebugSteps {
		steps += fmt.Sprintf("  %d. %s\n", i+1, step)
	}
	return steps
}
