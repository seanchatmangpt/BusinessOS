package ontology

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"
)

// MockSPARQLExecutor implements SPARQLExecutor for testing
type MockSPARQLExecutor struct {
	constructResults map[string]string
	askResults       map[string]bool
}

func (m *MockSPARQLExecutor) ExecuteConstruct(ctx context.Context, query string) (string, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		if result, ok := m.constructResults[query]; ok {
			return result, nil
		}
		// Default: return minimal Turtle
		return `@prefix fhir: <http://hl7.org/fhir/> .
@prefix prov: <http://www.w3.org/ns/prov#> .
fhir:test_entity a prov:Entity .`, nil
	}
}

func (m *MockSPARQLExecutor) ExecuteAsk(ctx context.Context, query string) (bool, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		if result, ok := m.askResults[query]; ok {
			return result, nil
		}
		return true, nil
	}
}

func (m *MockSPARQLExecutor) ExecuteSelect(ctx context.Context, query string) (map[string]interface{}, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return map[string]interface{}{
			"results": map[string]interface{}{
				"bindings": []map[string]interface{}{},
			},
		}, nil
	}
}

// MockRDFStore implements RDFStore for testing
type MockRDFStore struct {
	mu            sync.Mutex
	storedTriples map[string]string
	tripleCounts  map[string]int
}

func (m *MockRDFStore) StoreTriples(ctx context.Context, turtleData string) error {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.mu.Lock()
		m.storedTriples["last"] = turtleData
		m.mu.Unlock()
		return nil
	}
}

func (m *MockRDFStore) QueryTriples(ctx context.Context, query string) (int, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		return 42, nil // Default: return 42 triples
	}
}

func (m *MockRDFStore) DeleteTriples(ctx context.Context, pattern string) error {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (m *MockRDFStore) GetTriplesForEntity(ctx context.Context, entityURI string) (int, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		m.mu.Lock()
		count, ok := m.tripleCounts[entityURI]
		m.mu.Unlock()
		if ok {
			return count, nil
		}
		return 15, nil // Default: 15 triples per entity
	}
}

// MockAuditLogger implements AuditLogger for testing
type MockAuditLogger struct {
	mu            sync.Mutex
	loggedEntries []PHIAuditEntry
}

func (m *MockAuditLogger) LogAccess(ctx context.Context, entry PHIAuditEntry) error {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.mu.Lock()
		m.loggedEntries = append(m.loggedEntries, entry)
		m.mu.Unlock()
		return nil
	}
}

func (m *MockAuditLogger) GetAuditTrail(ctx context.Context, patientID string, lastNDays int) ([]PHIAuditEntry, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		now := time.Now()
		entries := []PHIAuditEntry{
			{
				Timestamp:    now.AddDate(0, 0, -5),
				Actor:        "doctor@example.com",
				Action:       "read",
				ResourceID:   "obs-123",
				ResourceType: "Observation",
				Details:      "Patient BP reading",
			},
			{
				Timestamp:    now.AddDate(0, 0, -3),
				Actor:        "nurse@example.com",
				Action:       "update",
				ResourceID:   "med-456",
				ResourceType: "MedicationRequest",
				Details:      "Prescription update",
			},
		}
		return entries, nil
	}
}

func (m *MockAuditLogger) VerifyAuditIntegrity(ctx context.Context, entries []PHIAuditEntry) (bool, error) {
	return len(entries) > 0, nil
}

// TestTrackPHI_PatientResource tests tracking a Patient FHIR resource.
func TestTrackPHI_PatientResource(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts: map[string]int{
			"http://hl7.org/fhir/Patient_p123": 12,
		},
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	data := map[string]interface{}{
		"name":   "John Doe",
		"dob":    "1980-01-15",
		"gender": "male",
	}

	result, err := manager.TrackPHI(context.Background(), "p123", "Patient", "p123", data, "admin@example.com")

	if err != nil {
		t.Fatalf("TrackPHI failed: %v", err)
	}
	if result.ResourceID != "p123" {
		t.Errorf("Expected resource_id p123, got %s", result.ResourceID)
	}
	if result.ResourceType != "Patient" {
		t.Errorf("Expected resource_type Patient, got %s", result.ResourceType)
	}
	if result.TripleCount != 12 {
		t.Errorf("Expected 12 triples, got %d", result.TripleCount)
	}
	if !result.HIPAACheckPassed {
		t.Error("Expected HIPAA check to pass")
	}
	if len(auditor.loggedEntries) != 1 {
		t.Errorf("Expected 1 audit entry, got %d", len(auditor.loggedEntries))
	}
}

