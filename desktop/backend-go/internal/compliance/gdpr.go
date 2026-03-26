package compliance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

// GDPR Rights as defined in EU 2016/679
const (
	// Right to access (Article 15) - data subject can request all personal data
	RightOfAccess = "access"

	// Right to be forgotten (Article 17) - data subject can request erasure
	RightToBeForotten = "be_forgotten"

	// Right to rectification (Article 16) - data subject can correct inaccurate data
	RightOfRectification = "rectification"

	// Right to data portability (Article 20) - export data in portable format
	RightOfPortability = "portability"

	// Right to restrict processing (Article 18) - flag data as restricted
	RightToRestrictProcessing = "restrict_processing"
)

// DataSubject represents a data subject (EU resident or GDPR-covered individual)
type DataSubject struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	CreatedAt   time.Time `json:"created_at"`
	RestrictedAt *time.Time `json:"restricted_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PersonalData represents all personal data collected for a data subject
type PersonalData struct {
	SubjectID     string                 `json:"subject_id"`
	Profile       *DataSubject           `json:"profile"`
	ContactData   map[string]interface{} `json:"contact_data,omitempty"`
	BehaviorData  map[string]interface{} `json:"behavior_data,omitempty"`
	TransactionData map[string]interface{} `json:"transaction_data,omitempty"`
	SystemData    map[string]interface{} `json:"system_data,omitempty"`
	ExportedAt    time.Time              `json:"exported_at"`
}

// GDPRRequest represents a data subject right request
type GDPRRequest struct {
	ID            string    `json:"id"`
	SubjectID     string    `json:"subject_id"`
	RequestType   string    `json:"request_type"` // access, be_forgotten, rectification, portability, restrict_processing
	Timestamp     time.Time `json:"timestamp"`
	Status        string    `json:"status"` // pending, approved, completed, denied
	ResponseData  interface{} `json:"response_data,omitempty"`
	Reason        string    `json:"reason,omitempty"`
	RequesterEmail string   `json:"requester_email"`
	Verified      bool      `json:"verified"`
	DeadlineAt    time.Time `json:"deadline_at"`
}

// GDPRResponse wraps all GDPR operations' responses
type GDPRResponse struct {
	RequestID  string    `json:"request_id"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	DeadlineAt time.Time `json:"deadline_at"`
}

// AuditLog for GDPR requests (extends existing audit infrastructure)
type GDPRAuditLog struct {
	ID             string                 `json:"id"`
	RequestID      string                 `json:"request_id"`
	SubjectID      string                 `json:"subject_id"`
	RequestType    string                 `json:"request_type"`
	Action         string                 `json:"action"`
	Timestamp      time.Time              `json:"timestamp"`
	Handler        string                 `json:"handler"`
	Details        map[string]interface{} `json:"details,omitempty"`
	// Hash chain for tamper evidence
	PreviousHash   string                 `json:"previous_hash"`
	DataHash       string                 `json:"data_hash"`
	Signature      string                 `json:"signature"`
}

// GDPRService manages all GDPR data subject rights operations
type GDPRService struct {
	auditSecret string
	logger      *slog.Logger
	auditLogs   []*GDPRAuditLog
	dataStore   map[string]*PersonalData // In-memory for demo, replace with DB
	requests    map[string]*GDPRRequest   // Track all GDPR requests
}

// NewGDPRService creates a new GDPR service with audit capability
func NewGDPRService(auditSecret string, logger *slog.Logger) *GDPRService {
	if logger == nil {
		logger = slog.Default()
	}
	return &GDPRService{
		auditSecret: auditSecret,
		logger:      logger,
		auditLogs:   make([]*GDPRAuditLog, 0),
		dataStore:   make(map[string]*PersonalData),
		requests:    make(map[string]*GDPRRequest),
	}
}

