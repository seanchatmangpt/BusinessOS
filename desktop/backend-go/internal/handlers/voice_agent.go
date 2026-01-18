package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
)

// MinimalUserContext - simplified context for minimal voice agent
type MinimalUserContext struct {
	Name string `json:"name"`
}

// HandleVoiceUserContext returns minimal user context for voice agent
// GET /api/voice/user-context/:user_id
// Called by Python voice agent to get user's name for personalization
func (h *Handlers) HandleVoiceUserContext(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}

	slog.Info("[VoiceAgent] Fetching user context", "user_id", userID)

	// Default context
	response := MinimalUserContext{Name: "User"}

	// Try to get name from database
	queries := sqlc.New(h.pool)
	user, err := queries.GetUserByID(c.Request.Context(), userID)
	if err == nil && user.Name != nil && *user.Name != "" {
		response.Name = *user.Name
		slog.Info("[VoiceAgent] Found user", "name", response.Name)
	}

	c.JSON(http.StatusOK, response)
}
