-- Migration: Create compliance controls table for multi-framework governance
-- Supports SOC2, GDPR, HIPAA, SOX compliance control definitions

CREATE TABLE IF NOT EXISTS compliance_controls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    framework VARCHAR(50) NOT NULL CHECK (framework IN ('SOC2', 'GDPR', 'HIPAA', 'SOX', 'PCI-DSS', 'ISO27001', 'CUSTOM')),
    control_id VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(10) NOT NULL CHECK (severity IN ('critical', 'high', 'medium', 'low')),
    control_type VARCHAR(50) CHECK (control_type IN ('preventive', 'detective', 'corrective', 'compensating')),
    enabled BOOLEAN DEFAULT true,
    responsible_team VARCHAR(200),
    test_frequency VARCHAR(50) CHECK (test_frequency IN ('continuous', 'daily', 'weekly', 'monthly', 'quarterly', 'annually', 'ad_hoc')),
    last_tested_at TIMESTAMP WITH TIME ZONE,
    last_tested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    test_result VARCHAR(20) CHECK (test_result IN ('pass', 'fail', 'pending', 'waived')),
    remediation_notes TEXT,
    evidence_location VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT control_id_unique UNIQUE(framework, control_id)
);

-- Indexes for querying controls
CREATE INDEX IF NOT EXISTS idx_compliance_controls_framework ON compliance_controls(framework);
CREATE INDEX IF NOT EXISTS idx_compliance_controls_control_id ON compliance_controls(control_id);
CREATE INDEX IF NOT EXISTS idx_compliance_controls_severity ON compliance_controls(severity);
CREATE INDEX IF NOT EXISTS idx_compliance_controls_enabled ON compliance_controls(enabled);
CREATE INDEX IF NOT EXISTS idx_compliance_controls_framework_severity ON compliance_controls(framework, severity);
CREATE INDEX IF NOT EXISTS idx_compliance_controls_last_tested ON compliance_controls(last_tested_at DESC);
CREATE INDEX IF NOT EXISTS idx_compliance_controls_test_result ON compliance_controls(test_result);

-- Table comments
COMMENT ON TABLE compliance_controls IS 'Master control catalog for multi-framework compliance governance (SOC2, GDPR, HIPAA, SOX)';
COMMENT ON COLUMN compliance_controls.framework IS 'Compliance framework identifier';
COMMENT ON COLUMN compliance_controls.control_id IS 'Framework-specific control identifier (e.g., CC6.1 for SOC2)';
COMMENT ON COLUMN compliance_controls.control_type IS 'Control operational approach';
COMMENT ON COLUMN compliance_controls.test_frequency IS 'Required testing cadence';
COMMENT ON COLUMN compliance_controls.test_result IS 'Latest test outcome from testing';
COMMENT ON COLUMN compliance_controls.evidence_location IS 'URI or reference to control evidence artifacts';

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_compliance_controls_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_compliance_controls_updated_at ON compliance_controls;
CREATE TRIGGER trigger_compliance_controls_updated_at
BEFORE UPDATE ON compliance_controls
FOR EACH ROW
EXECUTE FUNCTION update_compliance_controls_updated_at();

-- Create compliance_control_mappings for cross-framework mapping (e.g., SOC2 CC6.1 maps to GDPR Article 32)
CREATE TABLE IF NOT EXISTS compliance_control_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_control_id UUID NOT NULL REFERENCES compliance_controls(id) ON DELETE CASCADE,
    target_control_id UUID NOT NULL REFERENCES compliance_controls(id) ON DELETE CASCADE,
    mapping_type VARCHAR(50) CHECK (mapping_type IN ('equivalent', 'related', 'supports', 'overlaps')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT no_self_mapping CHECK (source_control_id != target_control_id),
    CONSTRAINT unique_mapping UNIQUE(source_control_id, target_control_id)
);

CREATE INDEX IF NOT EXISTS idx_control_mappings_source ON compliance_control_mappings(source_control_id);
CREATE INDEX IF NOT EXISTS idx_control_mappings_target ON compliance_control_mappings(target_control_id);

COMMENT ON TABLE compliance_control_mappings IS 'Maps controls across frameworks to show equivalence and overlap';
