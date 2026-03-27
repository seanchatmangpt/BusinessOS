-- Migration: Create healthcare PHI (Protected Health Information) tables
-- HIPAA-compliant storage for patient health records with encryption and consent management

CREATE TABLE IF NOT EXISTS phi_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    patient_id VARCHAR(100) NOT NULL,
    patient_id_hash VARCHAR(256) NOT NULL,
    resource_type VARCHAR(50) NOT NULL CHECK (resource_type IN (
        'Patient', 'Observation', 'Medication', 'MedicationAdministration',
        'Condition', 'DiagnosticReport', 'Procedure', 'Encounter',
        'AllergyIntolerance', 'Immunization', 'CarePlan', 'CareTeam',
        'DocumentReference', 'Goal', 'HealthcareService', 'Organization'
    )),
    data_hash VARCHAR(256) NOT NULL,
    data_encrypted BYTEA NOT NULL,
    encryption_algorithm VARCHAR(50) DEFAULT 'AES-256-GCM',
    data_classification VARCHAR(20) CHECK (data_classification IN ('de_identified', 'limited_dataset', 'phi')),
    consent_status VARCHAR(20) NOT NULL CHECK (consent_status IN ('granted', 'denied', 'pending', 'revoked')),
    consent_date TIMESTAMP WITH TIME ZONE,
    consent_type VARCHAR(50) CHECK (consent_type IN ('treatment', 'payment', 'operations', 'research', 'marketing')),
    consent_version INT DEFAULT 1,
    purpose_of_use VARCHAR(100),
    authorized_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    accessed_at TIMESTAMP WITH TIME ZONE,
    accessed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    access_reason VARCHAR(200),
    export_allowed BOOLEAN DEFAULT false,
    secondary_use_allowed BOOLEAN DEFAULT false,
    retention_end_date DATE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT patient_id_not_empty CHECK (patient_id != ''),
    CONSTRAINT created_before_accessed CHECK (accessed_at IS NULL OR accessed_at >= created_at),
    CONSTRAINT retention_after_created CHECK (retention_end_date IS NULL OR retention_end_date > CAST(created_at AS DATE))
);

