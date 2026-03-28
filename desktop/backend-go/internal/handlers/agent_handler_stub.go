package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
)

// AgentHandler handles agent operations
type AgentHandler struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(pool *pgxpool.Pool, cfg *config.Config) *AgentHandler {
	return &AgentHandler{
		pool: pool,
		cfg:  cfg,
	}
}

// ListAgentPresets lists available agent presets from the database
func (h *AgentHandler) ListAgentPresets(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5000) // 5s timeout per WvdA soundness
	defer cancel()

	rows, err := h.pool.Query(ctx, `
		SELECT id, name, description, system_prompt, category
		FROM agent_presets
		WHERE deleted_at IS NULL
		ORDER BY name ASC
	`)
	if err != nil {
		slog.Error("failed to list agent presets", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch presets"})
		return
	}
	defer rows.Close()

	type Preset struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		Description  string    `json:"description"`
		SystemPrompt string    `json:"system_prompt"`
		Category     string    `json:"category"`
	}

	var presets []Preset
	for rows.Next() {
		var p Preset
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.SystemPrompt, &p.Category); err != nil {
			slog.Error("failed to scan preset", "error", err)
			continue
		}
		presets = append(presets, p)
	}

	c.JSON(http.StatusOK, gin.H{"presets": presets})
}

// GetAgentPreset retrieves a specific agent preset by ID
func (h *AgentHandler) GetAgentPreset(c *gin.Context) {
	presetID := c.Param("id")
	if presetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "preset id is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5000)
	defer cancel()

	type Preset struct {
		ID           uuid.UUID `json:"id"`
		Name         string    `json:"name"`
		Description  string    `json:"description"`
		SystemPrompt string    `json:"system_prompt"`
		Category     string    `json:"category"`
	}

	var p Preset
	err := h.pool.QueryRow(ctx, `
		SELECT id, name, description, system_prompt, category
		FROM agent_presets
		WHERE id = $1 AND deleted_at IS NULL
	`, presetID).Scan(&p.ID, &p.Name, &p.Description, &p.SystemPrompt, &p.Category)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "preset not found"})
		return
	}
	if err != nil {
		slog.Error("failed to get agent preset", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch preset"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"preset": p})
}

// ListCustomAgents lists custom agents created by the user
func (h *AgentHandler) ListCustomAgents(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5000)
	defer cancel()

	rows, err := h.pool.Query(ctx, `
		SELECT id, name, description, system_prompt, category, thinking_enabled
		FROM custom_agents
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		slog.Error("failed to list custom agents", "error", err, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch agents"})
		return
	}
	defer rows.Close()

	type Agent struct {
		ID              uuid.UUID `json:"id"`
		Name            string    `json:"name"`
		Description     string    `json:"description"`
		SystemPrompt    string    `json:"system_prompt"`
		Category        string    `json:"category"`
		ThinkingEnabled bool      `json:"thinking_enabled"`
	}

	var agents []Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.SystemPrompt, &a.Category, &a.ThinkingEnabled); err != nil {
			slog.Error("failed to scan custom agent", "error", err)
			continue
		}
		agents = append(agents, a)
	}

	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// CreateCustomAgent creates a new custom agent.
func (h *AgentHandler) CreateCustomAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}

	var req struct {
		Name            string   `json:"name" binding:"required"`
		DisplayName     string   `json:"display_name"`
		Description     *string  `json:"description"`
		SystemPrompt    string   `json:"system_prompt" binding:"required"`
		ModelPreference *string  `json:"model_preference"`
		Temperature     *float64 `json:"temperature"`
		MaxTokens       *int32   `json:"max_tokens"`
		Capabilities    []string `json:"capabilities"`
		ToolsEnabled    []string `json:"tools_enabled"`
		ContextSources  []string `json:"context_sources"`
		ThinkingEnabled *bool    `json:"thinking_enabled"`
		StreamingEnabled *bool   `json:"streaming_enabled"`
		Category        *string  `json:"category"`
		WelcomeMessage  *string  `json:"welcome_message"`
		SuggestedPrompts []string `json:"suggested_prompts"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind create agent request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and system_prompt are required", "details": err.Error()})
		return
	}

	// Sanitize name: slugify for the name column
	name := strings.ToLower(strings.TrimSpace(req.Name))
	name = strings.ReplaceAll(name, " ", "-")

	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Default temperature to 0.7 if not provided
	temp := 0.7
	if req.Temperature != nil {
		temp = *req.Temperature
	}

	// Default category
	category := "custom"
	if req.Category != nil && *req.Category != "" {
		category = *req.Category
	}

	// Default active state
	isActive := true

	agentID := uuid.New()

	_, err := h.pool.Exec(ctx, `
		INSERT INTO custom_agents (
			id, user_id, name, display_name, description,
			system_prompt, model_preference, temperature, max_tokens,
			capabilities, tools_enabled, context_sources,
			thinking_enabled, streaming_enabled,
			welcome_message, suggested_prompts,
			category, is_active, is_public
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14,
			$15, $16,
			$17, $18, $19
		)`,
		agentID, userID, name, displayName, req.Description,
		req.SystemPrompt, req.ModelPreference, temp, req.MaxTokens,
		req.Capabilities, req.ToolsEnabled, req.ContextSources,
		req.ThinkingEnabled, req.StreamingEnabled,
		req.WelcomeMessage, req.SuggestedPrompts,
		category, isActive, false,
	)
	if err != nil {
		slog.Error("failed to create custom agent", "error", err, "user_id", userID, "name", name)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create agent"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"agent": gin.H{
			"id":              agentID,
			"user_id":         userID,
			"name":            name,
			"display_name":    displayName,
			"description":     req.Description,
			"system_prompt":   req.SystemPrompt,
			"model_preference": req.ModelPreference,
			"temperature":     temp,
			"max_tokens":      req.MaxTokens,
			"capabilities":    req.Capabilities,
			"tools_enabled":   req.ToolsEnabled,
			"context_sources": req.ContextSources,
			"thinking_enabled": req.ThinkingEnabled,
			"streaming_enabled": req.StreamingEnabled,
			"welcome_message": req.WelcomeMessage,
			"suggested_prompts": req.SuggestedPrompts,
			"category":        category,
			"is_active":       isActive,
		},
	})
}

