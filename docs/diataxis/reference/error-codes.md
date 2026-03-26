# Reference: Error Codes & Solutions

**Information-oriented.** HTTP error codes → HTTP status → description → how to fix.

**Format:** Code → HTTP Status → Meaning → Resolution

---

## 4xx Client Errors

### 400 Bad Request

**Description:** Request is invalid (missing fields, wrong format, validation failed).

**Common Causes:**
- Missing required field in request body
- Invalid JSON/form data
- Enum value not recognized
- Type mismatch (e.g., string instead of UUID)

**Response:**
```json
{
  "error": "validation_failed",
  "message": "Field 'email' is required",
  "details": {
    "field": "email",
    "reason": "required"
  }
}
```

**How to Fix:**
1. Check request body against API endpoint documentation
2. Ensure all required fields are present
3. Validate data types (strings, UUIDs, timestamps, enums)
4. Check for typos in field names
5. For JSON, use valid JSON formatting (test with `jq` or online validator)

**Examples:**
- Missing `title` in POST /api/v1/projects: Add `"title": "Project Name"`
- Invalid UUID: Use valid UUID format `550e8400-e29b-41d4-a716-446655440000`
- Invalid enum: Use `status: "ACTIVE"` not `status: "active"`

---

### 400 Bad Request — Invalid Token

**Description:** Request body or query parameter invalid.

**Response:**
```json
{
  "error": "invalid_token",
  "message": "Token malformed or expired"
}
```

**How to Fix:**
1. Get fresh token: POST /api/v1/auth/login
2. Add token to request: `Authorization: Bearer <token>`
3. Check token is valid base64 JWT
4. Verify token hasn't expired (check `exp` claim)

---

### 401 Unauthorized

**Description:** Missing or invalid authentication credentials.

**Common Causes:**
- Missing `Authorization` header
- Invalid or expired JWT token
- Wrong token format
- Token signing secret changed

**Response:**
```json
{
  "error": "unauthorized",
  "message": "Invalid or expired token"
}
```

**How to Fix:**
1. **Missing header:** Add `Authorization: Bearer <token>` to request
2. **Expired token:** Re-authenticate with POST /api/v1/auth/login
3. **Invalid token:**
   - Check token format: `Authorization: Bearer eyJhbGc...`
   - Verify token is complete (not truncated)
   - Re-generate if corrupted
4. **Server-side:** Check `SECRET_KEY` env var matches across all instances

**Debug Steps:**
```bash
# Decode JWT (without verification)
echo "eyJhbGc..." | jq -R 'split(".")[1] | @base64d | fromjson'

# Check expiration
echo "eyJhbGc..." | jq -R 'split(".")[1] | @base64d | fromjson | .exp'
```

---

### 403 Forbidden

**Description:** Authenticated but lacks permission for resource.

**Common Causes:**
- User not admin (trying admin endpoint)
- User not owner of resource
- Resource belongs to different user/organization

**Response:**
```json
{
  "error": "forbidden",
  "message": "You do not have permission to access this resource"
}
```

**How to Fix:**
1. Verify user role is sufficient (check user profile)
2. Check resource ownership (user_id matches current user)
3. For admin endpoints, use admin account
4. For shared resources, verify sharing permissions

---

### 404 Not Found

**Description:** Resource does not exist.

**Common Causes:**
- Wrong resource ID (typo, wrong UUID)
- Resource deleted
- Resource belongs to different user (filtering)
- Wrong endpoint path

**Response:**
```json
{
  "error": "not_found",
  "message": "Project with id 'abc123' not found"
}
```

**How to Fix:**
1. Verify resource ID is correct (copy from list endpoint)
2. Check resource still exists (GET /api/v1/{resource}/{id})
3. Verify you own/have access to resource
4. Check endpoint path spelling
5. Confirm resource type (is it a project or task?)

**Common Typos:**
- `/projects/123` vs `/project/123` (singular vs plural)
- `/tasks/:id/comment` vs `/tasks/:id/comments` (singular vs plural)
- `projects` vs `porjects` (spelling)

---

### 409 Conflict

**Description:** Resource conflict prevents operation.