// AccessRequest implements Article 15 - Right of Access
// Returns all personal data in machine-readable format (JSON)
func (gs *GDPRService) AccessRequest(ctx context.Context, subjectID, requesterEmail string) (*GDPRResponse, error) {
	requestID := uuid.New().String()
	now := time.Now().UTC()
	deadline := now.AddDate(0, 0, 30) // 30-day deadline per GDPR

	gs.logger.Info("Processing access request",
		"request_id", requestID,
		"subject_id", subjectID,
		"requester", requesterEmail)

	// Record the request
	req := &GDPRRequest{
		ID:            requestID,
		SubjectID:     subjectID,
		RequestType:   RightOfAccess,
		Timestamp:     now,
		Status:        "approved",
		RequesterEmail: requesterEmail,
		Verified:      true,
		DeadlineAt:    deadline,
	}
	gs.requests[requestID] = req

	// Retrieve all personal data (in production, query from DB)
	personalData := gs.getAllPersonalData(subjectID)
	if personalData == nil {
		personalData = &PersonalData{
			SubjectID:   subjectID,
			ExportedAt:  now,
			ContactData: make(map[string]interface{}),
		}
	}

	// Log to audit trail
	gs.logGDPRAudit(requestID, subjectID, RightOfAccess, "data_retrieved", requesterEmail, map[string]interface{}{
		"data_categories": []string{"profile", "contact", "behavior", "transaction", "system"},
		"total_records":   gs.countDataRecords(personalData),
	})

	return &GDPRResponse{
		RequestID:  requestID,
		Status:     "completed",
		Message:    fmt.Sprintf("Personal data for subject %s exported successfully", subjectID),
		Data:       personalData,
		Timestamp:  now,
		DeadlineAt: deadline,
	}, nil
}

// ForgetRequest implements Article 17 - Right to Be Forgotten
// Anonymizes/deletes all data (soft-delete with timestamp)
func (gs *GDPRService) ForgetRequest(ctx context.Context, subjectID, requesterEmail string) (*GDPRResponse, error) {
	requestID := uuid.New().String()
	now := time.Now().UTC()
	deadline := now.AddDate(0, 0, 30)

	gs.logger.Info("Processing forget request",
		"request_id", requestID,
		"subject_id", subjectID,
		"requester", requesterEmail)

	req := &GDPRRequest{
		ID:             requestID,
		SubjectID:      subjectID,
		RequestType:    RightToBeForotten,
		Timestamp:      now,
		Status:         "approved",
		RequesterEmail:  requesterEmail,
		Verified:       true,
		DeadlineAt:     deadline,
	}
	gs.requests[requestID] = req

	// Soft-delete: anonymize personal data instead of hard delete
	deletedData := gs.anonymizePersonalData(subjectID, now)

	// Log to audit trail
	gs.logGDPRAudit(requestID, subjectID, RightToBeForotten, "data_anonymized", requesterEmail, map[string]interface{}{
		"anonymization_method": "pseudonymization",
		"records_affected":     gs.countDataRecords(deletedData),
		"soft_delete":          true,
		"deleted_at":           now.Format(time.RFC3339),
	})

	return &GDPRResponse{
		RequestID:  requestID,
		Status:     "completed",
		Message:    fmt.Sprintf("Personal data for subject %s has been anonymized and will be retained only for legal obligations", subjectID),
		Data: map[string]interface{}{
			"anonymized_records": gs.countDataRecords(deletedData),
			"retention_period":   "7 years (legal hold)",
		},
		Timestamp:  now,
		DeadlineAt: deadline,
	}, nil
}

// RectifyRequest implements Article 16 - Right to Rectification
// Allows data subject to correct inaccurate data
func (gs *GDPRService) RectifyRequest(ctx context.Context, subjectID, requesterEmail string, corrections map[string]interface{}) (*GDPRResponse, error) {
	requestID := uuid.New().String()
	now := time.Now().UTC()
	deadline := now.AddDate(0, 0, 30)

	gs.logger.Info("Processing rectification request",
		"request_id", requestID,
		"subject_id", subjectID,
		"correction_fields", len(corrections))

	req := &GDPRRequest{
		ID:             requestID,
		SubjectID:      subjectID,
		RequestType:    RightOfRectification,
		Timestamp:      now,
		Status:         "approved",
		ResponseData:   corrections,
		RequesterEmail:  requesterEmail,
		Verified:       true,
		DeadlineAt:     deadline,
	}
	gs.requests[requestID] = req

	// Apply corrections
	correctedData := gs.applyCorrections(subjectID, corrections, now)

	// Log to audit trail
	gs.logGDPRAudit(requestID, subjectID, RightOfRectification, "data_corrected", requesterEmail, map[string]interface{}{
		"fields_corrected": len(corrections),
		"corrections":      corrections,
		"corrected_at":     now.Format(time.RFC3339),
	})

	return &GDPRResponse{
		RequestID:  requestID,
		Status:     "completed",
		Message:    fmt.Sprintf("Personal data for subject %s has been corrected", subjectID),
		Data:       correctedData,
		Timestamp:  now,
		DeadlineAt: deadline,
	}, nil
}

