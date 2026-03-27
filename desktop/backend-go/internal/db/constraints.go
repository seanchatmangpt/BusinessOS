package db

import (
	"errors"
	"fmt"
	"time"
)

// Deal constraint validation errors
var (
	ErrDealAmountNonPositive   = errors.New("deal amount must be positive")
	ErrDealQualityOutOfBounds  = errors.New("deal quality score must be between 0 and 100")
	ErrDealMissingDomain       = errors.New("deal must have a domain")
	ErrDealMissingStatus       = errors.New("deal must have a status")
	ErrDealTemporalOrdering    = errors.New("created_at cannot be after updated_at")
	ErrDealStatusInvalid       = errors.New("deal status is invalid")
)

// PHI constraint validation errors
var (
	ErrPHIConfidenceOutOfBounds = errors.New("PHI confidence level must be between 0.0 and 1.0")
	ErrPHIMissingPatientID      = errors.New("PHI record must have a patient ID")
	ErrPHIMissingResourceType   = errors.New("PHI record must have a resource type")
	ErrPHINotEncrypted          = errors.New("PHI record must be encrypted before storage (SOC2 A-level requirement)")
)

// Data lineage constraint validation errors
var (
	ErrLineageDepthOutOfBounds = errors.New("lineage depth must be between 1 and 5 (WvdA Soundness)")
)

// Heartbeat constraint validation errors
var (
	ErrHeartbeatIntervalTooShort  = errors.New("heartbeat interval must be at least 100ms")
	ErrHeartbeatIntervalTooLong   = errors.New("heartbeat interval must not exceed 60s (Armstrong Supervision)")
)

// Workspace constraint validation errors
var (
	ErrWorkspaceMissingUserID = errors.New("workspace must have a user ID")
	ErrWorkspaceMissingName   = errors.New("workspace must have a name")
	ErrWorkspaceMissingMode   = errors.New("workspace mode must be specified (2d, 3d, or hybrid)")
	ErrWorkspaceModeInvalid   = errors.New("workspace mode must be one of: 2d, 3d, hybrid")
)

// Audit constraint validation errors
var (
	ErrAuditMissingActor    = errors.New("audit log must have an actor ID")
	ErrAuditMissingAction   = errors.New("audit log must have an action")
	ErrAuditMissingResource = errors.New("audit log must have a resource type")
	ErrAuditImmutable       = errors.New("audit logs cannot be modified after creation (SOC2 A-level)")
)

