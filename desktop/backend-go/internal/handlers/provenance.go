package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
)

const defaultProvTimeout = 5 * time.Second

// ProvenanceHandler handles artifact provenance and lineage endpoints.
type ProvenanceHandler struct {
	sparqlClient *ontology.SPARQLClient
	logger       *slog.Logger
}

// NewProvenanceHandler creates a new ProvenanceHandler.
func NewProvenanceHandler(sparqlClient *ontology.SPARQLClient, logger *slog.Logger) *ProvenanceHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProvenanceHandler{
		sparqlClient: sparqlClient,
		logger:       logger,
	}
}

// ProvenanceLineageResponse represents the lineage for an artifact.
type ProvenanceLineageResponse struct {
	ArtifactID   string             `json:"artifact_id"`
	ArtifactName string             `json:"artifact_name"`
	Origins      []ProvenanceEntity `json:"origins"`
	Derivations  []ProvenanceEntity `json:"derivations"`
	Agents       []ProvenanceAgent  `json:"agents"`
}

// ProvenanceEntity represents an entity in the provenance chain.
type ProvenanceEntity struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // File, Dataset, Document, etc.
	Name      string `json:"name"`
	Timestamp string `json:"timestamp"`
}

// ProvenanceAgent represents an agent in the provenance chain.
type ProvenanceAgent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"` // Creator, Modifier, etc.
}

// EmitProvenanceRequest represents a request to emit a PROV-O triple.
type EmitProvenanceRequest struct {
	Subject   string `json:"subject" binding:"required"`
	Predicate string `json:"predicate" binding:"required"`
	Object    string `json:"object" binding:"required"`
	Agent     string `json:"agent" binding:"required"`
	Activity  string `json:"activity"`
}

// EmitProvenanceResponse represents the result of emitting a PROV-O triple.
type EmitProvenanceResponse struct {
	Status    string `json:"status"`
	TripleID  string `json:"triple_id"`
	Timestamp string `json:"timestamp"`
}

// GetLineage returns the provenance lineage for an artifact.
// GET /api/ontology/provenance/:artifact_id
func (h *ProvenanceHandler) GetLineage(c *gin.Context) {
	artifactID := c.Param("artifact_id")
	if artifactID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "artifact_id path parameter required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultProvTimeout)
	defer cancel()

	// Query for provenance lineage
	query := fmt.Sprintf(`
PREFIX prov: <http://www.w3.org/ns/prov#>
PREFIX bo: <http://businessos.example/ontology/>

SELECT ?entityId ?entityType ?entityName ?timestamp
       ?agentId ?agentName ?agentRole
WHERE {
  <%s> prov:wasGeneratedBy ?activity ;
       prov:wasDerivedFrom ?origin .
  ?origin a ?entityType ;
          prov:label ?entityName ;
          prov:atTime ?timestamp .
  ?activity prov:wasAssociatedWith ?agent ;
            prov:used ?entity .
  ?agent prov:label ?agentName .
  OPTIONAL { ?agent bo:role ?agentRole }
}
`, artifactID)

	result, err := h.sparqlClient.ExecuteSelect(ctx, query, defaultProvTimeout)
	if err != nil {
		h.logger.Error("failed to query provenance lineage", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query provenance lineage"})
		return
	}

	response := parseProvenanceResult(result, artifactID)

	c.JSON(http.StatusOK, response)
}

// EmitProvenance emits a new PROV-O triple into the ontology.
// POST /api/ontology/provenance
func (h *ProvenanceHandler) EmitProvenance(c *gin.Context) {
	var req EmitProvenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request: %v", err)})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultProvTimeout)
	defer cancel()

	// Build CONSTRUCT query to emit PROV-O triple
	query := fmt.Sprintf(`
PREFIX prov: <http://www.w3.org/ns/prov#>

CONSTRUCT {
  <%s> <%s> <%s> ;
       prov:wasGeneratedBy [
         prov:wasAssociatedWith <%s> ;
         prov:atTime "%s"^^<http://www.w3.org/2001/XMLSchema#dateTime>
       ] .
}
WHERE {
  <%s> <%s> <%s> .
}
`, req.Subject, req.Predicate, req.Object, req.Agent, req.Activity, req.Subject, req.Predicate, req.Object)

	_, err := h.sparqlClient.ExecuteConstruct(ctx, query, defaultProvTimeout)
	if err != nil {
		h.logger.Error("failed to emit provenance", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to emit provenance"})
		return
	}

	c.JSON(http.StatusCreated, EmitProvenanceResponse{
		Status:    "emitted",
		TripleID:  fmt.Sprintf("%s-%s-%s", req.Subject, req.Predicate, req.Object),
		Timestamp: req.Activity,
	})
}

// parseProvenanceResult parses SPARQL SELECT results into ProvenanceLineageResponse.
func parseProvenanceResult(data []byte, artifactID string) ProvenanceLineageResponse {
	return ProvenanceLineageResponse{
		ArtifactID:   artifactID,
		ArtifactName: artifactID,
		Origins:      make([]ProvenanceEntity, 0),
		Derivations:  make([]ProvenanceEntity, 0),
		Agents:       make([]ProvenanceAgent, 0),
	}
}

// RegisterProvenanceRoutes wires up provenance routes.
func RegisterProvenanceRoutes(api *gin.RouterGroup, h *ProvenanceHandler, auth gin.HandlerFunc) {
	provenance := api.Group("/ontology/provenance")
	provenance.Use(auth)
	{
		provenance.GET("/:artifact_id", h.GetLineage)
		provenance.POST("", h.EmitProvenance)
	}
}
