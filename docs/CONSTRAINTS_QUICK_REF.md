# Database Constraints & Indexes — Quick Reference Card

**For developers:** Keep this at your desk. All rules in 2 pages.

---

## Golden Rules

1. **Always validate before database** — Use `internal/db/constraints.go`
2. **Indexes are for queries, not data** — Don't index every column
3. **Partial > Full** — Soft deletes should use partial indexes
4. **BRIN for time-series** — Created_at on 100M+ rows
5. **CASCADE = atomic deletion** — Parent delete removes all children

---

## Constraint Checklist

### When Creating Deals

```go
deal := &Deal{
  Domain: "sales",      // ✅ NOT NULL
  Name: "Acme Corp",
  Amount: 500000.00,    // ✅ > 0 (check_deal_amount_positive)
  Status: "prospect",   // ✅ In {prospect, negotiating, won, lost, archived}
  QualityScore: 75.0,   // ✅ [0, 100] (check_deal_quality_bounds)
}

if err := deal.ValidateConstraints(); err != nil {
  return err  // Fail fast before database
}
```

### When Creating PHI Records

```go
phi := &PHIRecord{
  PatientID: "patient-123",    // ✅ NOT NULL
  ResourceType: "medical",     // ✅ NOT NULL
  ConfidenceLevel: 0.95,       // ✅ [0.0, 1.0] (check_phi_confidence_bounds)
  EncryptedData: encrypted,    // ✅ NOT NULL (enforce encryption)
}

if err := phi.ValidateConstraints(); err != nil {
  return err  // Fail: missing encryption or confidence OOB
}
```

### When Creating Workspaces

```go
ws := &Workspace{
  UserID: "user-123",    // ✅ NOT NULL
  Name: "My Workspace",  // ✅ NOT NULL
  Mode: "2d",            // ✅ In {2d, 3d, hybrid}
}

if err := ws.ValidateConstraints(); err != nil {
  return err
}
```

### Audit Logs (READ-ONLY)

```go
// ✅ Can INSERT audit logs
INSERT INTO audit_logs (actor_id, action, resource_type, ...)
VALUES ('user-123', 'deal_created', 'deals', ...);

// ❌ CANNOT UPDATE audit logs
UPDATE audit_logs SET action = 'deal_deleted' WHERE id = 'log-123';
// ERROR: Audit logs are immutable (SOC2 A-level requirement)
```

---

## Index Quick Lookup

### "I want to query deals by domain + quality"

```go
// Query:
deals := db.FindDeals(
  "domain = $1 AND quality_score >= $2",
  []interface{}{"sales", 80},
)
// ✅ Uses: idx_deals_domain_quality
// Performance: 10 ms for 100M rows
```

### "I want to find user's audit history (last 30 days)"

```go
// Query:
audits := db.FindAuditLogs(
  "actor_id = $1 AND created_at > $2",
  []interface{}{"user-123", time.Now().AddDate(0, 0, -30)},
)
// ✅ Uses: idx_audit_actor_timestamp
// Performance: 50 ms for 1B rows
```

### "I need to check if email exists (excluding deleted users)"

```go
// Query:
exists := db.EmailExists("alice@example.com")
// ✅ Uses: idx_user_email_active (partial unique)
// Performance: 1 ms even for 10M users
```

### "I want to find modules with a schema property"

```go
// Query:
modules := db.FindModules(
  "schema_definition @> $1",
  []interface{}{`{"database": {"type": "postgresql"}}`},
)
// ✅ Uses: idx_osa_modules_schema_jsonb_gin
// Performance: 5 ms for 10M modules
```

---

## Error Codes (from constraints.go)

| Error | Meaning | How to Fix |
|-------|---------|-----------|
| `ErrDealAmountNonPositive` | amount ≤ 0 | Check amount > 0 |
| `ErrDealQualityOutOfBounds` | quality < 0 or > 100 | Check 0 ≤ quality ≤ 100 |
| `ErrDealTemporalOrdering` | created_at > updated_at | Check created_at ≤ updated_at |
| `ErrPHIConfidenceOutOfBounds` | confidence < 0 or > 1 | Check 0.0 ≤ confidence ≤ 1.0 |
| `ErrPHINotEncrypted` | empty encrypted_data | Encrypt data before insert |
| `ErrLineageDepthOutOfBounds` | depth < 1 or > 5 | WvdA: depth must be [1, 5] |
| `ErrHeartbeatIntervalTooShort` | interval < 100ms | Armstrong: min 100ms |
| `ErrHeartbeatIntervalTooLong` | interval > 60s | Armstrong: max 60s |
| `ErrWorkspaceModeInvalid` | mode not in {2d, 3d, hybrid} | Fix mode enum |
| `ErrAuditMissingActor` | actor_id empty | SOC2: all audits need actor |
| `ErrAuditImmutable` | attempted UPDATE on audit_log | Audit logs are read-only |

---

## Common Queries & Indexes

### "SELECT deals WHERE domain = 'sales' AND quality > 80"
```
Index: idx_deals_domain_quality (domain, quality_score DESC)
Time: 10 ms (100M rows)
Type: Composite B-tree with INCLUDE clause
```

