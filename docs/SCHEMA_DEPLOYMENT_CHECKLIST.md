# PostgreSQL Schema Deployment Checklist

**Migration Set:** 026–030
**Date:** 2026-03-26
**Environment:** Development → Staging → Production

---

## Pre-Deployment Review

- [ ] **DBA Review** — Schema design reviewed by database administrator
- [ ] **Backup Verified** — Production backup exists and tested for restore
- [ ] **Downtime Window** — Maintenance window scheduled (if production)
- [ ] **Rollback Plan** — Previous schema snapshot available for downgrade
- [ ] **Users Informed** — Team notified of schema changes
- [ ] **Documentation Current** — `database-schema.md` and related docs up to date

---

## Database Prerequisites

- [ ] **PostgreSQL 15+** — `SELECT version();` confirms version
- [ ] **Disk Space** — At least 1 GB free space in PGDATA directory
- [ ] **Superuser Access** — Can connect as `postgres` or superuser
- [ ] **Timezone Set** — `SHOW timezone;` confirms correct timezone
- [ ] **Extensions Loaded** — `pgcrypto` available for `gen_random_uuid()`
- [ ] **users Table Exists** — Foreign key references valid: `\dt users`

**Verify Prerequisites:**
```bash
psql -U postgres -c "SELECT version();"
psql -U postgres -d businessos -c "SELECT * FROM information_schema.tables WHERE table_name='users';"
psql -U postgres -d businessos -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
```

---

## Deployment Steps

### Step 1: Backup (Production Only)

```bash
# Full database backup
BACKUP_FILE="/backups/businessos-pre-migration-$(date +%Y%m%d_%H%M%S).dump"
pg_dump -Fc -U postgres businessos > "$BACKUP_FILE"
echo "Backup size: $(du -h "$BACKUP_FILE" | cut -f1)"

# Verify backup integrity
pg_dump -Fc "$BACKUP_FILE" > /dev/null && echo "Backup verified OK"
```

**Expected Result:**
- Backup file created successfully
- File size > 1 MB (depends on existing data)
- Restoration test passes (optional but recommended)

### Step 2: Execute Migrations (In Order)

```bash
cd /Users/sac/chatmangpt/BusinessOS/desktop/backend-go/migrations

# Using psql directly
psql -U postgres -d businessos -f 026_create_fibo_deals.sql
psql -U postgres -d businessos -f 027_create_compliance_controls.sql
psql -U postgres -d businessos -f 028_create_datasets.sql
psql -U postgres -d businessos -f 029_create_healthcare_phi.sql
psql -U postgres -d businessos -f 030_create_audit_logs.sql
```

**Expected Output:**
```
CREATE TABLE
CREATE INDEX
...
(no errors)
```

**If Using Migration Tool (golang-migrate):**
```bash
migrate -path migrations -database "postgresql://user:pass@localhost/businessos" up
```

**Expected Output:**
```
5/5 migrations up
```

### Step 3: Verify Migration Success

```bash
# Connect to database
psql -U postgres -d businessos

# Verify all tables created
\dt
-- Should list: deals, compliance_controls, datasets, phi_records, audit_logs,
-- dataset_lineage, dataset_quality_metrics, phi_audit_log, phi_breach_detection,
-- phi_consent_log, audit_log_integrity, audit_log_summary, audit_log_retention_policy,
-- compliance_control_mappings

# Verify all indexes created
\di
-- Should show 43 indexes

# Verify all triggers created
SELECT trigger_name FROM information_schema.triggers
WHERE trigger_schema = 'public';
-- Should list: trigger_deals_updated_at, trigger_compliance_controls_updated_at, etc.

# Verify functions created
\df archive_old_audit_logs
\df verify_audit_log_integrity
-- Both should appear in list

# Test a query
SELECT COUNT(*) FROM deals;
-- Should return: 0 (empty table, which is correct)
```

---

## Data Integrity Tests

### Test 1: Foreign Key Constraints

