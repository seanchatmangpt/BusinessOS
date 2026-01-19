# BusinessOS Voice System Architecture

**COMPLETE DATA FLOW DOCUMENTATION**

## 🎯 Overview

The BusinessOS voice system enables real-time AI voice conversations with OSA through LiveKit WebRTC. Users speak to the voice orb in the frontend, which connects to a Python agent that processes speech, generates responses via the Go backend, and streams audio back.

## 📋 System Components

### 1. **Frontend (Svelte/TypeScript)**
- Location: `frontend/src/lib/`
- Purpose: Voice orb UI, LiveKit connection, audio playback

### 2. **Go Backend (API Server)**
- Location: `desktop/backend-go/`
- Purpose: Token generation, context enrichment, LLM orchestration

### 3. **Python Agent (Voice Processing)**
- Location: `voice-agent/`
- Purpose: STT (Groq Whisper), TTS (ElevenLabs), voice pipeline

### 4. **LiveKit Cloud (WebRTC)**
- URL: `wss://macstudiosystems-yn61tekm.livekit.cloud`
- Purpose: Real-time audio streaming, room management

### 5. **PostgreSQL (Database)**
- Tables: `voice_sessions`, `user_facts`, `user`
- Purpose: Session tracking, user context storage

---

## 🔄 COMPLETE DATA FLOW (Step-by-Step)

### PHASE 1: Connection Initialization

```
User clicks voice orb
    ↓
Frontend: livekitVoice.ts:connect()
    ↓
POST /api/livekit/token
    {
        agent_role: "groq-agent"
    }
    ↓
Go Backend: livekit.go:GenerateToken()
    ├─ Authenticates user via middleware
    ├─ Generates session_id (UUID)
    ├─ Creates room metadata:
    │   {
    │       "session_id": "abc123...",
    │       "user_id": "user_xyz",
    │       "workspace_id": "workspace_123"
    │   }
    ├─ Creates room in LiveKit with metadata
    ├─ Dispatches agent to room (async):
    │   POST https://livekit.cloud/twirp/.../CreateDispatch
    │   {
    │       "room": "ws_abc_xyz_1234567890",
    │       "agent_name": "groq-agent"
    │   }
    ├─ Saves session in voice_sessions table
    └─ Returns:
        {
            "token": "JWT...",
            "url": "wss://...",
            "room_name": "ws_abc_xyz_1234567890",
            "session_id": "abc123..."
        }
    ↓
Frontend receives token + URL
    ↓
Frontend connects to LiveKit room with token
    ↓
Frontend publishes microphone audio track
```

**Files Involved:**
- `frontend/src/lib/services/livekitVoice.ts` (lines 62-213)
- `desktop/backend-go/internal/handlers/livekit.go` (lines 53-177)
- `desktop/backend-go/internal/integrations/livekit/client.go` (lines 127-270)

---

### PHASE 2: Agent Dispatch & Session Extraction

```
LiveKit Cloud receives dispatch request
    ↓
LiveKit finds registered "groq-agent"
    ↓
LiveKit calls Python agent's request_fnc()
    ↓
Python: agent_groq.py:request_fnc() (lines 562-584)
    ├─ Checks agent_name == "groq-agent"
    ├─ Checks room not already handled
    └─ Accepts job
    ↓
LiveKit calls Python agent's entrypoint()
    ↓
Python: agent_groq.py:entrypoint() (lines 383-556)
    ├─ Connects to room: ctx.connect()
    ├─ Extracts session_id from room metadata:
    │   metadata = json.loads(ctx.room.metadata)
    │   session_id = metadata["session_id"]
    ├─ Creates AgentSession:
    │   ├─ VAD: Silero (voice activity detection)
    │   ├─ STT: Groq Whisper
    │   ├─ LLM: GoBackendLLM(session_id=session_id)
    │   └─ TTS: ElevenLabs
    └─ Starts listening: session.start(agent, room)
```

**Files Involved:**
- `voice-agent/agent_groq.py` (lines 383-556, 562-584)

---

### PHASE 3: User Speaks (Speech → Text)

