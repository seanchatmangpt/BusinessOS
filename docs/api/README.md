# API Reference

BusinessOS exposes a REST API used by the SvelteKit frontend and available to any HTTP client. This document covers authentication, request conventions, the main endpoint groups, and example requests for the most common operations.

---

## Base URL

```
http://localhost:8001/api/v1
```

In production, replace with your deployed domain.

---

## Authentication

### Cookie-Based (Browser Clients)

After a successful login, the backend sets an `httpOnly` session cookie. The browser sends this cookie automatically on every subsequent request — no manual token handling is needed for browser-based clients.

### CSRF Protection

All state-changing requests (`POST`, `PUT`, `PATCH`, `DELETE`) must include a valid CSRF token in the `X-CSRF-Token` header. The token is obtained from:

```
GET /api/v1/auth/csrf-token
```

```json
{
  "csrf_token": "abc123..."
}
```

Include this in the request header:

```http
X-CSRF-Token: abc123...
```

### Programmatic / API Clients

For non-browser clients, you can authenticate by sending the JWT token directly in the `Authorization` header:

```http
Authorization: Bearer <your_jwt_token>
```

Obtain a JWT token by calling the login endpoint:

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "you@example.com",
  "password": "yourpassword"
}
```

---

## Response Format

All endpoints return JSON. Successful responses contain the resource directly or a named wrapper. Errors follow a consistent format:

```json
{
  "error": "human-readable error message",
  "code": "error_code"
}
```

**Common HTTP status codes:**

| Status | Meaning |
|--------|---------|
| `200` | Success |
| `201` | Resource created |
| `400` | Invalid request body or parameters |
| `401` | Not authenticated |
| `403` | Authenticated but not authorized |
| `404` | Resource not found |
| `409` | Conflict (e.g., duplicate email) |
| `422` | Validation error |
| `500` | Internal server error |

---

## Endpoint Groups

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/auth/register` | Create a new account |
| `POST` | `/auth/login` | Log in with email and password |
| `POST` | `/auth/logout` | End the session |
| `GET` | `/auth/csrf-token` | Get a CSRF token |
| `GET` | `/auth/google` | Initiate Google OAuth flow |
| `GET` | `/auth/google/callback` | Handle Google OAuth callback |
| `POST` | `/auth/forgot-password` | Request a password reset email |
| `POST` | `/auth/reset-password` | Reset password with token |

### Conversations (Chat)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/conversations` | List conversations for the current user |
| `POST` | `/chat/send` | Send a message (returns SSE stream) |
| `GET` | `/conversations/:id` | Get a conversation with its messages |
| `DELETE` | `/conversations/:id` | Delete a conversation |

### Agents

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/agents` | List available agent types |
| `POST` | `/agents` | Create a custom agent |
| `GET` | `/agents/:id` | Get agent details |
| `PUT` | `/agents/:id` | Update a custom agent |
| `DELETE` | `/agents/:id` | Delete a custom agent |
| `GET` | `/agents/presets` | List built-in agent presets |

### Projects

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/projects` | List projects |
| `POST` | `/projects` | Create a project |
| `GET` | `/projects/:id` | Get a project |
| `PUT` | `/projects/:id` | Update a project |
| `DELETE` | `/projects/:id` | Delete a project |
| `GET` | `/projects/:id/members` | List project members |
| `POST` | `/projects/:id/members` | Add a member to a project |

### Tasks

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/tasks` | List tasks (filterable by project, status, assignee) |
| `POST` | `/tasks` | Create a task |
| `GET` | `/tasks/:id` | Get a task |
| `PUT` | `/tasks/:id` | Update a task |
| `DELETE` | `/tasks/:id` | Delete a task |

### Clients

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/clients` | List clients |
| `POST` | `/clients` | Create a client |
| `GET` | `/clients/:id` | Get a client |
| `PUT` | `/clients/:id` | Update a client |
| `DELETE` | `/clients/:id` | Delete a client |

### Knowledge / Contexts

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/contexts` | List knowledge contexts |
| `POST` | `/contexts` | Create a context |
| `POST` | `/contexts/:id/upload` | Upload a document to a context |
| `POST` | `/contexts/search` | Semantic search across contexts |

### Integrations

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/integrations` | List connected integrations |
| `GET` | `/integrations/:provider/connect` | Start OAuth authorization for a provider |
| `POST` | `/integrations/:provider/sync` | Trigger a manual sync |
| `DELETE` | `/integrations/:provider` | Disconnect an integration |

