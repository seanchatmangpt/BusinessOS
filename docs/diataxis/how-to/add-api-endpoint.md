# How To: Add a New API Endpoint

> **Add a REST endpoint to the BusinessOS backend.**
>
> Problem: You need a new endpoint like `GET /api/users/{id}/profile` — with handler, database query, routing, and tests.

---

## Quick Start

Add a new endpoint in 5 steps:

```bash
# Step 1: Create handler in handlers/<domain>.go
# Step 2: Wire the handler into the Handlers struct (handlers/handlers.go)
# Step 3: Add route to a register*Routes method in handlers/routes_<domain>.go
# Step 4: Add database query if needed (write SQL in queries/, run sqlc generate)
# Step 5: Test with curl
```

---

## Step 1: Create the Handler

Handlers live in `desktop/backend-go/internal/handlers/`. Follow the layering pattern:
**Handler → Service → Repository → Database**

Each handler type holds its own dependencies. Prefer injecting `*pgxpool.Pool`, `*config.Config`,
or specific service structs — not a single god-object. Look at the existing constructors for the
pattern your domain needs:

```go
package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// UserProfileHandler handles user profile operations.
type UserProfileHandler struct {
	pool *pgxpool.Pool
}

// NewUserProfileHandler creates a new UserProfileHandler.
// Accepts the dependencies it needs directly — not a *Handlers wrapper.
func NewUserProfileHandler(pool *pgxpool.Pool) *UserProfileHandler {
	return &UserProfileHandler{pool: pool}
}

// GetUserProfile retrieves a user's profile by ID.
// GET /api/users/:id/profile
func (h *UserProfileHandler) GetUserProfile(c *gin.Context) {
	// Authenticate user
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Validate input: extract ID from URL param
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// WvdA soundness: every blocking call has an explicit timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5000) // 5 s
	defer cancel()

	// Query the database (parameterized — never string interpolation)
	var name, email string
	err = h.pool.QueryRow(ctx,
		`SELECT name, email FROM users WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	).Scan(&name, &email)
	if err != nil {
		slog.Error("failed to get user profile",
			slog.Any("error", err),
			slog.Int64("user_id", userID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": userID, "name": name, "email": email})
}
```

**Handler Checklist:**
- [ ] Input validation at handler boundary (no invalid ID passes through)
- [ ] Authentication check via `middleware.GetCurrentUser`
- [ ] Every `QueryRow` / `Query` / `Exec` wrapped in `context.WithTimeout` (WvdA deadlock freedom)
- [ ] Parameterized queries only (`$1`, never `fmt.Sprintf`)
- [ ] `slog.Error` (not `fmt.Println`) for structured error logging
- [ ] Return appropriate HTTP status (200, 400, 401, 404, 500)

---

## Step 2: Implement a Stub First (501 Pattern)

If the handler needs a service or LLM integration that is not yet available, return
`501 Not Implemented` immediately instead of faking a success response. This follows the
Armstrong Let-It-Crash principle: fail visibly, never silently.

```go
// CreateUserProfile creates a user profile (stub — real implementation pending)
func (h *UserProfileHandler) CreateUserProfile(c *gin.Context) {
	// ARMSTRONG: Let-It-Crash — return 501 until request validation
	// and idempotency key logic are implemented.
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":  "create user profile not implemented",
		"reason": "requires request validation (name, bio) and idempotency key",
	})
}
```

Replace the stub with a real implementation only after the service layer exists and a
failing test (Red) has been written.

---

## Step 3: Register the Route

Routes are **not** registered in `main.go`. The registration flow is:

```
cmd/server/routes.go   → registerRoutes() calls app.handlers.RegisterRoutes(api)
handlers/routes.go     → Handlers.RegisterRoutes() calls h.register<Domain>Routes()
handlers/routes_<domain>.go  → register<Domain>Routes() mounts individual paths
```

### 3a. Add a `register*Routes` helper (or extend an existing one)

Create or update a domain-scoped file such as `handlers/routes_users.go`:

```go
package handlers

import "github.com/gin-gonic/gin"

