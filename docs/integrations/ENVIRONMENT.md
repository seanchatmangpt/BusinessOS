# Environment Variable Reference

This document lists every environment variable recognized by BusinessOS. Copy `backend/.env.example` as your starting point and fill in the values described here.

---

## Database

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string. Format: `postgres://user:password@host:5432/dbname?sslmode=disable` |
| `SUPABASE_DIRECT_HOST` | No | Direct PostgreSQL host for Supabase (bypasses the pooler). Format: `db.PROJECT_ID.supabase.co:5432` |

**Local development example:**
```env
DATABASE_URL=postgres://postgres:yourpassword@localhost:5432/business_os?sslmode=disable
```

**Supabase example:**
```env
DATABASE_URL=postgres://postgres.PROJECT_ID:PASSWORD@aws-0-us-east-1.pooler.supabase.com:6543/postgres?pgbouncer=true
SUPABASE_DIRECT_HOST=db.YOUR-PROJECT-ID.supabase.co:5432
```

---

## Redis

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `REDIS_URL` | Yes | — | Redis connection URL. Format: `redis://localhost:6379/0` |
| `REDIS_PASSWORD` | Yes | — | Redis authentication password. Must match the value used to start the Redis server |
| `REDIS_KEY_HMAC_SECRET` | Yes | — | HMAC secret used to derive Redis key names. Prevents key enumeration |
| `REDIS_TLS_ENABLED` | No | `false` | Set to `true` in production to enable TLS for Redis connections |

**Generate secrets:**
```bash
openssl rand -base64 32   # REDIS_PASSWORD
openssl rand -base64 32   # REDIS_KEY_HMAC_SECRET
```

---

## Authentication and Security

| Variable | Required | Description |
|----------|----------|-------------|
| `SECRET_KEY` | Yes | JWT signing key. Minimum 64 bytes. Rotate this to invalidate all existing sessions |
| `TOKEN_ENCRYPTION_KEY` | Yes | AES-256 key for encrypting OAuth tokens at rest. Minimum 32 bytes |
| `OSA_WEBHOOK_SECRET` | No | Shared secret for verifying incoming webhook requests from the OSA runtime |

**Generate secrets:**
```bash
openssl rand -base64 64   # SECRET_KEY
openssl rand -base64 32   # TOKEN_ENCRYPTION_KEY
openssl rand -base64 32   # OSA_WEBHOOK_SECRET
```

> Rotating `SECRET_KEY` invalidates all active user sessions. Rotating `TOKEN_ENCRYPTION_KEY` requires re-encrypting all stored OAuth tokens — plan accordingly.

---

## AI Providers

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `AI_PROVIDER` | No | `ollama_local` | Active AI provider. One of: `ollama_local`, `anthropic`, `openai`, `groq`, `xai` |
| `DEFAULT_MODEL` | No | `llama3.2:3b` | Default model name used when no model is specified in a request |

### Ollama (Local — no API key required)

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_LOCAL_URL` | `http://localhost:11434` | Base URL of the local Ollama instance |

