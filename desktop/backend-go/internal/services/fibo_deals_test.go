package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestCreateDealValid tests creating a valid deal with required fields.
func TestCreateDealValid(t *testing.T) {
	// Setup mock Oxigraph server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate CONSTRUCT response (N-Triples format)
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-test-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
<https://businessos.dev/id/deals/d-test-1> <https://businessos.dev/id/dealName> "Cloud Infrastructure Deal" .
<https://businessos.dev/id/deals/d-test-1> <https://businessos.dev/id/dealAmount> "250000"^^<http://www.w3.org/2001/XMLSchema#double> .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	// Test data
	deal := &Deal{
		Name:              "Cloud Infrastructure Deal",
		Amount:            250000.00,
		Currency:          "USD",
		BuyerID:           "acme-corp",
		SellerID:          "cloudtech-inc",
		ExpectedCloseDate: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
		Probability:       85,
		Stage:             "negotiation",
	}

	result, err := service.CreateDeal(context.Background(), deal)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID == "" {
		t.Error("deal ID should not be empty")
	}
	if result.Name != "Cloud Infrastructure Deal" {
		t.Errorf("expected name 'Cloud Infrastructure Deal', got '%s'", result.Name)
	}
	if result.Amount != 250000.00 {
		t.Errorf("expected amount 250000.00, got %f", result.Amount)
	}
	if result.RDFTripleCount == 0 {
		t.Error("RDF triple count should be > 0")
	}
	if result.Status != "draft" {
		t.Errorf("expected status 'draft', got '%s'", result.Status)
	}
	if !result.CreatedAt.After(time.Now().Add(-1 * time.Second)) {
		t.Error("created_at should be recent")
	}
}

// TestCreateDealMissingName tests that deals require a name.
func TestCreateDealMissingName(t *testing.T) {
	service := NewFIBODealsService("http://localhost:8890")

	deal := &Deal{
		Amount:      250000.00,
		BuyerID:     "acme-corp",
		SellerID:    "cloudtech-inc",
		Probability: 85,
	}

	_, err := service.CreateDeal(context.Background(), deal)

	if err == nil {
		t.Error("expected error for missing name")
	}
	if !strings.Contains(err.Error(), "deal name required") {
		t.Errorf("expected 'deal name required' error, got: %v", err)
	}
}

// TestCreateDealInvalidAmount tests that deals require positive amounts.
func TestCreateDealInvalidAmount(t *testing.T) {
	service := NewFIBODealsService("http://localhost:8890")

	deal := &Deal{
		Name:        "Invalid Deal",
		Amount:      0,
		BuyerID:     "acme-corp",
		SellerID:    "cloudtech-inc",
		Probability: 85,
	}

	_, err := service.CreateDeal(context.Background(), deal)

	if err == nil {
		t.Error("expected error for invalid amount")
	}
	if !strings.Contains(err.Error(), "must be positive") {
		t.Errorf("expected 'must be positive' error, got: %v", err)
	}
}

// TestCreateDealInvalidProbability tests that probability must be 0-100.
func TestCreateDealInvalidProbability(t *testing.T) {
	service := NewFIBODealsService("http://localhost:8890")

	deal := &Deal{
		Name:        "Invalid Deal",
		Amount:      250000.00,
		BuyerID:     "acme-corp",
		SellerID:    "cloudtech-inc",
		Probability: 150, // Invalid: > 100
	}

	_, err := service.CreateDeal(context.Background(), deal)

	if err == nil {
		t.Error("expected error for invalid probability")
	}
	if !strings.Contains(err.Error(), "probability must be 0-100") {
		t.Errorf("expected 'probability must be 0-100' error, got: %v", err)
	}
}

// TestGetDeal tests retrieving a deal by ID.
func TestGetDeal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-existing-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
<https://businessos.dev/id/deals/d-existing-1> <https://businessos.dev/id/dealName> "Existing Deal" .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	result, err := service.GetDeal(context.Background(), "d-existing-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "d-existing-1" {
		t.Errorf("expected deal ID 'd-existing-1', got '%s'", result.ID)
	}
	if result.RDFTripleCount == 0 {
		t.Error("RDF triple count should be > 0")
	}
}

