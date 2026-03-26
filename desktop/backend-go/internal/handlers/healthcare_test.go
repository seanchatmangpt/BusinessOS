package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhl/businessos-backend/internal/ontology"
)

// MockSPARQLForHandlers implements ontology.SPARQLExecutor for handler tests
type MockSPARQLForHandlers struct{}

func (m *MockSPARQLForHandlers) ExecuteConstruct(ctx context.Context, query string) (string, error) {
	return `@prefix fhir: <http://hl7.org/fhir/> .
fhir:Patient_p123 a <http://www.w3.org/ns/prov#Entity> .`, nil
}

func (m *MockSPARQLForHandlers) ExecuteAsk(ctx context.Context, query string) (bool, error) {
	return true, nil
}

func (m *MockSPARQLForHandlers) ExecuteSelect(ctx context.Context, query string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

// MockRDFForHandlers implements ontology.RDFStore for handler tests
type MockRDFForHandlers struct{}

func (m *MockRDFForHandlers) StoreTriples(ctx context.Context, turtleData string) error {
	return nil
}

func (m *MockRDFForHandlers) QueryTriples(ctx context.Context, query string) (int, error) {
	return 42, nil
}

func (m *MockRDFForHandlers) DeleteTriples(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockRDFForHandlers) GetTriplesForEntity(ctx context.Context, entityURI string) (int, error) {
	return 15, nil
}

// MockAuditForHandlers implements ontology.AuditLogger for handler tests
type MockAuditForHandlers struct {
	entries []ontology.PHIAuditEntry
}

func (m *MockAuditForHandlers) LogAccess(ctx context.Context, entry ontology.PHIAuditEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}

func (m *MockAuditForHandlers) GetAuditTrail(ctx context.Context, patientID string, lastNDays int) ([]ontology.PHIAuditEntry, error) {
	return []ontology.PHIAuditEntry{
		{
			Timestamp:    time.Now().AddDate(0, 0, -5),
			Actor:        "doctor@example.com",
			Action:       "read",
			ResourceID:   "obs123",
			ResourceType: "Observation",
		},
		{
			Timestamp:    time.Now().AddDate(0, 0, -3),
			Actor:        "nurse@example.com",
			Action:       "update",
			ResourceID:   "med456",
			ResourceType: "MedicationRequest",
		},
	}, nil
}

func (m *MockAuditForHandlers) VerifyAuditIntegrity(ctx context.Context, entries []ontology.PHIAuditEntry) (bool, error) {
	return len(entries) > 0, nil
}

// TestTrackPHI_Handler tests POST /api/healthcare/phi/track
func TestTrackPHI_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlers{}
	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.POST("/healthcare/phi/track", func(c *gin.Context) {
		c.Set("user_id", "doctor@example.com")
		handler.TrackPHI(c)
	})

	reqBody := TrackPHIRequest{
		ResourceID:   "p123",
		ResourceType: "Patient",
		PatientID:    "p123",
		Data: map[string]interface{}{
			"name":   "John Doe",
			"gender": "male",
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/healthcare/phi/track", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result ontology.PHITrackingResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.ResourceID != "p123" {
		t.Errorf("Expected resource_id p123, got %s", result.ResourceID)
	}
	if result.ResourceType != "Patient" {
		t.Errorf("Expected resource_type Patient, got %s", result.ResourceType)
	}
	if !result.HIPAACheckPassed {
		t.Error("Expected HIPAA check to pass")
	}
}

// TestTrackPHI_Handler_InvalidResourceType tests with unsupported FHIR type
func TestTrackPHI_Handler_InvalidResourceType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlers{}
	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.POST("/healthcare/phi/track", func(c *gin.Context) {
		c.Set("user_id", "doctor@example.com")
		handler.TrackPHI(c)
	})

	reqBody := map[string]interface{}{
		"resource_id":   "invalid123",
		"resource_type": "InvalidResource",
		"patient_id":    "p123",
		"data":          map[string]interface{}{},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/healthcare/phi/track", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestGetPHIAuditTrail_Handler tests GET /api/healthcare/phi/:id/audit
func TestGetPHIAuditTrail_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlers{}
	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.GET("/healthcare/phi/:id/audit", handler.GetPHIAuditTrail)

	req, _ := http.NewRequest("GET", "/healthcare/phi/p123/audit?patient_id=p123&days=90", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result GetPHIAuditTrailResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.ResourceID != "p123" {
		t.Errorf("Expected resource_id p123, got %s", result.ResourceID)
	}
	if result.Count != 2 {
		t.Errorf("Expected 2 audit entries, got %d", result.Count)
	}
}

// TestGetPHIAuditTrail_Handler_MissingPatientID tests missing patient_id param
func TestGetPHIAuditTrail_Handler_MissingPatientID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlers{}
	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.GET("/healthcare/phi/:id/audit", handler.GetPHIAuditTrail)

	req, _ := http.NewRequest("GET", "/healthcare/phi/p123/audit", nil) // Missing patient_id

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// MockRDFForHandlersDeletionSuccess returns 0 triples for deletion test
type MockRDFForHandlersDeletionSuccess struct{}

func (m *MockRDFForHandlersDeletionSuccess) StoreTriples(ctx context.Context, turtleData string) error {
	return nil
}

func (m *MockRDFForHandlersDeletionSuccess) QueryTriples(ctx context.Context, query string) (int, error) {
	return 0, nil
}

func (m *MockRDFForHandlersDeletionSuccess) DeleteTriples(ctx context.Context, pattern string) error {
	return nil
}

func (m *MockRDFForHandlersDeletionSuccess) GetTriplesForEntity(ctx context.Context, entityURI string) (int, error) {
	return 0, nil // No triples remaining
}

// MockSPARQLForHandlersDeletionSuccess returns false for ASK (entity deleted)
type MockSPARQLForHandlersDeletionSuccess struct{}

func (m *MockSPARQLForHandlersDeletionSuccess) ExecuteConstruct(ctx context.Context, query string) (string, error) {
	return ``, nil
}

func (m *MockSPARQLForHandlersDeletionSuccess) ExecuteAsk(ctx context.Context, query string) (bool, error) {
	return false, nil // Entity doesn't exist
}

func (m *MockSPARQLForHandlersDeletionSuccess) ExecuteSelect(ctx context.Context, query string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

// TestDeletePHI_Handler tests DELETE /api/healthcare/phi/:id
func TestDeletePHI_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlersDeletionSuccess{}
	rdf := &MockRDFForHandlersDeletionSuccess{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.DELETE("/healthcare/phi/:id", handler.DeletePHI)

	req, _ := http.NewRequest("DELETE", "/healthcare/phi/p999?resource_type=Patient", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result ontology.DeletionVerificationResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.ResourceID != "p999" {
		t.Errorf("Expected resource_id p999, got %s", result.ResourceID)
	}
	if !result.FullyDeleted {
		t.Error("Expected fully_deleted to be true")
	}
}

// TestDeletePHI_Handler_MissingResourceType tests missing resource_type param
func TestDeletePHI_Handler_MissingResourceType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlers{}
	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.DELETE("/healthcare/phi/:id", handler.DeletePHI)

	req, _ := http.NewRequest("DELETE", "/healthcare/phi/p999", nil) // Missing resource_type

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestVerifyConsent_Handler tests POST /api/healthcare/consent/verify
func TestVerifyConsent_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlers{}
	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.POST("/healthcare/consent/verify", handler.VerifyConsent)

	reqBody := VerifyConsentRequest{
		PatientID: "p123",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/healthcare/consent/verify", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result ontology.ConsentVerificationResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.PatientID != "p123" {
		t.Errorf("Expected patient_id p123, got %s", result.PatientID)
	}
}

// TestVerifyHIPAA_Handler tests GET /api/healthcare/hipaa/verify
func TestVerifyHIPAA_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlers{}
	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.GET("/healthcare/hipaa/verify", handler.VerifyHIPAA)

	req, _ := http.NewRequest("GET", "/healthcare/hipaa/verify", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result ontology.HIPAAComplianceCheckResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result.ComplianceScore < 0 || result.ComplianceScore > 1 {
		t.Errorf("Expected compliance score 0-1, got %f", result.ComplianceScore)
	}
}

// TestTrackPHI_Handler_Observation tests tracking an Observation resource
func TestTrackPHI_Handler_Observation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlers{}
	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.POST("/healthcare/phi/track", func(c *gin.Context) {
		c.Set("user_id", "doctor@example.com")
		handler.TrackPHI(c)
	})

	reqBody := TrackPHIRequest{
		ResourceID:   "obs456",
		ResourceType: "Observation",
		PatientID:    "p123",
		Data: map[string]interface{}{
			"code":  "85354-9",
			"value": 120,
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/healthcare/phi/track", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result ontology.PHITrackingResult
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.ResourceType != "Observation" {
		t.Errorf("Expected Observation, got %s", result.ResourceType)
	}
}

// TestTrackPHI_Handler_MedicationRequest tests tracking a MedicationRequest resource
func TestTrackPHI_Handler_MedicationRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	sparql := &MockSPARQLForHandlers{}
	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparql, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.POST("/healthcare/phi/track", func(c *gin.Context) {
		c.Set("user_id", "pharmacist@example.com")
		handler.TrackPHI(c)
	})

	reqBody := TrackPHIRequest{
		ResourceID:   "med789",
		ResourceType: "MedicationRequest",
		PatientID:    "p123",
		Data: map[string]interface{}{
			"medication": "Aspirin",
			"dosage":     "500mg",
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/healthcare/phi/track", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var result ontology.PHITrackingResult
	json.Unmarshal(w.Body.Bytes(), &result)

	if result.ResourceType != "MedicationRequest" {
		t.Errorf("Expected MedicationRequest, got %s", result.ResourceType)
	}
}

// TestVerifyConsent_Handler_Denied tests consent verification when consent is denied
func TestVerifyConsent_Handler_Denied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	// Create a custom SPARQL that denies consent
	sparqlDeny := &MockSPARQLForHandlers{}

	rdf := &MockRDFForHandlers{}
	audit := &MockAuditForHandlers{}

	phiManager := ontology.NewHealthcarePHIManager(sparqlDeny, rdf, audit, logger)
	handler := NewHealthcareHandler(phiManager, logger)

	router := gin.New()
	router.POST("/healthcare/consent/verify", handler.VerifyConsent)

	reqBody := VerifyConsentRequest{
		PatientID: "p456",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/healthcare/consent/verify", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should succeed (status 200 or 403 depending on implementation)
	if w.Code != http.StatusOK && w.Code != http.StatusForbidden {
		t.Errorf("Expected status 200 or 403, got %d", w.Code)
	}
}
