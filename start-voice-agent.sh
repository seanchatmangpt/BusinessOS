#!/bin/bash

# Voice Agent Startup Script - Prevents Duplicates
# This script ensures only ONE voice agent runs at a time

cd /Users/rhl/Desktop/BusinessOS2/python-voice-agent

# Kill any existing agents
echo "🧹 Checking for existing voice agents..."
if pgrep -f "agent.py dev" > /dev/null; then
    echo "⚠️  Found running agents - killing them..."
    pkill -9 -f "agent.py dev"
    sleep 2
fi

# Verify they're dead
if pgrep -f "agent.py dev" > /dev/null; then
    echo "❌ Failed to kill agents. Try manually: pkill -9 -f 'agent.py dev'"
    exit 1
fi

echo "✅ No duplicate agents running"
echo ""
echo "🚀 Starting voice agent..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Start the agent in foreground (so you can see console logs)
python3 agent.py dev

# If you prefer background mode, comment above line and uncomment below:
# nohup python3 agent.py dev > /tmp/voice-agent.log 2>&1 &
# echo "Agent started in background. View logs:"
# echo "  tail -f /tmp/voice-agent.log"
