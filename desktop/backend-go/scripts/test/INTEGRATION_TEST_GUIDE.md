# Integration Test Guide

## Quick Start

### Prerequisites
1. PostgreSQL database running and accessible
2. `.env` file configured with `DATABASE_URL`
3. Go 1.24+ installed
4. Ollama (optional, for embedding service testing)

### Run Tests
```bash
cd /Users/rhl/Desktop/BusinessOS2/desktop/backend-go

# Run all integration tests
bash scripts/test/run_integration_tests.sh

# Run individual tests
go run scripts/test/test_db_connectivity.go
go run scripts/test/test_voice_pipeline.go
```

---

## Test Descriptions

### 1. Database Connectivity Test (`test_db_connectivity.go`)

**What it does:**
- Connects to PostgreSQL using `DATABASE_URL`
- Tests basic query execution
- Verifies voice-related tables exist
- Checks connection pool statistics

**Expected output:**
```
✅ Database connectivity test completed successfully!
```

**If it fails:**

#### Error: "DATABASE_URL not set"
```bash
# Solution: Set DATABASE_URL in .env or environment
export DATABASE_URL="postgres://user:pass@localhost/businessos"
```

#### Error: "Failed to connect to database"
```bash
# Check PostgreSQL is running
pg_isready -h localhost

# Verify credentials
psql postgresql://user:pass@localhost/businessos -c "SELECT 1"
```

#### Error: "Table not found"
- Some tables like `agent_v2` and `embeddings` are created in Phase 5
- This is expected and non-blocking
- Core tables should exist: `user`, `workspaces`, `workspace_members`

### 2. Voice Pipeline Test (`test_voice_pipeline.go`)

**What it does:**
- Loads configuration from `.env`
- Initializes database connection pool
- Creates embedding service instance
- Verifies voice system database schema
- Checks sample data availability

**Expected output:**
```
✅ Voice System Pipeline Test Summary
All critical components initialized successfully!
```

**If it fails:**

#### Error: "Failed to load config"
```bash
# Ensure all required .env variables are set:
DATABASE_URL
ENVIRONMENT
AI_PROVIDER
```

#### Error: "Initializing embedding service" warning
- This is normal if Ollama is not running
- Voice system still works without embeddings
- To enable: Start Ollama on http://localhost:11434

#### Error: "agent_v2 table not found"
- This is expected - table created in Phase 5
- Not a blocker for current tests

---

## Environment Configuration

### Required Variables
```env
# Database
DATABASE_URL=postgres://user:password@localhost:5432/businessos

# Environment
ENVIRONMENT=development

# AI Provider
AI_PROVIDER=ollama_local  # or: ollama_cloud, anthropic, groq
OLLAMA_LOCAL_URL=http://localhost:11434
```

### Optional Variables
```env
# API Keys (only if using cloud providers)
ANTHROPIC_API_KEY=...
GROQ_API_KEY=...
OLLAMA_CLOUD_API_KEY=...

# Embedding
EMBEDDING_MODEL=nomic-embed-text
```

---

## Common Issues & Solutions

### Issue: Connection Pool Error
```
❌ Failed to create connection pool
Error: could not translate host name
```

**Solution:**
- Check PostgreSQL host/port in DATABASE_URL
- Default: `localhost:5432`
- For remote: `postgres://user:pass@host:port/db`

### Issue: Query Execution Failed
```
❌ Failed to query user table
Error: relation "user" does not exist
```

**Solution:**
- Run database migrations
- Check database is properly initialized
- Verify schema exists: `psql -c "\d user"`

### Issue: Voice Pipeline Slow
```
🧪 Testing Voice System Pipeline...
[hangs for 30+ seconds]
```

**Solution:**
- Embedding service initialization takes time
- If Ollama not running, service still initializes (optional)
- This is normal on first run

### Issue: Test Hangs
```
^C (stuck waiting)
```

**Solution:**
```bash
# Kill stuck go processes
killall go

# Run again with timeout
timeout 30 go run scripts/test/test_db_connectivity.go
```

---

## Database Schema Reference

### Voice-Related Tables

#### `workspace_members`
```sql
-- Members of a workspace with roles
SELECT * FROM workspace_members LIMIT 1;
-- Columns: id, workspace_id, user_id, role, created_at
```

#### `user_workspace_profiles`
```sql
-- User profiles within workspaces
SELECT * FROM user_workspace_profiles LIMIT 1;
-- Columns: user_id, workspace_id, display_name, metadata, created_at
```

#### `workspaces`
```sql
-- Workspace data
SELECT * FROM workspaces LIMIT 1;
-- Columns: id, name, owner_id, created_at, updated_at
```

#### `user` (Core)
```sql
-- User accounts
SELECT id, email FROM "user" LIMIT 1;
-- Columns: id, email, created_at, ...
-- Note: username column needs to be added in migration
```

---

## Test Output Interpretation

### Green Check (✅)
Indicates a successful test or component initialization.

### Yellow Warning (⚠️)
Indicates an optional component is missing or a known limitation.
- Example: `agent_v2` table not found (created in Phase 5)

### Red Error (❌)
Indicates a critical failure requiring attention.

---

## Advancing to Phase 5

After integration tests pass, Phase 5 adds:

1. **Database Migrations**
   - Create `agent_v2` table
   - Create `embeddings` table
   - Add user columns (username, etc.)

2. **Agent System**
   - Agent V2 registry implementation
   - Agent lifecycle management

3. **Embedding System**
   - Vector storage and retrieval
   - Similarity search

4. **Voice Server**
   - gRPC voice server
   - Protocol buffers
   - Voice handlers

---

## Debugging Commands

### Test Database Connectivity Directly
```bash
# Using psql
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM \"user\""

# Check all tables
psql "$DATABASE_URL" -c "\dt"

# Check specific table
psql "$DATABASE_URL" -c "\d workspace_members"
```

### Test Configuration Loading
```bash
# Go can read .env
export $(cat .env | grep -v '^#' | xargs) && \
  echo "DATABASE_URL=$DATABASE_URL"
```

### Check Embedding Service
```bash
# Test Ollama
curl http://localhost:11434/api/tags

# If returns models, Ollama is running
```

---

## Integration Test Checklist

Use this checklist before deployment:

- [ ] Database connectivity verified
- [ ] All core tables exist and accessible
- [ ] Configuration loads without errors
- [ ] User and workspace data available
- [ ] Embedding service responds (optional)
- [ ] Connection pool statistics normal
- [ ] No errors in test output

---

## Performance Metrics

Target execution times:
- Database connectivity test: < 2 seconds
- Voice pipeline test: < 5 seconds
- Full suite: < 10 seconds

If slower, check:
- Network latency to database
- Database query performance
- Local CPU/memory availability

---

## Getting Help

If tests fail:

1. Check `.env` configuration
2. Verify PostgreSQL is running
3. Review error message carefully
4. Check database connectivity directly with psql
5. Review logs in `INTEGRATION_TEST_RESULTS.md`

For issues:
- Check database migrations are applied
- Ensure schema matches expected structure
- Verify all tables exist
- Check user permissions on database

---

**Last Updated:** 2026-01-18
**Status:** All tests passing
**Ready for:** Phase 5 - Voice Server Implementation
