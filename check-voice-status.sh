#!/bin/bash

# Voice System Status Checker

echo "🔍 Voice System Status Check"
echo "═══════════════════════════════════════"
echo ""

# Check running agents
AGENT_COUNT=$(ps aux | grep "agent.py dev" | grep -v grep | wc -l | tr -d ' ')
echo -n "Python Voice Agents: "
if [ "$AGENT_COUNT" -eq "0" ]; then
    echo "❌ None running (start with: ./start-voice-agent.sh)"
elif [ "$AGENT_COUNT" -eq "1" ]; then
    echo "✅ 1 running (perfect!)"
else
    echo "⚠️  $AGENT_COUNT running (DUPLICATES! Kill with: pkill -9 -f 'agent.py dev')"
fi

# Check Go backend
echo -n "Go Backend: "
if curl -s http://localhost:8001/api/osa/health > /dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Not running"
fi

# Check Frontend
echo -n "Frontend: "
if curl -s http://localhost:5173 > /dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Not running"
fi

echo ""
echo "📝 Quick Commands:"
echo "  Start agent:  ./start-voice-agent.sh"
echo "  Kill agents:  pkill -9 -f 'agent.py dev'"
echo "  Check agents: ps aux | grep 'agent.py' | grep -v grep"
echo ""
