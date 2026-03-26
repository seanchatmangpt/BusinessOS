# Data Seeding Guide — Fortune 5 Development/Testing

**Version:** 1.0
**Last Updated:** 2026-03-26
**Status:** Production-Ready

This guide explains how to seed realistic test data for Fortune 5 systems (BusinessOS, Canopy, OSA, pm4py-rust).

---

## Quick Start

### Option 1: Shell Script (Recommended)
```bash
cd /Users/sac/chatmangpt
DATABASE_URL="postgresql://user:pass@localhost/business_os" \
  ./scripts/seed-dev-data.sh
```

### Option 2: Pure SQL
```bash
psql "$DATABASE_URL" -f scripts/seed-rdf-data.sql
```

### Option 3: Go Seeder (Integrated)
```bash
cd BusinessOS/desktop/backend-go
go run ./cmd/seed -email "user@example.com"
```

---

## What Gets Seeded

### 1. Deals (10 records)
Financial instruments across 7 domains:

| Domain | Count | Example | Status Mix |
|--------|-------|---------|-----------|
| **Equity** | 3 | TechCorp Series B, Healthcare Merger, GreenEnergy IPO | draft, negotiating, approved, executed |
| **Fixed Income** | 3 | Corporate Bonds, Government Debt, Municipal Bonds | draft, settled, executed, closed |
| **Derivatives** | 2 | Currency Swaps, Equity Options | proposed, executed |
| **Commodities** | 1 | Crude Oil Futures | executed |
| **Structured** | 1 | Real Estate Partnership | closed |

**Key Fields:**
- Amount in cents (avoids floating-point issues): $50M → 5,000,000,000 cents
- Currency: USD, EUR, GBP, etc. (ISO 4217)
- Status lifecycle: draft → proposed → negotiating → approved → executed → settled → closed
- Domain categorization for financial regulation
- Risk ratings: AAA through D (S&P-equivalent)
- Compliance flags: KYC, AML, SOX verification
- Deterministic internal_reference: SEED-DEAL-NNN for easy cleanup

**Example Record:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440001",
  "name": "TechCorp Series B Round",
  "amount_cents": 5000000000,
  "currency": "USD",
  "status": "executed",
  "domain": "equity",
  "risk_rating": "BB",
  "created_by": "<user_id>",
  "internal_reference": "SEED-DEAL-001",
  "created_at": "2026-02-25T...",
  "deal_date": "2026-02-15",
  "settlement_date": "2026-03-01"
}
```

---

### 2. Datasets (30 records)
Operational data across 3 domains:

| Domain | Count | Categories | Privacy Level |
|--------|-------|-----------|--------------|
| **Finance** | 10 | Ledgers, P&L, AR/AP | Internal |
| **Operations** | 10 | Supply Chain, Inventory | Internal |
| **Sales** | 10 | Pipeline, Regions, Forecast | Confidential |

**Key Fields:**
- Record count (realistic volumes): 250 to 10,000 rows
- Size in MB (for storage planning): 15 to 95 MB
- Privacy level: public, internal, confidential, restricted
- Metadata JSONB: source, category, format, retention_policy
- Timestamps spread over last 30 days (realistic distribution)

**Example Record:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440101",
  "name": "Financial_Ledger_Q1",
  "description": "Quarterly general ledger with P&L transactions",
  "domain": "Finance",
  "record_count": 1000,
  "size_mb": 50,
  "privacy_level": "internal",
  "created_by": "<user_id>",
  "metadata": {
    "source": "seed",
    "category": "ledger",
    "format": "csv",
    "retention_days": 2555
  }
}
```

---

### 3. PHI Records (20 records)
Protected Health Information for healthcare compliance testing:

| Record Type | Count | Example |
|-----------|-------|---------|
| **Patient Demographics** | 5 | 5 unique patients (MRN, DOB, name) |
| **Observations** | 8 | Vital signs, lab results |
| **Medications** | 7 | Drug, dose, frequency, route |

**Key Fields:**
- Patient ID: deterministic (PAT-001 to PAT-005)
- Encrypted data: stored as JSON text (simulated encryption)
- Access audit: `accessed_by` array tracks who viewed record
- HIPAA compliance: requires audit trail for each access
- Created over last 28 days (realistic patient interaction history)

