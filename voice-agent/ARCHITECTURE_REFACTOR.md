# Voice System Architecture Refactor Plan

## 🤔 The Problem: Split Intelligence

**You asked:** "wait so we need to give the go thing all the context all the shit we did in python for it right?"

**Answer:** YES! Currently intelligence is split between Python and Go (confusing). Should ALL be in Go.

---

## ❌ CURRENT ARCHITECTURE (Messy):

```
┌─────────────────────────────────────────────────────────┐
│                    PYTHON AGENT                         │
│  (Has TOO MUCH logic - should be dumb!)                 │
├─────────────────────────────────────────────────────────┤
│  ✅ VAD (voice activity detection)                      │
│  ✅ STT (Groq Whisper)                                  │
│  ❌ Navigation command detection ← SHOULDN'T BE HERE    │
│  ❌ Personality prompt (personality.py) ← DUPLICATE     │
│  ✅ Extracts session_id                                 │
│  ✅ TTS (ElevenLabs)                                    │
└─────────────────────────────────────────────────────────┘
                            │
                            │ POST /api/chat
                            │ {messages, session_id}
                            ↓
┌─────────────────────────────────────────────────────────┐
│                    GO BACKEND                           │
│  (Has SOME logic - should have ALL!)                    │
├─────────────────────────────────────────────────────────┤
│  ✅ Session lookup (voice_sessions table)              │
│  ✅ Context enrichment (user name, tasks, projects)    │
│  ✅ Personality prompt (voice_chat.go) ← DUPLICATE      │
│  ❌ NO tool definitions for LLM ← MISSING               │
│  ✅ Calls Groq LLM API                                  │
│  ❌ NO function calling ← MISSING                       │
└─────────────────────────────────────────────────────────┘

PROBLEMS:
1. Personality defined in TWO places (Python AND Go)
2. Navigation logic in Python (should be LLM function calling in Go)
3. No real tool system (create_task, etc.)
4. Python agent is "smart" (should be "dumb")
```

---

## ✅ IDEAL ARCHITECTURE (Clean):

```
┌─────────────────────────────────────────────────────────┐
│                    PYTHON AGENT                         │
│  (DUMB - only voice I/O pipeline)                       │
├─────────────────────────────────────────────────────────┤
│  ✅ VAD (voice activity detection)                      │
│  ✅ STT (Groq Whisper: audio → text)                    │
│  ✅ Extracts session_id from room metadata              │
│  ✅ TTS (ElevenLabs: text → audio)                      │
│                                                         │
│  NO LOGIC | NO TOOLS | NO PERSONALITY                   │
└─────────────────────────────────────────────────────────┘
                            │
                            │ POST /api/chat
                            │ {
                            │   "messages": [...],
                            │   "session_id": "abc123"
                            │ }
                            ↓
┌─────────────────────────────────────────────────────────┐
│                    GO BACKEND                           │
│  (SMART - ALL intelligence here!)                       │
├─────────────────────────────────────────────────────────┤
│  ✅ Session lookup (voice_sessions → user)             │
│  ✅ Context enrichment (BuildVoiceContext):            │
│     • User name (Roberto)                              │
│     • Workspace name                                   │
│     • Recent tasks                                     │
│     • Active projects                                  │
│     • User facts from DB                               │
│  ✅ Personality prompt (ONE place only!)               │
│  ✅ Tool definitions (Groq function calling):          │
│     • navigate_to_module(module)                       │
│     • create_task(title, due_date)                     │
│     • list_tasks(status)                               │
│     • create_project(name)                             │
│     • search_context(query)                            │
│  ✅ LLM call to Groq with tools                        │
│  ✅ Tool execution (handle function calls)             │
│  ✅ Return final response                              │
└─────────────────────────────────────────────────────────┘
                            │
                            │ If tool called:
                            │ execute tool → get result
                            │ → call LLM again with result
                            │
                            │ {
                            │   "response": "Opened tasks for you!"
                            │ }
                            ↓
                    Python plays audio

BENEFITS:
1. Single source of truth (ALL logic in Go)
2. Python is simple voice pipeline (easy to maintain)
3. LLM can call real tools via function calling
4. Context is automatically injected (user knows OSA)
5. Easy to add new tools (just add to Go)
```

---

## 🔧 What to Move/Add:

### 1. ✅ Context (Already in Go!)

```go
// voice_context.go - ALREADY EXISTS
func BuildVoiceContext(ctx context.Context, userID string) (*VoiceContext, error) {
    // Gets user data, workspace, tasks, projects from DB
    // Formats into system prompt
}
```

**Status**: ✅ DONE - No changes needed

---

### 2. ✅ Personality (Already in Go!)

```go
// voice_chat.go - ALREADY EXISTS (line 143-207)
basePrompt := `You are OSA...
🚨 CRITICAL RULE - NO PARENTHETICALS 🚨
NEVER use (laughs), (chuckles), etc...
`
```

