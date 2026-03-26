---
title: Your First API Call
type: tutorial
signal: S=(linguistic, tutorial, direct, markdown, step-by-step)
relates_to: [configuration, api-endpoints, error-codes]
prerequisites: [Go 1.24, Docker running, curl or Postman]
time: 10 minutes
difficulty: Beginner
---

# Your First API Call to BusinessOS

> **Make your first successful API request to BusinessOS in under 10 minutes.**
>
> Learn how the backend exposes business operations through REST endpoints.

---

## What You'll Do

By the end of this tutorial, you will:

1. Start BusinessOS backend (`make dev`)
2. Make your first HTTP request to `/api/agents`
3. Understand the response structure (JSON format, status codes)
4. Know how to check for errors

**Outcome:** See a 200 OK response with a list of agents in JSON format.

---

## Prerequisites

| Tool | Check With | Install |
|------|-----------|---------|
| **Go 1.24+** | `go version` | [golang.org](https://golang.org/doc/install) |
| **Docker** | `docker --version` | [Docker Desktop](https://www.docker.com/products/docker-desktop) |
| **curl or Postman** | `curl --version` | `brew install curl` or [Postman](https://www.postman.com/) |

---

## Step 1: Start BusinessOS (2 min)

Open a terminal and navigate to the BusinessOS directory:

```bash
cd /Users/sac/chatmangpt/BusinessOS
```

Start all services with the make command:

```bash
make dev
```

This command:
- Starts PostgreSQL (port 5432)
- Starts Redis (port 6379)
- Starts the Go backend (port 8001)
- Starts the SvelteKit frontend (port 5173)

**You'll see:**
```
Starting services...
✅ PostgreSQL running
✅ Redis running
✅ Backend running on http://localhost:8001
✅ Frontend running on http://localhost:5173
```

Wait until all services show ready. The backend is ready when you see:

```
Starting server on port 8001
```

---

## Step 2: Check Backend Health (1 min)

Before making requests to `/api/agents`, verify the backend is up:

**Using curl:**
```bash
curl -s http://localhost:8001/health
```

**Expected response:**
```json
{
  "status": "ok",
  "timestamp": "2026-03-25T10:30:45Z"
}
```

If you see this response with HTTP 200 — your backend is ready!

If you see a connection error, the backend hasn't started yet. Wait 5 seconds and try again.

---

## Step 3: Make Your First Request (1 min)

Now make a request to the `/api/agents` endpoint:

**Using curl:**
```bash
curl -s http://localhost:8001/api/agents | jq
```

(The `| jq` pipes the response through `jq` for pretty-printing. If you don't have `jq`, just remove it.)

**Using Postman:**
1. Open Postman
2. Create a new GET request
3. URL: `http://localhost:8001/api/agents`
4. Click Send

---

## Step 4: Understand the Response (3 min)

You'll see a JSON response like this:

```json
{
  "agents": [
    {
      "id": "agent-001",
      "name": "Architect",
      "role": "system_design",
      "status": "active",
      "created_at": "2026-03-25T10:00:00Z"
    },
    {
      "id": "agent-002",
      "name": "Builder",
      "role": "implementation",
      "status": "active",
      "created_at": "2026-03-25T10:05:00Z"
    }
  ],
  "total": 2,
  "timestamp": "2026-03-25T10:30:45Z"
}
```

### Response Structure Explained

| Field | Meaning |
|-------|---------|
| `agents` | Array of agent objects |
| `agents[].id` | Unique identifier (used in future requests) |
| `agents[].name` | Human-readable name |
| `agents[].role` | What the agent does (system_design, implementation, etc.) |
| `agents[].status` | Current state: active, inactive, paused |
| `agents[].created_at` | ISO 8601 timestamp when agent was created |
| `total` | How many agents returned |
| `timestamp` | When the API responded |

### What This Means

You just successfully:
- ✅ Connected to the backend
- ✅ Authenticated (no auth required for `/api/agents`)
- ✅ Retrieved business data
- ✅ Parsed a JSON response

---

## Step 5: Try a Second Request (2 min)

Now try getting details about a specific agent:

**Using curl:**
```bash
curl -s http://localhost:8001/api/agents/agent-001 | jq
```

**Expected response:**
```json
{
  "id": "agent-001",
  "name": "Architect",
  "role": "system_design",
  "status": "active",
  "capabilities": [
    "system_design",
    "technical_decision",
    "architecture_review"
  ],
  "created_at": "2026-03-25T10:00:00Z",
  "updated_at": "2026-03-25T10:30:45Z"
}
```

Notice the new fields:
- `capabilities` — an array of things this agent can do
- `updated_at` — when the agent was last modified

---

## Step 6: Test Error Handling (2 min)

APIs return error codes when something goes wrong. Let's test one:

**Request a non-existent agent:**
```bash
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:8001/api/agents/nonexistent
```

**Expected response:**
```json
{
  "error": "agent not found",
  "code": "AGENT_NOT_FOUND",
  "message": "Agent with ID 'nonexistent' does not exist"
}

HTTP Status: 404
```

### Common HTTP Status Codes

| Code | Meaning | Example |
|------|---------|---------|
| **200** | Success | Agent found and returned |
| **404** | Not found | Agent ID doesn't exist |
| **400** | Bad request | Invalid request format |
| **401** | Unauthorized | Missing authentication |
| **500** | Server error | Backend crashed or DB unavailable |

---

## Next Steps

Now that you've made your first API calls, you're ready for:

1. **[Tutorial: First Database Record](tutorial-first-database-record.md)** — Learn how data flows from API → Database → Response
2. **[Tutorial: First Frontend Component](tutorial-first-frontend-component.md)** — Build a UI that calls your API
3. **[Reference: API Endpoints](../reference/api-endpoints.md)** — See all available endpoints
4. **[How-to: Debug API Errors](../how-to/debug-api-errors.md)** — Advanced error diagnosis

---

## Troubleshooting

### "Connection refused" error

**Problem:** `curl: (7) Failed to connect to localhost port 8001`

**Solution:**
1. Check if backend is running: `docker ps | grep businessos`
2. Check logs: `docker logs businessos-backend`
3. Restart: `make down && make dev`

### "502 Bad Gateway" or timeout

**Problem:** Request hangs or returns 502

**Solution:**
1. Backend might be starting up — wait 10 seconds
2. Check if database is connected: `docker ps | grep postgres`
3. Check backend logs: `make logs`

### JSON parse error with `jq`

**Problem:** `parse error: Invalid numeric literal`

**Solution:**
- Run without `jq`: `curl -s http://localhost:8001/api/agents`
- Or use Postman which formats automatically

---

## What You Learned

✅ How to start BusinessOS services
✅ How to make HTTP requests to the backend
✅ How to read JSON responses
✅ How to interpret HTTP status codes
✅ How to handle API errors

**Key Insight:** BusinessOS is a REST API backend. Every action (creating agents, updating projects, tracking tasks) flows through HTTP endpoints that return JSON data.

---

*Your First API Call — Part of the BusinessOS Diataxis Tutorial Series*

**Word count: 520 words**

## See Also

- [Configuration Reference](../reference/configuration.md) — Backend environment variables
- [API Endpoints Reference](../reference/api-endpoints.md) — Complete endpoint catalog
- [Error Codes Reference](../reference/error-codes.md) — All possible error responses
- [Diátaxis Home](README.md) — Back to tutorial index
