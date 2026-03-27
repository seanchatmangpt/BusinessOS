# Reference: API Endpoints

**Information-oriented.** Complete list of all BusinessOS HTTP endpoints for lookup.

Format: `METHOD /path` → description, request body, response schema, status codes.

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
- **See also:** Troubleshooting health checks

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
- **See also:** Kubernetes liveness/readiness probes

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
- **See also:** How-to: Create a conversation

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
- **Response:** `200 OK`
  ```json
  {
    "id": "uuid",
    "title": "Project Planning",
    "context_id": "uuid",
    "created_at": "2026-03-25T10:00:00Z",
    "messages": [
      {
        "id": "uuid",
        "role": "user",
        "content": "What's the project status?",
        "created_at": "2026-03-25T10:00:00Z"
      }
    ]
  }
  ```
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
  ```json
  {
    "id": "uuid",
    "conversation_id": "uuid",
    "role": "user",
    "content": "What's the project status?",
    "created_at": "2026-03-25T10:00:00Z"
  }
  ```
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

### GET /api/v1/conversations/:id/sync
Sync conversation (real-time updates).
- **Auth:** Required
- **Query:** `since=2026-03-25T10:00:00Z` (last sync timestamp)
- **Response:** `200 OK` with updated messages
- **Status codes:** `200 OK`, `401 Unauthorized`, `404 Not Found`
- **See also:** Real-time synchronization guide

---

## Projects

### GET /api/v1/projects
List all projects for user.
- **Auth:** Required
- **Query:** `status=ACTIVE`, `limit=50`, `offset=0`
- **Response:** `200 OK`
  ```json
  {
    "projects": [
      {
        "id": "uuid",
        "name": "Client Portal",
        "description": "Build client-facing portal",
        "status": "ACTIVE",
        "priority": "HIGH",
        "start_date": "2026-03-01",
        "due_date": "2026-06-30",
        "created_at": "2026-03-25T10:00:00Z"
      }
    ],
    "total": 10
  }
  ```
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`
- **See also:** How-to: Create a project

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
- **See also:** Sync operations reference

---

## Tasks

### GET /api/v1/tasks
List all tasks for user.
- **Auth:** Required
- **Query:** `status=todo`, `project_id=uuid`, `limit=50`
- **Response:** `200 OK`
  ```json
  {
    "tasks": [
      {
        "id": "uuid",
        "title": "Design homepage",
        "description": "Create wireframes and mockups",
        "status": "in_progress",
        "priority": "high",
        "due_date": "2026-04-15",
        "project_id": "uuid",
        "assignee_id": "uuid",
        "created_at": "2026-03-25T10:00:00Z"
      }
    ],
    "total": 45
  }
  ```
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
- **See also:** Comments reference

### POST /api/v1/tasks/:id/comments
Add comment to task.
- **Auth:** Required
- **Request:**
  ```json
  {
    "content": "Need more details on requirements",
    "attachments": []
  }
  ```
- **Response:** `201 Created`
- **See also:** Comments reference

### GET /api/v1/tasks/sync
Sync all tasks (real-time).
- **Auth:** Required
- **Query:** `since=timestamp`
- **Response:** `200 OK`
- **See also:** Sync operations reference

---

## Workspace & Nodes

### GET /api/v1/nodes
List all workspace nodes.
- **Auth:** Required
- **Query:** `parent_id=uuid`, `type=PROJECT`
- **Response:** `200 OK`
  ```json
  {
    "nodes": [
      {
        "id": "uuid",
        "name": "Q1 Planning",
        "type": "BUSINESS",
        "health": "HEALTHY",
        "parent_id": "uuid",
        "created_at": "2026-03-25T10:00:00Z"
      }
    ]
  }
  ```
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
- **Response:** `200 OK`
  ```json
  {
    "id": "uuid",
    "name": "Q1 Planning",
    "type": "BUSINESS",
    "health": "HEALTHY",
    "linked_projects": [],
    "linked_contexts": [],
    "children": []
  }
  ```
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
- **See also:** How-to: Visualize project structure

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
  {
    "email": "user@example.com",
    "password": "secure_password"
  }
  ```
- **Response:** `200 OK`
  ```json
  {
    "access_token": "eyJhbGc...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "id": "user_id",
      "email": "user@example.com",
      "name": "John Doe"
    }
  }
  ```
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`
- **See also:** How-to: Authenticate with OAuth

### POST /api/v1/auth/logout
Logout user (invalidate token).
- **Auth:** Required
- **Response:** `204 No Content`
- **Status codes:** `204 No Content`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/users/me
Get current user profile.
- **Auth:** Required
- **Response:** `200 OK`
  ```json
  {
    "id": "user_id",
    "email": "user@example.com",
    "name": "John Doe",
    "image": "https://...",
    "created_at": "2026-03-25T10:00:00Z"
  }
  ```
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### PUT /api/v1/users/me
Update current user profile.
- **Auth:** Required
- **Request:**
  ```json
  {
    "name": "John Doe",
    "image": "https://..."
  }
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
- **Response:** `200 OK`
  ```json
  {
    "clients": [
      {
        "id": "uuid",
        "name": "Acme Corp",
        "type": "company",
        "status": "active",
        "email": "contact@acme.com",
        "phone": "+1-555-0100",
        "created_at": "2026-03-25T10:00:00Z"
      }
    ]
  }
  ```
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

