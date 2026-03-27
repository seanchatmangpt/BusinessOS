package compliance

import (
	"context"
	"crypto/tls"
	"log/slog"
	"os"
	"testing"
	"time"
)

// TestNewHIPAARuleValidator_DefaultConfiguration tests initialization.
func TestNewHIPAARuleValidator_DefaultConfiguration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	v := NewHIPAARuleValidator(logger)

	if v == nil {
		t.Fatal("NewHIPAARuleValidator returned nil")
	}
	if v.retentionDays != 2190 { // 6 years
		t.Errorf("want retentionDays=2190, got %d", v.retentionDays)
	}
	if v.tlsMinVersion != tls.VersionTLS12 {
		t.Errorf("want tlsMinVersion=0x%x, got 0x%x", tls.VersionTLS12, v.tlsMinVersion)
	}
	if v.encryptionAlgo != "AES-256" {
		t.Errorf("want encryptionAlgo=AES-256, got %s", v.encryptionAlgo)
	}
}

// TestRegisterAuthorizedUser_ValidRoles tests user registration with valid roles.
func TestRegisterAuthorizedUser_ValidRoles(t *testing.T) {
	v := NewHIPAARuleValidator(nil)

	tests := []struct {
		userID string
		roles  []string
		valid  bool
	}{
		{"user1", []string{"hipaa_admin"}, true},
		{"user2", []string{"hipaa_user", "phi_viewer"}, true},
		{"user3", []string{"hipaa_auditor"}, true},
		{"user4", []string{"phi_editor"}, true},
		{"", []string{"hipaa_admin"}, false}, // empty user_id
		{"user5", []string{}, false},          // no roles
		{"user6", []string{"invalid_role"}, false},
	}

	for _, tt := range tests {
		err := v.RegisterAuthorizedUser(tt.userID, tt.roles)
		if (err == nil) != tt.valid {
			t.Errorf("RegisterAuthorizedUser(%q, %v): got error=%v, want valid=%v", tt.userID, tt.roles, err, tt.valid)
		}
	}
}

// TestValidateAccessControl_AuthorizedAccess tests successful PHI access.
func TestValidateAccessControl_AuthorizedAccess(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	v.RegisterAuthorizedUser("doctor1", []string{"hipaa_user"})

	tests := []struct {
		userID string
		action string
		valid  bool
	}{
		{"doctor1", "read", true},
		{"doctor1", "write", true},
		{"doctor1", "delete", false}, // hipaa_user can't delete
		{"unknown_user", "read", false},
	}

	for _, tt := range tests {
		err := v.ValidateAccessControl(ctx, tt.userID, tt.action)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateAccessControl(%q, %q): got error=%v, want valid=%v", tt.userID, tt.action, err, tt.valid)
		}
	}
}

// TestValidateAccessControl_RoleBasedPermissions tests role-specific permissions.
func TestValidateAccessControl_RoleBasedPermissions(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	v.RegisterAuthorizedUser("admin", []string{"hipaa_admin"})
	v.RegisterAuthorizedUser("auditor", []string{"hipaa_auditor"})

	// Admin can do everything
	if err := v.ValidateAccessControl(ctx, "admin", "read"); err != nil {
		t.Errorf("admin read: got error=%v", err)
	}
	if err := v.ValidateAccessControl(ctx, "admin", "write"); err != nil {
		t.Errorf("admin write: got error=%v", err)
	}
	if err := v.ValidateAccessControl(ctx, "admin", "delete"); err != nil {
		t.Errorf("admin delete: got error=%v", err)
	}

	// Auditor can only read
	if err := v.ValidateAccessControl(ctx, "auditor", "read"); err != nil {
		t.Errorf("auditor read: got error=%v", err)
	}
	if err := v.ValidateAccessControl(ctx, "auditor", "write"); err == nil {
		t.Errorf("auditor write: should fail but didn't")
	}
}