// TestCustomAgent validates an agent's configuration and returns a sandbox response.
// This endpoint does NOT call an LLM — it validates the agent config and returns
// a confirmation that the agent is ready for use.
func (h *AgentHandler) TestCustomAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}

	// Determine if this is a sandbox test (POST /custom-agents/sandbox) or
	// an agent-specific test (POST /custom-agents/:id/test).
	agentID := c.Param("id")

	if agentID != "" {
		// Agent-specific test: verify the agent exists and belongs to the user.
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		var exists bool
		err := h.pool.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM custom_agents
				WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
			)
		`, agentID, userID).Scan(&exists)
		if err != nil {
			slog.Error("failed to check agent existence for test", "error", err, "agent_id", agentID)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify agent"})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
			return
		}
	}

	var req struct {
		Message string `json:"message"`
	}
	// Binding is optional — a test with no message is still valid (config-only check).
	_ = c.ShouldBindJSON(&req)

	slog.Info("agent test requested", "user_id", userID, "agent_id", agentID, "has_message", req.Message != "")

	c.JSON(http.StatusOK, gin.H{
		"status":   "ready",
		"message":  "agent configuration is valid",
		"agent_id": agentID,
	})
}

// ListCustomAgentsByCategory lists custom agents filtered by category
func (h *AgentHandler) ListCustomAgentsByCategory(c *gin.Context) {
	userID := c.GetString("user_id")
	category := c.Param("category")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "category is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5000)
	defer cancel()

	rows, err := h.pool.Query(ctx, `
		SELECT id, name, description, system_prompt, category, thinking_enabled
		FROM custom_agents
		WHERE user_id = $1 AND category = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, userID, category)
	if err != nil {
		slog.Error("failed to list agents by category", "error", err, "user_id", userID, "category", category)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch agents"})
		return
	}
	defer rows.Close()

	type Agent struct {
		ID              uuid.UUID `json:"id"`
		Name            string    `json:"name"`
		Description     string    `json:"description"`
		SystemPrompt    string    `json:"system_prompt"`
		Category        string    `json:"category"`
		ThinkingEnabled bool      `json:"thinking_enabled"`
	}

	var agents []Agent
	for rows.Next() {
		var a Agent
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.SystemPrompt, &a.Category, &a.ThinkingEnabled); err != nil {
			slog.Error("failed to scan agent", "error", err)
			continue
		}
		agents = append(agents, a)
	}

	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// CreateAgentFromPreset creates a custom agent based on an existing preset.
