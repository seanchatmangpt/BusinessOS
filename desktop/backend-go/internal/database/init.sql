-- BusinessOS Database Initialization Script
-- Run this on a fresh Cloud SQL database to set up all tables
-- Usage: psql -h <host> -U postgres -d business_os -f init.sql

-- ===== BETTER AUTH TABLES =====
-- These tables are managed by Better Auth but need to exist in the database

CREATE TABLE IF NOT EXISTS "user" (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE NOT NULL,
    "emailVerified" BOOLEAN DEFAULT FALSE,
    image TEXT,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS session (
    id VARCHAR(255) PRIMARY KEY,
    "userId" VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    "expiresAt" TIMESTAMP WITH TIME ZONE NOT NULL,
    "ipAddress" VARCHAR(45),
    "userAgent" TEXT,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS account (
    id VARCHAR(255) PRIMARY KEY,
    "userId" VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    "accountId" VARCHAR(255) NOT NULL,
    "providerId" VARCHAR(255) NOT NULL,
    "accessToken" TEXT,
    "refreshToken" TEXT,
    "accessTokenExpiresAt" TIMESTAMP WITH TIME ZONE,
    "refreshTokenExpiresAt" TIMESTAMP WITH TIME ZONE,
    scope TEXT,
    password TEXT,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS verification (
    id VARCHAR(255) PRIMARY KEY,
    identifier VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL,
    "expiresAt" TIMESTAMP WITH TIME ZONE NOT NULL,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_session_user ON session("userId");
CREATE INDEX IF NOT EXISTS idx_session_token ON session(token);
CREATE INDEX IF NOT EXISTS idx_account_user ON account("userId");
CREATE INDEX IF NOT EXISTS idx_account_provider ON account("providerId", "accountId");

-- ===== ENUM TYPES =====

DO $$ BEGIN
    CREATE TYPE messagerole AS ENUM ('USER', 'ASSISTANT', 'SYSTEM', 'user', 'assistant', 'system');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE artifacttype AS ENUM ('CODE', 'DOCUMENT', 'MARKDOWN', 'REACT', 'HTML', 'SVG');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE contexttype AS ENUM ('PERSON', 'BUSINESS', 'PROJECT', 'CUSTOM', 'document', 'DOCUMENT');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE projectstatus AS ENUM ('ACTIVE', 'PAUSED', 'COMPLETED', 'ARCHIVED');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE projectpriority AS ENUM ('CRITICAL', 'HIGH', 'MEDIUM', 'LOW');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE nodetype AS ENUM ('BUSINESS', 'PROJECT', 'LEARNING', 'OPERATIONAL');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE nodehealth AS ENUM ('HEALTHY', 'NEEDS_ATTENTION', 'CRITICAL', 'NOT_STARTED');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE taskstatus AS ENUM ('todo', 'in_progress', 'done', 'cancelled');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE taskpriority AS ENUM ('critical', 'high', 'medium', 'low');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE memberstatus AS ENUM ('AVAILABLE', 'BUSY', 'OVERLOADED', 'OOO');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE clienttype AS ENUM ('company', 'individual');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE clientstatus AS ENUM ('lead', 'prospect', 'active', 'inactive', 'churned');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE interactiontype AS ENUM ('call', 'email', 'meeting', 'note');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE dealstage AS ENUM ('qualification', 'proposal', 'negotiation', 'closed_won', 'closed_lost');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
    CREATE TYPE meetingtype AS ENUM (
        'team', 'sales', 'onboarding', 'kickoff', 'implementation',
        'standup', 'retrospective', 'planning', 'review', 'one_on_one',
        'client', 'internal', 'external', 'other'
    );
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

-- ===== APPLICATION TABLES =====

-- Clients table (created first for FK references)
CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type clienttype DEFAULT 'company',
    email VARCHAR(255),
    phone VARCHAR(50),
    website VARCHAR(255),
    industry VARCHAR(100),
    company_size VARCHAR(50),
    address VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    zip_code VARCHAR(20),
    country VARCHAR(100),
    status clientstatus DEFAULT 'lead',
    source VARCHAR(100),
    assigned_to VARCHAR(255),
    lifetime_value NUMERIC(12, 2),
    tags JSONB DEFAULT '[]',
    custom_fields JSONB DEFAULT '{}',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_contacted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_clients_user_id ON clients(user_id);

-- Contexts table
CREATE TABLE IF NOT EXISTS contexts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    type contexttype DEFAULT 'CUSTOM',
    content TEXT,
    structured_data JSONB,
    system_prompt_template TEXT,
    blocks JSONB DEFAULT '[]',
    cover_image VARCHAR(500),
    icon VARCHAR(50),
    parent_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    is_template BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    last_edited_at TIMESTAMP,
    word_count INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT FALSE,
    share_id VARCHAR(32) UNIQUE,
    property_schema JSONB DEFAULT '[]',
    properties JSONB DEFAULT '{}',
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contexts_user_id ON contexts(user_id);

-- Conversations table
CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) DEFAULT 'New Conversation',
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id);

-- Messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    role messagerole NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    message_metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status projectstatus DEFAULT 'ACTIVE',
    priority projectpriority DEFAULT 'MEDIUM',
    client_name VARCHAR(255),
    project_type VARCHAR(100) DEFAULT 'internal',
    project_metadata JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);

-- Artifacts table
CREATE TABLE IF NOT EXISTS artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    type artifacttype NOT NULL,
    language VARCHAR(50),
    content TEXT NOT NULL,
    summary VARCHAR(500),
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_artifacts_user_id ON artifacts(user_id);

-- Team members table
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    status memberstatus DEFAULT 'AVAILABLE',
    capacity INTEGER DEFAULT 0,
    manager_id UUID REFERENCES team_members(id) ON DELETE SET NULL,
    skills JSONB,
    hourly_rate NUMERIC(10, 2),
    share_calendar BOOLEAN DEFAULT FALSE,
    calendar_user_id VARCHAR(255),
    joined_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);

