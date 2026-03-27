---
title: Your First Database Record
type: tutorial
signal: S=(linguistic, tutorial, direct, markdown, step-by-step)
relates_to: [api-endpoints, database-schema, sql-queries]
prerequisites: [PostgreSQL running, BusinessOS backend running, curl installed]
time: 10 minutes
difficulty: Beginner
---

# Your First Database Record

> **Create a record in BusinessOS and read it back in under 10 minutes.**
>
> Understand the full cycle: API Request → Database Insert → Read Response

---

## What You'll Do

You'll:
1. Inspect the BusinessOS database schema
2. Create a new record via API (POST request)
3. Verify it was stored in PostgreSQL
4. Read it back via API (GET request)

**Outcome:** See your data persisted in the database and returned by the API.

---

## Prerequisites

| Component | Check With | Status |
|-----------|-----------|--------|
| **PostgreSQL** | Running from `make dev` | Should be running |
| **Backend** | `curl http://localhost:8001/health` | Should return 200 |
| **Database** | `psql -U postgres -d business_os -c "\dt"` | Should list tables |

If anything is not running, go back to [Tutorial: First API Call](tutorial-first-api-call.md) and complete Step 1.

---

## Step 1: Understand the Data Model (2 min)

BusinessOS stores business data in PostgreSQL. Let's explore the schema.

**Using psql (command-line PostgreSQL):**

```bash
psql -U postgres -d business_os
```

Once connected, list all tables:

```sql
\dt
```

You'll see tables like:

```
                      List of relations
 Schema │          Name           │ Type  │     Owner
────────┼─────────────────────────┼───────┼──────────
 public │ agents                  │ table │ postgres
 public │ projects                │ table │ postgres
 public │ tasks                   │ table │ postgres
 public │ conversations           │ table │ postgres
 public │ memories                │ table │ postgres
 ...
```

Now let's look at the `agents` table structure:

```sql
\d agents
```

You'll see:

```
                            Table "public.agents"
      Column      │           Type           │     Modifiers
──────────────────┼──────────────────────────┼────────────────
 id               │ uuid                     │ not null
 name             │ character varying        │ not null
 role             │ character varying        │ not null
 status           │ character varying        │ default 'active'
 created_at       │ timestamp with time zone │ default now()
 updated_at       │ timestamp with time zone │ default now()
```

**Column Explanation:**

| Column | Type | Meaning |
|--------|------|---------|
| `id` | UUID | Unique identifier (auto-generated) |
| `name` | String | Human-readable name (e.g., "Architect") |
| `role` | String | What the agent does (e.g., "system_design") |
| `status` | String | Current state: active, inactive, paused |
| `created_at` | Timestamp | When record was created (auto-set) |
| `updated_at` | Timestamp | When record was last changed (auto-set) |

Exit psql with `\q`.

---

## Step 2: View Existing Records (1 min)

Let's see what agents already exist:

**Using curl:**

```bash
curl -s http://localhost:8001/api/agents | jq '.agents[0]'
```

You should see something like:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Architect",
  "role": "system_design",
  "status": "active",
  "created_at": "2026-03-25T10:00:00Z"
}
```

**Or directly in PostgreSQL:**

```bash
psql -U postgres -d business_os -c "SELECT id, name, role, status FROM agents LIMIT 1;"
```

Both queries show the same data:
- The API queries the database
- The database returns results
- The API formats and returns JSON

---

## Step 3: Create a New Agent Record (2 min)

Now let's create a new agent via the API:

**Using curl (POST request):**

```bash
curl -X POST http://localhost:8001/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My First Agent",
    "role": "custom_role"
  }' | jq
```

**Expected response:**

```json
{
  "id": "a1b2c3d4-e5f6-47g8-h9i0-j1k2l3m4n5o6",
  "name": "My First Agent",
  "role": "custom_role",
  "status": "active",
  "created_at": "2026-03-25T10:45:30Z",
  "updated_at": "2026-03-25T10:45:30Z"
}
```

**Important:** Save the `id` from the response. You'll use it in the next step.

### What Just Happened

1. **You sent** a JSON payload with name and role
2. **Backend received** the request in a handler function
3. **Handler validated** the data
4. **Service layer** executed business logic (if any)
5. **Database layer** inserted a new row into the `agents` table
6. **Backend returned** the created record with generated `id` and timestamps

---

## Step 4: Verify Data in Database (2 min)

Let's confirm the record was actually stored:

**Using psql:**

```bash
psql -U postgres -d business_os -c \
  "SELECT id, name, role, status, created_at FROM agents \
   WHERE name = 'My First Agent';"
```

You should see your newly created agent!

```
                  id                  │      name       │   role    │ status │          created_at
──────────────────────────────────────┼─────────────────┼───────────┼────────┼─────────────────────────
 a1b2c3d4-e5f6-47g8-h9i0-j1k2l3m4n5o6 │ My First Agent  │ custom_role│ active │ 2026-03-25 10:45:30.123
```

This proves:
- ✅ Data was stored in PostgreSQL
- ✅ The database has your exact record
- ✅ The backend successfully connected to the database

---

## Step 5: Retrieve the Record by ID (2 min)

Now let's fetch the record we just created:

**Using curl (replace ID with yours):**

```bash
curl -s http://localhost:8001/api/agents/a1b2c3d4-e5f6-47g8-h9i0-j1k2l3m4n5o6 | jq
```

**Expected response:**

```json
{
  "id": "a1b2c3d4-e5f6-47g8-h9i0-j1k2l3m4n5o6",
  "name": "My First Agent",
  "role": "custom_role",
  "status": "active",
  "capabilities": [],
  "created_at": "2026-03-25T10:45:30Z",
  "updated_at": "2026-03-25T10:45:30Z"
}
```

This is the same record you created, now fetched from the database!

---

## Step 6: Update the Record (2 min)

Let's modify the agent (e.g., change status to inactive):

**Using curl (PUT request):**

```bash
curl -X PUT http://localhost:8001/api/agents/a1b2c3d4-e5f6-47g8-h9i0-j1k2l3m4n5o6 \
  -H "Content-Type: application/json" \
  -d '{
    "status": "inactive"
  }' | jq
