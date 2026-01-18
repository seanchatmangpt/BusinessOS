#!/bin/bash

# Voice Agent Minimal System - Quick Test Script
# Run this to verify the voice system is working

echo "🧹 Voice Agent Minimal System - Test Script"
echo "============================================"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Check Go Backend
echo -n "Testing Go Backend (port 8001)... "
if curl -s http://localhost:8001/api/osa/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Running${NC}"
else
    echo -e "${RED}❌ Not running${NC}"
    echo "Start with: cd desktop/backend-go && go run cmd/server/main.go"
fi

# Test 2: Check Frontend
echo -n "Testing Frontend (port 5173)... "
if curl -s http://localhost:5173 > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Running${NC}"
else
    echo -e "${RED}❌ Not running${NC}"
    echo "Start with: cd frontend && npm run dev"
fi

# Test 3: Check Python Voice Agent
echo -n "Testing Python Voice Agent... "
if ps aux | grep -q "[a]gent.py dev"; then
    echo -e "${GREEN}✅ Running${NC}"
else
    echo -e "${RED}❌ Not running${NC}"
    echo "Start with: cd python-voice-agent && python3 agent.py dev"
fi

# Test 4: LiveKit Token Generation
echo -n "Testing LiveKit Token Generation... "
RESPONSE=$(curl -s -X POST http://localhost:8001/api/livekit/token 2>&1)
if echo "$RESPONSE" | grep -q "token"; then
    echo -e "${GREEN}✅ Working${NC}"
    ROOM=$(echo "$RESPONSE" | grep -o '"room_name":"[^"]*"' | cut -d'"' -f4)
    echo "  Room: $ROOM"
else
    echo -e "${RED}❌ Failed${NC}"
fi

# Test 5: File Cleanup Verification
echo ""
echo "📊 Cleanup Verification:"
echo "----------------------"

cd /Users/rhl/Desktop/BusinessOS2/python-voice-agent
if [ ! -f tools.py ] && [ ! -f context.py ] && [ ! -d prompts ]; then
    echo -e "${GREEN}✅ Python bloat files deleted${NC}"
else
    echo -e "${RED}❌ Python bloat files still exist${NC}"
fi

AGENT_LINES=$(wc -l < agent.py 2>/dev/null || echo "0")
echo "  agent.py: $AGENT_LINES lines (was 249)"

cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go/internal/handlers
if [ ! -f voice_bus.go ] && [ ! -f voice_events.go ] && [ ! -f voice_ui.go ]; then
    echo -e "${GREEN}✅ Go bloat files deleted${NC}"
else
    echo -e "${RED}❌ Go bloat files still exist${NC}"
fi

VOICE_AGENT_LINES=$(wc -l < voice_agent.go 2>/dev/null || echo "0")
echo "  voice_agent.go: $VOICE_AGENT_LINES lines (was 119)"

cd /Users/rhl/Desktop/BusinessOS2/frontend/src/lib/components/desktop3d
if [ ! -f VoiceDebugPanel.svelte ] && [ ! -f PermissionPrompt.svelte ]; then
    echo -e "${GREEN}✅ Frontend bloat files deleted${NC}"
else
    echo -e "${RED}❌ Frontend bloat files still exist${NC}"
fi

# Summary
echo ""
echo "🎯 Next Steps:"
echo "-------------"
echo "1. Open http://localhost:5173 in your browser"
echo "2. Login (if needed)"
echo "3. Click '3D Desktop' button"
echo "4. Click the voice orb (silver circle)"
echo "5. Say 'Hello'"
echo ""
echo "Expected: You'll hear OSA respond in ~1-2 seconds"
echo ""
echo "📝 Check TEST_VOICE_MINIMAL.md for detailed testing guide"
