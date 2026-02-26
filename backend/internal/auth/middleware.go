package auth

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ModeMiddleware returns a Gin middleware that enforces the active auth mode.
//
// In single-user mode the default owner is auto-injected and the request
// always proceeds without a cookie check.
//
// In all other modes a valid session cookie is required. The implementation
// shares the same session table and cookie name as the existing
// middleware.AuthMiddleware so no schema changes are needed.
func ModeMiddleware(pool *pgxpool.Pool, mode AuthMode, singleSession *SingleUserSession) gin.HandlerFunc {
	switch mode {
	case AuthModeSingle:
		return singleUserMiddleware(pool, singleSession)
	default:
		return requireSessionMiddleware(pool)
	}
}

// singleUserMiddleware injects the permanent owner user into every request
// context without checking cookies. If the singleSession hasn't been
// initialised yet it falls back to a DB lookup so startup order doesn't
// matter.
func singleUserMiddleware(pool *pgxpool.Pool, sess *SingleUserSession) gin.HandlerFunc {
	return func(c *gin.Context) {
		if sess == nil {
			slog.Error("single-user session not initialised — falling back to DB lookup")
			injectUserFromDB(c, pool, singleUserID)
			return
		}
		user := &middleware.BetterAuthUser{
			ID:            sess.UserID,
			Name:          singleUserName,
			Email:         singleUserEmail,
			EmailVerified: true,
			CreatedAt:     time.Time{},
			UpdatedAt:     time.Time{},
		}
		c.Set(middleware.UserContextKey, user)
		c.Set("user_id", user.ID)
		c.Next()
	}
}

// requireSessionMiddleware validates the better-auth.session_token cookie
// against the session table. It is intentionally thin — the full Redis-cached
// variant lives in middleware.CachedAuthMiddleware. This one is used when
// mode-aware routing is needed at the auth package level.
func requireSessionMiddleware(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(middleware.SessionCookieName)
		if err != nil || cookie == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Not authenticated",
				"code":  "UNAUTHENTICATED",
			})
			return
		}

		cookie, err = url.QueryUnescape(cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid session cookie"})
			return
		}

		// Strip Better Auth HMAC signature (format: token.signature).
		token := cookie
		if idx := strings.Index(cookie, "."); idx != -1 {
			token = cookie[:idx]
		}

		injectUserFromDB(c, pool, token)
	}
}

// injectUserFromDB resolves a user from the session table using the provided
// token (in single-user mode) or validates the token against the session table
// (in multi-user modes). The single-user path skips the session table check
// and looks up the user directly.
func injectUserFromDB(c *gin.Context, pool *pgxpool.Pool, tokenOrUserID string) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var user middleware.BetterAuthUser
	var sessionExpiresAt time.Time

	err := pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u."emailVerified", u.image, u."createdAt", u."updatedAt", s."expiresAt"
		FROM session s
		JOIN "user" u ON s."userId" = u.id
		WHERE s.token = $1 AND s."expiresAt" > NOW()
	`, tokenOrUserID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.EmailVerified,
		&user.Image,
		&user.CreatedAt,
		&user.UpdatedAt,
		&sessionExpiresAt,
	)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired session",
			"code":  "UNAUTHENTICATED",
		})
		return
	}

	c.Set(middleware.UserContextKey, &user)
	c.Set("user_id", user.ID)
	c.Next()
}

// SetupGuard redirects unauthenticated requests to /setup when the application
// has no users yet. This middleware should sit in front of the login/register
// routes in local/oauth modes.
func SetupGuard(pool *pgxpool.Pool, mode AuthMode) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mode == AuthModeSingle {
			c.Next()
			return
		}

		status, err := CheckSetupStatus(c.Request.Context(), pool, mode)
		if err != nil {
			slog.Error("setup guard: failed to check setup status", "error", err)
			c.Next()
			return
		}

		if status.NeedsSetup && !strings.HasPrefix(c.Request.URL.Path, "/api/auth/setup") {
			c.AbortWithStatusJSON(http.StatusPreconditionRequired, gin.H{
				"error":     "First-boot setup required",
				"code":      "SETUP_REQUIRED",
				"setup_url": "/setup",
			})
			return
		}

		c.Next()
	}
}