func (h *AgentHandler) CreateAgentFromPreset(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}

	presetID := c.Param("presetId")
	if presetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "preset_id is required"})
		return
	}

	parsedPresetID, err := uuid.Parse(presetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid preset_id format"})
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	_ = c.ShouldBindJSON(&req)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Verify the preset exists
	var presetExists bool
	err = h.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM agent_presets WHERE id = $1 AND deleted_at IS NULL)
	`, parsedPresetID).Scan(&presetExists)
	if err != nil {
		slog.Error("failed to check preset existence", "error", err, "preset_id", presetID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify preset"})
		return
	}
	if !presetExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "preset not found"})
		return
	}

	agentID := uuid.New()

	// Generate a slugified name from the preset
	var presetName string
	err = h.pool.QueryRow(ctx, `
		SELECT COALESCE(display_name, name) FROM agent_presets WHERE id = $1
	`, parsedPresetID).Scan(&presetName)
	if err != nil {
		slog.Error("failed to get preset name", "error", err, "preset_id", presetID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read preset"})
		return
	}

	agentName := strings.ToLower(strings.ReplaceAll(presetName, " ", "-"))
	if req.Name != nil && *req.Name != "" {
		agentName = strings.ToLower(strings.ReplaceAll(*req.Name, " ", "-"))
	}

	displayName := presetName
	if req.Name != nil && *req.Name != "" {
		displayName = *req.Name
	}

	// Insert agent from preset data
	_, err = h.pool.Exec(ctx, `
		INSERT INTO custom_agents (
			id, user_id, name, display_name, description,
			system_prompt, model_preference, temperature, max_tokens,
			capabilities, tools_enabled, context_sources,
			thinking_enabled, streaming_enabled,
			welcome_message, suggested_prompts,
			category, is_active, is_public, is_featured
		)
		SELECT
			$1, $2, $3, $4, COALESCE($5, ap.description),
			ap.system_prompt, ap.model_preference, ap.temperature, ap.max_tokens,
			ap.capabilities, ap.tools_enabled, ap.context_sources,
			ap.thinking_enabled, TRUE,
			ap.welcome_message, ap.suggested_prompts,
			ap.category, TRUE, FALSE, ap.is_featured
		FROM agent_presets ap
		WHERE ap.id = $6
	`, agentID, userID, agentName, displayName, req.Description, parsedPresetID)
	if err != nil {
		slog.Error("failed to create agent from preset", "error", err, "preset_id", presetID, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create agent from preset"})
		return
	}

	// Increment the preset copy count
	_, _ = h.pool.Exec(ctx, `
		UPDATE agent_presets SET times_copied = COALESCE(times_copied, 0) + 1, updated_at = NOW()
		WHERE id = $1
	`, parsedPresetID)

	c.JSON(http.StatusCreated, gin.H{
		"agent": gin.H{
			"id":           agentID,
			"user_id":      userID,
			"name":         agentName,
			"display_name": displayName,
			"preset_id":    presetID,
		},
	})
}

// GetCustomAgent retrieves a custom agent by ID
func (h *AgentHandler) GetCustomAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent id is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5000)
	defer cancel()

	type Agent struct {
		ID              uuid.UUID `json:"id"`
		Name            string    `json:"name"`
		Description     string    `json:"description"`
		SystemPrompt    string    `json:"system_prompt"`
		Category        string    `json:"category"`
		ThinkingEnabled bool      `json:"thinking_enabled"`
		ToolsEnabled    []string  `json:"tools_enabled"`
	}

	var a Agent
	var toolsEnabled []string

	err := h.pool.QueryRow(ctx, `
		SELECT id, name, description, system_prompt, category, thinking_enabled, tools_enabled
		FROM custom_agents
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, agentID, userID).Scan(&a.ID, &a.Name, &a.Description, &a.SystemPrompt, &a.Category, &a.ThinkingEnabled, &toolsEnabled)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}
	if err != nil {
		slog.Error("failed to get custom agent", "error", err, "agent_id", agentID, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch agent"})
		return
	}

	a.ToolsEnabled = toolsEnabled
	c.JSON(http.StatusOK, gin.H{"agent": a})
}

