# Reference: API Endpoints

**Information-oriented.** Complete list of all BusinessOS HTTP endpoints for lookup.

**Last updated:** 2026-03-27 — Added BOS Gateway status route, corrected request schemas, added A2A agent protocol routes, added 2PC status route.

Format: `METHOD /path` → description, request body, response schema, status codes.

> **Path prefix note:** All `/api/*` routes are registered under both `/api/*` (deprecated, returns
> deprecation headers) and `/api/v1/*` (current). CSRF is skipped for `/api/bos/*` paths.

---

## Health & Status

### GET /health
Health check endpoint (no auth).
- **Response:** `200 OK`
  ```json
  {
    "status": "healthy",
    "timestamp": "2026-03-25T10:00:00Z"
  }
  ```

### GET /ready
Readiness probe (no auth).
- **Response:** `200 OK` (service ready) | `503 Service Unavailable` (not ready)
  ```json
  {
    "ready": true,
    "database": "connected",
    "redis": "connected"
  }
  ```

### GET /health/detailed
Detailed health report with component status (no auth).
- **Response:** `200 OK`
  ```json
  {
    "status": "healthy",
    "database": "connected",
    "redis": "connected",
    "osa_service": "healthy"
  }
  ```

### GET /healthz
Kubernetes liveness probe (no auth).
- **Response:** `200 OK` | `503 Service Unavailable`

### GET /readyz
Kubernetes readiness probe (no auth).
- **Response:** `200 OK` | `503 Service Unavailable`

---

## Chat & Conversations

### GET /api/v1/conversations
List all conversations for user.
- **Auth:** Required (JWT or session)
- **Query:** `limit=50`, `offset=0`
- **Response:** `200 OK`
  ```json
  {
    "conversations": [
      {
        "id": "uuid",
        "title": "Project Planning",
        "context_id": "uuid",
        "created_at": "2026-03-25T10:00:00Z",
        "updated_at": "2026-03-25T10:00:00Z"
      }
    ],
    "total": 150,
    "limit": 50,
    "offset": 0
  }
  ```
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/conversations
Create new conversation.
- **Auth:** Required
- **Request:**
  ```json
  {
    "title": "Project Planning",
    "context_id": "uuid"
  }
  ```
- **Response:** `201 Created`
  ```json
  {
    "id": "uuid",
    "title": "Project Planning",
    "context_id": "uuid",
    "created_at": "2026-03-25T10:00:00Z"
  }
  ```
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/conversations/:id
Retrieve single conversation.
- **Auth:** Required
- **Response:** `200 OK` with conversation + messages array
- **Status codes:** `200 OK`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### POST /api/v1/conversations/:id/messages
Add message to conversation.
- **Auth:** Required
- **Request:**
  ```json
  {
    "content": "What's the project status?",
    "role": "user"
  }
  ```
- **Response:** `201 Created`
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### GET /api/v1/conversations/:id/sync
Sync conversation (real-time updates).
- **Auth:** Required
- **Query:** `since=2026-03-25T10:00:00Z`
- **Response:** `200 OK` with updated messages
- **Status codes:** `200 OK`, `401 Unauthorized`, `404 Not Found`

---

## Projects

### GET /api/v1/projects
List all projects for user.
- **Auth:** Required
- **Query:** `status=ACTIVE`, `limit=50`, `offset=0`
- **Response:** `200 OK` with projects array and `total` count
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/projects
Create new project.
- **Auth:** Required
- **Request:**
  ```json
  {
    "name": "Client Portal",
    "description": "Build client-facing portal",
    "status": "ACTIVE",
    "priority": "HIGH",
    "start_date": "2026-03-01",
    "due_date": "2026-06-30",
    "client_id": "uuid"
  }
  ```
- **Response:** `201 Created`
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/projects/:id
Retrieve project details.
- **Auth:** Required
- **Response:** `200 OK` with full project object
- **Status codes:** `200 OK`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### PUT /api/v1/projects/:id
Update project.
- **Auth:** Required
- **Request:** Partial or complete project object
- **Response:** `200 OK` with updated project
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### DELETE /api/v1/projects/:id
Delete project.
- **Auth:** Required
- **Response:** `204 No Content`
- **Status codes:** `204 No Content`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### GET /api/v1/projects/sync
Sync all projects (real-time).
- **Auth:** Required
- **Query:** `since=timestamp`
- **Response:** `200 OK` with updated projects

---

## Tasks

