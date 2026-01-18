#!/bin/bash
# Voice Agent Startup Script - gRPC Thin Adapter
# Starts the new Hybrid Go-First voice system

cd /Users/rhl/Desktop/BusinessOS2/python-voice-agent

# Kill any existing adapters
echo "🧹 Checking for existing voice adapters..."
if pgrep -f "grpc_adapter.py dev" > /dev/null; then
    echo "⚠️  Found running adapter - killing it..."
    pkill -9 -f "grpc_adapter.py dev"
    sleep 2
fi

# Verify it's dead
if pgrep -f "grpc_adapter.py dev" > /dev/null; then
    echo "❌ Failed to kill adapter. Try manually: pkill -9 -f 'grpc_adapter.py dev'"
    exit 1
fi

echo "✅ No duplicate adapters running"
echo ""
echo "🎤 Starting gRPC Voice Adapter (Hybrid Go-First Architecture)..."
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Start the adapter in foreground (so you can see console logs)
python3 grpc_adapter.py dev

# If you prefer background mode, comment above line and uncomment below:
# nohup python3 grpc_adapter.py dev > /tmp/voice-adapter.log 2>&1 &
# echo "Adapter started in background. View logs:"
# echo "  tail -f /tmp/voice-adapter.log"
