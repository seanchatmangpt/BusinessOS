package errors

import (
	"errors"
	"fmt"
	"strings"
)

// HintForError provides helpful context for common Go errors, reducing debugging time.
// Returns actionable suggestions for the 20% of errors causing 80% of confusion.
func HintForError(err error) string {
	if err == nil {
		return ""
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	// Nil pointer dereference (panic: runtime error)
	if strings.Contains(lower, "nil pointer dereference") || strings.Contains(lower, "invalid memory address") {
		return "Nil pointer dereference — check if variable is nil before dereferencing. " +
			"Use if x != nil { ... } guard. " +
			"Tip: add nil checks at all function entry points."
	}

	// Undefined method
	if strings.Contains(lower, "undefined") && strings.Contains(lower, "method") {
		return "Undefined method — did you mean to use a different method name? " +
			"Check the receiver type and available methods. " +
			"Tip: hover in IDE or run 'go doc <type>' to list all methods."
	}

	// Import cycle
	if strings.Contains(lower, "import cycle") {
		return "Import cycle detected — package A imports B, B imports A. " +
			"Move shared types to a third package. " +
			"Tip: use 'go mod graph | grep -A2 <pkg>' to visualize imports."
	}

	// Type mismatch
	if strings.Contains(lower, "cannot use") && strings.Contains(lower, "type") {
		parts := strings.Split(msg, " ")
		if len(parts) >= 3 {
			return fmt.Sprintf("Type mismatch — %s. "+
				"Check variable assignment and function signatures. "+
				"Tip: cast with explicit type conversion if needed.", msg)
		}
		return "Type mismatch — ensure variable and expected types match. " +
			"Use explicit type conversion: myInt := int(myFloat64)"
	}

	// Slice bounds out of range
	if strings.Contains(lower, "out of range") || strings.Contains(lower, "slice bounds") {
		return "Slice bounds out of range — index >= len(slice) or negative index. " +
			"Check bounds: if i >= 0 && i < len(slice) { ... }. " +
			"Tip: use len() and cap() to check slice limits."
	}

	// Channel send on closed channel
	if strings.Contains(lower, "send on closed channel") {
		return "Send on closed channel — never send to closed channel. " +
			"Use defer close(ch) after spawning goroutines. " +
			"Tip: only close from sender side; receivers should not close."
	}

	// Concurrent map access
	if strings.Contains(lower, "concurrent map") {
		return "Concurrent map access — multiple goroutines read/write map unsafely. " +
			"Protect with sync.RWMutex or use sync.Map for concurrent reads. " +
			"Tip: add mutex locks around all map operations."
	}

	// Timeout/context deadline
	if strings.Contains(lower, "context deadline") || strings.Contains(lower, "context canceled") {
		return "Context canceled or deadline exceeded — operation took too long. " +
			"Check timeout duration and slow operations inside context. " +
			"Tip: increase context timeout or optimize blocking operations."
	}

	// Connection/network errors
	if strings.Contains(lower, "connection refused") {
		return "Connection refused — service not listening on target address:port. " +
			"Check if server is running and listening on the right port. " +
			"Tip: verify with 'lsof -i :PORT' to see what's listening. See docs/TROUBLESHOOTING.md#port-already-in-use"
	}

	if strings.Contains(lower, "connection reset") {
		return "Connection reset — server closed connection unexpectedly. " +
			"Check server logs and ensure server handles graceful shutdown. " +
			"Tip: add retry logic with exponential backoff."
	}

	// EOF errors
	if errors.Is(err, errors.New("EOF")) || strings.Contains(lower, "unexpected eof") {
		return "Unexpected EOF — connection closed before reading expected data. " +
			"Check if sender closed connection prematurely. " +
			"Tip: add length prefix or delimiter to messages. See docs/TROUBLESHOOTING.md for integration help."
	}

	// Invalid JSON/data parsing
	if strings.Contains(lower, "invalid") && (strings.Contains(lower, "json") || strings.Contains(lower, "syntax")) {
		return "Invalid JSON/data format — parse failed. " +
			"Check JSON syntax: commas, quotes, brackets. Use https://jsonlint.com/ " +
			"Tip: print raw data before parsing: fmt.Printf(\"%q\\n\", rawData)"
	}

	// Permission denied
	if strings.Contains(lower, "permission denied") || strings.Contains(lower, "access denied") {
		return "Permission denied — no read/write/execute permission. " +
			"Check file permissions: ls -la. Grant with chmod 644 (file) or 755 (dir). " +
			"Tip: or run with appropriate user (sudo if needed)."
	}

	// File not found
	if strings.Contains(lower, "no such file") || strings.Contains(lower, "not found") {
		return "File not found — path doesn't exist or is wrong. " +
			"Check path with pwd and ls. Use absolute paths in tests. " +
			"Tip: print path being checked: fmt.Printf(\"Path: %s\\n\", path)"
	}

	// Goroutine panic
	if strings.Contains(lower, "panic") {
		return "Goroutine panicked — unhandled exception in background goroutine. " +
			"Add recover() in goroutines: defer func() { recover() }(). " +
			"Tip: capture panic with logs for debugging."
	}

	// Generic fallback
	return fmt.Sprintf("Error: %s — check logs for stack trace and context. "+
		"Tip: run with -v flag for verbose output or add logging statements.", msg)
}

// WvdAViolationHint returns helpful context for van der Aalst soundness violations
// (deadlock, liveness, boundedness).
func WvdAViolationHint(violation string) string {
	lower := strings.ToLower(violation)

	// Deadlock-related
	if strings.Contains(lower, "deadlock") || strings.Contains(lower, "blocking") {
		return "Deadlock risk: blocking operation without timeout. " +
			"Add explicit timeout_ms to all await(), receive(), get() calls. " +
			"Example: select { case <-ch: ... case <-time.After(5*time.Second): ... }"
	}

	// Liveness-related
	if strings.Contains(lower, "liveness") || strings.Contains(lower, "infinite") || strings.Contains(lower, "loop") {
		return "Liveness violation: infinite loop detected. " +
			"Add explicit loop bounds or exit condition. " +
			"Example: for i := 0; i < maxIterations; i++ { ... }"
	}

	// Boundedness-related
	if strings.Contains(lower, "boundedness") || strings.Contains(lower, "unbounded") || strings.Contains(lower, "queue") {
		return "Boundedness violation: unbounded resource growth. " +
			"Add max_size or TTL to queues, caches, maps. " +
			"Example: if len(queue) >= maxSize { evict_oldest() }"
	}

	// Resource limits
	if strings.Contains(lower, "resource") || strings.Contains(lower, "exhausted") {
		return "Resource exhaustion: memory/goroutine/connection limit exceeded. " +
			"Check runtime.NumGoroutine() and connection pools. " +
			"Add graceful degradation: queue requests or return 429 when full."
	}

	return "Soundness violation: verify all blocking ops have timeouts, " +
		"loops have bounds, and resources have limits."
}

// ArmstrongViolationHint returns helpful context for Armstrong fault tolerance violations
// (supervision, let-it-crash, shared state, budget).
func ArmstrongViolationHint(violation string) string {
	lower := strings.ToLower(violation)

	// Supervision
	if strings.Contains(lower, "supervision") || strings.Contains(lower, "unsupervised") {
		return "Supervision violation: process not supervised. " +
			"Every goroutine must have a supervisor monitoring for crashes. " +
			"Use context.Context cancellation and sync.WaitGroup."
	}

	// Let-it-crash
	if strings.Contains(lower, "swallow") || strings.Contains(lower, "silent") || strings.Contains(lower, "catch") {
		return "Let-it-crash violation: error silently swallowed. " +
			"Don't hide errors with || null or defer recover(). " +
			"Instead: fail fast, log, and let supervisor restart."
	}

	// Shared state
	if strings.Contains(lower, "shared") || strings.Contains(lower, "race") || strings.Contains(lower, "mutex") {
		return "Shared mutable state detected: potential race condition. " +
			"Use channels for communication, not shared memory. " +
			"Or protect with sync.Mutex: mu.Lock(); defer mu.Unlock()"
	}

	// Budget
	if strings.Contains(lower, "budget") || strings.Contains(lower, "timeout") {
		return "Budget violation: operation exceeded time/memory budget. " +
			"Add per-operation timeouts and resource limits. " +
			"Example: ctx, cancel := context.WithTimeout(parent, 5*time.Second)"
	}

	return "Fault tolerance violation: verify supervision tree, " +
		"no silent errors, message-passing only, and explicit budgets."
}
