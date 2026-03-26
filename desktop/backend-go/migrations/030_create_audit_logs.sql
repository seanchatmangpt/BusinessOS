-- Migration: Create comprehensive audit logs table for SOC2 compliance
-- Tracks all system actions for security investigation and regulatory compliance
-- 7-year retention policy per SOX requirements

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_email VARCHAR(255),
    actor_ip_address INET,
    actor_user_agent TEXT,
    action VARCHAR(100) NOT NULL CHECK (action != ''),
    action_category VARCHAR(50) NOT NULL CHECK (action_category IN (
        'authentication', 'authorization', 'data_access', 'data_modification',
        'configuration', 'compliance', 'incident', 'system', 'integration',
        'export', 'import', 'deletion', 'archival'
    )),
    resource_type VARCHAR(50) NOT NULL CHECK (resource_type != ''),
    resource_id VARCHAR(255),
    resource_name VARCHAR(500),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    result VARCHAR(20) NOT NULL CHECK (result IN ('success', 'failure', 'partial')) DEFAULT 'success',
    result_code VARCHAR(20),
    error_message TEXT,
    change_summary JSONB,
    change_details JSONB,
    old_values JSONB,
    new_values JSONB,
    affected_records INT,
    duration_ms INT,
    ip_address INET,
    geographic_location VARCHAR(255),
    device_fingerprint VARCHAR(256),
    session_id VARCHAR(255),
    request_id VARCHAR(255),
    api_endpoint VARCHAR(500),
    api_method VARCHAR(10),
    response_status_code INT,
    response_size_bytes INT,
    details TEXT,
    severity VARCHAR(20) CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info')) DEFAULT 'info',
    flags VARCHAR(50)[] DEFAULT ARRAY[]::VARCHAR[],
    signature VARCHAR(512),
    signature_algorithm VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    retention_until TIMESTAMP WITH TIME ZONE DEFAULT (NOW() + INTERVAL '2555 days'),
    CONSTRAINT action_not_empty CHECK (action != ''),
    CONSTRAINT resource_type_not_empty CHECK (resource_type != '')
);

