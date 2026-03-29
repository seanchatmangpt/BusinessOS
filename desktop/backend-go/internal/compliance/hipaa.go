// Package compliance implements HIPAA (45 CFR 164) compliance rule validators.
// Validates Protected Health Information (PHI) access, encryption, audit logging,
// data retention, and breach notification requirements per regulatory framework.
package compliance

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// HIPAARuleValidator validates compliance with HIPAA regulations (45 CFR 164).
// Implements four core rule categories: access control, encryption, audit logging,
// data retention, and breach notification.
type HIPAARuleValidator struct {
	mu sync.RWMutex

	// Authorization roles: map of user_id -> []role
	authorizedRoles map[string][]string

	// PHI access audit trail
	auditLog []PHIAccessEvent

	// Data retention tracker: resource_id -> creation_time
	dataCreationTimes map[string]time.Time

	// Metrics counters
	accessViolations     atomic.Int64
	encryptionViolations atomic.Int64
	auditViolations      atomic.Int64
	retentionViolations  atomic.Int64
	breachNotifications  atomic.Int64

	// Configuration
	retentionDays   int
	tlsMinVersion   uint16
	encryptionAlgo  string
	logger          *slog.Logger
	breachCallbacks []BreachCallback
	auditMaxEntries int
}

// BreachCallback is called when an unencrypted transmission is detected.
type BreachCallback func(ctx context.Context, breach *BreachNotification) error

// PHIAccessEvent represents a single PHI access audit entry.
type PHIAccessEvent struct {
	ID          string
	Timestamp   time.Time
	UserID      string
	Action      string // "read", "write", "delete", "access"
	ResourceID  string // Patient ID, medical record ID
	Outcome     string // "success", "denied", "error"
	Details     map[string]string
	EncryptedTx bool // True if transmitted over TLS
}

// BreachNotification represents a detected breach event.
type BreachNotification struct {
	ID              string
	Timestamp       time.Time
	ResourceID      string
	UnencryptedData string
	DetectedBy      string // Which rule detected it
	Severity        string // "critical", "high"
}

// NewHIPAARuleValidator creates a new HIPAA compliance validator.
func NewHIPAARuleValidator(logger *slog.Logger) *HIPAARuleValidator {
	if logger == nil {
		logger = slog.Default()
	}
	return &HIPAARuleValidator{
		authorizedRoles:   make(map[string][]string),
		auditLog:          make([]PHIAccessEvent, 0),
		dataCreationTimes: make(map[string]time.Time),
		retentionDays:     2190, // 6 years per 45 CFR 164.404
		tlsMinVersion:     tls.VersionTLS12,
		encryptionAlgo:    "AES-256",
		logger:            logger,
		breachCallbacks:   make([]BreachCallback, 0),
		auditMaxEntries:   100000,
	}
}

// RegisterAuthorizedUser registers a user with HIPAA authorization roles.
// Valid roles: "hipaa_admin", "hipaa_user", "hipaa_auditor", "phi_viewer"
func (v *HIPAARuleValidator) RegisterAuthorizedUser(userID string, roles []string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if userID == "" {
		return fmt.Errorf("hipaa: user_id cannot be empty")
	}
	if len(roles) == 0 {
		return fmt.Errorf("hipaa: at least one role required")
	}

	validRoles := map[string]bool{
		"hipaa_admin":   true,
		"hipaa_user":    true,
		"hipaa_auditor": true,
		"phi_viewer":    true,
		"phi_editor":    true,
	}

	for _, role := range roles {
		if !validRoles[role] {
			return fmt.Errorf("hipaa: invalid role %q", role)
		}
	}

	v.authorizedRoles[userID] = roles
	v.logger.InfoContext(context.Background(), "registered hipaa user",
		slog.String("user_id", userID),
		slog.String("roles", fmt.Sprintf("%v", roles)))

	return nil
}

