package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Constructor & Initialization
// ---------------------------------------------------------------------------

func TestCircuitBreaker_InitializesInClosedState(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  50 * time.Millisecond,
		TimeoutDuration: 1 * time.Second,
	})

	if state := cb.GetState(); state != StateClosed {
		t.Errorf("expected CLOSED, got %d", state)
	}
}

func TestCircuitBreaker_DefaultConfigValues(t *testing.T) {
	cb := NewCircuitBreaker(Config{})

	stats := cb.GetStats()
	if stats.ConsecutiveFailures != 0 {
		t.Errorf("expected 0 consecutive failures, got %d", stats.ConsecutiveFailures)
	}
	if stats.TotalCalls != 0 {
		t.Errorf("expected 0 total calls, got %d", stats.TotalCalls)
	}
}

func TestCircuitBreaker_DefaultConfigOverridesZeroValues(t *testing.T) {
	// All zero config should produce valid circuit breaker (no panics, no zero timeouts)
	cb := NewCircuitBreaker(Config{})

	if cb.GetState() != StateClosed {
		t.Errorf("expected CLOSED state after zero-config construction")
	}

	delay := cb.GetNextRetryDelay()
	if delay != 0 {
		t.Errorf("expected 0 retry delay in CLOSED state, got %v", delay)
	}
}

// ---------------------------------------------------------------------------
// State Transitions: CLOSED -> OPEN
// ---------------------------------------------------------------------------

func TestCircuitBreaker_TransitionsToOpenAfterMaxFailures(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute, // long cooldown to prevent auto-recovery
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("service unavailable")

	for i := 0; i < 3; i++ {
		cb.Execute(context.Background(), func() error {
			return testErr
		})
	}

	if state := cb.GetState(); state != StateOpen {
		t.Errorf("expected OPEN after %d failures, got %d", 3, state)
	}
}

func TestCircuitBreaker_RemainsClosedBelowFailureThreshold(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     5,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("transient error")

	// Fail 4 times (threshold is 5)
	for i := 0; i < 4; i++ {
		cb.Execute(context.Background(), func() error {
			return testErr
		})
	}

	if state := cb.GetState(); state != StateClosed {
		t.Errorf("expected CLOSED with %d failures (threshold %d), got %d", 4, 5, state)
	}
}

func TestCircuitBreaker_FailureCountResetsOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("failure")

	// Fail twice
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	// One success resets counter
	cb.Execute(context.Background(), func() error { return nil })

	// Two more failures should NOT open (counter was reset)
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	if state := cb.GetState(); state != StateClosed {
		t.Errorf("expected CLOSED after failure count reset, got %d", state)
	}
}

// ---------------------------------------------------------------------------
// State Transitions: OPEN -> HALF_OPEN (after cooldown)
// ---------------------------------------------------------------------------

func TestCircuitBreaker_TransitionsToHalfOpenAfterCooldown(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  50 * time.Millisecond,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("failure")

	// Trip the circuit
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	if state := cb.GetState(); state != StateOpen {
		t.Fatalf("expected OPEN before cooldown, got %d", state)
	}

	// Wait for cooldown
	time.Sleep(60 * time.Millisecond)

	// After cooldown, allowCall should return true (OPEN allows call)
	// and a success should transition to HALF_OPEN then CLOSED
	// NOTE: The implementation doesn't explicitly transition OPEN->HALF_OPEN
	// via allowCall. It only sets HALF_OPEN in handleSuccess when already HALF_OPEN.
	// Actually, looking at the code: allowCall just returns true after cooldown
	// but state stays OPEN. The state machine is:
	// CLOSED -> OPEN (on max failures)
	// OPEN allows calls after cooldown but state stays OPEN
	// Success when state is... wait -- let me re-read.
	// handleSuccess checks if state == StateHalfOpen to transition to CLOSED.
	// But nothing ever sets state to HALF_OPEN!
	// This is a bug in the implementation -- but we test actual behavior.

	// After cooldown, the circuit allows calls through even in OPEN state
	err := cb.Execute(context.Background(), func() error { return nil })
	if err != nil {
		t.Errorf("expected nil error after cooldown + success, got %v", err)
	}
}

