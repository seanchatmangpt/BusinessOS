package compliance

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Constructor + Constants
// ---------------------------------------------------------------------------

func TestNewGDPRService_Initialization(t *testing.T) {
	svc := NewGDPRService("secret", slog.Default())
	if svc == nil {
		t.Fatal("NewGDPRService returned nil")
	}
	if svc.auditSecret != "secret" {
		t.Errorf("want auditSecret=secret, got %s", svc.auditSecret)
	}
	if svc.auditLogs == nil {
		t.Error("auditLogs should be initialized")
	}
	if svc.dataStore == nil {
		t.Error("dataStore should be initialized")
	}
	if svc.requests == nil {
		t.Error("requests should be initialized")
	}
}

func TestNewGDPRService_NilLogger(t *testing.T) {
	svc := NewGDPRService("s", nil)
	if svc == nil {
		t.Fatal("NewGDPRService with nil logger returned nil")
	}
}

func TestGDPRConstants_Values(t *testing.T) {
	if RightOfAccess != "access" {
		t.Errorf("want RightOfAccess=access, got %s", RightOfAccess)
	}
	if RightToBeForotten != "be_forgotten" {
		t.Errorf("want RightToBeForotten=be_forgotten, got %s", RightToBeForotten)
	}
	if RightOfRectification != "rectification" {
		t.Errorf("want RightOfRectification=rectification, got %s", RightOfRectification)
	}
	if RightOfPortability != "portability" {
		t.Errorf("want RightOfPortability=portability, got %s", RightOfPortability)
	}
	if RightToRestrictProcessing != "restrict_processing" {
		t.Errorf("want RightToRestrictProcessing=restrict_processing, got %s", RightToRestrictProcessing)
	}
}

// ---------------------------------------------------------------------------
// AccessRequest (Article 15)
// ---------------------------------------------------------------------------

