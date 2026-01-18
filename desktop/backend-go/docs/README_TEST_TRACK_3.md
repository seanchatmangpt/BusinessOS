# Test Track 3: Environment Validation - Complete Index

## Overview

Test Track 3 implements a comprehensive environment validation system for the BusinessOS voice system. This ensures all dependencies, services, and configurations are properly set up before development or deployment.

## Quick Start

```bash
# 1. Configure environment
source .env

# 2. Run validation
./scripts/validate_environment.sh

# 3. Fix any errors using guides below
```

## Documentation Files

### For New Users (Start Here)

1. **QUICK_START_VALIDATION.md** (5.5K)
   - 5-minute setup procedure
   - Quick troubleshooting fixes
   - API key shortcuts
   - One-liner validation command
   - **Read this first if you're new**

### Core Setup & Configuration

2. **ENVIRONMENT_SETUP.md** (7.4K)
   - Complete variable reference (11 variables)
   - Step-by-step setup instructions
   - Database configuration
   - API key acquisition
   - Development vs. production
   - Comprehensive troubleshooting

3. **VOICE_ENVIRONMENT_CONFIG.md** (5.9K)
   - Voice system-specific setup
   - OpenAI/Whisper configuration
   - ElevenLabs/TTS configuration
   - LiveKit setup (self-hosted & cloud)
   - Voice selection guide
   - Cost considerations

### Validation & Troubleshooting

4. **ENVIRONMENT_VALIDATION_GUIDE.md** (7.8K)
   - How to run validation script
   - Output interpretation
   - Common issues & solutions
   - Validation checklist
   - Detailed checks explained
   - CI/CD integration

5. **VOICE_VALIDATION_TESTS.md** (9.7K)
   - 12 test scenarios
   - Expected outputs
   - Manual testing procedures
   - Automated test examples
   - Performance validation

### Reference & Summary

6. **TEST_TRACK_3_SUMMARY.md** (11K)
   - Complete overview
   - All deliverables listed
   - Variables validated
   - System checks explained
   - Test results
   - Success metrics

7. **README_TEST_TRACK_3.md** (this file)
   - Documentation index
   - Navigation guide
   - Quick reference

## Validation Script

**File**: `scripts/validate_environment.sh` (3.6K)

**What it checks**:
- 8 required environment variables
- 3 optional environment variables
- Go installation (>= 1.21)
- Critical Go dependencies
- PostgreSQL connectivity
- LiveKit server connectivity

**Usage**:
```bash
./scripts/validate_environment.sh
```

**Exit codes**:
- 0 = Success (ready to run)
- 1 = Errors found (fix before running)

## Environment Variables

### Required (Must Have)
```
DATABASE_URL              PostgreSQL connection string
SECRET_KEY                JWT signing key (32+ chars)
AI_PROVIDER               anthropic, groq, ollama_local, ollama_cloud
OPENAI_API_KEY            Speech-to-text (Whisper)
ELEVENLABS_API_KEY        Text-to-speech
LIVEKIT_URL               WebSocket server URL
LIVEKIT_API_KEY           API key for token generation
LIVEKIT_API_SECRET        API secret
```

### Optional (Nice to Have)
```
OLLAMA_URL                Ollama API endpoint
REDIS_URL                 Redis for caching/pub-sub
GRPC_VOICE_PORT           gRPC port (default: 50051)
```

## Common Tasks

### I'm setting up for the first time
1. Read: `QUICK_START_VALIDATION.md`
2. Read: `VOICE_ENVIRONMENT_CONFIG.md`
3. Run: `./scripts/validate_environment.sh`

### I'm getting validation errors
1. Read: `ENVIRONMENT_VALIDATION_GUIDE.md`
2. Check the "Common Issues" section
3. Apply the suggested fix

### I need to configure the database
1. Read: `ENVIRONMENT_SETUP.md` → Database section
2. Set `DATABASE_URL` in `.env`
3. Run: `go run ./cmd/migrate`

### I need to set up voice services
1. Read: `VOICE_ENVIRONMENT_CONFIG.md`
2. Get API keys from:
   - OpenAI: https://platform.openai.com/api-keys
   - ElevenLabs: https://elevenlabs.io/app/voice-lab
   - LiveKit: https://cloud.livekit.io
3. Add to `.env` and validate

### I'm deploying to production
1. Read: `ENVIRONMENT_SETUP.md` → Production setup
2. Read: `TEST_TRACK_3_SUMMARY.md` → Next steps
3. Update `.env` with production values
4. Run: `./scripts/validate_environment.sh`
5. Verify: No errors, all variables set

