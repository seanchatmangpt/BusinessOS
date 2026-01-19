# 🚀 Quick Test Guide - Voice System Refactor

**Status**: ✅ Both services running, ready to test!

---

## Current Status

```bash
✅ Backend: Running on port 8001 with tools system
✅ Python Agent: Registered as "groq-agent" with LiveKit
✅ Build: Successful, no compilation errors
✅ Docs: Complete (5 markdown files)
```

---

## What Works NOW vs BEFORE

| Feature | Before | After | Status |
|---------|--------|-------|--------|
| **Voice Input** | ✅ Working | ✅ Working | No change |
| **Voice Output** | ✅ Working | ✅ Working | No change |
| **Navigation** | ❌ Pattern matching (fragile) | ✅ LLM function calling (smart) | **IMPROVED** |
| **Context Awareness** | ✅ User name, tasks | ✅ User name, tasks | No change |
| **Create Tasks** | ❌ Not possible | ✅ Voice → database | **NEW** |
| **List Tasks** | ❌ Not possible | ✅ Voice queries DB | **NEW** |
| **Tool System** | ❌ Fake (pattern matching) | ✅ Real (function calling) | **NEW** |
| **Intelligence Location** | ❌ Split (Python + Go) | ✅ Centralized (Go only) | **IMPROVED** |

---

## 5-Minute Test Plan

### Test 1: Simple Navigation (30 seconds)
```
1. Click voice orb
2. Say: "Open tasks"
3. Expected: OSA says "I've opened tasks" (or similar)
4. Check logs: tail -20 /tmp/backend-final-tools.log | grep "navigate_to_module"
```

✅ **Pass criteria**: Logs show tool call + execution

### Test 2: Create Task (1 minute)
```
1. Click voice orb
2. Say: "Create a task called buy milk"
3. Expected: OSA confirms task creation
4. Verify in database:
   psql $DATABASE_URL -c "SELECT title FROM tasks WHERE title ILIKE '%milk%';"
```

✅ **Pass criteria**: Task exists in database

### Test 3: List Tasks (1 minute)
```
1. Click voice orb
2. Say: "What are my tasks?"
3. Expected: OSA lists your tasks
4. Check logs: tail -20 /tmp/backend-final-tools.log | grep "list_tasks"
```

✅ **Pass criteria**: Logs show tool call + execution

### Test 4: Context Awareness (30 seconds)
```
1. Click voice orb
2. Say: "What's my name?"
3. Expected: OSA says "Roberto" (or your actual name)
4. No tool calls needed (context in prompt)
```

✅ **Pass criteria**: Correct name spoken

### Test 5: Multi-Turn (2 minutes)
```
1. Click voice orb
2. Say: "Open projects and create a project called test"
3. Expected: OSA does BOTH actions
4. Check logs for BOTH tool calls
```

✅ **Pass criteria**: Both tools executed

---

## Quick Log Checks

### Backend Logs (Tool Execution)
```bash
# See recent activity
tail -50 /tmp/backend-final-tools.log

# Filter for tool calls
tail -100 /tmp/backend-final-tools.log | grep -E "(Tool|tool_call)"

# See errors only
tail -100 /tmp/backend-final-tools.log | grep -i error
```

### Python Agent Logs (Audio Processing)
```bash
# See recent activity
tail -50 /tmp/agent-test-final.log

# Filter for STT/TTS
tail -100 /tmp/agent-test-final.log | grep -E "(STT|TTS)"

# Check registration
tail -100 /tmp/agent-test-final.log | grep "registered"
```

### Database Verification
```bash
# Check recent tasks
psql $DATABASE_URL -c "SELECT id, title, created_at FROM tasks ORDER BY created_at DESC LIMIT 5;"

# Check voice sessions
psql $DATABASE_URL -c "SELECT session_id, user_id, state FROM voice_sessions ORDER BY created_at DESC LIMIT 5;"
```

---

## If Something Breaks

### Backend Not Responding
```bash
# Check if running
ps aux | grep "[s]erver-refactored"

# Restart
killall server-refactored
/tmp/server-refactored > /tmp/backend-final-tools.log 2>&1 &
```

### Python Agent Not Running
```bash
# Check if running
ps aux | grep "[a]gent_groq"

# Restart
pkill -f agent_groq.py
cd /Users/rhl/Desktop/BusinessOS2/voice-agent
python3 agent_groq.py dev > /tmp/agent-test-final.log 2>&1 &
```

