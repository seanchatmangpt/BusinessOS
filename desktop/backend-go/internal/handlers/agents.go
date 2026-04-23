package handlers

import (
<<<<<<< HEAD
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
)

const defaultTimeout = 5 * time.Second

// AgentsHandler handles agent discovery endpoints.
type AgentsHandler struct {
	sparqlClient *ontology.SPARQLClient
	logger       *slog.Logger
}

// NewAgentsHandler creates a new AgentsHandler.
func NewAgentsHandler(sparqlClient *ontology.SPARQLClient, logger *slog.Logger) *AgentsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentsHandler{
		sparqlClient: sparqlClient,
		logger:       logger,
	}
}

// AgentListResponse represents a list of active agents.
type AgentListResponse struct {
	Agents []Agent `json:"agents"`
	Count  int     `json:"count"`
}

// Agent represents an active agent in the system.
type Agent struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Type          string   `json:"type"` // osa, businessos, canopy, pm4py-mcp
	Status        string   `json:"status"`
	LastHeartbeat string   `json:"last_heartbeat"`
	Capabilities  []string `json:"capabilities"`
}

// ListAgents returns all active agents from the ontology.
// GET /api/ontology/agents
func (h *AgentsHandler) ListAgents(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultTimeout)
	defer cancel()

	// Query for active agents in the ontology
	query := `
PREFIX schema: <http://schema.org/>
PREFIX bo: <http://businessos.example/ontology/>

SELECT ?agentId ?agentName ?agentType ?status ?lastHeartbeat ?capability
WHERE {
  ?agent a bo:Agent ;
         bo:agentId ?agentId ;
         bo:name ?agentName ;
         bo:type ?agentType ;
         bo:status ?status ;
         bo:lastHeartbeat ?lastHeartbeat .
  OPTIONAL { ?agent bo:capability ?capability }
}
ORDER BY ?agentId
`

	// Execute SPARQL SELECT query
	result, err := h.sparqlClient.ExecuteSelect(ctx, query, defaultTimeout)
	if err != nil {
		h.logger.Error("failed to query agents", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query agents"})
		return
	}

	// Parse results and build response
	agents := parseAgentResults(result)

	c.JSON(http.StatusOK, AgentListResponse{
		Agents: agents,
		Count:  len(agents),
	})
}

// parseAgentResults parses SPARQL SELECT results into Agent objects.
func parseAgentResults(data []byte) []Agent {
	// Lightweight parsing — full implementation would use a JSON SPARQL parser
	agents := make([]Agent, 0)

	// For now, return empty list
	// In production, parse JSON response from SPARQL endpoint
	return agents
}

// RegisterAgentsRoutes wires up agent discovery routes.
func RegisterAgentsRoutes(api *gin.RouterGroup, h *AgentsHandler, auth gin.HandlerFunc) {
	agents := api.Group("/ontology/agents")
	agents.Use(auth)
	{
		agents.GET("", h.ListAgents)
	}
=======
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
)

// AgentHandler handles custom AI agent CRUD operations
type AgentHandler struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

// NewAgentHandler creates a new AgentHandler
func NewAgentHandler(pool *pgxpool.Pool, cfg *config.Config) *AgentHandler {
	return &AgentHandler{pool: pool, cfg: cfg}
}

// CreateCustomAgentRequest represents request to create custom agent
type CreateCustomAgentRequest struct {
	Name                 string   `json:"name" binding:"required"`
	DisplayName          string   `json:"display_name" binding:"required"`
	Description          string   `json:"description"`
	Avatar               string   `json:"avatar"`
	SystemPrompt         string   `json:"system_prompt" binding:"required"`
	ModelPreference      string   `json:"model_preference"`
	Temperature          float64  `json:"temperature"`
	MaxTokens            int32    `json:"max_tokens"`
	Capabilities         []string `json:"capabilities"`
	ToolsEnabled         []string `json:"tools_enabled"`
	ContextSources       []string `json:"context_sources"`
	ThinkingEnabled      bool     `json:"thinking_enabled"`
	StreamingEnabled     bool     `json:"streaming_enabled"`
	ApplyPersonalization bool     `json:"apply_personalization"`
	WelcomeMessage       string   `json:"welcome_message"`
	SuggestedPrompts     []string `json:"suggested_prompts"`
	Category             string   `json:"category"`
	IsPublic             bool     `json:"is_public"`
	IsFeatured           bool     `json:"is_featured"`
}

// UpdateCustomAgentRequest represents request to update custom agent
type UpdateCustomAgentRequest struct {
	Name                 *string  `json:"name"`
	DisplayName          *string  `json:"display_name"`
	Description          *string  `json:"description"`
	Avatar               *string  `json:"avatar"`
	SystemPrompt         *string  `json:"system_prompt"`
	ModelPreference      *string  `json:"model_preference"`
	Temperature          *float64 `json:"temperature"`
	MaxTokens            *int32   `json:"max_tokens"`
	Capabilities         []string `json:"capabilities"`
	ToolsEnabled         []string `json:"tools_enabled"`
	ContextSources       []string `json:"context_sources"`
	ThinkingEnabled      *bool    `json:"thinking_enabled"`
	StreamingEnabled     *bool    `json:"streaming_enabled"`
	ApplyPersonalization *bool    `json:"apply_personalization"`
	WelcomeMessage       *string  `json:"welcome_message"`
	SuggestedPrompts     []string `json:"suggested_prompts"`
	Category             *string  `json:"category"`
	IsActive             *bool    `json:"is_active"`
	IsPublic             *bool    `json:"is_public"`
	IsFeatured           *bool    `json:"is_featured"`
}

// CreateAgentFromPresetRequest represents request to create agent from preset
type CreateAgentFromPresetRequest struct {
	Name string `json:"name" binding:"required"`
}

// TestAgentRequest represents request to test an agent prompt
type TestAgentRequest struct {
	SystemPrompt string   `json:"system_prompt"`
	TestMessage  string   `json:"test_message" binding:"required"`
	Model        *string  `json:"model"`
	Temperature  *float64 `json:"temperature"`
	MaxTokens    *int     `json:"max_tokens"`
}

// TestAgentResponse represents the sandbox test response
type TestAgentResponse struct {
	Response   string `json:"response"`
	TokensUsed int    `json:"tokens_used"`
	DurationMs int64  `json:"duration_ms"`
	Model      string `json:"model"`
>>>>>>> upstream/main
}
