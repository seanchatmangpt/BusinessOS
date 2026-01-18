# Voice System Validation Test Scenarios

This document outlines various test scenarios for validating the environment before running the voice system.

## TEST SCENARIO 1: Minimal Valid Setup

**Scenario**: Run with only required environment variables set.

**Prerequisites**:
```bash
# Unset all optional variables
unset OLLAMA_URL
unset REDIS_URL
unset GRPC_VOICE_PORT
```

**Commands**:
```bash
# Set required variables
export DATABASE_URL="postgres://user:pass@localhost/businessos"
export SECRET_KEY="your-32-character-minimum-secret-key"
export AI_PROVIDER="anthropic"
export OPENAI_API_KEY="sk-proj-test-key"
export ELEVENLABS_API_KEY="sk_test-key"
export LIVEKIT_URL="wss://localhost:7880"
export LIVEKIT_API_KEY="devkey"
export LIVEKIT_API_SECRET="secret"

# Run validation
./scripts/validate_environment.sh
```

**Expected Output**:
```
✅ All required environment variables set
⚠️  Some optional variables missing (OK for development)
✅ Go 1.25.0 installed
✅ All critical dependencies found
⚠️  Database connection test skipped (OK without database)
⚠️  LiveKit not reachable (expected for local testing)
```

**Expected Exit Code**: 0 (warnings are OK)

---

## TEST SCENARIO 2: Complete Production Setup

**Scenario**: Run with all variables including optional ones.

**Prerequisites**:
```bash
# Ensure production services are running/accessible
# - PostgreSQL at production URL
# - Redis configured
# - LiveKit server running
# - All API keys valid
```

**Commands**:
```bash
# Load all variables from .env
source .env

# Run validation with verbose output
./scripts/validate_environment.sh
```

**Expected Output**:
```
✅ All environment variables set
✅ Go 1.25.0 installed
✅ All critical dependencies found
✅ Database connection successful
✅ LiveKit server reachable
✅ Environment is ready for voice system!
```

**Expected Exit Code**: 0

---

## TEST SCENARIO 3: Missing Required Variable

**Scenario**: Run with one required variable missing.

**Commands**:
```bash
export DATABASE_URL="postgres://user:pass@localhost/businessos"
export SECRET_KEY="your-32-character-minimum-secret-key"
export AI_PROVIDER="anthropic"
export OPENAI_API_KEY="sk-proj-test-key"
# ELEVENLABS_API_KEY intentionally missing
export LIVEKIT_URL="wss://localhost:7880"
export LIVEKIT_API_KEY="devkey"
export LIVEKIT_API_SECRET="secret"

./scripts/validate_environment.sh
```

**Expected Output**:
```
❌ MISSING (REQUIRED): ELEVENLABS_API_KEY
...
Errors: 1
Warnings: 0
❌ Please fix 1 error(s) before running voice system
```

**Expected Exit Code**: 1 (failure)

---

## TEST SCENARIO 4: Missing Multiple Required Variables

**Scenario**: Run with multiple required variables missing.

**Commands**:
```bash
export DATABASE_URL="postgres://user:pass@localhost/businessos"
export SECRET_KEY="your-32-character-minimum-secret-key"
# AI_PROVIDER missing
# OPENAI_API_KEY missing
# ELEVENLABS_API_KEY missing
export LIVEKIT_URL="wss://localhost:7880"
export LIVEKIT_API_KEY="devkey"
export LIVEKIT_API_SECRET="secret"

./scripts/validate_environment.sh
```

**Expected Output**:
```
❌ MISSING (REQUIRED): AI_PROVIDER
❌ MISSING (REQUIRED): OPENAI_API_KEY
❌ MISSING (REQUIRED): ELEVENLABS_API_KEY
...
Errors: 3
Warnings: 0
❌ Please fix 3 error(s) before running voice system
```

**Expected Exit Code**: 1

---

## TEST SCENARIO 5: Go Version Check

**Scenario**: Verify Go version detection works correctly.

**Commands**:
```bash
# Check detected version
go version

# Run validation to see version detection
./scripts/validate_environment.sh
```

**Expected Output**:
```
✅ Go installed: go1.25.0
```

Note: Version 1.21.0 or newer is acceptable

---

## TEST SCENARIO 6: Go Dependency Check

**Scenario**: Verify critical dependencies are detected in go.mod.

**Commands**:
```bash
# Check dependencies manually
grep "github.com/livekit/server-sdk-go/v2" go.mod
grep "github.com/jackc/pgx/v5" go.mod
grep "gopkg.in/hraban/opus.v2" go.mod

# Run validation
./scripts/validate_environment.sh
```

**Expected Output**:
```
✅ Dependency: github.com/livekit/server-sdk-go/v2
✅ Dependency: github.com/jackc/pgx/v5
✅ Dependency: gopkg.in/hraban/opus.v2
```

---

## TEST SCENARIO 7: Database Connection (with running database)

**Scenario**: Test database connectivity when PostgreSQL is running.

**Prerequisites**:
```bash
# Start PostgreSQL
brew services start postgresql@15

# Create test database
createdb businessos

# Set valid connection string
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/businessos"
```

**Commands**:
```bash
# Run validation
./scripts/validate_environment.sh
```

**Expected Output**:
```
✅ Database connection successful
```

---

## TEST SCENARIO 8: Database Connection (without database)

**Scenario**: Test graceful handling when database is unavailable.

**Prerequisites**:
```bash
# Ensure PostgreSQL is NOT running
brew services stop postgresql@15
```

**Commands**:
```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/businessos"
./scripts/validate_environment.sh
```