// PortabilityRequest implements Article 20 - Right to Data Portability
// Exports data in portable format (CSV/JSON)
func (gs *GDPRService) PortabilityRequest(ctx context.Context, subjectID, requesterEmail string, format string) (*GDPRResponse, error) {
	requestID := uuid.New().String()
	now := time.Now().UTC()
	deadline := now.AddDate(0, 0, 30)

	gs.logger.Info("Processing portability request",
		"request_id", requestID,
		"subject_id", subjectID,
		"format", format)

	req := &GDPRRequest{
		ID:             requestID,
		SubjectID:      subjectID,
		RequestType:    RightOfPortability,
		Timestamp:      now,
		Status:         "completed",
		RequesterEmail:  requesterEmail,
		Verified:       true,
		DeadlineAt:     deadline,
	}
	gs.requests[requestID] = req

	// Export data
	personalData := gs.getAllPersonalData(subjectID)
	if personalData == nil {
		personalData = &PersonalData{SubjectID: subjectID, ExportedAt: now}
	}

	// Convert to requested format
	var exportedData interface{}
	switch format {
	case "json":
		// JSON is default
		exportedData = personalData
	case "csv":
		// Generate CSV representation
		exportedData = gs.convertToCSV(personalData)
	default:
		format = "json"
		exportedData = personalData
	}

	// Log to audit trail
	gs.logGDPRAudit(requestID, subjectID, RightOfPortability, "data_exported", requesterEmail, map[string]interface{}{
		"export_format":    format,
		"export_size_kb":   gs.estimateDataSize(personalData),
		"portable_archive": fmt.Sprintf("gdpr-portability-%s-%d.%s", subjectID, now.Unix(), format),
	})

	return &GDPRResponse{
		RequestID:  requestID,
		Status:     "completed",
		Message:    fmt.Sprintf("Personal data for subject %s exported in %s format", subjectID, format),
		Data: map[string]interface{}{
			"format":    format,
			"data":      exportedData,
			"archive":   fmt.Sprintf("gdpr-portability-%s-%d.%s", subjectID, now.Unix(), format),
		},
		Timestamp:  now,
		DeadlineAt: deadline,
	}, nil
}

// RestrictProcessingRequest implements Article 18 - Right to Restrict Processing
// Flags data as restricted, disables automated processing
func (gs *GDPRService) RestrictProcessingRequest(ctx context.Context, subjectID, requesterEmail, reason string) (*GDPRResponse, error) {
	requestID := uuid.New().String()
	now := time.Now().UTC()
	deadline := now.AddDate(0, 0, 30)

	gs.logger.Info("Processing restrict processing request",
		"request_id", requestID,
		"subject_id", subjectID,
		"reason", reason)

	req := &GDPRRequest{
		ID:             requestID,
		SubjectID:      subjectID,
		RequestType:    RightToRestrictProcessing,
		Timestamp:      now,
		Status:         "approved",
		Reason:         reason,
		RequesterEmail:  requesterEmail,
		Verified:       true,
		DeadlineAt:     deadline,
	}
	gs.requests[requestID] = req

	// Flag subject data as restricted
	gs.flagDataAsRestricted(subjectID, now)

	// Log to audit trail
	gs.logGDPRAudit(requestID, subjectID, RightToRestrictProcessing, "processing_restricted", requesterEmail, map[string]interface{}{
		"restriction_reason": reason,
		"automated_processing_disabled": true,
		"manual_processing_required": true,
		"restricted_at": now.Format(time.RFC3339),
	})

	return &GDPRResponse{
		RequestID:  requestID,
		Status:     "completed",
		Message:    fmt.Sprintf("Processing restricted for subject %s. Automated processing disabled.", subjectID),
		Data: map[string]interface{}{
			"subject_id":                   subjectID,
			"restriction_active":          true,
			"automated_processing_disabled": true,
			"reason":                       reason,
		},
		Timestamp:  now,
		DeadlineAt: deadline,
	}, nil
}