// TestValidateEncryption_TLSVersion tests TLS version validation.
func TestValidateEncryption_TLSVersion(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	tests := []struct {
		tlsVersion uint16
		algo       string
		valid      bool
	}{
		{tls.VersionTLS13, "AES-256", true},
		{tls.VersionTLS12, "AES-256", true},
		{tls.VersionTLS11, "AES-256", false}, // Too old
		{tls.VersionSSL30, "AES-256", false}, // Too old
	}

	for _, tt := range tests {
		err := v.ValidateEncryption(ctx, tt.tlsVersion, tt.algo)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateEncryption(0x%x, %q): got error=%v, want valid=%v", tt.tlsVersion, tt.algo, err, tt.valid)
		}
	}
}

// TestValidateEncryption_EncryptionAlgorithm tests encryption algorithm validation.
func TestValidateEncryption_EncryptionAlgorithm(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	tests := []struct {
		algo  string
		valid bool
	}{
		{"AES-256", true},
		{"AES-128", false},
		{"DES", false},
		{"RSA", false},
		{"", false},
	}

	for _, tt := range tests {
		err := v.ValidateEncryption(ctx, tls.VersionTLS12, tt.algo)
		if (err == nil) != tt.valid {
			t.Errorf("ValidateEncryption with algo=%q: got error=%v, want valid=%v", tt.algo, err, tt.valid)
		}
	}
}

// TestLogPHIAccess_AuditTrail tests audit logging functionality.
func TestLogPHIAccess_AuditTrail(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	event1 := PHIAccessEvent{
		UserID:     "doctor1",
		Action:     "read",
		ResourceID: "patient123",
		Outcome:    "success",
	}

	event2 := PHIAccessEvent{
		UserID:     "doctor1",
		Action:     "write",
		ResourceID: "patient123",
		Outcome:    "success",
	}

	if err := v.LogPHIAccess(ctx, event1); err != nil {
		t.Errorf("LogPHIAccess(event1): got error=%v", err)
	}
	if err := v.LogPHIAccess(ctx, event2); err != nil {
		t.Errorf("LogPHIAccess(event2): got error=%v", err)
	}

	log, _ := v.GetAuditLog(ctx)
	if len(log) != 2 {
		t.Errorf("GetAuditLog: got %d events, want 2", len(log))
	}
}

// TestValidateAuditLogging_AccessFound tests audit log verification.
func TestValidateAuditLogging_AccessFound(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	event := PHIAccessEvent{
		UserID:     "doctor1",
		Action:     "read",
		ResourceID: "patient123",
		Outcome:    "success",
	}

	v.LogPHIAccess(ctx, event)

	found, err := v.ValidateAuditLogging(ctx, "doctor1", "patient123")
	if !found || err != nil {
		t.Errorf("ValidateAuditLogging: got found=%v, error=%v", found, err)
	}

	// Non-existent access should not be found
	found, err = v.ValidateAuditLogging(ctx, "unknown", "patient999")
	if found || err == nil {
		t.Errorf("ValidateAuditLogging(unknown, patient999): got found=%v, error=%v", found, err)
	}
}

// TestRegisterPHIData_TrackingCreation tests PHI data registration.
func TestRegisterPHIData_TrackingCreation(t *testing.T) {
	v := NewHIPAARuleValidator(nil)

	tests := []struct {
		resourceID string
		valid      bool
	}{
		{"patient123", true},
		{"medicalrecord456", true},
		{"", false},
	}

	for _, tt := range tests {
		err := v.RegisterPHIData(tt.resourceID)
		if (err == nil) != tt.valid {
			t.Errorf("RegisterPHIData(%q): got error=%v, want valid=%v", tt.resourceID, err, tt.valid)
		}
	}
}

// TestValidateRetention_WithinLimit tests retention validation when within limit.
func TestValidateRetention_WithinLimit(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	v.RegisterPHIData("patient123")

	err := v.ValidateRetention(ctx, "patient123")
	if err != nil {
		t.Errorf("ValidateRetention (new data): got error=%v", err)
	}
}

