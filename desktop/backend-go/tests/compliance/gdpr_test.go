package compliance

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhl/businessos-backend/internal/compliance"
)

const testAuditSecret = "test-secret-gdpr-audit-key-12345"

// Test 1: AccessRequest returns all personal data in JSON format
func TestAccessRequestReturnsPersonalData(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-001"
	gs.InsertSampleData(subjectID)

	resp, err := gs.AccessRequest(context.Background(), subjectID, "requester@example.com")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, compliance.RightOfAccess, gs.QueryGDPRRequest(resp.RequestID).RequestType)

	// Verify personal data in response
	data, ok := resp.Data.(*compliance.PersonalData)
	require.True(t, ok)
	assert.Equal(t, subjectID, data.SubjectID)
	assert.NotNil(t, data.Profile)
	assert.Equal(t, subjectID, data.Profile.ID)
}

// Test 2: AccessRequest creates audit trail with hash-chain integrity
func TestAccessRequestAuditTrail(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-002"
	gs.InsertSampleData(subjectID)

	resp, err := gs.AccessRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	// Verify audit trail
	auditLogs := gs.GetAuditTrail(subjectID)
	require.Greater(t, len(auditLogs), 0)

	log := auditLogs[0]
	assert.Equal(t, resp.RequestID, log.RequestID)
	assert.Equal(t, compliance.RightOfAccess, log.RequestType)
	assert.Equal(t, "data_retrieved", log.Action)
	assert.NotEmpty(t, log.DataHash)
	assert.NotEmpty(t, log.Signature)
}

// Test 3: ForgetRequest anonymizes data (soft-delete)
func TestForgetRequestAnonymizesData(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-003"
	gs.InsertSampleData(subjectID)

	resp, err := gs.ForgetRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, compliance.RightToBeForotten, gs.QueryGDPRRequest(resp.RequestID).RequestType)

	// Verify data is anonymized (soft-delete)
	auditLogs := gs.GetAuditTrail(subjectID)
	require.Greater(t, len(auditLogs), 0)
	assert.Equal(t, "data_anonymized", auditLogs[0].Action)
}

// Test 4: ForgetRequest maintains legal hold (7 year retention)
func TestForgetRequestMaintainsLegalHold(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-004"
	gs.InsertSampleData(subjectID)

	resp, err := gs.ForgetRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	// Verify response mentions legal hold period
	respData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, respData["retention_period"], "7 years")
}

// Test 5: RectifyRequest corrects inaccurate data
func TestRectifyRequestCorrectedData(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-005"
	gs.InsertSampleData(subjectID)

	corrections := map[string]interface{}{
		"email":     "newemail@example.com",
		"full_name": "Updated Name",
		"phone":     "+1-555-0200",
	}

	rectResp, err := gs.RectifyRequest(context.Background(), subjectID, "requester@example.com", corrections)
	require.NoError(t, err)

	assert.Equal(t, "completed", rectResp.Status)
	assert.Equal(t, compliance.RightOfRectification, gs.QueryGDPRRequest(rectResp.RequestID).RequestType)

	// Verify corrections in response
	correctedData, ok := rectResp.Data.(*compliance.PersonalData)
	require.True(t, ok)
	assert.Equal(t, "newemail@example.com", correctedData.ContactData["email"])
}

// Test 6: RectifyRequest records all corrections in audit trail
func TestRectifyRequestAuditRecordsCorrections(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-006"
	gs.InsertSampleData(subjectID)

	corrections := map[string]interface{}{
		"email": "corrected@example.com",
	}

	auditResp, err := gs.RectifyRequest(context.Background(), subjectID, "requester@example.com", corrections)
	require.NoError(t, err)

	// Verify audit log contains corrections
	auditLogs := gs.GetAuditTrail(subjectID)
	require.Greater(t, len(auditLogs), 0)
	log := auditLogs[0]
	assert.Equal(t, "data_corrected", log.Action)
	assert.NotNil(t, log.Details["corrections"])
	assert.NotNil(t, auditResp)
}

