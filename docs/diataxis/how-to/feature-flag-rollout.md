# How To: Add a Feature Behind a Feature Flag

> **Roll out a new feature gradually using a feature flag.**
>
> Problem: You built a new feature and want to test it with a subset of users before releasing it to everyone.

---

## Quick Start

Implement a feature flag in 4 steps:

```bash
# Step 1: Add boolean flag to config
# Step 2: Check flag in handler before executing
# Step 3: Log flag usage for analytics
# Step 4: Verify flag can be toggled at runtime (no restart needed)
```

---

## Architecture: Feature Flags in BusinessOS

Feature flags live in the **config system**, not hardcoded. This enables:
- **Zero downtime**: Change flag without restarting
- **Per-user targeting**: Different users see different flags
- **Analytics**: Track who uses flagged features
- **Rollback**: Disable broken feature instantly

---

## Step 1: Add Flag to Config

BusinessOS stores config in PostgreSQL. Flags are consulted at runtime.

### Option A: Config YAML (Development)

Create or update `config/feature-flags.yaml`:

```yaml
features:
  user_profile_redesign:
    enabled: false
    description: "New user profile UI with Signal Theory integration"
    rollout_percent: 0  # 0% = disabled for everyone

  ai_content_generation:
    enabled: true
    description: "AI-powered content suggestions"
    rollout_percent: 50  # 50% = enabled for half the users

  advanced_analytics:
    enabled: true
    description: "Advanced dashboard analytics with predictions"
    rollout_percent: 100  # 100% = enabled for everyone
    excluded_users: [1, 2, 3]  # Except these user IDs
```

### Option B: Database-Backed (Production)

For hot reload, store flags in the database:

```sql
-- Create feature_flags table
CREATE TABLE feature_flags (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) UNIQUE NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  rollout_percent INTEGER DEFAULT 0,  -- 0-100
  description TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert your feature flag
INSERT INTO feature_flags (name, enabled, rollout_percent, description)
VALUES ('user_profile_redesign', FALSE, 0, 'New user profile UI');
```

---

## Step 2: Check Flag in Handler

In your handler, check the flag before executing the feature:

```go
package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
	"github.com/rhl/businessos-backend/internal/utils"
)

type UserHandler struct {
	userService  *services.UserService
	flagService  *services.FeatureFlagService
}

// GetUserProfile retrieves user profile with optional redesigned UI.
// GET /api/users/:id/profile
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	userID := user.ID

	// Check if user should see the new profile UI
	useRedesign, err := h.flagService.IsEnabledForUser(
		c.Request.Context(),
		"user_profile_redesign",
		userID,
	)
	if err != nil {
		slog.Error("failed to check feature flag", slog.Any("error", err))
		// Safe default: use old version if flag check fails
		useRedesign = false
	}

	// Fetch profile
	profile, err := h.userService.GetProfileByID(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to get user profile", slog.Any("error", err))
		utils.RespondInternalError(c, slog.Default())
		return
	}

	// Log flag usage for analytics
	slog.Info("feature flag used",
		slog.String("flag_name", "user_profile_redesign"),
		slog.Bool("enabled", useRedesign),
		slog.Int64("user_id", userID),
	)

	// Return different response based on flag
	if useRedesign {
		c.JSON(http.StatusOK, gin.H{
			"profile": profile,
			"ui_version": "v2",  // Signal to frontend
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"profile": profile,
			"ui_version": "v1",
		})
	}
}
```

---

## Step 3: Implement Feature Flag Service

Create `services/feature_flag_service.go`:

```go
package services

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/rhl/businessos-backend/internal/database"
)

type FeatureFlagService struct {
	repo *database.FeatureFlagRepository
}

// IsEnabledForUser checks if a feature flag is enabled for a specific user.
// Takes into account:
// 1. Flag enabled status
// 2. Rollout percentage (0-100)
// 3. Excluded users
func (s *FeatureFlagService) IsEnabledForUser(
	ctx context.Context,
	flagName string,
	userID int64,
) (bool, error) {
	// Fetch flag from database (or cache)
	flag, err := s.repo.GetByName(ctx, flagName)
	if err != nil {
		return false, fmt.Errorf("get feature flag: %w", err)
	}

	// If flag not found, default to disabled
	if flag == nil {
		return false, nil
	}

	// If globally disabled, return false
	if !flag.Enabled {
		return false, nil
	}

	// Check if user is excluded
	if flag.IsUserExcluded(userID) {
		slog.Debug("user excluded from flag", slog.String("flag", flagName), slog.Int64("user_id", userID))
		return false, nil
	}

	// Check rollout percentage (deterministic per user)
	// Hash user ID to get consistent result across requests
	userHash := int64(userID) % 100
	if userHash >= int64(flag.RolloutPercent) {
		return false, nil
	}

	// Flag is enabled for this user
	return true, nil
}

// ToggleFlag enables or disables a flag (hot reload, no restart needed)
func (s *FeatureFlagService) ToggleFlag(
	ctx context.Context,
	flagName string,
	enabled bool,
	rolloutPercent int,
) error {
	err := s.repo.Update(ctx, flagName, enabled, rolloutPercent)
	if err != nil {
		return fmt.Errorf("update feature flag: %w", err)
	}

	slog.Info("feature flag toggled",
		slog.String("flag", flagName),
		slog.Bool("enabled", enabled),
		slog.Int("rollout_percent", rolloutPercent),
	)

	return nil
}
```

---

## Step 4: Create Admin API to Toggle Flags

Add an admin handler to control flags without restart:

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

// ToggleFeatureFlag enables/disables a feature flag at runtime.
// POST /api/admin/flags/:name/toggle
// Body: { "enabled": true, "rollout_percent": 50 }
func (h *AdminHandler) ToggleFeatureFlag(c *gin.Context) {
	// Require admin role
	user := middleware.GetCurrentUser(c)
	if user == nil || !user.IsAdmin {
		utils.RespondForbidden(c, "admin role required", slog.Default())
		return
	}

	flagName := c.Param("name")

	var req struct {
		Enabled         bool `json:"enabled"`
		RolloutPercent  int  `json:"rollout_percent"`
	}

	if err := c.BindJSON(&req); err != nil {
		utils.RespondBadRequest(c, "invalid request body", slog.Default())
		return
	}

	// Validate rollout percent (0-100)
	if req.RolloutPercent < 0 || req.RolloutPercent > 100 {
		utils.RespondBadRequest(c, "rollout_percent must be 0-100", slog.Default())
		return
	}

	// Toggle the flag
	err := h.flagService.ToggleFlag(c.Request.Context(), flagName, req.Enabled, req.RolloutPercent)
	if err != nil {
		slog.Error("failed to toggle flag", slog.Any("error", err))
		utils.RespondInternalError(c, slog.Default())
		return
	}

	// Log for audit trail
	slog.Info("feature flag toggled by admin",
		slog.String("flag", flagName),
		slog.Bool("enabled", req.Enabled),
		slog.Int("rollout_percent", req.RolloutPercent),
		slog.Int64("admin_user_id", user.ID),
	)

	c.JSON(http.StatusOK, gin.H{
		"flag": flagName,
		"enabled": req.Enabled,
		"rollout_percent": req.RolloutPercent,
	})
}

// ListFeatureFlags returns all feature flags (admin only).
// GET /api/admin/flags
func (h *AdminHandler) ListFeatureFlags(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil || !user.IsAdmin {
		utils.RespondForbidden(c, "admin role required", slog.Default())
		return
	}

	flags, err := h.flagService.ListAll(c.Request.Context())
	if err != nil {
		slog.Error("failed to list flags", slog.Any("error", err))
		utils.RespondInternalError(c, slog.Default())
		return
	}

	c.JSON(http.StatusOK, gin.H{"flags": flags})
}
```

Register the routes:

```go
// In main.go setupRoutes()
admin := engine.Group("/api/admin")
admin.Use(middleware.AuthRequired)
admin.GET("/flags", adminHandler.ListFeatureFlags)
admin.POST("/flags/:name/toggle", adminHandler.ToggleFeatureFlag)
```

---

## Step 5: Test the Feature Flag

### Test with curl

```bash
# List all flags
curl -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  http://localhost:8001/api/admin/flags

# Enable a flag for 50% of users
curl -X POST \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled": true, "rollout_percent": 50}' \
  http://localhost:8001/api/admin/flags/user_profile_redesign/toggle