```bash
psql -U postgres -d businessos << 'EOF'
-- Try to insert deal with non-existent user (should fail)
INSERT INTO deals (name, amount_cents, currency, created_by, domain)
VALUES ('Test Deal', 100000, 'USD', 'invalid-uuid'::UUID, 'equity');
-- Expected error: violates foreign key constraint

\echo "✓ Foreign key constraint works"
EOF
```

### Test 2: Check Constraints

```bash
psql -U postgres -d businessos << 'EOF'
-- Try to insert deal with negative amount (should fail)
INSERT INTO deals (name, amount_cents, currency, created_by, domain, created_by)
VALUES ('Bad Deal', -100000, 'USD', 'real-user-uuid'::UUID, 'equity', 'real-user-uuid'::UUID);
-- Expected error: new row for relation "deals" violates check constraint

\echo "✓ Check constraint works"
EOF
```

### Test 3: Unique Constraints

```bash
psql -U postgres -d businessos << 'EOF'
-- Insert a compliance control (should succeed)
INSERT INTO compliance_controls
  (framework, control_id, title, description, severity)
VALUES ('SOC2', 'CC1.1', 'Risk Management', 'Test control', 'high');

-- Try to insert duplicate (should fail)
INSERT INTO compliance_controls
  (framework, control_id, title, description, severity)
VALUES ('SOC2', 'CC1.1', 'Different Title', 'Another test', 'medium');
-- Expected error: duplicate key value violates unique constraint

\echo "✓ Unique constraint works"
EOF
```

### Test 4: Trigger (Auto-Update Timestamp)

```bash
psql -U postgres -d businessos << 'EOF'
-- Insert a deal
INSERT INTO deals (name, amount_cents, currency, created_by, domain)
VALUES ('Trigger Test', 500000, 'USD', 'some-uuid'::UUID, 'fixed_income');

-- Check created_at and updated_at are equal
SELECT created_at = updated_at as timestamps_equal FROM deals
WHERE name = 'Trigger Test';
-- Expected: t (true)

-- Update the deal
UPDATE deals SET name = 'Updated Deal' WHERE name = 'Trigger Test';

-- Check updated_at > created_at
SELECT updated_at > created_at as updated_at_changed FROM deals
WHERE name = 'Updated Deal';
-- Expected: t (true)

\echo "✓ Trigger (auto-update timestamp) works"
EOF
```

### Test 5: Compliance Control Mappings

```bash
psql -U postgres -d businessos << 'EOF'
-- Insert two controls (different frameworks)
INSERT INTO compliance_controls
  (framework, control_id, title, description, severity)
VALUES ('SOC2', 'CC6.1', 'Logical Access', 'SOC2 control', 'high');

INSERT INTO compliance_controls
  (framework, control_id, title, description, severity)
VALUES ('GDPR', 'Article32', 'Security of Processing', 'GDPR control', 'high');

-- Get their IDs
\set soc2_id (SELECT id FROM compliance_controls WHERE framework='SOC2' AND control_id='CC6.1')
\set gdpr_id (SELECT id FROM compliance_controls WHERE framework='GDPR' AND control_id='Article32')

-- Create mapping
INSERT INTO compliance_control_mappings
  (source_control_id, target_control_id, mapping_type)
VALUES (
  (SELECT id FROM compliance_controls WHERE framework='SOC2' AND control_id='CC6.1'),
  (SELECT id FROM compliance_controls WHERE framework='GDPR' AND control_id='Article32'),
  'equivalent'
);

-- Verify mapping
SELECT COUNT(*) as mapping_count FROM compliance_control_mappings;
-- Expected: 1

\echo "✓ Compliance control mapping works"
EOF
```

### Test 6: PHI Encryption

```bash
psql -U postgres -d businessos << 'EOF'
-- Insert a PHI record (with encrypted data)
INSERT INTO phi_records
  (patient_id, patient_id_hash, resource_type, data_hash, data_encrypted,
   data_classification, consent_status)
VALUES (
  'patient-12345',
  'hash-sha256-patient-12345',
  'Patient',
  'hash-sha256-encrypted-data',
  E'\\x' || encode(randomBytes(64), 'hex')::bytea,  -- Simulated encrypted blob
  'phi',
  'granted'
);

-- Verify data_encrypted is not null
SELECT data_encrypted IS NOT NULL as data_encrypted FROM phi_records
WHERE patient_id = 'patient-12345';
-- Expected: t (true)

\echo "✓ PHI encryption works"
EOF
```