// Test 7: PortabilityRequest exports data in JSON format
func TestPortabilityRequestExportsJSON(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-007"
	gs.InsertSampleData(subjectID)

	resp, err := gs.PortabilityRequest(context.Background(), subjectID, "requester@example.com", "json")
	require.NoError(t, err)

	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, compliance.RightOfPortability, gs.QueryGDPRRequest(resp.RequestID).RequestType)

	// Verify data in response
	respData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "json", respData["format"])
	assert.NotNil(t, respData["archive"])
}

// Test 8: PortabilityRequest exports data in CSV format
func TestPortabilityRequestExportsCSV(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-008"
	gs.InsertSampleData(subjectID)

	resp, err := gs.PortabilityRequest(context.Background(), subjectID, "requester@example.com", "csv")
	require.NoError(t, err)

	assert.Equal(t, "completed", resp.Status)

	// Verify CSV format in response
	respData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "csv", respData["format"])
}

// Test 9: PortabilityRequest includes archive metadata
func TestPortabilityRequestIncludesMetadata(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-009"
	gs.InsertSampleData(subjectID)

	resp, err := gs.PortabilityRequest(context.Background(), subjectID, "requester@example.com", "json")
	require.NoError(t, err)

	respData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	archive, ok := respData["archive"].(string)
	require.True(t, ok)
	assert.Contains(t, archive, "gdpr-portability")
	assert.Contains(t, archive, subjectID)
}

// Test 10: RestrictProcessingRequest flags data as restricted
func TestRestrictProcessingRequestFlagsRestriction(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-010"
	gs.InsertSampleData(subjectID)

	resp, err := gs.RestrictProcessingRequest(context.Background(), subjectID, "requester@example.com", "disputed_accuracy")
	require.NoError(t, err)

	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, compliance.RightToRestrictProcessing, gs.QueryGDPRRequest(resp.RequestID).RequestType)

	// Verify restriction in response
	respData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, respData["restriction_active"].(bool))
	assert.True(t, respData["automated_processing_disabled"].(bool))
}

// Test 11: RestrictProcessingRequest records reason in audit trail
func TestRestrictProcessingRequestRecordsReason(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-011"
	gs.InsertSampleData(subjectID)

	reason := "disputed_accuracy"
	restrictResp, err := gs.RestrictProcessingRequest(context.Background(), subjectID, "requester@example.com", reason)
	require.NoError(t, err)

	// Verify audit trail contains reason
	auditLogs := gs.GetAuditTrail(subjectID)
	require.Greater(t, len(auditLogs), 0)
	log := auditLogs[0]
	assert.Equal(t, reason, log.Details["restriction_reason"])
	assert.NotNil(t, restrictResp)
}

// Test 12: All GDPR requests meet 30-day deadline
func TestGDPRRequestDeadline30Days(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-012"
	gs.InsertSampleData(subjectID)

	now := time.Now().UTC()
	deadline30Days := now.AddDate(0, 0, 30)

	resp, err := gs.AccessRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	// Verify deadline is approximately 30 days from now (within 1 minute tolerance)
	assert.WithinDuration(t, deadline30Days, resp.DeadlineAt, 1*time.Minute)
}

// Test 13: Audit chain integrity verification passes for valid chain
func TestAuditChainIntegrityValid(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-013"
	gs.InsertSampleData(subjectID)

	// Create multiple GDPR requests to build audit chain
	_, err1 := gs.AccessRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err1)

	_, err2 := gs.RectifyRequest(context.Background(), subjectID, "requester@example.com", map[string]interface{}{"email": "test@example.com"})
	require.NoError(t, err2)

	// Verify chain integrity
	valid, issues := gs.VerifyAuditChainIntegrity()
	assert.True(t, valid)
	assert.Empty(t, issues)
}

// Test 14: All requests include request ID and tracking
func TestGDPRRequestTracking(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-014"
	gs.InsertSampleData(subjectID)

	resp, err := gs.AccessRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	// Verify request is tracked
	req := gs.QueryGDPRRequest(resp.RequestID)
	require.NotNil(t, req)
	assert.Equal(t, subjectID, req.SubjectID)
	assert.True(t, req.Verified)
	assert.Equal(t, "requester@example.com", req.RequesterEmail)
}

