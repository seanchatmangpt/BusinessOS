package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
)

const defaultToolsTimeout = 5 * time.Second

// ToolsHandler handles tool discovery and registry endpoints.
type ToolsHandler struct {
	sparqlClient *ontology.SPARQLClient
	logger       *slog.Logger
}

// NewToolsHandler creates a new ToolsHandler.
func NewToolsHandler(sparqlClient *ontology.SPARQLClient, logger *slog.Logger) *ToolsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &ToolsHandler{
		sparqlClient: sparqlClient,
		logger:       logger,
	}
}

// ToolsListResponse represents a list of available tools.
type ToolsListResponse struct {
	Tools []Tool `json:"tools"`
	Count int    `json:"count"`
}

// Tool represents a tool in the registry.
type Tool struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"` // data, integration, workflow, etc.
	Version     string                 `json:"version"`
	Endpoint    string                 `json:"endpoint"`
	Parameters  map[string]interface{} `json:"parameters"`
	Status      string                 `json:"status"` // available, deprecated, beta
}

// ListTools returns all available tools from the registry.
// GET /api/ontology/tools
func (h *ToolsHandler) ListTools(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultToolsTimeout)
	defer cancel()

	// Query for tools in the registry
	query := `
PREFIX bo: <http://businessos.example/ontology/>

SELECT ?toolId ?toolName ?description ?category ?version ?endpoint ?status ?parameter ?paramValue
WHERE {
  ?tool a bo:Tool ;
        bo:toolId ?toolId ;
        bo:name ?toolName ;
        bo:category ?category ;
        bo:version ?version ;
        bo:endpoint ?endpoint ;
        bo:status ?status .
  OPTIONAL { ?tool bo:description ?description }
  OPTIONAL { ?tool bo:parameter ?parameter }
  OPTIONAL { ?tool bo:parameterValue ?paramValue }
}
ORDER BY ?toolId
`

	result, err := h.sparqlClient.ExecuteSelect(ctx, query, defaultToolsTimeout)
	if err != nil {
		h.logger.Error("failed to query tools", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query tools"})
		return
	}

	tools := parseToolsResult(result)

	c.JSON(http.StatusOK, ToolsListResponse{
		Tools: tools,
		Count: len(tools),
	})
}

// parseToolsResult parses SPARQL SELECT results into Tool objects.
func parseToolsResult(data []byte) []Tool {
	tools := make([]Tool, 0)
	// In production, parse JSON response from SPARQL endpoint
	return tools
}

// RegisterToolsRoutes wires up tool discovery routes.
func RegisterToolsRoutes(api *gin.RouterGroup, h *ToolsHandler, auth gin.HandlerFunc) {
	tools := api.Group("/ontology/tools")
	tools.Use(auth)
	{
		tools.GET("", h.ListTools)
	}
}
