package testing

import (
	"testing"
)

// AssertEqual checks equality with descriptive failure message.
// Reduces "got Y but expected X" confusion by providing context.
func AssertEqual(t *testing.T, expected, actual interface{}, context string) {
	t.Helper()
	if expected != actual {
		t.Errorf(
			"Assertion failed: %s\n  Expected: %v\n  Got: %v\n  Context: %s",
			context, expected, actual, context,
		)
	}
}

// AssertNil checks if value is nil with context.
func AssertNil(t *testing.T, value interface{}, context string) {
	t.Helper()
	if value != nil {
		t.Errorf(
			"Assertion failed: expected nil but got %v\n  Context: %s",
			value, context,
		)
	}
}

// AssertNotNil checks if value is not nil with context.
func AssertNotNil(t *testing.T, value interface{}, context string) {
	t.Helper()
	if value == nil {
		t.Errorf(
			"Assertion failed: expected non-nil value\n  Context: %s",
			context,
		)
	}
}

// AssertError checks if error is not nil and message matches pattern.
func AssertError(t *testing.T, err error, expectedPattern string) {
	t.Helper()
	if err == nil {
		t.Errorf(
			"Assertion failed: expected error matching %q but got nil",
			expectedPattern,
		)
		return
	}
	// Simple substring check
	if !contains(err.Error(), expectedPattern) {
		t.Errorf(
			"Assertion failed: error message does not match pattern\n"+
				"  Expected pattern: %s\n"+
				"  Got error: %s",
			expectedPattern, err.Error(),
		)
	}
}

// AssertNoError checks if error is nil with context.
func AssertNoError(t *testing.T, err error, context string) {
	t.Helper()
	if err != nil {
		t.Errorf(
			"Assertion failed: %s\n  Expected no error but got: %v",
			context, err,
		)
	}
}

// AssertLen checks slice/map/string length with context.
func AssertLen(t *testing.T, actual interface{}, expectedLen int, context string) {
	t.Helper()
	// This is a simplified version; a real implementation would use reflection
	switch v := actual.(type) {
	case []interface{}:
		if len(v) != expectedLen {
			t.Errorf(
				"Assertion failed: %s\n  Expected length: %d\n  Got length: %d",
				context, expectedLen, len(v),
			)
		}
	case string:
		if len(v) != expectedLen {
			t.Errorf(
				"Assertion failed: %s\n  Expected length: %d\n  Got length: %d",
				context, expectedLen, len(v),
			)
		}
	default:
		t.Errorf("AssertLen: unsupported type %T", actual)
	}
}

// AssertTrue checks if condition is true.
func AssertTrue(t *testing.T, condition bool, context string) {
	t.Helper()
	if !condition {
		t.Errorf("Assertion failed: %s\n  Expected true but got false", context)
	}
}

// AssertFalse checks if condition is false.
func AssertFalse(t *testing.T, condition bool, context string) {
	t.Helper()
	if condition {
		t.Errorf("Assertion failed: %s\n  Expected false but got true", context)
	}
}

// AssertBoundedLatency checks if operation latency is within bounds.
// Useful for performance assertions.
func AssertBoundedLatency(t *testing.T, actualMs int64, maxMs int64, operation string) {
	t.Helper()
	if actualMs > maxMs {
		t.Errorf(
			"Performance assertion failed: %s exceeded latency budget\n"+
				"  Budget: %dms\n"+
				"  Actual: %dms\n"+
				"  Tip: profile with pprof to find bottleneck",
			operation, maxMs, actualMs,
		)
	}
}

// AssertResourceBounded checks if resource usage is bounded.
// Useful for memory/goroutine assertions.
func AssertResourceBounded(t *testing.T, actual, max int64, resourceType string) {
	t.Helper()
	if actual > max {
		t.Errorf(
			"Resource assertion failed: %s exceeded limit\n"+
				"  Limit: %d\n"+
				"  Actual: %d\n"+
				"  Tip: check for leaks or increase limit",
			resourceType, max, actual,
		)
	}
}

// ---- Private helpers ----

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
