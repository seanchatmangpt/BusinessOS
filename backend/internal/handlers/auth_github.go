package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GitHubAuthHandler handles GitHub OAuth for authentication.
type GitHubAuthHandler struct {
	pool         *pgxpool.Pool
	cfg          *config.Config
	oauthConfig  *oauth2.Config
	sessionCache *middleware.SessionCache
}

// GitHubUserInfo represents the user info from the GitHub API.
type GitHubUserInfo struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

// NewGitHubAuthHandler creates a new GitHub OAuth handler.
// Returns nil if GitHub OAuth credentials are not configured, which causes
// the routes to return 404 — this is intentional for unconfigured installs.
func NewGitHubAuthHandler(pool *pgxpool.Pool, cfg *config.Config, sessionCache *middleware.SessionCache) *GitHubAuthHandler {
	if cfg.GitHubClientID == "" || cfg.GitHubClientSecret == "" {
		return nil
	}

	redirectURI := cfg.GitHubRedirectURI
	if redirectURI == "" {
		redirectURI = "http://localhost:8001/api/v1/auth/github/callback"
	}

	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
	}

	return &GitHubAuthHandler{
		pool:         pool,
		cfg:          cfg,
		oauthConfig:  oauthConfig,
		sessionCache: sessionCache,
	}
}

// InitiateGitHubLogin starts the GitHub OAuth flow.
func (h *GitHubAuthHandler) InitiateGitHubLogin(c *gin.Context) {
	state := generateRandomState()

	redirectAfter := c.Query("redirect")
	if !isValidRedirectURL(redirectAfter) {
		redirectAfter = "/dashboard"
	}

	isProduction := os.Getenv("ENVIRONMENT") == "production"
	c.SetCookie("oauth_state", state, 600, "/", "", isProduction, true)
	c.SetCookie("oauth_redirect", redirectAfter, 600, "/", "", isProduction, true)

	authURL := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// HandleGitHubCallback handles the OAuth callback from GitHub.
func (h *GitHubAuthHandler) HandleGitHubCallback(c *gin.Context) {
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		utils.RespondBadRequest(c, slog.Default(), "Invalid state parameter")
		return
	}

	redirectAfter, _ := c.Cookie("oauth_redirect")
	if !isValidRedirectURL(redirectAfter) {
		redirectAfter = "/dashboard"
	}

	if errMsg := c.Query("error"); errMsg != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/?error="+errMsg)
		return
	}

	code := c.Query("code")
	token, err := h.oauthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "exchange OAuth code", err)
		return
	}

	userInfo, err := h.getGitHubUserInfo(token.AccessToken)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "get GitHub user info", err)
		return
	}

	if userInfo.Email == "" {
		// GitHub may not expose email publicly; fetch it from the emails endpoint.
		email, emailErr := h.getGitHubPrimaryEmail(token.AccessToken)
		if emailErr != nil {
			utils.RespondInternalError(c, slog.Default(), "get GitHub email", emailErr)
			return
		}
		userInfo.Email = email
	}

	if userInfo.Email == "" {
		utils.RespondBadRequest(c, slog.Default(), "GitHub account has no accessible email address")
		return
	}

	userID, err := h.upsertGitHubUser(c.Request.Context(), userInfo)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "upsert GitHub user", err)
		return
	}

	sessionToken, err := h.createGitHubSession(c.Request.Context(), userID)
	if err != nil {
		utils.RespondInternalError(c, slog.Default(), "create session", err)
		return
	}

	isProduction := os.Getenv("ENVIRONMENT") == "production"
	c.SetCookie("oauth_state", "", -1, "/", "", isProduction, true)
	c.SetCookie("oauth_redirect", "", -1, "/", "", isProduction, true)

	domain := os.Getenv("COOKIE_DOMAIN")

	sameSite := http.SameSiteStrictMode
	if !isProduction {
		sameSite = http.SameSiteLaxMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "better-auth.session_token",
		Value:    sessionToken,
		Path:     "/",
		Domain:   domain,
		MaxAge:   60 * 60 * 24 * 7, // 7 days
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: sameSite,
	})

	c.Redirect(http.StatusTemporaryRedirect, redirectAfter)
}

// getGitHubUserInfo fetches the authenticated user's public profile.
func (h *GitHubAuthHandler) getGitHubUserInfo(accessToken string) (*GitHubUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github user API returned %d", resp.StatusCode)
	}

	var info GitHubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// getGitHubPrimaryEmail fetches the primary verified email from GitHub's
// email endpoint (needed when the public profile email is empty).
func (h *GitHubAuthHandler) getGitHubPrimaryEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	// Fallback: first verified email.
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", nil
}

// upsertGitHubUser creates or updates the user record for a GitHub login.
func (h *GitHubAuthHandler) upsertGitHubUser(ctx context.Context, info *GitHubUserInfo) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	displayName := info.Name
	if strings.TrimSpace(displayName) == "" {
		displayName = info.Login
	}

	// Try to find existing user by email.
	var existingID string
	err := h.pool.QueryRow(ctx,
		`SELECT id FROM "user" WHERE email = $1`, info.Email,
	).Scan(&existingID)

	if err == nil {
		// User exists — update avatar and name.
		_, err = h.pool.Exec(ctx, `
			UPDATE "user"
			SET name = $1, image = $2, "emailVerified" = true, "updatedAt" = NOW()
			WHERE id = $3
		`, displayName, info.AvatarURL, existingID)
		if err != nil {
			return "", fmt.Errorf("update user: %w", err)
		}
		return existingID, nil
	}

	// New user.
	userID := generateUserID()
	_, err = h.pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, "emailVerified", image, "createdAt", "updatedAt")
		VALUES ($1, $2, $3, true, $4, NOW(), NOW())
	`, userID, displayName, info.Email, info.AvatarURL)
	if err != nil {
		return "", fmt.Errorf("insert user: %w", err)
	}

	// Create the OAuth account record.
	accountID := generateUserID()
	_, err = h.pool.Exec(ctx, `
		INSERT INTO account (id, "userId", "accountId", "providerId", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, 'github', NOW(), NOW())
	`, accountID, userID, fmt.Sprintf("github-%d", info.ID))
	if err != nil {
		return "", fmt.Errorf("insert account: %w", err)
	}

	return userID, nil
}

// createGitHubSession creates a 7-day session for the given user.
func (h *GitHubAuthHandler) createGitHubSession(ctx context.Context, userID string) (string, error) {
	sessionToken := generateSessionToken()
	sessionID := generateSessionID()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	_, err := h.pool.Exec(ctx, `
		INSERT INTO session (id, "userId", token, "expiresAt", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, sessionID, userID, sessionToken, expiresAt)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return sessionToken, nil
}
