package handlers

import (
	"context"
	"log/slog"
	"net/http"

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

// CreateCustomAgent creates a new custom agent (requires real implementation)
func (h *AgentHandler) CreateCustomAgent(c *gin.Context) {
	// ARMSTRONG: Let-It-Crash principle — this feature needs a proper request model + validation
	// Return 501 Not Implemented instead of fake success
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":  "custom agent creation not implemented",
		"reason": "requires request validation (name, system_prompt, category) and idempotency key",
	})
}

// TestCustomAgent tests a custom agent (requires real LLM integration)
func (h *AgentHandler) TestCustomAgent(c *gin.Context) {
	// ARMSTRONG: Let-It-Crash principle — testing an agent requires LLM service integration
	// Not implemented until agent service can call LLM
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":  "custom agent testing not implemented",
		"reason": "requires live LLM service integration (anthropic, openai, etc.)",
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

// CreateAgentFromPreset creates a custom agent from a preset (requires real implementation)
func (h *AgentHandler) CreateAgentFromPreset(c *gin.Context) {
	// ARMSTRONG: Let-It-Crash principle — agent creation from preset needs proper service layer
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":  "create agent from preset not implemented",
		"reason": "requires preset service integration and custom_agents table insert with idempotency",
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

// UpdateCustomAgent updates a custom agent (requires real implementation)
func (h *AgentHandler) UpdateCustomAgent(c *gin.Context) {
	// ARMSTRONG: Let-It-Crash principle — updates need proper validation and service layer
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":  "update custom agent not implemented",
		"reason": "requires request validation and optimistic locking to prevent conflicts",
	})
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