## Integrations & A2A (Agent-to-Agent)

### POST /api/v1/integrations/a2a/agents
Create A2A (agent-to-agent) connection.
- **Auth:** Required (JWT)
- **Request:**
  ```json
  {
    "agent_id": "agent_123",
    "target_service": "osa",
    "action": "execute_task"
  }
  ```
- **Response:** `201 Created`
- **Status codes:** `201 Created`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`
- **See also:** A2A integration guide

### GET /api/v1/integrations/a2a/agents
List all A2A connections.
- **Auth:** Required
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/integrations/a2a/agents/:id/invoke
Invoke remote agent action.
- **Auth:** Required
- **Request:**
  ```json
  {
    "action": "execute_task",
    "parameters": {}
  }
  ```
- **Response:** `200 OK` with action result
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `404 Not Found`, `500 Internal Server Error`

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

## BOS Gateway (Process Mining)

### POST /api/v1/bos/discover
Forward discovery request to pm4py-rust.
- **Auth:** Required
- **Request:**
  ```json
  {
    "log_file": "base64_encoded_csv",
    "algorithm": "heuristics"
  }
  ```
- **Response:** `202 Accepted` (async processing)
- **Status codes:** `202 Accepted`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`
- **See also:** Process mining guide

### POST /api/v1/bos/conformance
Check log conformance against process model.
- **Auth:** Required
- **Request:**
  ```json
  {
    "log_file": "base64_encoded_csv",
    "model": "base64_encoded_model"
  }
  ```
- **Response:** `202 Accepted`
- **Status codes:** `202 Accepted`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/bos/progress
Receive progress updates from pm4py-rust (internal).
- **Auth:** Required (JWT)
- **Request:**
  ```json
  {
    "session_id": "uuid",
    "status": "processing",
    "progress": 45,
    "message": "Discovering patterns..."
  }
  ```
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### GET /api/v1/bos/stream/:session_id
Stream discovery results via Server-Sent Events (SSE).
- **Auth:** Required
- **Response:** `200 OK` with event stream
  ```
  data: {"progress": 45, "status": "processing"}
  ```
- **Status codes:** `200 OK`, `401 Unauthorized`, `404 Not Found`

### POST /api/v1/bos/tx/prepare
Prepare phase of two-phase commit.
- **Auth:** Required
- **Request:** Transaction details
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/bos/tx/commit
Commit phase of two-phase commit.
- **Auth:** Required
- **Request:** Transaction ID
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

### POST /api/v1/bos/tx/abort
Abort transaction (rollback).
- **Auth:** Required
- **Request:** Transaction ID
- **Response:** `200 OK`
- **Status codes:** `200 OK`, `400 Bad Request`, `401 Unauthorized`, `500 Internal Server Error`

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
- **See also:** SOC2 compliance configuration reference

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
- **Response:** `200 OK`
  ```json
  {
    "skills": [
      {
        "id": "skill_uuid",
        "name": "send_email",
        "description": "Send email via integrated mail service",
        "category": "communication"
      }
    ]
  }
  ```
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
      {
        "type": "project",
        "id": "uuid",
        "title": "Client Portal",
        "snippet": "Build client-facing portal..."
      }
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

All sync endpoints follow same pattern:

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

### GET /api/v1/osa/health
Check OSA service health (no auth).
- **Response:** `200 OK`
  ```json
  {
    "status": "healthy",
    "version": "1.0.0",
    "uptime_seconds": 86400
  }
  ```

### POST /api/v1/osa/config
Send OSA configuration (no auth).
- **Request:**
  ```json
  {
    "agent_pool_size": 10,
    "timeout_ms": 5000
  }
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
| `202 Accepted` | Request accepted for async processing (e.g., discovery job) |

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
| `503 Service Unavailable` | Service temporarily unavailable (e.g., database down) |

---

## Authentication

All endpoints except `/health`, `/ready`, `/healthz`, `/readyz`, and OAuth callbacks require:

```
Authorization: Bearer <JWT_TOKEN>
```

**Token Format:** JWT with claims:
```json
{
  "sub": "user_id",
  "email": "user@example.com",
  "exp": 1234567890
}
```

---

## See Also

- How-to: Authenticate with OAuth
- How-to: Create a conversation
- How-to: Integrate with A2A
- Error codes & troubleshooting
- Database schema reference
- Configuration options reference