```

**Expected response:**

```json
{
  "id": "a1b2c3d4-e5f6-47g8-h9i0-j1k2l3m4n5o6",
  "name": "My First Agent",
  "role": "custom_role",
  "status": "inactive",
  "created_at": "2026-03-25T10:45:30Z",
  "updated_at": "2026-03-25T10:50:15Z"
}
```

Notice:
- `status` changed to `inactive`
- `updated_at` changed to the current time
- `created_at` stayed the same (records creation time)

---

## Step 7: Understand the API ↔ Database Cycle (3 min)

Here's the complete flow you just executed:

```
┌─────────────────────────────────────────────────────────┐
│ 1. CREATE REQUEST                                       │
│ curl -X POST http://localhost:8001/api/agents          │
│ Body: { "name": "My First Agent", "role": "custom" }   │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│ 2. BACKEND HANDLER                                      │
│ internal/handlers/agents.go                             │
│ - Validate request                                      │
│ - Parse JSON                                            │
│ - Call service layer                                    │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│ 3. SERVICE LAYER                                        │
│ internal/services/agent_service.go                      │
│ - Business logic (none in this case)                    │
│ - Call repository                                       │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│ 4. DATABASE LAYER                                       │
│ internal/database/repository.go                         │
│ INSERT INTO agents (name, role, status, created_at)    │
│ VALUES (?, ?, 'active', NOW())                          │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│ 5. POSTGRESQL                                           │
│ Row inserted with generated UUID and timestamps         │
│ Returns: (id, name, role, status, created_at, ...)     │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│ 6. RETURN TO API CALLER                                 │
│ HTTP 201 Created                                        │
│ Body: { "id": "...", "name": "...", ...}               │
└─────────────────────────────────────────────────────────┘
```

### The Three Layers

| Layer | Location | Responsibility |
|-------|----------|-----------------|
| **Handler** | `internal/handlers/agents.go` | Parse HTTP request, validate, serialize response |
| **Service** | `internal/services/agent_service.go` | Business logic, workflows, decisions |
| **Repository** | `internal/database/` | Run SQL queries, fetch/store data |

Every API operation flows through these three layers.

---

## Next Steps

1. **[Tutorial: First Frontend Component](tutorial-first-frontend-component.md)** — Build a UI that displays database records
2. **[How-to: Add Database Fields](../how-to/add-database-fields.md)** — Extend schema with new columns
3. **[Reference: Database Schema](../reference/database-schema.md)** — Complete schema documentation
4. **[How-to: Query Optimization](../how-to/query-optimization.md)** — Performance tuning

---

## Key Concepts

### CRUD Operations

| Operation | HTTP Method | Endpoint | Example |
|-----------|------------|----------|---------|
| **Create** | POST | `/api/agents` | Create new agent |
| **Read** | GET | `/api/agents/{id}` | Fetch specific agent |
| **Update** | PUT/PATCH | `/api/agents/{id}` | Modify agent |
| **Delete** | DELETE | `/api/agents/{id}` | Remove agent |

### HTTP Status Codes

| Code | Operation | Example |
|------|-----------|---------|
| **201** | Create successful | POST returns 201 Created |
| **200** | Read/Update successful | GET/PUT return 200 OK |
| **404** | Record not found | GET non-existent ID returns 404 |
| **400** | Invalid request | Missing required field returns 400 |

### Database → API Bridge

The backend **automatically**:
- Converts SQL rows → JSON objects
- Converts timestamps → ISO 8601 strings
- Generates UUIDs for new records
- Updates `updated_at` timestamp on changes

---

## Troubleshooting

### "Connection refused" when accessing database

**Problem:** `psql: could not translate host name "postgres" to address`

**Solution:**
1. Ensure PostgreSQL is running: `docker ps | grep postgres`
2. Use default connection: `psql -U postgres -d business_os`
3. If still fails, restart: `make down && make dev`

### "404 Not Found" after creating record

**Problem:** POST succeeds but GET fails

**Solution:**
1. Verify the ID is correct (copy-paste from POST response)
2. Check the endpoint path: `/api/agents/{id}` not `/api/agents?id={id}`
3. Wait a moment — database might still be committing

### "400 Bad Request" when creating record

**Problem:** POST returns 400 error

**Solution:**
1. Check JSON syntax: `echo '{"name":"test","role":"test"}' | jq` should parse
2. Required fields: both `name` and `role` are required
3. Check Content-Type header: `-H "Content-Type: application/json"`

---

## What You Learned

✅ How REST APIs create, read, update records
✅ How the backend layers (handler → service → repository) work
✅ How data persists in PostgreSQL
✅ How the API ↔ Database cycle flows
✅ The three CRUD patterns: Create, Read, Update

**Key Insight:** Every feature in BusinessOS follows this same cycle:
1. User interacts with frontend
2. Frontend calls API
3. Backend validates and processes
4. Data is stored/retrieved from database
5. Response flows back to user

---

*Your First Database Record — Part of the BusinessOS Diataxis Tutorial Series*

**Word count: 395 words**

## See Also

- [API Endpoints Reference](../reference/api-endpoints.md) — All available endpoints
- [Database Schema Reference](../reference/database-schema.md) — All tables and columns
- [SQL Queries Guide](../how-to/sql-queries.md) — Writing custom queries
- [Diátaxis Home](README.md) — Back to tutorial index
