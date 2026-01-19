# 🧹 Voice System Cleanup & Testing Guide

**Created**: 2026-01-20
**Status**: Ready for cleanup and testing

---

## 📋 Table of Contents

1. [What Changed](#what-changed)
2. [Deprecated Code to Remove](#deprecated-code-to-remove)
3. [How It Works Now](#how-it-works-now)
4. [Testing Checklist](#testing-checklist)
5. [Verification Steps](#verification-steps)

---

## 🎯 What Changed

### BEFORE (Split Intelligence - Messy):
```
Python Agent:
- ✗ Pattern matching for navigation ("open tasks" → regex → module)
- ✗ Personality defined in personality.py
- ✗ Navigation responses generated in Python
- ✗ No real tool system - just pattern matching

Go Backend:
- ✓ Context enrichment (user name, tasks)
- ✓ LLM calls to Groq
- ✗ No tools system
```

**Problem**: Intelligence split between Python and Go. No real function calling.

### AFTER (Centralized Intelligence - Clean):
```
Python Agent:
- ✓ VAD (voice activity detection)
- ✓ STT (Groq Whisper: audio → text)
- ✓ TTS (ElevenLabs: text → audio)
- ✓ Session ID extraction
- ✗ NO LOGIC | NO TOOLS | NO PERSONALITY

Go Backend:
- ✓ Context enrichment (user name, tasks, projects)
- ✓ Personality prompt (ONE place!)
- ✓ Groq function calling with 5 tools
- ✓ Tool execution (DB writes, navigation)
- ✓ LLM orchestration with tool results
```

**Solution**: ALL intelligence in Go. Python is just voice I/O pipeline.

---

## 🗑️ Deprecated Code to Remove

### File: `voice-agent/agent_groq.py`

#### 1. Remove personality import (line 28)
```python
# ❌ DEPRECATED - Remove this:
from personality import build_system_prompt
```

**Why deprecated**: Personality is now in Go backend (`voice_chat.go` line 197-227).

#### 2. Remove MODULES dictionary (lines 49-101)
```python
# ❌ DEPRECATED - Remove this entire block:
MODULES = {
    "dashboard": "dashboard",
    "home": "dashboard",
    "chat": "chat",
    # ... entire dictionary
}
```

**Why deprecated**: Navigation is now handled by Go tool `navigate_to_module` in `tools/definitions.go`.

#### 3. Remove NAV_PATTERNS (lines 105-118)
```python
# ❌ DEPRECATED - Remove this:
NAV_PATTERNS = [
    r"(?:open|go to|switch to|show|navigate to|take me to)\s+(?:the\s+)?(.+)",
    # ... all patterns
]
```

**Why deprecated**: LLM with function calling detects navigation intent automatically.

#### 4. Remove all pattern lists (lines 121-168)
```python
# ❌ DEPRECATED - Remove ALL of these:
CLOSE_PATTERNS = [...]
BACK_PATTERNS = [...]
MINIMIZE_PATTERNS = [...]
DESKTOP_3D_PATTERNS = [...]
MAIN_WINDOW_PATTERNS = [...]
```

**Why deprecated**: No longer using pattern matching. Go tools handle everything.

#### 5. Remove detect_navigation_command() (lines 171-236)
```python
# ❌ DEPRECATED - Remove this entire function:
def detect_navigation_command(text: str) -> dict | None:
    """Detect if text is a navigation/control command."""
    # ... entire function
```

**Why deprecated**: LLM function calling replaces pattern matching.

#### 6. Remove get_nav_response() (line 239+)
```python
# ❌ DEPRECATED - Remove this function:
def get_nav_response(nav_command: dict) -> str:
    """Generate a spoken response for navigation command."""
    # ... entire function
```

**Why deprecated**: Go backend generates all responses with tool results.

#### 7. Remove any navigation detection in entrypoint (if present)
Check the `entrypoint()` function - if there's any code calling `detect_navigation_command()`, remove it.

### File: `voice-agent/personality.py`
```python
# ❌ DEPRECATED - Entire file can be deleted
```

**Why deprecated**: Personality moved to Go backend (`voice_chat.go` line 197-227).

---

## 🏗️ How It Works Now

### Complete Flow:

```
┌─────────────────────────────────────────────────────────────┐
│ 1. USER SPEAKS                                              │
│    "Open tasks and create a task called buy milk"          │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. PYTHON AGENT (voice-agent/agent_groq.py)                │
│    • VAD detects speech                                     │
│    • STT: Groq Whisper converts audio → text                │
│    • Sends text + session_id to Go backend                  │
└─────────────────────────────────────────────────────────────┘
                        ↓
                  POST /api/chat
                  {
                    "messages": [
                      {"role": "user", "content": "Open tasks..."}
                    ],
                    "session_id": "abc123"
                  }
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. GO BACKEND (desktop/backend-go/internal/handlers)       │
│                                                             │
│    voice_chat.go (line 188-240):                           │
│    • Lookup session → get user                             │
│    • Fetch user context (name, tasks, projects)            │
│    • Build personality prompt with context                  │
│    • Call callGroqAPIWithTools()                           │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. GROQ API WITH TOOLS (voice_chat_tools.go)              │
│                                                             │
│    callGroqAPIWithTools() does:                            │
│    1. Get tool definitions from tools/definitions.go       │
│    2. Send to Groq with tools array                        │
│    3. LLM responds with tool calls:                        │
│       [                                                     │
│         {                                                   │
│           "id": "call_123",                                 │
│           "type": "function",                               │
│           "function": {                                     │
│             "name": "navigate_to_module",                   │
│             "arguments": "{\"module\": \"tasks\"}"          │
│           }                                                 │
│         },                                                  │
│         {                                                   │
│           "id": "call_456",                                 │
│           "type": "function",                               │
│           "function": {                                     │
│             "name": "create_task",                          │
│             "arguments": "{\"title\": \"buy milk\"}"        │
│           }                                                 │
│         }                                                   │
│       ]                                                     │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ 5. TOOL EXECUTOR (tools/executor.go)                       │
│                                                             │
│    For EACH tool call:                                      │
│                                                             │
│    Tool 1: navigate_to_module                               │
│    • executeNavigation() runs                               │
│    • Returns: "Opened tasks module"                         │
│                                                             │
│    Tool 2: create_task                                      │
│    • executeCreateTask() runs                               │
│    • Inserts into PostgreSQL tasks table                    │
│    • Returns: "Created task: buy milk (ID: abc123)"         │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ 6. SECOND LLM CALL (voice_chat_tools.go)                  │
│                                                             │
│    Call Groq AGAIN with tool results:                      │
│    [                                                        │
│      {"role": "user", "content": "Open tasks..."},         │
│      {"role": "assistant", "tool_calls": [...]},           │
│      {"role": "tool", "content": "Opened tasks module"},   │
│      {"role": "tool", "content": "Created task: buy milk"} │
│    ]                                                        │
│                                                             │
│    LLM generates natural response:                         │
│    "I've opened the tasks module and created a task        │
│     called 'buy milk' for you!"                            │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ 7. PYTHON AGENT (agent_groq.py)                            │
│    • Receives text response from Go                         │
│    • TTS: ElevenLabs converts text → audio                  │
│    • Streams audio to user's speakers                       │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ 8. USER HEARS                                               │
│    "I've opened the tasks module and created a task        │
│     called 'buy milk' for you!"                            │
└─────────────────────────────────────────────────────────────┘
```

### Key Files and Their Roles:

| File | Responsibility | Status |
|------|----------------|--------|
| **voice-agent/agent_groq.py** | VAD, STT, TTS, session_id extraction | ✅ Keep (remove deprecated code) |
| **desktop/backend-go/internal/handlers/voice_chat.go** | Session lookup, context enrichment, personality prompt | ✅ Keep (line 240 modified) |
| **desktop/backend-go/internal/handlers/voice_chat_tools.go** | Groq API with function calling | ✅ NEW |
| **desktop/backend-go/internal/tools/definitions.go** | Tool definitions for LLM | ✅ NEW |
| **desktop/backend-go/internal/tools/executor.go** | Tool execution (DB, navigation) | ✅ NEW |
| **voice-agent/personality.py** | Old personality system | ❌ DELETE |

---

## ✅ Testing Checklist

### Pre-Test Verification:

```bash
# 1. Backend running?
curl http://localhost:8001/api/health
# Should respond (404 is OK - server is running)

# 2. Python agent registered?
tail -20 /tmp/agent-test-final.log | grep "registered worker"
# Should show: registered worker {"agent_name": "groq-agent"}

# 3. Check environment variables
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go
cat .env | grep -E "(GROQ_API_KEY|DATABASE_URL)"
# Should be set
```

### Test Cases:

#### Test 1: Simple Navigation
```
User says: "Open tasks"

Expected Flow:
1. Python STT → "Open tasks"
2. Go backend → LLM with tools
3. LLM calls: navigate_to_module(module="tasks")
4. Tool executes → "Opened tasks module"
5. LLM generates: "I've opened tasks for you"
6. Python TTS plays response

✅ Verify:
- User hears natural response
- Frontend tasks module opens (if frontend integration ready)
```

#### Test 2: Create Task
```
User says: "Create a task called buy milk"

Expected Flow:
1. Python STT → "Create a task called buy milk"
2. Go backend → LLM with tools
3. LLM calls: create_task(title="buy milk")
4. Tool executes → INSERT into tasks table → "Created task: buy milk (ID: xyz)"
5. LLM generates: "I've created a task called 'buy milk'"
6. Python TTS plays response

✅ Verify:
- User hears confirmation
- Task exists in database:
  psql $DATABASE_URL -c "SELECT title FROM tasks WHERE title = 'buy milk';"
```

#### Test 3: List Tasks
```
User says: "What are my tasks?"

Expected Flow:
1. Python STT → "What are my tasks?"
2. Go backend → LLM with tools
3. LLM calls: list_tasks()
4. Tool executes → SELECT from tasks → "You have 3 tasks: 1. Buy milk..."
5. LLM generates natural summary
6. Python TTS plays response

✅ Verify:
- User hears list of tasks
- Matches database:
  psql $DATABASE_URL -c "SELECT title FROM tasks LIMIT 10;"
```

#### Test 4: Multi-Turn Conversation
```
User says: "Open projects and list my projects"

Expected Flow:
1. Python STT → "Open projects and list my projects"
2. Go backend → LLM with tools
3. LLM calls TWO tools:
   - navigate_to_module(module="projects")
   - list_projects() [if implemented]
4. Tools execute → results returned
5. LLM generates combined response
6. Python TTS plays response

✅ Verify:
- User hears response addressing BOTH requests
- Both actions completed
```

#### Test 5: Context Awareness
```
User says: "What's my name?"

Expected Flow:
1. Python STT → "What's my name?"
2. Go backend enriches with context:
   - User name: "Roberto" (from database)
3. LLM has context in system prompt
4. LLM responds: "Your name is Roberto"
5. Python TTS plays response

✅ Verify:
- User hears correct name (Roberto)
- No tool calls needed (context in prompt)
```

---

## 🔍 Verification Steps

### 1. Check Backend Logs (Tool Calls)

```bash
tail -100 /tmp/backend-final-tools.log | grep "Tool"
```

**Expected output**:
```
INFO 🔧 LLM requested tool calls count=1
INFO ⚙️ Executing tool tool=navigate_to_module args={"module":"tasks"}
INFO ✅ Tool executed tool=navigate_to_module result=Opened tasks module
INFO 🔄 Calling LLM again with tool results tool_result_count=1
INFO ✅ Final response after tool execution
```

### 2. Check Python Agent Logs (Audio Processing)

```bash
tail -100 /tmp/agent-test-final.log
```

**Expected output**:
```
INFO STT: "open tasks"
INFO Backend response: "I've opened tasks for you"
INFO TTS: Playing audio
```

### 3. Verify Database Changes

```bash
# Check if tasks were created
psql $DATABASE_URL -c "SELECT id, title, created_at FROM tasks ORDER BY created_at DESC LIMIT 5;"
```

### 4. Test Without Frontend

You can test the ENTIRE system without frontend by:

```bash
# Send direct HTTP request to backend
curl -X POST http://localhost:8001/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [{"role": "user", "content": "create a task called test"}],
    "session_id": "test-session-123"
  }'
```

**Expected response**:
```json
{
  "response": "I've created a task called 'test' for you!",
  "tools_used": ["create_task"],
  "tool_results": ["Created task: test (ID: abc123)"]
}
```

---

## 🐛 Common Issues & Fixes

### Issue 1: Tools Not Being Called

**Symptom**: LLM responds normally but doesn't execute tools.

**Check**:
```bash
tail -100 /tmp/backend-final-tools.log | grep "tool"
```

**Possible Causes**:
- Tool definitions not being sent to Groq → Check `voice_chat_tools.go` line 65-73
- Groq model doesn't support tools → Verify using `mixtral-8x7b-32768`

**Fix**: Check logs for "🤖 Calling Groq API with tools" message.

### Issue 2: Context Not Working (OSA Doesn't Know User Name)

**Symptom**: OSA says "I don't know your name" when asked.

**Check**:
```bash
tail -100 /tmp/backend-final-tools.log | grep "Context"
```

**Possible Causes**:
- Session not linked to user → Check `voice_sessions` table
- User lookup failing → Check `voice_chat.go` line 201-208

**Fix**: Verify session exists:
```sql
SELECT * FROM voice_sessions WHERE session_id = 'your-session-id';
```

### Issue 3: Tasks Not Being Created

**Symptom**: Tool says "created task" but nothing in database.

**Check**:
```bash
# Check for SQL errors
tail -100 /tmp/backend-final-tools.log | grep -i "error"
```

**Possible Causes**:
- SQL query failing → Check `executor.go` line 126-135
- User ID not being passed → Check tool execution

**Fix**: Test direct database insert:
```sql
INSERT INTO tasks (user_id, title, status)
VALUES ('user-id-here', 'test', 'todo');
```

---

## 📊 Summary of Changes

| Component | Before | After | Change |
|-----------|--------|-------|--------|
| **Navigation Detection** | Python regex patterns | Go LLM function calling | ✅ More accurate |
| **Personality** | Python personality.py | Go system prompt | ✅ Centralized |
| **Tools System** | Pattern matching | Groq function calling | ✅ Real tools |
| **Context Enrichment** | Go backend | Go backend | ✓ Same (enhanced) |
| **Database Integration** | None | Tool executor with sqlc | ✅ Real actions |
| **Python Agent Role** | Smart (logic) | Dumb (I/O pipeline) | ✅ Simpler |

---

## 🎉 What You Get

### Benefits:
1. **Single Source of Truth**: All intelligence in Go backend
2. **Real Tool System**: LLM can call actual functions, not pattern matching
3. **Database Integration**: Tools can write to database (create tasks, projects, etc.)
4. **Extensible**: Add new tools by editing 2 files (`definitions.go`, `executor.go`)
5. **Maintainable**: Python is just voice I/O - easy to swap STT/TTS providers
6. **Context-Aware**: User name, tasks, projects all available to LLM
7. **Multi-Turn**: Tools can be chained (open module + create task in one request)

### What Still Works:
- ✅ Voice input/output
- ✅ Context awareness (user name, tasks, projects)
- ✅ Natural conversation
- ✅ LiveKit WebRTC connection
- ✅ Session management

### What's Better:
- ✅ Navigation is more accurate (LLM understands intent)
- ✅ Can create tasks from voice
- ✅ Can list tasks from voice
- ✅ Can ask about projects/context
- ✅ Extensible tool system for future features

---

## 🚀 Next Steps

1. **Test each test case** from the checklist above
2. **Verify logs** show tool calls and execution
3. **Check database** for created tasks
4. **Clean up deprecated code** (remove Python navigation patterns)
5. **Add new tools** as needed:
   - `update_task(id, status)` - Update task status
   - `delete_task(id)` - Delete task
   - `list_projects()` - Get projects
   - `search_clients(query)` - Search clients

---

**Last Updated**: 2026-01-20
**Status**: ✅ Ready to test and clean up deprecated code
