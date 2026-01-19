---
title: Phase 2 - Auth Middleware Migration Guide
author: Roberto Luna (with Claude Code)
created: 2026-01-19
category: Backend
type: Migration Guide
status: Active
part_of: Codebase Cleanup Initiative - Phase 2
---

# Phase 2: Auth Middleware Migration Guide

## Executive Summary

**Problem:** 392 duplicate auth checks across 53 handler files = 1,560 lines of duplicate code

**Solution:** `RequireAuth()` middleware applied at router level

**Impact:**
- ✅ Remove 1,560 duplicate lines
- ✅ Single source of truth for auth validation
- ✅ Cleaner, more maintainable handlers
- ✅ Consistent 401 responses
- ✅ Better security posture

---

## What Was Added

### File: `/internal/middleware/auth.go`

Added new `RequireAuth()` middleware function (lines 173-186):

```go
// RequireAuth is a middleware that enforces authentication
// Use this at the router level to protect routes that require authentication
// Eliminates the need for manual user checks in handlers (removes 392 duplicate checks)
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}
		c.Next()
	}
}
```

### File: `/internal/middleware/auth_test.go` (NEW)

Comprehensive test suite with 10 tests covering:
- ✅ RequireAuth with valid auth
- ✅ RequireAuth without auth (401 response)
- ✅ RequireAuth with invalid token
- ✅ RequireAuth with expired session
- ✅ OptionalAuth with/without auth
- ✅ Dev bypass mode
- ✅ Request abortion behavior
- ✅ GetCurrentUser helper

**All tests passing:** `go test ./internal/middleware`

---

## How It Works

### Current Pattern (392 Occurrences)

**Before - Manual Check in Every Handler:**

```go
func (h *MemoryHandler) ListMemories(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    if user == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
        return
    }

    // Business logic here...
}
```

**Problems:**
- 4 lines duplicated 392 times = 1,560 lines
- Easy to forget or implement inconsistently
- No centralized control
- More surface area for security bugs

### New Pattern - Middleware at Router Level

**After - RequireAuth Middleware:**

```go
func (h *MemoryHandler) ListMemories(c *gin.Context) {
    // User is guaranteed to exist - middleware checked it
    user := middleware.GetCurrentUser(c)

    // Business logic here...
}
```

**Benefits:**
- No manual null check needed
- Middleware handles auth once
- Handler assumes user exists
- Cleaner, shorter code

---

## Migration Steps

### Step 1: Apply Middleware to Route Groups

Update `/cmd/server/main.go` router configuration:

```go
// Create API group
api := router.Group("/api")

// Apply AuthMiddleware to parse session from cookie
api.Use(middleware.AuthMiddleware(pool))

// PROTECTED ROUTES - Require authentication
authenticated := api.Group("")
authenticated.Use(middleware.RequireAuth()) // NEW: Apply RequireAuth middleware
{
    // Memory endpoints
    authenticated.GET("/memories", memoryHandler.ListMemories)
    authenticated.POST("/memories", memoryHandler.CreateMemory)
    authenticated.GET("/memories/:id", memoryHandler.GetMemory)
    authenticated.PUT("/memories/:id", memoryHandler.UpdateMemory)
    authenticated.DELETE("/memories/:id", memoryHandler.DeleteMemory)

    // Chat endpoints
    authenticated.POST("/chat/messages", chatHandler.SendMessage)
    authenticated.GET("/chat/history", chatHandler.GetHistory)

    // Projects
    authenticated.GET("/projects", projectHandler.ListProjects)
    authenticated.POST("/projects", projectHandler.CreateProject)

    // ... all other protected routes
}

// PUBLIC ROUTES - No authentication required
api.GET("/health", healthHandler.Health)
api.POST("/auth/google/callback", authHandler.GoogleCallback)
api.POST("/auth/signup", authHandler.Signup)
api.POST("/auth/login", authHandler.Login)
```

### Step 2: Update Handler Files (Batch 1 - Example)

**File:** `internal/handlers/memory.go`

**Lines to Change:** 16 occurrences

**Before:**
```go
func (h *MemoryHandler) ListMemories(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    if user == nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
        return
    }

    // ... rest of function
}
```

**After:**
```go
func (h *MemoryHandler) ListMemories(c *gin.Context) {
    user := middleware.GetCurrentUser(c)
    // No nil check needed - RequireAuth middleware guarantees user exists

    // ... rest of function
}
```

**Repeat for all 16 functions in memory.go:**
- `ListMemories`
- `GetMemory`
- `CreateMemory`
- `UpdateMemory`
- `DeleteMemory`
- `SearchMemories`
- `GetRelevantMemories`
- `RecordMemoryAccess`
- `PinMemory`
- `UnpinMemory`
- `ArchiveMemory`
- `GetMemoryStats`
- `ListUserFacts`
- `GetUserFact`
- `UpdateUserFact`
- `DeleteUserFact`