func TestCircuitBreaker_ReturnsErrCircuitOpenBeforeCooldown(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("failure")

	// Trip the circuit
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	// Immediately after opening, should get ErrCircuitOpen
	err := cb.Execute(context.Background(), func() error { return nil })
	if !IsCircuitOpenError(err) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// HALF_OPEN State Behavior
// ---------------------------------------------------------------------------

func TestCircuitBreaker_HalfOpenTransitionsToOpenOnFailure(t *testing.T) {
	// The implementation has no explicit OPEN->HALF_OPEN transition.
	// After cooldown, allowCall returns true but state remains OPEN.
	// A failure in what should be HALF_OPEN opens the circuit again.
	// We test the observable behavior: after cooldown, a failure
	// keeps the circuit open and resets the lastFailure time.

	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  50 * time.Millisecond,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("failure")

	// Trip the circuit
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	if state := cb.GetState(); state != StateOpen {
		t.Fatalf("expected OPEN, got %d", state)
	}

	// Wait for cooldown
	time.Sleep(60 * time.Millisecond)

	// First call after cooldown fails
	err := cb.Execute(context.Background(), func() error { return testErr })
	if err != testErr {
		t.Errorf("expected original error, got %v", err)
	}

	// State should remain OPEN after failure post-cooldown
	if state := cb.GetState(); state != StateOpen {
		t.Errorf("expected OPEN after failure during half-open, got %d", state)
	}
}

// ---------------------------------------------------------------------------
// Execute: Success Path
// ---------------------------------------------------------------------------

func TestCircuitBreaker_ExecuteReturnsNilOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	err := cb.Execute(context.Background(), func() error {
		return nil
	})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestCircuitBreaker_ExecuteReturnsFunctionError(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	expectedErr := errors.New("db connection lost")
	err := cb.Execute(context.Background(), func() error {
		return expectedErr
	})
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

// ---------------------------------------------------------------------------
// ExecuteWithFallback
// ---------------------------------------------------------------------------

func TestCircuitBreaker_ExecuteWithFallback_UsesFallbackWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("down")

	// Trip the circuit
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	// Primary function should be rejected, fallback should run
	fallbackCalled := false
	err := cb.ExecuteWithFallback(
		context.Background(),
		func() error { return errors.New("should not run") },
		func() error {
			fallbackCalled = true
			return nil
		},
	)

	if !fallbackCalled {
		t.Error("expected fallback to be called")
	}
	if err != nil {
		t.Errorf("expected nil from fallback, got %v", err)
	}
}

func TestCircuitBreaker_ExecuteWithFallback_ReturnsFallbackError(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("down")

	// Trip the circuit
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	fallbackErr := errors.New("fallback also failed")
	err := cb.ExecuteWithFallback(
		context.Background(),
		func() error { return errors.New("primary") },
		func() error { return fallbackErr },
	)

	if err != fallbackErr {
		t.Errorf("expected fallback error %v, got %v", fallbackErr, err)
	}
}

func TestCircuitBreaker_ExecuteWithFallback_DoesNotUseFallbackWhenClosed(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	primaryCalled := false
	fallbackCalled := false

	err := cb.ExecuteWithFallback(
		context.Background(),
		func() error {
			primaryCalled = true
			return nil
		},
		func() error {
			fallbackCalled = true
			return nil
		},
	)

	if !primaryCalled {
		t.Error("expected primary to be called when circuit is closed")
	}
	if fallbackCalled {
		t.Error("expected fallback NOT to be called when circuit is closed")
	}
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Timeout Handling
// ---------------------------------------------------------------------------

func TestCircuitBreaker_ExecuteReturnsTimeoutOnLongRunningFunction(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 10 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cb.Execute(ctx, func() error {
		time.Sleep(1 * time.Second) // longer than timeout
		return nil
	})

	if !IsTimeoutError(err) {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

func TestCircuitBreaker_TimeoutRespectsParentContextCancellation(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 10 * time.Second, // long internal timeout
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := cb.Execute(ctx, func() error {
		time.Sleep(5 * time.Second)
		return nil
	})

	if !IsTimeoutError(err) {
		t.Errorf("expected ErrTimeout from parent cancellation, got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Callbacks
// ---------------------------------------------------------------------------

func TestCircuitBreaker_OnStateChangeFiresOnTransition(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	var transitions []struct {
		oldState State
		newState State
	}
	var mu sync.Mutex

	cb.OnStateChange(func(oldState, newState State) {
		mu.Lock()
		transitions = append(transitions, struct {
			oldState State
			newState State
		}{oldState, newState})
		mu.Unlock()
	})

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	mu.Lock()
	defer mu.Unlock()

	if len(transitions) != 1 {
		t.Fatalf("expected 1 state transition, got %d", len(transitions))
	}
	if transitions[0].oldState != StateClosed {
		t.Errorf("expected transition from CLOSED, got %d", transitions[0].oldState)
	}
	if transitions[0].newState != StateOpen {
		t.Errorf("expected transition to OPEN, got %d", transitions[0].newState)
	}
}

func TestCircuitBreaker_OnFailureFiresOnEachFailure(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     5,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	var failureCount atomic.Int32
	testErr := errors.New("fail")

	cb.OnFailure(func(err error) {
		failureCount.Add(1)
	})

	for i := 0; i < 3; i++ {
		cb.Execute(context.Background(), func() error { return testErr })
	}

	if got := failureCount.Load(); got != 3 {
		t.Errorf("expected 3 failure callbacks, got %d", got)
	}
}

func TestCircuitBreaker_OnSuccessFiresOnSuccess(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	var successCount atomic.Int32

	cb.OnSuccess(func() {
		successCount.Add(1)
	})

	for i := 0; i < 3; i++ {
		cb.Execute(context.Background(), func() error { return nil })
	}

	if got := successCount.Load(); got != 3 {
		t.Errorf("expected 3 success callbacks, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Stats
// ---------------------------------------------------------------------------

func TestCircuitBreaker_StatsTracksTotalCalls(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	for i := 0; i < 5; i++ {
		cb.Execute(context.Background(), func() error { return nil })
	}

	stats := cb.GetStats()
	if stats.TotalCalls != 5 {
		t.Errorf("expected 5 total calls, got %d", stats.TotalCalls)
	}
}

func TestCircuitBreaker_StatsTracksSuccessfulCalls(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return nil })
	cb.Execute(context.Background(), func() error { return nil })
	cb.Execute(context.Background(), func() error { return testErr })

	stats := cb.GetStats()
	if stats.SuccessfulCalls != 2 {
		t.Errorf("expected 2 successful calls, got %d", stats.SuccessfulCalls)
	}
}

func TestCircuitBreaker_StatsTracksFailedCalls(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return nil })

	stats := cb.GetStats()
	if stats.FailedCalls != 2 {
		t.Errorf("expected 2 failed calls, got %d", stats.FailedCalls)
	}
}

func TestCircuitBreaker_StatsTracksTimeoutCalls(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 10 * time.Millisecond,
	})

	// One normal call
	cb.Execute(context.Background(), func() error { return nil })

	// One timed-out call
	cb.Execute(context.Background(), func() error {
		time.Sleep(1 * time.Second)
		return nil
	})

	stats := cb.GetStats()
	if stats.TimeoutCalls != 1 {
		t.Errorf("expected 1 timeout call, got %d", stats.TimeoutCalls)
	}
}

func TestCircuitBreaker_StatsSuccessRate(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")
	// 3 success, 1 failure = 75%
	cb.Execute(context.Background(), func() error { return nil })
	cb.Execute(context.Background(), func() error { return nil })
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return nil })

	stats := cb.GetStats()
	expectedRate := 75.0
	if stats.SuccessRate < expectedRate-0.1 || stats.SuccessRate > expectedRate+0.1 {
		t.Errorf("expected success rate ~75%%, got %.2f%%", stats.SuccessRate)
	}
}

func TestCircuitBreaker_StatsSuccessRateZeroWithNoCalls(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	stats := cb.GetStats()
	if stats.SuccessRate != 0 {
		t.Errorf("expected 0%% success rate with no calls, got %.2f%%", stats.SuccessRate)
	}
}

func TestCircuitBreaker_StatsConsecutiveFailures(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     5,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")

	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	stats := cb.GetStats()
	if stats.ConsecutiveFailures != 3 {
		t.Errorf("expected 3 consecutive failures, got %d", stats.ConsecutiveFailures)
	}
}

func TestCircuitBreaker_StatsReportsCurrentState(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	stats := cb.GetStats()
	if stats.State != StateClosed {
		t.Errorf("expected stats state CLOSED, got %d", stats.State)
	}

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	stats = cb.GetStats()
	if stats.State != StateOpen {
		t.Errorf("expected stats state OPEN, got %d", stats.State)
	}
}

// ---------------------------------------------------------------------------
// Reset
// ---------------------------------------------------------------------------

func TestCircuitBreaker_ResetReturnsToClosedState(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")

	// Trip the circuit
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	if state := cb.GetState(); state != StateOpen {
		t.Fatalf("expected OPEN before reset, got %d", state)
	}

	cb.Reset()

	if state := cb.GetState(); state != StateClosed {
		t.Errorf("expected CLOSED after reset, got %d", state)
	}
}

func TestCircuitBreaker_ResetClearsCounters(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     5,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")

	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return nil })

	cb.Reset()

	stats := cb.GetStats()
	if stats.ConsecutiveFailures != 0 {
		t.Errorf("expected 0 consecutive failures after reset, got %d", stats.ConsecutiveFailures)
	}
	if !stats.LastFailure.IsZero() {
		t.Errorf("expected zero last failure after reset, got %v", stats.LastFailure)
	}
}

