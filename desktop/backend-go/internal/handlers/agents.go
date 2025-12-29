package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/middleware"
)

// ListCustomAgents returns all custom agents for the authenticated user
func (h *Handlers) ListCustomAgents(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	// Check if we want all agents or only active ones
	includeInactive := c.Query("include_inactive") == "true"

	var agents []sqlc.CustomAgent
	var err error

	if includeInactive {
		agents, err = queries.GetAllCustomAgents(ctx, user.ID)
	} else {
		agents, err = queries.ListCustomAgents(ctx, user.ID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list agents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// GetCustomAgent retrieves a specific custom agent
func (h *Handlers) GetCustomAgent(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	agent, err := queries.GetCustomAgent(ctx, sqlc.GetCustomAgentParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"agent": agent})
}

// CreateCustomAgentRequest represents request to create custom agent
type CreateCustomAgentRequest struct {
	Name             string   `json:"name" binding:"required"`
	DisplayName      string   `json:"display_name" binding:"required"`
	Description      string   `json:"description"`
	Avatar           string   `json:"avatar"`
	SystemPrompt     string   `json:"system_prompt" binding:"required"`
	ModelPreference  string   `json:"model_preference"`
	Temperature      float64  `json:"temperature"`
	MaxTokens        int32    `json:"max_tokens"`
	Capabilities     []string `json:"capabilities"`
	ToolsEnabled     []string `json:"tools_enabled"`
	ContextSources   []string `json:"context_sources"`
	ThinkingEnabled  bool     `json:"thinking_enabled"`
	StreamingEnabled bool     `json:"streaming_enabled"`
	Category         string   `json:"category"`
}

// CreateCustomAgent creates a new custom agent
func (h *Handlers) CreateCustomAgent(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var req CreateCustomAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate agent name (alphanumeric + hyphens only)
	name := strings.ToLower(strings.TrimSpace(req.Name))
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Agent name can only contain lowercase letters, numbers, and hyphens"})
			return
		}
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	// Prepare optional fields
	var desc *string
	if req.Description != "" {
		desc = &req.Description
	}
	var avatar *string
	if req.Avatar != "" {
		avatar = &req.Avatar
	}
	var modelPref *string
	if req.ModelPreference != "" {
		modelPref = &req.ModelPreference
	}
	var maxTokens *int32
	if req.MaxTokens > 0 {
		maxTokens = &req.MaxTokens
	}
	var category *string
	if req.Category != "" {
		category = &req.Category
	}

	// Convert temperature to pgtype.Numeric
	tempNumeric := pgtype.Numeric{}
	if req.Temperature > 0 {
		tempNumeric.Scan(req.Temperature)
	}

	thinkingEnabled := &req.ThinkingEnabled
	streamingEnabled := &req.StreamingEnabled
	isActive := boolPtr(true)

	agent, err := queries.CreateCustomAgent(ctx, sqlc.CreateCustomAgentParams{
		UserID:           user.ID,
		Name:             name,
		DisplayName:      req.DisplayName,
		Description:      desc,
		Avatar:           avatar,
		SystemPrompt:     req.SystemPrompt,
		ModelPreference:  modelPref,
		Temperature:      tempNumeric,
		MaxTokens:        maxTokens,
		Capabilities:     req.Capabilities,
		ToolsEnabled:     req.ToolsEnabled,
		ContextSources:   req.ContextSources,
		ThinkingEnabled:  thinkingEnabled,
		StreamingEnabled: streamingEnabled,
		Category:         category,
		IsActive:         isActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"agent": agent})
}

// UpdateCustomAgentRequest represents request to update custom agent
type UpdateCustomAgentRequest struct {
	Name             *string  `json:"name"`
	DisplayName      *string  `json:"display_name"`
	Description      *string  `json:"description"`
	Avatar           *string  `json:"avatar"`
	SystemPrompt     *string  `json:"system_prompt"`
	ModelPreference  *string  `json:"model_preference"`
	Temperature      *float64 `json:"temperature"`
	MaxTokens        *int32   `json:"max_tokens"`
	Capabilities     []string `json:"capabilities"`
	ToolsEnabled     []string `json:"tools_enabled"`
	ContextSources   []string `json:"context_sources"`
	ThinkingEnabled  *bool    `json:"thinking_enabled"`
	StreamingEnabled *bool    `json:"streaming_enabled"`
	Category         *string  `json:"category"`
	IsActive         *bool    `json:"is_active"`
}

// UpdateCustomAgent updates an existing custom agent
func (h *Handlers) UpdateCustomAgent(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	var req UpdateCustomAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	// Validate name if provided
	if req.Name != nil {
		name := strings.ToLower(strings.TrimSpace(*req.Name))
		for _, r := range name {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Agent name can only contain lowercase letters, numbers, and hyphens"})
				return
			}
		}
		req.Name = &name
	}

	// Convert temperature to pgtype.Numeric
	tempNumeric := pgtype.Numeric{}
	if req.Temperature != nil && *req.Temperature > 0 {
		tempNumeric.Scan(*req.Temperature)
	}

	agent, err := queries.UpdateCustomAgent(ctx, sqlc.UpdateCustomAgentParams{
		ID:               pgtype.UUID{Bytes: id, Valid: true},
		UserID:           user.ID,
		Name:             req.Name,
		DisplayName:      req.DisplayName,
		Description:      req.Description,
		Avatar:           req.Avatar,
		SystemPrompt:     req.SystemPrompt,
		ModelPreference:  req.ModelPreference,
		Temperature:      tempNumeric,
		MaxTokens:        req.MaxTokens,
		Capabilities:     req.Capabilities,
		ToolsEnabled:     req.ToolsEnabled,
		ContextSources:   req.ContextSources,
		ThinkingEnabled:  req.ThinkingEnabled,
		StreamingEnabled: req.StreamingEnabled,
		Category:         req.Category,
		IsActive:         req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"agent": agent})
}