**Common Causes:**
- Duplicate unique field (email, share_id, etc.)
- Concurrent modification (optimistic lock)
- Invalid state transition (e.g., delete archived project)

**Response:**
```json
{
  "error": "conflict",
  "message": "Email 'user@example.com' is already in use"
}
```

**How to Fix:**
1. **Duplicate:** Use different value for unique field
   - Email: Check existing users before creating new account
   - Unique name: Append timestamp or random suffix
2. **Concurrent modification:** Retry operation (idempotent)
3. **Invalid state:** Check resource current state before modifying

---

## 5xx Server Errors

### 500 Internal Server Error

**Description:** Unexpected server error (bug or configuration issue).

**Response:**
```json
{
  "error": "internal_server_error",
  "message": "An unexpected error occurred",
  "request_id": "req_12345"
}
```

**How to Fix:**
1. **Check logs:** Request ID in response → find in server logs
2. **Retry:** Transient error? Wait 30s and retry
3. **Check dependencies:**
   - Database: `GET /health` shows database connected?
   - Redis: Check Redis is running and accessible
   - External APIs: Google OAuth, pm4py-rust, etc. up?
4. **Report bug:** Include request ID, request body, and exact error message

**Common Causes:**
- Database connection lost
- Redis unavailable
- Missing environment variable
- Unhandled exception in business logic
- External service timeout

---

### 503 Service Unavailable

**Description:** Service temporarily unavailable (maintenance, overload, dependency down).

**Response:**
```json
{
  "error": "service_unavailable",
  "message": "Service temporarily unavailable. Please try again later."
}
```

**How to Fix:**
1. **Check status:** GET /health/detailed
   - If database unavailable: Wait for DB recovery or contact ops
   - If Redis unavailable: Restart Redis
   - If OSA unavailable: Check OSA service logs
2. **Retry strategy:** Exponential backoff (wait 1s, 2s, 4s, 8s... up to 60s)
3. **Wait for recovery:** Typically resolves within 5 minutes

**If Persistent:**
1. Check server logs for startup errors
2. Verify all services running (docker-compose ps)
3. Restart failed service: `docker-compose restart <service>`

---

## Error Code Categories

### Authentication Errors (401)

| Error | Fix |
|-------|-----|
| Missing Authorization header | Add `Authorization: Bearer <token>` |
| Invalid token format | Ensure format is `Authorization: Bearer eyJ...` |
| Expired token | Call POST /api/v1/auth/login for fresh token |
| Wrong signing secret | Verify `SECRET_KEY` matches server config |

---

### Validation Errors (400)

| Error | Fix |
|-------|-----|
| Missing required field | Add field to request body |
| Invalid UUID format | Use valid UUID: `550e8400-e29b-41d4-a716-446655440000` |
| Invalid enum value | Check allowed values in API docs |
| Invalid timestamp format | Use ISO8601: `2026-03-25T10:00:00Z` |
| Invalid JSON | Validate with `jq` or online JSON validator |

---

### Permission Errors (403)

| Error | Fix |
|-------|-----|
| Not authenticated | Add valid JWT token |
| Not admin | Use admin account or request permission |
| Not resource owner | Verify user_id matches current user |
| Resource archived/deleted | Unarchive or restore resource |

---

### Not Found Errors (404)

| Error | Fix |
|-------|-----|
| Wrong resource ID | Copy ID from GET list endpoint |
| Resource deleted | Check if resource was archived instead |
| Wrong endpoint path | Verify spelling: `/projects` not `/porjects` |
| Wrong resource type | Confirm you're using correct endpoint |

---

### Conflict Errors (409)

| Error | Fix |
|-------|-----|
| Duplicate unique field | Use different value (email, name, etc.) |
| Concurrent modification | Retry operation |
| Invalid state transition | Check current resource state first |

---

### Server Errors (500, 503)

| Error | Fix |
|-------|-----|
| Database disconnected | Wait for recovery or restart database |
| Redis unavailable | Restart Redis: `docker-compose restart redis` |
| Out of memory | Check server resources, increase allocation |
| Timeout on external API | Retry with backoff, check external service status |
| Unknown exception | Check logs with request ID, report bug |

---