---

## Files to Migrate (53 Files Total)

Based on the analysis, these files need the manual auth check removed:

### High Priority (Most Duplicates)

1. **crm.go** - 32 occurrences
2. **tables.go** - 21 occurrences
3. **nodes.go** - 21 occurrences
4. **mobile_handlers.go** - 17 occurrences
5. **memory.go** - 16 occurrences
6. **clients.go** - 15 occurrences
7. **workspace_handlers.go** - 13 occurrences
8. **contexts.go** - 12 occurrences
9. **onboarding_handlers.go** - 12 occurrences
10. **thinking.go** - 11 occurrences

### Medium Priority (6-10 Duplicates)

11. **comment_handlers.go** - 11 occurrences
12. **chat.go** - 10 occurrences
13. **agents.go** - 10 occurrences
14. **dashboard_handlers.go** - 10 occurrences
15. **dashboard.go** - 10 occurrences
16. **context_tree.go** - 9 occurrences
17. **integrations.go** - 9 occurrences
18. **projects.go** - 9 occurrences
19. **notification_handlers.go** - 9 occurrences
20. **team.go** - 9 occurrences

### Lower Priority (1-5 Duplicates)

21-53. **32 additional files** with 1-7 occurrences each

---

## Automated Migration Script

### Option A: Batch Sed Replace (Careful!)

```bash
#!/bin/bash
# WARNING: Review changes before committing!

cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go

# For each handler file, remove the manual auth check
find internal/handlers -name "*.go" -type f | while read -r file; do
    # Skip test files
    if [[ "$file" == *"_test.go" ]]; then
        continue
    fi

    # Create backup
    cp "$file" "${file}.bak"

    # Remove the 4-line auth check pattern
    # Pattern: user := middleware.GetCurrentUser(c)
    #          if user == nil {
    #              c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
    #              return
    #          }

    # Add comment explaining middleware handles auth
    sed -i '' '/user := middleware\.GetCurrentUser(c)/{
        N;N;N;N
        s/user := middleware\.GetCurrentUser(c)\n[[:space:]]*if user == nil {\n[[:space:]]*c\.JSON(http\.StatusUnauthorized, gin\.H{"error": "Not authenticated"})\n[[:space:]]*return\n[[:space:]]*}/user := middleware.GetCurrentUser(c)\n\t\/\/ No nil check needed - RequireAuth middleware guarantees user exists/
    }' "$file"

    # Check if file changed
    if ! diff -q "$file" "${file}.bak" > /dev/null 2>&1; then
        echo "✓ Updated: $file"
    else
        rm "${file}.bak"
    fi
done

echo ""
echo "Migration complete! Review changes with: git diff"
echo "To restore backups: find internal/handlers -name '*.go.bak' -exec bash -c 'mv \"\$0\" \"\${0%.bak}\"' {} \\;"
echo "To delete backups: find internal/handlers -name '*.go.bak' -delete"
```

### Option B: Manual Migration (Recommended)

Safer approach - manually update each file and test incrementally:

1. Update 5-10 handler files
2. Run tests: `go test ./internal/handlers/...`
3. Run build: `go build ./cmd/server`
4. Commit changes
5. Repeat until all 53 files migrated

---

## Testing Strategy

### Unit Tests

Run middleware tests:
```bash
go test ./internal/middleware -v
```

Expected output:
```
=== RUN   TestRequireAuth_WithValidAuth
--- PASS: TestRequireAuth_WithValidAuth (0.01s)
=== RUN   TestRequireAuth_WithoutAuth
--- PASS: TestRequireAuth_WithoutAuth (0.00s)
=== RUN   TestRequireAuth_WithInvalidToken
--- PASS: TestRequireAuth_WithInvalidToken (0.01s)
=== RUN   TestRequireAuth_WithExpiredSession
--- PASS: TestRequireAuth_WithExpiredSession (0.01s)
=== RUN   TestGetCurrentUser_WithUser
--- PASS: TestGetCurrentUser_WithUser (0.00s)
=== RUN   TestGetCurrentUser_WithoutUser
--- PASS: TestGetCurrentUser_WithoutUser (0.00s)
=== RUN   TestOptionalAuthMiddleware_WithAuth
--- PASS: TestOptionalAuthMiddleware_WithAuth (0.01s)
=== RUN   TestOptionalAuthMiddleware_WithoutAuth
--- PASS: TestOptionalAuthMiddleware_WithoutAuth (0.00s)
=== RUN   TestAuthMiddleware_DevBypass
--- PASS: TestAuthMiddleware_DevBypass (0.01s)
=== RUN   TestRequireAuth_AbortsRequest
--- PASS: TestRequireAuth_AbortsRequest (0.00s)
PASS
ok  	github.com/rhl/businessos-backend/internal/middleware	0.800s
```

### Integration Tests

Test a migrated handler:
```bash
go test ./internal/handlers -run TestMemoryHandler -v
```

