# Voice Agent Troubleshooting

**Updated**: 2026-01-20 (Post-Refactor)
**Architecture**: Go Backend with Tools + Python Voice I/O

---

## 🔍 Quick Diagnostics

### Check System Status

```bash
# 1. Backend running?
curl http://localhost:8001/api/health
# Response: anything (404 is OK - server is running)

# 2. Python agent registered?
tail -20 /tmp/agent-test-final.log | grep "registered worker"
# Expected: registered worker {"agent_name": "groq-agent"}

# 3. Recent errors?
tail -50 /tmp/backend-final-tools.log | grep -i error
tail -50 /tmp/agent-test-final.log | grep -i error
```

---

## 🐛 Common Issues (Post-Refactor)

### Issue 1: Tools Not Being Called

**Symptoms**:
- OSA responds normally but doesn't execute actions
- No tasks created in database
- Navigation doesn't work

**Diagnosis**:
```bash
# Check if tools are being called
tail -100 /tmp/backend-final-tools.log | grep "tool"
```

**Expected Output**:
```
INFO 🔧 LLM requested tool calls count=1
INFO ⚙️ Executing tool tool=navigate_to_module
INFO ✅ Tool executed
```

**Fixes**:

1. **Tool definitions not sent to Groq**:
   - Check: `voice_chat_tools.go` line 65-73
   - Verify: `toolDefinitions := tools.GetAllTools()`

2. **Groq model doesn't support tools**:
   - Check: `voice_chat_tools.go` line 69
   - Must be: `"mixtral-8x7b-32768"` (not llama3)

3. **Tool execution failing**:
   - Check: `executor.go` for database errors
   - Verify: User ID is being passed correctly

### Issue 2: Context Not Working (OSA Doesn't Know User Name)

**Symptoms**:
- OSA says "I don't know your name"
- No user-specific information available

**Diagnosis**:
```bash
# Check session lookup
tail -100 /tmp/backend-final-tools.log | grep "session"

# Check voice_sessions table
psql $DATABASE_URL -c "SELECT session_id, user_id FROM voice_sessions ORDER BY created_at DESC LIMIT 5;"
```

**Fixes**:

1. **Session not linked to user**:
   ```sql
   -- Verify session exists
   SELECT * FROM voice_sessions WHERE session_id = 'your-session-id';
   ```

2. **Context enrichment failing**:
   - Check: `voice_chat.go` line 201-227
   - Look for: "Context enrichment" in logs

3. **User not found**:
   ```sql
   -- Verify user exists
   SELECT id, name FROM users WHERE id = 'user-id-here';
   ```

### Issue 3: Tasks Not Being Created in Database

**Symptoms**:
- OSA says "I've created a task" but nothing in database
- No errors in logs

**Diagnosis**:
```bash
# Check for SQL errors
tail -100 /tmp/backend-final-tools.log | grep -E "(SQL|database|INSERT)"

# Check recent tasks
psql $DATABASE_URL -c "SELECT id, title, created_at FROM tasks ORDER BY created_at DESC LIMIT 5;"
```

**Fixes**:

1. **SQL query failing**:
   - Check: `executor.go` line 126-141
   - Look for: "Failed to create task" error

2. **User ID not being passed**:
   - Check: `voice_chat_tools.go` line 153
   - Verify: `userID` is not empty

3. **Test direct database insert**:
   ```sql
   INSERT INTO tasks (user_id, title, status)
   VALUES ('user-id-here', 'test task', 'todo')
   RETURNING id, title;
   ```

### Issue 4: Python Agent Not Receiving Jobs

**Symptoms**:
- Python agent shows "registered worker" but never processes rooms
- Logs show: `DEBUG: No job available`

**Diagnosis**:
```bash
# Check agent registration
tail -50 /tmp/agent-test-final.log | grep "registered"

# Check LiveKit dispatch
tail -100 /tmp/backend-final-tools.log | grep -i dispatch
```

**Fixes**:

1. **Agent name mismatch**:
   - Python: Must be `groq-agent` (check `agent_groq.py`)
   - Go dispatch: Must request `groq-agent` (check `livekit/client.go` line 222)

2. **LiveKit credentials wrong**:
   ```bash
   # Verify credentials
   cd /Users/rhl/Desktop/BusinessOS2/voice-agent
   cat .env | grep LIVEKIT
   ```

3. **Room not being created**:
   - Check: `livekit/client.go` line 47-83
   - Look for: "CreateRoom succeeded" in logs

### Issue 5: Audio Not Playing

**Symptoms**:
- Transcripts appear but no audio
- Frontend connects but silent

**Diagnosis**:
```bash
# Check TTS processing
tail -50 /tmp/agent-test-final.log | grep -i "tts"

# Check browser console for errors
# (Open browser DevTools)
```

**Fixes**:

1. **ElevenLabs API key missing**:
   ```bash
   cd /Users/rhl/Desktop/BusinessOS2/voice-agent
   cat .env | grep ELEVENLABS_API_KEY
   # Should not be empty
   ```

2. **Audio track not subscribed**:
   - Check browser console: Should see "TrackSubscribed" events
   - Check Python logs: Should see "TTS: Playing audio"

3. **Browser autoplay blocked**:
   - User must click page before connecting voice orb
   - Check console for: "Audio play failed: NotAllowedError"

### Issue 6: Deprecated Code Still Running

**Symptoms**:
- Pattern matching still happening
- Personality not using Go system prompt
- Tools not being called

**Diagnosis**:
```bash
# Check if deprecated code still present
grep -n "detect_navigation_command" /Users/rhl/Desktop/BusinessOS2/voice-agent/agent_groq.py
grep -n "from personality import" /Users/rhl/Desktop/BusinessOS2/voice-agent/agent_groq.py
```

