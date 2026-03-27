# Go Test Tagging Guide — BusinessOS Backend

How to tag Go tests for fast feedback loops and selective execution.

## Quick Start

### Run only unit tests (fast)
```bash
go test ./... -tags=unit -v
```

### Run only integration tests
```bash
go test ./... -tags=integration -v
```

### Run only short tests (<100ms each)
```bash
go test ./... -short -v
```

### Run all tests with race detector
```bash
go test ./... -race
```

---

## Test File Structure

### Unit Test (Fast, No External Dependencies)

```go
// auth_test.go
package auth

import "testing"

func TestValidateToken(t *testing.T) {
    token := "valid-token-123"
    valid := ValidateToken(token)
    if !valid {
        t.Errorf("expected valid, got invalid")
    }
}
```

**Run:**
```bash
go test ./internal/auth/... -short -v
```

---

### Integration Test (Requires External Service/Database)

```go
// auth_integration_test.go
// +build integration

package auth

import (
    "testing"
    "github.com/rhl/businessos-backend/internal/database"
)

func TestAuthWithDB(t *testing.T) {
    // Integration test — requires database
    db := database.NewTestDB(t)
    // ... test with real database
}
```

**Run:**
```bash
go test ./... -tags=integration -v
```

---

## Build Tags

### Conditional Compilation Based on Tag

```go
// +build unit

package mypackage

import "testing"

func TestUnitOnly(t *testing.T) {
    // This runs ONLY with: go test -tags=unit
}
```

### Multi-Tag Tests

```go
// +build !short

package mypackage

func TestLongRunning(t *testing.T) {
    // This test is SKIPPED with go test -short
}
```

---

## Tag Categories

| Tag | Meaning | When to Use | Run Command |
|-----|---------|------------|-------------|
| `unit` | No external dependencies | Pure logic tests | `go test -tags=unit` |
| `integration` | Requires external services | Database, API, cache | `go test -tags=integration` |
| `short` (built-in) | Fast tests (<100ms) | For quick feedback | `go test -short` |
| `race` (built-in) | Race detector enabled | Finding concurrency bugs | `go test -race` |

---

## Common Patterns

### Skip Long-Running Test by Default

```go
func TestExpensiveAlgorithm(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping expensive test in short mode")
    }
    // ... expensive test
}
```

**Run:**
```bash
go test -short ./...                    # Skipped
go test ./...                           # Runs
```

---

### Database Tests (Integration Tag)

```go
// +build integration

package handlers_test

import (
    "testing"
    "github.com/rhl/businessos-backend/internal/database"
)

func TestCreateUserInDB(t *testing.T) {
    db := database.NewTestDB(t)
    defer db.Close()

    user := CreateUser(db, "test@example.com")
    if user.ID == 0 {
        t.Fatal("expected user ID > 0")
    }
}
```

**Run:**
```bash
go test -tags=integration ./internal/handlers/...
```

---

### API Mock Tests (Unit Tag)

```go
// +build unit

package handlers_test

import (
    "testing"
    "net/http/httptest"
)

func TestLoginHandler(t *testing.T) {
    // No external dependencies — unit test
    req := httptest.NewRequest("POST", "/login", nil)
    w := httptest.NewRecorder()

    LoginHandler(w, req)
    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", w.Code)
    }
}
```

**Run:**
```bash
go test -tags=unit ./internal/handlers/...
```

---

## Testing Best Practices

### 1. Pure Logic Tests (Unit)
```go
func TestCalculation(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("expected 5, got %d", result)
    }
}
// No setup, no mocking, <1ms
```

### 2. Mocked External Services (Unit)
```go
func TestAPIWithMock(t *testing.T) {
    mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    }))
    defer mockAPI.Close()

    client := NewClient(mockAPI.URL)
    status, err := client.Check()
    if status != "ok" || err != nil {
        t.Errorf("expected ok, got %s, %v", status, err)
    }
}
// Fast (<100ms), deterministic, no external dependency
```

### 3. Real Database (Integration)
```go
// +build integration

func TestQueryUser(t *testing.T) {
    db := database.NewTestDB(t)
    defer db.Close()

    user := db.GetUser(1)
    if user.ID != 1 {
        t.Errorf("expected user 1")
    }
}
// Slower (>100ms), real database, requires setup
```

---

## Typical Workflow

### During Development (Fast Feedback)

```bash
# Make changes
echo "func NewFeature() { ... }" >> internal/feature/feature.go

# Quick test
go test ./internal/feature/... -short -v

# If green, run all unit tests
go test ./... -tags=unit -v

# If green, run full test suite
go test ./...
```

### Before Commit

```bash
# Run race detector
go test ./... -race

# Run full suite
go test ./...

# Check coverage
go test ./... -cover
```

### CI/CD Pipeline

```bash
# Phase 1: Fast feedback (1 min)
go test ./... -short -v

# Phase 2: Unit tests (2 min)
go test ./... -tags=unit -v

# Phase 3: Full validation (5 min)
go test ./... -race

# Phase 4: Coverage (2 min)
go test ./... -cover
```

---

## Test File Organization

```
internal/
├── handlers/
│   ├── handlers.go              # Implementation
│   ├── handlers_test.go         # Unit tests (default)
│   └── handlers_integration_test.go  # Integration tests (+build integration)
├── auth/
│   ├── auth.go
│   └── auth_test.go             # Default behavior: unit tests
└── database/
    ├── database.go
    ├── database_unit_test.go    # Explicit unit tests
    └── database_integration_test.go  # Explicit integration tests
```

---

## Run Examples

### Fast Feedback (5 seconds)
```bash
go test ./... -short -v
```

### Unit Tests Only (10 seconds)
```bash
go test ./... -tags=unit -v
```

### Integration Tests Only (30 seconds)
```bash
go test ./... -tags=integration -v
```

### All Tests (45 seconds)
```bash
go test ./...
```

### With Race Detector (60 seconds)
```bash
go test ./... -race
```

### Single Package
```bash
go test ./internal/handlers/... -v
```

### Single Test
```bash
go test ./internal/handlers/... -run TestCreateUser -v
```

### With Coverage
```bash
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out  # Open in browser
```

---

## Helpful Commands

### List all tests
```bash
go test ./... -list '.*'
```

### Show test output on success
```bash
go test ./... -v
```

### Verbose with panic details
```bash
go test ./... -v -race
```

### Count occurrences of a pattern
```bash
go test ./... -list '.*Create.*' | wc -l
```

### Run tests in parallel (different default)
```bash
go test ./... -parallel 10
```

---

## Troubleshooting

### Tests fail intermittently (flaky)
```bash
# Run multiple times
go test ./... -count=10

# Or with race detector
go test ./... -race -count=5
```

### One test hangs
```bash
# Run with timeout (requires custom runner)
timeout 30s go test ./... -run TestName

# Or run single test
go test -run TestName -timeout 5s
```

### Need to see println output
```bash
go test ./... -v -run TestName
```

### Test passes locally but fails in CI
```bash
# Try with race detector
go test -race ./...

# Try in different order
go test -p 1 ./...  # Sequential, not parallel
```

---

## Further Reading

- **Go Testing Official:** https://golang.org/pkg/testing/
- **BuildTags:** https://golang.org/cmd/go/#hdr-Build_constraints
- **Go Test Flags:** `go test -h`
- **BusinessOS Backend:** `internal/`

---

**Last Updated:** 2026-03-25
**Maintained by:** ChatmanGPT Build System
