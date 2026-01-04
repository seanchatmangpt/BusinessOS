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

type NotionOAuthHandler struct {
	notionService *services.NotionService
}

func NewNotionOAuthHandler(notionService *services.NotionService) *NotionOAuthHandler {
	return &NotionOAuthHandler{
		notionService: notionService,
	}
}

func (h *NotionOAuthHandler) InitiateNotionAuth(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	state, err := generateNotionSecureRandomState()
	if err != nil {
		log.Printf("CRITICAL: Failed to generate Notion OAuth state: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Security error"})
		return
	}

	// Determine if we're in production (use secure cookies)
	isProduction := os.Getenv("ENVIRONMENT") == "production"

	// SECURITY: Secure=true in production to prevent MitM attacks
	c.SetCookie("notion_oauth_state", state, 600, "/", "", isProduction, true)
	c.SetCookie("notion_oauth_user", user.ID, 600, "/", "", isProduction, true)

	authURL := h.notionService.GetAuthURL(state)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
	})
}

// HandleNotionCallback handles the OAuth callback from Notion
func (h *NotionOAuthHandler) HandleNotionCallback(c *gin.Context) {
	state := c.Query("state")
	storedState, err := c.Cookie("notion_oauth_state")
	if err != nil || state != storedState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	userID, err := c.Cookie("notion_oauth_user")
	if err != nil || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User session not found"})
		return
	}

	if errMsg := c.Query("error"); errMsg != "" {
		c.Redirect(http.StatusTemporaryRedirect, "/settings?notion_error="+errMsg)
		return
	}

	// Exchange code for tokens
	code := c.Query("code")
	response, err := h.notionService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		log.Printf("Notion OAuth exchange error: %v", err)
		c.Redirect(http.StatusTemporaryRedirect, "/settings?notion_error=exchange_failed")
		return
	}

	// Save tokens to database
	if err := h.notionService.SaveToken(c.Request.Context(), userID, response); err != nil {
		if err := h.notionService.UpdateToken(c.Request.Context(), userID, response); err != nil {
			log.Printf("Notion OAuth save token error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, "/settings?notion_error=save_failed")
			return
		}
	}

	// Bridge to user_integrations table for the new integrations module
	_ = h.notionService.SyncToUserIntegrations(c.Request.Context(), userID, response.WorkspaceName)

	// Clear OAuth cookies (use same Secure flag as when setting)
	isProduction := os.Getenv("ENVIRONMENT") == "production"
	c.SetCookie("notion_oauth_state", "", -1, "/", "", isProduction, true)
	c.SetCookie("notion_oauth_user", "", -1, "/", "", isProduction, true)

	// Redirect to integrations page with success
	c.Redirect(http.StatusTemporaryRedirect, "/integrations?notion_connected=true")
}

func (h *NotionOAuthHandler) GetNotionConnectionStatus(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	status, err := h.notionService.GetConnectionStatus(c.Request.Context(), user.ID)
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
		"workspace_icon": status.WorkspaceIcon,
		"owner_name":     status.OwnerUserName,
		"connected_at":   status.CreatedAt,
	})
}


func (h *NotionOAuthHandler) DisconnectNotion(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Delete OAuth tokens
	if err := h.notionService.DeleteToken(c.Request.Context(), user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect Notion workspace"})
		return
	}

	// Also clean up user_integrations table
	_ = h.notionService.DeleteUserIntegration(c.Request.Context(), user.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Notion workspace disconnected"})
}

// GetNotionDatabases returns all databases the user has access to
func (h *NotionOAuthHandler) GetNotionDatabases(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	cursor := c.Query("cursor")
	databases, nextCursor, hasMore, err := h.notionService.ListDatabases(c.Request.Context(), user.ID, 100, cursor)
	if err != nil {
		log.Printf("GetNotionDatabases error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch databases"})
		return
	}

	// Transform to simpler response format
	var result []map[string]interface{}
	for _, db := range databases {
		result = append(result, map[string]interface{}{
			"id":               db.ID,
			"title":            services.GetDatabaseTitle(&db),
			"url":              db.URL,
			"created_time":     db.CreatedTime,
			"last_edited_time": db.LastEditedTime,
			"icon":             db.Icon,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"databases":   result,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}

func (h *NotionOAuthHandler) GetNotionPages(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	databaseID := c.Query("database_id")
	if databaseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database_id is required"})
		return
	}

	cursor := c.Query("cursor")
	queryResp, err := h.notionService.QueryDatabase(c.Request.Context(), user.ID, databaseID, 100, cursor, nil, nil)
	if err != nil {
		log.Printf("GetNotionPages error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pages"})
		return
	}

	// Transform to simpler response format
	var result []map[string]interface{}
	for _, page := range queryResp.Results {
		result = append(result, map[string]interface{}{
			"id":               page.ID,
			"title":            services.GetPageTitle(&page),
			"url":              page.URL,
			"created_time":     page.CreatedTime,
			"last_edited_time": page.LastEditedTime,
			"icon":             page.Icon,
			"archived":         page.Archived,
		})
	}

	nextCursor := ""
	if queryResp.NextCursor != nil {
		nextCursor = *queryResp.NextCursor
	}

	c.JSON(http.StatusOK, gin.H{
		"pages":       result,
		"next_cursor": nextCursor,
		"has_more":    queryResp.HasMore,
	})
}

// SearchNotion searches for pages and databases
func (h *NotionOAuthHandler) SearchNotion(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	query := c.Query("query")
	filter := c.Query("filter") // "database" or "page"
	cursor := c.Query("cursor")

	searchResp, err := h.notionService.Search(c.Request.Context(), user.ID, query, filter, 20, cursor)
	if err != nil {
		log.Printf("SearchNotion error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}

	nextCursor := ""
	if searchResp.NextCursor != nil {
		nextCursor = *searchResp.NextCursor
	}

	c.JSON(http.StatusOK, gin.H{
		"results":     searchResp.Results,
		"next_cursor": nextCursor,
		"has_more":    searchResp.HasMore,
	})
}

// generateNotionSecureRandomState generates a cryptographically secure random state
// SECURITY: Returns error if crypto/rand fails - never silently continue with weak randomness
func generateNotionSecureRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand.Read failed: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Deprecated: Use generateNotionSecureRandomState instead
func generateNotionRandomState() string {
	state, err := generateNotionSecureRandomState()
	if err != nil {
		log.Fatalf("CRITICAL: Failed to generate secure random state: %v", err)
	}
	return state
}
