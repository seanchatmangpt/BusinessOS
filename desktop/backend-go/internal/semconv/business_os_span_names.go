package semconv

const (
	// business_os_audit_record is the span name for "business_os.audit.record".
	//
	// Recording an audit event in the BusinessOS immutable audit trail.
	// Kind: internal
	// Stability: development
	BusinessOsAuditRecord = "business_os.audit.record"
	// business_os_compliance_check is the span name for "business_os.compliance.check".
	//
	// Evaluating a SOC2/HIPAA/GDPR compliance rule against current system state.
	// Kind: internal
	// Stability: development
	BusinessOsComplianceCheck = "business_os.compliance.check"
)