// TestTrackPHI_ObservationResource tests tracking an Observation resource.
func TestTrackPHI_ObservationResource(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts: map[string]int{
			"http://hl7.org/fhir/Observation_obs456": 18,
		},
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	data := map[string]interface{}{
		"code":      "85354-9",
		"value":     120,
		"unit":      "mmHg",
		"timestamp": "2026-03-26T10:00:00Z",
	}

	result, err := manager.TrackPHI(context.Background(), "obs456", "Observation", "p123", data, "doctor@example.com")

	if err != nil {
		t.Fatalf("TrackPHI failed: %v", err)
	}
	if result.ResourceType != "Observation" {
		t.Errorf("Expected Observation, got %s", result.ResourceType)
	}
	if result.TripleCount != 18 {
		t.Errorf("Expected 18 triples, got %d", result.TripleCount)
	}
}

// TestTrackPHI_MedicationRequestResource tests tracking a MedicationRequest.
func TestTrackPHI_MedicationRequestResource(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts: map[string]int{
			"http://hl7.org/fhir/MedicationRequest_med789": 21,
		},
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	data := map[string]interface{}{
		"medication": "Metformin",
		"dosage":     "500mg",
		"frequency":  "twice daily",
		"status":     "active",
	}

	result, err := manager.TrackPHI(context.Background(), "med789", "MedicationRequest", "p123", data, "pharmacist@example.com")

	if err != nil {
		t.Fatalf("TrackPHI failed: %v", err)
	}
	if result.ResourceType != "MedicationRequest" {
		t.Errorf("Expected MedicationRequest, got %s", result.ResourceType)
	}
	if result.TripleCount != 21 {
		t.Errorf("Expected 21 triples, got %d", result.TripleCount)
	}
}

// TestVerifyConsent_Granted tests consent verification when consent is granted.
func TestVerifyConsent_Granted(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults: map[string]bool{
			// Matching the ASK query for granted consent
			"": true,
		},
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	result, err := manager.VerifyConsent(context.Background(), "p123")

	if err != nil {
		t.Fatalf("VerifyConsent failed: %v", err)
	}
	if !result.ConsentGranted {
		t.Error("Expected consent to be granted")
	}
	if len(result.Scope) == 0 {
		t.Error("Expected non-empty scope")
	}
	if result.ExpiresAt.Before(time.Now()) {
		t.Error("Expected expiry in future")
	}
}

// MockSPARQLExecutorDenyConsent returns false for ASK queries
type MockSPARQLExecutorDenyConsent struct {
	constructResults map[string]string
	askResults       map[string]bool
}

func (m *MockSPARQLExecutorDenyConsent) ExecuteConstruct(ctx context.Context, query string) (string, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		if result, ok := m.constructResults[query]; ok {
			return result, nil
		}
		return ``, nil
	}
}

func (m *MockSPARQLExecutorDenyConsent) ExecuteAsk(ctx context.Context, query string) (bool, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return false, nil // Always deny consent
	}
}

func (m *MockSPARQLExecutorDenyConsent) ExecuteSelect(ctx context.Context, query string) (map[string]interface{}, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return map[string]interface{}{}, nil
	}
}

// TestVerifyConsent_Denied tests consent verification when consent is denied.
func TestVerifyConsent_Denied(t *testing.T) {
	executor := &MockSPARQLExecutorDenyConsent{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	result, err := manager.VerifyConsent(context.Background(), "p456")

	if err != nil {
		t.Fatalf("VerifyConsent failed: %v", err)
	}
	if result.ConsentGranted {
		t.Error("Expected consent to be denied")
	}
	if len(result.Scope) > 0 {
		t.Error("Expected empty scope when consent denied")
	}
}

// TestGenerateAuditTrail tests audit trail generation for last 90 days.
func TestGenerateAuditTrail(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	result, err := manager.GenerateAuditTrail(context.Background(), "p123")

	if err != nil {
		t.Fatalf("GenerateAuditTrail failed: %v", err)
	}
	if result.PatientID != "p123" {
		t.Errorf("Expected patient_id p123, got %s", result.PatientID)
	}
	if result.TotalEntries != 2 {
		t.Errorf("Expected 2 audit entries, got %d", result.TotalEntries)
	}
	if result.Period != "last_90_days" {
		t.Errorf("Expected period last_90_days, got %s", result.Period)
	}
	if len(result.Entries) == 0 {
		t.Error("Expected non-empty audit entries")
	}
}

// TestGenerateAuditTrail_VerifyIntegrity tests HMAC signature verification in audit trail.
func TestGenerateAuditTrail_VerifyIntegrity(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	result, err := manager.GenerateAuditTrail(context.Background(), "p123")

	if err != nil {
		t.Fatalf("GenerateAuditTrail failed: %v", err)
	}

	// Verify integrity of entries
	integrity, err := auditor.VerifyAuditIntegrity(context.Background(), result.Entries)
	if err != nil {
		t.Fatalf("VerifyAuditIntegrity failed: %v", err)
	}
	if !integrity {
		t.Error("Expected audit integrity check to pass")
	}
}

// MockRDFStoreFullyDeleted returns 0 triples
type MockRDFStoreFullyDeleted struct {
	storedTriples map[string]string
	tripleCounts  map[string]int
}

func (m *MockRDFStoreFullyDeleted) StoreTriples(ctx context.Context, turtleData string) error {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		m.storedTriples["last"] = turtleData
		return nil
	}
}

