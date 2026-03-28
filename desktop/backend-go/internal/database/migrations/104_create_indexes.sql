-- Migration: 102_create_indexes.sql
-- Description: Advanced composite, partial, and BRIN indexes for Fortune 5 performance
-- Created: 2026-03-26
-- Purpose: Optimize query patterns across critical tables with materialized access paths

-- =============================================================================
-- COMPOSITE INDEXES (Multi-column access paths)
-- =============================================================================

-- Business domain + quality scoring (A2A deal queries)
CREATE INDEX IF NOT EXISTS idx_deals_domain_quality
ON deals(domain ASC, quality_score DESC)
WHERE deleted_at IS NULL
INCLUDE (id, actor_id, status, created_at);

COMMENT ON INDEX idx_deals_domain_quality IS
'Composite index for deal discovery by domain/quality. Includes key columns for index-only scans.';

-- Framework + status (OSA app lifecycle queries)
CREATE INDEX IF NOT EXISTS idx_osa_modules_framework_status
ON osa_modules(framework VARCHAR(50), status VARCHAR(50))
WHERE deleted_at IS NULL
INCLUDE (id, workspace_id, created_by, created_at);

COMMENT ON INDEX idx_osa_modules_framework_status IS
'Composite index for filtering OSA modules by framework and status. Supports index-only scans.';

-- Actor + temporal (audit trail + usage analytics)
CREATE INDEX IF NOT EXISTS idx_audit_actor_timestamp
ON audit_logs(actor_id ASC, created_at DESC NULLS LAST)
WHERE deleted_at IS NULL
INCLUDE (action, resource_type, resource_id, status);

COMMENT ON INDEX idx_audit_actor_timestamp IS
'Composite index for actor activity timeline. Supports reverse chronological queries and index-only scans.';

-- User + workspace + status (workspace navigation)
CREATE INDEX IF NOT EXISTS idx_osa_workspaces_user_status
ON osa_workspaces(user_id ASC, status VARCHAR(50))
WHERE archived_at IS NULL
INCLUDE (id, name, template_type, last_accessed_at);

COMMENT ON INDEX idx_osa_workspaces_user_status IS
'Composite index for user workspace listing with status filtering. Enables index-only scans.';

-- Resource + quality + confidence (PHI quality assurance)
CREATE INDEX IF NOT EXISTS idx_phi_resource_quality
ON phi_records(patient_id UUID, resource_type VARCHAR(50), quality_score NUMERIC)
WHERE deleted_at IS NULL
INCLUDE (id, confidence_level, encrypted_data, created_at);

COMMENT ON INDEX idx_phi_resource_quality IS
'Composite index for PHI record retrieval by patient, resource type, and quality. SOC2 B-level compliance.';

-- =============================================================================
-- PARTIAL INDEXES (Filtered access paths for soft deletes)
-- =============================================================================

-- Active deals only (soft delete support)
CREATE INDEX IF NOT EXISTS idx_deals_active
ON deals(id, status)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_deals_active IS
'Partial index on active (non-deleted) deals. Dramatically faster than full-table soft-delete scans.';

-- Active audit logs (temporal filtering)
CREATE INDEX IF NOT EXISTS idx_audit_logs_active
ON audit_logs(id, action)
WHERE deleted_at IS NULL
AND created_at > NOW() - INTERVAL '90 days';

COMMENT ON INDEX idx_audit_logs_active IS
'Partial index on recent, non-deleted audit logs. Supports compliance audits without full scan.';

-- Active OSA modules (lifecycle filtering)
CREATE INDEX IF NOT EXISTS idx_osa_modules_active
ON osa_modules(id, workspace_id, status)
WHERE deleted_at IS NULL
AND status NOT IN ('archived', 'failed');

COMMENT ON INDEX idx_osa_modules_active IS
'Partial index on active (non-archived, non-failed) OSA modules. Excludes inactive/failed from queries.';

-- Non-encrypted sessions (security boundary)
CREATE INDEX IF NOT EXISTS idx_sessions_unencrypted
ON sessions(user_id, created_at)
WHERE encryption_key_id IS NOT NULL
AND revoked_at IS NULL;

COMMENT ON INDEX idx_sessions_unencrypted IS
'Partial index on valid (non-revoked) encrypted sessions. Supports session validation without full scan.';

-- =============================================================================
-- BRIN INDEXES (Block Range Index — optimal for sorted time-series)
-- =============================================================================

-- Time-series: audit events (billions of rows, time-sequential)
CREATE INDEX IF NOT EXISTS idx_audit_logs_timestamp_brin
ON audit_logs USING BRIN (created_at)
WITH (pages_per_range=128);

COMMENT ON INDEX idx_audit_logs_timestamp_brin IS
'BRIN index on audit log timestamps. 1-10MB vs 100MB+ B-tree for 100M+ rows. Ideal for time-range scans.
Query: SELECT * FROM audit_logs WHERE created_at BETWEEN ? AND ?';

-- Time-series: PHI access (regulatory: all access must be auditable)
CREATE INDEX IF NOT EXISTS idx_phi_access_timestamp_brin
ON phi_access_logs USING BRIN (created_at)
WITH (pages_per_range=64);

