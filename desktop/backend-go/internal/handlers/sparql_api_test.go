package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSPARQLClient mocks the SPARQL client for testing
type MockSPARQLClient struct {
	ConstructResult []byte
	ConstructError  error
	AskResult       bool
	AskError        error
}

func (m *MockSPARQLClient) ExecuteConstruct(ctx context.Context, query string, timeout time.Duration) ([]byte, error) {
	if m.ConstructError != nil {
		return nil, m.ConstructError
	}
	return m.ConstructResult, nil
}

func (m *MockSPARQLClient) ExecuteAsk(ctx context.Context, query string, timeout time.Duration) (bool, error) {
	if m.AskError != nil {
		return false, m.AskError
	}
	return m.AskResult, nil
}

func (m *MockSPARQLClient) ParseTurtle(data []byte) (map[string]interface{}, error) {
	return map[string]interface{}{
		"format": "turtle",
		"bytes":  len(data),
	}, nil
}

func (m *MockSPARQLClient) ParseNTriples(data []byte) (map[string]interface{}, error) {
	return map[string]interface{}{
		"format":  "ntriples",
		"triples": bytes.Count(data, []byte("\n")),
	}, nil
}

func (m *MockSPARQLClient) ParseJSONLD(data []byte) (map[string]interface{}, error) {
	return map[string]interface{}{
		"format": "jsonld",
		"bytes":  len(data),
	}, nil
}

func (m *MockSPARQLClient) Close() error {
	return nil
}

// Setup test handler
func setupTestHandler() *SPARQLAPIHandler {
	mockClient := &MockSPARQLClient{
		ConstructResult: []byte("@prefix fibo: <http://example.org/> .\nfibo:example fibo:property fibo:value ."),
		AskResult:       true,
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return &SPARQLAPIHandler{
		sparqlClient: mockClient,
		registry:     nil,
		logger:       logger,
	}
}

// Helper to create test context with user
func setupTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Add user to context
		mockUser := &middleware.BetterAuthUser{
			ID:    "test-user-id",
			Email: "test@example.com",
		}
	c.Set("user", mockUser)

	return c, w
}

// Tests