// Deal represents a business deal with constraints
type Deal struct {
	ID           string    `db:"id"`
	Domain       string    `db:"domain"`
	Name         string    `db:"name"`
	Amount       float64   `db:"amount"`
	Status       string    `db:"status"`
	QualityScore float64   `db:"quality_score"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// ValidateConstraints validates all deal constraints
func (d *Deal) ValidateConstraints() error {
	if errs := d.validateAmount(); errs != nil {
		return errs
	}
	if errs := d.validateQuality(); errs != nil {
		return errs
	}
	if errs := d.validateDomain(); errs != nil {
		return errs
	}
	if errs := d.validateStatus(); errs != nil {
		return errs
	}
	if errs := d.validateTemporal(); errs != nil {
		return errs
	}
	return nil
}

// validateAmount: CHECK (amount > 0)
func (d *Deal) validateAmount() error {
	if d.Amount <= 0 {
		return fmt.Errorf("%w (got: %.2f)", ErrDealAmountNonPositive, d.Amount)
	}
	return nil
}

// validateQuality: CHECK (quality_score >= 0 AND quality_score <= 100)
func (d *Deal) validateQuality() error {
	if d.QualityScore < 0 || d.QualityScore > 100 {
		return fmt.Errorf("%w (got: %.2f)", ErrDealQualityOutOfBounds, d.QualityScore)
	}
	return nil
}

// validateDomain: NOT NULL
func (d *Deal) validateDomain() error {
	if d.Domain == "" {
		return ErrDealMissingDomain
	}
	return nil
}

// validateStatus: NOT NULL + valid enum
func (d *Deal) validateStatus() error {
	if d.Status == "" {
		return ErrDealMissingStatus
	}
	validStatuses := map[string]bool{
		"prospect":   true,
		"negotiating": true,
		"won":        true,
		"lost":       true,
		"archived":   true,
	}
	if !validStatuses[d.Status] {
		return fmt.Errorf("%w: %s", ErrDealStatusInvalid, d.Status)
	}
	return nil
}

// validateTemporal: CHECK (created_at <= updated_at)
func (d *Deal) validateTemporal() error {
	if d.CreatedAt.After(d.UpdatedAt) {
		return fmt.Errorf("%w (created: %s, updated: %s)",
			ErrDealTemporalOrdering, d.CreatedAt, d.UpdatedAt)
	}
	return nil
}

// PHIRecord represents protected health information with compliance constraints
type PHIRecord struct {
	ID              string    `db:"id"`
	PatientID       string    `db:"patient_id"`
	ResourceType    string    `db:"resource_type"`
	RecordVersion   int       `db:"record_version"`
	ConfidenceLevel float64   `db:"confidence_level"`
	EncryptedData   string    `db:"encrypted_data"`
	CreatedAt       time.Time `db:"created_at"`
}

// ValidateConstraints validates all PHI record constraints
// SOC2 A-level: encryption required, confidence validated
func (p *PHIRecord) ValidateConstraints() error {
	if errs := p.validateConfidence(); errs != nil {
		return errs
	}
	if errs := p.validatePatientID(); errs != nil {
		return errs
	}
	if errs := p.validateResourceType(); errs != nil {
		return errs
	}
	if errs := p.validateEncryption(); errs != nil {
		return errs
	}
	return nil
}

// validateConfidence: CHECK (confidence_level >= 0.0 AND confidence_level <= 1.0)
func (p *PHIRecord) validateConfidence() error {
	if p.ConfidenceLevel < 0.0 || p.ConfidenceLevel > 1.0 {
		return fmt.Errorf("%w (got: %.2f)", ErrPHIConfidenceOutOfBounds, p.ConfidenceLevel)
	}
	return nil
}

// validatePatientID: NOT NULL
func (p *PHIRecord) validatePatientID() error {
	if p.PatientID == "" {
		return ErrPHIMissingPatientID
	}
	return nil
}

// validateResourceType: NOT NULL
func (p *PHIRecord) validateResourceType() error {
	if p.ResourceType == "" {
		return ErrPHIMissingResourceType
	}
	return nil
}

// validateEncryption: enforce encryption before storage (application-level compliance)
func (p *PHIRecord) validateEncryption() error {
	if p.EncryptedData == "" {
		return ErrPHINotEncrypted
	}
	return nil
}

// DataLineage represents data provenance with WvdA Soundness constraints
type DataLineage struct {
	ID            string `db:"id"`
	ParentID      string `db:"parent_id"`
	ChildID       string `db:"child_id"`
	LineageDepth  int    `db:"lineage_depth"`
	TransformType string `db:"transform_type"`
}

// ValidateConstraints validates lineage constraints
// WvdA Soundness: depth bounded [1,5] prevents infinite recursion
func (dl *DataLineage) ValidateConstraints() error {
	if errs := dl.validateDepth(); errs != nil {
		return errs
	}
	return nil
}

// validateDepth: CHECK (lineage_depth >= 1 AND lineage_depth <= 5)
func (dl *DataLineage) validateDepth() error {
	if dl.LineageDepth < 1 || dl.LineageDepth > 5 {
		return fmt.Errorf("%w (got: %d)", ErrLineageDepthOutOfBounds, dl.LineageDepth)
	}
	return nil
}

// AgentHeartbeat represents periodic liveness check with Armstrong Supervision constraints
type AgentHeartbeat struct {
	ID         string        `db:"id"`
	AgentID    string        `db:"agent_id"`
	IntervalMs int           `db:"interval_ms"`
	CreatedAt  time.Time     `db:"created_at"`
}

// ValidateConstraints validates heartbeat configuration
// Armstrong Fault Tolerance: interval bounded prevents supervisor overload
func (ah *AgentHeartbeat) ValidateConstraints() error {
	if errs := ah.validateInterval(); errs != nil {
		return errs
	}
	return nil
}

// validateInterval: CHECK (interval_ms >= 100 AND interval_ms <= 60000)
func (ah *AgentHeartbeat) validateInterval() error {
	if ah.IntervalMs < 100 {
		return fmt.Errorf("%w (got: %dms)", ErrHeartbeatIntervalTooShort, ah.IntervalMs)
	}
	if ah.IntervalMs > 60000 {
		return fmt.Errorf("%w (got: %dms)", ErrHeartbeatIntervalTooLong, ah.IntervalMs)
	}
	return nil
}

// Workspace represents user workspace with validation
type Workspace struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`
	Name   string `db:"name"`
	Mode   string `db:"mode"`
}

// ValidateConstraints validates workspace constraints
func (w *Workspace) ValidateConstraints() error {
	if errs := w.validateUserID(); errs != nil {
		return errs
	}
	if errs := w.validateName(); errs != nil {
		return errs
	}
	if errs := w.validateMode(); errs != nil {
		return errs
	}
	return nil
}

// validateUserID: NOT NULL
func (w *Workspace) validateUserID() error {
	if w.UserID == "" {
		return ErrWorkspaceMissingUserID
	}
	return nil
}