## Troubleshooting Decision Tree

```
Error received
  ├─ 400 Bad Request?
  │  ├─ Validation failed?
  │  │  └─ Check request body against API docs
  │  └─ Invalid token?
  │     └─ Call POST /api/v1/auth/login
  │
  ├─ 401 Unauthorized?
  │  ├─ Missing Authorization header?
  │  │  └─ Add: Authorization: Bearer <token>
  │  └─ Expired token?
  │     └─ Call POST /api/v1/auth/login
  │
  ├─ 403 Forbidden?
  │  ├─ Not admin?
  │  │  └─ Use admin account
  │  └─ Not resource owner?
  │     └─ Verify user_id matches
  │
  ├─ 404 Not Found?
  │  ├─ Wrong resource ID?
  │  │  └─ Copy from GET list endpoint
  │  ├─ Resource deleted?
  │  │  └─ Check if archived instead
  │  └─ Wrong endpoint?
  │     └─ Verify spelling (e.g., /projects not /porjects)
  │
  ├─ 409 Conflict?
  │  ├─ Duplicate field?
  │  │  └─ Use different value
  │  └─ Concurrent modification?
  │     └─ Retry operation
  │
  └─ 5xx Server Error?
     ├─ GET /health/detailed
     │  ├─ Database unavailable?
     │  │  └─ Wait for recovery or restart database
     │  ├─ Redis unavailable?
     │  │  └─ Restart Redis
     │  └─ OSA unavailable?
     │     └─ Check OSA service logs
     └─ Check server logs with request ID
```

---

## Common Error Scenarios

### Scenario 1: "Invalid or expired token"
```
Error: 401 Unauthorized
Message: "Invalid or expired token"
```

**Steps:**
1. Get fresh token: `curl -X POST http://localhost:8001/api/v1/auth/login -d '{"email":"user@example.com","password":"..."}'`
2. Use new token: `Authorization: Bearer <new_token>`
3. If still fails, verify SECRET_KEY in .env matches

---

### Scenario 2: "Field 'title' is required"
```
Error: 400 Bad Request
Message: "Field 'title' is required"
```

**Steps:**
1. Check request body: `curl -X POST ... -d '{"title":"My Project",...}'`
2. Verify all required fields in request (check API docs)
3. Use valid JSON format

---

### Scenario 3: "Project with id 'abc' not found"
```
Error: 404 Not Found
Message: "Project with id 'abc' not found"
```

**Steps:**
1. Verify ID format is valid UUID (not short ID like 'abc')
2. Get correct ID: `curl http://localhost:8001/api/v1/projects -H "Authorization: Bearer <token>"`
3. Copy full UUID from response
4. Try again with correct UUID

---

### Scenario 4: "You do not have permission"
```
Error: 403 Forbidden
Message: "You do not have permission to access this resource"
```

**Steps:**
1. Check current user: `curl http://localhost:8001/api/v1/users/me -H "Authorization: Bearer <token>"`
2. Verify user owns resource (user_id field)
3. For admin endpoints, use admin account
4. Check resource sharing/permissions

---

### Scenario 5: "Service temporarily unavailable"
```
Error: 503 Service Unavailable
Message: "Service temporarily unavailable"
```

**Steps:**
1. Check status: `curl http://localhost:8001/health/detailed`
2. If database unavailable: Wait or restart database
3. If Redis unavailable: `docker-compose restart redis`
4. Retry request with exponential backoff

---

## HTTP Status Code Summary

| Code | Category | Meaning |
|------|----------|---------|
| 200 | Success | Request succeeded |
| 201 | Success | Resource created |
| 204 | Success | Request succeeded (no body) |
| 202 | Success | Request accepted (async) |
| 400 | Client Error | Bad request (validation failed) |
| 401 | Client Error | Unauthorized (auth failed) |
| 403 | Client Error | Forbidden (permission denied) |
| 404 | Client Error | Not found |
| 409 | Client Error | Conflict |
| 500 | Server Error | Internal error |
| 503 | Server Error | Service unavailable |

---

## See Also

- API endpoints reference
- Configuration options reference
- How-to: Authenticate with OAuth
- How-to: Debug API requests
