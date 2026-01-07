-- Migration 031: Fix Workspace Members
-- Ensures all workspace owners are registered as members
-- This is a safety migration to fix any workspaces created before workspace_members table

-- Add workspace owners as members if they're not already
INSERT INTO workspace_members (workspace_id, user_id, role, status, joined_at)
SELECT
    w.id as workspace_id,
    w.owner_id as user_id,
    'owner' as role,
    'active' as status,
    w.created_at as joined_at
FROM workspaces w
WHERE NOT EXISTS (
    SELECT 1 FROM workspace_members wm
    WHERE wm.workspace_id = w.id
    AND wm.user_id = w.owner_id
)
ON CONFLICT (workspace_id, user_id) DO UPDATE
SET
    role = EXCLUDED.role,
    status = EXCLUDED.status;

-- Log how many were fixed
DO $$
DECLARE
    fixed_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO fixed_count
    FROM workspaces w
    WHERE NOT EXISTS (
        SELECT 1 FROM workspace_members wm
        WHERE wm.workspace_id = w.id AND wm.user_id = w.owner_id
    );

    IF fixed_count > 0 THEN
        RAISE NOTICE 'Fixed % workspace(s) by adding owners as members', fixed_count;
    ELSE
        RAISE NOTICE 'All workspaces already have their owners as members';
    END IF;
END $$;