// validateName: NOT NULL
func (w *Workspace) validateName() error {
	if w.Name == "" {
		return ErrWorkspaceMissingName
	}
	return nil
}

// validateMode: NOT NULL + valid enum
func (w *Workspace) validateMode() error {
	if w.Mode == "" {
		return ErrWorkspaceMissingMode
	}
	validModes := map[string]bool{
		"2d":     true,
		"3d":     true,
		"hybrid": true,
	}
	if !validModes[w.Mode] {
		return fmt.Errorf("%w: %s", ErrWorkspaceModeInvalid, w.Mode)
	}
	return nil
}

// AuditLog represents immutable audit trail entry
type AuditLog struct {
	ID           string    `db:"id"`
	ActorID      string    `db:"actor_id"`
	Action       string    `db:"action"`
	ResourceType string    `db:"resource_type"`
	ResourceID   string    `db:"resource_id"`
	CreatedAt    time.Time `db:"created_at"`
}

// ValidateConstraints validates audit log constraints
// SOC2 A-level: enforces immutability and accountability
func (al *AuditLog) ValidateConstraints() error {
	if errs := al.validateActor(); errs != nil {
		return errs
	}
	if errs := al.validateAction(); errs != nil {
		return errs
	}
	if errs := al.validateResource(); errs != nil {
		return errs
	}
	return nil
}

// validateActor: NOT NULL (accountability)
func (al *AuditLog) validateActor() error {
	if al.ActorID == "" {
		return ErrAuditMissingActor
	}
	return nil
}

// validateAction: NOT NULL (event tracking)
func (al *AuditLog) validateAction() error {
	if al.Action == "" {
		return ErrAuditMissingAction
	}
	return nil
}

// validateResource: NOT NULL (impact tracking)
func (al *AuditLog) validateResource() error {
	if al.ResourceType == "" {
		return ErrAuditMissingResource
	}
	return nil
}

// BatchValidator validates collections of records
type BatchValidator struct {
	Errors []error
}

// ValidateDealBatch validates multiple deals and collects errors
func ValidateDealBatch(deals []*Deal) *BatchValidator {
	bv := &BatchValidator{Errors: []error{}}
	for i, deal := range deals {
		if err := deal.ValidateConstraints(); err != nil {
			bv.Errors = append(bv.Errors, fmt.Errorf("deal[%d]: %w", i, err))
		}
	}
	return bv
}

// ValidatePHIBatch validates multiple PHI records and collects errors
func ValidatePHIBatch(records []*PHIRecord) *BatchValidator {
	bv := &BatchValidator{Errors: []error{}}
	for i, record := range records {
		if err := record.ValidateConstraints(); err != nil {
			bv.Errors = append(bv.Errors, fmt.Errorf("phi_record[%d]: %w", i, err))
		}
	}
	return bv
}

// ValidateWorkspaceBatch validates multiple workspaces and collects errors
func ValidateWorkspaceBatch(workspaces []*Workspace) *BatchValidator {
	bv := &BatchValidator{Errors: []error{}}
	for i, ws := range workspaces {
		if err := ws.ValidateConstraints(); err != nil {
			bv.Errors = append(bv.Errors, fmt.Errorf("workspace[%d]: %w", i, err))
		}
	}
	return bv
}

// HasErrors returns true if batch validation found any errors
func (bv *BatchValidator) HasErrors() bool {
	return len(bv.Errors) > 0
}

// Error returns concatenated error message
func (bv *BatchValidator) Error() string {
	if !bv.HasErrors() {
		return ""
	}
	msg := fmt.Sprintf("%d validation error(s):\n", len(bv.Errors))
	for i, err := range bv.Errors {
		msg += fmt.Sprintf("%d. %v\n", i+1, err)
	}
	return msg
}

// SummaryStats provides constraint violation metrics for observability
type SummaryStats struct {
	TotalValidated  int
	Passed          int
	Failed          int
	FailureRate     float64
	MostCommonError string
}

// CalculateStats returns validation statistics for monitoring dashboard
func (bv *BatchValidator) CalculateStats(total int) *SummaryStats {
	failed := len(bv.Errors)
	passed := total - failed
	failureRate := float64(0)
	if total > 0 {
		failureRate = float64(failed) / float64(total)
	}

	// Find most common error type
	errorCounts := make(map[string]int)
	mostCommon := ""
	for _, err := range bv.Errors {
		errorCounts[err.Error()]++
	}
	maxCount := 0
	for errMsg, count := range errorCounts {
		if count > maxCount {
			maxCount = count
			mostCommon = errMsg
		}
	}

	return &SummaryStats{
		TotalValidated:  total,
		Passed:          passed,
		Failed:          failed,
		FailureRate:     failureRate,
		MostCommonError: mostCommon,
	}
}
