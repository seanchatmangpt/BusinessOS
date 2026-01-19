# Voice System Modular Architecture Guide

## 🎯 Making the System Modular & Customizable

This guide explains how the voice system is structured for easy customization and extension.

---

## 🧩 System Modules

### 1. **Personality Module** (Easily Swappable)

**File**: `voice-agent/personality.py`

**Purpose**: Defines OSA's personality, tone, behavior rules

**How to Customize**:
```python
# personality.py

def build_system_prompt() -> str:
    """Build OSA's personality prompt - CUSTOMIZE THIS!"""

    return """
You are OSA.

## YOUR PERSONALITY
- Tone: [friendly/professional/playful/serious]
- Response length: [brief/detailed/conversational]
- Emotional style: [warm/neutral/enthusiastic]

## YOUR RULES
- [Add your custom rules here]
- [Command behavior]
- [How to handle errors]
    """
```

**Use Cases**:
- Create different personas for different users
- A/B test personality variants
- Switch between formal/casual modes
- Create domain-specific assistants (e.g., medical OSA, legal OSA)

---

### 2. **STT Module** (Speech-to-Text Engine)

**Current**: Groq Whisper
**File**: `voice-agent/agent_groq.py` line 461

**How to Swap**:
```python
# Option 1: Use Deepgram instead
from livekit.plugins import deepgram
session = AgentSession(
    vad=silero.VAD.load(),
    stt=deepgram.STT(api_key=DEEPGRAM_API_KEY),  # ← Swap here
    llm=GoBackendLLM(session_id=session_id),
    tts=elevenlabs.TTS(...),
)

# Option 2: Use OpenAI Whisper
from livekit.plugins import openai
session = AgentSession(
    vad=silero.VAD.load(),
    stt=openai.STT(api_key=OPENAI_API_KEY, model="whisper-1"),
    llm=GoBackendLLM(session_id=session_id),
    tts=elevenlabs.TTS(...),
)
```

**Available Options**:
- Groq Whisper (fast, free tier available)
- Deepgram Nova-2 (highest accuracy, paid)
- OpenAI Whisper (good balance, paid)
- Azure Speech (enterprise, paid)

---

### 3. **LLM Module** (Language Model)

**Current**: Groq Mixtral (via Go backend)
**Files**:
- `voice-agent/agent_groq.py` lines 261-302 (wrapper)
- `desktop/backend-go/internal/handlers/voice_chat.go` lines 208-240 (actual API call)

**How to Swap**:

**Option A: Change backend LLM** (easiest):
```go
// voice_chat.go line 214
// Change this:
Model:    "mixtral-8x7b-32768",

// To this (faster):
Model:    "llama-3.3-70b-versatile",

// Or this (more accurate):
Model:    "llama-3.1-70b-versatile",
```

**Option B: Use different provider**:
```go
// Change groqBaseURL to:
anthropicURL := "https://api.anthropic.com/v1/messages"
openaiURL := "https://api.openai.com/v1/chat/completions"
```

**Option C: Local LLM** (privacy-focused):
```go
// voice_chat.go
ollamaURL := "http://localhost:11434/api/chat"
Model: "llama3.2:latest"
```

---

### 4. **TTS Module** (Text-to-Speech Engine)

**Current**: ElevenLabs
**File**: `voice-agent/agent_groq.py` lines 463-472

**How to Swap**:
```python
# Option 1: Use Cartesia (faster, cheaper)
from livekit.plugins import cartesia
session = AgentSession(
    vad=silero.VAD.load(),
    stt=groq.STT(api_key=GROQ_API_KEY),
    llm=GoBackendLLM(session_id=session_id),
    tts=cartesia.TTS(
        api_key=CARTESIA_API_KEY,
        voice_id="sonic-english",
        model="sonic-english",
    ),
)

# Option 2: Use OpenAI TTS
from livekit.plugins import openai
session = AgentSession(
    vad=silero.VAD.load(),
    stt=groq.STT(api_key=GROQ_API_KEY),
    llm=GoBackendLLM(session_id=session_id),
    tts=openai.TTS(
        api_key=OPENAI_API_KEY,
        voice="nova",  # alloy, echo, fable, onyx, nova, shimmer
        model="tts-1",
    ),
)
```