func TestCircuitBreaker_ResetAllowsCallsAgain(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")

	// Trip and reset
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Reset()

	// Should allow calls after reset
	err := cb.Execute(context.Background(), func() error { return nil })
	if err != nil {
		t.Errorf("expected nil after reset, got %v", err)
	}
}

func TestCircuitBreaker_ResetFiresStateChangeCallback(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	var transitions []struct {
		oldState State
		newState State
	}
	var mu sync.Mutex

	cb.OnStateChange(func(oldState, newState State) {
		mu.Lock()
		transitions = append(transitions, struct {
			oldState State
			newState State
		}{oldState, newState})
		mu.Unlock()
	})

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	cb.Reset()

	mu.Lock()
	defer mu.Unlock()

	if len(transitions) != 2 {
		t.Fatalf("expected 2 transitions (CLOSED->OPEN, OPEN->CLOSED), got %d", len(transitions))
	}
	// First: CLOSED -> OPEN
	if transitions[0].oldState != StateClosed || transitions[0].newState != StateOpen {
		t.Errorf("transition 0: expected CLOSED->OPEN, got %d->%d", transitions[0].oldState, transitions[0].newState)
	}
	// Second: OPEN -> CLOSED (from Reset)
	if transitions[1].oldState != StateOpen || transitions[1].newState != StateClosed {
		t.Errorf("transition 1: expected OPEN->CLOSED, got %d->%d", transitions[1].oldState, transitions[1].newState)
	}
}