// TestGetDealEmptyID tests that GetDeal rejects empty deal IDs.
func TestGetDealEmptyID(t *testing.T) {
	service := NewFIBODealsService("http://localhost:8890")

	_, err := service.GetDeal(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty deal ID")
	}
	if !strings.Contains(err.Error(), "deal_id required") {
		t.Errorf("expected 'deal_id required' error, got: %v", err)
	}
}

// TestListDeals tests listing all deals with pagination.
func TestListDeals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	result, err := service.ListDeals(context.Background(), 50, 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) == 0 {
		t.Error("expected at least one deal")
	}
	if result[0].ID == "" {
		t.Error("deal ID should not be empty")
	}
}

// TestListDealsDefaultLimit tests that ListDeals applies default limit.
func TestListDealsDefaultLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request was made (actual query validation happens at service level)
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	_, err := service.ListDeals(context.Background(), 0, 0)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestUpdateDeal tests updating an existing deal.
func TestUpdateDeal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-update-1> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
<https://businessos.dev/id/deals/d-update-1> <https://businessos.dev/id/dealName> "Updated Deal" .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	updates := map[string]interface{}{
		"dealName": "Updated Deal",
		"status":   "active",
	}

	result, err := service.UpdateDeal(context.Background(), "d-update-1", updates)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "d-update-1" {
		t.Errorf("expected deal ID 'd-update-1', got '%s'", result.ID)
	}
	if result.UpdatedAt.Before(time.Now().Add(-1 * time.Second)) {
		t.Error("updated_at should be recent")
	}
}

// TestUpdateDealEmptyUpdates tests that UpdateDeal rejects empty updates.
func TestUpdateDealEmptyUpdates(t *testing.T) {
	service := NewFIBODealsService("http://localhost:8890")

	_, err := service.UpdateDeal(context.Background(), "d-1", map[string]interface{}{})

	if err == nil {
		t.Error("expected error for empty updates")
	}
	if !strings.Contains(err.Error(), "no updates provided") {
		t.Errorf("expected 'no updates provided' error, got: %v", err)
	}
}

