# How To: Add a New API Endpoint

> **Add a REST endpoint to the BusinessOS backend.**
>
> Problem: You need a new endpoint like `GET /api/users/{id}/profile` — with handler, database query, routing, and tests.

---

## Quick Start

Add a new endpoint in 5 steps:

```bash
# Step 1: Create handler in handlers/users.go
# Step 2: Add route to router in cmd/server/main.go
# Step 3: Add database query if needed (write SQL in queries/)
# Step 4: Run code generation (sqlc, wire)
# Step 5: Test with curl
```

---

## Step 1: Create the Handler

Handlers live in `desktop/backend-go/internal/handlers/`. Follow the layering pattern:
**Handler → Service → Repository → Database**

Add your handler to `handlers/users.go`:

```go
package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// GetUserProfile retrieves a user's profile by ID.
// GET /api/users/:id/profile
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	// Authenticate user
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	// Validate input: extract ID from URL param
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.RespondBadRequest(c, "invalid user id", slog.Default())
		return
	}

	// Call service layer to fetch profile
	profile, err := h.userService.GetProfileByID(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to get user profile", slog.Any("error", err), slog.Int64("user_id", userID))
		utils.RespondInternalError(c, slog.Default())
		return
	}

	// Return 200 OK with profile data
	c.JSON(http.StatusOK, profile)
}
```

**Handler Checklist:**
- [ ] Input validation at handler boundary (no invalid ID)
- [ ] Authentication check (use `middleware.GetCurrentUser`)
- [ ] Call service layer (not directly database)
- [ ] Error handling with structured logging (`slog`)
- [ ] Return appropriate HTTP status (200, 400, 404, 500)

---

## Step 2: Add Service Layer Logic

Services live in `internal/services/`. They contain business logic (validation, authorization, orchestration).

Add to `services/user_service.go`:

```go
package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/rhl/businessos-backend/internal/database"
)

type UserService struct {
	repo *database.UserRepository
}

// GetProfileByID retrieves a user's profile.
// Returns error if user not found or access denied.
func (s *UserService) GetProfileByID(ctx context.Context, userID int64) (*UserProfile, error) {
	// Call repository (database access)
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch user profile: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found: id=%d", userID)
	}

	// Transform to response DTO
	profile := &UserProfile{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	return profile, nil
}
```

**Service Checklist:**
- [ ] Business logic (validation, transformation)
- [ ] Error wrapping with context (`fmt.Errorf`)
- [ ] Calls repository for data access
- [ ] No HTTP concerns (no gin.Context, no status codes)

---

## Step 3: Add Database Query (if needed)

If you need a database query, define it in SQL. BusinessOS uses `sqlc` to generate type-safe Go code from SQL.

Create or update `queries/users.sql`:

```sql
-- name: GetUserByID :one
SELECT id, name, email, created_at, updated_at
FROM users
WHERE id = $1;
```

Run code generation:

```bash
cd desktop/backend-go
sqlc generate
```

This generates `internal/database/querier.go` with:

```go
type Querier interface {
	GetUserByID(ctx context.Context, id int64) (User, error)
}
```

**Database Checklist:**
- [ ] SQL query uses parameterized `$1` (never string interpolation)
- [ ] Query has sqlc directive (`:one`, `:many`, `:exec`)
- [ ] Run `sqlc generate` after adding SQL
- [ ] Generated code is type-safe (compiler enforces types)

---

## Step 4: Register the Route

Routes live in `cmd/server/main.go` (or in a router initialization function). Use Gin's routing:

```go
package main

import "github.com/gin-gonic/gin"

func setupRoutes(engine *gin.Engine, userHandler *handlers.UserHandler) {
	// Public routes (no auth required)
	public := engine.Group("/api")
	// public.POST("/login", userHandler.Login)

	// Protected routes (auth required)
	protected := engine.Group("/api")
	protected.Use(middleware.AuthRequired)
	protected.GET("/users/:id/profile", userHandler.GetUserProfile)
	protected.POST("/users", userHandler.CreateUser)
	protected.PUT("/users/:id", userHandler.UpdateUser)
}
```