## Documentation Map

```
Test Track 3 Documentation
├── Quick Start (5 min)
│   └── QUICK_START_VALIDATION.md
│
├── Setup & Configuration
│   ├── ENVIRONMENT_SETUP.md (complete reference)
│   └── VOICE_ENVIRONMENT_CONFIG.md (voice-specific)
│
├── Validation & Testing
│   ├── ENVIRONMENT_VALIDATION_GUIDE.md (how to fix issues)
│   └── VOICE_VALIDATION_TESTS.md (test scenarios)
│
└── Reference
    ├── TEST_TRACK_3_SUMMARY.md (complete overview)
    └── README_TEST_TRACK_3.md (this file)
```

## File Organization

```
/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/

scripts/
└── validate_environment.sh (3.6K)
    The validation script

docs/
├── QUICK_START_VALIDATION.md (5.5K)
├── ENVIRONMENT_SETUP.md (7.4K)
├── VOICE_ENVIRONMENT_CONFIG.md (5.9K)
├── ENVIRONMENT_VALIDATION_GUIDE.md (7.8K)
├── VOICE_VALIDATION_TESTS.md (9.7K)
├── TEST_TRACK_3_SUMMARY.md (11K)
└── README_TEST_TRACK_3.md (this file)
```

## Quick Reference Commands

```bash
# Load environment variables
source .env

# Run validation
./scripts/validate_environment.sh

# Check specific variable
echo $OPENAI_API_KEY

# Test OpenAI API
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY" | head

# Test ElevenLabs API
curl https://api.elevenlabs.io/v1/voices \
  -H "xi-api-key: $ELEVENLABS_API_KEY" | head

# Test database
psql $DATABASE_URL -c "SELECT 1"

# Test LiveKit (if running)
curl $LIVEKIT_URL/health

# Start LiveKit Docker
docker run -d \
  -p 7880:7880 \
  -p 7881:7881 \
  -p 7882:7882 \
  -p 50051:50051 \
  livekit/livekit-server --dev

# Start server
go run ./cmd/server

# Run migrations
go run ./cmd/migrate
```

## Troubleshooting Quick Links

| Problem | Solution |
|---------|----------|
| Variables not loading | Try `source .env` |
| Go not found | Run `brew install go` |
| Database connection error | See ENVIRONMENT_SETUP.md → Troubleshooting |
| API key invalid | See VOICE_ENVIRONMENT_CONFIG.md → API Testing |
| LiveKit unreachable | See ENVIRONMENT_VALIDATION_GUIDE.md → LiveKit Issues |

## Success Checklist

Before running voice system:

- [ ] Environment variables configured (`.env`)
- [ ] Validation script passes: `./scripts/validate_environment.sh`
- [ ] All required variables have values
- [ ] Warnings about optional services are acceptable
- [ ] Can start server: `go run ./cmd/server`
- [ ] Server logs show no errors

## Contact & Support

For issues:

1. **Setup questions**: See `ENVIRONMENT_SETUP.md`
2. **Validation errors**: See `ENVIRONMENT_VALIDATION_GUIDE.md`
3. **Voice configuration**: See `VOICE_ENVIRONMENT_CONFIG.md`
4. **Test scenarios**: See `VOICE_VALIDATION_TESTS.md`
5. **General questions**: See `TEST_TRACK_3_SUMMARY.md`

## Version Information

- **Created**: January 18, 2026
- **Go Version Required**: >= 1.21 (tested with 1.25.0)
- **Voice System**: Production-ready
- **Status**: Complete

## Summary Statistics

- **Total Documentation**: 7 files, ~55 KB
- **Validation Script**: 1 file, 3.6 KB
- **Environment Variables**: 11 total (8 required, 3 optional)
- **System Checks**: 5 categories
- **Test Scenarios**: 12 documented
- **Quick Start Time**: 5 minutes

## Next Steps

1. **Immediate**: Read `QUICK_START_VALIDATION.md`
2. **Setup**: Follow steps in `VOICE_ENVIRONMENT_CONFIG.md`
3. **Validation**: Run `./scripts/validate_environment.sh`
4. **Development**: Start coding with voice system
5. **Deployment**: Follow `ENVIRONMENT_SETUP.md` production section

---

**Status**: ✅ Complete and Production-Ready

Start with `QUICK_START_VALIDATION.md` if you're new!
