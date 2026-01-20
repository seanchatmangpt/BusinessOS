-- Create workspace for ALL users who don't have one
-- This fixes users who completed onboarding but have no workspace

DO $$
DECLARE
    user_record RECORD;
    v_workspace_id UUID;
    workspace_count INT := 0;
BEGIN
    -- Loop through all users who don't have a workspace
    FOR user_record IN (
        SELECT DISTINCT u.id::TEXT as id, u.email
        FROM auth.users u
        WHERE NOT EXISTS (
            SELECT 1 FROM workspaces w WHERE w.owner_id = u.id::TEXT
        )
    )
    LOOP
        RAISE NOTICE 'Creating workspace for user: % (%)', user_record.email, user_record.id;

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
            'My Workspace',  -- Generic name
            'workspace-' || REPLACE(CAST(user_record.id AS TEXT), '-', ''),  -- Unique slug
            user_record.id,
            NOW(),
            NOW(),
            NOW()
        )
        RETURNING id INTO v_workspace_id;

        RAISE NOTICE '  -> Created workspace ID: %', v_workspace_id;

        -- Create default workspace roles
        INSERT INTO workspace_roles (workspace_id, name, display_name, is_system, hierarchy_level, permissions, is_default)
        VALUES
            (v_workspace_id, 'owner', 'Owner', true, 0, '{"workspace": {"manage": true, "delete": true}, "members": {"view": true, "invite": true, "manage": true, "remove": true}, "roles": {"view": true, "manage": true}, "projects": {"create": true, "manage": true, "delete": true}}'::jsonb, false),
            (v_workspace_id, 'admin', 'Admin', true, 10, '{"members": {"view": true, "invite": true}, "projects": {"create": true, "manage": true}}'::jsonb, false),
            (v_workspace_id, 'member', 'Member', true, 30, '{"projects": {"view": true}}'::jsonb, true)
        ON CONFLICT DO NOTHING;

        -- Add user as owner/member
        INSERT INTO workspace_members (
            workspace_id,
            user_id,
            role,
            status,
            joined_at
        )
        VALUES (
            v_workspace_id,
            user_record.id,
            'owner',
            'active',
            NOW()
        );

        workspace_count := workspace_count + 1;
        RAISE NOTICE '  -> SUCCESS! User added as workspace owner';
    END LOOP;

    IF workspace_count = 0 THEN
        RAISE NOTICE 'No users found without workspaces';
    ELSE
        RAISE NOTICE 'Created % workspace(s)', workspace_count;
    END IF;
END $$;