// TestValidateRetention_ExceedsLimit tests retention validation when exceeds limit.
func TestValidateRetention_ExceedsLimit(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	// Register PHI data then manually set creation time to > 6 years ago
	v.RegisterPHIData("patient123")
	v.mu.Lock()
	v.dataCreationTimes["patient123"] = time.Now().AddDate(-7, 0, 0) // 7 years ago
	v.mu.Unlock()

	err := v.ValidateRetention(ctx, "patient123")
	if err == nil {
		t.Errorf("ValidateRetention (old data): should fail but didn't")
	}

	metrics := v.GetMetrics()
	if metrics.RetentionViolations == 0 {
		t.Errorf("RetentionViolations counter: got 0, want > 0")
	}
}

// TestDeletePHIData_RemovalFromTracking tests PHI data deletion.
func TestDeletePHIData_RemovalFromTracking(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	v.RegisterPHIData("patient123")

	err := v.DeletePHIData("patient123")
	if err != nil {
		t.Errorf("DeletePHIData: got error=%v", err)
	}

	// After deletion, retention check should fail (not found)
	err = v.ValidateRetention(ctx, "patient123")
	if err == nil {
		t.Errorf("ValidateRetention (after delete): should fail but didn't")
	}
}

// TestDetectBreach_UnencryptedTransmission tests breach detection.
func TestDetectBreach_UnencryptedTransmission(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	// Old TLS version should trigger breach
	err := v.DetectBreach(ctx, tls.VersionTLS11, "sensitive_patient_data")
	if err == nil {
		t.Errorf("DetectBreach (TLS 1.1): should fail but didn't")
	}

	metrics := v.GetMetrics()
	if metrics.BreachNotifications == 0 {
		t.Errorf("BreachNotifications counter: got 0, want > 0")
	}

	// Proper TLS should not trigger
	err = v.DetectBreach(ctx, tls.VersionTLS12, "data")
	if err != nil {
		t.Errorf("DetectBreach (TLS 1.2): got error=%v", err)
	}
}

// TestRegisterBreachCallback_CallbackInvoked tests breach notification callback.
func TestRegisterBreachCallback_CallbackInvoked(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	callbackInvoked := false
	var capturedBreach *BreachNotification

	v.RegisterBreachCallback(func(ctx context.Context, breach *BreachNotification) error {
		callbackInvoked = true
		capturedBreach = breach
		return nil
	})

	v.DetectBreach(ctx, tls.VersionTLS11, "patient_ssn")

	if !callbackInvoked {
		t.Errorf("Breach callback: not invoked")
	}
	if capturedBreach == nil {
		t.Errorf("Breach callback: breach data is nil")
	}
	if capturedBreach.Severity != "critical" {
		t.Errorf("Breach severity: got %q, want critical", capturedBreach.Severity)
	}
}

// TestGetMetrics_ViolationTracking tests metrics collection.
func TestGetMetrics_ViolationTracking(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	// Trigger various violations
	v.ValidateAccessControl(ctx, "unknown_user", "read")
	v.ValidateEncryption(ctx, tls.VersionTLS11, "AES-256")
	v.ValidateAuditLogging(ctx, "unknown", "resource")
	v.RegisterPHIData("patient123")
	v.RegisterPHIData("old_patient")
	v.mu.Lock()
	v.dataCreationTimes["old_patient"] = time.Now().AddDate(-7, 0, 0) // 7 years ago
	v.mu.Unlock()
	v.ValidateRetention(ctx, "old_patient")
	v.DetectBreach(ctx, tls.VersionTLS10, "data")

	metrics := v.GetMetrics()

	if metrics.AccessViolations == 0 {
		t.Errorf("AccessViolations: got 0, want > 0")
	}
	if metrics.EncryptionViolations == 0 {
		t.Errorf("EncryptionViolations: got 0, want > 0")
	}
	if metrics.AuditViolations == 0 {
		t.Errorf("AuditViolations: got 0, want > 0")
	}
	if metrics.RetentionViolations == 0 {
		t.Errorf("RetentionViolations: got 0, want > 0")
	}
	if metrics.BreachNotifications == 0 {
		t.Errorf("BreachNotifications: got 0, want > 0")
	}
	if metrics.TotalViolations == 0 {
		t.Errorf("TotalViolations: got 0, want > 0")
	}
}

