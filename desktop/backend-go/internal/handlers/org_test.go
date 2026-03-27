package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
)

// TestNewOrgStructureHandler tests handler creation
func TestNewOrgStructureHandler(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewOrgStructureHandler(client, nil)

	if handler == nil {
		t.Fatal("failed to create org structure handler")
	}

	if handler.sparqlClient == nil {
		t.Error("expected sparqlClient to be set")
	}
}

// TestGetOrgStructure tests the GetOrgStructure endpoint
func TestGetOrgStructure(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewOrgStructureHandler(client, nil)

	router := gin.New()
	router.GET("/api/ontology/org", handler.GetOrgStructure)

	req := httptest.NewRequest("GET", "/api/ontology/org", nil)
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should succeed or return internal error depending on Oxigraph availability
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Errorf("expected status 500 or 200, got %d", w.Code)
	}
}

// TestParseOrgStructureResult tests org structure result parsing
func TestParseOrgStructureResult(t *testing.T) {
	data := []byte(`{
		"head": {"vars": ["deptId", "deptName"]},
		"results": {"bindings": []}
	}`)

	result := parseOrgStructureResult(data)

	if result.Organization == "" {
		t.Error("expected organization name to be set")
	}

	if result.Departments == nil {
		t.Error("expected departments list, got nil")
	}

	if result.Roles == nil {
		t.Error("expected roles list, got nil")
	}

	if result.ReportingLines == nil {
		t.Error("expected reporting lines list, got nil")
	}
}