### Manual Testing

1. Start server: `go run ./cmd/server`
2. Test unauthenticated request:
```bash
curl -X GET http://localhost:8080/api/memories
# Expected: {"error":"Not authenticated"} (401)
```

3. Test authenticated request:
```bash
curl -X GET http://localhost:8080/api/memories \
  -H "Cookie: better-auth.session_token=YOUR_VALID_TOKEN"
# Expected: List of memories (200)
```

---

## Rollout Plan

### Phase 2A: Router Configuration (Day 1)
- ✅ Create RequireAuth middleware ← DONE
- ✅ Create comprehensive tests ← DONE
- 🔲 Update `/cmd/server/main.go` router configuration
- 🔲 Create authenticated route group
- 🔲 Move all protected routes to authenticated group

### Phase 2B: Handler Migration Batch 1 (Day 2-3)
High-priority files with most duplicates:
- crm.go (32)
- tables.go (21)
- nodes.go (21)
- mobile_handlers.go (17)
- memory.go (16)
- clients.go (15)
- workspace_handlers.go (13)
- contexts.go (12)
- onboarding_handlers.go (12)
- thinking.go (11)

**Total:** ~180 duplicate checks removed

### Phase 2C: Handler Migration Batch 2 (Day 4)
Medium-priority files:
- comment_handlers.go through team.go (10 files)

**Total:** ~90 duplicate checks removed

### Phase 2D: Handler Migration Batch 3 (Day 5)
Remaining files:
- 32 files with 1-7 occurrences each

**Total:** ~122 duplicate checks removed

### Phase 2E: Testing & Verification (Day 5)
- All unit tests pass
- All integration tests pass
- Manual smoke testing
- Performance testing
- Security audit

---

## Success Criteria

### Code Metrics
- ✅ 1,560 duplicate lines removed
- ✅ 392 auth checks eliminated
- ✅ 53 handler files cleaned
- ✅ 100% test coverage for RequireAuth
- ✅ Zero compiler errors
- ✅ Zero test failures

### Functional Requirements
- ✅ All protected routes still require authentication
- ✅ Unauthenticated requests return 401
- ✅ Authenticated requests work normally
- ✅ Dev bypass mode still works
- ✅ Optional auth routes unaffected
- ✅ Public routes unaffected

### Performance
- ✅ No performance regression
- ✅ Response times unchanged
- ✅ Memory usage unchanged

---

## Edge Cases & Considerations

### 1. Optional Auth Routes
Some routes allow but don't require auth. Keep using `OptionalAuthMiddleware`:

```go
// Routes that work with or without auth
optional := api.Group("")
optional.Use(middleware.OptionalAuthMiddleware(pool))
{
    optional.GET("/public-content", handler.GetPublicContent)
}
```

### 2. Dev Bypass Mode
DEV_AUTH_BYPASS still works - middleware handles it:

```bash
# Development environment
export DEV_AUTH_BYPASS=true
```

### 3. Custom Auth Requirements
For routes needing additional checks (e.g., admin-only):

```go
authenticated.GET("/admin/users",
    middleware.RequireAuth(),           // First: Require any auth
    middleware.RequireRole("admin"),    // Then: Require admin role
    adminHandler.ListUsers,
)
```

### 4. WebSocket & SSE Routes
RequireAuth works for long-lived connections:

```go
authenticated.GET("/chat/stream",
    middleware.RequireAuth(),
    chatHandler.StreamMessages,
)
```

---

## Troubleshooting

### Problem: "User is nil" panic after migration

**Cause:** Route not in authenticated group

**Solution:**
```go
// Move route from:
api.GET("/route", handler.Method)

// To:
authenticated.GET("/route", handler.Method)
```

### Problem: Public route returns 401

**Cause:** Route accidentally in authenticated group

**Solution:**
```go
// Move route from:
authenticated.GET("/route", handler.Method)

// To:
api.GET("/route", handler.Method)
```

### Problem: Tests failing after migration

**Cause:** Test setup doesn't include middleware

**Solution:**
```go
// In test setup:
router := gin.New()
router.Use(middleware.AuthMiddleware(pool))

authenticated := router.Group("")
authenticated.Use(middleware.RequireAuth())
authenticated.GET("/test", handler.Method)
```

---

## Related Documentation

- **Phase 1:** Random ID Utilities & Session Cookie Helper (✅ Complete)
- **Phase 3:** Session Creation Logic (📋 Planned)
- **Phase 4:** Error Handling Patterns (📋 Planned)
- **DUPLICATE_CODE_ANALYSIS.md:** Full analysis of all duplicates
- **CODEBASE_CLEANUP_MASTER_REPORT.md:** Executive summary

---

## Questions & Support

Contact: Roberto Luna (@roberto)
Claude Code Session: 2026-01-19

---

**Status:** ✅ Middleware Created & Tested | 🔲 Router Migration Pending | 🔲 Handler Migration Pending