### GET /api/v1/tasks
List all tasks for user.
- **Auth:** Required
- **Query:** `status=todo`, `project_id=uuid`, `limit=50`
- **Response:** `200 OK` with tasks array and `total` count
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/tasks
Create new task.
- **Auth:** Required
- **Request:**
  ```json
  {
    "title": "Design homepage",
    "description": "Create wireframes and mockups",
    "status": "todo",
    "priority": "high",
    "due_date": "2026-04-15",
    "project_id": "uuid",
    "assignee_id": "uuid"
  }
  ```
- **Response:** `201 Created`
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/tasks/:id
Retrieve task details.
- **Auth:** Required
- **Response:** `200 OK` with full task object including subtasks
- **Status codes:** `200 OK`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### PUT /api/v1/tasks/:id
Update task.
- **Auth:** Required
- **Request:** Partial task object
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### DELETE /api/v1/tasks/:id
Delete task.
- **Auth:** Required
- **Response:** `204 No Content`
- **Status codes:** `204 No Content`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### GET /api/v1/tasks/:id/comments
List task comments.
- **Auth:** Required
- **Response:** `200 OK` with array of comments

### POST /api/v1/tasks/:id/comments
Add comment to task.
- **Auth:** Required
- **Request:**
  ```json
  { "content": "Need more details on requirements", "attachments": [] }
  ```
- **Response:** `201 Created`

### GET /api/v1/tasks/sync
Sync all tasks (real-time).
- **Auth:** Required
- **Query:** `since=timestamp`
- **Response:** `200 OK`

---

## Workspace & Nodes

### GET /api/v1/nodes
List all workspace nodes.
- **Auth:** Required
- **Query:** `parent_id=uuid`, `type=PROJECT`
- **Response:** `200 OK` with nodes array
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/nodes
Create new workspace node.
- **Auth:** Required
- **Request:**
  ```json
  {
    "name": "Q1 Planning",
    "type": "BUSINESS",
    "parent_id": "uuid",
    "purpose": "Plan Q1 activities"
  }
  ```
- **Response:** `201 Created`
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/nodes/:id
Retrieve node details with connections.
- **Auth:** Required
- **Response:** `200 OK` with node + `linked_projects`, `linked_contexts`, `children`
- **Status codes:** `200 OK`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### PUT /api/v1/nodes/:id
Update node.
- **Auth:** Required
- **Request:** Partial node object
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### GET /api/v1/nodes/graph
Retrieve full node dependency graph.
- **Auth:** Required
- **Response:** `200 OK` with graph structure (nodes, edges, relationships)
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/nodes/sync
Sync all nodes (real-time).
- **Auth:** Required
- **Query:** `since=timestamp`
- **Response:** `200 OK`

---

## Users & Authentication

### POST /api/v1/auth/login
Authenticate user (email/password or OAuth).
- **Auth:** Not required
- **Request:**
  ```json
  { "email": "user@example.com", "password": "secure_password" }
  ```
- **Response:** `200 OK`
  ```json
  {
    "access_token": "eyJhbGc...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": { "id": "user_id", "email": "user@example.com", "name": "John Doe" }
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/auth/logout
Logout user (invalidate token).
- **Auth:** Required
- **Response:** `204 No Content`
- **Status codes:** `204 No Content`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/users/me
Get current user profile.
- **Auth:** Required
- **Response:** `200 OK` with user object
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### PUT /api/v1/users/me
Update current user profile.
- **Auth:** Required
- **Request:**
  ```json
  { "name": "John Doe", "image": "https://..." }
  ```
- **Response:** `200 OK` with updated user object
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/users/:id
Get user by ID (admin only).
- **Auth:** Required (admin)
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`, `500 Internal Server Error`

---

## CRM & Clients

### GET /api/v1/clients
List all clients.
- **Auth:** Required
- **Query:** `status=active`, `type=company`, `limit=50`
- **Response:** `200 OK` with clients array
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/clients
Create new client.
- **Auth:** Required
- **Request:**
  ```json
  {
    "name": "Acme Corp",
    "type": "company",
    "status": "lead",
    "email": "contact@acme.com",
    "phone": "+1-555-0100"
  }
  ```
- **Response:** `201 Created`
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/clients/:id
Retrieve client details.
- **Auth:** Required
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### PUT /api/v1/clients/:id
Update client.
- **Auth:** Required
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### DELETE /api/v1/clients/:id
Delete client.
- **Auth:** Required
- **Response:** `204 No Content`
- **Status codes:** `204 No Content`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