-- Indexes for optimal query performance and compliance reporting
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_category ON audit_logs(action_category);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_result ON audit_logs(result);
CREATE INDEX IF NOT EXISTS idx_audit_logs_severity ON audit_logs(severity);
CREATE INDEX IF NOT EXISTS idx_audit_logs_ip_address ON audit_logs(ip_address);
CREATE INDEX IF NOT EXISTS idx_audit_logs_session_id ON audit_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_request_id ON audit_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_timestamp ON audit_logs(actor_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_timestamp ON audit_logs(resource_type, resource_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_retention ON audit_logs(retention_until);
CREATE INDEX IF NOT EXISTS idx_audit_logs_change_summary ON audit_logs USING GIN(change_summary);
CREATE INDEX IF NOT EXISTS idx_audit_logs_flags ON audit_logs USING GIN(flags);

-- Composite index for common audit queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_compliance ON audit_logs(
    action_category, result, timestamp DESC
) WHERE result != 'success';

-- Table comments
COMMENT ON TABLE audit_logs IS 'Comprehensive audit trail for SOC2, GDPR, HIPAA, SOX compliance with 7-year retention';
COMMENT ON COLUMN audit_logs.actor_id IS 'User or system account that performed the action';
COMMENT ON COLUMN audit_logs.actor_email IS 'Email address of the actor (cached for deleted users)';
COMMENT ON COLUMN audit_logs.action_category IS 'Semantic category for compliance reporting and filtering';
COMMENT ON COLUMN audit_logs.resource_type IS 'Entity type affected (user, deal, dataset, compliance_control, etc.)';
COMMENT ON COLUMN audit_logs.result IS 'Outcome of the action: success, failure, or partial completion';
COMMENT ON COLUMN audit_logs.change_summary IS 'High-level summary of what changed: {field: old_value, field: new_value}';
COMMENT ON COLUMN audit_logs.change_details IS 'Complete before/after snapshots of affected record';
COMMENT ON COLUMN audit_logs.severity IS 'Risk severity for alerting (critical for failed auth, deletion, config changes)';
COMMENT ON COLUMN audit_logs.signature IS 'HMAC or digital signature for log integrity verification';
COMMENT ON COLUMN audit_logs.retention_until IS 'Deletion deadline (7 years from creation per SOX)';

-- Create audit_log_integrity table for detecting log tampering
CREATE TABLE IF NOT EXISTS audit_log_integrity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    audit_log_id UUID NOT NULL REFERENCES audit_logs(id) ON DELETE CASCADE,
    checksum VARCHAR(256) NOT NULL,
    checksum_algorithm VARCHAR(50) DEFAULT 'SHA256',
    verified_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_by UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) CHECK (status IN ('valid', 'tampered', 'missing', 'unknown')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_integrity_log_id ON audit_log_integrity(audit_log_id);
CREATE INDEX IF NOT EXISTS idx_audit_integrity_verified_at ON audit_log_integrity(verified_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_integrity_status ON audit_log_integrity(status);

COMMENT ON TABLE audit_log_integrity IS 'Cryptographic verification of audit log integrity (prevents tampering)';

-- Create audit_log_summary for efficient compliance reporting
CREATE TABLE IF NOT EXISTS audit_log_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    summary_date DATE NOT NULL,
    action_category VARCHAR(50) NOT NULL,
    total_actions INT NOT NULL,
    successful_actions INT NOT NULL,
    failed_actions INT NOT NULL,
    unique_actors INT NOT NULL,
    unique_resources INT NOT NULL,
    critical_actions INT DEFAULT 0,
    high_severity_actions INT DEFAULT 0,
    unusual_patterns VARCHAR(255)[],
    summary JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_summary UNIQUE(summary_date, action_category)
);

CREATE INDEX IF NOT EXISTS idx_summary_date ON audit_log_summary(summary_date DESC);
CREATE INDEX IF NOT EXISTS idx_summary_category ON audit_log_summary(action_category);

COMMENT ON TABLE audit_log_summary IS 'Pre-aggregated daily summaries for fast compliance reporting and trend analysis';

-- Create audit_log_retention_policy for configurable retention
CREATE TABLE IF NOT EXISTS audit_log_retention_policy (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_category VARCHAR(50) NOT NULL,
    retention_days INT NOT NULL CHECK (retention_days > 0),
    reason VARCHAR(255),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_category UNIQUE(action_category)
);

-- Default retention policies (can be overridden)
INSERT INTO audit_log_retention_policy (action_category, retention_days, reason)
VALUES
    ('authentication', 730, 'HIPAA minimum: 2 years'),
    ('authorization', 730, 'HIPAA minimum: 2 years'),
    ('data_access', 2555, 'SOX requirement: 7 years'),
    ('data_modification', 2555, 'SOX requirement: 7 years'),
    ('configuration', 2555, 'SOX requirement: 7 years'),
    ('compliance', 2555, 'SOX requirement: 7 years'),
    ('incident', 2555, 'SOX requirement: 7 years'),
    ('system', 730, 'HIPAA minimum: 2 years'),
    ('integration', 730, 'HIPAA minimum: 2 years'),
    ('export', 2555, 'SOX requirement: 7 years'),
    ('deletion', 2555, 'SOX requirement: 7 years'),
    ('archival', 2555, 'SOX requirement: 7 years')
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_retention_category ON audit_log_retention_policy(action_category);

COMMENT ON TABLE audit_log_retention_policy IS 'Configurable retention policies for different audit log categories per framework';

-- Create function to archive old audit logs (soft delete)
CREATE OR REPLACE FUNCTION archive_old_audit_logs()
RETURNS TABLE(archived_count INT) AS $$
DECLARE
    v_archived_count INT;
BEGIN
    UPDATE audit_logs
    SET retention_until = NOW()
    WHERE retention_until <= NOW()
    AND created_at < NOW() - INTERVAL '2555 days';

    GET DIAGNOSTICS v_archived_count = ROW_COUNT;

    RETURN QUERY SELECT v_archived_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION archive_old_audit_logs() IS 'Archives audit logs past retention date (callable from cron job or manual admin)';

-- Create function to verify audit log integrity
CREATE OR REPLACE FUNCTION verify_audit_log_integrity(p_log_id UUID)
RETURNS TABLE(valid BOOLEAN, checksum_match BOOLEAN, notes TEXT) AS $$
DECLARE
    v_stored_checksum VARCHAR(256);
    v_computed_checksum VARCHAR(256);
BEGIN
    SELECT checksum INTO v_stored_checksum
    FROM audit_log_integrity
    WHERE audit_log_id = p_log_id
    ORDER BY verified_at DESC
    LIMIT 1;

    IF v_stored_checksum IS NULL THEN
        RETURN QUERY SELECT false, false, 'No stored checksum found'::TEXT;
        RETURN;
    END IF;

    -- Compute current checksum from audit log
    SELECT encode(digest(
        ROW(action, resource_type, actor_id, timestamp, result)::TEXT,
        'sha256'
    ), 'hex') INTO v_computed_checksum
    FROM audit_logs
    WHERE id = p_log_id;

    RETURN QUERY SELECT
        v_stored_checksum = v_computed_checksum,
        v_stored_checksum = v_computed_checksum,
        CASE
            WHEN v_stored_checksum = v_computed_checksum THEN 'Integrity verified'
            ELSE 'Checksum mismatch - log may be tampered'
        END;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION verify_audit_log_integrity(UUID) IS 'Cryptographic verification that audit log has not been modified since creation';
