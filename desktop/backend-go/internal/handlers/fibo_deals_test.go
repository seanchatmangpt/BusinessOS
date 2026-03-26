package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/services"
)

// setupFIBODealsTest creates a test router with deals handlers.
func setupFIBODealsTest(dealsService *services.FIBODealsService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock auth middleware (passes through)
	mockAuth := func(c *gin.Context) {
		c.Next()
	}

	api := router.Group("/api")
	dealsHandler := NewFIBODealsHandler(dealsService)
	RegisterFIBODealsRoutes(api, dealsHandler, mockAuth)

	return router
}

// TestHandlerCreateDealSuccess tests POST /api/deals with valid input.
func TestHandlerCreateDealSuccess(t *testing.T) {
	// Create mock Oxigraph server
	oxigraphServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-test-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer oxigraphServer.Close()

	service := services.NewFIBODealsService(oxigraphServer.URL)
	router := setupFIBODealsTest(service)

	// Prepare request
	payload := createDealRequest{
		Name:        "Test Deal",
		Amount:      100000.00,
		Currency:    "USD",
		BuyerID:     "buyer-1",
		SellerID:    "seller-1",
		Probability: 75,
		Stage:       "negotiation",
	}
	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/deals", bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response dealResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Name != "Test Deal" {
		t.Errorf("expected name 'Test Deal', got '%s'", response.Name)
	}
	if response.Amount != 100000.00 {
		t.Errorf("expected amount 100000.00, got %f", response.Amount)
	}
}