// DeleteCustomAgent deletes a custom agent
func (h *Handlers) DeleteCustomAgent(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid agent ID"})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	err = queries.DeleteCustomAgent(ctx, sqlc.DeleteCustomAgentParams{
		ID:     pgtype.UUID{Bytes: id, Valid: true},
		UserID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListAgentPresets returns all available agent presets
func (h *Handlers) ListAgentPresets(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	presets, err := queries.ListAgentPresets(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list presets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"presets": presets})
}

// GetAgentPreset retrieves a specific agent preset
func (h *Handlers) GetAgentPreset(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preset ID"})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	preset, err := queries.GetAgentPreset(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preset not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preset": preset})
}

// CreateAgentFromPresetRequest represents request to create agent from preset
type CreateAgentFromPresetRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateAgentFromPreset creates a new custom agent based on a preset
func (h *Handlers) CreateAgentFromPreset(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	presetIDStr := c.Param("presetId")
	presetID, err := uuid.Parse(presetIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preset ID"})
		return
	}

	var req CreateAgentFromPresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate agent name
	name := strings.ToLower(strings.TrimSpace(req.Name))
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Agent name can only contain lowercase letters, numbers, and hyphens"})
			return
		}
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	// Create agent from preset
	agent, err := queries.CreateAgentFromPreset(ctx, sqlc.CreateAgentFromPresetParams{
		UserID:   user.ID,
		Name:     name,
		PresetID: pgtype.UUID{Bytes: presetID, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent from preset: " + err.Error()})
		return
	}

	// Increment preset copy count
	_ = queries.IncrementPresetCopyCount(ctx, pgtype.UUID{Bytes: presetID, Valid: true})

	c.JSON(http.StatusCreated, gin.H{"agent": agent})
}

// ListCustomAgentsByCategory returns custom agents filtered by category
func (h *Handlers) ListCustomAgentsByCategory(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category is required"})
		return
	}

	ctx := context.Background()
	queries := sqlc.New(h.pool)

	agents, err := queries.ListCustomAgentsByCategory(ctx, sqlc.ListCustomAgentsByCategoryParams{
		UserID:   user.ID,
		Category: &category,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list agents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"agents": agents})
}