# Call endpoint to check if flag is enabled for your user
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8001/api/users/123/profile
```

### Test with Frontend

In your SvelteKit component, check the `ui_version` in response:

```svelte
<script>
  let profile = $state(null);
  let uiVersion = $state('v1');

  onMount(async () => {
    const res = await fetch('/api/users/123/profile');
    const data = await res.json();
    profile = data.profile;
    uiVersion = data.ui_version;
  });
</script>

{#if uiVersion === 'v2'}
  <!-- New redesigned UI -->
  <UserProfileV2 user={profile} />
{:else}
  <!-- Old UI -->
  <UserProfileV1 user={profile} />
{/if}
```

---

## Rollout Strategy

### Phase 1: Test with Yourself (0% Rollout)
1. Add user ID to `excluded_users` list
2. Set `rollout_percent: 0`
3. Check only your user ID passes flag check
4. Test feature manually

### Phase 2: Canary Release (10% Rollout)
1. Set `rollout_percent: 10`
2. Monitor logs for errors: `grep "user_profile_redesign" logs/`
3. Check metrics (API latency, error rate)
4. If OK, proceed to phase 3

### Phase 3: Staged Rollout (50%, 75%, 100%)
1. Increase `rollout_percent` gradually
2. Monitor each phase for 24 hours
3. If issues found, set `enabled: false` to disable instantly

### Phase 4: Full Release (100% Rollout)
1. Set `rollout_percent: 100`
2. Keep flag in code for easy rollback
3. Remove flag after 2 weeks (code cleanup)

---

## Monitoring Flag Usage

Check logs to see which users hit the flag:

```bash
# View all feature flag usage
docker-compose logs businessos-backend | grep "feature flag used"

# Output:
# 2026-03-25T10:30:45Z ... flag_name=user_profile_redesign enabled=true user_id=123
# 2026-03-25T10:30:46Z ... flag_name=user_profile_redesign enabled=false user_id=456
```

---

## Common Patterns

### Pattern: A/B Testing

Split users 50/50 between two versions:

```go
// Show version A to 50%, version B to other 50%
versionA := h.flagService.IsEnabledForUser(ctx, "experiment_v2a", userID)

if versionA {
  // A/B version A
  return sendVersionA()
} else {
  // A/B version B
  return sendVersionB()
}
```

### Pattern: Time-Based Rollout

Enable only during testing window:

```go
// Enable only Monday-Friday, 9am-5pm (for live testing)
now := time.Now()
isBusinessHours := now.Weekday() >= time.Monday && now.Weekday() <= time.Friday &&
  now.Hour() >= 9 && now.Hour() < 17

enabled := isBusinessHours && isEnabledForUser(ctx, "live_feature", userID)
```

### Pattern: Geolocation-Based

Enable only for specific regions:

```go
// Add region to flag check
isEnabledForUser := h.flagService.IsEnabledForUserInRegion(
  ctx,
  "new_payment_processor",
  userID,
  userRegion,  // "US", "EU", "APAC"
)
```

---

## Cleanup: Removing a Flag

After 2 weeks at 100% with no issues:

1. Remove flag check from handler
2. Remove flag from database
3. Remove dead code (old UI version)
4. Test again
5. Commit cleanup

---

## Full Checklist: Feature Flag Rollout

- [ ] Config created (YAML or database)
- [ ] Feature flag service implemented with `IsEnabledForUser()`
- [ ] Handler checks flag before executing feature
- [ ] Flag usage logged with user ID (for analytics)
- [ ] Admin API to toggle flag created
- [ ] Admin API requires auth/admin role
- [ ] Frontend reads `ui_version` and renders accordingly
- [ ] Tested with `curl` at 0%, 50%, 100%
- [ ] Gradual rollout plan documented (0% → 10% → 50% → 100%)
- [ ] Monitoring setup (grep logs for flag usage)
- [ ] Rollback procedure verified (disable flag instantly)
- [ ] Cleanup plan for after 2 weeks

---

## Next Steps

- **Add flag caching**: Cache flag state in Redis for performance
- **Add UI dashboard**: Admin interface to toggle flags without curl
- **Add metrics**: Export flag usage to Prometheus for analytics
- **Add webhooks**: Notify services when flag changes

---

*See also: [Code Standards](../../CLAUDE.md#code-standards-go-backend), [Admin Handlers](../reference/api-endpoints.md#admin-endpoints)*
