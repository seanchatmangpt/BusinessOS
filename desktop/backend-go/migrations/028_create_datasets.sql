-- Migration: Create datasets table for data mesh and lineage tracking
-- Supports Finance, Operations, Marketing, Sales, HR domains
-- Tracks data quality, lineage depth, and ownership

CREATE TABLE IF NOT EXISTS datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    domain VARCHAR(50) NOT NULL CHECK (domain IN ('Finance', 'Operations', 'Marketing', 'Sales', 'HR', 'Legal', 'Engineering', 'Other')),
    name VARCHAR(500) NOT NULL,
    description TEXT,
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    quality_score SMALLINT DEFAULT 0 CHECK (quality_score >= 0 AND quality_score <= 100),
    lineage_depth SMALLINT DEFAULT 0 CHECK (lineage_depth >= 0 AND lineage_depth <= 10),
    source_count INT DEFAULT 0 CHECK (source_count >= 0),
    dependent_count INT DEFAULT 0 CHECK (dependent_count >= 0),
    record_count BIGINT DEFAULT 0,
    size_bytes BIGINT DEFAULT 0,
    sensitivity_level VARCHAR(20) CHECK (sensitivity_level IN ('public', 'internal', 'confidential', 'restricted')),
    pii_present BOOLEAN DEFAULT false,
    pii_types VARCHAR(200)[],
    last_ingested_at TIMESTAMP WITH TIME ZONE,
    last_ingested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    retention_days INT CHECK (retention_days IS NULL OR retention_days > 0),
    retention_policy VARCHAR(50) CHECK (retention_policy IN ('permanent', 'tiered', 'delete_after', 'archive_after')),
    tags VARCHAR(100)[] DEFAULT ARRAY[]::VARCHAR[],
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for dataset discovery and management
CREATE INDEX IF NOT EXISTS idx_datasets_domain ON datasets(domain);
CREATE INDEX IF NOT EXISTS idx_datasets_owner_id ON datasets(owner_id);
CREATE INDEX IF NOT EXISTS idx_datasets_quality_score ON datasets(quality_score DESC);
CREATE INDEX IF NOT EXISTS idx_datasets_lineage_depth ON datasets(lineage_depth);
CREATE INDEX IF NOT EXISTS idx_datasets_created_at ON datasets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_datasets_sensitivity ON datasets(sensitivity_level);
CREATE INDEX IF NOT EXISTS idx_datasets_pii ON datasets(pii_present);
CREATE INDEX IF NOT EXISTS idx_datasets_tags ON datasets USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_datasets_metadata ON datasets USING GIN(metadata);
CREATE INDEX IF NOT EXISTS idx_datasets_name_search ON datasets(name);

-- Table comments
COMMENT ON TABLE datasets IS 'Data mesh dataset registry with lineage tracking and quality metrics';
COMMENT ON COLUMN datasets.domain IS 'Data domain classification for organization';
COMMENT ON COLUMN datasets.quality_score IS 'Data quality score (0-100): completeness, accuracy, consistency';
COMMENT ON COLUMN datasets.lineage_depth IS 'Number of transformation steps in data lineage';
COMMENT ON COLUMN datasets.sensitivity_level IS 'Data classification level for access control';
COMMENT ON COLUMN datasets.pii_present IS 'Flag indicating presence of personally identifiable information';
COMMENT ON COLUMN datasets.pii_types IS 'Array of PII data types present (email, ssn, phone, address, etc.)';
COMMENT ON COLUMN datasets.retention_policy IS 'Data retention and archival strategy';
COMMENT ON COLUMN datasets.metadata IS 'Flexible JSONB for dataset-specific metadata';

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_datasets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_datasets_updated_at ON datasets;
CREATE TRIGGER trigger_datasets_updated_at
BEFORE UPDATE ON datasets
FOR EACH ROW
EXECUTE FUNCTION update_datasets_updated_at();

-- Create dataset_lineage table for tracking data dependencies
CREATE TABLE IF NOT EXISTS dataset_lineage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    target_dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    transformation_type VARCHAR(50) CHECK (transformation_type IN ('copy', 'aggregate', 'join', 'filter', 'enrich', 'custom')),
    transformation_description TEXT,
    transformation_code TEXT,
    last_executed_at TIMESTAMP WITH TIME ZONE,
    execution_duration_ms INT,
    status VARCHAR(20) CHECK (status IN ('success', 'failed', 'pending', 'skipped')),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT no_self_lineage CHECK (source_dataset_id != target_dataset_id),
    CONSTRAINT unique_lineage UNIQUE(source_dataset_id, target_dataset_id)
);

CREATE INDEX IF NOT EXISTS idx_lineage_source ON dataset_lineage(source_dataset_id);
CREATE INDEX IF NOT EXISTS idx_lineage_target ON dataset_lineage(target_dataset_id);
CREATE INDEX IF NOT EXISTS idx_lineage_executed_at ON dataset_lineage(last_executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_lineage_status ON dataset_lineage(status);

COMMENT ON TABLE dataset_lineage IS 'Tracks data transformations and dependencies between datasets';
COMMENT ON COLUMN dataset_lineage.transformation_type IS 'Category of data transformation applied';
COMMENT ON COLUMN dataset_lineage.status IS 'Status of the most recent transformation execution';

-- Create dataset_quality_metrics for tracking quality dimensions
CREATE TABLE IF NOT EXISTS dataset_quality_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    metric_name VARCHAR(100) NOT NULL,
    metric_type VARCHAR(50) CHECK (metric_type IN ('completeness', 'accuracy', 'consistency', 'timeliness', 'validity', 'uniqueness')),
    value NUMERIC(10, 4) NOT NULL,
    threshold NUMERIC(10, 4),
    status VARCHAR(20) CHECK (status IN ('pass', 'warning', 'fail')),
    measured_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    measured_by UUID REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT
);

CREATE INDEX IF NOT EXISTS idx_quality_metrics_dataset ON dataset_quality_metrics(dataset_id);
CREATE INDEX IF NOT EXISTS idx_quality_metrics_type ON dataset_quality_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_quality_metrics_measured_at ON dataset_quality_metrics(measured_at DESC);

COMMENT ON TABLE dataset_quality_metrics IS 'Detailed quality measurements for datasets across six dimensions';