---

## Integrations (OAuth Callbacks)

### GET /api/v1/integrations/google/callback
Google OAuth callback endpoint (no auth).
- **Query:** `code=auth_code`, `state=state_token`
- **Response:** Redirect to frontend with session token
- **Status codes:** `302 Found`, `400 Bad Request`, `500 Internal Server Error`

### GET /api/v1/integrations/microsoft/callback
Microsoft OAuth callback endpoint (no auth).
- **Query:** `code=auth_code`, `state=state_token`
- **Response:** Redirect to frontend with session token
- **Status codes:** `302 Found`, `400 Bad Request`, `500 Internal Server Error`

---

## A2A Agent Protocol

Agent-to-agent endpoints use shared-secret authentication via the `X-Shared-Secret` header.
The calling agent must also supply `X-Agent-ID` to identify itself in the audit trail.

All responses include an `audit_entry` field with PROV-O hash-chain signature for governance.

### POST /api/integrations/a2a/crm/deals
Create a CRM deal via A2A call.
- **Auth:** `X-Shared-Secret` header required; `X-Agent-ID` header identifies calling agent
- **Request:**
  ```json
  {
    "name": "Enterprise License",
    "value": 50000.00,
    "extra": { "source": "agent-7", "probability": 0.85 }
  }
  ```
- **Response:** `201 Created`
  ```json
  {
    "deal": {
      "id": "uuid",
      "name": "Enterprise License",
      "value": 50000.00
    },
    "audit_entry": {
      "id": "uuid",
      "agent": "agent-7",
      "action": "create",
      "resource_type": "deal",
      "resource_id": "uuid",
      "sn_score": 0.9,
      "hash": "sha256:..."
    }
  }
  ```
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/integrations/a2a/crm/leads
Update a CRM lead status via A2A call.
- **Auth:** `X-Shared-Secret` header required
- **Request:**
  ```json
  {
    "lead_id": "uuid",
    "status": "qualified",
    "extra": { "score": 92 }
  }
  ```