**Expected Output**:
```
⚠️  Database connection failed (might be permissions)
```

Note: This is a warning, not an error - OK for local development

---

## TEST SCENARIO 9: LiveKit Connection (with running server)

**Scenario**: Test LiveKit connectivity when server is running.

**Prerequisites**:
```bash
# Start LiveKit Docker container
docker run -d \
  -p 7880:7880 \
  -p 7881:7881 \
  -p 7882:7882 \
  -p 50051:50051 \
  livekit/livekit-server --dev

export LIVEKIT_URL="ws://localhost:7880"
```

**Commands**:
```bash
./scripts/validate_environment.sh
```

**Expected Output**:
```
✅ LiveKit server reachable
```

---

## TEST SCENARIO 10: LiveKit Connection (without server)

**Scenario**: Test graceful handling when LiveKit is unavailable.

**Prerequisites**:
```bash
# Ensure LiveKit is NOT running
docker ps | grep livekit  # Should be empty

export LIVEKIT_URL="wss://localhost:7880"
```

**Commands**:
```bash
./scripts/validate_environment.sh
```

**Expected Output**:
```
⚠️  LiveKit server not reachable (might be down or wrong URL)
```

Note: This is a warning - OK for initial setup

---

## TEST SCENARIO 11: All Optional Variables Set

**Scenario**: Verify optional variables don't cause errors.

**Commands**:
```bash
# Set all variables including optional
export OLLAMA_URL="http://localhost:11434"
export REDIS_URL="redis://localhost:6379"
export GRPC_VOICE_PORT="50051"

# Set required variables
export DATABASE_URL="postgres://user:pass@localhost/businessos"
export SECRET_KEY="your-32-character-minimum-secret-key"
export AI_PROVIDER="anthropic"
export OPENAI_API_KEY="sk-proj-test-key"
export ELEVENLABS_API_KEY="sk_test-key"
export LIVEKIT_URL="wss://localhost:7880"
export LIVEKIT_API_KEY="devkey"
export LIVEKIT_API_SECRET="secret"

./scripts/validate_environment.sh
```

**Expected Output**:
```
✅ SET: OLLAMA_URL
✅ SET: REDIS_URL
✅ SET: GRPC_VOICE_PORT
✅ SET: [all required variables]
```

---

## TEST SCENARIO 12: Environment Variable Loading from .env

**Scenario**: Test that validation works when variables are loaded from .env file.

**Prerequisites**:
```bash
# Create test .env file
cat > /tmp/test.env << 'EOF'
DATABASE_URL=postgres://user:pass@localhost/businessos
SECRET_KEY=your-32-character-minimum-secret-key
AI_PROVIDER=anthropic
OPENAI_API_KEY=sk-proj-test-key
ELEVENLABS_API_KEY=sk_test-key
LIVEKIT_URL=wss://localhost:7880
LIVEKIT_API_KEY=devkey
LIVEKIT_API_SECRET=secret
EOF

# Load variables
source /tmp/test.env
```

**Commands**:
```bash
./scripts/validate_environment.sh
```

**Expected Output**:
```
✅ All required variables set
✅ Environment is ready for voice system!
```

---

## Test Execution Automation

Create a test runner script:

```bash
#!/bin/bash
# run_validation_tests.sh

echo "Running validation test scenarios..."
TESTS_PASSED=0
TESTS_FAILED=0

run_test() {
    local test_name=$1
    local expected_exit_code=$2

    echo ""
    echo "Running: $test_name"

    # Run validation (implement test logic)
    if [ $? -eq $expected_exit_code ]; then
        echo "✅ PASSED: $test_name"
        ((TESTS_PASSED++))
    else
        echo "❌ FAILED: $test_name"
        ((TESTS_FAILED++))
    fi
}

# Run tests
run_test "Scenario 1: Minimal Setup" 0
run_test "Scenario 3: Missing Required" 1
# ... more tests

echo ""
echo "Summary: $TESTS_PASSED passed, $TESTS_FAILED failed"
```

---

## Manual Testing Checklist

Before deploying voice system, manually verify:

- [ ] Run validation script: `./scripts/validate_environment.sh`
- [ ] Check for 0 errors and acceptable warnings
- [ ] Test each API key manually:
  - [ ] OpenAI: `curl https://api.openai.com/v1/models -H "Authorization: Bearer $OPENAI_API_KEY"`
  - [ ] ElevenLabs: `curl https://api.elevenlabs.io/v1/voices -H "xi-api-key: $ELEVENLABS_API_KEY"`
- [ ] Test database: `psql $DATABASE_URL -c "SELECT 1"`
- [ ] Test LiveKit: `curl $LIVEKIT_URL/health` (if reachable)
- [ ] Start server: `go run ./cmd/server`
- [ ] Check logs for errors
- [ ] Test voice endpoint: `curl http://localhost:8080/api/voice/health`

---

## Performance Validation

After environment validation passes, check performance:

```bash
# Build for performance
go build -o bin/server ./cmd/server

# Check binary size
ls -lh bin/server

# Profile startup time
time ./bin/server -test

# Monitor resource usage
top -p $(pgrep -f bin/server)
```

---

## Continuous Integration

For CI/CD pipelines, use exit code:

```yaml
# GitHub Actions
- name: Validate Environment
  run: ./scripts/validate_environment.sh
  # Fails if exit code != 0
```

---

## Summary

The validation script provides comprehensive environment checking across:
- Required vs. optional variables
- Go toolchain
- Dependencies
- External services

Use these test scenarios to verify all edge cases are handled correctly.
