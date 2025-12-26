# Slack Integration Documentation

## Overview

This document describes the Slack integration implementation for BusinessOS, enabling users to connect their Slack workspace and allowing the AI (OSA) to interact with Slack on their behalf.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    SLACK INTEGRATION STACK                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│   ┌──────────────────┐                                         │
│   │  Config Layer    │  config.go                              │
│   │  - SLACK_CLIENT_ID                                         │
│   │  - SLACK_CLIENT_SECRET                                     │
│   │  - SLACK_REDIRECT_URI                                      │
│   └────────┬─────────┘                                         │
│            │                                                    │
│   ┌────────▼─────────┐                                         │
│   │  Database Layer  │  schema.sql + slack_oauth.sql           │
│   │  - slack_oauth_tokens table                                │
│   │  - workspace_id, team_name                                 │
│   │  - bot_token, user_token                                   │
│   └────────┬─────────┘                                         │
│            │                                                    │
│   ┌────────▼─────────┐                                         │
│   │  Service Layer   │  services/slack.go                      │
│   │  - OAuth flow                                              │
│   │  - Token management                                        │
│   │  - Slack API calls                                         │
│   └────────┬─────────┘                                         │
│            │                                                    │
│   ┌────────▼─────────┐     ┌────────────────────┐             │
│   │  Handler Layer   │     │  MCP Tools Layer   │             │
│   │  slack_oauth.go  │     │  mcp_slack.go      │             │
│   │  - /auth         │     │  - slack_list_channels           │
│   │  - /callback     │     │  - slack_send_message            │
│   │  - /status       │     │  - slack_search_messages         │
│   │  - /disconnect   │     │  - slack_get_users               │
│   └──────────────────┘     └────────────────────┘             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Files Created/Modified

### Created Files

| File | Purpose |
|------|---------|
| `queries/slack_oauth.sql` | SQLC query definitions |
| `sqlc/slack_oauth.sql.go` | Generated Go code (auto) |
| `services/slack.go` | Slack API service |
| `services/mcp_slack.go` | MCP tools for AI |
| `handlers/slack_oauth.go` | OAuth HTTP endpoints |

### Modified Files

| File | Changes |
|------|---------|
| `config/config.go` | Added Slack credentials |
| `database/schema.sql` | Added slack_oauth_tokens table |
| `database/init.sql` | Added table for production |
| `services/mcp.go` | Integrated Slack service |
| `handlers/handlers.go` | Registered Slack routes |

---

## Database Schema

```sql
CREATE TABLE slack_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    workspace_id VARCHAR(255) NOT NULL,
    workspace_name VARCHAR(255),
    bot_token TEXT NOT NULL,
    user_token TEXT,
    bot_user_id VARCHAR(255),
    authed_user_id VARCHAR(255),
    bot_scopes TEXT[],
    user_scopes TEXT[],
    incoming_webhook_url TEXT,
    incoming_webhook_channel VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Token Types

- **bot_token**: Used by your app to send messages, list channels, etc.
- **user_token**: Used for search functionality (requires user authorization)

---

## API Endpoints

| Endpoint | Method | Auth | Purpose |
|----------|--------|------|---------|
| `/api/integrations/slack/auth` | GET | Required | Initiate OAuth flow |
| `/api/integrations/slack/callback` | GET | None | Handle OAuth redirect |
| `/api/integrations/slack/status` | GET | Required | Check connection status |
| `/api/integrations/slack` | DELETE | Required | Disconnect workspace |

### Example: Check Status Response

```json
{
  "connected": true,
  "workspace_id": "T1234567890",
  "workspace_name": "My Company",
  "connected_at": "2024-12-23T10:00:00Z"
}
```

---

## MCP Tools (AI-Accessible)

| Tool | Description |
|------|-------------|
| `slack_list_channels` | List accessible channels with member counts |
| `slack_send_message` | Send message to channel or DM, supports threads |
| `slack_get_channel_history` | Get recent messages from a channel |
| `slack_search_messages` | Search messages across workspace |
| `slack_list_users` | List workspace members |
| `slack_get_user_info` | Get details about a specific user |

### Example: slack_send_message

```json
{
  "tool": "slack_send_message",
  "arguments": {
    "channel": "#general",
    "text": "Hello from BusinessOS!",
    "thread_ts": "1234567890.123456"
  }
}
```

---

## OAuth Scopes

### Bot Scopes (App Permissions)

| Scope | Purpose |
|-------|---------|
| `channels:read` | View channel info |
| `channels:history` | Read public channel messages |
| `chat:write` | Send messages |
| `users:read` | View user profiles |
| `users:read.email` | See user emails |
| `groups:read` | View private channels |
| `groups:history` | Read private channel messages |
| `im:read` | View direct messages |
| `im:history` | Read DM messages |
| `mpim:read` | View group DMs |
| `mpim:history` | Read group DM messages |

### User Scopes

| Scope | Purpose |
|-------|---------|
| `search:read` | Search messages across workspace |

---

## Environment Variables

```bash
SLACK_CLIENT_ID=your-slack-client-id
SLACK_CLIENT_SECRET=your-slack-client-secret
SLACK_REDIRECT_URI=http://localhost:8000/api/integrations/slack/callback
```

---

## Setup Instructions

### 1. Create Slack App

1. Go to https://api.slack.com/apps
2. Click "Create New App" → "From scratch"
3. Enter app name and select workspace
4. Navigate to "OAuth & Permissions"
5. Add Redirect URL: `http://localhost:8000/api/integrations/slack/callback`
6. Add the bot scopes listed above
7. Copy Client ID and Client Secret

### 2. Configure Environment

Add to your `.env` file:

```bash
SLACK_CLIENT_ID=your-client-id
SLACK_CLIENT_SECRET=your-client-secret
SLACK_REDIRECT_URI=http://localhost:8000/api/integrations/slack/callback
```

### 3. Run Database Migration

Execute the table creation SQL on your PostgreSQL database.

### 4. Test the Integration

1. Start the backend server
2. Call `GET /api/integrations/slack/auth` to get the OAuth URL
3. Redirect user to the auth_url
4. User approves in Slack
5. Callback saves tokens automatically
6. MCP tools are now available for AI

---

## Library Used

- **slack-go/slack** v0.17.3 - Official Slack SDK for Go
  - Supports OAuth 2.0, Web API, Events API, Socket Mode
  - Type-safe with full API coverage

---

## Related Files

- Google Calendar Integration: Similar OAuth pattern in `google_calendar.go`
- MCP Service: `mcp.go` orchestrates all tool execution
- Handler Registration: `handlers.go` registers all routes

---

*Last Updated: December 2024*
