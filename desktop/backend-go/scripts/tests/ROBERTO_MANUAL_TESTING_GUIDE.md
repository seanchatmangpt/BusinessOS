# 🎯 Roberto's Complete Manual Testing Guide

**Date**: 2026-01-19
**Purpose**: Step-by-step guide to complete manual integration testing
**Estimated Time**: 4-5 hours total
**Status**: All automated tests ✅ COMPLETE - Ready for manual testing

---

## 📊 **Quick Status**

```
✅ AUTOMATED (Done by Claude)     6/6 tasks   100% COMPLETE
⏳ MANUAL (Roberto to do)         0/3 tasks     0% COMPLETE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 OVERALL PROGRESS                6/9 tasks    67% COMPLETE
```

---

## 🎯 **What Roberto Needs To Do**

Three main tasks, in this order:

1. **SETUP** (30 min) - Configure OAuth credentials
2. **OAUTH TESTING** (1-2 hours) - Connect 3-5 providers via browser
3. **FRONTEND TESTING** (2-3 hours) - Test UI with connected integrations

Optional:
4. **MCP TOOLS** (30 min) - Test tool execution with connected services

---

## 🚀 **PHASE 1: SETUP** (30 minutes)

### Step 1.1: Check Backend Server Running

```bash
cd ~/Desktop/BusinessOS2/desktop/backend-go
curl http://localhost:8001/health

# Expected: {"status":"healthy"}
# If not running: go run ./cmd/server
```

### Step 1.2: Configure OAuth Credentials

**Current Status**: ❌ No `.env` file found

**What you need**:
- Google OAuth credentials (Calendar, Gmail)
- Slack OAuth credentials (Workspace integrations)
- Linear OAuth credentials (Issue tracking)
- Notion OAuth credentials (Databases)
- HubSpot OAuth credentials (CRM)

**How to get credentials**:

#### 1️⃣ Google OAuth Setup

1. Go to https://console.cloud.google.com
2. Create/select project
3. Enable APIs:
   - Google Calendar API
   - Gmail API
   - Google Drive API
4. Create OAuth 2.0 credentials
5. Add redirect URI: `http://localhost:8001/api/integrations/google/callback`
6. Copy Client ID and Client Secret

#### 2️⃣ Slack OAuth Setup

1. Go to https://api.slack.com/apps
2. Create new app
3. Add OAuth scopes (channels:read, chat:write, etc.)
4. Add redirect URI: `http://localhost:8001/api/integrations/slack/callback`
5. Copy Client ID and Client Secret

#### 3️⃣ Linear OAuth Setup

1. Go to https://linear.app/settings/api
2. Create OAuth application
3. Add redirect URI: `http://localhost:8001/api/integrations/linear/callback`
4. Copy Client ID and Client Secret

#### 4️⃣ Notion OAuth Setup

1. Go to https://www.notion.so/my-integrations
2. Create new integration
3. Add redirect URI: `http://localhost:8001/api/integrations/notion/callback`
4. Copy OAuth Client ID and Secret

#### 5️⃣ HubSpot OAuth Setup

1. Go to https://developers.hubspot.com
2. Create app
3. Add redirect URI: `http://localhost:8001/api/integrations/hubspot/callback`
4. Copy Client ID and Client Secret

### Step 1.3: Create .env File

Create `~/Desktop/BusinessOS2/desktop/backend-go/.env`:

```bash
# Copy from .env.example if it exists
cp .env.example .env

# Or create new file with these OAuth credentials:
cat > .env << 'EOF'
# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Slack OAuth
SLACK_CLIENT_ID=your-slack-client-id
SLACK_CLIENT_SECRET=your-slack-client-secret

# Linear OAuth
LINEAR_CLIENT_ID=your-linear-client-id
LINEAR_CLIENT_SECRET=your-linear-client-secret

# Notion OAuth
NOTION_CLIENT_ID=your-notion-client-id
NOTION_CLIENT_SECRET=your-notion-client-secret

# HubSpot OAuth
HUBSPOT_CLIENT_ID=your-hubspot-client-id
HUBSPOT_CLIENT_SECRET=your-hubspot-client-secret

# Database (if not set)
DATABASE_URL=postgresql://localhost:5432/businessos

# Server
SERVER_PORT=8001
ENVIRONMENT=development
EOF

# Edit the file and replace placeholders with real credentials
nano .env
```

### Step 1.4: Restart Server

