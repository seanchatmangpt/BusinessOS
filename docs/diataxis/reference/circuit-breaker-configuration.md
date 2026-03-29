---
title: "Reference: Circuit Breaker Configuration"
type: reference
signal: "S=(linguistic, reference, inform, markdown, table-driven)"
relates_to: [hipaa-compliance-validator, api-endpoints, bos-gateway-pattern]
version: "1.0.0"
created: "2026-03-27"
---

# Circuit Breaker — Reference

**Package:** `github.com/rhl/businessos-backend/internal/circuitbreaker`
**Core type:** `circuitbreaker.CircuitBreaker`
**Wrapper type:** `services.CircuitBreakerWrapper` (compliance service)
**Source files:**
- `desktop/backend-go/internal/circuitbreaker/circuit_breaker.go`
- `desktop/backend-go/internal/services/compliance_service_circuitbreaker.go`

---

## Overview

BusinessOS uses a custom circuit breaker (`circuitbreaker.CircuitBreaker`) to guard outbound calls to the OSA service from the compliance layer. The implementation follows the classic three-state model with exponential backoff and jitter. When the circuit opens, `CircuitBreakerWrapper` activates per-method fallback strategies so compliance handlers continue to return degraded-but-valid responses instead of 5xx errors. The breaker is instantiated at bootstrap via `NewCircuitBreakerWrapper` using the `ComplianceServiceConfig()` preset.

---

## States

| State | Integer | Description | Entry Condition | Exit Condition |
|-------|---------|-------------|-----------------|----------------|
| `StateClosed` | `0` | Normal operation. All calls pass through. | Initial state; or after `halfOpenMaxCalls` consecutive successes in HALF-OPEN. | `consecutiveFailures >= maxAttempts` → transitions to OPEN. |
| `StateOpen` | `1` | Circuit tripped. Calls are rejected immediately without executing. | `consecutiveFailures >= maxAttempts`. | `time.Since(lastFailure) >= cooldownPeriod` → allows probe call (re-evaluated per call in `allowCall`). |
| `StateHalfOpen` | `2` | Probe mode. A limited number of calls are permitted to test recovery. | Implicit: first call allowed after OPEN cooldown expires. | Success: `successCount >= halfOpenMaxCalls` → CLOSED. Failure: any error → OPEN. |

---

## `Config` Struct Fields

All fields are in package `circuitbreaker`. Zero values trigger defaults (applied in `NewCircuitBreaker`).

| Field | Go Type | Default | Description |
|-------|---------|---------|-------------|
| `MaxAttempts` | `int` | `5` | Consecutive failures before the circuit opens. |
| `BaseDelay` | `time.Duration` | `100ms` | Starting delay for exponential backoff (used by `GetNextRetryDelay`). |
| `MaxDelay` | `time.Duration` | `10s` | Cap on exponential backoff delay. |
| `TimeoutDuration` | `time.Duration` | `5s` | Per-call context timeout applied inside `executeWithTimeout`. |
| `CooldownPeriod` | `time.Duration` | `30s` | How long the circuit stays OPEN before allowing a probe call. |
| `HalfOpenMaxCalls` | `int` | `3` | Consecutive successes in HALF-OPEN required to close the circuit. |

---

## Pre-defined Configuration Presets

Three `Config` factory functions are provided in `circuit_breaker.go`:

| Function | `MaxAttempts` | `BaseDelay` | `MaxDelay` | `TimeoutDuration` | `CooldownPeriod` | `HalfOpenMaxCalls` | Intended Use |
|----------|--------------|-------------|------------|-------------------|------------------|--------------------|--------------|
| `ComplianceServiceConfig()` | `3` | `1s` | `30s` | `5s` | `60s` | `1` | OSA compliance calls (critical; few retries, long cooldown) |
| `DatabaseConfig()` | `5` | `100ms` | `5s` | `2s` | `10s` | `3` | Database operations (fast recovery expected) |
| `ExternalAPIConfig()` | `5` | `200ms` | `10s` | `10s` | `30s` | `2` | Generic external HTTP API calls |

---

## `Stats` Struct Fields

Returned by `(*CircuitBreaker).GetStats()` and surfaced in `HealthCheck()`.