- **Response:** `200 OK`
  ```json
  {
    "lead": { "id": "uuid", "status": "qualified" },
    "audit_entry": { ... }
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/integrations/a2a/projects/tasks
Assign a task via A2A call.
- **Auth:** `X-Shared-Secret` header required
- **Request:**
  ```json
  {
    "title": "Review proposal",
    "assignee": "user@example.com",
    "extra": { "priority": "high" }
  }
  ```
- **Response:** `201 Created`
  ```json
  {
    "task": { "id": "uuid", "title": "Review proposal", "assignee": "user@example.com" },
    "audit_entry": { ... }
  }
  ```
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/integrations/a2a/projects/progress
Update project progress via A2A call.
- **Auth:** `X-Shared-Secret` header required
- **Request:**
  ```json
  {
    "project_id": "uuid",
    "status": "on_track",
    "percent": 65,
    "extra": { "milestone": "Phase 2 complete" }
  }
  ```
- **Response:** `200 OK`
  ```json
  {
    "status": "updated",
    "audit_entry": { ... }
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/integrations/a2a/audit/query
Query A2A audit trail for a resource.
- **Auth:** `X-Shared-Secret` header required
- **Query:** `resource_type=deal` (required), `resource_id=uuid` (optional)
- **Response:** `200 OK`
  ```json
  {
    "entries": [
      {
        "id": "uuid",
        "agent": "agent-7",
        "action": "create",
        "resource_type": "deal",
        "resource_id": "uuid",
        "sn_score": 0.9,
        "hash": "sha256:..."
      }
    ],
    "count": 1
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

---

## BOS Gateway (Process Mining)

Routes are registered at `/api/bos/` (no versioning prefix). CSRF is skipped for this group.
The gateway proxies requests to pm4py-rust (default: `http://localhost:8090`, override via `PM4PY_RUST_URL`).

### GET /api/bos/status
Gateway health and statistics.
- **Auth:** None required
- **Response:** `200 OK`
  ```json
  {
    "status": "ok",
    "database_ready": true,
    "latency_ms": 2,
    "requests_total": 142,
    "requests_failed": 3,
    "average_latency_ms": 185.4,
    "uptime_seconds": 86400
  }
  ```
- **Status codes:** `200 OK`

### POST /api/bos/discover
Trigger process model discovery on an event log via pm4py-rust.
Results are persisted via write-ahead log before the response is returned.
- **Auth:** None required (CSRF skipped)
- **Request:**
  ```json
  {
    "log_path": "/data/events/order-process.json",
    "algorithm": "inductive_miner"
  }
  ```
  | Field | Type | Required | Default | Values |
  |-------|------|----------|---------|--------|
  | `log_path` | string | yes | — | Path to JSON or XES log file |
  | `algorithm` | string | no | `inductive_miner` | `alpha`, `heuristics`, `inductive_miner` |

- **Response:** `200 OK`
  ```json
  {
    "model_id": "uuid",
    "algorithm": "inductive_miner",
    "places": 12,
    "transitions": 8,
    "arcs": 24,
    "model_data": { ... },
    "latency_ms": 310
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request` (missing `log_path`), `503 Service Unavailable` (pm4py-rust down), `500 Internal Server Error`
- **See also:** Process mining guide; pm4py-rust `/api/discovery/alpha`

### POST /api/bos/conformance
Check event log conformance against a process model via pm4py-rust.
- **Auth:** None required (CSRF skipped)
- **Request:**
  ```json
  {
    "log_path": "/data/events/order-process.json",
    "model_id": "uuid",
    "model_path": "/data/models/order-model.pnml"
  }
  ```
  | Field | Type | Required |
  |-------|------|----------|
  | `log_path` | string | yes |
  | `model_id` | string | yes |
  | `model_path` | string | no |

- **Response:** `200 OK`
  ```json
  {
    "traces_checked": 1200,
    "fitting_traces": 1140,
    "fitness": 0.95,
    "precision": 0.88,
    "generalization": 0.91,
    "simplicity": 0.82,
    "latency_ms": 520
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `503 Service Unavailable`, `500 Internal Server Error`

### POST /api/bos/statistics
Extract statistics from an event log via pm4py-rust.
- **Auth:** None required (CSRF skipped)
- **Request:**
  ```json
  {
    "log_path": "/data/events/order-process.json"
  }
  ```
- **Response:** `200 OK`
  ```json
  {
    "log_name": "order-process",
    "num_traces": 1200,
    "num_events": 14400,
    "num_unique_activities": 12,
    "num_variants": 38,
    "avg_trace_length": 12.0,
    "min_trace_length": 3,
    "max_trace_length": 24,
    "activity_frequency": [
      { "activity": "Place Order", "frequency": 1200, "percentage": 8.33 }
    ],
    "case_duration": {
      "min_seconds": 300,
      "max_seconds": 172800,
      "avg_seconds": 14400.0,
      "median_seconds": 10800.0
    },
    "latency_ms": 210
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `503 Service Unavailable`, `500 Internal Server Error`

---

## BOS Progress & Streaming

### POST /api/bos/progress
Receive progress events from pm4py-rust during discovery/conformance operations.
Events are broadcast to SSE subscribers. This endpoint is called by pm4py-rust, not browsers.
- **Auth:** JWT Bearer token required (`Authorization: Bearer <token>`)
- **Request:**
  ```json
  {
    "progress": 50,
    "algorithm": "inductive_miner",
    "elapsed_ms": 2500,
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }
  ```
  | Field | Type | Required | Notes |
  |-------|------|----------|-------|
  | `progress` | uint32 | yes | 0–100 |
  | `algorithm` | string | yes | Name of running algorithm |
  | `elapsed_ms` | uint64 | yes | Wall-clock ms since start |
  | `session_id` | string | no | UUID; auto-generated if omitted |

- **Response:** `200 OK`
  ```json
  { "status": "received", "session_id": "550e8400-e29b-41d4-a716-446655440000" }
  ```
- **Status codes:** `200 OK`, `400 Bad Request` (invalid body or progress > 100), `401 Unauthorized`

### GET /api/bos/stream/discover/:session_id
Server-Sent Events stream for real-time discovery progress. Connect from browser to follow a running discovery job.
- **Auth:** JWT Bearer token required (user must be authenticated)
- **Path param:** `session_id` — UUID returned by or used in the discovery request
- **Response:** `200 OK` with `Content-Type: text/event-stream`
  ```
  data: {"id":"uuid","event_type":"discovery_progress","session_id":"...","progress":{"events_processed":5000,"percent_complete":42,"current_step":"inductive_miner"},"timestamp_ms":1711500000000}

  data: {"event_type":"discovery_complete","session_id":"...","timestamp_ms":1711500010000}
  ```
- **Status codes:** `200 OK` (stream open), `400 Bad Request` (invalid session UUID), `401 Unauthorized`, `500 Internal Server Error`

---

## BOS Transactions (Two-Phase Commit)

These endpoints implement the prepare/commit/abort protocol for distributed process mining transactions across pm4py-rust participants.

### POST /api/bos/tx/prepare
Initiate prepare phase. Validates input and locks the transaction participant.
- **Auth:** None required (CSRF skipped)
- **Request:**
  ```json
  {
    "transaction_id": "uuid",
    "algorithm": "inductive_miner",
    "log_data": {
      "log_path": "/data/events/order-process.json",
      "format": "json"
    },
    "parameters": {
      "noise_threshold": 0.2,
      "max_iterations": 100
    },
    "timeout_ms": 30000
  }
  ```
  | Field | Type | Required |
  |-------|------|----------|
  | `transaction_id` | string | yes |
  | `algorithm` | string | yes |
  | `log_data` | object | yes |
  | `parameters` | object | yes |
  | `timeout_ms` | int64 | no |

- **Response:** `200 OK`
  ```json
  {
    "transaction_id": "uuid",
    "status": "prepared",
    "vote": "commit",
    "version": 1,
    "model": { "model_id": "uuid", "algorithm": "inductive_miner" },
    "timestamp": "2026-03-27T10:00:00Z"
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `500 Internal Server Error`

### POST /api/bos/tx/commit
Commit a previously prepared transaction. Must be called after successful prepare.
- **Auth:** None required (CSRF skipped)
- **Request:**
  ```json
  { "transaction_id": "uuid" }
  ```
- **Response:** `200 OK`
  ```json
  {
    "transaction_id": "uuid",
    "status": "committed",
    "version": 2,
    "timestamp": "2026-03-27T10:00:01Z"
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `500 Internal Server Error`

### POST /api/bos/tx/abort
Abort and roll back a transaction (prepared or in-progress).
- **Auth:** None required (CSRF skipped)
- **Request:**
  ```json
  {
    "transaction_id": "uuid",
    "reason": "participant timeout"
  }
  ```
- **Response:** `200 OK`
  ```json
  {
    "transaction_id": "uuid",
    "status": "aborted",
    "version": 2,
    "timestamp": "2026-03-27T10:00:02Z"
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `500 Internal Server Error`

### GET /api/bos/tx/status/:xid
Retrieve current status of a transaction by its ID.
- **Auth:** None required
- **Path param:** `xid` — transaction UUID from prepare response
- **Response:** `200 OK`
  ```json
  {
    "transaction_id": "uuid",
    "status": "prepared",
    "started_at": "2026-03-27T10:00:00Z",
    "timestamp": "2026-03-27T10:00:00Z"
  }
  ```
- **Status codes:** `200 OK`, `404 Not Found`, `500 Internal Server Error`

---

## Compliance & Governance

### GET /api/v1/compliance/rules
List all compliance rules (SOC2, HIPAA, GDPR).
- **Auth:** Required (admin)
- **Query:** `framework=SOC2`, `severity=critical`
- **Response:** `200 OK`
  ```json
  {
    "rules": [
      {
        "id": "soc2.cc6.1",
        "title": "Logical Access Control",
        "framework": "SOC2",
        "severity": "critical",
        "enabled": true
      }
    ]
  }
  ```
- **Status codes:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `500 Internal Server Error`

### POST /api/v1/compliance/rules/reload
Hot-reload compliance rules (no restart).
- **Auth:** Required (admin)
- **Response:** `200 OK`
  ```json
  {
    "status": "reloaded",
    "rules_loaded": 25,
    "timestamp": "2026-03-25T10:00:00Z"
  }
  ```
- **Status codes:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `500 Internal Server Error`

### GET /api/v1/compliance/audit
Retrieve audit trail (SOC2 A1).
- **Auth:** Required (admin)
- **Query:** `since=timestamp`, `limit=1000`
- **Response:** `200 OK` with audit log entries
- **Status codes:** `200 OK`, `401 Unauthorized`, `403 Forbidden`, `500 Internal Server Error`

---

## Skills & Tools

### GET /api/v1/skills
List available skills.
- **Auth:** Required
- **Response:** `200 OK` with skills array
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/skills/:id/execute
Execute a skill.
- **Auth:** Required
- **Request:**
  ```json
  {
    "parameters": {
      "to": "user@example.com",
      "subject": "Meeting Update",
      "body": "The meeting has been rescheduled..."
    }
  }
  ```
- **Response:** `200 OK` with skill result
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

---

## Search & Memory

### GET /api/v1/search
Full-text search across all resources.
- **Auth:** Required
- **Query:** `q=project`, `limit=20`
- **Response:** `200 OK`
  ```json
  {
    "results": [
      { "type": "project", "id": "uuid", "title": "Client Portal", "snippet": "Build client-facing portal..." }
    ],
    "total": 42
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/memory
Retrieve knowledge base entries.
- **Auth:** Required
- **Query:** `category=technical`, `limit=50`
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/memory
Add new knowledge base entry.
- **Auth:** Required
- **Request:**
  ```json
  {
    "category": "technical",
    "key": "api_authentication",
    "content": "All API endpoints require JWT token in Authorization header..."
  }
  ```
- **Response:** `201 Created`
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

---

## Contexts & Documents

### GET /api/v1/contexts
List all document contexts.
- **Auth:** Required
- **Query:** `type=DOCUMENT`, `is_archived=false`
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/contexts
Create new context (document, profile, etc).
- **Auth:** Required
- **Request:**
  ```json
  {
    "name": "Client Profile",
    "type": "DOCUMENT",
    "content": "Company details...",
    "is_template": false
  }
  ```
- **Response:** `201 Created`
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/contexts/:id
Retrieve context details.
- **Auth:** Required (or public if `is_public=true` and `share_id` valid)
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### GET /api/v1/contexts/sync
Sync all contexts (real-time).
- **Auth:** Required
- **Query:** `since=timestamp`
- **Response:** `200 OK`

---

## Synchronization (Real-Time)

All sync endpoints follow the same pattern:

### GET /api/v1/{resource}/sync
Sync {resource} changes since last sync.
- **Auth:** Required
- **Query:** `since=ISO8601_timestamp` (optional, defaults to epoch)
- **Response:** `200 OK`
  ```json
  {
    "changes": [
      {
        "type": "created|updated|deleted",
        "resource_id": "uuid",
        "resource_type": "projects|tasks|...",
        "data": {},
        "timestamp": "2026-03-25T10:00:00Z"
      }
    ],
    "cursor": "next_sync_timestamp"
  }
  ```
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`
- **Resources:** contexts, conversations, projects, tasks, nodes, clients, calendar_events, daily_logs, team_members, artifacts, focus_items, user_settings

---

## OSA Health & Config (Public)

### GET /api/osa/health
Check OSA service health (no auth).
- **Response:** `200 OK`
  ```json
  { "status": "healthy", "version": "1.0.0", "uptime_seconds": 86400 }
  ```

### POST /api/osa/config
Send OSA configuration (no auth).
- **Request:**
  ```json
  { "agent_pool_size": 10, "timeout_ms": 5000 }
  ```
- **Response:** `200 OK`

---

## Responses & Status Codes

### Success Codes
| Code | Meaning |
|------|---------|
| `200 OK` | Request succeeded, response body contains data |
| `201 Created` | Resource created, response body contains new resource |
| `204 No Content` | Request succeeded, no response body (e.g., DELETE) |
| `202 Accepted` | Request accepted for async processing |

### Client Error Codes
| Code | Meaning |
|------|---------|
| `400 Bad Request` | Invalid request body or missing required fields |
| `401 Unauthorized` | Missing or invalid authentication token |
| `403 Forbidden` | Authenticated but insufficient permissions |
| `404 Not Found` | Resource not found |
| `409 Conflict` | Resource conflict (e.g., duplicate unique key) |

### Server Error Codes
| Code | Meaning |
|------|---------|
| `500 Internal Server Error` | Server error, check logs |
| `503 Service Unavailable` | Service temporarily unavailable (e.g., pm4py-rust down) |

---

## Authentication

All endpoints except `/health`, `/ready`, `/healthz`, `/readyz`, OSA health, and OAuth callbacks require:

```
Authorization: Bearer <JWT_TOKEN>
```

A2A endpoints use a separate shared-secret scheme:

```
X-Shared-Secret: <SECRET>
X-Agent-ID: <calling-agent-id>
```

**JWT Token Format:**
```json
{ "sub": "user_id", "email": "user@example.com", "exp": 1234567890 }
```

---

## See Also

- How-to: Authenticate with OAuth
- How-to: Create a conversation
- How-to: Integrate with A2A
- Error codes & troubleshooting
- Database schema reference
- Configuration options reference
- Process mining guide (BOS Gateway + pm4py-rust)
