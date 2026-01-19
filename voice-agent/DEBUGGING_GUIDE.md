# Voice System Debugging Guide

## 🔍 How to Debug Context Awareness Issues

### Quick Diagnosis Script

Run this to check system health:

```bash
#!/bin/bash
echo "=== VOICE SYSTEM HEALTH CHECK ==="

# 1. Check backend
echo -n "Backend: "
curl -s http://localhost:8001/api/health > /dev/null && echo "✅ Running" || echo "❌ Not running"

# 2. Check Python agent
echo -n "Python Agent: "
ps aux | grep -q "[a]gent_groq.py" && echo "✅ Running" || echo "❌ Not running"

# 3. Check database
echo "Database voice_sessions table:"
psql "$DATABASE_URL" -c "SELECT COUNT(*) as session_count FROM voice_sessions;" 2>/dev/null || echo "❌ Cannot connect"

# 4. Check logs
echo ""
echo "Recent backend LiveKit activity:"
grep -i "livekit\|dispatch" /tmp/backend-*.log 2>/dev/null | tail -5 || echo "No activity"

echo ""
echo "Recent Python agent activity:"
tail -10 /tmp/voice-*.log 2>/dev/null | grep -i "request\|job\|room" || echo "No jobs received"
```

### Log Locations

| Component | Log File | What to Look For |
|-----------|----------|------------------|
| Go Backend | `/tmp/backend-dispatch-debug.log` | "dispatching agent", "Room created", "generated LiveKit token" |
| Python Agent | `/tmp/voice-agent-dispatch-test.log` | "[REQUEST_FNC]", "Job started for room", "Using session ID" |
| Frontend | Browser DevTools Console | "[LiveKit] Connected", "[LiveKit] Published audio track" |

### Critical Checkpoints

#### ✅ Checkpoint 1: Token Generation
**Backend log should show:**
```
✅ generated LiveKit token
   user_id=abc123
   session_id=def456
   room_name=ws_abc_xyz_1234567890
```

**How to test:**
```bash
# Must be authenticated first
curl -X POST http://localhost:8001/api/livekit/token \
  -H "Content-Type: application/json" \
  -H "Cookie: session=YOUR_SESSION_TOKEN" \
  -d '{"agent_role":"groq-agent"}'
```

#### ✅ Checkpoint 2: Agent Dispatch
**Backend log should show:**
```
dispatching agent asynchronously room=ws_abc_xyz_1234567890 agent=groq-agent
agent dispatch response status=200
```

#### ✅ Checkpoint 3: Agent Request Received
**Python log should show:**
```
============================================================
[REQUEST_FNC] 🎯 JOB REQUEST RECEIVED!
[REQUEST_FNC] Room: ws_abc_xyz_1234567890
[REQUEST_FNC] Agent Name: 'groq-agent'
============================================================
Accepting job for room: ws_abc_xyz_1234567890
```

**If this is MISSING**: LiveKit is not dispatching to the agent!

#### ✅ Checkpoint 4: Agent Connects & Extracts Session
**Python log should show:**
```
Job started for room: ws_abc_xyz_1234567890
Connected to LiveKit room
[GROQ-WHISPER] ✅ Using session ID for context: def456...
```

#### ✅ Checkpoint 5: LLM Request with Context
**Python log should show:**
```
[GROQ-WHISPER-LLM] 🔍 DEBUG: session_id variable = def456...
[GROQ-WHISPER-LLM] 🔍 DEBUG: Sending POST to http://localhost:8080/api/chat
```

**Backend log should show:**
```
✅ voice chat with session auth session_id=def456...
✅ Found voice session user_id=abc123
✅ Successfully loaded user from session user_name="Roberto Huacuja Luna"
```

### Common Issues

#### Issue 1: Python agent never receives job request
**Symptoms:**
- No `[REQUEST_FNC]` logs
- Agent registered but idle
- User connects but agent doesn't respond