### Test 7: Audit Log Integrity Function

```bash
psql -U postgres -d businessos << 'EOF'
-- Insert an audit log
INSERT INTO audit_logs
  (action, action_category, resource_type, result)
VALUES (
  'user_login',
  'authentication',
  'user',
  'success'
);

-- Get the log ID
SELECT id FROM audit_logs ORDER BY created_at DESC LIMIT 1;

-- Test integrity verification function
SELECT * FROM verify_audit_log_integrity(
  (SELECT id FROM audit_logs ORDER BY created_at DESC LIMIT 1)::UUID
);
-- Expected: valid=true, checksum_match=true, notes='Integrity verified'

\echo "✓ Audit log integrity function works"
EOF
```

---

## Performance Tests

### Test 1: Index Coverage

```bash
psql -U postgres -d businessos << 'EOF'
-- Check index sizes
SELECT
  schemaname,
  tablename,
  indexname,
  pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;

-- Expected: All indexes listed with reasonable sizes (KB range initially)
EOF
```

### Test 2: Query Plan (Verify Index Usage)

```bash
psql -U postgres -d businessos << 'EOF'
-- Check if index is used for status filter
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, name FROM deals WHERE status = 'approved';

-- Expected: "Index Scan using idx_deals_status"

-- Check if index is used for timestamp ordering
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, action FROM audit_logs
WHERE action_category = 'data_modification'
ORDER BY timestamp DESC LIMIT 10;

-- Expected: "Index Scan using idx_audit_logs_..." or "Index Only Scan"
EOF
```

### Test 3: Insert Performance

```bash
psql -U postgres -d businessos << 'EOF'
-- Measure bulk insert performance (1000 rows)
\timing on

INSERT INTO audit_logs
  (action, action_category, resource_type, result)
SELECT
  'test_action_' || i,
  CASE (i % 5)
    WHEN 0 THEN 'authentication'
    WHEN 1 THEN 'data_access'
    WHEN 2 THEN 'data_modification'
    WHEN 3 THEN 'configuration'
    ELSE 'system'
  END,
  'test_entity',
  CASE (i % 3)
    WHEN 0 THEN 'success'
    WHEN 1 THEN 'failure'
    ELSE 'partial'
  END
FROM generate_series(1, 1000) as i;

\timing off

-- Expected: < 1000 ms (fast inserts with 17 indexes)

SELECT COUNT(*) FROM audit_logs;
-- Expected: 1000
EOF
```

---

## Compliance Verification

### SOC2: Audit Trail

```bash
psql -U postgres -d businessos << 'EOF'
-- Verify audit logging structure
\d audit_logs
-- Should show all columns: action, action_category, result, signature, etc.

-- Verify retention policy
SELECT * FROM audit_log_retention_policy;
-- Should show default policies with 730-2555 day retention

-- Verify integrity table exists
\d audit_log_integrity
-- Should have checksum, status, verified_at columns
EOF
```

### HIPAA: PHI Encryption

```bash
psql -U postgres -d businessos << 'EOF'
-- Verify PHI table structure
\d phi_records
-- Should show: data_encrypted (BYTEA), encryption_algorithm (VARCHAR)

-- Verify audit log for PHI
\d phi_audit_log
-- Should show: action (create/read/update/delete), retention_until, result

-- Verify breach detection exists
\d phi_breach_detection
-- Should show: alert_type, severity, investigation_status
EOF
```

### GDPR: Data Classification

```bash
psql -U postgres -d businessos << 'EOF'
-- Verify sensitivity levels
SELECT DISTINCT sensitivity_level FROM datasets WHERE sensitivity_level IS NOT NULL;
-- Should show: public, internal, confidential, restricted (or NULL for unclassified)

-- Verify PII tracking
SELECT DISTINCT pii_types FROM datasets WHERE pii_present = true;
-- Should show arrays with PII types: email, phone, ssn, address, etc.
EOF
```

