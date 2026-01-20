-- Create workspace for user who completed onboarding but has no workspace
-- This fixes the issue where onboarding was marked complete but workspace wasn't created

-- First, get the user ID (replace with actual user email if different)
DO $$
DECLARE
    v_user_id TEXT;
    v_workspace_id UUID;
BEGIN
    -- Get user ID from email (adjust if needed)
    SELECT id INTO v_user_id
    FROM auth.users
    WHERE email = 'rhuacujaluna@gmail.com'  -- Change this to your actual email
    LIMIT 1;

    IF v_user_id IS NULL THEN
        RAISE EXCEPTION 'User not found';
    END IF;

    RAISE NOTICE 'Found user ID: %', v_user_id;

    -- Create workspace
    INSERT INTO workspaces (
        name,
        slug,
        owner_id,
        onboarding_completed_at,
        created_at,
        updated_at
    )
    VALUES (
        'Roberto Workspace',
        'roberto-workspace',
        v_user_id,
        NOW(),
        NOW(),
        NOW()
    )
    RETURNING id INTO v_workspace_id;

    RAISE NOTICE 'Created workspace ID: %', v_workspace_id;

    -- Create default workspace roles
    INSERT INTO workspace_roles (workspace_id, name, display_name, is_system, hierarchy_level, permissions, is_default)
    VALUES
        (v_workspace_id, 'owner', 'Owner', true, 0, '{"workspace": {"manage": true, "delete": true}, "members": {"view": true, "invite": true, "manage": true, "remove": true}, "roles": {"view": true, "manage": true}, "projects": {"create": true, "manage": true, "delete": true}}'::jsonb, false),
        (v_workspace_id, 'admin', 'Admin', true, 10, '{"members": {"view": true, "invite": true}, "projects": {"create": true, "manage": true}}'::jsonb, false),
        (v_workspace_id, 'member', 'Member', true, 30, '{"projects": {"view": true}}'::jsonb, true)
    ON CONFLICT DO NOTHING;

    RAISE NOTICE 'Created workspace roles';

    -- Add user as owner/member
    INSERT INTO workspace_members (
        workspace_id,
        user_id,
        role_name,
        status,
        joined_at
    )
    VALUES (
        v_workspace_id,
        v_user_id,
        'owner',
        'active',
        NOW()
    );

    RAISE NOTICE 'Added user as workspace owner';
    RAISE NOTICE 'SUCCESS! Workspace created: %', v_workspace_id;

END $$;