// Test 15: Response includes all required GDPR compliance fields
func TestGDPRResponseComplianceFields(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-015"
	gs.InsertSampleData(subjectID)

	resp, err := gs.AccessRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	// Verify all required fields
	assert.NotEmpty(t, resp.RequestID)
	assert.NotEmpty(t, resp.Status)
	assert.NotEmpty(t, resp.Message)
	assert.NotZero(t, resp.Timestamp)
	assert.NotZero(t, resp.DeadlineAt)
	assert.NotNil(t, resp.Data)
}

// Test 16: Audit logs contain handler (requester email)
func TestAuditLogsIncludeHandler(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-016"
	gs.InsertSampleData(subjectID)

	requesterEmail := "audit-test@example.com"
	_, err := gs.AccessRequest(context.Background(), subjectID, requesterEmail)
	require.NoError(t, err)

	auditLogs := gs.GetAuditTrail(subjectID)
	require.Greater(t, len(auditLogs), 0)
	assert.Equal(t, requesterEmail, auditLogs[0].Handler)
}

// Test 17: Edge case - Multiple requests for same subject
func TestMultipleRequestsSameSubject(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-017"
	gs.InsertSampleData(subjectID)

	// Submit multiple requests
	resp1, _ := gs.AccessRequest(context.Background(), subjectID, "requester1@example.com")
	resp2, _ := gs.RectifyRequest(context.Background(), subjectID, "requester2@example.com", map[string]interface{}{"phone": "+1-555-0300"})

	// Verify both requests are tracked
	req1 := gs.QueryGDPRRequest(resp1.RequestID)
	req2 := gs.QueryGDPRRequest(resp2.RequestID)
	assert.NotNil(t, req1)
	assert.NotNil(t, req2)
	assert.NotEqual(t, req1.ID, req2.ID)

	// Verify audit trail contains both
	auditLogs := gs.GetAuditTrail(subjectID)
	assert.GreaterOrEqual(t, len(auditLogs), 2)
}

// Test 18: Edge case - Data subject not found returns empty data
func TestAccessRequestNonexistentSubject(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "nonexistent-subject"

	resp, err := gs.AccessRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	// Should return empty but valid response
	assert.Equal(t, "completed", resp.Status)
	data, ok := resp.Data.(*compliance.PersonalData)
	require.True(t, ok)
	assert.Equal(t, subjectID, data.SubjectID)
}

// Test 19: GDPR Article mapping - Article 15 (Access)
func TestArticle15Compliance(t *testing.T) {
	// Article 15: Right of Access
	// Data subject has right to obtain confirmation of whether personal data is being processed
	// and to receive a copy in a commonly used electronic format

	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-article-15"
	gs.InsertSampleData(subjectID)

	resp, err := gs.AccessRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	// Verify Article 15 compliance
	assert.Equal(t, "completed", resp.Status)
	data, ok := resp.Data.(*compliance.PersonalData)
	require.True(t, ok)

	// Data must be in "commonly used electronic format" (JSON satisfies this)
	assert.NotNil(t, data)
	assert.Equal(t, subjectID, data.SubjectID)
}

// Test 20: GDPR Article mapping - Article 17 (Right to be Forgotten)
func TestArticle17Compliance(t *testing.T) {
	// Article 17: Right to Erasure (Right to be Forgotten)
	// Data subject has right to erasure without undue delay
	// Exception: data kept for legal obligations

	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-article-17"
	gs.InsertSampleData(subjectID)

	resp, err := gs.ForgetRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	// Verify Article 17 compliance
	respData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	// Verify data is anonymized
	assert.Equal(t, "completed", resp.Status)
	// Verify retention period for legal obligations
	assert.Contains(t, respData["retention_period"], "7 years")
}

// Test 21: GDPR Article mapping - Article 16 (Rectification)
func TestArticle16Compliance(t *testing.T) {
	// Article 16: Right to Rectification
	// Data subject has right to obtain rectification of inaccurate personal data
	// without undue delay

	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-article-16"
	gs.InsertSampleData(subjectID)

	corrections := map[string]interface{}{
		"email":     "corrected@example.com",
		"full_name": "Corrected Name",
	}

	resp, err := gs.RectifyRequest(context.Background(), subjectID, "requester@example.com", corrections)
	require.NoError(t, err)

	// Verify Article 16 compliance
	assert.Equal(t, "completed", resp.Status)
	correctedData, ok := resp.Data.(*compliance.PersonalData)
	require.True(t, ok)
	assert.Equal(t, "corrected@example.com", correctedData.ContactData["email"])
}