**Fix**: See `CLEANUP_AND_TESTING.md` for removal instructions.

---

## 🔧 Service Management

### Restart Backend

```bash
# Kill old process
ps aux | grep "[s]erver-refactored" | awk '{print $2}' | xargs kill

# Start new
/tmp/server-refactored > /tmp/backend-final-tools.log 2>&1 &

# Verify
curl http://localhost:8001/api/health
```

### Restart Python Agent

```bash
# Kill old process
pkill -f agent_groq.py

# Start new
cd /Users/rhl/Desktop/BusinessOS2/voice-agent
python3 agent_groq.py dev > /tmp/agent-test-final.log 2>&1 &

# Verify
tail -20 /tmp/agent-test-final.log | grep "registered"
```

### Check Logs in Real-Time

```bash
# Backend logs (tool execution)
tail -f /tmp/backend-final-tools.log | grep --line-buffered -E "(Tool|tool_call|Error)"

# Python logs (audio processing)
tail -f /tmp/agent-test-final.log | grep --line-buffered -E "(STT|TTS|Error)"
```

---

## 📊 Verification Commands

### Database Queries

```sql
-- Check recent tasks
SELECT id, title, status, created_at
FROM tasks
ORDER BY created_at DESC
LIMIT 10;

-- Check voice sessions
SELECT session_id, user_id, state, created_at
FROM voice_sessions
ORDER BY created_at DESC
LIMIT 5;

-- Check user context
SELECT id, name, email
FROM users
WHERE id = 'user-id-here';

-- Check if task was created from voice
SELECT id, title, created_at
FROM tasks
WHERE title ILIKE '%milk%'
ORDER BY created_at DESC;
```

### HTTP Requests (Test Backend Directly)

```bash
# Test backend without voice
curl -X POST http://localhost:8001/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": "create a task called test"}],
    "session_id": "test-session-123"
  }'

# Expected response:
# {"response": "I've created a task called 'test'..."}
```

---

## 🎯 Expected vs Actual Behavior

### When User Says "Open Tasks"

#### Expected:
```
Backend Logs:
  INFO 🤖 Calling Groq API with tools
  INFO 🔧 LLM requested tool calls count=1
  INFO ⚙️ Executing tool tool=navigate_to_module args={"module":"tasks"}
  INFO ✅ Tool executed result=Opened tasks module
  INFO 🔄 Calling LLM again with tool results
  INFO ✅ Final response

Python Logs:
  INFO STT: "open tasks"
  INFO Backend response received
  INFO TTS: Playing audio
```

#### If Something's Wrong:
- No tool calls → Check `voice_chat_tools.go` integration
- Tool not executing → Check `executor.go` implementation
- No audio → Check ElevenLabs API key

### When User Says "Create a Task Called Buy Milk"

#### Expected:
```
Backend Logs:
  INFO 🔧 LLM requested tool calls count=1
  INFO ⚙️ Executing tool tool=create_task args={"title":"buy milk"}
  INFO ✅ Creating task title=buy milk
  INFO ✅ Tool executed result=Created task: buy milk (ID: abc123)

Database:
  SELECT * FROM tasks WHERE title = 'buy milk';
  -- Returns 1 row
```

#### If Something's Wrong:
- Task not created → Check `executor.go` line 93-141
- SQL error → Check database schema matches code
- User ID wrong → Check session lookup

---

## 🔄 Architecture Flow (Reference)

```
USER SPEAKS
    ↓
[Python Agent] VAD + STT
    ↓
POST /api/chat → [Go Backend]
    ↓
[voice_chat.go] Session lookup, context enrichment
    ↓
[voice_chat_tools.go] Call Groq with tools
    ↓
[Groq LLM] Returns tool calls
    ↓
[executor.go] Execute tools (DB writes, navigation)
    ↓
[voice_chat_tools.go] Call Groq again with results
    ↓
Response → [Python Agent]
    ↓
[TTS] ElevenLabs
    ↓
USER HEARS
```

**Key Files**:
- `voice_chat.go` - Session + context
- `voice_chat_tools.go` - Groq API with function calling
- `definitions.go` - Tool definitions
- `executor.go` - Tool execution

---

## 📚 Related Documentation

- `CLEANUP_AND_TESTING.md` - Full testing guide
- `QUICK_TEST.md` - 5-minute test plan
- `REFACTOR_COMPLETE.md` - What changed
- `VOICE_SYSTEM_ARCHITECTURE.md` - Complete architecture
- `DEBUGGING_GUIDE.md` - Deep debugging

---

## 🚨 Emergency Reset

If everything is broken:

```bash
# 1. Kill all processes
pkill -f agent_groq.py
pkill -f server-refactored

# 2. Rebuild backend
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
go build -o /tmp/server-refactored ./cmd/server

# 3. Restart backend
/tmp/server-refactored > /tmp/backend-final-tools.log 2>&1 &

# 4. Restart Python agent
cd /Users/rhl/Desktop/BusinessOS2/voice-agent
python3 agent_groq.py dev > /tmp/agent-test-final.log 2>&1 &

# 5. Verify both running
curl http://localhost:8001/api/health
tail -20 /tmp/agent-test-final.log | grep "registered"

# 6. Check logs for errors
tail -50 /tmp/backend-final-tools.log | grep -i error
tail -50 /tmp/agent-test-final.log | grep -i error
```

---

**Last Updated**: 2026-01-20 02:36
**Architecture**: Refactored (Go tools + Python I/O)
**Status**: Post-refactor troubleshooting guide