| Field | Go Type | Description |
|-------|---------|-------------|
| `TotalCalls` | `int64` | Cumulative calls attempted since creation or last `Reset()`. |
| `SuccessfulCalls` | `int64` | Calls that returned `nil` error. |
| `FailedCalls` | `int64` | Calls that returned a non-nil error (includes circuit-open rejections). |
| `TimeoutCalls` | `int64` | Calls terminated by `TimeoutDuration` context expiry. |
| `SuccessRate` | `float64` | `SuccessfulCalls / TotalCalls * 100`. Zero when `TotalCalls == 0`. |
| `State` | `State` | Current state integer (`0`=CLOSED, `1`=OPEN, `2`=HALF-OPEN). |
| `LastFailure` | `time.Time` | Timestamp of most recent failure; zero value if no failures. |
| `ConsecutiveFailures` | `int` | Current run of consecutive failures without a success. Reset to `0` on any success. |

---

## Backoff Formula

`GetNextRetryDelay()` returns the recommended wait before the next call attempt:

| State | Formula |
|-------|---------|
| `StateClosed` | `0` (no delay) |
| `StateOpen` | `min(BaseDelay × 2^consecutiveFailures, MaxDelay) ± 20% jitter` |
| `StateHalfOpen` | `BaseDelay / 2` |

Jitter coefficient: `delay × 0.2 × (rand.Float64() - 0.5)` — prevents thundering herd after cooldown.

---

## Integration Points

| Integration | Type | Constructor | Preset Used |
|-------------|------|-------------|-------------|
| **Compliance service** | `services.CircuitBreakerWrapper` | `NewCircuitBreakerWrapper(osaBaseURL, logger)` | `ComplianceServiceConfig()` |
| **OSA resilient client** | `osa.ResilientClientConfig.CircuitBreakerConfig` | Wired in `cmd/server/bootstrap.go` | `osa.DefaultCircuitBreakerConfig()` |
| **App generation orchestrator** | Internal stub; metrics exposed via `GetCircuitBreakerMetrics()` | `services.AppGenerationOrchestrator` | N/A (stub) |

The compliance wrapper protects these `ComplianceService` methods:

| Method | Fallback on OPEN |
|--------|-----------------|
| `GetStatus` | `getFallbackStatus()` — returns `OverallScore: 0.5`, all domains at `0.5` |
| `GetAuditTrail` | `ExecuteWithFallback` — returns empty `AuditTrailResponse` |
| `CollectEvidence` | `ExecuteWithFallback` — returns single `degraded_mode` evidence item |
| `GetGapAnalysis` | `getFallbackGapAnalysis()` — returns one placeholder gap, `Score: 0.5` |
| `VerifyAuditChain` | Returns `VerifyResult{Verified: false}` with error message in `Issues` |
| `EvaluateAuditEvent` | Silently skips (returns `nil`) |
| `ReloadRules` | Silently skips (returns `nil`) |

---

## Error Types

| Sentinel | Type | Trigger | `Is*` Helper |
|----------|------|---------|--------------|
| `ErrCircuitOpen` | `*CircuitBreakerError` | `allowCall()` returns `false` in OPEN/HALF-OPEN | `IsCircuitOpenError(err)` |
| `ErrTimeout` | `*CircuitBreakerError` | `executeWithTimeout` context deadline exceeded | `IsTimeoutError(err)` |

Both values implement the `error` interface. Identity comparison (`err == ErrCircuitOpen`) is valid because they are package-level pointers.

---

## HTTP Status Codes (Compliance Endpoints)

| Condition | HTTP Status | Body |
|-----------|-------------|------|
| Circuit CLOSED, OSA responds 200 | `200 OK` | Normal JSON response |
| Circuit CLOSED, OSA unavailable | `503 Service Unavailable` | `{"error": "..."}` |
| Circuit OPEN — method has fallback | `200 OK` | Degraded/cached JSON response |
| Circuit OPEN — no fallback, error propagated | `503 Service Unavailable` | `{"error": "circuit breaker protected ..."}` |
| Call exceeds `TimeoutDuration` | `503 Service Unavailable` | `{"error": "request timeout"}` |

---

## Health Endpoint

`CircuitBreakerWrapper.CircuitBreakerHealthHandler()` returns a Gin handler mounted on the compliance health route. Response shape:

```json
{
  "status": "healthy | degraded | recovering",
  "timestamp": "<RFC3339>",
  "circuit_breaker": {
    "state": 0,
    "success_rate": 98.5,
    "total_calls": 1042,
    "successful_calls": 1026,
    "failed_calls": 16,
    "timeout_calls": 3,
    "consecutive_failures": 0,
    "last_failure": "<RFC3339 | zero>",
    "next_retry_delay": 0
  },
  "compliance_service": {
    "osa_base_url": "http://osa:8089",
    "last_refresh": "<RFC3339>"
  }
}
```

| `status` value | Circuit state |
|----------------|---------------|
| `"healthy"` | `StateClosed` |
| `"degraded"` | `StateOpen` |
| `"recovering"` | `StateHalfOpen` |

---

## Monitoring Callbacks

Set via fluent methods on `*CircuitBreaker`. All are optional.

| Method | Callback Signature | Fired When |
|--------|--------------------|------------|
| `OnStateChange(fn)` | `func(oldState, newState State)` | Any state transition |
| `OnFailure(fn)` | `func(error)` | Each failed call |
| `OnSuccess(fn)` | `func()` | Each successful call |
| `OnTimeout(fn)` | `func()` | Each timeout (note: increments `timeoutCalls` counter but current `onTimeout` hook is stored and not called from `executeWithTimeout`; `OnFailure` fires instead via `recordResult`) |

---

## OTEL Attributes (Semconv)

Defined in `desktop/backend-go/internal/semconv/healing_attributes.go` and `healing_span_names.go`.

| Attribute Key | Go Constant | Type | Description |
|---------------|-------------|------|-------------|
| `healing.circuit_breaker.state` | `HealingCircuitBreakerStateKey` | `string` | Current state: `"closed"`, `"open"`, `"half_open"` |
| `healing.circuit_breaker.failure_count` | `HealingCircuitBreakerFailureCountKey` | `int64` | Consecutive failures that triggered the circuit |
| `healing.circuit_breaker.call_count` | `HealingCircuitBreakerCallCountKey` | `int64` | Total calls in the current window |
| `healing.circuit_breaker.reset_ms` | `HealingCircuitBreakerResetMsKey` | `int64` | Time in ms before half-open reset attempt |

| Span Name | Go Constant | Description |
|-----------|-------------|-------------|
| `healing.circuit_breaker.trip` | `HealingCircuitBreakerTripSpan` | Emitted when circuit transitions to OPEN |

State value enum: `HealingCircuitBreakerStateValues.Closed = "closed"`, `.Open = "open"`, `.HalfOpen = "half_open"`.

---

## Example: Reading Circuit State

```go
import "github.com/rhl/businessos-backend/internal/circuitbreaker"

cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.ComplianceServiceConfig())

state := cb.GetState()
switch state {
case circuitbreaker.StateClosed:
    // Normal path
case circuitbreaker.StateOpen:
    delay := cb.GetNextRetryDelay()
    // Wait delay before retry
case circuitbreaker.StateHalfOpen:
    // Probe in progress
}

stats := cb.GetStats()
// stats.SuccessRate, stats.ConsecutiveFailures, stats.LastFailure
```

To reset programmatically (e.g. after operator intervention):

```go
cb.Reset() // transitions to StateClosed, zeroes counters
```

---

## Builder API

`circuitbreaker.NewBuilder()` provides a fluent alternative to `NewCircuitBreaker`:

| Method | Sets |
|--------|------|
| `WithConfig(Config)` | All fields at once |
| `WithMaxAttempts(int)` | `maxAttempts` |
| `WithBackoff(base, max time.Duration)` | `baseDelay`, `maxDelay` |
| `WithTimeout(time.Duration)` | `timeoutDuration` |
| `WithCooldown(time.Duration)` | `cooldownPeriod` |
| `Build()` | Returns `*CircuitBreaker` |

---

## See Also

- [`hipaa-compliance-validator.md`](./hipaa-compliance-validator.md) — HIPAA rule validator used within the compliance service
- [`api-endpoints.md`](./api-endpoints.md) — Full compliance API endpoint reference
- [`configuration-options.md`](./configuration-options.md) — Server-wide environment variables including `OSA_BASE_URL`
- `desktop/backend-go/internal/circuitbreaker/circuit_breaker.go` — Core implementation
- `desktop/backend-go/internal/services/compliance_service_circuitbreaker.go` — Compliance wrapper with fallbacks
