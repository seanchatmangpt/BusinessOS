package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
)

// TestNewToolsHandler tests handler creation
func TestNewToolsHandler(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewToolsHandler(client, nil)

	if handler == nil {
		t.Fatal("failed to create tools handler")
	}

	if handler.sparqlClient == nil {
		t.Error("expected sparqlClient to be set")
	}
}

// TestListTools tests the ListTools endpoint
func TestListTools(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewToolsHandler(client, nil)

	router := gin.New()
	router.GET("/api/ontology/tools", handler.ListTools)

	req := httptest.NewRequest("GET", "/api/ontology/tools", nil)
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return internal error (Oxigraph not running) or success
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Errorf("expected status 500 or 200, got %d", w.Code)
	}
}

// TestParseToolsResult tests tool result parsing
func TestParseToolsResult(t *testing.T) {
	data := []byte(`{
		"head": {"vars": ["toolId", "toolName"]},
		"results": {"bindings": []}
	}`)

	tools := parseToolsResult(data)

	if tools == nil {
		t.Error("expected tools list, got nil")
	}

	if len(tools) != 0 {
		t.Errorf("expected empty tools list, got %d", len(tools))
	}
}
