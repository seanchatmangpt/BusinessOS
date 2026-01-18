# Environment Validation Guide

This guide walks through running and interpreting the environment validation script for the BusinessOS voice system.

## Quick Start

```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
./scripts/validate_environment.sh
```

## Validation Script Overview

The `scripts/validate_environment.sh` script checks:

1. **Environment Variables** - All required and optional voice system variables
2. **Go Environment** - Go version and installation
3. **Go Dependencies** - Critical packages in `go.mod`
4. **Database Connection** - PostgreSQL connectivity
5. **LiveKit Server** - WebSocket connectivity

## Output Interpretation

### Success Output

When all checks pass:

```
✅ Environment is ready for voice system!
```

### Typical Output Structure

```
🔍 BusinessOS Voice System - Environment Validation
==================================================

📋 Environment Variables Check:
-------------------------------
✅ SET: DATABASE_URL
✅ SET: SECRET_KEY
❌ MISSING (REQUIRED): OPENAI_API_KEY
⚠️  MISSING (OPTIONAL): REDIS_URL

🔧 Go Environment:
------------------
✅ Go installed: go1.25.0

📦 Go Dependencies:
-------------------
✅ Dependency: github.com/livekit/server-sdk-go/v2
✅ Dependency: github.com/jackc/pgx/v5
❌ Missing: gopkg.in/hraban/opus.v2

🗄️  Database Connection:
------------------------
✅ Database connection successful

🎙️  LiveKit Server:
-------------------
⚠️  LiveKit server not reachable

📊 Summary:
-----------
Errors: 1
Warnings: 1
```

## Fixing Common Issues

### Missing Required Environment Variables

**Problem:**
```
❌ MISSING (REQUIRED): OPENAI_API_KEY
❌ MISSING (REQUIRED): ELEVENLABS_API_KEY
```

**Solution:**

1. Check `.env` file exists:
```bash
ls -la .env
```

2. Load `.env` into current shell:
```bash
source .env
```

3. Verify variable is set:
```bash
echo $OPENAI_API_KEY
```

4. If still empty, add to `.env`:
```bash
echo "OPENAI_API_KEY=sk-..." >> .env
source .env
```

### Go Not Installed

**Problem:**
```
❌ Go not installed
```

**Solution:**

```bash
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Go Version Too Old

**Problem:**
```
⚠️  Go version might be too old (need >= 1.21)
```

**Solution:**

Update to Go 1.25.0 or later:

```bash
# Check current version
go version

# Update (macOS)
brew upgrade go

# Update (Linux) - download and replace installation
# See https://golang.org/doc/install
```

### Missing Go Dependencies

**Problem:**
```
❌ Missing: github.com/livekit/server-sdk-go/v2
```

**Solution:**

```bash
# Download missing dependencies
go mod tidy

# Verify all dependencies
go mod verify
```

### Database Connection Failed

**Problem:**
```
⚠️  Database connection failed (might be permissions)
```

**Causes:**
1. PostgreSQL service not running
2. Wrong connection string
3. Database doesn't exist
4. Wrong password

**Solutions:**

```bash
# Start PostgreSQL (macOS)
brew services start postgresql@15

# Test connection manually
psql $DATABASE_URL -c "SELECT 1"

# Check connection string format
echo $DATABASE_URL

# Create database if missing
createdb businessos

# Run migrations
go run ./cmd/migrate
```

### LiveKit Server Not Reachable

**Problem:**
```
⚠️  LiveKit server not reachable
```

**Causes:**
1. LiveKit service not running
2. Wrong URL
3. Network connectivity issue
4. Firewall blocking

**Solutions:**

```bash
# Check LiveKit URL
echo $LIVEKIT_URL

# Test with curl
curl -i "${LIVEKIT_URL/wss/https}"

# If using Docker, check container is running
docker ps | grep livekit

# Start LiveKit Docker container
docker run -d \
  -p 7880:7880 \
  -p 7881:7881 \
  -p 7882:7882 \
  -p 50051:50051 \
  livekit/livekit-server --dev

# Update LIVEKIT_URL in .env
echo "LIVEKIT_URL=ws://localhost:7880" >> .env
```

### psql Not Installed

**Problem:**
```
⚠️  psql not installed, skipping DB test
```

**Solution:**

Install PostgreSQL client:

```bash
# macOS
brew install postgresql