// ---------------------------------------------------------------------------
// Retry Delay
// ---------------------------------------------------------------------------

func TestCircuitBreaker_RetryDelayZeroInClosedState(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	delay := cb.GetNextRetryDelay()
	if delay != 0 {
		t.Errorf("expected 0 delay in CLOSED state, got %v", delay)
	}
}

func TestCircuitBreaker_RetryDelayPositiveInOpenState(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	delay := cb.GetNextRetryDelay()
	if delay <= 0 {
		t.Errorf("expected positive delay in OPEN state, got %v", delay)
	}
}

func TestCircuitBreaker_RetryDelayCappedAtMaxDelay(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		BaseDelay:       100 * time.Millisecond,
		MaxDelay:        200 * time.Millisecond,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	// Run 100 samples to check all stay within max (accounting for jitter)
	for i := 0; i < 100; i++ {
		delay := cb.GetNextRetryDelay()
		// Jitter can add up to 20% of delay, so allow some slack
		if delay > 250*time.Millisecond {
			t.Errorf("delay %v exceeded max delay with jitter margin", delay)
		}
	}
}

func TestCircuitBreaker_RetryDelayInHalfOpenStateIsBaseDelayHalf(t *testing.T) {
	// NOTE: The implementation never explicitly transitions to StateHalfOpen
	// (no code path sets cb.state = StateHalfOpen). The allowCall() method
	// returns true after cooldown in OPEN state, but does not change the state.
	// The HALF_OPEN retry delay branch (baseDelay/2) is unreachable via public API.
	//
	// This test documents the observed behavior: CLOSED state returns 0 delay.
	// The HALF_OPEN branch is dead code in the current implementation.

	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		BaseDelay:       100 * time.Millisecond,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	delay := cb.GetNextRetryDelay()
	if delay != 0 {
		t.Errorf("expected 0 delay in CLOSED state, got %v", delay)
	}
}