**Routing Checklist:**
- [ ] Route matches your handler method name
- [ ] HTTP method correct (GET, POST, PUT, DELETE)
- [ ] Auth middleware applied if protected (`middleware.AuthRequired`)
- [ ] URL parameter names match handler (`c.Param("id")`)

---

## Step 5: Test with curl

Test your new endpoint from the command line:

```bash
# Start the server
cd BusinessOS && make dev

# In another terminal, test GET
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8001/api/users/123/profile

# Expected response:
# {
#   "id": 123,
#   "name": "John Doe",
#   "email": "john@example.com"
# }
```

**Test Checklist:**
- [ ] Server starts without errors (`make dev`)
- [ ] Request returns 200 OK
- [ ] Response body matches expected shape
- [ ] Try invalid ID (should return 400 or 404)
- [ ] Try without auth header (should return 401 Unauthorized)

---

## Writing Integration Tests

Add tests in `handlers/users_test.go`:

```go
package handlers_test

import (
	"testing"
	"net/http"
	"net/http/httptest"

	"github.com/rhl/businessos-backend/internal/handlers"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/gin-gonic/gin"
)

func TestGetUserProfile_Success(t *testing.T) {
	// Setup mock service
	mockService := &services.MockUserService{}
	mockService.On("GetProfileByID", mock.Anything, int64(123)).
		Return(&services.UserProfile{
			ID:    123,
			Name:  "Test User",
			Email: "test@example.com",
		}, nil)

	// Create handler
	handler := handlers.NewUserHandler(mockService)

	// Create request
	req := httptest.NewRequest("GET", "/api/users/123/profile", nil)
	w := httptest.NewRecorder()

	// Call handler
	c, _ := gin.CreateTestContextForRequest(req)
	c.Params = gin.Params{{Key: "id", Value: "123"}}
	handler.GetUserProfile(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test User")
}

func TestGetUserProfile_InvalidID(t *testing.T) {
	// Invalid ID should return 400
	handler := handlers.NewUserHandler(mockService)
	req := httptest.NewRequest("GET", "/api/users/invalid/profile", nil)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContextForRequest(req)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}
	handler.GetUserProfile(c)

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
**Cause:** Route not registered in `main.go`.
**Fix:** Add the route to `setupRoutes()` function, matching your handler method name exactly.

### Error: "missing Authorization header"
**Cause:** Applied `middleware.AuthRequired` but didn't send JWT token.
**Fix:** Add `Authorization: Bearer YOUR_TOKEN` header to request.

---

## Full Checklist: From Problem to Deployed

Before merging your endpoint:

- [ ] Handler validates input at boundaries
- [ ] Handler calls service (not database directly)
- [ ] Service contains business logic
- [ ] Service calls repository
- [ ] Database query uses parameterized `$1` (no string interpolation)
- [ ] `sqlc generate` run and types match
- [ ] Route registered in main.go with correct HTTP method
- [ ] Auth middleware applied (if protected endpoint)
- [ ] Test passes locally (`make test-backend`)
- [ ] curl test passes with valid and invalid input
- [ ] Error logging includes context (`slog`)
- [ ] Status codes correct (200, 400, 401, 404, 500)

---

## Next Steps

- **Add validation**: Use `github.com/go-playground/validator` for struct tags
- **Add pagination**: Parse `?limit=10&offset=0` query params in handler
- **Cache responses**: Use Redis via `cache.Client` for expensive queries
- **Document OpenAPI**: Add `// @Summary` comments above handlers for Swagger generation

---

*See also: [API Endpoints Reference](../reference/api-endpoints.md), [Code Standards](../../CLAUDE.md#code-standards-go-backend)*