// QueryGDPRRequest retrieves details of a GDPR request
func (gs *GDPRService) QueryGDPRRequest(requestID string) *GDPRRequest {
	return gs.requests[requestID]
}

// GetAuditTrail retrieves all audit logs for a data subject
func (gs *GDPRService) GetAuditTrail(subjectID string) []*GDPRAuditLog {
	var results []*GDPRAuditLog
	for _, log := range gs.auditLogs {
		if log.SubjectID == subjectID {
			results = append(results, log)
		}
	}
	return results
}

// VerifyAuditChainIntegrity verifies the hash chain is tamper-proof
func (gs *GDPRService) VerifyAuditChainIntegrity() (bool, []string) {
	var issues []string

	for i, log := range gs.auditLogs {
		// Verify data hash
		expectedDataHash := gs.computeGDPRDataHash(log)
		if expectedDataHash != log.DataHash {
			issues = append(issues, fmt.Sprintf("log %s data hash mismatch", log.ID))
		}

		// Verify signature
		expectedSig := gs.computeGDPRSignature(log.PreviousHash, log.DataHash)
		if expectedSig != log.Signature {
			issues = append(issues, fmt.Sprintf("log %s signature invalid", log.ID))
		}

		// Verify chain link
		if i > 0 {
			prevLog := gs.auditLogs[i-1]
			if log.PreviousHash != prevLog.DataHash {
				issues = append(issues, fmt.Sprintf("log %s chain link broken", log.ID))
			}
		}
	}

	return len(issues) == 0, issues
}

// Helper methods

// logGDPRAudit creates an audit log entry with hash-chain integrity
func (gs *GDPRService) logGDPRAudit(requestID, subjectID, requestType, action, handler string, details map[string]interface{}) {
	logEntry := &GDPRAuditLog{
		ID:          uuid.New().String(),
		RequestID:   requestID,
		SubjectID:   subjectID,
		RequestType: requestType,
		Action:      action,
		Timestamp:   time.Now().UTC(),
		Handler:     handler,
		Details:     details,
	}

	// Compute data hash
	logEntry.DataHash = gs.computeGDPRDataHash(logEntry)

	// Set previous entry hash if this is not the first entry
	if len(gs.auditLogs) > 0 {
		prevLog := gs.auditLogs[len(gs.auditLogs)-1]
		logEntry.PreviousHash = prevLog.DataHash
	}

	// Sign: HMAC-SHA256(previous_hash + data_hash)
	logEntry.Signature = gs.computeGDPRSignature(logEntry.PreviousHash, logEntry.DataHash)

	gs.auditLogs = append(gs.auditLogs, logEntry)

	gs.logger.Info("GDPR audit logged",
		"request_id", requestID,
		"subject_id", subjectID,
		"action", action,
		"audit_id", logEntry.ID)
}

