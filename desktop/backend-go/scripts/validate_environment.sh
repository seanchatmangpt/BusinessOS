#!/bin/bash

echo "🔍 BusinessOS Voice System - Environment Validation"
echo "=================================================="
echo ""

ERRORS=0
WARNINGS=0

# Function to check env var
check_env() {
    local var_name=$1
    local required=$2

    if [ -z "${!var_name}" ]; then
        if [ "$required" = "true" ]; then
            echo "❌ MISSING (REQUIRED): $var_name"
            ((ERRORS++))
        else
            echo "⚠️  MISSING (OPTIONAL): $var_name"
            ((WARNINGS++))
        fi
    else
        echo "✅ SET: $var_name"
    fi
}

echo "📋 Environment Variables Check:"
echo "-------------------------------"

# Required
check_env "DATABASE_URL" "true"
check_env "SECRET_KEY" "true"
check_env "AI_PROVIDER" "true"

# Voice-specific (required)
check_env "OPENAI_API_KEY" "true"
check_env "ELEVENLABS_API_KEY" "true"
check_env "LIVEKIT_URL" "true"
check_env "LIVEKIT_API_KEY" "true"
check_env "LIVEKIT_API_SECRET" "true"

# Optional
check_env "OLLAMA_URL" "false"
check_env "REDIS_URL" "false"
check_env "GRPC_VOICE_PORT" "false"

echo ""
echo "🔧 Go Environment:"
echo "------------------"

# Check Go version
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo "✅ Go installed: $GO_VERSION"

    # Check if version is >= 1.21
    REQUIRED_VERSION="go1.21"
    if [[ "$GO_VERSION" < "$REQUIRED_VERSION" ]]; then
        echo "⚠️  Go version might be too old (need >= 1.21)"
        ((WARNINGS++))
    fi
else
    echo "❌ Go not installed"
    ((ERRORS++))
fi

echo ""
echo "📦 Go Dependencies:"
echo "-------------------"

# Check go.mod exists
if [ -f "go.mod" ]; then
    echo "✅ go.mod found"

    # Check critical dependencies
    DEPS=("github.com/livekit/server-sdk-go/v2" "github.com/jackc/pgx/v5" "gopkg.in/hraban/opus.v2")
    for dep in "${DEPS[@]}"; do
        if grep -q "$dep" go.mod; then
            echo "✅ Dependency: $dep"
        else
            echo "❌ Missing: $dep"
            ((ERRORS++))
        fi
    done
else
    echo "❌ go.mod not found"
    ((ERRORS++))
fi

echo ""
echo "🗄️  Database Connection:"
echo "------------------------"

if [ -n "$DATABASE_URL" ]; then
    # Try to connect using psql if available
    if command -v psql &> /dev/null; then
        if psql "$DATABASE_URL" -c "SELECT 1" &> /dev/null; then
            echo "✅ Database connection successful"
        else
            echo "⚠️  Database connection failed (might be permissions)"
            ((WARNINGS++))
        fi
    else
        echo "⚠️  psql not installed, skipping DB test"
        ((WARNINGS++))
    fi
else
    echo "⚠️  DATABASE_URL not set, skipping DB test"
fi

echo ""
echo "🎙️  LiveKit Server:"
echo "-------------------"

if [ -n "$LIVEKIT_URL" ]; then
    # Try to reach LiveKit server
    if command -v curl &> /dev/null; then
        # Convert ws:// to http:// for curl test
        HTTP_URL=$(echo "$LIVEKIT_URL" | sed 's/^ws/http/')

        if curl -s --max-time 5 "$HTTP_URL" &> /dev/null; then
            echo "✅ LiveKit server reachable"
        else
            echo "⚠️  LiveKit server not reachable (might be down or wrong URL)"
            ((WARNINGS++))
        fi
    else
        echo "⚠️  curl not installed, skipping LiveKit test"
        ((WARNINGS++))
    fi
else
    echo "❌ LIVEKIT_URL not set"
    ((ERRORS++))
fi

echo ""
echo "📊 Summary:"
echo "-----------"
echo "Errors: $ERRORS"
echo "Warnings: $WARNINGS"

if [ $ERRORS -eq 0 ]; then
    echo ""
    echo "✅ Environment is ready for voice system!"
    exit 0
else
    echo ""
    echo "❌ Please fix $ERRORS error(s) before running voice system"
    exit 1
fi