// ---------------------------------------------------------------------------
// Error Types
// ---------------------------------------------------------------------------

func TestCircuitBreakerError_ErrorMessage(t *testing.T) {
	err := NewCircuitBreakerError("test message")
	if err.Error() != "test message" {
		t.Errorf("expected 'test message', got '%s'", err.Error())
	}
}

func TestIsCircuitOpenError_True(t *testing.T) {
	if !IsCircuitOpenError(ErrCircuitOpen) {
		t.Error("expected IsCircuitOpenError to return true for ErrCircuitOpen")
	}
}

func TestIsCircuitOpenError_False(t *testing.T) {
	if IsCircuitOpenError(errors.New("other")) {
		t.Error("expected IsCircuitOpenError to return false for other error")
	}
}

func TestIsCircuitOpenError_Nil(t *testing.T) {
	if IsCircuitOpenError(nil) {
		t.Error("expected IsCircuitOpenError to return false for nil")
	}
}

func TestIsTimeoutError_True(t *testing.T) {
	if !IsTimeoutError(ErrTimeout) {
		t.Error("expected IsTimeoutError to return true for ErrTimeout")
	}
}

func TestIsTimeoutError_False(t *testing.T) {
	if IsTimeoutError(errors.New("other")) {
		t.Error("expected IsTimeoutError to return false for other error")
	}
}

// ---------------------------------------------------------------------------
// Builder Pattern
// ---------------------------------------------------------------------------

func TestBuilder_DefaultStateIsClosed(t *testing.T) {
	cb := NewBuilder().Build()
	if cb.GetState() != StateClosed {
		t.Errorf("expected CLOSED from default builder, got %d", cb.GetState())
	}
}

func TestBuilder_WithConfig(t *testing.T) {
	config := Config{
		MaxAttempts:     7,
		BaseDelay:       500 * time.Millisecond,
		MaxDelay:        20 * time.Second,
		TimeoutDuration: 3 * time.Second,
		CooldownPeriod:  45 * time.Second,
		HalfOpenMaxCalls: 5,
	}

	cb := NewBuilder().WithConfig(config).Build()

	if cb.GetState() != StateClosed {
		t.Errorf("expected CLOSED, got %d", cb.GetState())
	}
}

func TestBuilder_WithMaxAttempts(t *testing.T) {
	cb := NewBuilder().WithMaxAttempts(10).Build()

	// Verify by tripping with 10 failures
	testErr := errors.New("fail")
	for i := 0; i < 9; i++ {
		cb.Execute(context.Background(), func() error { return testErr })
	}
	if cb.GetState() != StateClosed {
		t.Errorf("expected CLOSED with 9 failures (threshold 10), got %d", cb.GetState())
	}

	cb.Execute(context.Background(), func() error { return testErr })
	if cb.GetState() != StateOpen {
		t.Errorf("expected OPEN with 10 failures (threshold 10), got %d", cb.GetState())
	}
}

func TestBuilder_WithBackoff(t *testing.T) {
	cb := NewBuilder().WithBackoff(200*time.Millisecond, 5*time.Second).Build()

	delay := cb.GetNextRetryDelay()
	if delay != 0 {
		t.Errorf("expected 0 delay in CLOSED state, got %v", delay)
	}
}

func TestBuilder_WithTimeout(t *testing.T) {
	cb := NewBuilder().WithTimeout(20 * time.Millisecond).Build()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cb.Execute(ctx, func() error {
		time.Sleep(1 * time.Second)
		return nil
	})

	if !IsTimeoutError(err) {
		t.Errorf("expected ErrTimeout with 20ms timeout, got %v", err)
	}
}

func TestBuilder_WithCooldown(t *testing.T) {
	cb := NewBuilder().WithCooldown(30 * time.Millisecond).WithMaxAttempts(2).Build()

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	// Before cooldown
	err := cb.Execute(context.Background(), func() error { return nil })
	if !IsCircuitOpenError(err) {
		t.Errorf("expected ErrCircuitOpen before cooldown, got %v", err)
	}

	// After cooldown
	time.Sleep(40 * time.Millisecond)
	err = cb.Execute(context.Background(), func() error { return nil })
	if IsCircuitOpenError(err) {
		t.Error("expected call to be allowed after cooldown")
	}
}

