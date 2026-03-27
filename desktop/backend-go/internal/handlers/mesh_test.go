package handlers

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupMeshTestRouter creates a test Gin router with DataMesh routes.
func setupMeshTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock auth middleware that just continues (no auth required)
	mockAuth := func(c *gin.Context) {
		c.Next()
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	handler := NewDataMeshHandler("http://localhost:3030", logger)

	api := router.Group("/api")
	// Manually register routes without RequireAuth for testing
	mesh := api.Group("/mesh")
	mesh.Use(mockAuth)
	{
		mesh.POST("/domains", handler.RegisterDomain)
		mesh.POST("/contracts", handler.DefineContract)
		mesh.GET("/discover", handler.DiscoverDatasets)
		mesh.GET("/lineage", handler.QueryLineage)
		mesh.GET("/quality", handler.CheckQuality)
		mesh.GET("/domains/list", handler.ListDomains)
	}

	return router
}

// TestRegisterDomainHandler tests POST /api/mesh/domains
func TestRegisterDomainHandler(t *testing.T) {
	router := setupMeshTestRouter(t)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "register valid finance domain",
			payload: RegisterDomainRequest{
				Name:        "Finance",
				Description: "Financial data domain",
				Owner:       "finance-team",
				IRI:         "http://data.example.com/domain/finance",
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"domain_id", "domain_name", "iri", "created_at", "status"},
		},
		{
			name: "register valid operations domain",
			payload: RegisterDomainRequest{
				Name:        "Operations",
				Description: "Operations domain",
				Owner:       "ops-team",
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"domain_id", "status"},
		},
		{
			name:           "missing required fields",
			payload:        RegisterDomainRequest{Name: "Finance"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "unsupported domain",
			payload: RegisterDomainRequest{
				Name:  "UnsupportedDomain",
				Owner: "someone",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			payload:        "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.payload.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}

			req := httptest.NewRequest("POST", "/api/mesh/domains",
				bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusCreated {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				for _, field := range tt.expectedFields {
					if _, exists := resp[field]; !exists {
						t.Errorf("expected field '%s' not found in response", field)
					}
				}
			}
		})
	}
}

// TestDefineContractHandler tests POST /api/mesh/contracts
func TestDefineContractHandler(t *testing.T) {
	router := setupMeshTestRouter(t)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "define valid contract",
			payload: map[string]interface{}{
				"domain_id":    "domain_finance",
				"name":         "Transaction Contract",
				"description":  "Standard transaction contract",
				"entities":     []string{"http://data.example.com/entity/Transaction"},
				"constraints": []map[string]string{
					{
						"name":        "Amount Required",
						"type":        "required_field",
						"description": "Amount is required",
						"expression":  "EXISTS(?amount)",
						"severity":    "error",
					},
				},
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"contract_id", "contract_name", "domain_id", "status"},
		},
		{
			name: "contract missing domain",
			payload: DefineContractRequest{
				Name: "SomeContract",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "contract missing constraints",
			payload: DefineContractRequest{
				DomainID: "domain_finance",
				Name:     "EmptyContract",
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}

			req := httptest.NewRequest("POST", "/api/mesh/contracts",
				bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusCreated && len(tt.expectedFields) > 0 {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				for _, field := range tt.expectedFields {
					if _, exists := resp[field]; !exists {
						t.Errorf("expected field '%s' not found in response", field)
					}
				}
			}
		})
	}
}

// TestDiscoverDatasetsHandler tests GET /api/mesh/discover
func TestDiscoverDatasetsHandler(t *testing.T) {
	router := setupMeshTestRouter(t)

	tests := []struct {
		name           string
		domainID       string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "discover finance datasets",
			domainID:       "domain_finance",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"domain_id", "datasets", "count"},
		},
		{
			name:           "discover operations datasets",
			domainID:       "domain_operations",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"domain_id", "count"},
		},
		{
			name:           "missing domain id",
			domainID:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/mesh/discover"
			if tt.domainID != "" {
				url += "?domain_id=" + tt.domainID
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK && len(tt.expectedFields) > 0 {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				for _, field := range tt.expectedFields {
					if _, exists := resp[field]; !exists {
						t.Errorf("expected field '%s' not found in response", field)
					}
				}
			}
		})
	}
}

