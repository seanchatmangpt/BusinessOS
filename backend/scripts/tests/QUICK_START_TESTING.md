# Quick Start: OSA Worker Testing

## 1-Minute Database Setup

```sql
-- Copy-paste this entire block into your PostgreSQL client

-- Step 1: Create test workspace
INSERT INTO workspaces (name, slug, owner_id, description)
VALUES (
    'OSA Worker Test',
    'osa-worker-test',
    'system-test',
    'Created for OSA queue worker testing'
)
ON CONFLICT (slug) DO UPDATE SET updated_at = NOW()
RETURNING id;

-- Copy the returned workspace ID, then replace <WORKSPACE_ID> below

-- Step 2: Insert test templates
INSERT INTO app_templates (
    template_name, category, display_name, description,
    icon_type, priority_score, generation_prompt, scaffold_type
) VALUES
(
    'test_crm_basic',
    'crm',
    'Test CRM System',
    'A basic CRM system for testing OSA worker functionality',
    'users',
    80,
    'Generate a CRM application with contact management, deal tracking, and task management.',
    'full-stack'
),
(
    'test_todo_app',
    'productivity',
    'Test Todo Application',
    'A simple todo app for testing worker',
    'check-square',
    60,
    'Generate a todo list application with task creation and completion tracking.',
    'frontend-focused'
)
ON CONFLICT (template_name) DO UPDATE SET updated_at = NOW();

-- Step 3: Insert test queue items (replace <WORKSPACE_ID>)
INSERT INTO app_generation_queue (
    workspace_id,
    template_id,
    status,
    priority,
    generation_context
)
SELECT
    '<WORKSPACE_ID>'::uuid,  -- REPLACE THIS
    t.id,
    'pending',
    8,
    '{"app_name": "Test CRM from Worker", "description": "Testing OSA queue worker"}'::jsonb
FROM app_templates t
WHERE t.template_name = 'test_crm_basic'
RETURNING id, workspace_id, status, created_at;

-- Step 4: Verify queue status
SELECT
    q.id,
    q.status,
    q.priority,
    q.generation_context->>'app_name' as app_name,
    t.template_name,
    q.created_at
FROM app_generation_queue q
LEFT JOIN app_templates t ON q.template_id = t.id
ORDER BY q.created_at DESC
LIMIT 5;
```

## Start Server with Worker

```bash
cd C:/Users/Pichau/Desktop/BusinessOS-main-dev/desktop/backend-go
./bin/server-with-worker
```

Look for this line in logs:
```
✅ OSA queue worker started (polling every 5s)
```

## Monitor Queue (Run Every 5 Seconds)

```sql
SELECT
    id,
    status,
    priority,
    generation_context->>'app_name' as app_name,
    retry_count,
    CASE
        WHEN started_at IS NOT NULL THEN
            'Started ' || EXTRACT(EPOCH FROM (NOW() - started_at))::int || 's ago'
        ELSE 'Waiting'
    END as status_info,
    error_message
FROM app_generation_queue
WHERE status IN ('pending', 'processing', 'completed', 'failed')
ORDER BY created_at DESC
LIMIT 10;
```

## Expected Behavior

**T+0s:** Queue item created
```
status = 'pending'
```

**T+5s:** Worker picks up item
```
status = 'processing'
started_at = NOW()
```

**T+15-60s:** Processing completes
```
status = 'completed' OR 'failed'
completed_at = NOW()
```

## Cleanup After Test

```sql
-- Delete test queue items
DELETE FROM app_generation_queue
WHERE template_id IN (
    SELECT id FROM app_templates WHERE template_name LIKE 'test_%'
);

-- Delete test templates
DELETE FROM app_templates WHERE template_name LIKE 'test_%';

-- Verify cleanup
SELECT COUNT(*) FROM app_generation_queue WHERE workspace_id IN (
    SELECT id FROM workspaces WHERE slug = 'osa-worker-test'
);
```

## Troubleshooting

### Issue: No workspace ID returned
**Solution:**
```sql
-- Get existing workspace
SELECT id FROM workspaces LIMIT 1;
-- Use this ID in Step 3
```

### Issue: Worker not starting
**Check:** Server logs for errors
**Verify:** Migration 089 applied
```sql
SELECT COUNT(*) FROM app_templates;  -- Should work
SELECT COUNT(*) FROM app_generation_queue;  -- Should work
```

### Issue: Items staying in "pending"
**Possible causes:**
1. Server not running
2. Worker crashed (check logs)
3. Database connection issue

**Debug query:**
```sql
SELECT * FROM pg_stat_activity
WHERE application_name LIKE '%businessos%';
```

## Alternative: Use Go Test Script

If you prefer automated testing:

```bash
# First, manually create a workspace (Step 1 above)
# Then run:
cd C:/Users/Pichau/Desktop/BusinessOS-main-dev/desktop/backend-go
go run scripts/tests/run_osa_worker_exec.go
```

This will:
1. ✅ Setup test templates
2. ✅ Find workspace automatically
3. ✅ Insert queue items
4. ✅ Monitor for 30 seconds
5. ✅ Show final results

---

**Estimated Time:** 5 minutes (2 min setup + 3 min observation)