-- Tasks table
CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    status taskstatus DEFAULT 'todo',
    priority taskpriority DEFAULT 'medium',
    due_date TIMESTAMP,
    completed_at TIMESTAMP,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    assignee_id UUID REFERENCES team_members(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);

-- Focus items table
CREATE TABLE IF NOT EXISTS focus_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    text VARCHAR(500) NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    focus_date TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_focus_items_user_id ON focus_items(user_id);

-- Daily logs table
CREATE TABLE IF NOT EXISTS daily_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    content TEXT NOT NULL,
    transcription_source VARCHAR(50),
    extracted_actions JSONB,
    extracted_patterns JSONB,
    energy_level INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, date)
);

CREATE INDEX IF NOT EXISTS idx_daily_logs_user_id ON daily_logs(user_id);

-- Nodes table
CREATE TABLE IF NOT EXISTS nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    type nodetype NOT NULL,
    health nodehealth DEFAULT 'NOT_STARTED',
    purpose TEXT,
    current_status TEXT,
    this_week_focus JSONB,
    decision_queue JSONB,
    delegation_ready JSONB,
    is_active BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_nodes_user_id ON nodes(user_id);

-- User settings table
CREATE TABLE IF NOT EXISTS user_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    default_model VARCHAR(100),
    email_notifications BOOLEAN DEFAULT TRUE,
    daily_summary BOOLEAN DEFAULT FALSE,
    theme VARCHAR(20) DEFAULT 'light',
    sidebar_collapsed BOOLEAN DEFAULT FALSE,
    share_analytics BOOLEAN DEFAULT TRUE,
    custom_settings JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);

-- Client contacts
CREATE TABLE IF NOT EXISTS client_contacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    role VARCHAR(100),
    is_primary BOOLEAN DEFAULT FALSE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Client interactions
CREATE TABLE IF NOT EXISTS client_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    contact_id UUID REFERENCES client_contacts(id) ON DELETE SET NULL,
    type interactiontype NOT NULL,
    subject VARCHAR(255) NOT NULL,
    description TEXT,
    outcome VARCHAR(255),
    occurred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- DROPPED in migration 099: client_deals

-- Google OAuth tokens for calendar integration
CREATE TABLE IF NOT EXISTS google_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expiry TIMESTAMP WITH TIME ZONE NOT NULL,
    scopes TEXT[],
    google_email VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Slack OAuth tokens for workspace integration
CREATE TABLE IF NOT EXISTS slack_oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) UNIQUE NOT NULL,
    workspace_id VARCHAR(255) NOT NULL,
    workspace_name VARCHAR(255),
    bot_token TEXT NOT NULL,
    user_token TEXT,
    bot_user_id VARCHAR(255),
    authed_user_id VARCHAR(255),
    bot_scopes TEXT[],
    user_scopes TEXT[],
    incoming_webhook_url TEXT,
    incoming_webhook_channel VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_slack_oauth_user_id ON slack_oauth_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_slack_oauth_workspace ON slack_oauth_tokens(workspace_id);