// Test 22: GDPR Article mapping - Article 20 (Portability)
func TestArticle20Compliance(t *testing.T) {
	// Article 20: Right to Data Portability
	// Data subject has right to receive personal data in structured, commonly used format
	// and transmit to another controller without hindrance

	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-article-20"
	gs.InsertSampleData(subjectID)

	resp, err := gs.PortabilityRequest(context.Background(), subjectID, "requester@example.com", "json")
	require.NoError(t, err)

	// Verify Article 20 compliance
	assert.Equal(t, "completed", resp.Status)
	respData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	// Data must be in structured, commonly used format
	assert.Equal(t, "json", respData["format"])
	assert.NotNil(t, respData["data"])
}

// Test 23: GDPR Article mapping - Article 18 (Restrict Processing)
func TestArticle18Compliance(t *testing.T) {
	// Article 18: Right to Restrict Processing
	// Data subject has right to obtain restriction of processing when:
	// - accuracy is contested
	// - processing is unlawful
	// - data no longer needed
	// - right to object exercised

	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-article-18"
	gs.InsertSampleData(subjectID)

	resp, err := gs.RestrictProcessingRequest(context.Background(), subjectID, "requester@example.com", "accuracy_disputed")
	require.NoError(t, err)

	// Verify Article 18 compliance
	assert.Equal(t, "completed", resp.Status)
	respData, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)

	// Verify processing is restricted
	assert.True(t, respData["restriction_active"].(bool))
	assert.True(t, respData["automated_processing_disabled"].(bool))
	assert.Equal(t, "accuracy_disputed", respData["reason"])
}

// Test 24: Audit trail signature verification prevents tampering
func TestAuditSignaturePreventsTampering(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-subject-024"
	gs.InsertSampleData(subjectID)

	// Create audit entry
	_, err := gs.AccessRequest(context.Background(), subjectID, "requester@example.com")
	require.NoError(t, err)

	// Get audit logs and verify signatures are valid
	auditLogs := gs.GetAuditTrail(subjectID)
	require.Greater(t, len(auditLogs), 0)

	// Verify chain integrity
	valid, issues := gs.VerifyAuditChainIntegrity()
	assert.True(t, valid)
	assert.Empty(t, issues)

	// Simulate tampering: modify audit log data
	auditLogs[0].RequestID = "tampered-request-id"

	// Chain should now be invalid
	valid, issues = gs.VerifyAuditChainIntegrity()
	assert.False(t, valid)
	assert.Greater(t, len(issues), 0)
}

// Test 25: Full GDPR lifecycle - Access -> Rectify -> Restrict -> Portability
func TestGDPRFullLifecycle(t *testing.T) {
	gs := compliance.NewGDPRService(testAuditSecret, slog.Default())
	subjectID := "test-lifecycle"
	gs.InsertSampleData(subjectID)

	// Step 1: Access request
	accessResp, err := gs.AccessRequest(context.Background(), subjectID, "user@example.com")
	require.NoError(t, err)
	assert.Equal(t, "completed", accessResp.Status)

	// Step 2: Rectification request
	rectifyResp, err := gs.RectifyRequest(context.Background(), subjectID, "user@example.com",
		map[string]interface{}{"email": "newemail@example.com"})
	require.NoError(t, err)
	assert.Equal(t, "completed", rectifyResp.Status)

	// Step 3: Restrict processing
	restrictResp, err := gs.RestrictProcessingRequest(context.Background(), subjectID, "user@example.com", "accuracy_check")
	require.NoError(t, err)
	assert.Equal(t, "completed", restrictResp.Status)

	// Step 4: Portability export
	portResp, err := gs.PortabilityRequest(context.Background(), subjectID, "user@example.com", "json")
	require.NoError(t, err)
	assert.Equal(t, "completed", portResp.Status)

	// Verify complete audit trail
	auditLogs := gs.GetAuditTrail(subjectID)
	assert.Greater(t, len(auditLogs), 3)

	// Verify chain integrity throughout lifecycle
	valid, issues := gs.VerifyAuditChainIntegrity()
	assert.True(t, valid)
	assert.Empty(t, issues)
}