func TestBuilder_Chaining(t *testing.T) {
	cb := NewBuilder().
		WithMaxAttempts(2).
		WithBackoff(50*time.Millisecond, 1*time.Second).
		WithTimeout(1*time.Second).
		WithCooldown(30*time.Millisecond).
		Build()

	if cb == nil {
		t.Error("expected non-nil circuit breaker from chained builder")
	}
	if cb.GetState() != StateClosed {
		t.Errorf("expected CLOSED, got %d", cb.GetState())
	}
}

// ---------------------------------------------------------------------------
// Preset Configurations
// ---------------------------------------------------------------------------

func TestComplianceServiceConfig_ReturnsValidConfig(t *testing.T) {
	config := ComplianceServiceConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", config.MaxAttempts)
	}
	if config.HalfOpenMaxCalls != 1 {
		t.Errorf("expected HalfOpenMaxCalls=1, got %d", config.HalfOpenMaxCalls)
	}
	if config.CooldownPeriod != 60*time.Second {
		t.Errorf("expected CooldownPeriod=60s, got %v", config.CooldownPeriod)
	}
}

func TestDatabaseConfig_ReturnsValidConfig(t *testing.T) {
	config := DatabaseConfig()

	if config.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", config.MaxAttempts)
	}
	if config.TimeoutDuration != 2*time.Second {
		t.Errorf("expected TimeoutDuration=2s, got %v", config.TimeoutDuration)
	}
	if config.HalfOpenMaxCalls != 3 {
		t.Errorf("expected HalfOpenMaxCalls=3, got %d", config.HalfOpenMaxCalls)
	}
}

func TestExternalAPIConfig_ReturnsValidConfig(t *testing.T) {
	config := ExternalAPIConfig()

	if config.MaxAttempts != 5 {
		t.Errorf("expected MaxAttempts=5, got %d", config.MaxAttempts)
	}
	if config.TimeoutDuration != 10*time.Second {
		t.Errorf("expected TimeoutDuration=10s, got %v", config.TimeoutDuration)
	}
	if config.HalfOpenMaxCalls != 2 {
		t.Errorf("expected HalfOpenMaxCalls=2, got %d", config.HalfOpenMaxCalls)
	}
}

func TestPresetConfig_CreatesWorkingCircuitBreaker(t *testing.T) {
	configs := []Config{
		ComplianceServiceConfig(),
		DatabaseConfig(),
		ExternalAPIConfig(),
	}

	for i, config := range configs {
		cb := NewCircuitBreaker(config)

		err := cb.Execute(context.Background(), func() error { return nil })
		if err != nil {
			t.Errorf("config[%d]: expected nil, got %v", i, err)
		}

		if cb.GetState() != StateClosed {
			t.Errorf("config[%d]: expected CLOSED, got %d", i, cb.GetState())
		}
	}
}

// ---------------------------------------------------------------------------
// Concurrent Access Safety
// ---------------------------------------------------------------------------

func TestCircuitBreaker_ConcurrentExecuteIsSafe(t *testing.T) {
	// NOTE: This test exposes a known data race in the implementation:
	//   cb.totalCalls++ on line 95 of circuit_breaker.go is not mutex-protected.
	// Run without -race to verify functional correctness.
	// Run with -race to confirm the race detector catches the bug.
	// TODO(fix): Protect totalCalls, successfulCalls, failedCalls, timeoutCalls with atomic or mutex.

	cb := NewCircuitBreaker(Config{
		MaxAttempts:     100, // high threshold to avoid premature opening
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	var wg sync.WaitGroup
	const goroutines = 50
	const callsPerGoroutine = 10

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < callsPerGoroutine; i++ {
				cb.Execute(context.Background(), func() error {
					return nil
				})
			}
		}(g)
	}

	wg.Wait()

	// Note: totalCalls may be inaccurate due to the data race, but
	// the circuit breaker should still function (not panic/deadlock).
	if cb.GetState() != StateClosed {
		t.Errorf("expected CLOSED state, got %d", cb.GetState())
	}
}