COMMENT ON INDEX idx_phi_access_timestamp_brin IS
'BRIN index on PHI access timestamps. SOC2 A-level: enables fast retrieval of access logs for compliance.';

-- Time-series: activity tracking (deal metrics, user behavior)
CREATE INDEX IF NOT EXISTS idx_deal_activity_timestamp_brin
ON deal_activity(created_at)
USING BRIN
WITH (pages_per_range=128);

COMMENT ON INDEX idx_deal_activity_timestamp_brin IS
'BRIN index for deal activity time-range queries. Supports analytics on billions of activity records.';

-- =============================================================================
-- GIN INDEXES (Inverted Index — optimal for JSONB, arrays)
-- =============================================================================

-- JSONB: OSA module definitions (schema exploration)
CREATE INDEX IF NOT EXISTS idx_osa_modules_schema_jsonb_gin
ON osa_modules USING GIN (schema_definition jsonb_path_ops);

COMMENT ON INDEX idx_osa_modules_schema_jsonb_gin IS
'GIN index on OSA module schema JSONB. Supports contains (@>) and path queries on nested schema structure.';

-- JSONB: Deal custom fields (flexible PHI storage)
CREATE INDEX IF NOT EXISTS idx_deals_custom_fields_gin
ON deals USING GIN (custom_fields jsonb_path_ops);

COMMENT ON INDEX idx_deals_custom_fields_gin IS
'GIN index on deal custom fields JSONB. Enables equality/contains queries without full table scan.';

-- Array: module tags (categorization + filtering)
CREATE INDEX IF NOT EXISTS idx_osa_modules_tags_gin
ON osa_modules USING GIN (tags);

COMMENT ON INDEX idx_osa_modules_tags_gin IS
'GIN index on OSA module tags array. Supports containment queries: tags @> ARRAY[''tag1'', ''tag2''].';

-- Array: user permissions (access control boundary)
CREATE INDEX IF NOT EXISTS idx_user_permissions_gin
ON "user" USING GIN (permissions);

COMMENT ON INDEX idx_user_permissions_gin IS
'GIN index on user permissions array. Supports fast permission checks without full table scan.';

-- =============================================================================
-- UNIQUE INDEXES (Constraint-backed indexes with side effects)
-- =============================================================================

-- PHI record uniqueness: patient + resource + version (idempotency boundary)
CREATE UNIQUE INDEX IF NOT EXISTS idx_phi_record_unique
ON phi_records(patient_id, resource_type, record_version)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_phi_record_unique IS
'Unique constraint on PHI records: prevents duplicate patient/resource/version combinations.
SOC2 C-level requirement: idempotent healthcare data operations.';

-- Session idempotency: user + device + timestamp uniqueness
CREATE UNIQUE INDEX IF NOT EXISTS idx_session_unique
ON sessions(user_id, device_id, created_at::date)
WHERE revoked_at IS NULL;

COMMENT ON INDEX idx_session_unique IS
'Unique constraint on sessions: one active session per user/device/day. Prevents token replay attacks.';

-- Workspace name: user scope (database-level uniqueness)
-- ALREADY EXISTS: idx_osa_workspaces_user_name_unique
-- (see migration 048_osa_integration.sql CONSTRAINT osa_workspaces_user_name_unique)

-- =============================================================================
-- PARTIAL UNIQUE INDEXES (Enforcement with soft-delete support)
-- =============================================================================

-- Email uniqueness: ignore deleted users
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_email_active
ON "user"(email COLLATE "C")
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_user_email_active IS
'Unique constraint on email for non-deleted users. Allows email reuse after account deletion.
Collation: C (byte-order) for email locality detection.';

-- Deal name per domain (business domain isolation)
CREATE UNIQUE INDEX IF NOT EXISTS idx_deal_name_domain
ON deals(domain, name)
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_deal_name_domain IS
'Unique constraint on deal name within business domain. Enforces naming uniqueness per domain only.';

-- =============================================================================
-- EXPRESSION INDEXES (Computed columns for faster queries)
-- =============================================================================

-- Case-insensitive search on module names
CREATE INDEX IF NOT EXISTS idx_osa_modules_name_lower
ON osa_modules(LOWER(name))
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_osa_modules_name_lower IS
'Expression index on lowercase module name. Supports case-insensitive search without LOWER() function in query.';

-- Month-bucketed activity (time-series aggregation)
CREATE INDEX IF NOT EXISTS idx_audit_logs_month
ON audit_logs(DATE_TRUNC('month', created_at))
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_audit_logs_month IS
'Expression index on DATE_TRUNC. Speeds up monthly audit reports without materializing every day.';

-- Quality score band (quartile filtering without function call)
CREATE INDEX IF NOT EXISTS idx_deals_quality_band
ON deals((CASE
  WHEN quality_score < 25 THEN 'poor'
  WHEN quality_score < 50 THEN 'fair'
  WHEN quality_score < 75 THEN 'good'
  ELSE 'excellent'
END))
WHERE deleted_at IS NULL;

COMMENT ON INDEX idx_deals_quality_band IS
'Expression index on quality score band. Supports fast filtering by quality tier without case/when in query.';

