# Database Migration Guide

**BusinessOS Schema Versioning & Evolution Framework**

This guide covers how to create, apply, and rollback database migrations using the BusinessOS migration framework.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Migration Concepts](#migration-concepts)
3. [Creating Migrations](#creating-migrations)
4. [Running Migrations](#running-migrations)
5. [Rollback Procedures](#rollback-procedures)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)
8. [Monitoring](#monitoring)
9. [Version Control](#version-control)

---

## Quick Start

### Apply All Pending Migrations

```bash
cd BusinessOS
./scripts/migrate.sh up
```

### Check Migration Status

```bash
./scripts/migrate.sh status
```

### Rollback Last Migration

```bash
./scripts/migrate.sh down 1
```

### Show Current Schema Version

```bash
./scripts/migrate.sh version
```

---

## Migration Concepts

### What is a Migration?

A migration is a pair of SQL files that define schema changes:

- **up migration** (`NNN_name.sql`): Applies the schema change
- **down migration** (`rollback_NNN_name.sql`): Reverts the schema change

### Why Migrations?

1. **Version Control**: Track all schema changes with Git history
2. **Reproducibility**: Rebuild schema from scratch in any environment
3. **Rollback**: Safely revert changes if needed
4. **Audit Trail**: `schema_migrations` table records all applied versions
5. **Team Collaboration**: Multiple developers can work on schema changes safely

### Migration Lifecycle

```
Development Environment
  → Create migration pair (up.sql + down.sql)
  → Test locally: ./scripts/migrate.sh up
  → Test rollback: ./scripts/migrate.sh down 1
  → Commit to Git

Staging Environment
  → Deploy code
  → Run: ./scripts/migrate.sh up
  → Test application with new schema
  → Monitor for issues

Production Environment
  → Backup database (CRITICAL!)
  → Run: ./scripts/migrate.sh up
  → Monitor query performance
  → If issues: ./scripts/migrate.sh down 1
```

### Schema Migrations Table

Every database contains a `schema_migrations` table that tracks applied migrations:

```sql
SELECT * FROM schema_migrations ORDER BY applied_at DESC;

-- Output:
version  | applied_at                | checksum         | status
---------+---------------------------+------------------+--------
010      | 2026-03-26 14:32:10+00   | a1b2c3d4...     | success
009      | 2026-03-26 14:31:45+00   | e5f6g7h8...     | success
008      | 2026-03-26 14:31:20+00   | i9j0k1l2...     | success
```

---

## Creating Migrations

### Step 1: Create the Up Migration

Create a new file with 3-digit version number:

```sql
-- BusinessOS/migrations/011_add_audit_log_table.sql

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID NOT NULL,
    old_values JSONB,
    new_values JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT
);

CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);
CREATE INDEX idx_audit_log_resource ON audit_log(resource_type, resource_id);
```

### Step 2: Create the Down Migration (Rollback)

Create the corresponding rollback file:

```sql
-- BusinessOS/migrations/rollback_011_add_audit_log_table.sql

DROP TABLE IF EXISTS audit_log CASCADE;
```

### Step 3: Name Your Migration Carefully

**Naming Convention:**
- `NNN_` — Three-digit version number (001, 002, 003, ..., 100, etc.)
- `description_in_snake_case` — Clear, descriptive name
- `.sql` — File extension

**Good names:**
- `011_add_audit_log_table.sql`
- `012_add_encryption_at_rest.sql`
- `013_add_process_mining_indexes.sql`

**Bad names:**
- `migration.sql` (not descriptive)
- `add_table.sql` (no version)
- `AddTable.sql` (wrong case)

### Step 4: Update schema.yaml

Add entry to `migrations/schema.yaml`:

```yaml
migrations:
  - version: "011"
    name: "add_audit_log_table"
    filename: "011_add_audit_log_table.sql"
    checksum: "TO_BE_COMPUTED"  # Will be computed by migration runner
    description: "Create audit_log table for compliance tracking"
    timestamp: "2026-03-26T14:00:00Z"
    breaking: false
    dependencies: ["001", "002", "003", "004", "005", "006", "007", "008", "009", "010"]
```

### Step 5: Test Locally

```bash
# Apply migration
./scripts/migrate.sh up

# Verify new schema
psql $DATABASE_URL -c "SELECT * FROM audit_log LIMIT 1;"

# Test rollback
./scripts/migrate.sh down 1

# Verify table is gone
psql $DATABASE_URL -c "SELECT * FROM audit_log LIMIT 1;" # Should error

# Re-apply for further testing
./scripts/migrate.sh up
```

---

## Running Migrations

### Prerequisites

1. **Database running**: PostgreSQL must be accessible
2. **DATABASE_URL set**: `export DATABASE_URL="postgres://user:pass@localhost:5432/businessos"`
3. **Migrations directory exists**: `BusinessOS/migrations/`

### Apply All Pending Migrations

```bash
./scripts/migrate.sh up
```

Output:
```
[INFO] Applying migrations...
[INFO] Migrations directory: /Users/sac/chatmangpt/BusinessOS/migrations
[INFO] Found 5 migration file(s)
[INFO] Migrations to apply:
  - 006_add_encryption_at_rest.sql
  - 007_add_cache_tables.sql
  - 008_add_process_mining_tables.sql
  - 009_add_mcp_a2a_tables.sql
  - 010_add_governance_tables.sql

[INFO] Applying migration 006_add_encryption_at_rest.sql...
[SUCCESS] Migration applied successfully (duration: 234ms)
[INFO] Applying migration 007_add_cache_tables.sql...
[SUCCESS] Migration applied successfully (duration: 156ms)
...
[SUCCESS] All migrations applied (total: 5)
```

### Apply N Pending Migrations

```bash
./scripts/migrate.sh up 2  # Apply first 2 pending migrations only
```

---

## Rollback Procedures

### Rollback Last Migration

```bash
./scripts/migrate.sh down 1
```

### Rollback Last N Migrations

```bash
./scripts/migrate.sh down 5  # Rollback last 5 migrations
```

**CRITICAL WARNING:** Rollbacks are destructive. Always:

1. Backup database first: `pg_dump $DATABASE_URL > backup.sql`
2. Test in staging first
3. Schedule during low-traffic window
4. Have rollback plan documented

### Rollback Strategy

Migrations roll back **in reverse order**:

```
Applied order:   001 → 002 → 003 → 004 → 005
Rollback order:  005 ← 004 ← 003 ← 002 ← 001
```

If you apply 5 migrations then run `./scripts/migrate.sh down 3`, it will:

1. Rollback 005 (using rollback_005.sql)
2. Rollback 004 (using rollback_004.sql)
3. Rollback 003 (using rollback_003.sql)

### Foreign Key Constraints During Rollback

**Important:** If rollback fails due to foreign key constraints:

```sql
-- Option 1: Drop child table first, then parent
DROP TABLE IF EXISTS child_table CASCADE;
DROP TABLE IF EXISTS parent_table CASCADE;

-- Option 2: Disable FK checks temporarily (use carefully!)
ALTER TABLE child_table DISABLE TRIGGER ALL;
DROP TABLE IF EXISTS parent_table;
ALTER TABLE child_table ENABLE TRIGGER ALL;
```

Always add this to your rollback file if you created FK relationships:

```sql
-- rollback_011_add_audit_log_table.sql
DROP TABLE IF EXISTS audit_log CASCADE;
```

The `CASCADE` keyword handles dependent objects.

---

## Best Practices

### 1. Always Write Rollback FIRST

Before writing the up migration, write the down migration:

**Step 1 — Write rollback:**
```sql
-- rollback_011_add_audit_log_table.sql
DROP TABLE IF EXISTS audit_log CASCADE;
```

**Step 2 — Write up migration:**
```sql
-- 011_add_audit_log_table.sql
CREATE TABLE audit_log (...)
```

**Why?** Forces you to think about consequences before making changes.

### 2. One Logical Change Per Migration

**Good:**
```sql
-- 011_add_audit_log_table.sql
CREATE TABLE audit_log (...);
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
```

**Bad:**
```sql
-- 011_megamigration.sql
CREATE TABLE audit_log (...);
CREATE TABLE notifications (...);
CREATE TABLE webhooks (...);
-- ... 500 lines later ...
ALTER TABLE users ADD COLUMN ...;
```

### 3. Use Explicit Schema Names

```sql
-- GOOD: Explicit schema
CREATE TABLE IF NOT EXISTS public.audit_log (...)

-- BAD: Implicit schema
CREATE TABLE audit_log (...)
```

### 4. Use Idempotent Migrations

Migrations should be safe to re-run:

```sql
-- GOOD: Idempotent
CREATE TABLE IF NOT EXISTS audit_log (...)
CREATE INDEX IF NOT EXISTS idx_audit_log_user_id ON audit_log(user_id)

-- BAD: Not idempotent (fails on re-run)
CREATE TABLE audit_log (...)  -- Error: table already exists!
CREATE INDEX idx_audit_log_user_id ON audit_log(user_id)
```

### 5. Never Modify Applied Migrations

Once a migration is applied to production:

- **NEVER** modify the .sql file
- **NEVER** change the version number
- **NEVER** rename the file

If you need to fix something: **Create a new migration.**

```
Bad:
  - Applied: 011_add_audit_log_table.sql
  - Modify file (WRONG!)
  - Checksum mismatch error (deserved!)

Good:
  - Applied: 011_add_audit_log_table.sql
  - Create: 012_fix_audit_log_table_indexes.sql
  - Run: ./scripts/migrate.sh up
```

### 6. Large Schema Changes: Break Into Smaller Migrations

```
Bad:
  - 011_refactor_entire_schema.sql (500 lines, refactors 20 tables)

Good:
  - 011_add_user_profiles_table.sql
  - 012_migrate_user_data_to_profiles.sql
  - 013_update_user_fk_to_profiles.sql
  - 014_drop_old_user_columns.sql
```

Smaller migrations = easier to debug, faster rollback.

### 7. Data Migrations: Plan Carefully

```sql
-- 013_migrate_user_data_to_profiles.sql

-- Step 1: Create new table with data
CREATE TABLE user_profiles AS
SELECT
    user_id,
    email,
    phone,
    address
FROM users;

-- Step 2: Add foreign key
ALTER TABLE user_profiles
    ADD CONSTRAINT fk_user_profiles_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Step 3: Create indexes
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);

-- Step 4: Verify migration
DO $$
DECLARE
    count_users INT;
    count_profiles INT;
BEGIN
    SELECT COUNT(*) INTO count_users FROM users;
    SELECT COUNT(*) INTO count_profiles FROM user_profiles;
    IF count_users != count_profiles THEN
        RAISE EXCEPTION 'Data migration failed: % users, % profiles', count_users, count_profiles;
    END IF;
END $$;
```

### 8. Backup Before Production Migrations

```bash
# Backup before migration
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql

# Run migration
./scripts/migrate.sh up

# If problems occur, restore from backup
psql $DATABASE_URL < backup_20260326_143210.sql
```

### 9. Test Migrations in All Environments

```
Development (local machine)
  → ./scripts/migrate.sh up
  → Test application
  → ./scripts/migrate.sh down 1
  → ./scripts/migrate.sh up

Staging (mirrors production)
  → Backup database
  → ./scripts/migrate.sh up
  → Run integration tests
  → Run performance tests
  → Monitor logs for 24 hours
  → If OK: approve for production

Production
  → Backup database
  → Schedule low-traffic window
  → ./scripts/migrate.sh up
  → Monitor application
  → Watch query performance
```

### 10. Document Breaking Changes

```yaml
# migrations/schema.yaml
- version: "006"
  name: "add_encryption_at_rest"
  breaking: true  # <-- Mark as breaking!
  migration_notes: |
    BREAKING CHANGE: All sensitive columns now encrypted.

    Application Changes Required:
    1. Update model: add `encryption_key_id UUID` field
    2. Update queries: use `pgcrypto.pgp_sym_decrypt()` when reading
    3. Update inserts: use `pgcrypto.pgp_sym_encrypt()` when writing
    4. Deploy new code BEFORE running this migration

    Rollback: Loses encrypted data (use backup to restore)
```

---

## Troubleshooting

### Checksum Mismatch Error

```
[ERROR] Migration 011: checksum mismatch
  Expected: a1b2c3d4...
  Actual: e5f6g7h8...
```

**Cause:** Migration file was modified after being applied to production.

**Fix:**
1. Don't modify applied migrations!
2. If you must fix something, create a new migration

```bash
# WRONG: Don't do this
vim BusinessOS/migrations/011_add_audit_log_table.sql

# RIGHT: Create a new migration
cp BusinessOS/migrations/011_add_audit_log_table.sql BusinessOS/migrations/012_fix_audit_log_table.sql
vim BusinessOS/migrations/012_fix_audit_log_table.sql
```

### Migration Fails Mid-Way

```
[ERROR] Migration 011: execute migration: ERROR: column "user_id" already exists
```

**Cause:** Migration tried to add column that already exists, or FK constraint failed.

**Fix:**

1. Make migration idempotent:
```sql
-- WRONG
ALTER TABLE users ADD COLUMN email TEXT;

-- RIGHT
ALTER TABLE users ADD COLUMN IF NOT EXISTS email TEXT;
```

2. Check for dependencies:
```bash
# List applied migrations
./scripts/migrate.sh status

# Check if prior migrations succeeded
psql $DATABASE_URL -c "SELECT * FROM schema_migrations ORDER BY version;"
```

3. Manual recovery (last resort):
```bash
# Restore from backup
pg_dump $DATABASE_URL > backup_$(date +%Y%m%d_%H%M%S).sql
pg_restore $DATABASE_URL < backup_20260326_140000.sql

# Remove failed migration from schema_migrations
psql $DATABASE_URL -c "DELETE FROM schema_migrations WHERE version = '011';"

# Fix migration file and retry
./scripts/migrate.sh up
```

### Connection Refused

```
[ERROR] Applying migration: connection refused
```

**Cause:** PostgreSQL not running or DATABASE_URL is wrong.

**Fix:**
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Check DATABASE_URL
echo $DATABASE_URL

# Test connection directly
psql $DATABASE_URL -c "SELECT 1"
```

### Permission Denied

```
[ERROR] Read migration file: permission denied
```

**Fix:**
```bash
# Make migration files readable
chmod 644 BusinessOS/migrations/*.sql

# Make script executable
chmod +x BusinessOS/scripts/migrate.sh
```

### "No rollback file found" Error

```
[ERROR] Migration 011: no rollback file (rollback_*.sql not found)
```

**Cause:** You tried to rollback but rollback file doesn't exist.

**Fix:**
```bash
# Create the rollback file
cat > BusinessOS/migrations/rollback_011_add_audit_log_table.sql << 'EOF'
DROP TABLE IF EXISTS audit_log CASCADE;
EOF

# Now rollback should work
./scripts/migrate.sh down 1
```

### Table Still Exists After Rollback

```bash
# Verify table was dropped
psql $DATABASE_URL -c "\dt audit_log"
# Did NOT exist: no output
# Still exists: shows table definition (rollback didn't work!)
```

**Fix:**
```sql
-- Check rollback_011.sql for syntax errors
cat BusinessOS/migrations/rollback_011_add_audit_log_table.sql

-- Manual rollback if necessary
psql $DATABASE_URL -c "DROP TABLE IF EXISTS audit_log CASCADE;"
```

---

## Monitoring

### Check Migration Status

```bash
./scripts/migrate.sh status
```

### Verify All Migrations Applied

```bash
./scripts/migrate.sh verify
```

### Current Schema Version

```bash
./scripts/migrate.sh version
# Output: Current schema version: 010
```

### Monitor Applied Migrations

```sql
-- List all applied migrations with times
SELECT
    version,
    applied_at,
    EXTRACT(EPOCH FROM (NOW() - applied_at)) AS seconds_ago,
    checksum
FROM schema_migrations
ORDER BY applied_at DESC;

-- Find slow migrations
SELECT
    version,
    duration_ms,
    CASE
        WHEN duration_ms > 5000 THEN 'SLOW'
        WHEN duration_ms > 1000 THEN 'MODERATE'
        ELSE 'FAST'
    END AS speed
FROM schema_migrations
ORDER BY duration_ms DESC;

-- Check for failed migrations
SELECT * FROM schema_migrations WHERE status != 'success';
```

### Monitor Query Performance During Migration

```bash
# Watch for slow queries while migration is running
watch -n 1 'psql $DATABASE_URL -c "SELECT query, mean_exec_time, calls FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 5;"'
```

---

## Version Control

### Committing Migrations to Git

```bash
# Create feature branch for schema change
git checkout -b feat/add-audit-log

# Create migration files
touch BusinessOS/migrations/011_add_audit_log_table.sql
touch BusinessOS/migrations/rollback_011_add_audit_log_table.sql

# Edit migration files with your changes
vim BusinessOS/migrations/011_add_audit_log_table.sql
vim BusinessOS/migrations/rollback_011_add_audit_log_table.sql

# Test locally
./scripts/migrate.sh up
./scripts/migrate.sh down 1
./scripts/migrate.sh up

# Stage migration files
git add BusinessOS/migrations/011_add_audit_log_table.sql
git add BusinessOS/migrations/rollback_011_add_audit_log_table.sql
git add BusinessOS/migrations/schema.yaml

# Commit with descriptive message
git commit -m "feat(db): add audit log table for compliance tracking

- Create audit_log table with columns: id, user_id, action, resource_type, resource_id, old_values, new_values, created_at, ip_address, user_agent
- Add indexes on user_id, created_at, and resource (type, id) for query performance
- Add rollback migration to drop table

Requires:
- Users must have 'id' column (added in migration 001)
- Migration runs in transaction, auto-rolls back on error
- No data loss on rollback (table will be dropped)"
```

### Version Control Best Practices

1. **Never skip migrations**: Don't cherry-pick migrations from Git history
2. **Linear history**: Maintain linear migration sequence (001, 002, 003, ...)
3. **Document breaking changes**: Note in commit message if migration breaks API
4. **Include dependencies**: Commit related app code changes with migration
5. **Code review**: Have migrations reviewed before merge

```bash
# Example PR (GitHub):

## Description
Add audit logging for compliance tracking

## Schema Changes
- Creates audit_log table
- Adds indexes for performance
- Rollback supported

## Migration
- Version: 011_add_audit_log_table.sql
- Backward compatible: ✅ (non-breaking)
- Tested locally: ✅
- Tested in staging: ✅
- Rollback tested: ✅

## Application Changes
- model/audit_log.go: +50 lines
- service/audit_service.go: +200 lines
- handler/audit_handler.go: +100 lines

## Testing
- Unit tests: 25 new tests, all passing
- Integration tests: 5 new tests, all passing
- Manual test in staging: ✅

## Merge Requirements
- [ ] Code review approved
- [ ] Migration tested in staging
- [ ] Rollback tested locally
- [ ] No other pending migrations
```

---

## Summary

**Key Takeaways:**

1. **Always write rollback first** — forces thinking about consequences
2. **One change per migration** — easier to debug and rollback
3. **Test locally first** — apply up, test app, rollback, apply up again
4. **Backup before production** — last resort recovery
5. **Never modify applied migrations** — creates new one instead
6. **Monitor during migration** — watch for slow queries, locks
7. **Version control everything** — Git tracks all schema changes

---

## Additional Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Migration Best Practices](https://wiki.postgresql.org/wiki/Safely_renaming_a_table_containing_generated_columns)
- [Data Migration Patterns](https://www.depesz.com/2010/08/22/why-is-my-query-slow/)
- BusinessOS Architecture: `docs/architecture.md`
