package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)


type SlackOAuthHandler struct {
	slackService *services.SlackService
}

// NewSlackOAuthHandler creates a new Slack OAuth handler
func NewSlackOAuthHandler(slackService *services.SlackService) *SlackOAuthHandler {
	return &SlackOAuthHandler{
		slackService: slackService,
	}
}

// InitiateSlackAuth initiates the Slack OAuth flow
func (h *SlackOAuthHandler) InitiateSlackAuth(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Generate random state for CSRF protection
	state, err := generateSlackSecureRandomState()
	if err != nil {
		log.Printf("CRITICAL: Failed to generate Slack OAuth state: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security error"})
		return
	}

	// Determine if we're in production (use secure cookies)
	isProduction := os.Getenv("ENVIRONMENT") == "production"

	// SECURITY: Secure=true in production to prevent MitM attacks
	c.SetCookie("slack_oauth_state", state, 600, "/", "", isProduction, true)
	c.SetCookie("slack_oauth_user", user.ID, 600, "/", "", isProduction, true)

	authURL := h.slackService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

func (h *SlackOAuthHandler) HandleSlackCallback(c *gin.Context) {
	// Verify state
	state := c.Query("state")
	storedState, err := c.Cookie("slack_oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	userID, err := c.Cookie("slack_oauth_user")
	if err != nil || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User session not found"})
		return
	}

	if errMsg := c.Query("error"); errMsg != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/settings?slack_error="+errMsg)
		return
	}

	// Exchange code for tokens
	code := c.Query("code")
	response, err := h.slackService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code: " + err.Error()})
		return
	}

	// Save tokens to database
	if err := h.slackService.SaveToken(c.Request.Context(), userID, response); err != nil {
		// If token already exists, update it
		if err := h.slackService.UpdateToken(c.Request.Context(), userID, response); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
			return
		}
	}

	// Bridge to user_integrations table for the new integrations module
	scopes := []string{}
	if response.Scope != "" {
		scopes = append(scopes, response.Scope)
	}
	_ = h.slackService.SyncToUserIntegrations(c.Request.Context(), userID, response.Team.Name, scopes)

	// Clear OAuth cookies (use same Secure flag as when setting)
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	c.SetCookie("slack_oauth_state", "", -1, "/", "", isProduction, true)
	c.SetCookie("slack_oauth_user", "", -1, "/", "", isProduction, true)

	// Redirect to integrations page with success
	c.Redirect(http.StatusTemporaryRedirect, "/integrations?slack_connected=true")
}

// GetSlackConnectionStatus returns the Slack connection status for a user
func (h *SlackOAuthHandler) GetSlackConnectionStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	status, err := h.slackService.GetConnectionStatus(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":      true,
		"workspace_id":   status.WorkspaceID,
		"workspace_name": status.WorkspaceName,
		"connected_at":   status.CreatedAt,
	})
}

// DisconnectSlack disconnects the user's Slack workspace
func (h *SlackOAuthHandler) DisconnectSlack(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Delete OAuth tokens
	if err := h.slackService.DeleteToken(c.Request.Context(), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect Slack workspace"})
		return
	}

	// Also clean up user_integrations table
	_ = h.slackService.DeleteUserIntegration(c.Request.Context(), user.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Slack workspace disconnected"})
}

// generateSlackSecureRandomState generates a cryptographically secure random state
// SECURITY: Returns error if crypto/rand fails - never silently continue with weak randomness
func generateSlackSecureRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand.Read failed: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Deprecated: Use generateSlackSecureRandomState instead
func generateSlackRandomState() string {
	state, err := generateSlackSecureRandomState()
	if err != nil {
		log.Fatalf("CRITICAL: Failed to generate secure random state: %v", err)
	}
	return state
}