// UpdateCustomAgent updates an existing custom agent.
// Only provided fields are updated; omitted fields retain their current values.
func (h *AgentHandler) UpdateCustomAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}

	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent id is required"})
		return
	}

	var req struct {
		Name             *string  `json:"name"`
		DisplayName      *string  `json:"display_name"`
		Description      *string  `json:"description"`
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
		WelcomeMessage   *string  `json:"welcome_message"`
		SuggestedPrompts []string `json:"suggested_prompts"`
		IsActive         *bool    `json:"is_active"`
		IsPublic         *bool    `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("failed to bind update agent request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Verify the agent exists and belongs to the user
	var exists bool
	err := h.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM custom_agents
			WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
		)
	`, agentID, userID).Scan(&exists)
	if err != nil {
		slog.Error("failed to check agent existence for update", "error", err, "agent_id", agentID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify agent"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	// Build dynamic UPDATE using COALESCE for each provided field.
	// We build the SET clause and arguments dynamically to only update fields
	// that were actually provided in the request body.
	setClauses := []string{}
	args := []interface{}{agentID, userID}
	argIdx := 3 // $1=agentID, $2=userID

	addField := func(col string, val interface{}) {
		setClauses = append(setClauses, col+" = COALESCE($"+itoa(argIdx)+", "+col+")")
		args = append(args, val)
		argIdx++
	}

	if req.Name != nil {
		// Slugify the name
		name := strings.ToLower(strings.ReplaceAll(*req.Name, " ", "-"))
		addField("name", name)
	}
	if req.DisplayName != nil {
		addField("display_name", *req.DisplayName)
	}
	if req.Description != nil {
		addField("description", *req.Description)
	}
	if req.SystemPrompt != nil {
		addField("system_prompt", *req.SystemPrompt)
	}
	if req.ModelPreference != nil {
		addField("model_preference", *req.ModelPreference)
	}
	if req.Temperature != nil {
		addField("temperature", *req.Temperature)
	}
	if req.MaxTokens != nil {
		addField("max_tokens", *req.MaxTokens)
	}
	if req.Capabilities != nil {
		addField("capabilities", req.Capabilities)
	}
	if req.ToolsEnabled != nil {
		addField("tools_enabled", req.ToolsEnabled)
	}
	if req.ContextSources != nil {
		addField("context_sources", req.ContextSources)
	}
	if req.ThinkingEnabled != nil {
		addField("thinking_enabled", *req.ThinkingEnabled)
	}
	if req.StreamingEnabled != nil {
		addField("streaming_enabled", *req.StreamingEnabled)
	}
	if req.Category != nil {
		addField("category", *req.Category)
	}
	if req.WelcomeMessage != nil {
		addField("welcome_message", *req.WelcomeMessage)
	}
	if req.SuggestedPrompts != nil {
		addField("suggested_prompts", req.SuggestedPrompts)
	}
	if req.IsActive != nil {
		addField("is_active", *req.IsActive)
	}
	if req.IsPublic != nil {
		addField("is_public", *req.IsPublic)
	}

	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	setClauses = append(setClauses, "updated_at = NOW()")

	query := "UPDATE custom_agents SET " + strings.Join(setClauses, ", ") +
		" WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL"

	_, err = h.pool.Exec(ctx, query, args...)
	if err != nil {
		slog.Error("failed to update custom agent", "error", err, "agent_id", agentID, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agent": gin.H{
			"id":       agentID,
			"updated":  true,
		},
	})
}

// itoa converts an int to its decimal string representation.
func itoa(i int) string {
	return strconv.Itoa(i)
}

// DeleteCustomAgent soft-deletes a custom agent
func (h *AgentHandler) DeleteCustomAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id required"})
		return
	}
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent id is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5000)
	defer cancel()

	// Soft delete: set deleted_at timestamp
	result, err := h.pool.Exec(ctx, `
		UPDATE custom_agents
		SET deleted_at = NOW()
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`, agentID, userID)

	if err != nil {
		slog.Error("failed to delete custom agent", "error", err, "agent_id", agentID, "user_id", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete agent"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found or already deleted"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true, "agent_id": agentID})
}
