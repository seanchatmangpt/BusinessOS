package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// GoogleOAuthHandler handles Google OAuth endpoints
type GoogleOAuthHandler struct {
	calendarService *services.GoogleCalendarService
}

// NewGoogleOAuthHandler creates a new Google OAuth handler
func NewGoogleOAuthHandler(calendarService *services.GoogleCalendarService) *GoogleOAuthHandler {
	return &GoogleOAuthHandler{
		calendarService: calendarService,
	}
}

// InitiateGoogleAuth initiates the Google OAuth flow
func (h *GoogleOAuthHandler) InitiateGoogleAuth(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Generate random state for CSRF protection
	state, err := generateSecureRandomState()
	if err != nil {
		log.Printf("CRITICAL: Failed to generate OAuth state: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security error"})
		return
	}

	// Determine if we're in production (use secure cookies)
	isProduction := os.Getenv("ENVIRONMENT") == "production"

	// Store state in session/cookie for verification
	// SECURITY: Secure=true in production to prevent MitM attacks
	c.SetCookie("oauth_state", state, 600, "/", "", isProduction, true)
	c.SetCookie("oauth_user", user.ID, 600, "/", "", isProduction, true)

	authURL := h.calendarService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// HandleGoogleCallback handles the OAuth callback from Google
func (h *GoogleOAuthHandler) HandleGoogleCallback(c *gin.Context) {
	// Verify state
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Get user ID from cookie
	userID, err := c.Cookie("oauth_user")
	if err != nil || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User session not found"})
		return
	}

	// Check for error from Google
	if errMsg := c.Query("error"); errMsg != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/settings?google_error="+errMsg)
		return
	}

	// Exchange code for tokens
	code := c.Query("code")
	token, err := h.calendarService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code"})
		return
	}

	// Get user's email from Google
	email, err := getGoogleUserEmail(token.AccessToken)
	if err != nil {
		email = ""
	}

	// Get scopes from token
	scopes := []string{}
	if token.Extra("scope") != nil {
		scopes = append(scopes, token.Extra("scope").(string))
	}

	// Save tokens to database
	if err := h.calendarService.SaveToken(c.Request.Context(), userID, token, email); err != nil {
		// If token already exists, update it
		if err := h.calendarService.UpdateToken(c.Request.Context(), userID, token); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
			return
		}
	}

	// Bridge to user_integrations table for the new integrations module
	if err := h.calendarService.SyncToUserIntegrations(c.Request.Context(), userID, email, scopes); err != nil {
		// Log but don't fail - the OAuth tokens are saved
		// This just means the new integrations UI won't see it until next sync
		_ = err
	}

	// Clear OAuth cookies (use same Secure flag as when setting)
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	c.SetCookie("oauth_state", "", -1, "/", "", isProduction, true)
	c.SetCookie("oauth_user", "", -1, "/", "", isProduction, true)

	// Redirect to integrations page with success (or settings for backwards compatibility)
	c.Redirect(http.StatusTemporaryRedirect, "/integrations?google_connected=true")
}

// GetGoogleConnectionStatus returns the Google connection status for a user
func (h *GoogleOAuthHandler) GetGoogleConnectionStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	status, err := h.calendarService.GetConnectionStatus(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"connected": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"connected":    true,
		"email":        status.GoogleEmail,
		"connected_at": status.CreatedAt,
	})
}

// DisconnectGoogle disconnects the user's Google account
func (h *GoogleOAuthHandler) DisconnectGoogle(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Delete OAuth tokens
	if err := h.calendarService.DeleteToken(c.Request.Context(), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect Google account"})
		return
	}

	// Also clean up user_integrations table
	_ = h.calendarService.DeleteUserIntegration(c.Request.Context(), user.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Google account disconnected"})
}

// Helper functions

// generateSecureRandomState generates a cryptographically secure random state
// SECURITY: Returns error if crypto/rand fails - never silently continue with weak randomness
func generateSecureRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand.Read failed: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Deprecated: Use generateSecureRandomState instead
func generateRandomState() string {
	state, err := generateSecureRandomState()
	if err != nil {
		// Log and panic - this is a critical security failure
		log.Fatalf("CRITICAL: Failed to generate secure random state: %v", err)
	}
	return state
}

// getGoogleUserEmail fetches the user's email from Google using the access token
// SECURITY: Uses Authorization header instead of URL parameter to prevent token leakage in logs
func getGoogleUserEmail(accessToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// SECURITY: Use Authorization header instead of URL parameter
	// This prevents token from appearing in server logs, proxy logs, browser history
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Google API returned status %d", resp.StatusCode)
	}

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return userInfo.Email, nil
}