**Voice Customization (ElevenLabs)**:
```python
# Change voice by updating ELEVENLABS_VOICE_ID in .env
# Find voices at: https://elevenlabs.io/voice-library

# Tweak voice characteristics:
tts=elevenlabs.TTS(
    api_key=ELEVENLABS_API_KEY,
    voice_id=ELEVENLABS_VOICE_ID,
    model_id="eleven_turbo_v2_5",
    stability=0.5,  # 0.0 = expressive, 1.0 = robotic
    similarity_boost=0.75,  # 0.0 = creative, 1.0 = accurate
    style=0.2,  # 0.0 = neutral, 1.0 = exaggerated
)
```

---

### 5. **Context Enrichment Module**

**File**: `desktop/backend-go/internal/services/voice_context.go`

**Purpose**: Injects user-specific data into system prompt

**How to Customize**:
```go
// Add custom context sections
func (vc *VoiceContext) FormatForPrompt() string {
    var sb strings.Builder

    sb.WriteString("\n\n## YOUR CURRENT CONTEXT\n\n")
    sb.WriteString(fmt.Sprintf("**User**: %s\n", vc.UserName))

    // ADD YOUR CUSTOM SECTIONS HERE:

    // Example: User's timezone
    sb.WriteString(fmt.Sprintf("**Timezone**: %s\n", vc.UserTimezone))

    // Example: User's company info
    sb.WriteString(fmt.Sprintf("**Company**: %s (%s industry)\n",
        vc.CompanyName, vc.Industry))

    // Example: Recent activity
    sb.WriteString("**Recent Activity**:\n")
    for _, activity := range vc.RecentActivities {
        sb.WriteString(fmt.Sprintf("- %s\n", activity))
    }

    return sb.String()
}
```

**Data Sources You Can Add**:
- Calendar events
- Recent emails
- Slack messages
- CRM contacts
- Task deadlines
- Financial data
- Custom fields from your DB

---

### 6. **Command Detection Module**

**File**: `voice-agent/agent_groq.py` lines 103-236

**Purpose**: Detects voice commands to open modules

**How to Customize**:
```python
# Add new modules
MODULES = {
    # Your custom modules:
    "calendar": "calendar",
    "email": "email",
    "analytics": "analytics",

    # Synonyms:
    "schedule": "calendar",
    "inbox": "email",
    "reports": "analytics",
}

# Add new command types
def detect_navigation_command(transcript: str) -> dict | None:
    lower = transcript.lower()

    # Add custom commands:
    if "create new" in lower or "add a" in lower:
        return {"action": "create", "type": extract_type(lower)}

    if "search for" in lower:
        return {"action": "search", "query": extract_query(lower)}

    # ... existing navigation logic
```

---

## 🔌 Extension Points

### Adding New Features

#### 1. **Add Voice Commands**

```python
# agent_groq.py

# Add to detect_navigation_command():
if "remind me" in lower:
    return {
        "action": "reminder",
        "text": extract_reminder_text(lower),
        "time": extract_time(lower)
    }

# Handle in on_user_speech():
if command["action"] == "reminder":
    asyncio.create_task(send_tool_command("reminder", command))
```

#### 2. **Add Context Data**

```go
// voice_context.go

type VoiceContext struct {
    UserName        string
    WorkspaceName   string
    // Add your fields:
    UserTimezone    string
    CompanySize     string
    UserRole        string
    RecentMeetings  []Meeting
}

func BuildVoiceContext(ctx context.Context, user *User) (*VoiceContext, error) {
    // ... existing code

    // Add your queries:
    meetings, _ := queries.GetRecentMeetings(ctx, user.ID)

    return &VoiceContext{
        // ... existing fields
        RecentMeetings: meetings,
    }, nil
}
```

#### 3. **Add Tool Integrations**

```python
# Create new file: voice-agent/tools/calendar.py

import aiohttp

async def create_calendar_event(title: str, time: str):
    """Tool for creating calendar events"""
    async with aiohttp.ClientSession() as session:
        await session.post(f"{BACKEND_URL}/api/calendar/events", json={
            "title": title,
            "scheduled_at": time,
        })
        return f"Created event: {title}"

# Register in agent_groq.py:
from tools.calendar import create_calendar_event

tools = [
    create_calendar_event,
    # ... other tools
]

session = AgentSession(
    vad=silero.VAD.load(),
    stt=groq.STT(api_key=GROQ_API_KEY),
    llm=GoBackendLLM(session_id=session_id),
    tts=elevenlabs.TTS(...),
    tools=tools,  # ← Enable tools
)
```