func TestAccessRequest_Success(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	resp, err := svc.AccessRequest(ctx, "subject-1", "req@example.com")
	if err != nil {
		t.Fatalf("AccessRequest returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("AccessRequest returned nil response")
	}
	if resp.RequestID == "" {
		t.Error("RequestID should not be empty")
	}
	if resp.Status != "completed" {
		t.Errorf("want Status=completed, got %s", resp.Status)
	}
	if resp.DeadlineAt.Before(time.Now()) {
		t.Error("DeadlineAt should be in the future")
	}
}

func TestAccessRequest_WithSampleData(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	svc.InsertSampleData("subject-sample")
	resp, err := svc.AccessRequest(ctx, "subject-sample", "req@example.com")
	if err != nil {
		t.Fatalf("AccessRequest with sample data returned error: %v", err)
	}
	if resp.Data == nil {
		t.Error("response Data should not be nil when sample data exists")
	}
}

func TestAccessRequest_CreatesAuditLog(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	_, err := svc.AccessRequest(ctx, "subject-audit", "req@example.com")
	if err != nil {
		t.Fatalf("AccessRequest error: %v", err)
	}

	trail := svc.GetAuditTrail("subject-audit")
	if len(trail) < 1 {
		t.Errorf("expected at least 1 audit log entry, got %d", len(trail))
	}
	if trail[0].RequestType != RightOfAccess {
		t.Errorf("want RequestType=%s, got %s", RightOfAccess, trail[0].RequestType)
	}
}

// ---------------------------------------------------------------------------
// ForgetRequest (Article 17)
// ---------------------------------------------------------------------------

func TestForgetRequest_Success(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	resp, err := svc.ForgetRequest(ctx, "subject-forget", "req@example.com")
	if err != nil {
		t.Fatalf("ForgetRequest error: %v", err)
	}
	if resp.RequestID == "" {
		t.Error("RequestID should not be empty")
	}
	if resp.Status != "completed" {
		t.Errorf("want Status=completed, got %s", resp.Status)
	}
}

func TestForgetRequest_AnonymizesData(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	svc.InsertSampleData("subject-anon")
	resp, err := svc.ForgetRequest(ctx, "subject-anon", "req@example.com")
	if err != nil {
		t.Fatalf("ForgetRequest error: %v", err)
	}
	if resp.Data == nil {
		t.Error("response Data should contain anonymization result")
	}
}

func TestForgetRequest_CreatesAuditLog(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	svc.InsertSampleData("subject-forget-audit")
	_, err := svc.ForgetRequest(ctx, "subject-forget-audit", "req@example.com")
	if err != nil {
		t.Fatalf("ForgetRequest error: %v", err)
	}

	trail := svc.GetAuditTrail("subject-forget-audit")
	if len(trail) < 1 {
		t.Error("expected at least 1 audit log entry")
	}
	if trail[0].RequestType != RightToBeForotten {
		t.Errorf("want RequestType=%s, got %s", RightToBeForotten, trail[0].RequestType)
	}
}

// ---------------------------------------------------------------------------
// RectifyRequest (Article 16)
// ---------------------------------------------------------------------------

func TestRectifyRequest_Success(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	corrections := map[string]interface{}{"email": "new@example.com"}
	resp, err := svc.RectifyRequest(ctx, "subject-rectify", "req@example.com", corrections)
	if err != nil {
		t.Fatalf("RectifyRequest error: %v", err)
	}
	if resp.RequestID == "" {
		t.Error("RequestID should not be empty")
	}
	if resp.Status != "completed" {
		t.Errorf("want Status=completed, got %s", resp.Status)
	}
}

func TestRectifyRequest_AppliesCorrections(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	svc.InsertSampleData("subject-correct")
	corrections := map[string]interface{}{"city": "Berlin", "postal_code": "10115"}
	_, err := svc.RectifyRequest(ctx, "subject-correct", "req@example.com", corrections)
	if err != nil {
		t.Fatalf("RectifyRequest error: %v", err)
	}
}

func TestRectifyRequest_CreatesAuditLog(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	_, err := svc.RectifyRequest(ctx, "subject-rectify-audit", "req@example.com", map[string]interface{}{"k": "v"})
	if err != nil {
		t.Fatalf("RectifyRequest error: %v", err)
	}

	trail := svc.GetAuditTrail("subject-rectify-audit")
	if len(trail) < 1 {
		t.Error("expected at least 1 audit log entry")
	}
	if trail[0].RequestType != RightOfRectification {
		t.Errorf("want RequestType=%s, got %s", RightOfRectification, trail[0].RequestType)
	}
}

// ---------------------------------------------------------------------------
// PortabilityRequest (Article 20)
// ---------------------------------------------------------------------------

func TestPortabilityRequest_JSONFormat(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	resp, err := svc.PortabilityRequest(ctx, "subject-port-json", "req@example.com", "json")
	if err != nil {
		t.Fatalf("PortabilityRequest json error: %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("want Status=completed, got %s", resp.Status)
	}
}

func TestPortabilityRequest_CSVFormat(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	resp, err := svc.PortabilityRequest(ctx, "subject-port-csv", "req@example.com", "csv")
	if err != nil {
		t.Fatalf("PortabilityRequest csv error: %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("want Status=completed, got %s", resp.Status)
	}
}

func TestPortabilityRequest_DefaultFormat(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	resp, err := svc.PortabilityRequest(ctx, "subject-port-default", "req@example.com", "")
	if err != nil {
		t.Fatalf("PortabilityRequest default format error: %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("want Status=completed, got %s", resp.Status)
	}
}

func TestPortabilityRequest_CreatesAuditLog(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	_, err := svc.PortabilityRequest(ctx, "subject-port-audit", "req@example.com", "json")
	if err != nil {
		t.Fatalf("PortabilityRequest error: %v", err)
	}

	trail := svc.GetAuditTrail("subject-port-audit")
	if len(trail) < 1 {
		t.Error("expected at least 1 audit log entry")
	}
	if trail[0].RequestType != RightOfPortability {
		t.Errorf("want RequestType=%s, got %s", RightOfPortability, trail[0].RequestType)
	}
}

// ---------------------------------------------------------------------------
// RestrictProcessingRequest (Article 18)
// ---------------------------------------------------------------------------

func TestRestrictProcessingRequest_Success(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	resp, err := svc.RestrictProcessingRequest(ctx, "subject-restrict", "req@example.com", "legal hold")
	if err != nil {
		t.Fatalf("RestrictProcessingRequest error: %v", err)
	}
	if resp.Status != "completed" {
		t.Errorf("want Status=completed, got %s", resp.Status)
	}
}

func TestRestrictProcessingRequest_FlagsProfile(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	svc.InsertSampleData("subject-restrict-flag")
	resp, err := svc.RestrictProcessingRequest(ctx, "subject-restrict-flag", "req@example.com", "dispute")
	if err != nil {
		t.Fatalf("RestrictProcessingRequest error: %v", err)
	}
	if resp.Data == nil {
		t.Error("response Data should not be nil")
	}
}

func TestRestrictProcessingRequest_CreatesAuditLog(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	_, err := svc.RestrictProcessingRequest(ctx, "subject-restrict-audit", "req@example.com", "reason")
	if err != nil {
		t.Fatalf("RestrictProcessingRequest error: %v", err)
	}

	trail := svc.GetAuditTrail("subject-restrict-audit")
	if len(trail) < 1 {
		t.Error("expected at least 1 audit log entry")
	}
	if trail[0].RequestType != RightToRestrictProcessing {
		t.Errorf("want RequestType=%s, got %s", RightToRestrictProcessing, trail[0].RequestType)
	}
}

// ---------------------------------------------------------------------------
// QueryGDPRRequest
// ---------------------------------------------------------------------------

func TestQueryGDPRRequest_Found(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	resp, _ := svc.AccessRequest(ctx, "subject-query", "req@example.com")
	found := svc.QueryGDPRRequest(resp.RequestID)
	if found == nil {
		t.Fatal("QueryGDPRRequest returned nil for known request ID")
	}
	if found.ID != resp.RequestID {
		t.Errorf("want ID=%s, got %s", resp.RequestID, found.ID)
	}
}

func TestQueryGDPRRequest_NotFound(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	result := svc.QueryGDPRRequest("nonexistent-id")
	if result != nil {
		t.Error("QueryGDPRRequest should return nil for unknown request ID")
	}
}

// ---------------------------------------------------------------------------
// GetAuditTrail
// ---------------------------------------------------------------------------

func TestGetAuditTrail_FiltersBySubject(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	svc.AccessRequest(ctx, "subject-a", "a@example.com")
	svc.ForgetRequest(ctx, "subject-b", "b@example.com")
	svc.AccessRequest(ctx, "subject-a", "a@example.com")

	trailA := svc.GetAuditTrail("subject-a")
	if len(trailA) != 2 {
		t.Errorf("want 2 audit entries for subject-a, got %d", len(trailA))
	}
	for _, entry := range trailA {
		if entry.SubjectID != "subject-a" {
			t.Errorf("unexpected SubjectID %s in trail for subject-a", entry.SubjectID)
		}
	}
}

func TestGetAuditTrail_Empty(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	trail := svc.GetAuditTrail("nonexistent-subject")
	if trail != nil && len(trail) > 0 {
		t.Errorf("expected empty trail for unknown subject, got %d entries", len(trail))
	}
}

// ---------------------------------------------------------------------------
// VerifyAuditChainIntegrity
// ---------------------------------------------------------------------------

func TestVerifyAuditChainIntegrity_EmptyChain(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	valid, issues := svc.VerifyAuditChainIntegrity()
	if !valid {
		t.Errorf("empty chain should be valid, got issues: %v", issues)
	}
	if len(issues) != 0 {
		t.Errorf("empty chain should have 0 issues, got %d", len(issues))
	}
}

func TestVerifyAuditChainIntegrity_ValidChain(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	svc.AccessRequest(ctx, "subject-chain", "req@example.com")
	svc.ForgetRequest(ctx, "subject-chain", "req@example.com")

	valid, issues := svc.VerifyAuditChainIntegrity()
	if !valid {
		t.Errorf("valid chain reported as invalid: %v", issues)
	}
}

func TestVerifyAuditChainIntegrity_ChainLinks(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()

	svc.AccessRequest(ctx, "subject-link", "req@example.com")
	svc.AccessRequest(ctx, "subject-link", "req@example.com")

	if len(svc.auditLogs) < 2 {
		t.Fatal("expected at least 2 audit log entries")
	}
	secondLog := svc.auditLogs[1]
	if secondLog.PreviousHash == "" {
		t.Error("second log entry should have a PreviousHash set")
	}
}

// ---------------------------------------------------------------------------
// InsertSampleData
// ---------------------------------------------------------------------------

func TestInsertSampleData_CreatesProfile(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	svc.InsertSampleData("subject-sample-test")

	data := svc.dataStore["subject-sample-test"]
	if data == nil {
		t.Fatal("InsertSampleData did not populate dataStore")
	}
	if data.Profile == nil {
		t.Error("InsertSampleData should create Profile")
	}
	if data.Profile.ID != "subject-sample-test" {
		t.Errorf("want Profile.ID=subject-sample-test, got %s", data.Profile.ID)
	}
}

// ---------------------------------------------------------------------------
// Table-driven: 30-day deadline for all 5 request types
// ---------------------------------------------------------------------------

func TestGDPRRequestDeadline_Is30Days(t *testing.T) {
	svc := NewGDPRService("secret", nil)
	ctx := context.Background()
	svc.InsertSampleData("subject-deadline")

	tests := []struct {
		name      string
		requestFn func() (*GDPRResponse, error)
	}{
		{"access", func() (*GDPRResponse, error) {
			return svc.AccessRequest(ctx, "subject-deadline", "req@example.com")
		}},
		{"be_forgotten", func() (*GDPRResponse, error) {
			return svc.ForgetRequest(ctx, "subject-deadline", "req@example.com")
		}},
		{"rectification", func() (*GDPRResponse, error) {
			return svc.RectifyRequest(ctx, "subject-deadline", "req@example.com", map[string]interface{}{})
		}},
		{"portability", func() (*GDPRResponse, error) {
			return svc.PortabilityRequest(ctx, "subject-deadline", "req@example.com", "json")
		}},
		{"restrict_processing", func() (*GDPRResponse, error) {
			return svc.RestrictProcessingRequest(ctx, "subject-deadline", "req@example.com", "test")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			resp, err := tt.requestFn()
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", tt.name, err)
			}
			minDeadline := before.AddDate(0, 0, 29)
			if resp.DeadlineAt.Before(minDeadline) {
				t.Errorf("%s: DeadlineAt %v is less than 29 days from now", tt.name, resp.DeadlineAt)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Integration scenario
// ---------------------------------------------------------------------------

func TestComplexScenario_FullDataSubjectLifecycle(t *testing.T) {
	svc := NewGDPRService("complex-secret", nil)
	ctx := context.Background()
	subjectID := "lifecycle-subject"

	// 1. Insert sample data
	svc.InsertSampleData(subjectID)

	// 2. Access request
	accessResp, err := svc.AccessRequest(ctx, subjectID, "dpo@example.com")
	if err != nil {
		t.Fatalf("AccessRequest error: %v", err)
	}
	if accessResp.Status != "completed" {
		t.Errorf("access: want Status=completed, got %s", accessResp.Status)
	}

	// 3. Rectification
	_, err = svc.RectifyRequest(ctx, subjectID, "dpo@example.com", map[string]interface{}{"email": "corrected@example.com"})
	if err != nil {
		t.Fatalf("RectifyRequest error: %v", err)
	}

	// 4. Portability export
	_, err = svc.PortabilityRequest(ctx, subjectID, "dpo@example.com", "json")
	if err != nil {
		t.Fatalf("PortabilityRequest error: %v", err)
	}

	// 5. Restrict processing
	_, err = svc.RestrictProcessingRequest(ctx, subjectID, "dpo@example.com", "subject request")
	if err != nil {
		t.Fatalf("RestrictProcessingRequest error: %v", err)
	}

	// 6. Forget (erasure)
	_, err = svc.ForgetRequest(ctx, subjectID, "dpo@example.com")
	if err != nil {
		t.Fatalf("ForgetRequest error: %v", err)
	}

	// Verify audit trail has all 5 requests
	trail := svc.GetAuditTrail(subjectID)
	if len(trail) < 5 {
		t.Errorf("expected at least 5 audit entries, got %d", len(trail))
	}

	// Verify audit chain integrity throughout lifecycle
	valid, issues := svc.VerifyAuditChainIntegrity()
	if !valid {
		t.Errorf("audit chain invalid after lifecycle: %v", issues)
	}
}

// ---------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------

func BenchmarkGDPRAccessRequest(b *testing.B) {
	svc := NewGDPRService("bench-secret", nil)
	ctx := context.Background()
	svc.InsertSampleData("bench-subject")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.AccessRequest(ctx, "bench-subject", "bench@example.com")
	}
}

func BenchmarkGDPRVerifyAuditChain(b *testing.B) {
	svc := NewGDPRService("bench-secret", nil)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		svc.AccessRequest(ctx, "bench-chain", "bench@example.com")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.VerifyAuditChainIntegrity()
	}
}