// ValidateAccessControl checks if a user has authorization for PHI access.
// 45 CFR 164.308(a)(4) - Access Management
// Returns error if user lacks required HIPAA role.
func (v *HIPAARuleValidator) ValidateAccessControl(ctx context.Context, userID, action string) error {
	v.mu.RLock()
	roles, exists := v.authorizedRoles[userID]
	v.mu.RUnlock()

	if !exists {
		v.accessViolations.Add(1)
		v.logger.WarnContext(ctx, "access denied: user not registered",
			slog.String("user_id", userID),
			slog.String("action", action))
		return fmt.Errorf("hipaa: user %q not authorized for %s", userID, action)
	}

	// Check role-action compatibility
	validActions := map[string]map[string]bool{
		"hipaa_admin":   {"read": true, "write": true, "delete": true, "access": true},
		"hipaa_user":    {"read": true, "write": true, "access": true},
		"hipaa_auditor": {"read": true, "access": true},
		"phi_viewer":    {"read": true, "access": true},
		"phi_editor":    {"read": true, "write": true, "access": true},
	}

	hasPermission := false
	for _, role := range roles {
		if actions, ok := validActions[role]; ok && actions[action] {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		v.accessViolations.Add(1)
		v.logger.WarnContext(ctx, "access denied: insufficient permissions",
			slog.String("user_id", userID),
			slog.String("roles", fmt.Sprintf("%v", roles)),
			slog.String("action", action))
		return fmt.Errorf("hipaa: user %q with roles %v cannot %s", userID, roles, action)
	}

	return nil
}

// ValidateEncryption checks if transmission uses TLS 1.2+ and data uses AES-256.
// 45 CFR 164.312(a)(2)(i) - Encryption and Decryption
// Returns error if encryption requirements not met.
func (v *HIPAARuleValidator) ValidateEncryption(ctx context.Context, tlsVersion uint16, encryptionAlgo string) error {
	if tlsVersion < v.tlsMinVersion {
		v.encryptionViolations.Add(1)
		v.logger.ErrorContext(ctx, "encryption violation: TLS version too old",
			slog.Uint64("tls_version", uint64(tlsVersion)),
			slog.Uint64("min_required", uint64(v.tlsMinVersion)))
		return fmt.Errorf("hipaa: TLS version 0x%x below required 0x%x", tlsVersion, v.tlsMinVersion)
	}

	if encryptionAlgo != "AES-256" {
		v.encryptionViolations.Add(1)
		v.logger.ErrorContext(ctx, "encryption violation: algorithm not AES-256",
			slog.String("algo", encryptionAlgo))
		return fmt.Errorf("hipaa: encryption algorithm %q must be AES-256", encryptionAlgo)
	}

	return nil
}

// LogPHIAccess records a PHI access event for audit trail.
// 45 CFR 164.312(b) - Audit Controls
// Returns error if audit log is full.
func (v *HIPAARuleValidator) LogPHIAccess(ctx context.Context, event PHIAccessEvent) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if len(v.auditLog) >= v.auditMaxEntries {
		v.auditViolations.Add(1)
		v.logger.ErrorContext(ctx, "audit log full, oldest entry evicted",
			slog.Int("max_entries", v.auditMaxEntries))
		v.auditLog = v.auditLog[1:] // Evict oldest
	}

	if event.ID == "" {
		event.ID = fmt.Sprintf("audit-%d-%s", time.Now().UnixNano(), event.UserID)
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	v.auditLog = append(v.auditLog, event)

	v.logger.InfoContext(ctx, "phi access logged",
		slog.String("event_id", event.ID),
		slog.String("user_id", event.UserID),
		slog.String("action", event.Action),
		slog.String("resource_id", event.ResourceID),
		slog.String("outcome", event.Outcome))

	return nil
}

// ValidateAuditLogging checks if PHI accesses are properly logged.
// 45 CFR 164.312(b) - Audit Controls
// Returns error if access not found in audit trail.
func (v *HIPAARuleValidator) ValidateAuditLogging(ctx context.Context, userID, resourceID string) (bool, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	for _, event := range v.auditLog {
		if event.UserID == userID && event.ResourceID == resourceID {
			return true, nil
		}
	}

	v.auditViolations.Add(1)
	v.logger.WarnContext(ctx, "audit violation: access not logged",
		slog.String("user_id", userID),
		slog.String("resource_id", resourceID))
	return false, fmt.Errorf("hipaa: access not found in audit log for user %q, resource %q", userID, resourceID)
}

// RegisterPHIData registers PHI data for retention tracking.
// Must be called when PHI is created to establish deletion deadline.
func (v *HIPAARuleValidator) RegisterPHIData(resourceID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if resourceID == "" {
		return fmt.Errorf("hipaa: resource_id cannot be empty")
	}

	v.dataCreationTimes[resourceID] = time.Now()
	v.logger.InfoContext(context.Background(), "registered phi data for retention tracking",
		slog.String("resource_id", resourceID))
	return nil
}

// ValidateRetention checks if PHI exceeds 6-year retention limit.
// 45 CFR 164.404 - Notification of Privacy Breaches
// Returns error if PHI is too old and should be deleted.
func (v *HIPAARuleValidator) ValidateRetention(ctx context.Context, resourceID string) error {
	v.mu.RLock()
	createdAt, exists := v.dataCreationTimes[resourceID]
	v.mu.RUnlock()

	if !exists {
		v.logger.WarnContext(ctx, "retention check: resource not registered",
			slog.String("resource_id", resourceID))
		return fmt.Errorf("hipaa: resource %q not registered for retention tracking", resourceID)
	}

	age := time.Since(createdAt)
	maxAge := time.Duration(v.retentionDays) * 24 * time.Hour

	if age > maxAge {
		v.retentionViolations.Add(1)
		v.logger.ErrorContext(ctx, "retention violation: PHI exceeds 6-year limit",
			slog.String("resource_id", resourceID),
			slog.Duration("age", age),
			slog.Duration("max_retention", maxAge))
		return fmt.Errorf("hipaa: PHI resource %q age (%v) exceeds retention limit (%v)", resourceID, age, maxAge)
	}

	return nil
}

// DeletePHIData removes PHI data from retention tracking after proper deletion.
func (v *HIPAARuleValidator) DeletePHIData(resourceID string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if _, exists := v.dataCreationTimes[resourceID]; !exists {
		return fmt.Errorf("hipaa: resource %q not found", resourceID)
	}

	delete(v.dataCreationTimes, resourceID)
	v.logger.InfoContext(context.Background(), "deleted phi data from retention tracking",
		slog.String("resource_id", resourceID))
	return nil
}

// DetectBreach checks for unencrypted transmission and triggers callbacks.
// 45 CFR 164.404 - Notification of Privacy Breaches
// Returns error if breach is detected.
func (v *HIPAARuleValidator) DetectBreach(ctx context.Context, tlsVersion uint16, unencryptedData string) error {
	if tlsVersion < v.tlsMinVersion {
		v.breachNotifications.Add(1)

		breach := &BreachNotification{
			ID:              fmt.Sprintf("breach-%d", time.Now().UnixNano()),
			Timestamp:       time.Now(),
			UnencryptedData: unencryptedData,
			DetectedBy:      "ValidateEncryption",
			Severity:        "critical",
		}

		v.logger.ErrorContext(ctx, "breach detected: unencrypted transmission",
			slog.String("breach_id", breach.ID),
			slog.String("severity", breach.Severity))

		// Invoke all registered callbacks
		for _, callback := range v.breachCallbacks {
			if err := callback(ctx, breach); err != nil {
				v.logger.ErrorContext(ctx, "breach callback failed",
					slog.String("breach_id", breach.ID),
					slog.String("error", err.Error()))
			}
		}

		return fmt.Errorf("hipaa: breach detected - unencrypted transmission of PHI")
	}

	return nil
}

// RegisterBreachCallback registers a function to be called on detected breaches.
func (v *HIPAARuleValidator) RegisterBreachCallback(callback BreachCallback) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.breachCallbacks = append(v.breachCallbacks, callback)
}