**Possible Causes:**
1. **Wrong agent_name in dispatch** - Backend dispatching to "deepgram-agent" but Python is "groq-agent"
2. **LiveKit dispatch API failing** - Check backend logs for HTTP errors
3. **Agent not properly registered** - Check Python startup logs for "registered worker"

**Fix:**
```bash
# Check backend dispatch logs
grep -A 5 "dispatching agent" /tmp/backend-*.log

# Verify agent registration
grep "registered worker" /tmp/voice-*.log

# Test dispatch manually via LiveKit API
curl -X POST https://macstudiosystems-yn61tekm.livekit.cloud/twirp/livekit.AgentDispatchService/CreateDispatch \
  -H "Authorization: Bearer $LIVEKIT_TOKEN" \
  -d '{"room":"test_room","agent_name":"groq-agent"}'
```

#### Issue 2: session_id is empty at backend
**Symptoms:**
- Python log shows: `session_id variable = None`
- Backend log shows: `session_id from request body value=""`
- OSA doesn't know user's name

**Possible Causes:**
1. **Room metadata not set** - Backend didn't attach metadata to room
2. **Python not extracting metadata** - JSON parsing failed
3. **Wrong metadata format** - LiveKit expecting different structure

**Fix:**
```python
# Add to agent_groq.py entrypoint():
print(f"Room metadata raw: {ctx.room.metadata}")
print(f"Room metadata parsed: {json.loads(ctx.room.metadata)}")
```

#### Issue 3: Backend can't find voice session
**Symptoms:**
- Backend log: `❌ invalid session_id`
- Database query fails

**Fix:**
```bash
# Check database
psql "$DATABASE_URL" -c "SELECT * FROM voice_sessions ORDER BY created_at DESC LIMIT 5;"

# Verify session was created
grep "Voice session created" /tmp/backend-*.log | tail -5
```

### Manual Testing Flow

```bash
# 1. Start backend with logging
cd desktop/backend-go
go run ./cmd/server 2>&1 | tee /tmp/backend-test.log &

# 2. Start Python agent with logging
cd voice-agent
python agent_groq.py dev 2>&1 | tee /tmp/agent-test.log &

# 3. Tail both logs in separate terminals
tail -f /tmp/backend-test.log
tail -f /tmp/agent-test.log

# 4. Connect voice orb in browser

# 5. Watch for all 5 checkpoints above
```

### Debugging Dispatch Issue

If Python agent NEVER receives job requests, the issue is in the dispatch mechanism.

**Verify dispatch API call:**

```go
// Add to desktop/backend-go/internal/integrations/livekit/client.go
// In createAgentDispatch() function:

respBody, _ := ioutil.ReadAll(resp.Body)
c.logger.Info("🔍 DEBUG: Dispatch API response",
    "status", resp.StatusCode,
    "body", string(respBody),
    "room", roomName,
    "agent_name", agentName,
)
```

**Check LiveKit dashboard:**
- Go to https://cloud.livekit.io
- Check "Agents" tab
- Verify "groq-agent" is listed and online

### Environment Variable Check

```bash
# Backend .env
grep -E "LIVEKIT_URL|LIVEKIT_API_KEY" desktop/backend-go/.env

# Python .env
grep -E "LIVEKIT_URL|LIVEKIT_API_KEY|BACKEND_URL" voice-agent/.env

# Verify they match!
```

### Quick Fixes

```bash
# Reset everything
pkill -f "cmd/server"
pkill -f "agent_groq"
rm /tmp/*.log

# Clear database sessions
psql "$DATABASE_URL" -c "DELETE FROM voice_sessions WHERE created_at < NOW() - INTERVAL '1 hour';"

# Restart fresh
cd desktop/backend-go && go run ./cmd/server > /tmp/backend.log 2>&1 &
cd voice-agent && python agent_groq.py dev > /tmp/agent.log 2>&1 &

# Test connection
# Click voice orb, watch logs
```

---

**Still broken?** Check `VOICE_SYSTEM_ARCHITECTURE.md` for complete data flow diagram.
