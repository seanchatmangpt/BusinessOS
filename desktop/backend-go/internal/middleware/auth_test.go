package middleware

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a test router with auth middleware
func setupTestRouter(pool *pgxpool.Pool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Public route (no auth)
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public"})
	})

	// Protected route group with RequireAuth
	auth := r.Group("/protected")
	auth.Use(AuthMiddleware(pool))
	auth.Use(RequireAuth())
	{
		auth.GET("/resource", func(c *gin.Context) {
			user := GetCurrentUser(c)
			c.JSON(http.StatusOK, gin.H{
				"message": "protected resource",
				"user_id": user.ID,
			})
		})
	}

	// Optional auth route
	optional := r.Group("/optional")
	optional.Use(OptionalAuthMiddleware(pool))
	{
		optional.GET("/resource", func(c *gin.Context) {
			user := GetCurrentUser(c)
			if user != nil {
				c.JSON(http.StatusOK, gin.H{"user_id": user.ID})
			} else {
				c.JSON(http.StatusOK, gin.H{"user_id": "anonymous"})
			}
		})
	}

	return r
}

// createTestSession creates a test session in the database and returns the token
func createTestSession(t *testing.T, pool *pgxpool.Pool) (string, string) {
	ctx := context.Background()

	// Create test user
	userID := "test-user-" + base64.URLEncoding.EncodeToString([]byte(time.Now().String()))[:16]
	_, err := pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, "emailVerified", "createdAt", "updatedAt")
		VALUES ($1, 'Test User', 'test@example.com', true, NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, userID)
	require.NoError(t, err)

	// Create session
	sessionToken := base64.URLEncoding.EncodeToString([]byte("test-session-" + userID))
	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = pool.Exec(ctx, `
		INSERT INTO session (id, token, "userId", "expiresAt", "ipAddress", "userAgent")
		VALUES ($1, $2, $3, $4, '127.0.0.1', 'test-agent')
	`, "session-"+userID, sessionToken, userID, expiresAt)
	require.NoError(t, err)

	return sessionToken, userID
}

// cleanupTestSession removes test data
func cleanupTestSession(t *testing.T, pool *pgxpool.Pool, userID string) {
	ctx := context.Background()
	_, _ = pool.Exec(ctx, `DELETE FROM session WHERE "userId" = $1`, userID)
	_, _ = pool.Exec(ctx, `DELETE FROM "user" WHERE id = $1`, userID)
}

func TestRequireAuth_WithValidAuth(t *testing.T) {
	// Skip if no database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	defer pool.Close()

	sessionToken, userID := createTestSession(t, pool)
	defer cleanupTestSession(t, pool, userID)

	router := setupTestRouter(pool)

	// Make request to protected route WITH valid session cookie
	req := httptest.NewRequest(http.MethodGet, "/protected/resource", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: sessionToken,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "protected resource")
	assert.Contains(t, w.Body.String(), userID)
}

func TestRequireAuth_WithoutAuth(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	defer pool.Close()

	router := setupTestRouter(pool)

	// Make request to protected route WITHOUT session cookie
	req := httptest.NewRequest(http.MethodGet, "/protected/resource", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Not authenticated")
}

func TestRequireAuth_WithInvalidToken(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	defer pool.Close()

	router := setupTestRouter(pool)

	// Make request with INVALID session cookie
	req := httptest.NewRequest(http.MethodGet, "/protected/resource", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: "invalid-token-12345",
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired session")
}

func TestRequireAuth_WithExpiredSession(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	defer pool.Close()

	ctx := context.Background()

	// Create user
	userID := "expired-user-" + base64.URLEncoding.EncodeToString([]byte(time.Now().String()))[:16]
	_, err = pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, "emailVerified", "createdAt", "updatedAt")
		VALUES ($1, 'Expired User', 'expired@example.com', true, NOW(), NOW())
	`, userID)
	require.NoError(t, err)
	defer cleanupTestSession(t, pool, userID)

	// Create EXPIRED session
	sessionToken := base64.URLEncoding.EncodeToString([]byte("expired-session-" + userID))
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
	_, err = pool.Exec(ctx, `
		INSERT INTO session (id, token, "userId", "expiresAt", "ipAddress", "userAgent")
		VALUES ($1, $2, $3, $4, '127.0.0.1', 'test-agent')
	`, "expired-session-"+userID, sessionToken, userID, expiresAt)
	require.NoError(t, err)

	router := setupTestRouter(pool)

	// Make request with expired session
	req := httptest.NewRequest(http.MethodGet, "/protected/resource", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: sessionToken,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired session")
}

func TestGetCurrentUser_WithUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	testUser := &BetterAuthUser{
		ID:    "test-123",
		Name:  "Test User",
		Email: "test@example.com",
	}
	c.Set(UserContextKey, testUser)

	user := GetCurrentUser(c)
	assert.NotNil(t, user)
	assert.Equal(t, "test-123", user.ID)
	assert.Equal(t, "Test User", user.Name)
}

func TestGetCurrentUser_WithoutUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	user := GetCurrentUser(c)
	assert.Nil(t, user)
}

func TestOptionalAuthMiddleware_WithAuth(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	defer pool.Close()

	sessionToken, userID := createTestSession(t, pool)
	defer cleanupTestSession(t, pool, userID)

	router := setupTestRouter(pool)

	// Make request to optional auth route WITH valid session
	req := httptest.NewRequest(http.MethodGet, "/optional/resource", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: sessionToken,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), userID)
}

func TestOptionalAuthMiddleware_WithoutAuth(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	defer pool.Close()

	router := setupTestRouter(pool)

	// Make request to optional auth route WITHOUT session
	req := httptest.NewRequest(http.MethodGet, "/optional/resource", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "anonymous")
}

func TestAuthMiddleware_DevBypass(t *testing.T) {
	// Set dev bypass mode
	originalValue := os.Getenv("DEV_AUTH_BYPASS")
	os.Setenv("DEV_AUTH_BYPASS", "true")
	defer os.Setenv("DEV_AUTH_BYPASS", originalValue)

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	require.NoError(t, err)
	defer pool.Close()

	router := setupTestRouter(pool)

	// Make request WITHOUT session cookie (dev bypass should allow it)
	req := httptest.NewRequest(http.MethodGet, "/protected/resource", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "protected resource")
}

func TestRequireAuth_AbortsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware chain: RequireAuth should abort before handler runs
	handlerCalled := false
	router.GET("/test", RequireAuth(), func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, handlerCalled, "Handler should not have been called after RequireAuth abort")
}