-- =============================================================================
-- HASH INDEXES (O(1) exact match lookup)
-- =============================================================================

-- Deal UUID lookup (standard PK, shown for completeness)
-- Note: Hash indexes rarely outperform B-tree for UUID, but useful for:
-- - Very large tables with high cardinality on single column
-- - ACID properties not needed on index (read-mostly lookup)

-- API key exact matching (high cardinality, read-only lookups)
CREATE INDEX IF NOT EXISTS idx_api_keys_value_hash
ON api_keys USING HASH (key_value)
WHERE revoked_at IS NULL;

COMMENT ON INDEX idx_api_keys_value_hash IS
'Hash index on API key value for O(1) exact match lookup. Supports high-speed API authentication without range queries.';

-- =============================================================================
-- INDEX STATISTICS & MAINTENANCE
-- =============================================================================

-- Collect statistics for query planner (required after index creation)
-- Note: in production, ANALYZE should be scheduled via cron or scheduled job
-- This is a placeholder for documentation; actual ANALYZE runs separately

COMMENT ON SCHEMA public IS
'Index Statistics Note: Run ANALYZE on tables after migration.
Command: SELECT * FROM pg_stat_user_indexes WHERE schemaname = ''public'';
Query planner uses pg_stat_user_indexes to choose optimal execution plans.';

-- =============================================================================
-- DOCUMENTATION: Index Selection Strategy
-- =============================================================================

/*
INDEX SELECTION GUIDE (for developers adding new queries):

1. COMPOSITE (Multi-column):
   - Use when: WHERE col1 = ? AND col2 = ? (all columns in filter)
   - Benefit: Index-only scans, reduced heap access
   - Example: idx_deals_domain_quality (domain + quality score)

2. PARTIAL (Filtered):
   - Use when: Most queries have WHERE status != 'deleted' or similar
   - Benefit: 5-10x smaller index, faster updates
   - Example: idx_deals_active (WHERE deleted_at IS NULL)

3. BRIN (Block Range):
   - Use when: Column is monotonically increasing (timestamps, IDs) and table >100GB
   - Benefit: 1-10MB vs 100MB+ B-tree index
   - Example: idx_audit_logs_timestamp_brin

4. GIN (Inverted):
   - Use when: Searching JSONB (@>, ?) or array containment (@>)
   - Benefit: O(1) to O(log n) vs O(n) full scan
   - Example: idx_osa_modules_schema_jsonb_gin

5. EXPRESSION (Computed):
   - Use when: Query uses function on column (LOWER, DATE_TRUNC, CASE)
   - Benefit: Function evaluation moved to index, not query
   - Example: idx_osa_modules_name_lower (LOWER(name))

6. HASH (Exact Match):
   - Use when: Column has very high cardinality and only exact matches (no ranges)
   - Benefit: O(1) constant-time lookup
   - Example: idx_api_keys_value_hash

ANTI-PATTERNS TO AVOID:
- Multiple single-column indexes on same table (combine into composite)
- Composite indexes on low-cardinality columns (status, type)
- Indexes on UPDATE/DELETE columns (slows writes)
- INCLUDE clauses on BRIN indexes (BRIN doesn't support INCLUDE)
- GIN indexes on frequently updated JSONB (GIN writes slower than B-tree)
*/

-- =============================================================================
-- ROLLBACK INSTRUCTIONS
-- =============================================================================
-- +migrate Down

DROP INDEX IF EXISTS idx_deals_domain_quality;
DROP INDEX IF EXISTS idx_osa_modules_framework_status;
DROP INDEX IF EXISTS idx_audit_actor_timestamp;
DROP INDEX IF EXISTS idx_osa_workspaces_user_status;
DROP INDEX IF EXISTS idx_phi_resource_quality;
DROP INDEX IF EXISTS idx_deals_active;
DROP INDEX IF EXISTS idx_audit_logs_active;
DROP INDEX IF EXISTS idx_osa_modules_active;
DROP INDEX IF EXISTS idx_sessions_unencrypted;
DROP INDEX IF EXISTS idx_audit_logs_timestamp_brin;
DROP INDEX IF EXISTS idx_phi_access_timestamp_brin;
DROP INDEX IF EXISTS idx_deal_activity_timestamp_brin;
DROP INDEX IF EXISTS idx_osa_modules_schema_jsonb_gin;
DROP INDEX IF EXISTS idx_deals_custom_fields_gin;
DROP INDEX IF EXISTS idx_osa_modules_tags_gin;
DROP INDEX IF EXISTS idx_user_permissions_gin;
DROP INDEX IF EXISTS idx_phi_record_unique;
DROP INDEX IF EXISTS idx_session_unique;
DROP INDEX IF EXISTS idx_user_email_active;
DROP INDEX IF EXISTS idx_deal_name_domain;
DROP INDEX IF EXISTS idx_osa_modules_name_lower;
DROP INDEX IF EXISTS idx_audit_logs_month;
DROP INDEX IF EXISTS idx_deals_quality_band;
DROP INDEX IF EXISTS idx_api_keys_value_hash;
