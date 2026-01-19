# Pure Go Voice Agent - Integration Testing Guide

This guide explains how to run and write integration tests for the Pure Go Voice Agent with LiveKit.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Test Structure](#test-structure)
- [Running Tests Locally](#running-tests-locally)
- [Running Tests in CI/CD](#running-tests-in-cicd)
- [Writing New Integration Tests](#writing-new-integration-tests)
- [Troubleshooting](#troubleshooting)
- [Performance Benchmarks](#performance-benchmarks)

## Overview

Integration tests validate the Pure Go Voice Agent's interaction with real LiveKit infrastructure:

- **Unit tests** (67 tests): Mock all dependencies, fast execution (<1s)
- **Integration tests** (20+ tests): Real LiveKit server, slower execution (30s-5min)

### What Integration Tests Cover

| Test Category | Tests | Description |
|---------------|-------|-------------|
| **Room Connection** | 5 | Join, leave, multiple rooms, reconnection |
| **Audio Tracks** | 3 | Subscribe, publish, RTP packet reading |
| **End-to-End Voice** | 4 | STT, TTS, full conversation, multi-turn |
| **Performance** | 3 | Latency, concurrent sessions, memory usage |
| **Error Handling** | 5 | Network failure, missing room, track errors |

## Prerequisites

### Local Development

1. **Docker Desktop** (or Docker Engine)
   ```bash
   docker --version  # Should be 20.x or higher
   ```

2. **Go 1.24+**
   ```bash
   go version
   ```

3. **LiveKit CLI** (optional, for debugging)
   ```bash
   brew install livekit-cli  # macOS
   # or download from https://github.com/livekit/livekit-cli
   ```

### Environment Variables

Required for integration tests:

```bash
export INTEGRATION_TEST=true
export LIVEKIT_URL=ws://localhost:7880
export LIVEKIT_API_KEY=test-key
export LIVEKIT_API_SECRET=test-secret
```

Optional (skip tests):

```bash
export SKIP_INTEGRATION_TESTS=true
```

## Quick Start

### Run All Integration Tests

```bash
cd desktop/backend-go
./scripts/run_integration_tests.sh
```

This script will:
1. Start LiveKit server in Docker
2. Wait for server to be ready
3. Run integration tests
4. Show results
5. Clean up containers

### Run Specific Test

```bash
# Set environment variables
export INTEGRATION_TEST=true
export LIVEKIT_URL=ws://localhost:7880
export LIVEKIT_API_KEY=test-key
export LIVEKIT_API_SECRET=test-secret

# Start LiveKit manually
docker-compose -f docker-compose.test.yml up -d livekit

# Run specific test
go test -tags=integration -v ./internal/livekit/ -run TestIntegration_JoinRoom

# Cleanup
docker-compose -f docker-compose.test.yml down
```

## Test Structure

### Build Tag

All integration tests use the `//go:build integration` build tag:

```go
//go:build integration

package livekit

func TestIntegration_JoinRoom(t *testing.T) {
    if !isLiveKitAvailable(t) {
        t.Skip("LiveKit server not available")
    }
    // Test implementation
}
```

This prevents integration tests from running during normal `go test ./...`

### Test Files

```
internal/livekit/
├── voice_agent_integration_test.go    # Main integration tests
├── integration_helpers_test.go        # Helper functions
├── voice_agent_test.go                # Unit tests
├── voice_agent_mocks_test.go          # Mock objects
└── ...
```

### Helper Functions

```go
// Check if LiveKit is available
isLiveKitAvailable(t) bool

// Create test room (auto-cleanup)
roomName, cleanup := createTestRoom(t)
defer cleanup()

// Join as user participant
userRoom, cleanup := joinRoomAsUser(t, roomName, "user-123", "Test User")
defer cleanup()

// Setup integration agent
agent, cleanup := setupIntegrationAgent(t)
defer cleanup()

// Wait helpers
waitForTrackPublished(ctx, roomName, participantID, timeout)
waitForParticipantJoin(ctx, roomName, participantID, timeout)

// Audio generation
samples := generateTestAudioTrack(durationSec, sampleRate, frequency)
silence := generateSilence(durationSec, sampleRate)
```

## Running Tests Locally

### Option 1: Automated Script (Recommended)

```bash
./scripts/run_integration_tests.sh
```

**Options:**
- `--verbose, -v`: Show detailed test output
- `--skip-cleanup`: Leave containers running after tests
- `--help, -h`: Show help message

**Example with verbose output:**
```bash
./scripts/run_integration_tests.sh --verbose
```

### Option 2: Manual Control

**1. Start LiveKit:**
```bash
docker-compose -f docker-compose.test.yml up -d livekit
```

**2. Wait for LiveKit to be ready:**
```bash
# Check logs
docker-compose -f docker-compose.test.yml logs -f livekit

# Or use health check
docker-compose -f docker-compose.test.yml ps
```

**3. Run tests:**
```bash
export INTEGRATION_TEST=true
export LIVEKIT_URL=ws://localhost:7880
export LIVEKIT_API_KEY=test-key
export LIVEKIT_API_SECRET=test-secret

go test -tags=integration -v -timeout=10m ./internal/livekit/...
```

**4. Cleanup:**
```bash
docker-compose -f docker-compose.test.yml down -v
```

### Option 3: Run with Docker Compose Services

If you also need Redis or PostgreSQL:

```bash
docker-compose -f docker-compose.test.yml up -d
export INTEGRATION_TEST=true
export LIVEKIT_URL=ws://localhost:7880
export LIVEKIT_API_KEY=test-key
export LIVEKIT_API_SECRET=test-secret
export REDIS_URL=redis://localhost:6379
export DATABASE_URL=postgres://test:test@localhost:5432/businessos_test

go test -tags=integration -v ./internal/livekit/...
docker-compose -f docker-compose.test.yml down -v
```

## Running Tests in CI/CD

### GitHub Actions

Integration tests run automatically on:
- Pushes to `main` branch
- PRs with `integration` label
- Manual workflow dispatch

**To run on a PR:**
1. Add the `integration` label to your PR
2. Tests will run on next push

**Workflow file:** `.github/workflows/integration-tests.yml`

### Local CI Simulation

Test the CI workflow locally:

```bash
# Install act (GitHub Actions local runner)
brew install act  # macOS

# Run integration tests workflow
act -j integration-tests \
  -s INTEGRATION_TEST=true \
  -s LIVEKIT_URL=ws://localhost:7880 \
  -s LIVEKIT_API_KEY=test-key \
  -s LIVEKIT_API_SECRET=test-secret
```

## Writing New Integration Tests

### Template

```go
//go:build integration

package livekit

func TestIntegration_YourFeature(t *testing.T) {
    if !isLiveKitAvailable(t) {
        t.Skip("LiveKit server not available")
    }

    // Arrange
    agent, cleanup := setupIntegrationAgent(t)
    defer cleanup()

    roomName, roomCleanup := createTestRoom(t)
    defer roomCleanup()

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Act
    err := agent.JoinRoom(ctx, roomName, "test-user", "Test User")

    // Assert
    require.NoError(t, err)
    assert.Contains(t, agent.GetActiveRooms(), roomName)
}
```

### Best Practices

**1. Always check LiveKit availability:**
```go
if !isLiveKitAvailable(t) {
    t.Skip("LiveKit server not available")
}
```

**2. Use timeouts:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

**3. Clean up resources:**
```go
roomName, cleanup := createTestRoom(t)
defer cleanup()  // Always defer cleanup
```

**4. Use descriptive test names:**
```go
func TestIntegration_JoinRoom_WhenDuplicateJoin_ShouldHandleGracefully(t *testing.T)
```

**5. Skip tests that require unavailable services:**
```go
if !isWhisperAvailable() {
    t.Skip("Requires Whisper service")
}
```

**6. Test timeout should be generous:**
```go
// Voice operations are slow - allow 30s per test
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
```

## Troubleshooting

### LiveKit Server Not Starting

**Check Docker:**
```bash
docker ps
docker logs livekit-test
```

**Common issues:**
- Port 7880 already in use: `lsof -i :7880`
- Docker daemon not running: Start Docker Desktop
- Insufficient memory: Increase Docker memory limit

### Tests Timing Out

**Increase test timeout:**
```bash
go test -tags=integration -timeout=20m ./internal/livekit/...
```

**Check LiveKit health:**
```bash
curl http://localhost:7881/
```

### Connection Refused

**Verify environment variables:**
```bash
echo $LIVEKIT_URL
echo $LIVEKIT_API_KEY
```

**Test LiveKit connection:**
```bash
livekit-cli list-rooms --url ws://localhost:7880 --api-key test-key --api-secret test-secret
```

### Tests Pass Locally but Fail in CI

**Check GitHub Actions logs:**
1. Go to Actions tab in GitHub
2. Click on failing workflow
3. Expand "Run integration tests" step

**Common CI issues:**
- Different timeout values
- Network configuration differences
- Service startup timing

**Debug in CI:**
Add debug output to workflow:
```yaml
- name: Debug LiveKit
  run: |
    curl -v http://localhost:7881/
    docker logs livekit-test
```

## Performance Benchmarks

### Expected Test Execution Times

| Test Type | Duration | Description |
|-----------|----------|-------------|
| Room Connection | 2-5s | Join/leave operations |
| Audio Track | 5-10s | Track subscription/publishing |
| E2E Voice | 30-60s | Full STT→LLM→TTS pipeline |
| Concurrent Sessions | 30-45s | 10 simultaneous sessions |

### Target Metrics

- **Latency**: End-to-end < 3000ms (STT + LLM + TTS)
- **Memory**: < 40MB per session
- **Concurrent Sessions**: 200+ supported
- **Test Suite**: < 5 minutes total

### Profiling Integration Tests

```bash
# CPU profile
go test -tags=integration -cpuprofile=cpu.prof ./internal/livekit/...
go tool pprof cpu.prof

# Memory profile
go test -tags=integration -memprofile=mem.prof ./internal/livekit/...
go tool pprof mem.prof
```

## Continuous Improvement

### Adding More Tests

Priority areas for new integration tests:
1. **VAD edge cases** - Voice Activity Detection boundary conditions
2. **Network resilience** - Packet loss, jitter, reconnection
3. **Long conversations** - 10+ turn conversations
4. **Stress testing** - 50+ concurrent sessions
5. **Audio quality** - SNR, bitrate, codec verification

### Metrics to Track

Monitor these metrics over time:
- Test execution time (should stay < 5 min)
- Flaky test rate (target: 0%)
- Test coverage (integration + unit > 80%)
- CI success rate (target: > 95%)

## Resources

- [LiveKit Documentation](https://docs.livekit.io/)
- [LiveKit Server SDK Go](https://github.com/livekit/server-sdk-go)
- [Go Testing Guide](https://go.dev/doc/tutorial/add-a-test)
- [Docker Compose Documentation](https://docs.docker.com/compose/)

## Support

For issues or questions:
1. Check existing tests in `voice_agent_integration_test.go`
2. Review logs: `docker-compose -f docker-compose.test.yml logs`
3. Test manually with LiveKit CLI
4. Ask in #backend-go channel

---

Last updated: January 2026
