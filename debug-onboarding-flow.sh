#!/bin/bash
# Debug script for onboarding flow issue
# Run this to check if everything is set up correctly

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  OSA Build Onboarding Flow - Debug Script                   ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# Check if DATABASE_URL is set
if [ -z "$DATABASE_URL" ]; then
    echo "❌ DATABASE_URL is not set!"
    echo ""
    echo "Please export your database URL:"
    echo "export DATABASE_URL='your-postgres-connection-string'"
    echo ""
    exit 1
fi

echo "✅ DATABASE_URL is set"
echo ""

# Check if onboarding_completed column exists
echo "Checking database schema..."
echo "─────────────────────────────────────────────────────────────"
COLUMN_CHECK=$(psql "$DATABASE_URL" -t -c "SELECT column_name FROM information_schema.columns WHERE table_name = 'user' AND column_name = 'onboarding_completed';" 2>&1)

if echo "$COLUMN_CHECK" | grep -q "onboarding_completed"; then
    echo "✅ onboarding_completed column EXISTS in user table"
else
    echo "❌ onboarding_completed column DOES NOT EXIST in user table"
    echo ""
    echo "You need to apply migration 052:"
    echo "cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go"
    echo "psql \$DATABASE_URL < internal/database/migrations/052_add_onboarding_completed.sql"
    echo ""
    exit 1
fi

echo ""

# Check how many users have onboarding_completed = false
echo "Checking user onboarding status..."
echo "─────────────────────────────────────────────────────────────"
USER_COUNT=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM \"user\" WHERE onboarding_completed = FALSE;" 2>&1)

if echo "$USER_COUNT" | grep -q "relation.*does not exist"; then
    echo "❌ User table doesn't exist - database issue"
    exit 1
fi

echo "Users with onboarding incomplete: $USER_COUNT"
echo ""

# Check if backend server is running
echo "Checking if backend server is running..."
echo "─────────────────────────────────────────────────────────────"
if curl -s http://localhost:8001/health > /dev/null 2>&1; then
    echo "✅ Backend server is running on port 8001"
else
    echo "❌ Backend server is NOT running on port 8001"
    echo ""
    echo "Start the backend server:"
    echo "cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go"
    echo "go run cmd/server/main.go"
    echo ""
fi

echo ""

# Check if frontend server is running
echo "Checking if frontend server is running..."
echo "─────────────────────────────────────────────────────────────"
if curl -s http://localhost:5173 > /dev/null 2>&1; then
    echo "✅ Frontend server is running on port 5173"
else
    echo "❌ Frontend server is NOT running on port 5173"
    echo ""
    echo "Start the frontend server:"
    echo "cd /Users/rhl/Desktop/BusinessOS2/frontend"
    echo "npm run dev"
    echo ""
fi

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║  Next Steps                                                  ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""
echo "1. Make sure both servers are running"
echo "2. Restart them if they were already running (to pick up code changes)"
echo "3. Open browser in INCOGNITO mode"
echo "4. Go to http://localhost:5173/login"
echo "5. Click 'Continue with Google'"
echo "6. Sign in with a NEW Google account"
echo "7. Open DevTools (F12) → Console tab"
echo "8. Should redirect to /onboarding (not /window)"
echo ""
echo "If it still goes to /window, check:"
echo "- Backend terminal logs for 'new_user' cookie being set"
echo "- Browser DevTools → Application → Cookies → look for 'new_user'"
echo "- Network tab → Headers → look for Set-Cookie: new_user=true"
echo ""