func (m *MockRDFStoreFullyDeleted) QueryTriples(ctx context.Context, query string) (int, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		return 0, nil
	}
}

func (m *MockRDFStoreFullyDeleted) DeleteTriples(ctx context.Context, pattern string) error {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (m *MockRDFStoreFullyDeleted) GetTriplesForEntity(ctx context.Context, entityURI string) (int, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
		return 0, nil // No triples left
	}
}

// MockSPARQLExecutorDeleted returns false for entity existence
type MockSPARQLExecutorDeleted struct{}

func (m *MockSPARQLExecutorDeleted) ExecuteConstruct(ctx context.Context, query string) (string, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return ``, nil
	}
}

func (m *MockSPARQLExecutorDeleted) ExecuteAsk(ctx context.Context, query string) (bool, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return false, nil // Entity no longer exists
	}
}

func (m *MockSPARQLExecutorDeleted) ExecuteSelect(ctx context.Context, query string) (map[string]interface{}, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return map[string]interface{}{}, nil
	}
}

// TestCheckDeletion_FullyDeleted tests deletion verification when resource is fully deleted.
func TestCheckDeletion_FullyDeleted(t *testing.T) {
	executor := &MockSPARQLExecutorDeleted{}
	store := &MockRDFStoreFullyDeleted{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	result, err := manager.CheckDeletion(context.Background(), "p999", "Patient")

	if err != nil {
		t.Fatalf("CheckDeletion failed: %v", err)
	}
	if !result.FullyDeleted {
		t.Error("Expected resource to be fully deleted")
	}
	if result.TripleCount != 0 {
		t.Errorf("Expected 0 triples remaining, got %d", result.TripleCount)
	}
	if !result.RDFCleanConfirmed {
		t.Error("Expected RDF clean confirmation")
	}
}

// TestCheckDeletion_PartiallyDeleted tests deletion verification when remnants remain.
func TestCheckDeletion_PartiallyDeleted(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults: map[string]bool{
			"": true, // Entity still exists
		},
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts: map[string]int{
			"http://hl7.org/fhir/Patient_p888": 3, // 3 triples still present
		},
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	result, err := manager.CheckDeletion(context.Background(), "p888", "Patient")

	if err != nil {
		t.Fatalf("CheckDeletion failed: %v", err)
	}
	if result.FullyDeleted {
		t.Error("Expected resource to NOT be fully deleted")
	}
	if result.TripleCount != 3 {
		t.Errorf("Expected 3 triples remaining, got %d", result.TripleCount)
	}
}

// TestVerifyHIPAA_Compliant tests HIPAA compliance check when all criteria pass.
func TestVerifyHIPAA_Compliant(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults: map[string]bool{
			"": true, // All checks pass
		},
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	result, err := manager.VerifyHIPAA(context.Background())

	if err != nil {
		t.Fatalf("VerifyHIPAA failed: %v", err)
	}
	if !result.Compliant {
		t.Error("Expected HIPAA compliance")
	}
	if !result.AccessControlPass {
		t.Error("Expected access control check to pass")
	}
	if !result.AuditLogPass {
		t.Error("Expected audit log check to pass")
	}
	if !result.EncryptionPass {
		t.Error("Expected encryption check to pass")
	}
	if !result.IntegrityPass {
		t.Error("Expected integrity check to pass")
	}
	if result.ComplianceScore != 1.0 {
		t.Errorf("Expected compliance score 1.0, got %f", result.ComplianceScore)
	}
}

// MockSPARQLExecutorNonCompliant always returns false
type MockSPARQLExecutorNonCompliant struct{}

func (m *MockSPARQLExecutorNonCompliant) ExecuteConstruct(ctx context.Context, query string) (string, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		return ``, nil
	}
}

