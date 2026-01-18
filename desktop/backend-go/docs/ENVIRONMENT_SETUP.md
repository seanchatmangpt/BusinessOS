# Environment Setup for BusinessOS Voice System

## Overview

The voice system requires a complete set of environment variables and dependencies to function properly. This guide walks through validating and setting up your environment.

## Quick Validation

Run the environment validation script to check if everything is configured:

```bash
./scripts/validate_environment.sh
```

Expected output if successful:
```
✅ Environment is ready for voice system!
```

## Required Environment Variables

### Core System

| Variable | Purpose | Required | Example |
|----------|---------|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Yes | `postgres://user:pass@localhost/businessos` |
| `SECRET_KEY` | JWT signing key (min 32 bytes) | Yes | `your-secret-key-here-min-32-chars` |
| `AI_PROVIDER` | AI model provider | Yes | `anthropic` or `ollama_local` |

### Voice System - Audio

| Variable | Purpose | Required | Example |
|----------|---------|----------|---------|
| `OPENAI_API_KEY` | OpenAI API key for speech-to-text | Yes | `sk-...` |
| `ELEVENLABS_API_KEY` | ElevenLabs API key for text-to-speech | Yes | `sk_...` |

### Voice System - Real-time Communication

| Variable | Purpose | Required | Example |
|----------|---------|----------|---------|
| `LIVEKIT_URL` | LiveKit server WebSocket URL | Yes | `wss://livekit.example.com` |
| `LIVEKIT_API_KEY` | LiveKit API key for token generation | Yes | `APxxxxx` |
| `LIVEKIT_API_SECRET` | LiveKit API secret | Yes | `secret-key-here` |

### AI Providers (choose one)

#### Anthropic
```bash
export AI_PROVIDER=anthropic
export ANTHROPIC_API_KEY=sk-ant-...
```

#### Groq
```bash
export AI_PROVIDER=groq
export GROQ_API_KEY=gsk_...
```

#### Ollama (Local)
```bash
export AI_PROVIDER=ollama_local
export OLLAMA_URL=http://localhost:11434
```

#### Ollama (Cloud)
```bash
export AI_PROVIDER=ollama_cloud
export OLLAMA_URL=https://api.ollama.example.com
```

### Optional Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `REDIS_URL` | Redis for caching/pub-sub | Not used (falls back to DB) |
| `OLLAMA_URL` | Ollama API endpoint | http://localhost:11434 |
| `GRPC_VOICE_PORT` | gRPC port for voice services | 50051 |
| `CORS_ORIGINS` | Allowed CORS origins | http://localhost:5173 |
| `LOG_LEVEL` | Logging level | info |

## Step-by-Step Setup

### 1. Create `.env` File

```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
cp .env.example .env
```

Edit `.env` with your actual values:

```bash
# Core
DATABASE_URL=postgres://user:pass@localhost/businessos
SECRET_KEY=your-super-secret-key-minimum-32-characters-long
AI_PROVIDER=anthropic

# Voice System - Audio
OPENAI_API_KEY=sk-...
ELEVENLABS_API_KEY=sk_...

# Voice System - Communication
LIVEKIT_URL=wss://your-livekit-server.com
LIVEKIT_API_KEY=APxxxxx
LIVEKIT_API_SECRET=secret-key-here

# AI Provider (Anthropic example)
ANTHROPIC_API_KEY=sk-ant-...

# Optional
REDIS_URL=redis://localhost:6379
LOG_LEVEL=debug
```

### 2. Validate Environment

```bash
./scripts/validate_environment.sh
```

Check the output:
- ✅ All required variables set
- ✅ Go 1.21+ installed
- ✅ All dependencies in go.mod
- ✅ Database connection successful (if database is running)
- ✅ LiveKit server reachable (if running)

### 3. Install System Dependencies

#### macOS
```bash
# Go (if not already installed)
brew install go

# PostgreSQL client (if you have a remote database)
brew install postgresql

# Optional: PostgreSQL server locally
brew install postgresql@15
brew services start postgresql@15
```

