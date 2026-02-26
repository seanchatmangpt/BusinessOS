# Migration Scripts

This directory contains scripts for testing, deploying, and managing database migrations.

---

## Available Scripts

### test_migrations.sh
**Automated migration testing framework**

Tests migrations 052-054, 088-089 in an isolated environment.

**Usage:**
```bash
# Test all migrations
./test_migrations.sh

# Test specific migration
./test_migrations.sh --specific 052

# Test rollback capability
./test_migrations.sh --rollback

# Test on staging (requires confirmation)
./test_migrations.sh --staging
```

**What it does:**
1. Creates temporary test database
2. Applies migrations in correct order
3. Verifies schema integrity
4. Tests CRUD operations
5. Checks for conflicts
6. Cleans up automatically

**Exit codes:**
- `0`: Success
- `1`: Failure (check output for details)

---

## Other Scripts

### run_migration_036.go
Example Go program for running a specific migration. Can be used as template for custom migration runners.

### run_rag_benchmarks.sh / .ps1
RAG (Retrieval-Augmented Generation) benchmark scripts for testing vector search performance.

### run_workspace_tests.sh / .ps1
Workspace functionality test scripts.

---

## Documentation

For comprehensive migration documentation, see:

- **[MIGRATIONS.md](../../docs/MIGRATIONS.md)** - Complete migration guide
- **[STAGING_MIGRATION_PLAN.md](../../docs/STAGING_MIGRATION_PLAN.md)** - Staging deployment plan
- **[MIGRATION_QUICKSTART.md](../../docs/MIGRATION_QUICKSTART.md)** - Quick reference
- **[MIGRATION_VALIDATION_SUMMARY.md](../../docs/MIGRATION_VALIDATION_SUMMARY.md)** - Migration analysis

---

## Prerequisites

### For test_migrations.sh
- PostgreSQL client (`psql`) installed
- `DATABASE_URL` environment variable set
- Access to test/staging database

### Installation

**macOS (Homebrew):**
```bash
brew install postgresql@15
```

**Ubuntu/Debian:**
```bash
sudo apt-get install postgresql-client
```

**Windows:**
- Download from [PostgreSQL Downloads](https://www.postgresql.org/download/windows/)
- Or use WSL with Ubuntu instructions above

---

## Environment Variables

Required environment variables (typically in `.env`):

```bash
DATABASE_URL=postgres://user:password@host:port/database
```

Optional:
```bash
STAGING_DB_URL=postgres://user:password@staging.host:port/database
```

---

## Migration Files

Migration files are located in:
```
desktop/backend-go/internal/database/migrations/
```

Current migrations being tested:
- `052_workspace_versions.sql`
- `053_onboarding_email_metadata.sql`
- `054_custom_modules.sql`
- `088_seed_builtin_templates.sql`
- `089_app_generation_system.sql`

---

## Troubleshooting

### Permission Denied
```bash
chmod +x test_migrations.sh
```

### psql: command not found
Install PostgreSQL client (see Prerequisites above).

### DATABASE_URL not set
```bash
# Load from .env
export $(grep -v '^#' ../../.env | xargs)

# Or set directly
export DATABASE_URL="postgres://..."
```

### Connection Refused
- Check if PostgreSQL is running
- Verify connection string (host, port, credentials)
- Check firewall rules

### Migration Fails
1. Check logs in `/tmp/migration_output.log`
2. Review migration file for syntax errors
3. Verify dependencies (tables, foreign keys)
4. Consult [MIGRATIONS.md troubleshooting section](../../docs/MIGRATIONS.md#troubleshooting)

---

## Best Practices

1. **Always test locally first**
   ```bash
   ./test_migrations.sh
   ```

2. **Test rollback before deploying**
   ```bash
   ./test_migrations.sh --rollback
   ```

3. **Backup before staging deployment**
   ```bash
   pg_dump "$DATABASE_URL" > backup_$(date +%Y%m%d).sql
   ```

4. **Regenerate sqlc code after schema changes**
   ```bash
   cd ../..
   sqlc generate
   ```

5. **Verify application compiles**
   ```bash
   cd ../..
   go build ./...
   ```

---

## Contributing

When adding new migration scripts:

1. Follow naming convention: `action_description.sh`
2. Include usage documentation (header comments)
3. Add exit code handling (`set -e`)
4. Include cleanup logic (traps)
5. Update this README

Example header:
```bash
#!/usr/bin/env bash
# ============================================================================
# Script Name: test_new_feature.sh
# Description: Tests new feature migrations
# Usage: ./test_new_feature.sh [options]
# ============================================================================
```

---

## Support

For questions or issues:

1. Check documentation in `docs/` directory
2. Review existing scripts for examples
3. Contact database team: #database-migrations

---

**Last Updated**: 2026-01-26
**Maintained By**: BusinessOS Database Team