func TestExecuteConstructQuery(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	query := "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o LIMIT 10 }"
	req := QueryRequest{
		Query:   query,
		Timeout: 5000,
		Format:  "turtle",
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/api/v1/sparql", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteQuery(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp QueryResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "CONSTRUCT", resp.QueryType)
	assert.Equal(t, "turtle", resp.Format)
	assert.NotEmpty(t, resp.Data)
}

func TestExecuteAskQuery(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	query := "ASK { ?s <http://example.org/hasValue> ?o }"
	req := QueryRequest{
		Query:   query,
		Timeout: 3000,
		Format:  "json",
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/api/v1/sparql", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteQuery(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp QueryResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "ASK", resp.QueryType)
	assert.True(t, resp.Result)
}

func TestEmptyQuery(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	req := QueryRequest{
		Query:   "   ",
		Timeout: 5000,
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/api/v1/sparql", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteQuery(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTimeoutExceeded(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	req := QueryRequest{
		Query:   "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }",
		Timeout: 40000, // > 30000ms max
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/api/v1/sparql", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteQuery(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestFormatNegotiation_Turtle(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	req := QueryRequest{
		Query:   "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o LIMIT 5 }",
		Timeout: 5000,
		Format:  "ttl",
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/api/v1/sparql", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteQuery(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp QueryResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "turtle", resp.Format)
}

func TestFormatNegotiation_JSONLD(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	req := QueryRequest{
		Query:   "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o LIMIT 5 }",
		Timeout: 5000,
		Format:  "jsonld",
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/api/v1/sparql", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteQuery(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp QueryResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "jsonld", resp.Format)
}

func TestListOntologies(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	c.Request = httptest.NewRequest("GET", "/api/v1/sparql/ontologies", nil)

	handler.ListOntologies(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Ontologies []map[string]interface{} `json:"ontologies"`
		Total      int                      `json:"total"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, 4, resp.Total)
	assert.True(t, len(resp.Ontologies) > 0)
	assert.Equal(t, "FIBO", resp.Ontologies[0]["name"])
}

func TestListOntologiesHasFIBO(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	c.Request = httptest.NewRequest("GET", "/api/v1/sparql/ontologies", nil)
	handler.ListOntologies(c)

	var resp struct {
		Ontologies []map[string]interface{} `json:"ontologies"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	found := false
	for _, o := range resp.Ontologies {
		if o["name"] == "FIBO" {
			found = true
			assert.True(t, o["loaded"].(bool))
			break
		}
	}
	assert.True(t, found, "FIBO ontology not found")
}

func TestGetStats(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	c.Request = httptest.NewRequest("GET", "/api/v1/sparql/stats", nil)

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)
	assert.NotNil(t, stats["endpoint"])
	assert.NotNil(t, stats["queries_executed"])
	assert.NotNil(t, stats["avg_latency_ms"])
}

func TestGetStatsHasMetrics(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	c.Request = httptest.NewRequest("GET", "/api/v1/sparql/stats", nil)
	handler.GetStats(c)

	var stats map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &stats)

	assert.Contains(t, stats, "construct_queries")
	assert.Contains(t, stats, "ask_queries")
	assert.Contains(t, stats, "cache_hit_rate")
	assert.Greater(t, stats["queries_executed"].(float64), float64(0))
}

func TestGetSupportedFormats(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	c.Request = httptest.NewRequest("GET", "/api/v1/sparql/formats", nil)

	handler.GetSupportedFormats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Formats []map[string]interface{} `json:"formats"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Greater(t, len(resp.Formats), 0)
}

func TestGetSupportedFormatsIncludesTurtle(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	c.Request = httptest.NewRequest("GET", "/api/v1/sparql/formats", nil)
	handler.GetSupportedFormats(c)

	var resp struct {
		Formats []map[string]interface{} `json:"formats"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	found := false
	for _, f := range resp.Formats {
		if f["name"] == "Turtle" {
			found = true
			assert.True(t, f["supported"].(bool))
			break
		}
	}
	assert.True(t, found, "Turtle format not found")
}

func TestDetermineQueryType_Construct(t *testing.T) {
	handler := setupTestHandler()
	query := "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o }"
	queryType := handler.determineQueryType(query)
	assert.Equal(t, "CONSTRUCT", queryType)
}

func TestDetermineQueryType_Ask(t *testing.T) {
	handler := setupTestHandler()
	query := "ASK WHERE { ?s ?p ?o }"
	queryType := handler.determineQueryType(query)
	assert.Equal(t, "ASK", queryType)
}

func TestDetermineQueryType_Select(t *testing.T) {
	handler := setupTestHandler()
	query := "SELECT ?s ?p WHERE { ?s ?p ?o }"
	queryType := handler.determineQueryType(query)
	assert.Equal(t, "SELECT", queryType)
}

func TestNormalizeFormat_Turtle(t *testing.T) {
	handler := setupTestHandler()
	assert.Equal(t, "turtle", handler.normalizeFormat("turtle"))
	assert.Equal(t, "turtle", handler.normalizeFormat("ttl"))
	assert.Equal(t, "turtle", handler.normalizeFormat("text/turtle"))
}

func TestNormalizeFormat_NTriples(t *testing.T) {
	handler := setupTestHandler()
	assert.Equal(t, "ntriples", handler.normalizeFormat("ntriples"))
	assert.Equal(t, "ntriples", handler.normalizeFormat("nt"))
	assert.Equal(t, "ntriples", handler.normalizeFormat("application/n-triples"))
}

func TestNormalizeFormat_JSONLD(t *testing.T) {
	handler := setupTestHandler()
	assert.Equal(t, "jsonld", handler.normalizeFormat("jsonld"))
	assert.Equal(t, "jsonld", handler.normalizeFormat("json-ld"))
	assert.Equal(t, "jsonld", handler.normalizeFormat("application/ld+json"))
}

func TestConstructQueryWithFilters(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	query := `CONSTRUCT { ?s ?p ?o }
             WHERE {
               ?s ?p ?o .
               FILTER (?p = <http://example.org/name>)
             }
             LIMIT 100`
	req := QueryRequest{
		Query:   query,
		Timeout: 5000,
		Format:  "turtle",
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/api/v1/sparql", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteQuery(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAskQueryComplexPattern(t *testing.T) {
	handler := setupTestHandler()
	c, w := setupTestContext()

	query := `ASK {
               ?s <http://example.org/hasAddress> ?address .
               ?address <http://example.org/country> "US"
             }`
	req := QueryRequest{
		Query:   query,
		Timeout: 3000,
	}

	body, _ := json.Marshal(req)
	c.Request = httptest.NewRequest("POST", "/api/v1/sparql", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.ExecuteQuery(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp QueryResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "ASK", resp.QueryType)
}

func BenchmarkExecuteConstructQuery(b *testing.B) {
	handler := setupTestHandler()

	query := "CONSTRUCT { ?s ?p ?o } WHERE { ?s ?p ?o LIMIT 10 }"
	req := QueryRequest{
		Query:   query,
		Timeout: 5000,
		Format:  "turtle",
	}

	body, _ := json.Marshal(req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, w := setupTestContext()
		c.Request = httptest.NewRequest("POST", "/api/v1/sparql", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.ExecuteQuery(c)
		if w.Code != http.StatusOK {
			b.Fatalf("Unexpected status code: %d", w.Code)
		}
	}
}