### Settings

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/settings` | Get workspace settings |
| `PUT` | `/settings` | Update workspace settings |
| `GET` | `/settings/ai` | Get AI configuration |
| `PUT` | `/settings/ai` | Update AI configuration |

### Analytics

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/analytics/usage` | Get token usage and request counts |
| `GET` | `/analytics/activity` | Get activity log |

### Intent Router

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/router/analyze` | Classify intent of a single message |
| `POST` | `/router/batch` | Classify intents of multiple messages |
| `GET` | `/router/intents` | List available intent types |

---

## Example Requests

### Register a New Account

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "Jane Smith"
}
```

**Response `201`:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "Jane Smith",
  "created_at": "2026-01-15T10:00:00Z"
}
```

---

### Send a Chat Message

```http
POST /api/v1/chat/send
Content-Type: application/json
X-CSRF-Token: your_csrf_token

{
  "message": "Summarize my open projects",
  "focus_mode": "analyze"
}
```

This endpoint returns a **Server-Sent Events stream** (`Content-Type: text/event-stream`). Each event contains a chunk of the AI response:

```
data: {"type":"content","delta":"Here are your "}
data: {"type":"content","delta":"3 open projects:"}
data: {"type":"done","conversation_id":"abc123"}
```

**Event types:**

| Type | Payload | Description |
|------|---------|-------------|
| `content` | `{"delta": "..."}` | A text chunk of the AI response |
| `thinking` | `{"delta": "..."}` | Chain of Thought reasoning (if enabled) |
| `tool_use` | `{"name": "...", "input": {...}}` | Agent tool call |
| `tool_result` | `{"content": "..."}` | Tool result |
| `done` | `{"conversation_id": "..."}` | Stream complete |
| `error` | `{"message": "..."}` | Error during streaming |

---

### Create a Project

```http
POST /api/v1/projects
Content-Type: application/json
X-CSRF-Token: your_csrf_token

{
  "name": "Website Redesign",
  "description": "Redesign the company website for Q2",
  "status": "active",
  "client_id": "550e8400-e29b-41d4-a716-446655440001"
}
```

**Response `201`:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "name": "Website Redesign",
  "description": "Redesign the company website for Q2",
  "status": "active",
  "client_id": "550e8400-e29b-41d4-a716-446655440001",
  "created_at": "2026-01-15T10:00:00Z"
}
```

---

### Create a Task

```http
POST /api/v1/tasks
Content-Type: application/json
X-CSRF-Token: your_csrf_token

{
  "title": "Design wireframes",
  "project_id": "550e8400-e29b-41d4-a716-446655440002",
  "priority": "high",
  "due_date": "2026-02-01T00:00:00Z"
}
```

---

### List Conversations with Filter

```http
GET /api/v1/conversations?context_id=abc123&limit=20&offset=0
```

---

### Semantic Search

```http
POST /api/v1/contexts/search
Content-Type: application/json
X-CSRF-Token: your_csrf_token

{
  "query": "customer onboarding process",
  "limit": 5
}
```

**Response `200`:**
```json
{
  "results": [
    {
      "id": "doc-123",
      "title": "Onboarding Handbook",
      "excerpt": "The customer onboarding process begins with...",
      "score": 0.92,
      "context_id": "ctx-456"
    }
  ]
}
```

---

### Classify Message Intent

```http
POST /api/v1/router/analyze
Content-Type: application/json
X-CSRF-Token: your_csrf_token

{
  "message": "Search the web for the latest AI news"
}
```

**Response `200`:**
```json
{
  "intent": "search",
  "confidence": 0.95,
  "requires_search": true,
  "suggested_agent": null,
  "suggested_focus_mode": "quick"
}
```

---

## Health Check

The backend exposes a health check endpoint that does not require authentication:

```http
GET /health
```

**Response `200`:**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

This endpoint is used by load balancers and container orchestrators to verify the service is running.

---

## Rate Limiting

The API applies rate limiting per IP address. Authentication endpoints have stricter limits:

| Endpoint Group | Limit |
|---------------|-------|
| General API | 100 requests per 15 minutes |
| Auth endpoints | 5 requests per minute |

Requests that exceed the limit receive a `429 Too Many Requests` response.

---

## See Also

- [Getting Started](../getting-started/README.md) — set up a local instance
- [Architecture Overview](../architecture/README.md) — understand the backend structure
- [Integrations](../integrations/README.md) — connect external services