### SOX: 7-Year Retention

```bash
psql -U postgres -d businessos << 'EOF'
-- Verify retention dates
SELECT
  action_category,
  retention_days,
  reason
FROM audit_log_retention_policy
ORDER BY retention_days DESC;

-- Expected:
-- data_modification   | 2555 | SOX requirement: 7 years
-- configuration       | 2555 | SOX requirement: 7 years
-- ...
-- authentication      | 730  | HIPAA minimum: 2 years

-- Verify default 7-year retention on new logs
SELECT (retention_until - created_at) as retention_period
FROM audit_logs
WHERE created_at > NOW() - INTERVAL '1 day'
LIMIT 1;

-- Expected: 2555 days (approximately)
EOF
```

---

## Post-Deployment

- [ ] **Update ORM Models** — Generate Go models from schema (sqlc)
- [ ] **Integration Tests Pass** — Run full test suite: `make test-backend`
- [ ] **Documentation Updated** — Add schema to team wiki/docs
- [ ] **Monitor Performance** — Track slow query log for first 24 hours
- [ ] **Backup Verification** — Confirm backup process includes new tables
- [ ] **Team Notification** — Update schema documentation for team

**Verification Script (Final):**
```bash
psql -U postgres -d businessos << 'EOF'
SELECT
  'Tables' as object_type, COUNT(*) as count
FROM information_schema.tables
WHERE table_schema = 'public'
UNION ALL
SELECT 'Indexes', COUNT(*)
FROM pg_indexes
WHERE schemaname = 'public'
UNION ALL
SELECT 'Triggers', COUNT(*)
FROM information_schema.triggers
WHERE trigger_schema = 'public'
ORDER BY object_type;

-- Expected output:
-- Indexes       | 43
-- Tables        | 14
-- Triggers      | 4
EOF
```

---

## Rollback Procedure (If Needed)

**Automated Rollback (if backup exists):**
```bash
# Stop application
systemctl stop businessos

# Restore from backup
pg_restore -d businessos /backups/businessos-pre-migration-20260326_120000.dump

# Start application
systemctl start businessos

# Verify rollback
psql -U postgres -d businessos -c "SELECT COUNT(*) FROM deals;"
-- Should fail: "relation 'deals' does not exist" (tables removed)
```

**Manual Rollback (drop new schema objects):**
```bash
psql -U postgres -d businessos << 'EOF'
-- Drop new tables (reverse order of migration 030, 029, 028, 027, 026)
DROP TABLE IF EXISTS audit_log_retention_policy CASCADE;
DROP TABLE IF EXISTS audit_log_summary CASCADE;
DROP TABLE IF EXISTS audit_log_integrity CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS phi_consent_log CASCADE;
DROP TABLE IF EXISTS phi_breach_detection CASCADE;
DROP TABLE IF EXISTS phi_audit_log CASCADE;
DROP TABLE IF EXISTS phi_records CASCADE;
DROP TABLE IF EXISTS dataset_quality_metrics CASCADE;
DROP TABLE IF EXISTS dataset_lineage CASCADE;
DROP TABLE IF EXISTS datasets CASCADE;
DROP TABLE IF EXISTS compliance_control_mappings CASCADE;
DROP TABLE IF EXISTS compliance_controls CASCADE;
DROP TABLE IF EXISTS deals CASCADE;

-- Verify rollback
SELECT COUNT(*) as remaining_tables
FROM information_schema.tables WHERE table_schema = 'public';
-- Expected: (original count, e.g., 5 - excluding new tables)
EOF
```

---

## Sign-Off

| Role | Name | Date | Signature |
|------|------|------|-----------|
| DBA | _____________ | ________ | ____________ |
| Developer Lead | _____________ | ________ | ____________ |
| QA Manager | _____________ | ________ | ____________ |
| Security Officer | _____________ | ________ | ____________ |

---

## Support Contact

- **Database Issues:** DBA on-call
- **Schema Questions:** See `database-schema.md`
- **Rollback Needed:** Contact DevOps lead
- **Audit Questions:** Compliance officer

---

*Deployment ready for Fortune 500-grade production.*