// TestQueryLineageHandler tests GET /api/mesh/lineage
func TestQueryLineageHandler(t *testing.T) {
	router := setupMeshTestRouter(t)

	tests := []struct {
		name           string
		datasetID      string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "query lineage for dataset",
			datasetID:      "dataset_transactions",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"dataset_id", "lineage", "depth"},
		},
		{
			name:           "missing dataset id",
			datasetID:      "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/mesh/lineage"
			if tt.datasetID != "" {
				url += "?dataset_id=" + tt.datasetID
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK && len(tt.expectedFields) > 0 {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				for _, field := range tt.expectedFields {
					if _, exists := resp[field]; !exists {
						t.Errorf("expected field '%s' not found in response", field)
					}
				}

				// Verify depth is bounded
				if depth, ok := resp["depth"].(float64); ok && depth > 5 {
					t.Errorf("lineage depth exceeds limit: %f > 5", depth)
				}
			}
		})
	}
}

// TestCheckQualityHandler tests GET /api/mesh/quality
func TestCheckQualityHandler(t *testing.T) {
	router := setupMeshTestRouter(t)

	tests := []struct {
		name           string
		datasetID      string
		expectedStatus int
		expectedFields []string
	}{
		{
			name:           "check quality for dataset",
			datasetID:      "dataset_ledger",
			expectedStatus: http.StatusOK,
			expectedFields: []string{"dataset_id", "quality", "metrics"},
		},
		{
			name:           "missing dataset id",
			datasetID:      "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/mesh/quality"
			if tt.datasetID != "" {
				url += "?dataset_id=" + tt.datasetID
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				// Verify metrics structure
				if metrics, ok := resp["metrics"].(map[string]interface{}); ok {
					requiredMetrics := []string{"completeness", "accuracy", "consistency", "timeliness", "overall"}
					for _, metric := range requiredMetrics {
						if _, exists := metrics[metric]; !exists {
							t.Errorf("expected metric '%s' not found", metric)
						}
					}

					// Verify metrics are bounded [0, 100]
					for metric, val := range metrics {
						if score, ok := val.(float64); ok {
							if score < 0 || score > 100 {
								t.Errorf("metric '%s' out of bounds: %f", metric, score)
							}
						}
					}
				}
			}
		})
	}
}

// TestListDomainsHandler tests GET /api/mesh/domains/list
func TestListDomainsHandler(t *testing.T) {
	router := setupMeshTestRouter(t)

	req := httptest.NewRequest("GET", "/api/mesh/domains/list", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify response structure
	if _, exists := resp["domains"]; !exists {
		t.Error("expected field 'domains' not found in response")
	}

	if _, exists := resp["count"]; !exists {
		t.Error("expected field 'count' not found in response")
	}

	// Verify all 5 domains present
	if domains, ok := resp["domains"].([]interface{}); ok {
		if len(domains) < 5 {
			t.Errorf("expected 5 domains, got %d", len(domains))
		}
	}
}

// TestMeshHandlerContentType verifies JSON content type on all responses
func TestMeshHandlerContentType(t *testing.T) {
	router := setupMeshTestRouter(t)

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/api/mesh/domains/list"},
		{"GET", "/api/mesh/discover?domain_id=finance"},
		{"GET", "/api/mesh/lineage?dataset_id=test"},
		{"GET", "/api/mesh/quality?dataset_id=test"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}

// TestMeshHandlerErrorResponse verifies error response format
func TestMeshHandlerErrorResponse(t *testing.T) {
	router := setupMeshTestRouter(t)

	// Query with missing required field
	req := httptest.NewRequest("GET", "/api/mesh/discover", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Error response should have error and message fields
	if _, exists := resp["error"]; !exists {
		t.Error("expected field 'error' in error response")
	}
}

// TestQualityMetricsBounded verifies quality metrics stay within [0, 100]
func TestQualityMetricsBounded(t *testing.T) {
	router := setupMeshTestRouter(t)

	req := httptest.NewRequest("GET", "/api/mesh/quality?dataset_id=test_dataset", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if metrics, ok := resp["metrics"].(map[string]interface{}); ok {
		for metric, val := range metrics {
			if score, ok := val.(float64); ok {
				if score < 0 || score > 100 {
					t.Errorf("metric '%s' out of bounds: %f", metric, score)
				}
			}
		}
	}
}