-- Calendar events
CREATE TABLE IF NOT EXISTS calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    google_event_id VARCHAR(255),
    calendar_id VARCHAR(255) DEFAULT 'primary',
    title VARCHAR(500),
    description TEXT,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    all_day BOOLEAN DEFAULT FALSE,
    location VARCHAR(500),
    attendees JSONB DEFAULT '[]',
    status VARCHAR(50) DEFAULT 'confirmed',
    visibility VARCHAR(50) DEFAULT 'default',
    html_link TEXT,
    source VARCHAR(50) DEFAULT 'google',
    meeting_type meetingtype DEFAULT 'other',
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    client_id UUID REFERENCES clients(id) ON DELETE SET NULL,
    recording_url TEXT,
    meeting_link TEXT,
    external_links JSONB DEFAULT '[]',
    meeting_notes TEXT,
    action_items JSONB DEFAULT '[]',
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, google_event_id)
);

CREATE INDEX IF NOT EXISTS idx_calendar_events_user_id ON calendar_events(user_id);

-- AI usage logs
CREATE TABLE IF NOT EXISTS ai_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    provider VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    agent_name VARCHAR(100),
    delegated_to VARCHAR(100),
    parent_request_id UUID,
    request_type VARCHAR(50),
    context_ids UUID[],
    node_id UUID REFERENCES nodes(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    duration_ms INTEGER DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    estimated_cost NUMERIC(10, 6),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_usage_user_id ON ai_usage_logs(user_id);

-- MCP usage logs
CREATE TABLE IF NOT EXISTS mcp_usage_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    tool_name VARCHAR(255) NOT NULL,
    server_name VARCHAR(255),
    input_params JSONB,
    output_result JSONB,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    duration_ms INTEGER DEFAULT 0,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    ai_request_id UUID REFERENCES ai_usage_logs(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_mcp_usage_user_id ON mcp_usage_logs(user_id);

-- Usage daily summary
CREATE TABLE IF NOT EXISTS usage_daily_summary (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    ai_requests INTEGER DEFAULT 0,
    ai_input_tokens INTEGER DEFAULT 0,
    ai_output_tokens INTEGER DEFAULT 0,
    ai_total_tokens INTEGER DEFAULT 0,
    ai_estimated_cost NUMERIC(10, 4) DEFAULT 0,
    provider_breakdown JSONB DEFAULT '{}',
    model_breakdown JSONB DEFAULT '{}',
    agent_breakdown JSONB DEFAULT '{}',
    mcp_requests INTEGER DEFAULT 0,
    mcp_tool_breakdown JSONB DEFAULT '{}',
    conversations_created INTEGER DEFAULT 0,
    messages_sent INTEGER DEFAULT 0,
    artifacts_created INTEGER DEFAULT 0,
    documents_created INTEGER DEFAULT 0,
    contexts_accessed UUID[],
    nodes_accessed UUID[],
    projects_accessed UUID[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, date)
);

-- User custom commands
CREATE TABLE IF NOT EXISTS user_commands (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(10),
    system_prompt TEXT NOT NULL,
    context_sources TEXT[] DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_user_commands_user_id ON user_commands(user_id);

-- Voice notes
CREATE TABLE IF NOT EXISTS voice_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    transcript TEXT NOT NULL,
    duration_seconds INTEGER NOT NULL,
    word_count INTEGER NOT NULL,
    words_per_minute NUMERIC(10, 2),
    language VARCHAR(10) DEFAULT 'en',
    audio_file_path VARCHAR(500),
    context_id UUID REFERENCES contexts(id) ON DELETE SET NULL,
    project_id UUID REFERENCES projects(id) ON DELETE SET NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_voice_notes_user_id ON voice_notes(user_id);

-- Additional tables (project notes, conversation tags, etc.)
CREATE TABLE IF NOT EXISTS project_notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- DROPPED in migration 098: conversation_tags, project_conversations

CREATE TABLE IF NOT EXISTS artifact_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    artifact_id UUID NOT NULL REFERENCES artifacts(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- DROPPED in migration 098: node_metrics

CREATE TABLE IF NOT EXISTS team_member_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id UUID NOT NULL REFERENCES team_members(id) ON DELETE CASCADE,
    activity_type VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS system_event_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    event_data JSONB,
    module VARCHAR(100),
    resource_type VARCHAR(100),
    resource_id UUID,
    session_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_system_events_user_id ON system_event_logs(user_id);

-- ===== USAGE METERING =====

-- Usage metering tables
CREATE TABLE IF NOT EXISTS usage_meters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id),
    meter_type VARCHAR(50) NOT NULL, -- 'ai_calls', 'storage_bytes', 'compute_cpu_seconds', 'compute_memory_gb_seconds'
    quantity BIGINT NOT NULL DEFAULT 0,
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unique constraint enables the upsert in RecordAICall and similar helpers.
CREATE UNIQUE INDEX IF NOT EXISTS idx_usage_meters_upsert ON usage_meters(workspace_id, meter_type, period_start);
CREATE INDEX IF NOT EXISTS idx_usage_meters_workspace_period ON usage_meters(workspace_id, period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_usage_meters_type ON usage_meters(meter_type);

-- Plan limits table
CREATE TABLE IF NOT EXISTS plan_limits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_name VARCHAR(50) NOT NULL UNIQUE, -- 'free', 'pro', 'enterprise'
    ai_calls_per_day INT NOT NULL DEFAULT 100,
    ai_model_tier VARCHAR(20) NOT NULL DEFAULT 'haiku', -- 'haiku', 'sonnet', 'opus'
    max_modules INT NOT NULL DEFAULT 3,
    storage_bytes_limit BIGINT NOT NULL DEFAULT 524288000, -- 500MB
    max_team_members INT NOT NULL DEFAULT 1,
    osa_modes TEXT[] NOT NULL DEFAULT '{chat}',
    compute_cpu_hours_per_month DECIMAL NOT NULL DEFAULT 10,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default plans
INSERT INTO plan_limits (plan_name, ai_calls_per_day, ai_model_tier, max_modules, storage_bytes_limit, max_team_members, osa_modes, compute_cpu_hours_per_month)
VALUES
    ('free',       100,  'haiku',  3,  524288000,    1,  '{chat}',                          10),
    ('pro',        5000, 'sonnet', 25, 10737418240,  5,  '{chat,build,focus,research,plan}', 100),
    ('enterprise', -1,   'opus',   -1, 107374182400, -1, '{chat,build,focus,research,plan}', -1)
ON CONFLICT (plan_name) DO NOTHING;

-- Workspace plan assignment
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS plan_name VARCHAR(50) NOT NULL DEFAULT 'free';
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS stripe_customer_id VARCHAR(255);
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(255);

-- Done!
SELECT 'BusinessOS database initialized successfully!' as status;

-- Migration 059: Background Jobs System
-- Migration 036: Background Jobs System
-- Description: Implements reliable task queue for async operations with retry logic and scheduling

-- ============================================================================
-- BACKGROUND JOBS TABLE
-- ============================================================================
-- Stores individual job execution records with retry logic and worker locking

CREATE TABLE IF NOT EXISTS background_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Job Definition
    job_type VARCHAR(100) NOT NULL,           -- 'email_send', 'report_generate', 'sync_calendar', etc.
    payload JSONB NOT NULL,                   -- Job-specific parameters

    -- Scheduling
    scheduled_at TIMESTAMPTZ DEFAULT NOW(),   -- When the job should run
    priority INTEGER DEFAULT 0,               -- Higher number = higher priority

    -- Execution Status
    status VARCHAR(50) DEFAULT 'pending',     -- 'pending', 'running', 'completed', 'failed', 'cancelled'
    started_at TIMESTAMPTZ,                   -- When execution began
    completed_at TIMESTAMPTZ,                 -- When execution finished

    -- Worker Management
    worker_id VARCHAR(100),                   -- ID of worker processing this job
    locked_until TIMESTAMPTZ,                 -- Lock expiry time (prevents duplicate processing)

    -- Retry Logic
    attempt_count INTEGER DEFAULT 0,          -- Number of attempts made
    max_attempts INTEGER DEFAULT 3,           -- Maximum retry attempts
    last_error TEXT,                          -- Error from last failed attempt

    -- Result Storage
    result JSONB,                             -- Job execution result

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for efficient job polling by workers
CREATE INDEX IF NOT EXISTS idx_background_jobs_status
ON background_jobs(status, scheduled_at, priority DESC);

-- Index for filtering by job type
CREATE INDEX IF NOT EXISTS idx_background_jobs_type
ON background_jobs(job_type);

-- Index for worker queries
CREATE INDEX IF NOT EXISTS idx_background_jobs_worker
ON background_jobs(worker_id) WHERE worker_id IS NOT NULL;

-- Index for cleanup queries
CREATE INDEX IF NOT EXISTS idx_background_jobs_created
ON background_jobs(created_at) WHERE status IN ('completed', 'failed', 'cancelled');

-- ============================================================================
-- SCHEDULED JOBS TABLE
-- ============================================================================
-- Stores recurring job definitions with cron-like scheduling

CREATE TABLE IF NOT EXISTS scheduled_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Job Definition
    job_type VARCHAR(100) NOT NULL,           -- Type of job to create
    payload JSONB NOT NULL,                   -- Default payload for created jobs

    -- Schedule Configuration
    cron_expression VARCHAR(100) NOT NULL,    -- "0 9 * * 1-5" (9am weekdays), "*/5 * * * *" (every 5 min)
    timezone VARCHAR(50) DEFAULT 'UTC',       -- Timezone for cron calculation

    -- Status & Execution Tracking
    is_active BOOLEAN DEFAULT TRUE,           -- Whether this scheduled job is enabled
    last_run_at TIMESTAMPTZ,                  -- When it last created a background_job
    next_run_at TIMESTAMPTZ,                  -- When it should next run (calculated from cron)

    -- Metadata
    name VARCHAR(255),                        -- Optional friendly name
    description TEXT,                         -- Optional description
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for scheduler polling
CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_next_run
ON scheduled_jobs(next_run_at) WHERE is_active = TRUE;

-- Index for filtering by job type
CREATE INDEX IF NOT EXISTS idx_scheduled_jobs_type
ON scheduled_jobs(job_type);

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Function to acquire next available job (with atomic locking)
CREATE OR REPLACE FUNCTION acquire_background_job(
    p_worker_id VARCHAR(100),
    p_lock_duration_seconds INTEGER DEFAULT 300  -- 5 minutes default lock
)
RETURNS TABLE (
    job_id UUID,
    job_type VARCHAR(100),
    payload JSONB,
    attempt_count INTEGER,
    max_attempts INTEGER
) AS $$
DECLARE
    v_job_id UUID;
BEGIN
    -- Find and lock the next available job atomically
    SELECT id INTO v_job_id
    FROM background_jobs
    WHERE status = 'pending'
      AND scheduled_at <= NOW()
      AND (locked_until IS NULL OR locked_until < NOW())
      AND attempt_count < max_attempts
    ORDER BY priority DESC, scheduled_at ASC
    LIMIT 1
    FOR UPDATE SKIP LOCKED;

    -- If no job found, return empty
    IF v_job_id IS NULL THEN
        RETURN;
    END IF;

    -- Update job with lock and status
    UPDATE background_jobs
    SET
        status = 'running',
        worker_id = p_worker_id,
        locked_until = NOW() + (p_lock_duration_seconds || ' seconds')::INTERVAL,
        started_at = CASE WHEN started_at IS NULL THEN NOW() ELSE started_at END,
        attempt_count = attempt_count + 1
    WHERE id = v_job_id;

    -- Return job details
    RETURN QUERY
    SELECT
        b.id,
        b.job_type,
        b.payload,
        b.attempt_count,
        b.max_attempts
    FROM background_jobs b
    WHERE b.id = v_job_id;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate next retry time with exponential backoff
CREATE OR REPLACE FUNCTION calculate_retry_time(
    p_attempt_count INTEGER
)
RETURNS TIMESTAMPTZ AS $$
BEGIN
    -- Exponential backoff: 1min, 5min, 15min, 15min, ...
    RETURN CASE
        WHEN p_attempt_count <= 1 THEN NOW() + INTERVAL '1 minute'
        WHEN p_attempt_count = 2 THEN NOW() + INTERVAL '5 minutes'
        ELSE NOW() + INTERVAL '15 minutes'
    END;
END;
$$ LANGUAGE plpgsql;

-- Function to release stuck jobs (for cleanup)
CREATE OR REPLACE FUNCTION release_stuck_jobs()
RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER;
BEGIN
    -- Release jobs that have been locked too long
    UPDATE background_jobs
    SET
        status = 'pending',
        locked_until = NULL,
        worker_id = NULL
    WHERE status = 'running'
      AND locked_until < NOW();

    GET DIAGNOSTICS v_count = ROW_COUNT;
    RETURN v_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE background_jobs IS 'Stores background job execution records with retry logic and worker locking';
COMMENT ON TABLE scheduled_jobs IS 'Stores recurring job definitions with cron-like scheduling';
COMMENT ON FUNCTION acquire_background_job IS 'Atomically acquires next available job for a worker';
COMMENT ON FUNCTION calculate_retry_time IS 'Calculates next retry time with exponential backoff';
COMMENT ON FUNCTION release_stuck_jobs IS 'Releases jobs stuck in running state (for cleanup)';
