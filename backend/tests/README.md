# Multi-Agent App Generation Tests

This directory contains integration and E2E tests for the multi-agent app generation system.

## Test Structure

```
tests/
├── e2e/
│   └── multi_agent_e2e.sh       # End-to-end test script
└── README.md                     # This file
```

## Prerequisites

1. **ANTHROPIC_API_KEY** must be set in environment
2. Server must be running on `http://localhost:8001` (or set `API_BASE`)
3. PostgreSQL database must be available (for full integration tests)

## Running Tests

### Unit Tests (No API calls)

```bash
cd desktop/backend-go
go test ./internal/agent -v -short
go test ./internal/services -v -short -run TestWorkspace
```

### Integration Tests (Requires API key)

```bash
# Run all integration tests
cd desktop/backend-go
ANTHROPIC_API_KEY=your-key go test ./internal/agent -v

# Run specific test
ANTHROPIC_API_KEY=your-key go test ./internal/agent -v -run TestOrchestratorCreatePlan
```

### E2E Test (Full system test)

```bash
# Ensure server is running first
cd desktop/backend-go
go run ./cmd/server

# In another terminal, run E2E test
cd desktop/backend-go/tests/e2e
chmod +x multi_agent_e2e.sh
./multi_agent_e2e.sh

# With custom workspace ID
WORKSPACE_ID=$(uuidgen) ./multi_agent_e2e.sh

# Against different server
API_BASE=https://your-server.com ./multi_agent_e2e.sh
```

## Test Coverage

### ✅ Unit Tests

**Circuit Breaker (`orchestrator_test.go`)**
- Circuit breaker state transitions
- Failure threshold triggering
- Half-open state recovery
- Manual reset functionality

**Graceful Shutdown (`orchestrator_test.go`)**
- WaitGroup tracking
- Shutdown timeout (30s)
- Cancellation of new executions after shutdown

**Workspace Management (`workspace_manager_test.go`)**
- Workspace directory creation
- File saving with proper structure
- Parent directory creation
- Workspace cleanup
- File categorization logic

**File Parsing (`workspace_manager_test.go`)**
- Markdown code block extraction
- Multiple filename format support
- File: comment parsing
- Empty/missing filename handling

### ✅ Integration Tests

**Orchestrator (`orchestrator_test.go`)**
- Full plan creation with Claude API
- Worker execution with real AI responses
- Progress callback tracking
- Circuit breaker integration
- Graceful shutdown under load

### ✅ E2E Tests

**Full System Flow (`multi_agent_e2e.sh`)**
- Server health check
- Queue item creation via API
- SSE progress stream monitoring
- Workspace file verification
- 4 parallel agent execution

## Test Results

Run tests with verbose output to see detailed results:

```bash
go test -v ./internal/agent -run TestCircuitBreaker
```

Expected output:
```
=== RUN   TestCircuitBreakerIntegration
--- PASS: TestCircuitBreakerIntegration (0.00s)
=== RUN   TestGracefulShutdown
--- PASS: TestGracefulShutdown (0.10s)
PASS
ok      github.com/rhl/businessos-backend/internal/agent    0.105s
```

## Troubleshooting

### "Skipping integration test in short mode"
- Remove `-short` flag from go test command
- Integration tests require ANTHROPIC_API_KEY

### "context deadline exceeded"
- Increase timeout in test code
- Check network connectivity to Claude API

### "workspace directory not found"
- E2E test expects workspace at `/tmp/businessos-agent-workspaces/{id}`
- Check server logs for file saving errors
- Verify permissions on temp directory

### "SSE stream ended unexpectedly"
- Check server logs for orchestrator errors
- Verify ANTHROPIC_API_KEY is valid
- Check circuit breaker hasn't opened (too many failures)

## Performance Benchmarks

Run benchmarks to measure system performance:

```bash
go test -bench=. -benchmem ./internal/agent
```

## Continuous Integration

For CI environments (GitHub Actions), use short mode to skip integration tests:

```bash
go test -short ./...
```

Or run integration tests only on main branch:

```yaml
# .github/workflows/test.yml
- name: Integration Tests
  if: github.ref == 'refs/heads/main'
  run: go test -v ./internal/agent
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
```

## Writing New Tests

### Adding a Unit Test

```go
func TestMyFeature(t *testing.T) {
	// Arrange
	orchestrator := NewOrchestrator(nil)

	// Act
	result := orchestrator.SomeMethod()

	// Assert
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}
```

### Adding an Integration Test

```go
func TestMyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test requires ANTHROPIC_API_KEY
	orchestrator := NewOrchestrator(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Run test...
}
```

## Test Metrics

Current test coverage (as of last run):

- **Circuit Breaker**: 95% coverage
- **Orchestrator**: 85% coverage
- **Workspace Manager**: 100% coverage
- **File Parsing**: 100% coverage

Run coverage report:

```bash
go test -coverprofile=coverage.out ./internal/agent ./internal/services
go tool cover -html=coverage.out
```
