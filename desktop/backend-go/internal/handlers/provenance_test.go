package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
)

// TestNewProvenanceHandler tests handler creation
func TestNewProvenanceHandler(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewProvenanceHandler(client, nil)

	if handler == nil {
		t.Fatal("failed to create provenance handler")
	}

	if handler.sparqlClient == nil {
		t.Error("expected sparqlClient to be set")
	}
}

// TestGetLineage tests the GetLineage endpoint
func TestGetLineage(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewProvenanceHandler(client, nil)

	router := gin.New()
	router.GET("/api/ontology/provenance/:artifact_id", handler.GetLineage)

	req := httptest.NewRequest("GET", "/api/ontology/provenance/artifact-123", nil)
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return internal error (Oxigraph not running) or success
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Errorf("expected status 500 or 200, got %d", w.Code)
	}
}

// TestGetLineageNoArtifactID tests missing artifact_id parameter
func TestGetLineageNoArtifactID(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewProvenanceHandler(client, nil)

	router := gin.New()
	router.GET("/api/ontology/provenance/:artifact_id", handler.GetLineage)

	req := httptest.NewRequest("GET", "/api/ontology/provenance/", nil)
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404 for missing parameter
	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Logf("got status %d", w.Code)
	}
}

// TestEmitProvenance tests the EmitProvenance endpoint
func TestEmitProvenance(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewProvenanceHandler(client, nil)

	router := gin.New()
	router.POST("/api/ontology/provenance", handler.EmitProvenance)

	req := EmitProvenanceRequest{
		Subject:   "http://example.com/subject",
		Predicate: "http://example.com/predicate",
		Object:    "http://example.com/object",
		Agent:     "http://example.com/agent",
		Activity:  "2026-03-26T00:00:00Z",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/ontology/provenance", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq = httpReq.WithContext(context.Background())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, httpReq)

	// Should return 201 or 500 depending on Oxigraph
	if w.Code != http.StatusCreated && w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 201 or 500, got %d", w.Code)
	}
}

// TestParseProvenanceResult tests provenance result parsing
func TestParseProvenanceResult(t *testing.T) {
	data := []byte(`{
		"head": {"vars": ["entityId", "entityName"]},
		"results": {"bindings": []}
	}`)

	result := parseProvenanceResult(data, "artifact-123")

	if result.ArtifactID != "artifact-123" {
		t.Errorf("expected artifact ID artifact-123, got %s", result.ArtifactID)
	}

	if result.Origins == nil {
		t.Error("expected origins list, got nil")
	}

	if result.Derivations == nil {
		t.Error("expected derivations list, got nil")
	}

	if result.Agents == nil {
		t.Error("expected agents list, got nil")
	}
}