**Example Records:**
```json
{
  "id": "...",
  "patient_id": "PAT-001",
  "record_type": "patient_demographics",
  "encrypted_data": "{\"name\":\"John Doe\",\"dob\":\"1960-05-15\",\"mrn\":\"MRN-45821\"}",
  "created_by": "<user_id>",
  "created_at": "2026-02-27T..."
}
```

---

### 4. Audit Events (50 records)
Immutable audit trail with chain integrity (hash chain for compliance):

| Event Type | Count | Category |
|-----------|-------|----------|
| **Deal Created** | 10 | ProcessMining |
| **Deal Updated** | 10 | ProcessMining |
| **Dataset Analyzed** | 15 | ProcessMining |
| **Audit Log Accessed** | 15 | Compliance |

**Key Fields:**
- Sequence number: ordered (1 to 50)
- Entry hash / Previous hash: SHA-256 chain for integrity
- Event type: deal_created, deal_updated, dataset_analyzed, audit_log_accessed
- Severity: info, warning, critical (for alerting)
- Resource tracking: links to deals, datasets, PHI records
- Payload: JSON context (what changed, why, by whom)

**Chain Integrity Example:**
```
Event 1: entry_hash=0000...001, previous_hash=0000...000
Event 2: entry_hash=0000...002, previous_hash=0000...001
Event 3: entry_hash=0000...003, previous_hash=0000...002
... (chain continues)
```

---

## Installation & Setup

### Prerequisites
- PostgreSQL 12+ running
- Database `business_os` created
- User with full database access
- Environment variable `DATABASE_URL` set or pass `--db-url`

### Database URL Format
```bash
# Local development
DATABASE_URL="postgresql://postgres:password@localhost:5432/business_os"

# Cloud SQL (Supabase)
DATABASE_URL="postgresql://user:pass@db.supabase.co/postgres"

# Docker Compose (BusinessOS)
DATABASE_URL="postgresql://postgres:postgres@postgres:5432/business_os"
```

### Load .env (Optional)
If using Docker Compose, create `.env` in root:
```bash
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/business_os
```

Then source it:
```bash
source .env
```

---

## Running Seed Scripts

### 1. Shell Script with Progress

```bash
./scripts/seed-dev-data.sh
```

**Output:**
```
Checking database connectivity...
Database connected
Using user: 12345678-1234-1234-1234-123456789012
Created workspace: 87654321-4321-4321-4321-210987654321

Generating seed data...
Seeding deals...
Seeded 10 deals
Seeding datasets...
Seeded 30 datasets
Seeding PHI (Protected Health Information) records...
Seeded 20 PHI records
Seeding audit log entries...
Seeded 50 audit log entries
Verifying seed data...

=== Seed Complete ===
Deals:        10 (target: 10)
Datasets:     30 (target: 30)
PHI Records:  20 (target: 20)
Audit Events: 50 (target: 50)

Total records: 110
```

### 2. Pure SQL Script

```bash
psql "$DATABASE_URL" -f scripts/seed-rdf-data.sql
```

**Benefits:**
- Zero dependencies (no shell tools needed)
- Directly insertable into PostgreSQL
- Works with any SQL client
- Idempotent (safe to run multiple times)

### 3. Go Seeder (Integrated CLI)

```bash
cd BusinessOS/desktop/backend-go

# By email (auto-lookup)
go run ./cmd/seed -email "user@example.com"

# By user ID
go run ./cmd/seed -user-id "12345678-1234-1234-1234-123456789012"

# Re-seed (delete existing, insert fresh)
go run ./cmd/seed -email "user@example.com" -force
```

**Benefits:**
- Type-safe (uses Go structs)
- Integrated with codebase (same dependencies)
- Handles user lookup automatically
- Better error messages

---

## Idempotency & Cleanup

### Why Idempotent?
All seed records use deterministic IDs and internal_references (e.g., SEED-DEAL-001). Running the seeder twice produces the same result—no duplicates.

### Checking Seed Status

```bash
# Count deals
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM deals WHERE internal_reference LIKE 'SEED-%';"

# Count datasets
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM datasets WHERE metadata->>'source' = 'seed';"

# Count PHI records
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM phi_records LIMIT 1;"

# Count audit events
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM audit_events WHERE payload->>'action' = 'seed_data';"
```

