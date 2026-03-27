package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
)

// TestNewAgentsHandler tests handler creation
func TestNewAgentsHandler(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewAgentsHandler(client, nil)

	if handler == nil {
		t.Fatal("failed to create agents handler")
	}

	if handler.sparqlClient == nil {
		t.Error("expected sparqlClient to be set")
	}

	if handler.logger == nil {
		t.Error("expected logger to be set")
	}
}

// TestListAgents tests the ListAgents endpoint
func TestListAgents(t *testing.T) {
	client := ontology.NewSPARQLClient("http://localhost:7878", nil)
	handler := NewAgentsHandler(client, nil)

	router := gin.New()
	router.GET("/api/ontology/agents", handler.ListAgents)

	req := httptest.NewRequest("GET", "/api/ontology/agents", nil)
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 500 since Oxigraph is not running
	if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
		t.Errorf("expected status 500 or 200, got %d", w.Code)
	}
}

// TestParseAgentResults tests agent result parsing
func TestParseAgentResults(t *testing.T) {
	data := []byte(`{
		"head": {"vars": ["agentId", "agentName", "agentType"]},
		"results": {"bindings": []}
	}`)

	agents := parseAgentResults(data)

	if agents == nil {
		t.Error("expected agents list, got nil")
	}

	if len(agents) != 0 {
		t.Errorf("expected empty agents list, got %d", len(agents))
	}
}
