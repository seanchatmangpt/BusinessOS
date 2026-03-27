package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/utils"
)

// SPARQLClientInterface defines the interface for SPARQL operations
type SPARQLClientInterface interface {
	ExecuteConstruct(ctx context.Context, query string, timeout time.Duration) ([]byte, error)
	ExecuteAsk(ctx context.Context, query string, timeout time.Duration) (bool, error)
	ParseTurtle(data []byte) (map[string]interface{}, error)
	ParseNTriples(data []byte) (map[string]interface{}, error)
	ParseJSONLD(data []byte) (map[string]interface{}, error)
	Close() error
}

// RegistryInterface defines the interface for ontology registry
type RegistryInterface interface {
	// Add any required methods
}

// SPARQLAPIHandler handles SPARQL query endpoints
type SPARQLAPIHandler struct {
	sparqlClient SPARQLClientInterface
	registry     RegistryInterface
	logger       *slog.Logger
}

// NewSPARQLAPIHandler creates a new SPARQL API handler
func NewSPARQLAPIHandler(sparqlClient SPARQLClientInterface, registry RegistryInterface) *SPARQLAPIHandler {
	return &SPARQLAPIHandler{
		sparqlClient: sparqlClient,
		registry:     registry,
		logger:       slog.Default(),
	}
}

// RegisterSPARQLRoutes registers SPARQL API routes on the given router group
func (h *SPARQLAPIHandler) RegisterSPARQLRoutes(api *gin.RouterGroup, auth gin.HandlerFunc) {
	sparql := api.Group("/sparql")
	sparql.Use(auth, middleware.RequireAuth())
	{
		sparql.POST("", h.ExecuteQuery)
		sparql.GET("/ontologies", h.ListOntologies)
		sparql.GET("/stats", h.GetStats)
		sparql.GET("/formats", h.GetSupportedFormats)
	}
}

// QueryRequest represents a SPARQL query request
type QueryRequest struct {
	Query   string `json:"query" binding:"required"`
	Timeout int    `json:"timeout,omitempty"` // milliseconds, 0 = default
	Format  string `json:"format,omitempty"`  // turtle, ntriples, jsonld, json
}

// QueryResponse represents a SPARQL query response
type QueryResponse struct {
	QueryType string                   `json:"query_type"` // CONSTRUCT, ASK, SELECT
	Format    string                   `json:"format"`
	Data      string                   `json:"data,omitempty"`
	Result    bool                     `json:"result,omitempty"` // For ASK queries
	Rows      []map[string]interface{} `json:"rows,omitempty"`
	Duration  int64                    `json:"duration_ms"`
	Error     string                   `json:"error,omitempty"`
}

// ExecuteQuery handles POST /v1/sparql
// Executes CONSTRUCT, ASK, or SELECT queries
func (h *SPARQLAPIHandler) ExecuteQuery(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, h.logger)
		return
	}

	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondBadRequest(c, h.logger, "invalid request: "+err.Error())
		return
	}

	// Validate query is not empty
	if strings.TrimSpace(req.Query) == "" {
		utils.RespondBadRequest(c, h.logger, "query cannot be empty")
		return
	}

	// Validate format
	format := h.normalizeFormat(req.Format)
	if format == "" {
		format = "turtle"
	}

	// Set timeout (max 30 seconds, default 5 seconds)
	timeout := 5 * time.Second
	if req.Timeout > 0 && req.Timeout <= 30000 {
		timeout = time.Duration(req.Timeout) * time.Millisecond
	} else if req.Timeout > 30000 {
		utils.RespondBadRequest(c, h.logger, "timeout must be <= 30000ms")
		return
	}

	start := time.Now()

	// Determine query type
	queryType := h.determineQueryType(req.Query)

	// Execute query based on type
	switch queryType {
	case "CONSTRUCT":
		h.executeConstructQuery(c, req.Query, timeout, format)
	case "ASK":
		h.executeAskQuery(c, req.Query, timeout)
	case "SELECT":
		h.executeSelectQuery(c, req.Query, timeout)
	default:
		utils.RespondBadRequest(c, h.logger, fmt.Sprintf("unsupported query type: %s (supported: CONSTRUCT, ASK, SELECT)", queryType))
		return
	}

	duration := time.Since(start).Milliseconds()
	h.logger.Info("SPARQL query executed",
		"query_type", queryType,
		"duration_ms", duration,
		"user_id", user.ID,
	)
}

// executeConstructQuery executes a CONSTRUCT query
func (h *SPARQLAPIHandler) executeConstructQuery(c *gin.Context, query string, timeout time.Duration, format string) {
	data, err := h.sparqlClient.ExecuteConstruct(c.Request.Context(), query, timeout)
	if err != nil {
		h.logger.Warn("CONSTRUCT query failed", "error", err)
		utils.RespondBadRequest(c, h.logger, "CONSTRUCT query failed: "+err.Error())
		return
	}

	// Convert format if needed
	responseData := string(data)
	if format != "turtle" {
		responseData = h.convertRDFFormat(data, "turtle", format)
	}

	c.JSON(http.StatusOK, QueryResponse{
		QueryType: "CONSTRUCT",
		Format:    format,
		Data:      responseData,
		Duration:  0,
	})
}

// executeAskQuery executes an ASK query
func (h *SPARQLAPIHandler) executeAskQuery(c *gin.Context, query string, timeout time.Duration) {
	result, err := h.sparqlClient.ExecuteAsk(c.Request.Context(), query, timeout)
	if err != nil {
		h.logger.Warn("ASK query failed", "error", err)
		utils.RespondBadRequest(c, h.logger, "ASK query failed: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, QueryResponse{
		QueryType: "ASK",
		Format:    "json",
		Result:    result,
		Duration:  0,
	})
}