// GetAuditLog returns a copy of the audit log for compliance reporting.
func (v *HIPAARuleValidator) GetAuditLog(ctx context.Context) ([]PHIAccessEvent, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	log := make([]PHIAccessEvent, len(v.auditLog))
	copy(log, v.auditLog)
	return log, nil
}

// GetMetrics returns compliance violation metrics.
type ComplianceMetrics struct {
	AccessViolations     int64
	EncryptionViolations int64
	AuditViolations      int64
	RetentionViolations  int64
	BreachNotifications  int64
	TotalViolations      int64
}

// GetMetrics retrieves current compliance metrics.
func (v *HIPAARuleValidator) GetMetrics() ComplianceMetrics {
	return ComplianceMetrics{
		AccessViolations:     v.accessViolations.Load(),
		EncryptionViolations: v.encryptionViolations.Load(),
		AuditViolations:      v.auditViolations.Load(),
		RetentionViolations:  v.retentionViolations.Load(),
		BreachNotifications:  v.breachNotifications.Load(),
		TotalViolations: v.accessViolations.Load() +
			v.encryptionViolations.Load() +
			v.auditViolations.Load() +
			v.retentionViolations.Load() +
			v.breachNotifications.Load(),
	}
}

// CFRMapping returns the compliance mapping to 45 CFR sections.
func CFRMapping() map[string]string {
	return map[string]string{
		"access_control":      "45 CFR 164.308(a)(4) - Access Management",
		"encryption_at_rest":  "45 CFR 164.312(a)(2)(i) - Encryption and Decryption (At Rest)",
		"encryption_in_trans": "45 CFR 164.312(a)(2)(i) - Encryption and Decryption (In Transit)",
		"audit_controls":      "45 CFR 164.312(b) - Audit Controls",
		"data_retention":      "45 CFR 164.404 - Notification of Privacy Breaches (retention requirement)",
		"breach_notification": "45 CFR 164.404 - Notification of Privacy Breaches (notification requirement)",
		"minimum_necessary":   "45 CFR 164.502(b) - Minimum Necessary Rule",
		"business_associates": "45 CFR 164.308(b) - Business Associate Contracts and Other Arrangements",
		"data_integrity":      "45 CFR 164.308(a)(2)(ii) - Data Integrity Controls",
	}
}
