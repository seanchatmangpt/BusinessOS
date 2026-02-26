package handlers

// auth_setup.go — Handlers for first-boot setup and team invite management.
//
// Routes (registered conditionally based on auth mode):
//
//   GET  /api/auth/mode              — Returns active auth mode + setup status
//   POST /api/auth/setup             — Completes first-boot setup (creates admin)
//   POST /api/auth/invites           — Admin creates an invite link
//   GET  /api/auth/invites/:token    — Validates an invite token (used by register page)
//   POST /api/auth/logout-all        — Logout all sessions (already in auth_google.go)

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/auth"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// AuthSetupHandler exposes first-boot and invite endpoints.
type AuthSetupHandler struct {
	pool *pgxpool.Pool
	cfg  *config.Config
	mode auth.AuthMode
}

// NewAuthSetupHandler creates a new AuthSetupHandler.
func NewAuthSetupHandler(pool *pgxpool.Pool, cfg *config.Config, mode auth.AuthMode) *AuthSetupHandler {
	return &AuthSetupHandler{pool: pool, cfg: cfg, mode: mode}
}

// ─── /api/auth/mode ─────────────────────────────────────────────────────────

// GetAuthMode returns the active authentication mode and setup status.
// This is a public endpoint — no auth required.
func (h *AuthSetupHandler) GetAuthMode(c *gin.Context) {
	status, err := auth.CheckSetupStatus(c.Request.Context(), h.pool, h.mode)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "check setup status", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mode":         string(h.mode),
		"needs_setup":  status.NeedsSetup,
		"has_users":    status.HasUsers,
		"allows_local": h.mode.AllowsLocalAuth(),
		"allows_oauth": h.mode.AllowsOAuth(),
		"oauth_providers": gin.H{
			"google": h.cfg.GoogleClientID != "",
			"github": h.cfg.GitHubClientID != "",
		},
	})
}

// ─── /api/auth/setup ────────────────────────────────────────────────────────

// SetupRequest is the body for completing first-boot setup.
type SetupRequest struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// CompleteSetup creates the first admin user. Only callable when no users
// exist yet. After this call the app transitions to normal auth mode.
func (h *AuthSetupHandler) CompleteSetup(c *gin.Context) {
	if !h.mode.RequiresLogin() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setup is not needed in single-user mode"})
		return
	}

	// Guard: refuse if users already exist.
	status, err := auth.CheckSetupStatus(c.Request.Context(), h.pool, h.mode)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "check setup status", err)
		return
	}
	if !status.NeedsSetup {
		c.JSON(http.StatusConflict, gin.H{"error": "Setup already completed"})
		return
	}

	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	if msg := validatePasswordStrength(req.Password); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	hashedPw, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "hash password", err)
		return
	}

	ctx := c.Request.Context()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "begin tx", err)
		return
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	userID := generateUserID()
	_, err = tx.Exec(ctx, `
		INSERT INTO "user" (id, name, email, "emailVerified", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, false, NOW(), NOW())
	`, userID, req.Name, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create admin user", err)
		return
	}

	accountID := generateUserID()
	_, err = tx.Exec(ctx, `
		INSERT INTO account (id, "userId", "accountId", "providerId", password, "createdAt", "updatedAt")
		VALUES ($1, $2, $3, 'credential', $4, NOW(), NOW())
	`, accountID, userID, userID, string(hashedPw))
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create account", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		utils.RespondInternalError(c, slog.Default(), "commit", err)
		return
	}

	// Issue a session immediately so the admin lands on the dashboard.
	sessionHandler := &EmailAuthHandler{pool: h.pool, cfg: h.cfg, logger: slog.Default()}
	sessionToken, err := sessionHandler.createSession(ctx, userID)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create session", err)
		return
	}

	setSessionCookie(c, sessionToken)

	slog.Info("first-boot setup completed", "user_id", userID, "email", req.Email)
	c.JSON(http.StatusCreated, gin.H{
		"user":    gin.H{"id": userID, "email": req.Email, "name": req.Name},
		"message": "Setup complete. Welcome to BusinessOS.",
	})
}

// ─── /api/auth/invites ──────────────────────────────────────────────────────

// CreateInviteRequest describes a request to generate a new invite link.
type CreateInviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"  binding:"omitempty,oneof=admin member viewer"`
}

// CreateInvite generates a new team invite link. Admin only.
func (h *AuthSetupHandler) CreateInvite(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, slog.Default())
		return
	}

	var req CreateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondInvalidRequest(c, slog.Default(), err)
		return
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	rawToken, err := auth.CreateInvite(c.Request.Context(), h.pool, req.Email, role, user.ID)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create invite", err)
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}

	inviteURL := baseURL + "/register?invite=" + rawToken

	c.JSON(http.StatusCreated, gin.H{
		"invite_url": inviteURL,
		"token":      rawToken,
		"email":      req.Email,
		"role":       role,
		"expires_at": time.Now().Add(7 * 24 * time.Hour).UTC(),
	})
}

// ValidateInvite checks whether an invite token is still valid without
// consuming it. Used by the register page to pre-fill the email field.
func (h *AuthSetupHandler) ValidateInvite(c *gin.Context) {
	rawToken := c.Param("token")
	if rawToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing invite token"})
		return
	}

	record, err := auth.ValidateInvite(c.Request.Context(), h.pool, rawToken)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error(), "valid": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"email":      record.Email,
		"role":       record.Role,
		"expires_at": record.ExpiresAt.UTC(),
	})
}

// ─── helpers ────────────────────────────────────────────────────────────────

// setSessionCookie sets the Better Auth session cookie following the same
// conventions used by auth_email.go and auth_google.go.
func setSessionCookie(c *gin.Context, token string) {
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	var domain string
	if isProduction {
		domain = os.Getenv("COOKIE_DOMAIN")
	}

	sameSite := http.SameSiteStrictMode
	if !isProduction {
		sameSite = http.SameSiteLaxMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "better-auth.session_token",
		Value:    token,
		Path:     "/",
		Domain:   domain,
		MaxAge:   60 * 60 * 24 * 7,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})
}

// validatePasswordStrength mirrors validatePassword in auth_email.go so that
// setup uses the same rules as normal registration.
func validatePasswordStrength(password string) string {
	runeLen := utf8.RuneCountInString(password)
	if runeLen < 8 {
		return "Password must be at least 8 characters"
	}
	if runeLen > 128 {
		return "Password must not exceed 128 characters"
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	if !hasUpper {
		return "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return "Password must contain at least one lowercase letter"
	}
	if !hasDigit {
		return "Password must contain at least one digit"
	}
	if !hasSpecial {
		return "Password must contain at least one special character"
	}
	return ""
}