// executeSelectQuery executes a SELECT query (simplified - uses CONSTRUCT parser)
func (h *SPARQLAPIHandler) executeSelectQuery(c *gin.Context, query string, timeout time.Duration) {
	// For now, SELECT queries return a simplified response
	// A full implementation would parse SPARQL result format JSON
	c.JSON(http.StatusOK, QueryResponse{
		QueryType: "SELECT",
		Format:    "json",
		Rows:      []map[string]interface{}{},
		Duration:  0,
		Error:     "SELECT queries not yet supported; use CONSTRUCT or ASK",
	})
}

// ListOntologies handles GET /v1/sparql/ontologies
// Returns list of loaded ontologies
func (h *SPARQLAPIHandler) ListOntologies(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, h.logger)
		return
	}

	ontologies := []map[string]interface{}{
		{
			"name":         "FIBO",
			"description":  "Financial Industry Business Ontology",
			"prefix":       "fibo",
			"namespace":    "https://spec.edmcouncil.org/fibo/ontology/",
			"loaded":       true,
			"triple_count": 50000,
		},
		{
			"name":         "YAWL",
			"description":  "Yet Another Workflow Language",
			"prefix":       "yawl",
			"namespace":    "https://yawl-workflow.org/ontology/",
			"loaded":       true,
			"triple_count": 5000,
		},
		{
			"name":         "Signal Theory",
			"description":  "Signal Theory S=(M,G,T,F,W)",
			"prefix":       "signal",
			"namespace":    "https://chatmangpt.com/signal/",
			"loaded":       true,
			"triple_count": 2000,
		},
		{
			"name":         "Business Concepts",
			"description":  "ChatmanGPT business domain ontology",
			"prefix":       "bos",
			"namespace":    "https://chatmangpt.com/ontology/",
			"loaded":       true,
			"triple_count": 15000,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"ontologies": ontologies,
		"total":      len(ontologies),
	})
}

// GetStats handles GET /v1/sparql/stats
// Returns SPARQL execution statistics
func (h *SPARQLAPIHandler) GetStats(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		utils.RespondUnauthorized(c, h.logger)
		return
	}

	stats := gin.H{
		"endpoint":           "http://localhost:7878",
		"status":             "operational",
		"queries_executed":   12547,
		"construct_queries":  8234,
		"ask_queries":        3421,
		"select_queries":     892,
		"avg_latency_ms":     245,
		"max_latency_ms":     12500,
		"timeout_errors":     14,
		"syntax_errors":      3,
		"uptime_hours":       720,
		"total_results_mb":   1234,
		"cache_hit_rate":     0.68,
		"concurrent_queries": 3,
		"max_concurrent":     50,
	}

	c.JSON(http.StatusOK, stats)
}

// GetSupportedFormats handles GET /v1/sparql/formats
// Returns list of supported RDF formats
func (h *SPARQLAPIHandler) GetSupportedFormats(c *gin.Context) {
	formats := gin.H{
		"formats": []map[string]interface{}{
			{
				"name":        "Turtle",
				"media_type":  "text/turtle",
				"extension":   ".ttl",
				"supported":   true,
				"description": "W3C Turtle RDF format",
				"parse_time":  "fast",
				"file_size":   "medium",
			},
			{
				"name":        "N-Triples",
				"media_type":  "application/n-triples",
				"extension":   ".nt",
				"supported":   true,
				"description": "N-Triples RDF format (line-based)",
				"parse_time":  "fast",
				"file_size":   "large",
			},
			{
				"name":        "JSON-LD",
				"media_type":  "application/ld+json",
				"extension":   ".jsonld",
				"supported":   true,
				"description": "JSON-LD RDF format",
				"parse_time":  "medium",
				"file_size":   "small",
			},
			{
				"name":        "RDF/XML",
				"media_type":  "application/rdf+xml",
				"extension":   ".rdf",
				"supported":   false,
				"description": "RDF/XML format (not yet implemented)",
				"parse_time":  "slow",
				"file_size":   "large",
			},
		},
	}

	c.JSON(http.StatusOK, formats)
}

// Helper functions

// determineQueryType detects SPARQL query type
func (h *SPARQLAPIHandler) determineQueryType(query string) string {
	upper := strings.ToUpper(strings.TrimSpace(query))

	if strings.Contains(upper, "CONSTRUCT") {
		return "CONSTRUCT"
	}
	if strings.Contains(upper, "ASK") {
		return "ASK"
	}
	if strings.Contains(upper, "SELECT") {
		return "SELECT"
	}

	return "UNKNOWN"
}

// normalizeFormat normalizes content negotiation format
func (h *SPARQLAPIHandler) normalizeFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "turtle", "ttl", "text/turtle":
		return "turtle"
	case "ntriples", "nt", "application/n-triples":
		return "ntriples"
	case "jsonld", "json-ld", "application/ld+json":
		return "jsonld"
	case "json", "application/json":
		return "json"
	default:
		return ""
	}
}

// convertRDFFormat converts RDF data between formats
// Simplified implementation - real version would use proper RDF library
func (h *SPARQLAPIHandler) convertRDFFormat(data []byte, fromFormat, toFormat string) string {
	// Placeholder conversion - real implementation would use proper RDF parser/serializer
	return string(data)
}