```
User speaks into microphone
    ↓
Frontend publishes audio to LiveKit room
    ↓
Python Agent VAD detects speech
    ↓
Python Agent STT (Groq Whisper):
    ├─ Transcribes audio to text
    └─ Returns transcript
    ↓
Python Agent fires on_user_speech() callback (lines 509-535)
    ├─ Sends transcript to frontend:
    │   send_transcript("user", text)
    ├─ Detects navigation commands:
    │   detect_navigation_command(transcript)
    │   ├─ Matches against MODULES dict (lines 46-99)
    │   └─ Sends tool command if matched
    └─ Transcript displayed in LiveCaptions component
```

**Files Involved:**
- `voice-agent/agent_groq.py` (lines 509-535, 103-236)
- `frontend/src/lib/components/desktop3d/LiveCaptions.svelte`

---

### PHASE 4: LLM Processing (Text → Response)

```
Python Agent LLM.chat() called
    ↓
Python: GoBackendLLM.chat() (lines 268-302)
    ├─ Converts chat context to messages
    └─ Returns GoBackendLLMStream
    ↓
Python: GoBackendLLMStream._run() (lines 321-380)
    ├─ Prepares request:
    │   {
    │       "messages": [...],
    │       "session_id": "abc123..."
    │   }
    ├─ Adds header: X-Session-ID: abc123...
    └─ POST http://localhost:8080/api/chat
    ↓
Go Backend: voice_chat.go:HandleVoiceChat() (lines 37-244)
    ├─ Extracts session_id from:
    │   1. X-Session-ID header, OR
    │   2. Request body session_id field
    ├─ Looks up voice_sessions table:
    │   GetVoiceSessionBySessionID(session_id)
    ├─ Gets user details:
    │   GetUserByID(user_id)
    ├─ Builds voice context:
    │   voiceContext := BuildVoiceContext(ctx, user, workspace)
    │   ├─ User name (first name only)
    │   ├─ Workspace name
    │   ├─ Recent tasks
    │   ├─ Active projects
    │   └─ User facts from DB
    ├─ Enriches system prompt:
    │   basePrompt + voiceContext.FormatForPrompt()
    ├─ Calls Groq LLM:
    │   POST https://api.groq.com/v1/chat/completions
    │   Model: mixtral-8x7b-32768
    └─ Returns response:
        {
            "response": "Hey Roberto! I see you have 3 tasks..."
        }
    ↓
Python receives response
    ↓
Python sends agent transcript to frontend:
    send_transcript("agent", response_text)
```

**Files Involved:**
- `voice-agent/agent_groq.py` (lines 261-380)
- `desktop/backend-go/internal/handlers/voice_chat.go` (lines 37-244)
- `desktop/backend-go/internal/services/voice_context.go` (all)
- `desktop/backend-go/internal/database/queries/voice_sessions.sql`

---

### PHASE 5: Response Playback (Text → Speech → Audio)

```
Python Agent TTS (ElevenLabs):
    ├─ Converts response text to audio
    ├─ Model: eleven_turbo_v2_5
    ├─ Voice: ELEVENLABS_VOICE_ID
    └─ Settings:
        ├─ stability: 0.75 (consistent pitch)
        ├─ similarity_boost: 0.85 (voice consistency)
        └─ style: 0.0 (natural speech)
    ↓
Python Agent publishes audio track to room
    ↓
Frontend: livekitVoice.ts:RoomEvent.TrackSubscribed (lines 150-171)
    ├─ Detects agent audio track
    ├─ Attaches track to <audio> element
    ├─ Sets autoplay + volume 1.0
    ├─ Calls audioElement.play()
    └─ User hears OSA's voice
    ↓
Voice orb changes color (purple glow = speaking)
```

**Files Involved:**
- `voice-agent/agent_groq.py` (lines 463-472)
- `frontend/src/lib/services/livekitVoice.ts` (lines 150-171)
- `frontend/src/lib/components/desktop3d/VoiceOrbPanel.svelte`

---

## 📁 FILE RESPONSIBILITIES

### Frontend Files

