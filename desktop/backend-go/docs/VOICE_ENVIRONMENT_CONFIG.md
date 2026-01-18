# Voice System Environment Configuration

This document describes the additional environment variables needed for the voice system beyond the standard BusinessOS configuration.

## Voice System Environment Variables

Add these to your `.env` file for voice system functionality:

### Audio Processing (Speech-to-Text & Text-to-Speech)

```bash
# OpenAI API key for speech-to-text (Whisper)
# Used for converting user audio to text
# Get from: https://platform.openai.com/api-keys
OPENAI_API_KEY=sk-proj-your-openai-key-here

# ElevenLabs API key for text-to-speech
# Used for converting AI responses to natural audio
# Get from: https://elevenlabs.io/app/voice-lab
ELEVENLABS_API_KEY=sk_your-elevenlabs-key-here

# ElevenLabs voice ID (default: Rachel)
# Available voices: Rachel, Adam, Bella, Charlie, Ethan, Freya, George, Grace, Harry, Liam, Mimi, Naomi, Ollie, Patricia, River, Roman, Sam, Sarah, Stella, Steve, Thomas
# List all voices: curl -H "xi-api-key: $ELEVENLABS_API_KEY" https://api.elevenlabs.io/v1/voices
# Or via web: https://elevenlabs.io/app/voice-lab
ELEVENLABS_VOICE_ID=21m00Tcm4TlvDq8ikWAM
```

### Real-time Communication (LiveKit)

```bash
# LiveKit server WebSocket URL
# Self-hosted: wss://your-livekit-server.com
# Managed (free): wss://[project].livekit.cloud
LIVEKIT_URL=wss://livekit.example.com

# LiveKit API key (for token generation)
# Get from LiveKit dashboard → Settings → API Keys
# For local testing: APxxxxxxxxxxxxxxxxxxxxx
LIVEKIT_API_KEY=APxxxxxxxxxxxxxxxxxxxxx

# LiveKit API secret
# Get from LiveKit dashboard → Settings → API Keys
# For local testing: your-secret-here
LIVEKIT_API_SECRET=your-livekit-secret-here
```

### Voice Service Configuration

```bash
# gRPC port for voice services (optional)
# Default: 50051
GRPC_VOICE_PORT=50051
```

## Setup Instructions

### Step 1: Get API Keys

#### OpenAI (Whisper)
1. Visit https://platform.openai.com/api-keys
2. Create new API key
3. Copy key to `OPENAI_API_KEY`

#### ElevenLabs (Text-to-Speech)
1. Visit https://elevenlabs.io/app/voice-lab
2. Sign up or login
3. Go to Voice Library to explore voices
4. Go to API Keys (in profile)
5. Copy API key to `ELEVENLABS_API_KEY`
6. Choose a voice ID and set `ELEVENLABS_VOICE_ID`

#### LiveKit
**Option A: Self-Hosted** (recommended for development)
```bash
# Using Docker
docker run -d \
  -p 7880:7880 \
  -p 7881:7881 \
  -p 7882:7882 \
  -p 50051:50051 \
  livekit/livekit-server \
  --dev

# Set environment variables
export LIVEKIT_URL=ws://localhost:7880
export LIVEKIT_API_KEY=devkey
export LIVEKIT_API_SECRET=secret
```

**Option B: LiveKit Cloud (free tier)**
1. Visit https://cloud.livekit.io
2. Create account and project
3. Get API credentials from Settings
4. Copy to `LIVEKIT_URL`, `LIVEKIT_API_KEY`, `LIVEKIT_API_SECRET`

### Step 2: Update .env

```bash
# Add to your .env file
cat >> .env << 'EOF'

# Voice System
OPENAI_API_KEY=sk-proj-your-key-here
ELEVENLABS_API_KEY=sk_your-key-here
ELEVENLABS_VOICE_ID=21m00Tcm4TlvDq8ikWAM
LIVEKIT_URL=wss://your-livekit-server.com
LIVEKIT_API_KEY=APxxxxxx
LIVEKIT_API_SECRET=your-secret
GRPC_VOICE_PORT=50051
EOF
```

### Step 3: Validate

```bash
./scripts/validate_environment.sh
```

Expected output:
```
✅ SET: OPENAI_API_KEY
✅ SET: ELEVENLABS_API_KEY
✅ SET: ELEVENLABS_VOICE_ID
✅ SET: LIVEKIT_URL
✅ SET: LIVEKIT_API_KEY
✅ SET: LIVEKIT_API_SECRET
```

## Voice Selection

### Available ElevenLabs Voices

Get the list of available voices:

```bash
curl -H "xi-api-key: $ELEVENLABS_API_KEY" \
  https://api.elevenlabs.io/v1/voices | jq '.voices[] | {name, voice_id}'
```

Popular voices:
- **Rachel** (21m00Tcm4TlvDq8ikWAM) - Friendly, warm female voice
- **Adam** (pFAgoVWLHKQ9Rm0R8XnC) - Clear male voice
- **Bella** (EXAVITQu4vr4xnSDxMaL) - Professional female voice
- **Charlie** (IZSifLlHbfdyhBlXOcZl) - Energetic male voice
- **Ethan** (g5CIjZEefAQxTkQ5MS1H) - Calm male voice

Set the voice ID:
```bash
export ELEVENLABS_VOICE_ID=pFAgoVWLHKQ9Rm0R8XnC  # Adam
```

## Cost Considerations

### OpenAI Whisper
- Pay-as-you-go: $0.02 per minute of audio
- Or use your organization's plan if you have one

### ElevenLabs Text-to-Speech
- Free tier: 10,000 characters/month
- Paid: Starting at ~$5-10/month for 100K+ characters
- Production: Check pricing for volume discounts

### LiveKit
- Self-hosted: Free (just infrastructure costs)
- Cloud: Free tier (limited), then pay-as-you-go for transcoding

## Troubleshooting

### OpenAI Key Issues
```bash
# Test API key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY" | head -20
```

### ElevenLabs Key Issues
```bash
# Test API key
curl https://api.elevenlabs.io/v1/voices \
  -H "xi-api-key: $ELEVENLABS_API_KEY" | head -20
```

### LiveKit Connection Issues
```bash
# Test LiveKit connection (replace URL)
curl -i "$LIVEKIT_URL/health" 2>&1 | head -10

# If using docker:
docker logs <container-id>
```

## Development vs. Production

### Development Setup
```bash
# Use free/test keys
OPENAI_API_KEY=sk-proj-test-key
ELEVENLABS_API_KEY=sk_test-key
LIVEKIT_URL=ws://localhost:7880  # Local docker
LIVEKIT_API_KEY=devkey
LIVEKIT_API_SECRET=secret
```

### Production Setup
```bash
# Use production keys
OPENAI_API_KEY=sk-proj-real-key
ELEVENLABS_API_KEY=sk_real-key
LIVEKIT_URL=wss://your-managed-livekit.cloud  # Production service
LIVEKIT_API_KEY=<production-key>
LIVEKIT_API_SECRET=<production-secret>

# Store in secure vault (not in .env)
# Use environment variables or secrets management
```

## Next Steps

1. Gather all required API keys
2. Update `.env` with voice system variables
3. Run `./scripts/validate_environment.sh`
4. Test voice endpoints with sample audio
5. Configure voice preferences (voice ID, etc.)

## References

- [OpenAI Whisper API](https://platform.openai.com/docs/guides/speech-to-text)
- [ElevenLabs Documentation](https://docs.elevenlabs.io/)
- [LiveKit Documentation](https://docs.livekit.io/)
- [Voice System Architecture](./VOICE_SYSTEM_ARCHITECTURE.md)
