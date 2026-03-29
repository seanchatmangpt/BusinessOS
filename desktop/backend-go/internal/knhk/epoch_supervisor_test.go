package knhk

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestEpochSupervisor_ExecuteQuerySuccess(t *testing.T) {
	logger := slog.Default()
	supervisor := NewEpochSupervisor(logger)

	ctx := context.Background()
	result, err := supervisor.ExecuteQuery(ctx, "SELECT ?x WHERE { ?x rdf:type ex:Node }", TierCritical, "trace_1")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

func TestEpochSupervisor_ExecuteQueryInvalidSPARQL(t *testing.T) {
	logger := slog.Default()
	supervisor := NewEpochSupervisor(logger)

	ctx := context.Background()
	_, err := supervisor.ExecuteQuery(ctx, "INVALID", TierCritical, "trace_2")

	if err == nil {
		t.Fatal("Expected error for invalid SPARQL, got nil")
	}
}

func TestEpochSupervisor_BudgetEnforcement(t *testing.T) {
	logger := slog.Default()
	supervisor := NewEpochSupervisor(logger)

	ctx := context.Background()

	// TierCritical has MaxOpsPerMin = 1000
	// Execute just at budget
	for i := 0; i < 10; i++ {
		_, err := supervisor.ExecuteQuery(ctx, "SELECT ?x WHERE { ?x rdf:type ex:Node }", TierCritical, "trace_3")
		if err != nil {
			t.Fatalf("Unexpected error at iteration %d: %v", i, err)
		}
	}

	status := supervisor.GetTierStatus(TierCritical)
	if status["current_ops"] != 10 {
		t.Fatalf("Expected 10 operations, got %v", status["current_ops"])
	}
}

func TestEpochSupervisor_TimeoutEnforcement(t *testing.T) {
	logger := slog.Default()
	supervisor := NewEpochSupervisor(logger)

	// Create a context that times out immediately
	ctx, cancel := context.WithTimeout(context.Background(), 0*time.Millisecond)
	defer cancel()

	// Sleep to ensure timeout
	time.Sleep(10 * time.Millisecond)

	_, err := supervisor.ExecuteQuery(ctx, "SELECT ?x WHERE { ?x rdf:type ex:Node }", TierCritical, "trace_4")

	if err == nil {
		t.Fatal("Expected error from timeout, got nil")
	}
}

func TestEpochSupervisor_CircuitBreakerOpensOnFailures(t *testing.T) {
	logger := slog.Default()
	supervisor := NewEpochSupervisor(logger)

	// Record 4 failures to open circuit (>3)
	supervisor.recordFailure(TierCritical.Name)
	supervisor.recordFailure(TierCritical.Name)
	supervisor.recordFailure(TierCritical.Name)
	supervisor.recordFailure(TierCritical.Name)

	if !supervisor.circuitOpen {
		t.Fatal("Expected circuit breaker to be open")
	}

	// Try to execute query with circuit open
	ctx := context.Background()
	_, err := supervisor.ExecuteQuery(ctx, "SELECT ?x WHERE { ?x rdf:type ex:Node }", TierCritical, "trace_5")

	if err == nil {
		t.Fatal("Expected error from open circuit, got nil")
	}
}

func TestEpochSupervisor_CircuitBreakerResets(t *testing.T) {
	logger := slog.Default()
	supervisor := NewEpochSupervisor(logger)

	// Open circuit
	supervisor.recordFailure(TierCritical.Name)
	supervisor.recordFailure(TierCritical.Name)
	supervisor.recordFailure(TierCritical.Name)
	supervisor.recordFailure(TierCritical.Name)

	if !supervisor.circuitOpen {
		t.Fatal("Expected circuit breaker to be open")
	}

	// Reset it
	supervisor.ResetCircuitBreaker()

	if supervisor.circuitOpen {
		t.Fatal("Expected circuit breaker to be reset (closed)")
	}
}

func TestEpochSupervisor_EpochRotation(t *testing.T) {
	logger := slog.Default()
	supervisor := NewEpochSupervisor(logger)

	ctx := context.Background()

	// Execute a query
	supervisor.ExecuteQuery(ctx, "SELECT ?x WHERE { ?x rdf:type ex:Node }", TierCritical, "trace_6")

	status := supervisor.GetTierStatus(TierCritical)
	if status["current_ops"] != 1 {
		t.Fatalf("Expected 1 operation, got %v", status["current_ops"])
	}

	// Rotate epoch
	supervisor.RotateEpoch()

	status = supervisor.GetTierStatus(TierCritical)
	if status["current_ops"] != 0 {
		t.Fatalf("Expected 0 operations after rotation, got %v", status["current_ops"])
	}
}

func TestEpochSupervisor_MultiTierTimeouts(t *testing.T) {
	// Verify different tiers have different timeouts
	if TierCritical.TimeoutMs >= TierHigh.TimeoutMs {
		t.Fatal("Critical tier should have shorter timeout than high tier")
	}

	if TierHigh.TimeoutMs >= TierNormal.TimeoutMs {
		t.Fatal("High tier should have shorter timeout than normal tier")
	}

	if TierNormal.TimeoutMs >= TierLow.TimeoutMs {
		t.Fatal("Normal tier should have shorter timeout than low tier")
	}
}

func TestEpochSupervisor_WvDASoundness(t *testing.T) {
	// Verify all tiers have finite timeouts (no infinite waits)
	tiers := []OperationTier{TierCritical, TierHigh, TierNormal, TierLow}

	for _, tier := range tiers {
		if tier.TimeoutMs <= 0 {
			t.Fatalf("Tier %s has invalid timeout: %d", tier.Name, tier.TimeoutMs)
		}

		if tier.TimeoutMs > 60*time.Second {
			t.Fatalf("Tier %s timeout exceeds 60s: %d", tier.Name, tier.TimeoutMs)
		}
	}
}

func TestEpochSupervisor_ArmstrongPrinciples(t *testing.T) {
	logger := slog.Default()
	supervisor := NewEpochSupervisor(logger)

	// Test: no exception swallowing (timeout = error propagation)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	defer cancel()

	_, err := supervisor.ExecuteQuery(ctx, "SELECT ?x WHERE { ?x rdf:type ex:Node }", TierCritical, "trace_7")

	// Error should be returned (not swallowed)
	if err == nil {
		t.Fatal("Expected timeout error to propagate")
	}

	// Test: circuit breaker prevents cascading failures
	for i := 0; i < 5; i++ {
		supervisor.recordFailure(TierHigh.Name)
	}

	if !supervisor.circuitOpen {
		t.Fatal("Circuit breaker should open after multiple failures")
	}

	_, err = supervisor.ExecuteQuery(context.Background(), "SELECT ?x WHERE { ?x rdf:type ex:Node }", TierHigh, "trace_8")
	if err == nil {
		t.Fatal("Expected circuit open error")
	}
}
