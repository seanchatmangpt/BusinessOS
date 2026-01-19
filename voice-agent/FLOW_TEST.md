# Voice Agent Flow Test - Complete Pipeline

## Expected Flow (Your Voice → Audio Output)

### Step 1: Frontend Connects
```
[Frontend] User clicks voice orb
[Frontend] POST /api/livekit/token
[Backend] Creates room: ws_{workspace}_{user}_{timestamp}
[Backend] Dispatches agent: "groq-agent" to room
[Backend] Returns token + room info
[Frontend] Connects to LiveKit room with token
[Frontend] Publishes microphone audio track
```

**Expected Logs:**
- Backend: `INFO dispatching agent asynchronously room=ws_...`
- Frontend Console: `[LiveKit] Connected to room successfully`
- Frontend Console: `[LiveKit] Published audio track`

---

### Step 2: Agent Receives Job
```
[LiveKit Cloud] Sends job request to Python agent
[Agent] request_fnc() checks agent_name == "groq-agent"
[Agent] Accepts job
[Agent] entrypoint() runs
```

**Expected Logs:**
- Agent: `Received job request for room: ws_... agent_name: 'groq-agent'`
- Agent: `Accepting job for room: ws_...`
- Agent: `Job started for room: ws_...`
- Agent: `Connected to LiveKit room`
- Agent: `[GROQ-WHISPER MODE] Starting agent session...`
- Agent: `[GROQ-WHISPER] Agent is now listening... say something!`

---

### Step 3: You Speak → STT
```
[You] Speak into microphone
[Frontend] Sends audio via WebRTC
[Agent] Silero VAD detects speech
[Agent] Groq Whisper STT transcribes audio
[Agent] Sends transcript to frontend via data channel
[Frontend] Receives transcript, displays in LiveCaptions
```

**Expected Logs:**
- Agent: `[GROQ-WHISPER] Sent transcript: [user] Hello, can you hear me?...`
- Frontend Console: `[GROQ-WHISPER] user: Hello, can you hear me?`

---

### Step 4: LLM Processing
```
[Agent] GoBackendLLM._run() called
[Agent] POST http://localhost:8001/api/chat
[Backend] Receives chat request
[Backend] Calls Groq LLM API (llama-3.3-70b-versatile)
[Backend] Returns response text
[Agent] Receives LLM response
```

**Expected Logs:**
- Agent: `[GROQ-WHISPER-LLM] === START _run() ===`
- Agent: `[GROQ-WHISPER-LLM] Sending to backend: [...]`
- Backend: `INFO LLM chat request received`
- Agent: `[GROQ-WHISPER-LLM] Got response: Yes, I can hear you perfectly!...`

---

### Step 5: TTS Audio Generation
```
[Agent] Sends ChatChunk to TTS pipeline
[Agent] ElevenLabs TTS receives text
[ElevenLabs API] Generates audio from text
[Agent] Receives audio data from ElevenLabs
[Agent] Publishes audio track to room
```

**Expected Logs:**
- Agent: `[GROQ-WHISPER-LLM] Sending ChatChunk to TTS pipeline...`
- Agent: `[GROQ-WHISPER-LLM] ChatChunk sent successfully!`
- Agent: (LiveKit SDK logs about publishing audio track)

---

### Step 6: Frontend Audio Playback
```
[Frontend] Receives RoomEvent.TrackSubscribed
[Frontend] track.kind === 'audio'
[Frontend] Attaches audio track to <audio> element
[Frontend] Calls audioElement.play()
[Browser] Plays audio through speakers
[You] HEAR the agent's voice!
```

**Expected Logs:**
- Frontend Console: `[LiveKit] Track subscribed: audio from agent-AJ_BMkPiaFVkNR`
- Frontend Console: `[LiveKit] Agent audio track subscribed`
- Frontend Console: `[LiveKit] Audio element state: {paused: false, muted: false, volume: 1, ...}`
- Frontend Console: `[LiveKit] ✅ Audio playing successfully!`
- Frontend Console: `[LiveKit] 🔊 Audio actually playing now!`

---

## Potential Failure Points

### 1. Agent Never Receives Job
**Symptoms:** No "Job started" log in agent
**Causes:**
- Explicit dispatch not working
- request_fnc rejecting job
- LiveKit Cloud not routing job

### 2. STT Not Transcribing
**Symptoms:** No user transcript in frontend
**Causes:**
- Microphone not publishing
- VAD not detecting speech
- Groq Whisper API error

### 3. LLM Not Responding
**Symptoms:** Transcript appears but no agent response
**Causes:**
- Backend /api/chat endpoint failing
- Groq LLM API error
- Empty session_id causing crash

### 4. TTS Not Generating Audio
**Symptoms:** Agent transcript appears but no audio
**Causes:**
- ElevenLabs API error
- ChatChunk not sent to TTS
- TTS pipeline broken

### 5. Audio Not Playing
**Symptoms:** "Audio playing successfully" but no sound
**Causes:**
- Browser autoplay policy blocking
- Audio element muted
- Wrong audio output device
- MediaStream not attached properly

---

## Testing Instructions

1. **Open browser console** (F12 → Console tab)
2. **Click voice orb**
3. **Say clearly:** "Hello, can you hear me?"
4. **Check logs:**
   - Agent terminal: Look for "Job started"
   - Agent terminal: Look for "[GROQ-WHISPER-LLM]" logs
   - Browser console: Look for "[LiveKit]" logs
   - Browser console: Look for "Audio element state"
5. **Screenshot all logs if audio doesn't play**

---

## Current Status

✅ Backend running with explicit dispatch
✅ Agent registered as "groq-agent"
✅ Frontend has audio debugging logs
❓ Need to test complete flow