### Cleaning Up Seed Data

```bash
# Option 1: Shell script with --force
./scripts/seed-dev-data.sh --force

# Option 2: Go seeder with -force
go run ./cmd/seed -email "user@example.com" -force

# Option 3: Manual SQL
psql "$DATABASE_URL" <<EOF
DELETE FROM audit_events WHERE payload->>'action' = 'seed_data';
DELETE FROM deals WHERE internal_reference LIKE 'SEED-%';
DELETE FROM datasets WHERE metadata->>'source' = 'seed';
DELETE FROM phi_records;  -- No cleanup needed (separate table)
EOF
```

---

## Customization & Extension

### Add More Deals
Edit `scripts/seed-rdf-data.sql` and add INSERT statements:

```sql
INSERT INTO deals (id, name, amount_cents, currency, status, domain, ...)
VALUES
  (gen_random_uuid(), 'Your Deal Name', 1000000000, 'USD', 'draft', 'equity', ...);
```

Or use Go seeder (`BusinessOS/desktop/backend-go/cmd/seed/deals.go`):

```go
deals := []Deal{
  // ... existing deals
  {
    uuid.New(),
    "Your Custom Deal",
    1000000000,
    "USD",
    "draft",
    "equity",
    userID,
    // ... other fields
  },
}
```

### Add More Datasets
Edit SQL or Go seeder similarly:

```sql
INSERT INTO datasets (id, name, description, domain, record_count, size_mb, ...)
VALUES (gen_random_uuid(), 'Custom_Dataset_Name', ..., 'Finance', 5000, 100, ...);
```

### Add More PHI Records
Follow the same pattern. Ensure patient IDs are deterministic for reproducibility:

```sql
INSERT INTO phi_records (patient_id, record_type, encrypted_data, created_by, created_at)
VALUES ('PAT-099', 'observation', '{"type":"vital_signs","bp":"120/80"}', %USER_ID%, NOW());
```

### Add More Audit Events
Manually insert with incrementing sequence numbers:

```sql
INSERT INTO audit_events (
  sequence_number, entry_hash, previous_hash, event_id, event_type,
  event_category, created_at, severity, user_id, resource_type, payload
) VALUES
  (51, '0000...033', '0000...032', gen_random_uuid(), 'deal_created', ...);
```

---

## Testing & Validation

### Unit Tests (Go Seeder)

```bash
cd BusinessOS/desktop/backend-go
go test ./cmd/seed/... -v
```

### Integration Tests (Full Stack)

```bash
# 1. Seed data
DATABASE_URL="..." ./scripts/seed-dev-data.sh

# 2. Run API tests
cd BusinessOS/frontend
npm test

# 3. Verify in UI
# Navigate to dashboard and check deals, datasets are visible
```

### Manual Verification

```bash
# Check deal distribution by domain
psql "$DATABASE_URL" -c "
  SELECT domain, COUNT(*) FROM deals
  WHERE internal_reference LIKE 'SEED-%'
  GROUP BY domain
  ORDER BY COUNT(*) DESC;"

# Check audit chain integrity
psql "$DATABASE_URL" -c "
  SELECT sequence_number, entry_hash, previous_hash
  FROM audit_events
  WHERE payload->>'action' = 'seed_data'
  ORDER BY sequence_number
  LIMIT 10;"

# Check PHI record types
psql "$DATABASE_URL" -c "
  SELECT record_type, COUNT(*)
  FROM phi_records
  GROUP BY record_type
  ORDER BY COUNT(*) DESC;"
```

---

## Seeding in Docker Compose

### Using docker-compose.yml

```yaml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: business_os
      POSTGRES_PASSWORD: postgres

  seed:
    image: golang:1.24
    working_dir: /app/BusinessOS/desktop/backend-go
    command: |
      sh -c "
        sleep 10 &&  # Wait for postgres to start
        go run ./cmd/seed -email user@example.com
      "
    depends_on:
      - postgres
    volumes:
      - .:/app
    environment:
      DATABASE_URL: postgresql://postgres:postgres@postgres:5432/business_os
```

### Run Seeding