-- Indexes for HIPAA compliance and query performance
CREATE INDEX IF NOT EXISTS idx_phi_patient_id_hash ON phi_records(patient_id_hash);
CREATE INDEX IF NOT EXISTS idx_phi_resource_type ON phi_records(resource_type);
CREATE INDEX IF NOT EXISTS idx_phi_created_at ON phi_records(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_phi_deleted_at ON phi_records(deleted_at);
CREATE INDEX IF NOT EXISTS idx_phi_consent_status ON phi_records(consent_status);
CREATE INDEX IF NOT EXISTS idx_phi_accessed_at ON phi_records(accessed_at DESC);
CREATE INDEX IF NOT EXISTS idx_phi_patient_resource ON phi_records(patient_id_hash, resource_type);
CREATE INDEX IF NOT EXISTS idx_phi_retention ON phi_records(retention_end_date);
CREATE INDEX IF NOT EXISTS idx_phi_data_classification ON phi_records(data_classification);

-- Table comments
COMMENT ON TABLE phi_records IS 'HIPAA-compliant storage for Protected Health Information with encryption and audit trail';
COMMENT ON COLUMN phi_records.patient_id_hash IS 'One-way hash of patient ID for privacy (not reversible)';
COMMENT ON COLUMN phi_records.data_hash IS 'SHA-256 hash of unencrypted data for integrity verification';
COMMENT ON COLUMN phi_records.data_encrypted IS 'Encrypted PHI data blob (decrypted only when authorized)';
COMMENT ON COLUMN phi_records.encryption_algorithm IS 'Algorithm used for encryption (AES-256-GCM standard)';
COMMENT ON COLUMN phi_records.consent_status IS 'Patient consent state for this specific resource';
COMMENT ON COLUMN phi_records.access_reason IS 'Clinical or administrative reason for data access';
COMMENT ON COLUMN phi_records.secondary_use_allowed IS 'Whether data can be used for research, analytics, etc.';
COMMENT ON COLUMN phi_records.deleted_at IS 'Soft-delete timestamp (hard deletion after retention period)';

-- Create phi_audit_log for HIPAA compliance and breach detection
CREATE TABLE IF NOT EXISTS phi_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phi_record_id UUID NOT NULL REFERENCES phi_records(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL CHECK (action IN ('create', 'read', 'update', 'delete', 'export', 'access_denied')),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_type VARCHAR(50) CHECK (actor_type IN ('user', 'system', 'api', 'batch_job')),
    action_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    result VARCHAR(20) CHECK (result IN ('success', 'failure', 'denied')),
    reason_code VARCHAR(100),
    system_change JSONB,
    retention_until TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '7 years'),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for audit trail queries and compliance reporting
CREATE INDEX IF NOT EXISTS idx_phi_audit_phi_record ON phi_audit_log(phi_record_id);
CREATE INDEX IF NOT EXISTS idx_phi_audit_action ON phi_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_phi_audit_timestamp ON phi_audit_log(action_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_phi_audit_actor ON phi_audit_log(actor_id);
CREATE INDEX IF NOT EXISTS idx_phi_audit_result ON phi_audit_log(result);
CREATE INDEX IF NOT EXISTS idx_phi_audit_retention ON phi_audit_log(retention_until);

COMMENT ON TABLE phi_audit_log IS 'Complete audit trail of all access, modifications, and denials of PHI for HIPAA compliance';
COMMENT ON COLUMN phi_audit_log.actor_type IS 'Type of actor performing the action (user, automated system, API service, etc.)';
COMMENT ON COLUMN phi_audit_log.reason_code IS 'Structured code for reason (e.g., TREATMENT, PAYMENT, RESEARCH_CONSENTED)';
COMMENT ON COLUMN phi_audit_log.system_change IS 'Before/after data for update actions (for change tracking)';

-- Create phi_breach_detection for monitoring unusual access patterns
CREATE TABLE IF NOT EXISTS phi_breach_detection (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_type VARCHAR(50) NOT NULL CHECK (alert_type IN (
        'mass_access', 'after_hours', 'access_denied_pattern', 'unusual_export',
        'unauthorized_secondary_use', 'geographic_anomaly', 'encryption_failure'
    )),
    severity VARCHAR(20) CHECK (severity IN ('info', 'warning', 'critical')),
    description TEXT NOT NULL,
    affected_records INT,
    affected_patients INT,
    detected_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    affected_phi_ids UUID[],
    investigation_status VARCHAR(20) DEFAULT 'open' CHECK (investigation_status IN ('open', 'investigating', 'resolved', 'false_positive')),
    investigation_notes TEXT,
    remediation_steps TEXT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_breach_alert_type ON phi_breach_detection(alert_type);
CREATE INDEX IF NOT EXISTS idx_breach_severity ON phi_breach_detection(severity);
CREATE INDEX IF NOT EXISTS idx_breach_detected_at ON phi_breach_detection(detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_breach_status ON phi_breach_detection(investigation_status);

COMMENT ON TABLE phi_breach_detection IS 'Detects potential HIPAA breaches through anomaly detection on access patterns';

-- Create phi_consent_log for tracking consent changes over time
CREATE TABLE IF NOT EXISTS phi_consent_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phi_record_id UUID NOT NULL REFERENCES phi_records(id) ON DELETE CASCADE,
    previous_consent VARCHAR(20),
    new_consent VARCHAR(20) NOT NULL CHECK (new_consent IN ('granted', 'denied', 'pending', 'revoked')),
    previous_types VARCHAR(50)[],
    new_types VARCHAR(50)[],
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    change_reason TEXT,
    patient_signed_date DATE,
    signature_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_consent_phi_record ON phi_consent_log(phi_record_id);
CREATE INDEX IF NOT EXISTS idx_consent_changed_at ON phi_consent_log(changed_at DESC);
CREATE INDEX IF NOT EXISTS idx_consent_status ON phi_consent_log(new_consent);

COMMENT ON TABLE phi_consent_log IS 'Historical log of all consent status changes for compliance auditing';

-- Trigger for updated_at on phi_records
CREATE OR REPLACE FUNCTION update_phi_records_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_phi_records_updated_at ON phi_records;
CREATE TRIGGER trigger_phi_records_updated_at
BEFORE UPDATE ON phi_records
FOR EACH ROW
EXECUTE FUNCTION update_phi_records_updated_at();