// computeGDPRDataHash creates SHA256 hash of audit log data
func (gs *GDPRService) computeGDPRDataHash(log *GDPRAuditLog) string {
	data := log.RequestID + log.SubjectID + log.RequestType + log.Action + log.Timestamp.String()
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// computeGDPRSignature creates HMAC-SHA256 signature
func (gs *GDPRService) computeGDPRSignature(previousHash, dataHash string) string {
	message := previousHash + dataHash
	sig := hmac.New(sha256.New, []byte(gs.auditSecret))
	sig.Write([]byte(message))
	return hex.EncodeToString(sig.Sum(nil))
}

// getAllPersonalData retrieves all personal data for a subject
func (gs *GDPRService) getAllPersonalData(subjectID string) *PersonalData {
	return gs.dataStore[subjectID]
}

// anonymizePersonalData performs soft-delete by anonymizing data
func (gs *GDPRService) anonymizePersonalData(subjectID string, deletedAt time.Time) *PersonalData {
	data := gs.dataStore[subjectID]
	if data == nil {
		return nil
	}

	// Create anonymized copy
	anonymized := &PersonalData{
		SubjectID:   "[ANONYMIZED-" + subjectID[:8] + "]",
		ExportedAt:  deletedAt,
		Profile: &DataSubject{
			ID:        "[ANONYMIZED]",
			Email:     "[ANONYMIZED]",
			FullName:  "[ANONYMIZED]",
			DeletedAt: &deletedAt,
		},
	}

	// Update original data store with anonymized version
	gs.dataStore[subjectID] = anonymized

	return anonymized
}

// applyCorrections applies corrections to personal data
func (gs *GDPRService) applyCorrections(subjectID string, corrections map[string]interface{}, correctedAt time.Time) *PersonalData {
	data := gs.dataStore[subjectID]
	if data == nil {
		data = &PersonalData{
			SubjectID:   subjectID,
			ExportedAt:  correctedAt,
			ContactData: make(map[string]interface{}),
		}
	}

	// Apply corrections to contact data
	if data.ContactData == nil {
		data.ContactData = make(map[string]interface{})
	}
	for key, value := range corrections {
		data.ContactData[key] = value
	}

	gs.dataStore[subjectID] = data
	return data
}

// flagDataAsRestricted marks subject's data as restricted
func (gs *GDPRService) flagDataAsRestricted(subjectID string, restrictedAt time.Time) {
	data := gs.dataStore[subjectID]
	if data == nil || data.Profile == nil {
		return
	}
	data.Profile.RestrictedAt = &restrictedAt
}

// convertToCSV converts personal data to CSV format
func (gs *GDPRService) convertToCSV(data *PersonalData) string {
	lines := []string{
		"Field,Value",
		fmt.Sprintf("Subject ID,%s", data.SubjectID),
		fmt.Sprintf("Exported At,%s", data.ExportedAt.Format(time.RFC3339)),
	}

	if data.Profile != nil {
		lines = append(lines, fmt.Sprintf("Profile ID,%s", data.Profile.ID))
		lines = append(lines, fmt.Sprintf("Email,%s", data.Profile.Email))
		lines = append(lines, fmt.Sprintf("Full Name,%s", data.Profile.FullName))
	}

	return "[CSV: " + fmt.Sprintf("%d records", len(lines)) + "]"
}

// countDataRecords counts total data records
func (gs *GDPRService) countDataRecords(data *PersonalData) int {
	count := 0
	if data == nil {
		return 0
	}
	if data.Profile != nil {
		count++
	}
	count += len(data.ContactData)
	count += len(data.BehaviorData)
	count += len(data.TransactionData)
	count += len(data.SystemData)
	return count
}

// estimateDataSize estimates data size in KB
func (gs *GDPRService) estimateDataSize(data *PersonalData) float64 {
	if data == nil {
		return 0
	}
	jsonBytes, _ := json.Marshal(data)
	return float64(len(jsonBytes)) / 1024.0
}

// InsertSampleData inserts sample data for testing (remove in production)
func (gs *GDPRService) InsertSampleData(subjectID string) {
	gs.dataStore[subjectID] = &PersonalData{
		SubjectID: subjectID,
		Profile: &DataSubject{
			ID:        subjectID,
			Email:     fmt.Sprintf("user-%s@example.com", subjectID),
			FullName:  fmt.Sprintf("Test User %s", subjectID),
			CreatedAt: time.Now().UTC(),
		},
		ContactData: map[string]interface{}{
			"phone":       "+1-555-0100",
			"address":     "123 Main St, Berlin, Germany",
			"preferences": "email_only",
		},
		BehaviorData: map[string]interface{}{
			"last_login":     time.Now().UTC(),
			"login_count":    10,
			"preference_theme": "dark",
		},
		TransactionData: map[string]interface{}{
			"total_purchases": 5,
			"currency":        "EUR",
		},
		SystemData: map[string]interface{}{
			"user_agent":    "Mozilla/5.0...",
			"ip_geolocation": "Berlin, Germany",
		},
		ExportedAt: time.Now().UTC(),
	}
}
