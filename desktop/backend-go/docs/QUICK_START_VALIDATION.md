# Quick Start: Environment Validation

## 5-Minute Setup

### 1. Configure Environment

```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go

# Copy example config
cp .env.example .env

# Edit with your API keys
nano .env

# Required to set:
# - DATABASE_URL (postgres://...)
# - SECRET_KEY (32+ characters)
# - AI_PROVIDER (anthropic, groq, etc)
# - OPENAI_API_KEY (sk-...)
# - ELEVENLABS_API_KEY (sk_...)
# - LIVEKIT_URL (wss://...)
# - LIVEKIT_API_KEY (AP...)
# - LIVEKIT_API_SECRET (...)
```

### 2. Load Variables

```bash
source .env
```

### 3. Validate Environment

```bash
./scripts/validate_environment.sh
```

### Expected Success Output

```
✅ SET: DATABASE_URL
✅ SET: SECRET_KEY
✅ SET: AI_PROVIDER
✅ SET: OPENAI_API_KEY
✅ SET: ELEVENLABS_API_KEY
✅ SET: LIVEKIT_URL
✅ SET: LIVEKIT_API_KEY
✅ SET: LIVEKIT_API_SECRET

✅ Go installed: go1.25.0

✅ Dependency: github.com/livekit/server-sdk-go/v2
✅ Dependency: github.com/jackc/pgx/v5
✅ Dependency: gopkg.in/hraban/opus.v2

Errors: 0
Warnings: 0

✅ Environment is ready for voice system!
```

### 4. Start Server (if validation passes)

```bash
go run ./cmd/server
```

---

## Troubleshooting Quick Links

| Issue | Command | Doc |
|-------|---------|-----|
| Missing variables | `echo $VARIABLE_NAME` | ENVIRONMENT_SETUP.md |
| Go not found | `brew install go` | ENVIRONMENT_SETUP.md |
| DB connection error | `psql $DATABASE_URL -c "SELECT 1"` | ENVIRONMENT_VALIDATION_GUIDE.md |
| API key invalid | `curl -H "Authorization: Bearer $OPENAI_API_KEY" https://api.openai.com/v1/models` | VOICE_ENVIRONMENT_CONFIG.md |
| LiveKit unreachable | `curl $LIVEKIT_URL` | ENVIRONMENT_VALIDATION_GUIDE.md |

---

## Common Issues & Quick Fixes

### "❌ MISSING (REQUIRED): OPENAI_API_KEY"
```bash
# 1. Check if .env is loaded
echo $OPENAI_API_KEY

# 2. If empty, load it
source .env

# 3. If still empty, check .env file
grep OPENAI_API_KEY .env

# 4. If missing, add it
echo "OPENAI_API_KEY=sk-..." >> .env
source .env
```

### "❌ Go not installed"
```bash
# Install Go
brew install go

# Verify installation
go version  # Should show go1.21 or higher
```

### "⚠️  Database connection failed"
```bash
# This is OK for development without a database
# To test with database:

# 1. Start PostgreSQL
brew services start postgresql@15

# 2. Create database
createdb businessos

# 3. Run migrations
go run ./cmd/migrate

# 4. Run validation again
./scripts/validate_environment.sh
```

### "⚠️  LiveKit server not reachable"
```bash
# This is OK for initial setup
# To run with LiveKit:

# 1. Start LiveKit Docker
docker run -d \
  -p 7880:7880 \
  -p 7881:7881 \
  -p 7882:7882 \
  -p 50051:50051 \
  livekit/livekit-server --dev

# 2. Update .env
echo "LIVEKIT_URL=ws://localhost:7880" >> .env
source .env

# 3. Run validation again
./scripts/validate_environment.sh
```

---

## API Key Acquisition

### OpenAI (Whisper - Speech-to-Text)
1. Go to https://platform.openai.com/api-keys
2. Create new API key
3. Copy to `OPENAI_API_KEY` in `.env`

### ElevenLabs (TTS - Text-to-Speech)
1. Go to https://elevenlabs.io/app/voice-lab
2. Sign up or login
3. Navigate to API Keys
4. Copy API key to `ELEVENLABS_API_KEY` in `.env`
5. Choose voice ID from Voice Library
6. Set `ELEVENLABS_VOICE_ID` (default: Rachel)

### LiveKit (Real-time Communication)
**Option A: Local (Development)**
```bash
docker run -d \
  -p 7880:7880 \
  livekit/livekit-server --dev

# Use in .env:
LIVEKIT_URL=ws://localhost:7880
LIVEKIT_API_KEY=devkey
LIVEKIT_API_SECRET=secret
```

**Option B: Cloud (https://cloud.livekit.io)**
1. Create account
2. Create project
3. Get credentials from Settings
4. Copy to `.env`

---

## Development Checklist

- [ ] Copy `.env.example` to `.env`
- [ ] Fill in all required variables
- [ ] Run `source .env`
- [ ] Run `./scripts/validate_environment.sh`
- [ ] Verify 0 errors (warnings OK)
- [ ] Run `go run ./cmd/server`
- [ ] Check logs for errors
- [ ] Test voice endpoint: `curl http://localhost:8080/api/voice/health`

---

## Documentation Map

```
Quick Help:
└─ This file (you are here)

Setup & Configuration:
├─ ENVIRONMENT_SETUP.md         ← Complete setup guide
├─ VOICE_ENVIRONMENT_CONFIG.md  ← Voice-specific config
└─ .env.example                 ← Example environment file

Validation & Troubleshooting:
├─ ENVIRONMENT_VALIDATION_GUIDE.md  ← How to fix issues
└─ VOICE_VALIDATION_TESTS.md        ← Test scenarios

Reference:
└─ TEST_TRACK_3_SUMMARY.md      ← Complete overview
```

---

## One-Liner Test

```bash
source .env && ./scripts/validate_environment.sh && echo "✅ Ready to go!" || echo "❌ Fix errors above"
```

---

## For Production

Before deploying:

1. Use production API keys
2. Use production database URL
3. Use production LiveKit instance
4. Set `LOG_LEVEL=info` (not debug)
5. Ensure all required variables set
6. Run validation one final time
7. Monitor logs after startup

```bash
# Final validation before deployment
./scripts/validate_environment.sh

# Check for 0 errors
echo "Exit code: $?"  # Must be 0
```

---

## Support

- Setup issues: See `ENVIRONMENT_SETUP.md`
- Validation issues: See `ENVIRONMENT_VALIDATION_GUIDE.md`
- Voice config: See `VOICE_ENVIRONMENT_CONFIG.md`
- Test scenarios: See `VOICE_VALIDATION_TESTS.md`

---

## Quick Stats

- **Setup time**: 5 minutes
- **Required variables**: 8 (plus optional)
- **System checks**: 5 categories
- **Documentation**: 6 files
- **Test scenarios**: 12

---

**Status**: ✅ Ready to use
**Last Updated**: January 18, 2026
**Version**: 1.0