#### Linux (Ubuntu/Debian)
```bash
# Go
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# PostgreSQL
sudo apt-get install postgresql-client

# Optional: PostgreSQL server
sudo apt-get install postgresql
```

### 4. Verify Go Dependencies

```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go mod tidy
go mod verify
```

Expected output:
```
all modules verified
```

### 5. Test Database Connection

If using local PostgreSQL:

```bash
# Create database
createdb businessos

# Run migrations
go run ./cmd/migrate
```

If using remote database (Supabase, etc.):

```bash
# Simply set DATABASE_URL and test connection
psql $DATABASE_URL -c "SELECT version();"
```

### 6. Test Voice Dependencies

Verify critical voice packages compile:

```bash
# Check LiveKit SDK
go list -m github.com/livekit/server-sdk-go/v2

# Check Opus codec
go list -m gopkg.in/hraban/opus.v2
```

### 7. Verify API Keys

```bash
# Test OpenAI API Key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY" | head -20

# Test ElevenLabs API Key
curl https://api.elevenlabs.io/v1/voices \
  -H "xi-api-key: $ELEVENLABS_API_KEY" | head -20

# Test LiveKit Server (replace URL)
curl -i $LIVEKIT_URL/health
```

## Troubleshooting

### Missing Environment Variables

If the validation script shows missing variables:

1. Check `.env` file exists and is readable
2. Ensure variables are properly set: `export VARIABLE_NAME=value`
3. Load `.env` manually if not auto-loaded: `source .env`

### Database Connection Failed

```bash
# Check connection string format
echo $DATABASE_URL

# Test connection directly
psql $DATABASE_URL -c "SELECT 1"

# Common issues:
# - Wrong host/port
# - Missing credentials
# - Firewall blocking connection
# - Database service not running
```

### Go Dependency Issues

```bash
# Clean up module cache
go clean -modcache

# Reinstall dependencies
go mod tidy
go mod download

# Verify dependencies
go mod verify

# Check specific package
go list -m github.com/livekit/server-sdk-go/v2
```

### LiveKit Server Unreachable

```bash
# Check if LiveKit is running
curl -v wss://your-livekit-url

# If self-hosted, check:
# - Network connectivity
# - Firewall rules
# - DNS resolution

# Test with curl (convert ws:// to http://)
curl http://your-livekit-url/health
```

### API Key Issues

```bash
# Verify OpenAI key is valid
curl -s https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY" | jq .

# Verify ElevenLabs key is valid
curl -s https://api.elevenlabs.io/v1/voices \
  -H "xi-api-key: $ELEVENLABS_API_KEY" | jq .
```

## Development vs. Production

### Development Setup

```bash
# Use local services
AI_PROVIDER=ollama_local
OLLAMA_URL=http://localhost:11434

# Use local database
DATABASE_URL=postgres://postgres:postgres@localhost/businessos

# Verbose logging
LOG_LEVEL=debug
```

### Production Setup

```bash
# Use production AI provider
AI_PROVIDER=anthropic
ANTHROPIC_API_KEY=sk-ant-...

# Use production database (e.g., Supabase)
DATABASE_URL=postgres://...@aws-0-us-east-1.db.supabase.co/postgres

# Use production LiveKit
LIVEKIT_URL=wss://production-livekit.example.com

# Appropriate logging
LOG_LEVEL=info
```

## Running Voice System Tests

After environment validation:

```bash
# Run all tests
go test ./...

# Run voice system tests specifically
go test -v ./internal/handlers/voice_...

# Run with coverage
go test -cover ./...
```

## Next Steps

1. Run `./scripts/validate_environment.sh` to verify setup
2. Review output and fix any errors
3. Start the server: `go run ./cmd/server`
4. Check logs for successful startup
5. Test voice endpoints with sample requests

## Support

For issues with specific services:

- **OpenAI**: https://platform.openai.com/docs
- **ElevenLabs**: https://elevenlabs.io/docs
- **LiveKit**: https://docs.livekit.io
- **PostgreSQL**: https://www.postgresql.org/docs
- **Go**: https://golang.org/doc