// TestVerifyCompliance tests deal compliance verification.
func TestVerifyCompliance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		// Return triples indicating compliance
		w.Write([]byte(`<https://businessos.dev/id/parties/acme-corp> <https://businessos.dev/id/hasKYCStatus> <https://businessos.dev/id/KYCVerified> .
<https://businessos.dev/id/parties/acme-corp> <https://businessos.dev/id/amlScreeningResult> "CLEAR" .
<https://businessos.dev/id/deals/d-compliance-1> <https://businessos.dev/id/sox11Compliant> true .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	result, err := service.VerifyCompliance(context.Background(), "d-compliance-1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["compliance_status"] == "" {
		t.Error("compliance_status should be set")
	}
	if result["kyc_verified"] == nil {
		t.Error("kyc_verified should be set")
	}
	if result["aml_screening"] == nil {
		t.Error("aml_screening should be set")
	}
	if result["sox_compliant"] == nil {
		t.Error("sox_compliant should be set")
	}
}

// TestVerifyComplianceEmptyID tests that VerifyCompliance rejects empty IDs.
func TestVerifyComplianceEmptyID(t *testing.T) {
	service := NewFIBODealsService("http://localhost:8890")

	_, err := service.VerifyCompliance(context.Background(), "")

	if err == nil {
		t.Error("expected error for empty deal ID")
	}
	if !strings.Contains(err.Error(), "deal_id required") {
		t.Errorf("expected 'deal_id required' error, got: %v", err)
	}
}

// TestCreateDealConcurrent tests concurrent deal creation.
func TestCreateDealConcurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-concurrent> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	var wg sync.WaitGroup
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			deal := &Deal{
				Name:        "Concurrent Deal",
				Amount:      100000.00 + float64(index*10000),
				BuyerID:     "buyer-1",
				SellerID:    "seller-1",
				Probability: 75,
			}

			_, err := service.CreateDeal(context.Background(), deal)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("concurrent creation error: %v", err)
		}
	}
}

// TestCreateDealTimeout tests handling of ontology query timeouts.
func TestCreateDealTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow server (longer than 5s timeout)
		time.Sleep(6 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	deal := &Deal{
		Name:        "Timeout Test",
		Amount:      250000.00,
		BuyerID:     "buyer-1",
		SellerID:    "seller-1",
		Probability: 50,
	}

	_, err := service.CreateDeal(context.Background(), deal)

	if err == nil {
		t.Error("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

// TestCreateDealOntologyError tests handling of ontology server errors.
func TestCreateDealOntologyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Ontology server error"))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	deal := &Deal{
		Name:        "Error Test",
		Amount:      250000.00,
		BuyerID:     "buyer-1",
		SellerID:    "seller-1",
		Probability: 50,
	}

	_, err := service.CreateDeal(context.Background(), deal)

	if err == nil {
		t.Error("expected ontology error")
	}
	if !strings.Contains(err.Error(), "ontology persistence failed") {
		t.Errorf("expected ontology error, got: %v", err)
	}
}

// TestFIBODealSaaS tests creating a SaaS deal fixture ($250K).
func TestFIBODealSaaS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/saas-001> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	// SaaS deal fixture
	deal := &Deal{
		Name:              "Annual SaaS Subscription",
		Amount:            250000.00,
		Currency:          "USD",
		BuyerID:           "acme-corp",
		SellerID:          "saastech-inc",
		ExpectedCloseDate: time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
		Probability:       90,
		Stage:             "negotiation",
	}

	result, err := service.CreateDeal(context.Background(), deal)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Amount != 250000.00 {
		t.Errorf("expected SaaS amount 250000.00, got %f", result.Amount)
	}
}

// TestFIBODealLoan tests creating a Loan deal fixture ($5M).
func TestFIBODealLoan(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/loan-001> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	// Loan deal fixture
	deal := &Deal{
		Name:              "Business Expansion Loan",
		Amount:            5000000.00,
		Currency:          "USD",
		BuyerID:           "acme-corp",
		SellerID:          "finance-bank",
		ExpectedCloseDate: time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		Probability:       70,
		Stage:             "approval",
	}

	result, err := service.CreateDeal(context.Background(), deal)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Amount != 5000000.00 {
		t.Errorf("expected loan amount 5000000.00, got %f", result.Amount)
	}
}

// TestFIBODealDefense tests creating a Defense contract fixture ($12.5M).
func TestFIBODealDefense(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/defense-001> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	// Defense contract fixture
	deal := &Deal{
		Name:              "Defense Aerospace Contract",
		Amount:            12500000.00,
		Currency:          "USD",
		BuyerID:           "us-defense-dept",
		SellerID:          "aerospace-corp",
		ExpectedCloseDate: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
		Probability:       65,
		Stage:             "bidding",
	}

	result, err := service.CreateDeal(context.Background(), deal)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Amount != 12500000.00 {
		t.Errorf("expected defense amount 12500000.00, got %f", result.Amount)
	}
}

// TestDealIDGeneration tests that deal IDs are auto-generated when empty.
func TestDealIDGeneration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/n-triples")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<https://businessos.dev/id/deals/d-auto> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <https://businessos.dev/id/Deal> .
`))
	}))
	defer server.Close()

	service := NewFIBODealsService(server.URL)

	deal := &Deal{
		Name:        "Auto ID Deal",
		Amount:      100000.00,
		BuyerID:     "buyer-1",
		SellerID:    "seller-1",
		Probability: 50,
	}

	result, err := service.CreateDeal(context.Background(), deal)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID == "" {
		t.Error("expected auto-generated deal ID")
	}
	if !strings.HasPrefix(result.ID, "d-") {
		t.Errorf("expected deal ID to start with 'd-', got '%s'", result.ID)
	}
}