// registerUserRoutes wires /api/profile, /api/users, /api/mcp routes.
func (h *Handlers) registerUserRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	profileH := NewUserProfileHandler(h.pool)

	users := api.Group("/users")
	{
		// Public
		users.GET("/check-username/:username", profileH.CheckUsernameAvailability)

		// Protected
		users.GET("/:id/profile", auth, profileH.GetUserProfile)
		users.POST("/:id/profile", auth, profileH.CreateUserProfile)
	}
}
```

### 3b. Call your helper from `handlers/routes.go`

`handlers/routes.go` already calls `h.registerUserRoutes(api, auth)`. If you added a new
domain (e.g., billing), add one line to `RegisterRoutes`:

```go
func (h *Handlers) RegisterRoutes(api *gin.RouterGroup) {
	// ... existing calls ...
	h.registerBillingRoutes(api, auth)   // ← add here
}
```

### 3c. Both `/api` and `/api/v1` are wired automatically

`cmd/server/routes.go` calls `app.handlers.RegisterRoutes` **twice** — once for the
deprecated `/api` group and once for `/api/v1`. You do not need to duplicate route
registrations; a single `register*Routes` call covers both.

```go
// From cmd/server/routes.go (already present — shown for context only):
app.handlers.RegisterRoutes(api)    // /api/* — deprecated, warns client
app.handlers.RegisterRoutes(apiv1) // /api/v1/* — current versioned path
```

**Routing Checklist:**
- [ ] Route mounted in a `register<Domain>Routes` method
- [ ] `register<Domain>Routes` called from `handlers/routes.go RegisterRoutes()`
- [ ] HTTP method matches semantics (GET read, POST create, PUT replace, PATCH update, DELETE remove)
- [ ] `auth` middleware passed for protected routes
- [ ] URL parameter names match `c.Param("id")` in handler

---

## Step 4: Add Database Query (if needed)

If you need a database query, define it in SQL. BusinessOS uses `sqlc` to generate
type-safe Go code from SQL.

Create or update a file under `queries/`, e.g. `queries/users.sql`:

```sql
-- name: GetUserByID :one
SELECT id, name, email, created_at, updated_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;
```

Run code generation:

```bash
cd desktop/backend-go
sqlc generate
```

This regenerates `internal/database/sqlc/querier.go` with:

```go
type Querier interface {
	GetUserByID(ctx context.Context, id int64) (User, error)
}
```

**Database Checklist:**
- [ ] SQL query uses parameterized `$1` (never string interpolation)
- [ ] Query has an sqlc directive (`:one`, `:many`, `:exec`)
- [ ] `sqlc generate` run after adding SQL
- [ ] Generated code is type-safe (compiler enforces types)
- [ ] Generated files committed alongside the `.sql` source

---

## Step 5: Test with curl

Test your new endpoint from the command line:

```bash
# Start the server
cd BusinessOS && make dev

# In another terminal, test GET
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8001/api/users/123/profile

# Expected 200 response:
# { "id": 123, "name": "John Doe", "email": "john@example.com" }

# Test stub (501):
curl -X POST -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8001/api/users/123/profile
# Expected:
# { "error": "create user profile not implemented", "reason": "..." }
```

**Test Checklist:**
- [ ] Server starts without errors (`make dev`)
- [ ] GET returns 200 with expected body
- [ ] POST stub returns 501 until implemented
- [ ] Invalid ID returns 400 Bad Request
- [ ] Missing auth header returns 401 Unauthorized

---

## Writing Integration Tests

Add tests in `handlers/<domain>_test.go`:

```go
package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUserProfile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler with test pool (or use httptest + real DB in integration tests)
	handler := NewUserProfileHandler(testPool)

	// Build request
	req := httptest.NewRequest(http.MethodGet, "/api/users/123/profile", nil)
	w := httptest.NewRecorder()

	r := gin.New()
	r.GET("/api/users/:id/profile", handler.GetUserProfile)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "email")
}

func TestGetUserProfile_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewUserProfileHandler(testPool)

	req := httptest.NewRequest(http.MethodGet, "/api/users/not-a-number/profile", nil)
	w := httptest.NewRecorder()

	r := gin.New()
	r.GET("/api/users/:id/profile", handler.GetUserProfile)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

Run tests:

```bash
cd desktop/backend-go
go test ./internal/handlers/... -run TestGetUserProfile -v
```

---

## Common Errors

### Error: "cannot find type User in database"
**Cause:** Forgot to run `sqlc generate` after adding SQL query.
**Fix:** Run `sqlc generate` in `desktop/backend-go/`, then rebuild.

### Error: "handler not found" when calling endpoint
**Cause:** Route not registered — either the `register<Domain>Routes` call is missing from
`handlers/routes.go`, or the route group path has a typo.
**Fix:** Add or verify the call in `Handlers.RegisterRoutes()` in `handlers/routes.go`.

### Error: "missing Authorization header"
**Cause:** Applied auth middleware but did not send a JWT token.
**Fix:** Add `Authorization: Bearer YOUR_TOKEN` header to the request, or call
`/api/auth/login` first to obtain a token.

### Error: "501 Not Implemented"
**Cause:** You called a stub endpoint. The handler intentionally returns 501 until the
service layer is wired.
**Fix:** Implement the service logic, write the Red test first, then replace the stub body.

---

## Full Checklist: From Problem to Deployed

Before merging your endpoint:

- [ ] Handler validates input at boundaries
- [ ] Handler calls service (not database directly for complex logic)
- [ ] Every blocking DB call wrapped in `context.WithTimeout` (WvdA soundness)
- [ ] Stubs return `501 Not Implemented` with descriptive `reason` (not fake 200)
- [ ] Database query uses parameterized `$1` (no string interpolation)
- [ ] `sqlc generate` run and types match (if SQL added)
- [ ] Route registered in `handlers/routes_<domain>.go` via `register<Domain>Routes`
- [ ] `register<Domain>Routes` called from `handlers/routes.go`
- [ ] Auth middleware applied for protected routes
- [ ] Test passes locally (`make test-backend`)
- [ ] curl test passes with valid and invalid input
- [ ] Error logging uses `slog` with structured key-value pairs
- [ ] Status codes correct (200, 400, 401, 404, 500, 501)

---

## Next Steps

- **Add validation**: Use `github.com/go-playground/validator` for struct tags
- **Add pagination**: Parse `?limit=10&offset=0` query params in handler
- **Cache responses**: Use Redis via `cache.Client` for expensive queries
- **Document OpenAPI**: Add `// @Summary` comments above handlers for Swagger generation

---

*See also: [API Endpoints Reference](../reference/api-endpoints.md), [Code Standards](../../CLAUDE.md#code-standards-go-backend)*

*updated: 2026-03-27*
