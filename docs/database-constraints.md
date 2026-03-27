# Database Constraints & Indexes — Fortune 5 Performance Reference

**Last Updated:** 2026-03-26
**Scope:** PostgreSQL 15+ | BusinessOS schema
**Standards:** SOC2 A/B/C compliance, WvdA Soundness, Armstrong Supervision

---

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Constraint Types](#constraint-types)
3. [Index Selection Guide](#index-selection-guide)
4. [Performance Impact](#performance-impact)
5. [Migration Guide](#migration-guide)
6. [Troubleshooting](#troubleshooting)

---

## Quick Reference

### Constraint Summary

| Table | Constraint | Type | Business Rule |
|-------|-----------|------|---------------|
| `deals` | `check_deal_amount_positive` | CHECK | amount > 0 (financial integrity) |
| `deals` | `check_deal_quality_bounds` | CHECK | 0 ≤ quality_score ≤ 100 (percentile metric) |
| `deals` | `check_temporal_ordering` | CHECK | created_at ≤ updated_at (logical time) |
| `deals` | `idx_deal_name_domain` | UNIQUE | (domain, name) unique per domain |
| `phi_records` | `check_phi_confidence_bounds` | CHECK | 0.0 ≤ confidence ≤ 1.0 (probability) |
| `data_lineage` | `check_lineage_depth_bounded` | CHECK | 1 ≤ depth ≤ 5 (prevent recursion) |
| `agent_heartbeats` | `check_heartbeat_interval_bounds` | CHECK | 100ms ≤ interval ≤ 60s (supervision) |
| `audit_logs` | `prevent_audit_log_update()` | TRIGGER | Immutable (SOC2 A-level) |

### Index Summary

| Name | Type | Table | Columns | Use Case |
|------|------|-------|---------|----------|
| `idx_deals_domain_quality` | COMPOSITE | deals | (domain, quality_score DESC) | Deal discovery by quality |
| `idx_audit_actor_timestamp` | COMPOSITE | audit_logs | (actor_id, created_at DESC) | User activity timeline |
| `idx_audit_logs_timestamp_brin` | BRIN | audit_logs | created_at | Time-range queries on 100M+ rows |
| `idx_osa_modules_schema_jsonb_gin` | GIN | osa_modules | schema_definition | JSONB schema search |
| `idx_user_email_active` | PARTIAL UNIQUE | user | email | Email uniqueness (active only) |
| `idx_osa_modules_active` | PARTIAL | osa_modules | (id, status) | Filter non-archived modules |

---

## Constraint Types

### 1. CHECK Constraints (Value Validation)

**Purpose:** Enforce domain-level rules at database boundary. Prevents invalid state from entering system.

#### Deal Constraints

##### Amount Must Be Positive
```sql
ALTER TABLE deals
ADD CONSTRAINT check_deal_amount_positive
CHECK (amount > 0);
```

**Migration:** `103_add_constraints.sql`
**Application:** `internal/db/constraints.go::Deal.validateAmount()`
**Rationale:**
- Financial integrity: no negative/zero amounts in system
- Caught before persistence: invalid amounts fail fast
- Audit trail: every amount in database is valid

**Example Violation:**
```go
deal := &Deal{Amount: -100.00, Domain: "sales"}
if err := deal.ValidateConstraints(); err != nil {
  // err = ErrDealAmountNonPositive
  log.Error("Invalid deal amount", "amount", deal.Amount)
}
```

##### Quality Score Bounds [0, 100]
```sql
ALTER TABLE deals
ADD CONSTRAINT check_deal_quality_bounds
CHECK (quality_score >= 0 AND quality_score <= 100);
```

**Why:** Quality is a percentile metric. Bounds ensure:
- Machine learning models produce [0, 100] output
- No -50 or 150 scores (nonsensical metrics)
- Visual UI can assume bounded scale

**Default:** 50 (neutral/unknown). Updated by AI assessment engine.

#### PHI Confidence [0.0, 1.0]

```sql
ALTER TABLE phi_records
ADD CONSTRAINT check_phi_confidence_bounds
CHECK (confidence_level >= 0.0 AND confidence_level <= 1.0);
```

**Compliance:** SOC2 B-level. Confidence is a probability — precision requirement.

**Query:** Find high-confidence PHI (ready for use):
```sql
SELECT * FROM phi_records
WHERE confidence_level >= 0.9 AND deleted_at IS NULL;
```

#### Lineage Depth [1, 5] — WvdA Soundness

```sql
ALTER TABLE data_lineage
ADD CONSTRAINT check_lineage_depth_bounded
CHECK (lineage_depth >= 1 AND lineage_depth <= 5);
```

**WvdA Property:** Bounds prevent infinite recursion in data provenance graphs.

| Depth | Example | Purpose |
|-------|---------|---------|
| 1 | Raw data (no parents) | Source record |
| 2 | Cleaned from raw | First transformation |
| 3 | Aggregated from cleaned | Second transformation |
| 4 | Derived for ML | Third transformation |
| 5 | Model output | Final product |
| >5 | (forbidden) | Would indicate circular dependency |

**Proof of Soundness:**
- Bounded depth (1-5) guaranteed to terminate
- Prevents cycle detection algorithms from hanging
- Formal model: DAG (directed acyclic graph)

#### Heartbeat Interval [100ms, 60s] — Armstrong Supervision

```sql
ALTER TABLE agent_heartbeats
ADD CONSTRAINT check_heartbeat_interval_bounds
CHECK (interval_ms >= 100 AND interval_ms <= 60000);
```

**Armstrong Principle:** Supervisor monitors child process liveness via periodic heartbeats.

| Interval | Use Case | Rationale |
|----------|----------|-----------|
| 100ms | Critical services | Fast failure detection |
| 1000ms (default) | Standard services | Balances responsiveness + CPU |
| 5000ms | Background services | Low overhead |
| 60000ms | Max | Prevents supervisor queue buildup |

**Violation Example:**
```go
hb := &AgentHeartbeat{IntervalMs: 150000}  // 150 seconds
err := hb.ValidateConstraints()
// err = ErrHeartbeatIntervalTooLong
// "heartbeat interval must not exceed 60s (Armstrong Supervision)"
```

#### Temporal Ordering (created_at ≤ updated_at)

```sql
ALTER TABLE deals
ADD CONSTRAINT check_temporal_ordering
CHECK (created_at <= updated_at);
```

**Prevents:** Logical impossibility (record updated before creation).

**When Violated:**
```sql
-- Impossible: created 2026-03-26 10:00, updated 2026-03-26 09:00
INSERT INTO deals (id, created_at, updated_at, ...)
VALUES ('123', '2026-03-26 10:00', '2026-03-26 09:00', ...) -- FAILS

-- ✅ CORRECT
UPDATE deals SET updated_at = NOW() WHERE id = '123';
```

---

### 2. Foreign Key Constraints (Referential Integrity)

**Purpose:** Maintain data consistency across related tables. Define cascade behavior for deletions.

#### Cascade Delete Strategy

**Rule:** When parent record deleted, all dependent child records deleted automatically.

##### Deal → Audit Logs
```sql
ALTER TABLE audit_logs
ADD CONSTRAINT fk_audit_logs_deal_cascade
FOREIGN KEY (deal_id) REFERENCES deals(id)
ON DELETE CASCADE
ON UPDATE CASCADE;
```

**Behavior:**
```sql
-- Delete a deal
DELETE FROM deals WHERE id = 'deal-123';

-- ✅ Cascade effect: automatically deletes all audit logs for that deal
-- SELECT COUNT(*) FROM audit_logs WHERE deal_id = 'deal-123'; -- Returns 0
```

**Why Cascade?**
- Audit logs are child of deal (deal lifecycle owns audit history)
- No external system references audit logs
- Deletion must be atomic (deal + history together)

##### Deal → PHI Records
```sql
ALTER TABLE phi_records
ADD CONSTRAINT fk_phi_records_deal_cascade
FOREIGN KEY (deal_id) REFERENCES deals(id)
ON DELETE CASCADE;
```

**Compliance:** SOC2 A-level. Ensures PHI cleanup when deal deleted.

**Query to verify:**
```sql
-- Count orphaned PHI records (should be 0)
SELECT COUNT(*) FROM phi_records
WHERE deal_id NOT IN (SELECT id FROM deals);
```

##### Workspace → Modules
```sql
ALTER TABLE osa_modules
ADD CONSTRAINT fk_osa_modules_workspace_cascade
FOREIGN KEY (workspace_id) REFERENCES osa_workspaces(id)
ON DELETE CASCADE;
```

**Behavior:** Deleting workspace removes all OSA modules in it (atomic workspace deletion).

#### Restrict Strategy (When NOT to Cascade)

**Anti-pattern to avoid:**
```sql
-- DON'T DO THIS: Cascade without understanding consequences
ALTER TABLE payments
ADD CONSTRAINT fk_payments_deal
FOREIGN KEY (deal_id) REFERENCES deals(id)
ON DELETE CASCADE;  -- ← DANGEROUS!
```

**Problem:** Deleting a deal accidentally voids all payments. Should RESTRICT instead:
```sql
-- CORRECT: Prevent deal deletion if payments exist
ALTER TABLE payments
ADD CONSTRAINT fk_payments_deal
FOREIGN KEY (deal_id) REFERENCES deals(id)
ON DELETE RESTRICT;  -- Can't delete deal with payments
```

---

### 3. UNIQUE Constraints (Preventing Duplicates)

#### Email Uniqueness (with soft-delete support)

```sql
CREATE UNIQUE INDEX idx_user_email_active
ON "user"(email COLLATE "C")
WHERE deleted_at IS NULL;
```

**Effect:** Only non-deleted users must have unique email.

**Use Case:**
```sql
-- User deletes account
UPDATE "user" SET deleted_at = NOW() WHERE id = 'user-123';

-- New user can reuse same email
INSERT INTO "user" (email, ...)
VALUES ('alice@example.com', ...);  -- ✅ Allowed (old alice deleted)
```

**Why PARTIAL UNIQUE?**
- Allows soft deletes (deleted_at IS NOT NULL)
- No need to append "-deleted-2026-03-26" to email
- Clean user experience for account reactivation

#### Deal Name Uniqueness per Domain

```sql
CREATE UNIQUE INDEX idx_deal_name_domain
ON deals(domain, name)
WHERE deleted_at IS NULL;
```

**Effect:** Deal names must be unique within domain, not globally.

**Use Case:**
```sql
-- Same deal name allowed in different domains
INSERT INTO deals (domain, name, ...) VALUES ('sales', 'Acme Corp', ...);  -- ✅
INSERT INTO deals (domain, name, ...) VALUES ('services', 'Acme Corp', ...);  -- ✅

-- But not twice in same domain
INSERT INTO deals (domain, name, ...) VALUES ('sales', 'Acme Corp', ...);  -- FAILS (duplicate)
```

---

### 4. Trigger-Based Constraints (Complex Logic)

#### Audit Log Immutability (SOC2 A-level)

```plpgsql
CREATE OR REPLACE FUNCTION prevent_audit_log_update()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'UPDATE' THEN
    RAISE EXCEPTION 'Audit logs are immutable (SOC2 A-level requirement)';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_log_immutability
BEFORE UPDATE ON audit_logs
FOR EACH ROW
EXECUTE FUNCTION prevent_audit_log_update();
```

**Behavior:**
```sql
-- ✅ Inserts work
INSERT INTO audit_logs (actor_id, action, ...)
VALUES ('user-123', 'deal_created', ...);

-- ❌ Updates fail
UPDATE audit_logs SET action = 'deal_deleted' WHERE id = 'log-123';
-- ERROR: Audit logs are immutable (SOC2 A-level requirement)
```

**Why Immutable?**
- Compliance: audit trail must not be tampered with
- Security: logged events are permanent record
- Governance: no retroactive deletion of audit history

---

## Index Selection Guide

### 1. Composite Indexes (Multi-Column)

**When:** Query filters on multiple columns with AND.

#### Example: Deal Domain + Quality
```sql
CREATE INDEX idx_deals_domain_quality
ON deals(domain ASC, quality_score DESC)
WHERE deleted_at IS NULL
INCLUDE (id, actor_id, status, created_at);
```

**Typical Query:**
```sql
SELECT id, actor_id, status FROM deals
WHERE domain = 'sales' AND quality_score >= 80
ORDER BY quality_score DESC
LIMIT 10;
```

**Why Composite?**
- Single index serves both `domain` and `quality_score` filters
- INCLUDE clause adds extra columns for index-only scans (no heap access)
- BRIN alternative: if table >100GB, use BRIN on timestamp instead

**Performance Impact:**
- Before index: O(n) full table scan (100M rows = 10 seconds)
- After index: O(log n) + index seek (100M rows = 10 milliseconds)
- 1000x speedup

---

### 2. BRIN Indexes (Time-Series)

**When:** Column is monotonically increasing (timestamps, IDs) and table >100GB.

```sql
CREATE INDEX idx_audit_logs_timestamp_brin
ON audit_logs USING BRIN (created_at)
WITH (pages_per_range=128);
```

**Size Comparison (100M rows):**
| Index Type | Size | Query Time |
|-----------|------|-----------|
| B-tree | ~150 MB | 5 ms |
| BRIN | ~5 MB | 10 ms |
| No index | 0 MB | 30 seconds |

**Use Case:** Audit logs (billions of historical records).

**Typical Query:**
```sql
-- Find all audit logs from last 7 days
SELECT * FROM audit_logs
WHERE created_at > NOW() - INTERVAL '7 days';
```

**Why BRIN?**
- Logs are inserted in chronological order
- Queries almost always use time range (last N days)
- BRIN index is 30x smaller than B-tree
- Trades 100% query speed for 98% index size reduction

---

### 3. GIN Indexes (JSONB & Arrays)

**When:** Searching JSONB fields or array containment.

```sql
CREATE INDEX idx_osa_modules_schema_jsonb_gin
ON osa_modules USING GIN (schema_definition jsonb_path_ops);
```

**Typical Queries:**
```sql
-- Find modules with specific schema field
SELECT * FROM osa_modules
WHERE schema_definition @> '{"database": {"type": "postgresql"}}';

-- Find modules with tag in array
SELECT * FROM osa_modules
WHERE tags @> ARRAY['production', 'critical'];
```

**Performance:**
- Without GIN: O(n) full scan (JSONB parsing every row)
- With GIN: O(log n) inverted index lookup

---

### 4. Partial Indexes (Soft Deletes)

**When:** Majority of queries filter on WHERE status != 'deleted'.

```sql
CREATE INDEX idx_deals_active
ON deals(id, status)
WHERE deleted_at IS NULL;
```

**Behavior:**
```sql
-- Index only includes non-deleted rows
SELECT * FROM deals WHERE deleted_at IS NULL;  -- Uses index ✅
SELECT * FROM deals WHERE deleted_at IS NOT NULL;  -- Full scan ❌
```

**Size Benefit:**
- Full index: 100M rows = 500 MB
- Partial index (10% deleted): 90M rows = 450 MB
- 10% savings (larger savings with higher deletion rate)

---

### 5. Expression Indexes (Computed Columns)

**When:** Query uses function on column.

```sql
CREATE INDEX idx_osa_modules_name_lower
ON osa_modules(LOWER(name))
WHERE deleted_at IS NULL;
```

**Typical Query:**
```sql
-- Case-insensitive search
SELECT * FROM osa_modules
WHERE LOWER(name) = 'my-module'
AND deleted_at IS NULL;
```

**Why Expression Index?**
- Function LOWER() evaluated during index creation, not query
- Query planner can use index without re-evaluating function
- 10-100x faster than query-time function evaluation

---

## Performance Impact

### Query Latency (Expected)

| Table Size | Without Index | B-tree Index | BRIN Index | GIN Index |
|-----------|---|---|---|---|
| 1M rows | 100 ms | 1 ms | 2 ms | 1 ms |
| 100M rows | 10 sec | 10 ms | 20 ms | 10 ms |
| 1B rows | 100 sec | 100 ms | 200 ms | 100 ms |

### Write Performance (Trade-off)

| Operation | Cost | Notes |
|-----------|------|-------|
| INSERT (no indexes) | 1x | Baseline |
| INSERT (1 B-tree) | 1.2x | Minimal overhead |
| INSERT (3 B-tree + 1 GIN) | 1.8x | Noticeable overhead |
| INSERT (10 indexes) | 3-5x | Significant slowdown |

**Recommendation:** Don't over-index. Index queries, not "just in case" scenarios.

---

### Disk Space (Maintenance)

| Index Type | Relative Size | Storage Cost |
|-----------|---|---|
| Full table | 100% | Baseline |
| B-tree index | 30-50% | ~40% additional |
| BRIN index | 3-5% | ~4% additional |
| GIN index | 50-80% | ~60% additional |

**Strategy:** Use BRIN for time-series, B-tree for business queries, GIN for JSONB.

---

## Migration Guide

### Adding New Constraints (Production)

**Step 1: Test in Staging**
```bash
cd /Users/sac/chatmangpt/BusinessOS
make db-migrate  # Runs migrations against staging DB
go test ./internal/db -run TestConstraints -v
```

**Step 2: Deploy to Production**
```bash
# Migration applied automatically by deployment pipeline
# (see docker-compose.yml migrate service)
```

**Step 3: Validate Application Handles Constraint Violations**
```go
// internal/handlers/deals.go
func CreateDeal(c *gin.Context) {
  var deal db.Deal
  c.BindJSON(&deal)

  // Validate BEFORE database (application layer)
  if err := deal.ValidateConstraints(); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
  }

  // Database will also check (defense in depth)
  if err := db.CreateDeal(deal); err != nil {
    // Constraint violation error from database
    c.JSON(400, gin.H{"error": "Invalid deal data"})
    return
  }
}
```

### Adding New Indexes (Low-Risk)

**Pros:**
- Non-blocking (PostgreSQL >= 11 supports concurrent index creation)
- No write-ahead log needed
- Safe to roll back (drop index)

**Command:**
```sql
-- CONCURRENT avoids exclusive lock on table
CREATE INDEX CONCURRENTLY idx_new_index ON table(column);
```

**Verification:**
```sql
-- Check index was created
SELECT * FROM pg_stat_user_indexes WHERE indexname = 'idx_new_index';

-- Run EXPLAIN to verify query planner uses it
EXPLAIN SELECT * FROM table WHERE column = 'value';
```

---

## Troubleshooting

### Problem: Constraint Violation on Insert

**Symptom:**
```
ERROR: new row for relation "deals" violates check constraint "check_deal_amount_positive"
```

**Root Cause:** Application sent negative amount to database.

**Fix:**
```go
// Bad (relies on database to catch)
deal := Deal{Amount: -100}
db.Create(&deal)  // Fails with constraint error

// Good (validates before database)
deal := Deal{Amount: -100}
if err := deal.ValidateConstraints(); err != nil {
  return err  // "deal amount must be positive"
}
```

### Problem: Slow Query Despite Index

**Symptom:**
```
SELECT * FROM deals WHERE domain = 'sales' TAKES 5 SECONDS
(despite idx_deals_domain_quality exists)
```

**Diagnosis:**
```sql
EXPLAIN ANALYZE SELECT * FROM deals
WHERE domain = 'sales'
ORDER BY quality_score DESC;
```

**Common Causes:**

1. **Statistics Outdated:** Index exists but query planner thinks it's slower
   ```sql
   ANALYZE deals;  -- Collect statistics
   ```

2. **Wrong Columns:** Filter uses different column
   ```sql
   -- Index: (domain, quality_score)
   -- Query: WHERE domain = 'sales' AND actor_id = 'user-123'
   -- ✗ Index doesn't help with actor_id
   ```

3. **INCLUDE Columns:** Not in filter
   ```sql
   -- Index: (domain) INCLUDE (quality_score)
   -- Query: SELECT quality_score FROM deals WHERE domain = 'sales'
   -- ✅ Index-only scan (no heap access)
   ```

### Problem: Cascade Delete Deleted Wrong Data

**Symptom:**
```
DELETE FROM deals WHERE id = '123';
-- Oops! Also deleted all deal_activity, audit_logs, phi_records
```

**Root Cause:** CASCADE was too aggressive.

**Prevention:**
```sql
-- RESTRICT instead (default)
ALTER TABLE audit_logs
ADD CONSTRAINT fk_audit_logs_deal
FOREIGN KEY (deal_id) REFERENCES deals(id)
ON DELETE RESTRICT;  -- Can't delete deal if audit logs exist

-- To delete deal:
DELETE FROM audit_logs WHERE deal_id = '123';  -- Must do this first
DELETE FROM deals WHERE id = '123';  -- Now this works
```

### Problem: Index Bloat

**Symptom:**
```
Index size keeps growing; queries getting slower.
```

**Diagnosis:**
```sql
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_size DESC;
```

**Solution:**
```sql
-- Rebuild index (locks table briefly)
REINDEX INDEX CONCURRENTLY idx_deals_domain_quality;
```

---

## SOC2 Compliance Mapping

| Control | Constraint | Evidence |
|---------|-----------|----------|
| **A-1** (Audit Trail) | Audit log immutability | `prevent_audit_log_update()` trigger |
| **A-2** (Data Integrity) | CHECK constraints | `check_deal_amount_positive`, etc. |
| **B-1** (PHI Security) | Encryption requirement | `ErrPHINotEncrypted` validation |
| **B-2** (Access Logging) | PHI access tracking | `phi_access_logs` table + trigger |
| **C-1** (Confidentiality) | Confidence bounds | `check_phi_confidence_bounds` |

---

## WvdA Soundness Mapping

| Property | Constraint | Evidence |
|----------|-----------|----------|
| **Deadlock-Free** | Heartbeat interval bounded | `check_heartbeat_interval_bounds [100ms, 60s]` |
| **Liveness** | Lineage depth bounded | `check_lineage_depth_bounded [1, 5]` |
| **Boundedness** | Temporal ordering | `check_temporal_ordering (created_at ≤ updated_at)` |

---

## Armstrong Fault Tolerance Mapping

| Principle | Constraint | Evidence |
|-----------|-----------|----------|
| **Supervision** | Heartbeat config | Interval limits + default 5000ms |
| **Let-It-Crash** | Cascade deletes | Parent deletion triggers cleanup |
| **Isolation** | Foreign key constraints | No orphaned records across boundaries |

---

## References

- PostgreSQL Official: [Indexes](https://www.postgresql.org/docs/current/indexes.html)
- PostgreSQL Official: [Constraints](https://www.postgresql.org/docs/current/ddl-constraints.html)
- van der Aalst, W.: *Process Mining* (2016), Chapter 2
- Armstrong, J.: *Making Reliable Distributed Systems* (2014)

---

**End of Document**

**Related Files:**
- Migration: `internal/database/migrations/102_create_indexes.sql`
- Migration: `internal/database/migrations/103_add_constraints.sql`
- Validation: `internal/db/constraints.go`