# Ubuntu/Debian
sudo apt-get install postgresql-client

# Then rerun validation
./scripts/validate_environment.sh
```

## Validation Checklist

### Before Running Voice System

Use this checklist to ensure all systems are ready:

- [ ] `.env` file exists: `ls -la .env`
- [ ] Environment variables loaded: `source .env`
- [ ] Validation script passes: `./scripts/validate_environment.sh`
- [ ] Go version >= 1.21: `go version`
- [ ] Database connected: `psql $DATABASE_URL -c "SELECT 1"`
- [ ] LiveKit reachable: `curl $LIVEKIT_URL/health`
- [ ] API keys valid:
  - [ ] OpenAI: Test with `curl https://api.openai.com/v1/models -H "Authorization: Bearer $OPENAI_API_KEY"`
  - [ ] ElevenLabs: Test with `curl https://api.elevenlabs.io/v1/voices -H "xi-api-key: $ELEVENLABS_API_KEY"`

## Detailed Checks

### Environment Variables Check

Validates all required and optional variables:

```
Required (must be set):
  - DATABASE_URL
  - SECRET_KEY
  - AI_PROVIDER
  - OPENAI_API_KEY
  - ELEVENLABS_API_KEY
  - LIVEKIT_URL
  - LIVEKIT_API_KEY
  - LIVEKIT_API_SECRET

Optional (nice to have):
  - OLLAMA_URL
  - REDIS_URL
  - GRPC_VOICE_PORT
```

### Go Environment Check

Verifies:
- Go is installed
- Go version >= 1.21
- Correct Go path configuration

```bash
# Manual check
go version
which go
$GOPATH
```

### Go Dependencies Check

Verifies critical packages:
- `github.com/livekit/server-sdk-go/v2` - LiveKit integration
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `gopkg.in/hraban/opus.v2` - Opus codec for audio

```bash
# Manual check
go list -m github.com/livekit/server-sdk-go/v2
go list -m github.com/jackc/pgx/v5
go list -m gopkg.in/hraban/opus.v2
```

### Database Connection Check

Tests PostgreSQL connectivity:

```bash
# Manual check
psql $DATABASE_URL -c "SELECT version();"
psql $DATABASE_URL -c "SELECT COUNT(*) FROM information_schema.tables;"
```

### LiveKit Server Check

Tests WebSocket connectivity:

```bash
# Manual check
curl -v $LIVEKIT_URL

# Or convert to HTTP for curl
curl -v "${LIVEKIT_URL/wss/https}"

# Docker verification
docker ps | grep livekit
docker logs <container-id>
```

## Running After Setup

After initial setup, run validation before each development session:

```bash
# Quick validation
./scripts/validate_environment.sh

# Full diagnostics
./scripts/validate_environment.sh && echo "✅ Ready to develop"
```

## CI/CD Integration

The validation script can be integrated into CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Validate Environment
  run: |
    chmod +x ./scripts/validate_environment.sh
    ./scripts/validate_environment.sh
```

## Troubleshooting Tips

1. **Always run from correct directory:**
   ```bash
   cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
   ./scripts/validate_environment.sh
   ```

2. **Load environment variables first:**
   ```bash
   source .env
   ./scripts/validate_environment.sh
   ```

3. **Check exit code:**
   ```bash
   ./scripts/validate_environment.sh
   echo "Exit code: $?"
   ```
   - Exit code 0 = All checks passed
   - Exit code 1 = Errors found

4. **Run specific checks manually:**
   ```bash
   # Check variable
   echo $OPENAI_API_KEY

   # Check service
   curl $LIVEKIT_URL

   # Check package
   go list -m github.com/livekit/server-sdk-go/v2
   ```

## Next Steps

After validation passes:

1. Start the backend server:
   ```bash
   go run ./cmd/server
   ```

2. Test voice endpoints:
   ```bash
   curl http://localhost:8080/api/voice/health
   ```

3. Check logs:
   ```bash
   # Server logs should show voice system initialized
   ```

## Support

For issues:
1. Review this guide's troubleshooting section
2. Check the detailed documentation: `docs/ENVIRONMENT_SETUP.md`
3. Check the voice system documentation: `docs/VOICE_ENVIRONMENT_CONFIG.md`
4. Review specific service documentation (OpenAI, ElevenLabs, LiveKit)