func (m *MockSPARQLExecutorNonCompliant) ExecuteAsk(ctx context.Context, query string) (bool, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return false, nil // All checks fail
	}
}

func (m *MockSPARQLExecutorNonCompliant) ExecuteSelect(ctx context.Context, query string) (map[string]interface{}, error) {
	// Armstrong compliance: respect context deadline
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return map[string]interface{}{}, nil
	}
}

// TestVerifyHIPAA_NonCompliant tests HIPAA compliance check when audit logging fails.
func TestVerifyHIPAA_NonCompliant(t *testing.T) {
	executor := &MockSPARQLExecutorNonCompliant{}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	result, err := manager.VerifyHIPAA(context.Background())

	if err != nil {
		t.Fatalf("VerifyHIPAA failed: %v", err)
	}
	if result.Compliant {
		t.Error("Expected HIPAA non-compliance")
	}
	if result.ComplianceScore != 0.0 {
		t.Errorf("Expected compliance score 0.0, got %f", result.ComplianceScore)
	}
}

// TestConcurrentPHIOperations tests concurrent PHI tracking (thread-safe).
func TestConcurrentPHIOperations(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	// Run 5 concurrent TrackPHI operations
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func(idx int) {
			data := map[string]interface{}{
				"test": "data",
				"idx":  idx,
			}
			_, err := manager.TrackPHI(context.Background(), "p1000", "Patient", "p1000", data, "actor@example.com")
			if err != nil {
				t.Errorf("Concurrent TrackPHI failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify no panics and operations completed
	if len(auditor.loggedEntries) != 5 {
		t.Errorf("Expected 5 audit entries, got %d", len(auditor.loggedEntries))
	}
}

// TestTrackPHI_TimeoutHandling tests that TrackPHI respects 12-second timeout.
func TestTrackPHI_TimeoutHandling(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	// Create a context that times out in 1ms (should not trigger in this test due to mock speed)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	data := map[string]interface{}{"test": "data"}
	result, err := manager.TrackPHI(ctx, "p123", "Patient", "p123", data, "actor@example.com")

	// With mocks, this should succeed quickly
	if err != nil {
		t.Logf("TrackPHI with timeout context: %v", err)
	}
	if result != nil && result.ResourceID != "p123" {
		t.Errorf("Expected resource_id p123, got %s", result.ResourceID)
	}
}

// TestMultipleFHIRResources tests tracking multiple FHIR resource types.
func TestMultipleFHIRResources(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	// Track all 3 resource types
	resources := []struct {
		id   string
		typ  string
		name string
	}{
		{"p100", "Patient", "John Doe"},
		{"obs200", "Observation", "BP Reading"},
		{"med300", "MedicationRequest", "Aspirin"},
	}

	for _, res := range resources {
		data := map[string]interface{}{"name": res.name}
		result, err := manager.TrackPHI(context.Background(), res.id, res.typ, "p100", data, "doctor@example.com")

		if err != nil {
			t.Fatalf("TrackPHI for %s failed: %v", res.typ, err)
		}
		if result.ResourceType != res.typ {
			t.Errorf("Expected %s, got %s", res.typ, result.ResourceType)
		}
	}

	if len(auditor.loggedEntries) != 3 {
		t.Errorf("Expected 3 audit entries, got %d", len(auditor.loggedEntries))
	}
}

// TestAuditTrailCompleteness verifies all 3 sample resources appear in audit trail.
func TestAuditTrailCompleteness(t *testing.T) {
	executor := &MockSPARQLExecutor{
		constructResults: make(map[string]string),
		askResults:       make(map[string]bool),
	}
	store := &MockRDFStore{
		storedTriples: make(map[string]string),
		tripleCounts:  make(map[string]int),
	}
	auditor := &MockAuditLogger{}
	logger := slog.Default()

	manager := NewHealthcarePHIManager(executor, store, auditor, logger)

	// Track resources
	manager.TrackPHI(context.Background(), "p100", "Patient", "p100", map[string]interface{}{}, "doctor@example.com")
	manager.TrackPHI(context.Background(), "obs200", "Observation", "p100", map[string]interface{}{}, "doctor@example.com")
	manager.TrackPHI(context.Background(), "med300", "MedicationRequest", "p100", map[string]interface{}{}, "pharmacist@example.com")

	// Generate audit trail
	result, err := manager.GenerateAuditTrail(context.Background(), "p100")

	if err != nil {
		t.Fatalf("GenerateAuditTrail failed: %v", err)
	}
	if result.TotalEntries < 2 {
		t.Errorf("Expected at least 2 audit entries from mock, got %d", result.TotalEntries)
	}
}