```bash
docker-compose up seed
```

---

## Performance & Scalability

### Seeding Time
| Script | Records | Time |
|--------|---------|------|
| Shell script | 110 | < 1 second |
| SQL script | 110 | < 1 second |
| Go seeder | 110+ (varies) | 2-5 seconds |

### Database Impact
- No full-table locks (uses `ON CONFLICT ... DO NOTHING`)
- Idempotent (safe to run anytime)
- All indices updated automatically
- No trigger-based cascades (deals table only)

### Optimization Tips
1. **Batch inserts:** Use `INSERT INTO ... VALUES (...), (...), ...` (already done)
2. **Disable triggers:** Not needed for seed data (lightweight triggers)
3. **Parallel execution:** Seeds are independent (no blocking)
4. **Index reuse:** Don't drop indices before seeding

---

## Troubleshooting

### Error: "No users in database"
Create a user first:
```bash
psql "$DATABASE_URL" -c "
  INSERT INTO \"user\" (id, name, email, \"emailVerified\")
  VALUES ('test-user-123', 'Test User', 'test@example.com', true);"
```

### Error: "Table phi_records does not exist"
The shell script creates it automatically. If using pure SQL, ensure the CREATE TABLE statement runs first:
```bash
psql "$DATABASE_URL" -f scripts/seed-rdf-data.sql
```

### Error: "ON CONFLICT not supported"
Update PostgreSQL to 9.5+ (the codebase requires 12+).

### Error: "Permission denied"
Ensure your user has:
- CREATE TABLE (for phi_records)
- INSERT on all tables
- SELECT on "user" table (for lookup)

```bash
psql "$DATABASE_URL" -c "GRANT ALL ON ALL TABLES IN SCHEMA public TO your_user;"
```

### Records Not Showing Up
1. Verify database connection:
   ```bash
   psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM deals;"
   ```
2. Check for ON CONFLICT (idempotent):
   ```bash
   psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM deals WHERE internal_reference LIKE 'SEED-%';"
   ```
3. Inspect audit log:
   ```bash
   psql "$DATABASE_URL" -c "SELECT * FROM audit_events LIMIT 1;"
   ```

---

## Production Considerations

### Sensitive Data
- PHI records are NOT encrypted (sample data only)
- For production, use real encryption (pgcrypto, transparent encryption)
- Audit log hashes are deterministic (for testing only)

### Compliance
- Seed data is marked with `source: 'seed'` in metadata
- Use date range queries to exclude seed data from reports:
  ```sql
  SELECT * FROM deals
  WHERE internal_reference NOT LIKE 'SEED-%'
    AND created_at >= DATE_TRUNC('year', CURRENT_DATE);
  ```
- Audit events are immutable (hash chain prevents modification)

### Cleanup Strategy
1. **Before production deployment:** Delete all seed data
2. **Test environment:** Keep seed data for regression testing
3. **Development environment:** Reseed daily if needed

---

## Related Documentation

- **CLAUDE.md** — Project architecture and build commands
- **BusinessOS/CLAUDE.md** — Backend structure and DB patterns
- **docs/DEPENDENCIES.md** — PostgreSQL versions and compatibility
- **docs/VERSION_COMPATIBILITY_SUMMARY.md** — Supported language versions

---

## FAQ

**Q: How do I seed data without losing existing records?**
A: The seeder uses `ON CONFLICT ... DO NOTHING`, so it's idempotent. Running twice = same result.

**Q: Can I seed different amounts (20 deals instead of 10)?**
A: Yes. Edit the SQL or Go seeder files and add/remove records. The IDs and references are deterministic.

**Q: What if I need custom data for a specific test?**
A: Use the Go seeder and extend `deals.go`, `datasets.go`, etc. with your custom logic.

**Q: Can I seed to production?**
A: Not recommended. The shell script marks all records with "SEED-" prefix for easy cleanup. Use only in dev/test.

**Q: What's the order of seeding?**
A: 1. Deals, 2. Datasets, 3. PHI Records, 4. Audit Events. No dependencies between tables.

---

**Version History:**
- v1.0 (2026-03-26) — Initial release with 110 seed records (10 deals, 30 datasets, 20 PHI, 50 audit events)