```bash
# Stop old server
pkill -f "cmd/server"

# Start with new .env
go run ./cmd/server

# Verify OAuth endpoints are configured
curl -s http://localhost:8001/api/integrations/providers | \
  jq -r '.providers[] | select(.id == "google" or .id == "slack") | "\(.id): \(.oauth_url)"'

# Expected output:
# google: /api/integrations/google/connect
# slack: /api/integrations/slack/connect
```

### Step 1.5: Start Frontend

```bash
cd ~/Desktop/BusinessOS2/frontend
npm install  # if first time
npm run dev

# Verify frontend is running
open http://localhost:5173
```

✅ **Setup complete!** You should now have:
- Backend running on port 8001
- Frontend running on port 5173
- OAuth credentials configured
- OAuth endpoints ready

---

## 🔐 **PHASE 2: OAUTH FLOW TESTING** (1-2 hours)

### Overview

Test OAuth integration flows by connecting 3-5 providers manually.

**Goal**: Verify OAuth authorization works end-to-end.

### Test 2.1: Connect Google Integration

**Time**: 15-20 minutes

1. **Initiate OAuth Flow**:
   ```bash
   # Option A: Via browser
   open http://localhost:8001/api/integrations/google/connect

   # Option B: Via frontend
   # Navigate to http://localhost:5173/integrations
   # Click "Connect" on Google card
   ```

2. **Complete Authorization**:
   - Google login page appears
   - Select your Google account
   - Review permissions (Calendar, Gmail, Drive)
   - Click "Allow"

3. **Verify Redirect**:
   - You're redirected back to BusinessOS
   - Should see success message
   - Check URL has `?success=true` or similar

4. **Verify in Database**:
   ```bash
   # Connect to PostgreSQL
   psql $DATABASE_URL

   # Check OAuth token saved
   SELECT provider, user_id, expires_at, created_at
   FROM oauth_tokens
   WHERE provider = 'google'
   ORDER BY created_at DESC
   LIMIT 1;

   # Expected: 1 row with your user_id and future expires_at
   ```

5. **Verify in Frontend**:
   - Go to http://localhost:5173/integrations
   - Click "Connected" tab
   - Google should appear in connected list

**✅ Success Criteria**:
- [  ] OAuth flow completed without errors
- [  ] Token saved in `oauth_tokens` table
- [  ] Google appears in "Connected" tab
- [  ] No console errors in browser

**📸 Screenshot**: Connected integrations showing Google

---

### Test 2.2: Connect Slack Integration

**Time**: 15-20 minutes

Same process as Google:

1. Navigate to `/api/integrations/slack/connect`
2. Authorize with Slack workspace
3. Verify redirect and success message
4. Check database: `SELECT * FROM oauth_tokens WHERE provider='slack';`
5. Verify in frontend "Connected" tab

**✅ Success Criteria**:
- [  ] Slack OAuth flow completed
- [  ] Token saved in database
- [  ] Slack appears in connected list

---

### Test 2.3: Connect Linear Integration

**Time**: 15-20 minutes

1. Navigate to `/api/integrations/linear/connect`
2. Authorize with Linear
3. Verify redirect
4. Check database: `SELECT * FROM oauth_tokens WHERE provider='linear';`
5. Verify in frontend

**✅ Success Criteria**:
- [  ] Linear OAuth flow completed
- [  ] Token saved in database
- [  ] Linear appears in connected list

---

### Test 2.4: Connect Notion Integration

**Time**: 15-20 minutes

1. Navigate to `/api/integrations/notion/connect`
2. Authorize with Notion
3. Select pages/databases to share
4. Verify redirect
5. Check database: `SELECT * FROM oauth_tokens WHERE provider='notion';`

**✅ Success Criteria**:
- [  ] Notion OAuth flow completed
- [  ] Token saved in database
- [  ] Notion appears in connected list

---

### Test 2.5: Connect HubSpot Integration (Optional)

**Time**: 15-20 minutes

1. Navigate to `/api/integrations/hubspot/connect`
2. Authorize with HubSpot
3. Verify redirect
4. Check database: `SELECT * FROM oauth_tokens WHERE provider='hubspot';`

---

### OAuth Testing Summary

After completing 3-5 provider connections:

```bash
# Verify all OAuth tokens
psql $DATABASE_URL -c "
SELECT
  provider,
  user_id,
  expires_at,
  created_at,
  CASE
    WHEN expires_at > NOW() THEN '✅ Valid'
    ELSE '❌ Expired'
  END as status
FROM oauth_tokens
WHERE deleted_at IS NULL
ORDER BY created_at DESC;
"
```