---

## 🎛️ Configuration Management

### Environment-Based Configuration

**File**: `voice-agent/config.py` (create this)

```python
import os
from dataclasses import dataclass

@dataclass
class VoiceConfig:
    # STT
    stt_provider: str = "groq"  # groq | deepgram | openai
    stt_language: str = "en"

    # LLM
    backend_url: str = os.getenv("BACKEND_URL", "http://localhost:8080")

    # TTS
    tts_provider: str = "elevenlabs"  # elevenlabs | cartesia | openai
    tts_voice_id: str = os.getenv("ELEVENLABS_VOICE_ID")
    tts_stability: float = 0.75
    tts_similarity: float = 0.85

    # Behavior
    response_style: str = "conversational"  # brief | conversational | detailed
    personality_mode: str = "friendly"  # friendly | professional | playful

    # Features
    enable_commands: bool = True
    enable_context: bool = True
    enable_tools: bool = False

# Load config
config = VoiceConfig()
```

**Usage**:
```python
# agent_groq.py
from config import config

# Dynamic STT selection
if config.stt_provider == "groq":
    stt = groq.STT(api_key=GROQ_API_KEY)
elif config.stt_provider == "deepgram":
    stt = deepgram.STT(api_key=DEEPGRAM_API_KEY)

# Dynamic TTS selection
if config.tts_provider == "elevenlabs":
    tts = elevenlabs.TTS(
        api_key=ELEVENLABS_API_KEY,
        voice_id=config.tts_voice_id,
        stability=config.tts_stability,
        similarity_boost=config.tts_similarity,
    )
```

---

## 📦 Plugin System (Advanced)

### Creating Plugin Architecture

**File**: `voice-agent/plugins/__init__.py`

```python
from abc import ABC, abstractmethod

class VoicePlugin(ABC):
    """Base class for voice plugins"""

    @abstractmethod
    async def on_user_speech(self, text: str) -> dict | None:
        """Called when user speaks"""
        pass

    @abstractmethod
    async def on_agent_response(self, text: str) -> str | None:
        """Called before agent responds (can modify response)"""
        pass

# Example plugin:
class SentimentPlugin(VoicePlugin):
    async def on_user_speech(self, text: str):
        sentiment = await analyze_sentiment(text)
        if sentiment == "angry":
            return {"mood": "angry", "action": "apologize"}
        return None

    async def on_agent_response(self, text: str):
        # Make responses more empathetic if user is upset
        if hasattr(self, 'user_mood') and self.user_mood == "angry":
            return f"I understand you're frustrated. {text}"
        return text

# Load plugins:
plugins = [
    SentimentPlugin(),
    CalendarPlugin(),
    NotificationPlugin(),
]

# Apply in agent:
for plugin in plugins:
    result = await plugin.on_user_speech(transcript)
    if result:
        # Handle plugin result
        pass
```

---

## 🔄 Quick Customization Checklist

### Want to change OSA's personality?
→ Edit `voice-agent/personality.py`

### Want a different voice?
→ Change `ELEVENLABS_VOICE_ID` in `.env`

### Want faster/cheaper STT?
→ Swap to `deepgram.STT()` in `agent_groq.py:461`

### Want a different LLM?
→ Change `Model` in `voice_chat.go:214`

### Want to add user data to context?
→ Edit `BuildVoiceContext()` in `voice_context.go`

### Want new voice commands?
→ Add to `MODULES` dict and `detect_navigation_command()` in `agent_groq.py`

### Want to add tools (calendar, email, etc.)?
→ Create tools following LiveKit Agents tool format, add to `AgentSession(..., tools=[])`

---

## 📚 Further Reading

- [LiveKit Agents SDK](https://docs.livekit.io/agents/)
- [ElevenLabs Docs](https://elevenlabs.io/docs)
- [Groq API Reference](https://console.groq.com/docs)
- [Voice System Architecture](./VOICE_SYSTEM_ARCHITECTURE.md)
- [Debugging Guide](./DEBUGGING_GUIDE.md)

---

**TL;DR**: The system is designed with **plug-and-play modules**. Swap STT, LLM, TTS providers by changing a few lines. Add features via plugins. Customize personality with one file.