### Anthropic (Claude)

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | API key from [console.anthropic.com](https://console.anthropic.com/settings/keys) |
| `ANTHROPIC_MODEL` | Model ID. Example: `claude-sonnet-4-20250514` |

### OpenAI

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | API key from [platform.openai.com](https://platform.openai.com/api-keys) |
| `OPENAI_MODEL` | Model ID. Example: `gpt-4o` |

### Groq

| Variable | Description |
|----------|-------------|
| `GROQ_API_KEY` | API key from [console.groq.com](https://console.groq.com/keys) |
| `GROQ_MODEL` | Model ID. Example: `llama-3.3-70b-versatile` |

### xAI (Grok)

| Variable | Description |
|----------|-------------|
| `XAI_API_KEY` | API key from [console.x.ai](https://console.x.ai/) |
| `XAI_MODEL` | Model ID. Example: `grok-beta` |
| `XAI_BASE_URL` | Base URL. Default: `https://api.x.ai/v1` |

### Ollama Cloud (Optional)

| Variable | Description |
|----------|-------------|
| `OLLAMA_CLOUD_API_KEY` | API key for Ollama Cloud |
| `OLLAMA_CLOUD_MODEL` | Model name. Example: `llama3.2` |

---

## Google Integration

| Variable | Required | Description |
|----------|----------|-------------|
| `GOOGLE_CLIENT_ID` | For Google features | OAuth 2.0 Client ID from Google Cloud Console |
| `GOOGLE_CLIENT_SECRET` | For Google features | OAuth 2.0 Client Secret |
| `GOOGLE_REDIRECT_URI` | For Google features | OAuth callback URL. Default: `http://localhost:8001/api/auth/google/callback` |

---

## Microsoft 365 Integration

| Variable | Required | Description |
|----------|----------|-------------|
| `MICROSOFT_CLIENT_ID` | For Microsoft features | Azure Application (client) ID |
| `MICROSOFT_CLIENT_SECRET` | For Microsoft features | Azure client secret value |
| `MICROSOFT_REDIRECT_URI` | For Microsoft features | OAuth callback URL. Default: `http://localhost:8001/api/auth/microsoft/callback` |

---

## Slack Integration

| Variable | Required | Description |
|----------|----------|-------------|
| `SLACK_CLIENT_ID` | For Slack features | Slack app Client ID |
| `SLACK_CLIENT_SECRET` | For Slack features | Slack app Client Secret |
| `SLACK_SIGNING_SECRET` | For Slack features | Slack app Signing Secret (used to verify webhook payloads) |
| `SLACK_REDIRECT_URI` | For Slack features | OAuth callback URL. Default: `http://localhost:8001/api/auth/slack/callback` |

---

## Notion Integration

| Variable | Required | Description |
|----------|----------|-------------|
| `NOTION_CLIENT_ID` | For Notion features | Notion integration OAuth client ID |
| `NOTION_CLIENT_SECRET` | For Notion features | Notion integration OAuth client secret |
| `NOTION_REDIRECT_URI` | For Notion features | OAuth callback URL. Default: `http://localhost:8001/api/auth/notion/callback` |

---

## Linear Integration

| Variable | Required | Description |
|----------|----------|-------------|
| `LINEAR_CLIENT_ID` | For Linear features | Linear OAuth application Client ID |
| `LINEAR_CLIENT_SECRET` | For Linear features | Linear OAuth application Client Secret |
| `LINEAR_REDIRECT_URI` | For Linear features | OAuth callback URL. Default: `http://localhost:8001/api/auth/linear/callback` |
| `LINEAR_WEBHOOK_SECRET` | No | Secret for verifying Linear webhook payloads |

---

## HubSpot Integration

| Variable | Required | Description |
|----------|----------|-------------|
| `HUBSPOT_CLIENT_ID` | For HubSpot features | HubSpot app Client ID |
| `HUBSPOT_CLIENT_SECRET` | For HubSpot features | HubSpot app Client Secret |
| `HUBSPOT_REDIRECT_URI` | For HubSpot features | OAuth callback URL. Default: `http://localhost:8001/api/auth/hubspot/callback` |

---

## ClickUp Integration

| Variable | Required | Description |
|----------|----------|-------------|
| `CLICKUP_CLIENT_ID` | For ClickUp features | ClickUp app Client ID |
| `CLICKUP_CLIENT_SECRET` | For ClickUp features | ClickUp app Client Secret |
| `CLICKUP_REDIRECT_URI` | For ClickUp features | OAuth callback URL. Default: `http://localhost:8001/api/auth/clickup/callback` |

---

## Airtable Integration

| Variable | Required | Description |
|----------|----------|-------------|
| `AIRTABLE_CLIENT_ID` | For Airtable features | Airtable OAuth Client ID |
| `AIRTABLE_CLIENT_SECRET` | For Airtable features | Airtable OAuth Client Secret |
| `AIRTABLE_REDIRECT_URI` | For Airtable features | OAuth callback URL. Default: `http://localhost:8001/api/auth/airtable/callback` |

---

## Fathom Analytics Integration

| Variable | Required | Description |
|----------|----------|-------------|
| `FATHOM_API_KEY` | For Fathom features | Fathom API key from app.usefathom.com/api |

---

## Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8001` | TCP port the backend HTTP server listens on |
| `ENVIRONMENT` | `development` | Runtime environment. Set to `production` in deployed instances |
| `ALLOWED_ORIGINS` | `http://localhost:5173,...` | Comma-separated list of allowed CORS origins |
| `APP_URL` | `http://localhost:5173` | Public URL of the frontend application. Used in emails and OAuth redirects |

---

## Email Service (Resend)

| Variable | Required | Description |
|----------|----------|-------------|
| `RESEND_API_KEY` | For email | API key from [resend.com](https://resend.com/api-keys) |
| `RESEND_FROM_EMAIL` | No | From address for transactional emails. Default: `noreply@osa.dev` |

---

## Web Push Notifications (Optional)

VAPID keys are required to send Web Push notifications to browsers.

```bash
# Generate VAPID keys
npx web-push generate-vapid-keys
```

| Variable | Description |
|----------|-------------|
| `VAPID_PUBLIC_KEY` | VAPID public key (shared with browsers) |
| `VAPID_PRIVATE_KEY` | VAPID private key (kept server-side) |

---

## NATS Messaging (Optional)

NATS provides real-time message streaming between services. It is disabled by default.

| Variable | Default | Description |
|----------|---------|-------------|
| `NATS_ENABLED` | `false` | Enable NATS integration |
| `NATS_URL` | `nats://localhost:4222` | NATS server URL |
| `NATS_TTL_HOURS` | `24` | Message TTL in hours |

---

## Feature Flags

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_LOCAL_MODELS` | `true` | Allow users to select local Ollama models in the AI settings |

---

## `.env.example` Template

```env
# Database
DATABASE_URL=
SUPABASE_DIRECT_HOST=

# Security (generate with openssl rand -base64 N)
SECRET_KEY=
TOKEN_ENCRYPTION_KEY=
REDIS_PASSWORD=
REDIS_KEY_HMAC_SECRET=
OSA_WEBHOOK_SECRET=

# Redis
REDIS_URL=redis://localhost:6379/0
REDIS_TLS_ENABLED=false

# AI Provider
AI_PROVIDER=ollama_local
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=llama3.2:3b
ANTHROPIC_API_KEY=
ANTHROPIC_MODEL=claude-sonnet-4-20250514
OPENAI_API_KEY=
OPENAI_MODEL=gpt-4o
GROQ_API_KEY=
GROQ_MODEL=llama-3.3-70b-versatile
XAI_API_KEY=
XAI_MODEL=grok-beta
XAI_BASE_URL=https://api.x.ai/v1

# Google OAuth
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URI=http://localhost:8001/api/auth/google/callback

# Microsoft 365
MICROSOFT_CLIENT_ID=
MICROSOFT_CLIENT_SECRET=
MICROSOFT_REDIRECT_URI=http://localhost:8001/api/auth/microsoft/callback

# Slack
SLACK_CLIENT_ID=
SLACK_CLIENT_SECRET=
SLACK_SIGNING_SECRET=
SLACK_REDIRECT_URI=http://localhost:8001/api/auth/slack/callback

# Notion
NOTION_CLIENT_ID=
NOTION_CLIENT_SECRET=
NOTION_REDIRECT_URI=http://localhost:8001/api/auth/notion/callback

# Linear
LINEAR_CLIENT_ID=
LINEAR_CLIENT_SECRET=
LINEAR_REDIRECT_URI=http://localhost:8001/api/auth/linear/callback
LINEAR_WEBHOOK_SECRET=

# HubSpot
HUBSPOT_CLIENT_ID=
HUBSPOT_CLIENT_SECRET=
HUBSPOT_REDIRECT_URI=http://localhost:8001/api/auth/hubspot/callback

# ClickUp
CLICKUP_CLIENT_ID=
CLICKUP_CLIENT_SECRET=
CLICKUP_REDIRECT_URI=http://localhost:8001/api/auth/clickup/callback

# Airtable
AIRTABLE_CLIENT_ID=
AIRTABLE_CLIENT_SECRET=
AIRTABLE_REDIRECT_URI=http://localhost:8001/api/auth/airtable/callback

# Fathom Analytics
FATHOM_API_KEY=

# Server
SERVER_PORT=8001
ENVIRONMENT=development
ALLOWED_ORIGINS=http://localhost:5173,http://localhost:5174,http://localhost:3000
APP_URL=http://localhost:5173

# Email (Resend)
RESEND_API_KEY=
RESEND_FROM_EMAIL=noreply@osa.dev

# Web Push (optional)
VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=

# NATS (optional)
NATS_ENABLED=false
NATS_URL=nats://localhost:4222
NATS_TTL_HOURS=24

# Feature Flags
ENABLE_LOCAL_MODELS=true
```