**Expected Output**:
```
provider  | user_id | expires_at | created_at | status
----------|---------|------------|------------|--------
google    | abc123  | 2026-01-20 | 2026-01-19 | ✅ Valid
slack     | abc123  | 2026-01-20 | 2026-01-19 | ✅ Valid
linear    | abc123  | 2026-01-20 | 2026-01-19 | ✅ Valid
notion    | abc123  | 2026-01-20 | 2026-01-19 | ✅ Valid
```

✅ **Phase 2 Complete!** - OAuth flows working correctly

---

## 🎨 **PHASE 3: FRONTEND VERIFICATION** (2-3 hours)

### Condensed Testing Checklist

Use this quick checklist instead of the full 650-line guide.

#### 3.1: Integration Hub Page (10 minutes)

1. Navigate to `/integrations`
2. Verify 4 tabs visible: Connected, Available, AI Models, Decisions
3. Click each tab - no errors
4. Provider cards display correctly
5. Search/filter works (if implemented)

**✅ Checklist**:
- [  ] Page loads without errors
- [  ] All 4 tabs functional
- [  ] Provider grid displays correctly
- [  ] Connected providers show in "Connected" tab
- [  ] No console errors

---

#### 3.2: OAuth Flow via UI (15 minutes)

1. Go to "Available" tab
2. Click "Connect" on an unconnected provider
3. Completes OAuth flow via UI (not direct URL)
4. Redirects back to integrations page
5. Provider moves to "Connected" tab

**✅ Checklist**:
- [  ] "Connect" button triggers OAuth
- [  ] Popup/redirect works correctly
- [  ] Success message appears
- [  ] Provider appears in Connected tab
- [  ] "Connect" button changes to "Connected" or "Disconnect"

---

#### 3.3: Integration Settings Page (20 minutes)

1. Click on a connected integration
2. Settings/details page loads
3. Displays:
   - Connection status
   - Last sync time
   - Sync statistics
   - "Disconnect" button
   - Manual sync trigger (if implemented)

**✅ Checklist**:
- [  ] Settings page loads
- [  ] Shows connection details
- [  ] "Disconnect" button visible
- [  ] Can trigger manual sync (if implemented)

---

#### 3.4: Disconnect Integration (10 minutes)

1. On integration settings page
2. Click "Disconnect"
3. Confirmation dialog appears
4. Confirm disconnect
5. Integration removed from "Connected" tab
6. Check database: Token soft-deleted

```bash
# Verify soft delete
psql $DATABASE_URL -c "
SELECT provider, deleted_at
FROM oauth_tokens
WHERE provider = 'google';
"
# Expected: deleted_at has a timestamp
```

**✅ Checklist**:
- [  ] Disconnect button works
- [  ] Confirmation dialog appears
- [  ] Integration removed from UI
- [  ] Token soft-deleted in database

---

#### 3.5: Error Handling (15 minutes)

Test error scenarios:

1. **Denied OAuth**: Start OAuth flow, click "Deny" on provider
   - Should show error message
   - Should not add to connected list

2. **Server Error**: Stop backend server, try to connect
   - Should show connection error
   - Should not crash frontend

3. **Invalid Token**: Manually set token expires_at to past
   - Should show "expired" status or re-auth prompt

**✅ Checklist**:
- [  ] OAuth denial handled gracefully
- [  ] Server errors shown to user
- [  ] Expired tokens detected
- [  ] No unhandled JavaScript errors

---

#### 3.6: Responsive Design (10 minutes)

1. Open DevTools (F12)
2. Toggle device toolbar (Ctrl+Shift+M)
3. Test different screen sizes:
   - Desktop (1920x1080)
   - Tablet (768x1024)
   - Mobile (375x667)

**✅ Checklist**:
- [  ] Layout adapts to screen size
- [  ] Provider cards stack on mobile
- [  ] Tabs usable on mobile
- [  ] No horizontal scrolling

---

### Frontend Testing Summary

After completing all frontend tests:

**Total Tests**: ~15 manual checks
**Expected Pass Rate**: 100%

**Document any failures**:
```
Test Failed: [describe what didn't work]
Expected: [what should happen]
Actual: [what actually happened]
Screenshot: [attach screenshot]
Console Errors: [paste any errors]
```

✅ **Phase 3 Complete!** - Frontend UI fully verified

---

## 🔧 **PHASE 4: MCP TOOLS TESTING** (Optional - 30 minutes)

Test MCP tool execution with connected integrations.

### Step 4.1: Get AUTH_TOKEN

```bash
# Option 1: From browser DevTools
# 1. Open http://localhost:5173
# 2. Open DevTools → Application → Local Storage
# 3. Find "authToken" or "session" key
# 4. Copy the token value

# Option 2: From login response
# Watch Network tab during login, copy token from response

# Set as environment variable
export AUTH_TOKEN="your-jwt-token-here"
```