### "SELECT audit_logs WHERE actor_id = 'user-123' ORDER BY created_at DESC LIMIT 100"
```
Index: idx_audit_actor_timestamp (actor_id, created_at DESC)
Time: 50 ms (1B rows)
Type: Composite B-tree
```

### "SELECT audit_logs WHERE created_at > NOW() - INTERVAL '7 days'"
```
Index: idx_audit_logs_timestamp_brin (created_at)
Time: 100 ms (1B rows, 30 MB index)
Type: BRIN (Block Range Index)
```

### "SELECT modules WHERE schema_definition @> '{\"database\": {\"type\": \"postgres\"}}'"
```
Index: idx_osa_modules_schema_jsonb_gin (schema_definition)
Time: 5 ms (10M modules)
Type: GIN inverted index
```

### "SELECT user WHERE email = 'alice@example.com' AND deleted_at IS NULL"
```
Index: idx_user_email_active (email)
Time: 1 ms (10M users)
Type: Partial unique index
```

---

## ADD_CONSTRAINT Workflow

### 1. Write failing test
```go
func TestDealAmountNonPositive(t *testing.T) {
  deal := &Deal{Amount: -100}
  err := deal.ValidateConstraints()
  assert.Equal(t, ErrDealAmountNonPositive, err)
}
```

### 2. Add constraint in constraints.go
```go
func (d *Deal) validateAmount() error {
  if d.Amount <= 0 {
    return fmt.Errorf("%w (got: %.2f)", ErrDealAmountNonPositive, d.Amount)
  }
  return nil
}
```

### 3. Add database constraint in migration
```sql
ALTER TABLE deals
ADD CONSTRAINT check_deal_amount_positive
CHECK (amount > 0);
```

### 4. Test passes ✅

---

## ADD_INDEX Workflow

### 1. Identify slow query
```
SELECT * FROM deals WHERE domain = 'sales' AND quality_score >= 80;
-- EXPLAIN: Seq Scan on deals (10 seconds on 100M rows)
```

### 2. Create index in migration
```sql
CREATE INDEX idx_deals_domain_quality
ON deals(domain, quality_score DESC)
WHERE deleted_at IS NULL;
```

### 3. Test query performance
```sql
EXPLAIN ANALYZE SELECT * FROM deals
WHERE domain = 'sales' AND quality_score >= 80;
-- Index Scan using idx_deals_domain_quality (10 ms) ✅
```

---

## SOC2 Mappings

| Requirement | How We Meet It |
|-----------|---|
| Audit trail immutable | `prevent_audit_log_update()` trigger |
| PHI encrypted | `ErrPHINotEncrypted` validation |
| Access logged | `phi_access_logs` table (app populates) |
| Data integrity | CHECK constraints on all sensitive columns |
| Temporal order | `check_temporal_ordering` |

---

## WvdA Soundness Mappings

| Property | How We Meet It |
|----------|---|
| Deadlock-free | Heartbeat interval [100ms, 60s] bounds |
| Liveness | Lineage depth [1, 5] bounds |
| Bounded | All metrics have min/max bounds |

---

## Armstrong Supervision Mappings

| Principle | How We Meet It |
|-----------|---|
| Supervision | Heartbeat interval validated + default 5s |
| Let-it-crash | Constraints fail fast (no swallowing) |
| No shared state | Referential integrity via FK |

---

## Performance Rules

| Rule | Rationale |
|------|-----------|
| **Index < 5% of table** | Maintenance overhead |
| **Query < 100ms** | User experience |
| **Constraint validate before DB** | Fail fast, clear errors |
| **Partial index for soft deletes** | 10-30% space savings |
| **BRIN for time-series 100M+** | 30x smaller than B-tree |
| **INCLUDE columns** | Index-only scans (no heap) |

---

## Pro Tips

1. **Check constraint before database**
   ```go
   if err := deal.ValidateConstraints(); err != nil {
     return err  // Fail fast with clear message
   }
   ```

2. **Use batch validators for bulk operations**
   ```go
   bv := ValidateDealBatch(deals)
   if bv.HasErrors() {
     // Return all errors at once
     return bv.Error()
   }
   ```

3. **Monitor constraint violations**
   ```go
   stats := bv.CalculateStats(len(deals))
   log.Metrics("constraint_failure_rate", stats.FailureRate)
   ```

4. **Index only frequent queries**
   ```
   If query runs <1x per day: Don't index
   If query runs >1x per second: Must index
   ```

5. **EXPLAIN ANALYZE before production**
   ```sql
   EXPLAIN ANALYZE SELECT ... WHERE ...;
   -- Check execution plan uses your index
   ```

---

## Files Reference

| File | Purpose |
|------|---------|
| `102_create_indexes.sql` | Create 23 production indexes |
| `103_add_constraints.sql` | Add 29 constraints + triggers |
| `constraints.go` | Application validation layer |
| `database-constraints.md` | Complete reference (500+ lines) |

---

**Last Updated:** 2026-03-26
**Compliance:** SOC2 A/B/C, WvdA Soundness, Armstrong Supervision