| File | Responsibility | Lines to Know |
|------|----------------|---------------|
| `livekitVoice.ts` | LiveKit connection, token fetch, audio handling | 62-213 (connect), 150-171 (audio) |
| `VoiceOrbPanel.svelte` | Voice orb UI, drag/drop, state colors | All |
| `LiveCaptions.svelte` | Transcript display for user/agent messages | All |
| `+layout.svelte` | Mounts voice orb on all authenticated pages | 105-108 |

### Backend Files (Go)

| File | Responsibility | Lines to Know |
|------|----------------|---------------|
| `handlers/livekit.go` | Token generation, session creation | 53-177 (GenerateToken) |
| `handlers/voice_chat.go` | LLM chat with context enrichment | 37-244 (HandleVoiceChat) |
| `integrations/livekit/client.go` | LiveKit SDK wrapper, dispatch | 127-270 (GenerateRoomToken) |
| `services/voice_context.go` | Context building from DB | 24-194 (BuildVoiceContext) |
| `database/queries/voice_sessions.sql` | Session CRUD operations | All |
| `database/migrations/057_voice_sessions.sql` | Session table schema | All |

### Agent Files (Python)

| File | Responsibility | Lines to Know |
|------|----------------|---------------|
| `agent_groq.py` | Main voice agent logic | 383-556 (entrypoint), 261-380 (LLM) |
| `personality.py` | OSA personality prompt | All |

---

## 🔧 KEY CONFIGURATION

### Environment Variables

**Go Backend** (`.env`):
```bash
LIVEKIT_ENABLED=true
LIVEKIT_URL=wss://macstudiosystems-yn61tekm.livekit.cloud
LIVEKIT_API_KEY=<key>
LIVEKIT_API_SECRET=<secret>
```

**Python Agent** (`.env`):
```bash
LIVEKIT_URL=wss://macstudiosystems-yn61tekm.livekit.cloud
LIVEKIT_API_KEY=<key>
LIVEKIT_API_SECRET=<secret>
BACKEND_URL=http://localhost:8080
GROQ_API_KEY=<key>
ELEVENLABS_API_KEY=<key>
ELEVENLABS_VOICE_ID=<voice_id>
```

---

## 🐛 CURRENT ISSUE (Context Awareness Broken)

**Symptom**: OSA doesn't know user's name (Roberto)

**Expected Flow**:
1. Backend creates room with metadata containing `session_id`
2. Python agent extracts `session_id` from room metadata
3. Python agent sends `session_id` to `/api/chat`
4. Backend looks up user from `voice_sessions` table
5. Backend enriches prompt with user context

**Actual Flow**:
1. ✅ Backend creates room with metadata ✅
2. ❌ **Python agent entrypoint NEVER CALLED** ❌
3. ❌ No session extraction happens
4. ❌ Backend receives EMPTY session_id
5. ❌ No context enrichment

**Root Cause**: LiveKit is NOT dispatching jobs to the Python agent, despite the agent being registered. The `request_fnc()` and `entrypoint()` functions are never called.

---

## 🔍 DEBUGGING CHECKLIST

- [x] Backend generates token correctly
- [x] Backend creates room with metadata
- [x] Backend dispatches agent (async call to LiveKit API)
- [x] Python agent is running and registered
- [x] Database `voice_sessions` table exists
- [ ] **LiveKit actually calls Python agent** ❌ FAILING HERE
- [ ] Python agent extracts session_id from metadata
- [ ] Python agent sends session_id to backend
- [ ] Backend receives session_id and looks up context

---

## 📝 NEXT STEPS

1. **Verify LiveKit dispatch** - Check if LiveKit API call succeeds
2. **Test agent manually** - Force a room connection to trigger agent
3. **Check LiveKit dashboard** - See if agent is connected to rooms
4. **Add request_fnc logging** - Confirm if LiveKit is calling it

---

## 🚀 HOW TO RUN

```bash
# 1. Start Go backend
cd desktop/backend-go
go run ./cmd/server

# 2. Start Python agent
cd voice-agent
python agent_groq.py dev

# 3. Start frontend
cd frontend
npm run dev

# 4. Open browser to http://localhost:5173
# 5. Click voice orb (cloud icon)
# 6. Speak!
```

---

**Status**: ⚠️ Voice connection works, but context awareness broken due to agent dispatch issue.