### LiveKit Connection Issues
```bash
# Check environment variables
cd /Users/rhl/Desktop/BusinessOS2/voice-agent
cat .env | grep LIVEKIT

# Verify credentials work
python3 -c "import os; from dotenv import load_dotenv; load_dotenv(); print('URL:', os.getenv('LIVEKIT_URL')[:40]+'...'); print('Key set:', bool(os.getenv('LIVEKIT_API_KEY')))"
```

---

## Expected Behavior

### When User Says "Open Tasks"

**Backend logs should show**:
```
INFO 🤖 Calling Groq API with tools
INFO 🔧 LLM requested tool calls count=1
INFO ⚙️ Executing tool tool=navigate_to_module args={"module":"tasks"}
INFO ✅ Tool executed tool=navigate_to_module result=Opened tasks module
INFO 🔄 Calling LLM again with tool results
INFO ✅ Final response after tool execution
```

**Python logs should show**:
```
INFO STT: "open tasks"
INFO Backend response received
INFO TTS: Playing audio
```

### When User Says "Create a Task"

**Backend logs should show**:
```
INFO 🤖 Calling Groq API with tools
INFO 🔧 LLM requested tool calls count=1
INFO ⚙️ Executing tool tool=create_task args={"title":"buy milk"}
INFO ✅ Creating task title=buy milk user_id=...
INFO ✅ Tool executed tool=create_task result=Created task: buy milk (ID: ...)
INFO 🔄 Calling LLM again with tool results
INFO ✅ Final response after tool execution
```

**Database should have**:
```sql
SELECT * FROM tasks WHERE title = 'buy milk';
-- Should return 1 row
```

---

## Architecture Diagram (Current)

```
USER SPEAKS
    ↓
[Python Agent] VAD + STT (Groq Whisper)
    ↓
POST /api/chat {messages, session_id}
    ↓
[Go Backend]
    ├─ Lookup session → user
    ├─ Fetch context (name, tasks, projects)
    ├─ Build personality prompt
    └─ Call Groq API with tools
        ↓
    [Groq LLM]
        ├─ Understands intent
        ├─ Calls tools (navigate, create_task, list_tasks)
        └─ Returns tool calls
            ↓
        [Tool Executor]
            ├─ Execute navigate_to_module()
            ├─ Execute create_task() → INSERT into DB
            ├─ Execute list_tasks() → SELECT from DB
            └─ Return results
                ↓
            [Second LLM Call]
                └─ Generate natural response with tool results
                    ↓
                Response → Python Agent
                    ↓
                [TTS] ElevenLabs
                    ↓
                USER HEARS RESPONSE
```

---

## Files Changed Summary

### NEW Files Created:
1. `desktop/backend-go/internal/tools/definitions.go` - Tool definitions for Groq
2. `desktop/backend-go/internal/tools/executor.go` - Tool execution with DB
3. `desktop/backend-go/internal/handlers/voice_chat_tools.go` - Groq API with function calling
4. `voice-agent/CLEANUP_AND_TESTING.md` - Comprehensive guide
5. `voice-agent/QUICK_TEST.md` - This file

### MODIFIED Files:
1. `desktop/backend-go/internal/handlers/voice_chat.go` - Line 240 (use tools)
2. `desktop/backend-go/internal/integrations/livekit/client.go` - Added io import

### DEPRECATED Files (not deleted yet):
1. `voice-agent/personality.py` - Entire file (personality moved to Go)
2. `voice-agent/agent_groq.py` - Lines 28, 49-236 (navigation code)

---

## Next Steps

1. ✅ **TEST**: Run the 5 test cases above (5 minutes)
2. 🧹 **CLEANUP**: Remove deprecated code from `agent_groq.py` (see CLEANUP_AND_TESTING.md)
3. 🗑️ **DELETE**: Remove `personality.py` file
4. 📝 **DOCUMENT**: Update any other docs that reference old architecture
5. 🚀 **DEPLOY**: Once tested, deploy to production

---

## Key Points

### What You Should NOT See Anymore:
- ❌ OSA saying "(laughs)" instead of laughing
- ❌ OSA not knowing your name (context should work)
- ❌ Pattern matching errors ("tasks" not matching "task")
- ❌ Fake navigation responses

### What You SHOULD See Now:
- ✅ Accurate navigation (LLM understands intent)
- ✅ Real task creation (voice → database)
- ✅ Task listing from voice
- ✅ Context-aware responses (knows your name)
- ✅ Natural multi-turn conversations

---

**Last Updated**: 2026-01-20 02:35
**Backend**: Running on port 8001
**Python Agent**: Registered as "groq-agent"
**Status**: ✅ READY TO TEST