### Step 4.2: Run MCP Tools Test Script

```bash
cd ~/Desktop/BusinessOS2/desktop/backend-go/scripts/tests

# Run MCP tools test
./e2e_mcp_tools_test.sh

# Expected: Tests for Google Calendar, Slack, Notion, Linear
```

### Step 4.3: Manual Tool Testing

Test individual tools:

```bash
# Test Google Calendar list events
curl -X POST http://localhost:8001/api/mcp/tools/calendar_list_events \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "timeMin": "2026-01-19T00:00:00Z",
    "timeMax": "2026-01-26T00:00:00Z"
  }'

# Expected: List of calendar events

# Test Slack list channels
curl -X POST http://localhost:8001/api/mcp/tools/slack_list_channels \
  -H "Authorization: Bearer $AUTH_TOKEN"

# Expected: List of Slack channels

# Test Notion search
curl -X POST http://localhost:8001/api/mcp/tools/notion_search \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "test"
  }'

# Expected: Notion search results
```

**✅ Checklist**:
- [  ] AUTH_TOKEN obtained successfully
- [  ] Calendar tools work
- [  ] Slack tools work
- [  ] Notion tools work
- [  ] Linear tools work
- [  ] All responses are valid JSON

---

## 📊 **FINAL VERIFICATION**

### Database Health Check

```bash
psql $DATABASE_URL << 'EOF'

-- OAuth Tokens Summary
SELECT
  provider,
  COUNT(*) as token_count,
  COUNT(*) FILTER (WHERE expires_at > NOW()) as valid_tokens,
  COUNT(*) FILTER (WHERE deleted_at IS NOT NULL) as deleted_tokens
FROM oauth_tokens
GROUP BY provider
ORDER BY provider;

-- Recent Sync Activity
SELECT
  provider,
  COUNT(*) as event_count,
  MAX(created_at) as last_sync
FROM synced_calendar_events
GROUP BY provider;

-- Webhook Subscriptions
SELECT
  provider,
  resource_type,
  status,
  event_count,
  last_event_at
FROM webhook_subscriptions
WHERE status = 'active'
ORDER BY provider;

EOF
```

### Test Results Summary

Fill out this summary:

```
========================================
BUSINESSOS INTEGRATION TESTING COMPLETE
========================================

Date: 2026-01-19
Tester: Roberto
Environment: Development (localhost)

AUTOMATED TESTS (Done by Claude):
✅ Unit Tests (25+ tests)
✅ Providers API (34 providers)
✅ Webhook Endpoints (9 endpoints)
✅ Bash Compatibility
✅ OAuth Test Script
✅ Documentation

MANUAL TESTS (Done by Roberto):
[  ] OAuth Flow - Google
[  ] OAuth Flow - Slack
[  ] OAuth Flow - Linear
[  ] OAuth Flow - Notion
[  ] OAuth Flow - HubSpot
[  ] Frontend - Integration Hub
[  ] Frontend - OAuth via UI
[  ] Frontend - Settings Page
[  ] Frontend - Disconnect
[  ] Frontend - Error Handling
[  ] Frontend - Responsive
[  ] MCP Tools - Calendar
[  ] MCP Tools - Slack
[  ] MCP Tools - Notion
[  ] MCP Tools - Linear

ISSUES FOUND:
[List any issues discovered]

OVERALL RESULT: [ PASS / FAIL / PARTIAL ]

READY FOR PRODUCTION: [ YES / NO ]
```

---

## 🎯 **SUCCESS CRITERIA**

Mark CUS-118 complete when:

- [  ] All automated tests passing ✅ (DONE)
- [  ] 3+ OAuth flows tested manually and working
- [  ] Frontend UI verified with connected integrations
- [  ] Database contains valid OAuth tokens
- [  ] No critical bugs found
- [  ] Test results documented

---

## 📞 **Need Help?**

If you encounter issues:

1. **Check server logs**: `tail -f ~/Desktop/BusinessOS2/desktop/backend-go/server.log`
2. **Check browser console**: F12 → Console tab
3. **Check database**: Connect with `psql $DATABASE_URL`
4. **Restart services**: Kill and restart backend/frontend

Common issues:
- **OAuth redirect fails**: Check callback URLs match in provider settings
- **Tokens not saving**: Check database connection and migrations
- **Frontend errors**: Check backend is running on port 8001
- **CORS errors**: Check CORS middleware configuration

---

**Good luck, Roberto! You got this! 🚀**