**Action**: Delete `personality.py` from Python (it's duplicate/unused)

---

### 3. ❌ Tools (Need to Add to Go)

**Current Python** (manual pattern matching):
```python
# agent_groq.py
def detect_navigation_command(transcript):
    if "open tasks" in lower:
        return {"action": "navigate", "module": "tasks"}
```

**Should be Go** (LLM function calling):

```go
// internal/handlers/voice_chat.go

// Add tools array to Groq request
tools := []map[string]interface{}{
    {
        "type": "function",
        "function": {
            "name": "navigate_to_module",
            "description": "Open a module in the BusinessOS interface when the user asks to navigate somewhere",
            "parameters": {
                "type": "object",
                "properties": {
                    "module": {
                        "type": "string",
                        "enum": [
                            "dashboard", "chat", "tasks", "projects",
                            "team", "clients", "terminal", "settings",
                        ],
                        "description": "The module to open",
                    },
                },
                "required": ["module"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "create_task",
            "description": "Create a new task when user asks to add a task or reminder",
            "parameters": {
                "type": "object",
                "properties": {
                    "title": {
                        "type": "string",
                        "description": "Task title",
                    },
                    "due_date": {
                        "type": "string",
                        "description": "Due date in YYYY-MM-DD format (optional)",
                    },
                },
                "required": ["title"],
            },
        },
    },
}

// Call Groq with tools
reqBody := map[string]interface{}{
    "model":    "mixtral-8x7b-32768",
    "messages": enrichedMessages,
    "tools":    tools,  // ← Enable function calling
    "tool_choice": "auto",
}
```

---

### 4. ❌ Tool Execution (Need to Add to Go)

```go
// internal/handlers/voice_chat.go

// After getting response from Groq:
resp := groqResponse.Choices[0]

// Check if LLM wants to call a tool
if resp.Message.ToolCalls != nil && len(resp.Message.ToolCalls) > 0 {
    for _, toolCall := range resp.Message.ToolCalls {
        toolName := toolCall.Function.Name
        toolArgs := toolCall.Function.Arguments

        // Execute tool
        var toolResult string
        switch toolName {
        case "navigate_to_module":
            var args struct {
                Module string `json:"module"`
            }
            json.Unmarshal([]byte(toolArgs), &args)

            // Send navigation command to frontend via database or WebSocket
            toolResult = executeNavigation(args.Module)

        case "create_task":
            var args struct {
                Title   string `json:"title"`
                DueDate string `json:"due_date"`
            }
            json.Unmarshal([]byte(toolArgs), &args)

            // Create task in database
            task, _ := h.queries.CreateTask(ctx, sqlc.CreateTaskParams{
                Title:   args.Title,
                UserID:  userID,
                DueDate: args.DueDate,
            })
            toolResult = fmt.Sprintf("Created task: %s", task.Title)
        }

        // Call LLM again with tool result
        messages = append(messages, map[string]interface{}{
            "role": "tool",
            "tool_call_id": toolCall.ID,
            "content": toolResult,
        })

        // Make second LLM call with tool result
        // LLM will now say "I created the task for you!"
    }
}
```

---

## 🔀 Migration Steps:

### Step 1: Test Current System ⏳
- Connect voice orb
- Capture dispatch logs
- Verify context awareness works

### Step 2: Remove Python Logic ❌
```bash
# Remove these from agent_groq.py:
- MODULES dict (lines 46-99)
- detect_navigation_command() (lines 103-236)
- Command detection in on_user_speech() (lines 520-532)

# Remove file:
- personality.py (duplicate, Go has it)
```

### Step 3: Add Tools to Go ✅
```go
// Create: internal/tools/definitions.go
// Create: internal/tools/navigation.go
// Create: internal/tools/tasks.go
// Update: internal/handlers/voice_chat.go (add tools array)
```

### Step 4: Add Tool Execution ✅
```go
// Update: internal/handlers/voice_chat.go
// Add: HandleToolCalls() function
// Add: executeNavigation() function
// Add: executeCreateTask() function
```

### Step 5: Test Function Calling 🧪
```
User: "Open tasks"
→ LLM: tool_call(navigate_to_module, module="tasks")
→ Go: Executes navigation
→ LLM: "I've opened tasks for you!"
→ User hears response
```

---

## 🎯 Final Architecture Benefits:

1. ✅ **Python is dumb** - just voice I/O (easy to maintain)
2. ✅ **Go has all logic** - single source of truth
3. ✅ **Context automatic** - OSA knows your name, tasks, projects
4. ✅ **Real tools** - LLM can create tasks, not just navigate
5. ✅ **Extensible** - add new tools by adding to Go array

---

## 🚀 Next Actions:

1. **Test current system** - verify dispatch works
2. **Implement tools in Go** - add function calling
3. **Remove Python logic** - make it dumb
4. **Test new flow** - verify LLM uses tools correctly

**Do you want me to implement the Go tools now, or test the current system first?**