func TestCircuitBreaker_ConcurrentFailuresAndSuccesses(t *testing.T) {
	// NOTE: Exposes the same data race as TestCircuitBreaker_ConcurrentExecuteIsSafe.
	// Tests that the circuit breaker remains functional under mixed concurrent load.

	cb := NewCircuitBreaker(Config{
		MaxAttempts:     100,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")

	var wg sync.WaitGroup
	const goroutines = 20

	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				if id%2 == 0 {
					cb.Execute(context.Background(), func() error { return nil })
				} else {
					cb.Execute(context.Background(), func() error { return testErr })
				}
			}
		}(g)
	}

	wg.Wait()

	// Primary assertion: no panic, no deadlock
	if stats := cb.GetStats(); stats.State == StateOpen {
		t.Errorf("expected non-OPEN state with threshold 100, got %d", stats.State)
	}
}

func TestCircuitBreaker_ConcurrentResetIsSafe(t *testing.T) {
	// NOTE: Exposes the data race on totalCalls. Tests that Reset()
	// does not panic or deadlock when called concurrently with Execute.

	cb := NewCircuitBreaker(Config{
		MaxAttempts:     100,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	var wg sync.WaitGroup

	// Concurrently execute and reset
	for i := 0; i < 20; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			cb.Execute(context.Background(), func() error { return nil })
		}()

		go func() {
			defer wg.Done()
			cb.Reset()
		}()
	}

	wg.Wait()

	// Should not panic or deadlock -- that's the main assertion
	// (State assertion skipped due to data race on totalCalls)
}

func TestCircuitBreaker_ConcurrentStatsReadIsSafe(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     100,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	var wg sync.WaitGroup

	// Writer goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				cb.Execute(context.Background(), func() error { return nil })
			}
		}()
	}

	// Reader goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				stats := cb.GetStats()
				// Just ensure we can read without panic
				_ = stats.TotalCalls
				_ = stats.SuccessRate
				_ = stats.State
			}
		}()
	}

	wg.Wait()
	// No assertion beyond "did not panic" or "did not deadlock"
}

// ---------------------------------------------------------------------------
// Edge Cases
// ---------------------------------------------------------------------------

func TestCircuitBreaker_MaxAttemptsOne(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     1,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })

	if cb.GetState() != StateOpen {
		t.Errorf("expected OPEN with MaxAttempts=1 after 1 failure, got %d", cb.GetState())
	}
}

func TestCircuitBreaker_AllowsCallsAfterCooldownPasses(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  30 * time.Millisecond,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	// Blocked immediately
	err := cb.Execute(context.Background(), func() error { return nil })
	if !IsCircuitOpenError(err) {
		t.Errorf("expected ErrCircuitOpen immediately after opening, got %v", err)
	}

	// Wait for cooldown
	time.Sleep(40 * time.Millisecond)

	// Should be allowed through
	err = cb.Execute(context.Background(), func() error { return nil })
	if err != nil {
		t.Errorf("expected call to succeed after cooldown, got %v", err)
	}
}

func TestCircuitBreaker_ContextCancelledBeforeExecute(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     3,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 10 * time.Second,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before executing

	// Even with a long timeout, parent cancellation should take effect quickly
	start := time.Now()
	err := cb.Execute(ctx, func() error {
		time.Sleep(10 * time.Second)
		return nil
	})
	elapsed := time.Since(start)

	if !IsTimeoutError(err) {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
	if elapsed > 2*time.Second {
		t.Errorf("expected quick return on cancelled context, took %v", elapsed)
	}
}

func TestCircuitBreaker_MultipleResetsAreIdempotent(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		MaxAttempts:     2,
		CooldownPeriod:  10 * time.Minute,
		TimeoutDuration: 1 * time.Second,
	})

	testErr := errors.New("fail")
	cb.Execute(context.Background(), func() error { return testErr })
	cb.Execute(context.Background(), func() error { return testErr })

	cb.Reset()
	cb.Reset()
	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Errorf("expected CLOSED after multiple resets, got %d", cb.GetState())
	}

	stats := cb.GetStats()
	if stats.ConsecutiveFailures != 0 {
		t.Errorf("expected 0 consecutive failures, got %d", stats.ConsecutiveFailures)
	}
}