// TestHandlerCreateDealMissingRequired tests POST /api/deals with missing required fields.
func TestHandlerCreateDealMissingRequired(t *testing.T) {
	service := services.NewFIBODealsService("http://localhost:8890")
	router := setupFIBODealsTest(service)

	// Missing "name" field
	payload := createDealRequest{
		Amount:      100000.00,
		BuyerID:     "buyer-1",
		SellerID:    "seller-1",
		Probability: 75,
	}
	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/deals", bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestHandlerCreateDealInvalidAmount tests POST /api/deals with invalid amount.
func TestHandlerCreateDealInvalidAmount(t *testing.T) {
	service := services.NewFIBODealsService("http://localhost:8890")
	router := setupFIBODealsTest(service)

	payload := createDealRequest{
		Name:        "Bad Amount Deal",
		Amount:      -100.00, // Invalid: negative
		BuyerID:     "buyer-1",
		SellerID:    "seller-1",
		Probability: 75,
	}
	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/deals", bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestHandlerGetDealSuccess tests GET /api/deals/:id.
func TestHandlerGetDealSuccess(t *testing.T) {
	oxigraphServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-123> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer oxigraphServer.Close()

	service := services.NewFIBODealsService(oxigraphServer.URL)
	router := setupFIBODealsTest(service)

	req := httptest.NewRequest("GET", "/api/deals/d-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response dealResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.ID != "d-123" {
		t.Errorf("expected ID 'd-123', got '%s'", response.ID)
	}
}

// TestHandlerListDealsSuccess tests GET /api/deals.
func TestHandlerListDealsSuccess(t *testing.T) {
	oxigraphServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer oxigraphServer.Close()

	service := services.NewFIBODealsService(oxigraphServer.URL)
	router := setupFIBODealsTest(service)

	req := httptest.NewRequest("GET", "/api/deals?limit=50&offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["count"] == nil {
		t.Error("expected 'count' field in response")
	}
	if response["deals"] == nil {
		t.Error("expected 'deals' field in response")
	}
}

// TestHandlerUpdateDealSuccess tests PATCH /api/deals/:id.
func TestHandlerUpdateDealSuccess(t *testing.T) {
	oxigraphServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-update-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer oxigraphServer.Close()

	service := services.NewFIBODealsService(oxigraphServer.URL)
	router := setupFIBODealsTest(service)

	// Prepare update request
	newName := "Updated Name"
	payload := updateDealRequest{
		Name: &newName,
	}
	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("PATCH", "/api/deals/d-update-1", bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response dealResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.ID != "d-update-1" {
		t.Errorf("expected ID 'd-update-1', got '%s'", response.ID)
	}
}

// TestHandlerUpdateDealMissingID tests PATCH /api/deals without ID.
func TestHandlerUpdateDealMissingID(t *testing.T) {
	service := services.NewFIBODealsService("http://localhost:8890")
	router := setupFIBODealsTest(service)

	payload := updateDealRequest{}
	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("PATCH", "/api/deals/", bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Gin treats missing parameter differently, so just verify not success
	if w.Code == http.StatusOK {
		t.Error("expected non-200 status for missing ID")
	}
}

// TestHandlerVerifyComplianceSuccess tests POST /api/deals/:id/verify-compliance.
func TestHandlerVerifyComplianceSuccess(t *testing.T) {
	oxigraphServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/parties/acme-corp> <https://businessos.dev/id/hasKYCStatus> <https://businessos.dev/id/KYCVerified> .
`))
	}))
	defer oxigraphServer.Close()

	service := services.NewFIBODealsService(oxigraphServer.URL)
	router := setupFIBODealsTest(service)

	req := httptest.NewRequest("POST", "/api/deals/d-compliance-1/verify-compliance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["deal_id"] == nil {
		t.Error("expected 'deal_id' field in response")
	}
	if response["compliance"] == nil {
		t.Error("expected 'compliance' field in response")
	}
}

// TestHandlerInvalidJSON tests POST /api/deals with invalid JSON.
func TestHandlerInvalidJSON(t *testing.T) {
	service := services.NewFIBODealsService("http://localhost:8890")
	router := setupFIBODealsTest(service)

	req := httptest.NewRequest("POST", "/api/deals", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestHandlerListDealsWithPagination tests GET /api/deals with custom pagination.
func TestHandlerListDealsWithPagination(t *testing.T) {
	oxigraphServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer oxigraphServer.Close()

	service := services.NewFIBODealsService(oxigraphServer.URL)
	router := setupFIBODealsTest(service)

	req := httptest.NewRequest("GET", "/api/deals?limit=100&offset=25", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if limit, ok := response["limit"].(float64); !ok || limit != 100 {
		t.Errorf("expected limit 100, got %v", response["limit"])
	}
	if offset, ok := response["offset"].(float64); !ok || offset != 25 {
		t.Errorf("expected offset 25, got %v", response["offset"])
	}
}

// TestHandlerCreateDealWithCloseDateSuccess tests POST /api/deals with expected_close_date.
func TestHandlerCreateDealWithCloseDateSuccess(t *testing.T) {
	oxigraphServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-date-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer oxigraphServer.Close()

	service := services.NewFIBODealsService(oxigraphServer.URL)
	router := setupFIBODealsTest(service)

	payload := createDealRequest{
		Name:              "Deal with Date",
		Amount:            100000.00,
		BuyerID:           "buyer-1",
		SellerID:          "seller-1",
		Probability:       75,
		ExpectedCloseDate: "2026-06-30T00:00:00Z",
	}
	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/deals", bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response dealResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.ExpectedCloseDate == "" {
		t.Error("expected close date should not be empty")
	}
}

// TestHandlerCreateDealValidateProbability tests probability bounds.
func TestHandlerCreateDealValidateProbability(t *testing.T) {
	service := services.NewFIBODealsService("http://localhost:8890")
	router := setupFIBODealsTest(service)

	// Test probability > 100
	payload := createDealRequest{
		Name:        "Bad Probability",
		Amount:      100000.00,
		BuyerID:     "buyer-1",
		SellerID:    "seller-1",
		Probability: 150, // Invalid
	}
	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/deals", bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for probability > 100, got %d", w.Code)
	}
}

// TestHandlerUpdateDealMultipleFields tests PATCH /api/deals/:id with multiple fields.
func TestHandlerUpdateDealMultipleFields(t *testing.T) {
	oxigraphServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-multi-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer oxigraphServer.Close()

	service := services.NewFIBODealsService(oxigraphServer.URL)
	router := setupFIBODealsTest(service)

	newName := "Multi Update"
	newAmount := 250000.0
	newStatus := "active"
	payload := updateDealRequest{
		Name:   &newName,
		Amount: &newAmount,
		Status: &newStatus,
	}
	jsonPayload, _ := json.Marshal(payload)

	req := httptest.NewRequest("PATCH", "/api/deals/d-multi-1", bytes.NewReader(jsonPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response dealResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.UpdatedAt == "" {
		t.Error("updated_at should be set after update")
	}

	// Verify it's recent
	updatedTime, err := time.Parse(time.RFC3339, response.UpdatedAt)
	if err == nil {
		if updatedTime.Before(time.Now().Add(-5 * time.Second)) {
			t.Error("updated_at should be recent")
		}
	}
}