// TestCFRMapping_ComplianceReferences tests CFR mapping table.
func TestCFRMapping_ComplianceReferences(t *testing.T) {
	mapping := CFRMapping()

	expectedKeys := []string{
		"access_control",
		"encryption_at_rest",
		"encryption_in_trans",
		"audit_controls",
		"data_retention",
		"breach_notification",
		"minimum_necessary",
		"business_associates",
		"data_integrity",
	}

	for _, key := range expectedKeys {
		if value, exists := mapping[key]; !exists {
			t.Errorf("CFRMapping missing key %q", key)
		} else if value == "" {
			t.Errorf("CFRMapping[%q] is empty", key)
		}
	}
}

// TestGetAuditLog_CopyReturned tests that audit log copy is returned.
func TestGetAuditLog_CopyReturned(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	event := PHIAccessEvent{
		UserID:     "doctor1",
		Action:     "read",
		ResourceID: "patient123",
		Outcome:    "success",
	}

	v.LogPHIAccess(ctx, event)

	log1, _ := v.GetAuditLog(ctx)
	log2, _ := v.GetAuditLog(ctx)

	if len(log1) != len(log2) {
		t.Errorf("GetAuditLog: lengths differ %d vs %d", len(log1), len(log2))
	}

	if len(log1) > 0 && log1[0].UserID != log2[0].UserID {
		t.Errorf("GetAuditLog: copies differ")
	}
}

// TestComplexScenario_MultiUserMultiAction tests realistic multi-user workflow.
func TestComplexScenario_MultiUserMultiAction(t *testing.T) {
	v := NewHIPAARuleValidator(nil)
	ctx := context.Background()

	// Register multiple users with different roles
	v.RegisterAuthorizedUser("doctor1", []string{"hipaa_user"})
	v.RegisterAuthorizedUser("nurse1", []string{"phi_editor"})
	v.RegisterAuthorizedUser("auditor1", []string{"hipaa_auditor"})

	// Register PHI data
	v.RegisterPHIData("patient_john_doe")

	// Doctor reads patient data
	if err := v.ValidateAccessControl(ctx, "doctor1", "read"); err != nil {
		t.Errorf("doctor1 read: got error=%v", err)
	}

	v.LogPHIAccess(ctx, PHIAccessEvent{
		UserID:     "doctor1",
		Action:     "read",
		ResourceID: "patient_john_doe",
		Outcome:    "success",
	})

	// Nurse edits patient data with proper encryption
	if err := v.ValidateAccessControl(ctx, "nurse1", "write"); err != nil {
		t.Errorf("nurse1 write: got error=%v", err)
	}

	if err := v.ValidateEncryption(ctx, tls.VersionTLS12, "AES-256"); err != nil {
		t.Errorf("encryption validation: got error=%v", err)
	}

	v.LogPHIAccess(ctx, PHIAccessEvent{
		UserID:     "nurse1",
		Action:     "write",
		ResourceID: "patient_john_doe",
		Outcome:    "success",
	})

	// Auditor reviews logs
	found, _ := v.ValidateAuditLogging(ctx, "doctor1", "patient_john_doe")
	if !found {
		t.Errorf("auditor: doctor1 read not found in log")
	}

	found, _ = v.ValidateAuditLogging(ctx, "nurse1", "patient_john_doe")
	if !found {
		t.Errorf("auditor: nurse1 write not found in log")
	}

	// Verify retention is within limit
	if err := v.ValidateRetention(ctx, "patient_john_doe"); err != nil {
		t.Errorf("retention check: got error=%v", err)
	}

	metrics := v.GetMetrics()
	if metrics.TotalViolations != 0 {
		t.Errorf("complex scenario: got %d violations, want 0", metrics.TotalViolations)
	}
}
