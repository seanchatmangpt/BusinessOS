-- Migration: 057_voice_sessions
-- Create tables for voice session management and analytics

-- Voice Sessions Table
-- Stores metadata about voice conversation sessions
CREATE TABLE IF NOT EXISTS voice_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(255) NOT NULL UNIQUE, -- External session ID from client
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    agent_role VARCHAR(100) DEFAULT 'assistant',

    -- Session state
    state VARCHAR(50) DEFAULT 'active', -- active, idle, ended
    last_activity_at TIMESTAMPTZ DEFAULT NOW(),

    -- Session metadata
    total_messages INTEGER DEFAULT 0,
    total_duration_seconds INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

-- Voice Session Events Table
-- Stores detailed events for analytics and debugging
CREATE TABLE IF NOT EXISTS voice_session_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES voice_sessions(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL, -- session_start, utterance_processed, agent_response, session_end, error
    event_data JSONB DEFAULT '{}'::jsonb, -- Flexible storage for event-specific data

    -- Event metadata
    duration_ms INTEGER, -- Event processing duration
    error_message TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- User Facts Table (for personalization)
-- Stores user preferences and learned facts from conversations
CREATE TABLE IF NOT EXISTS user_facts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(255) NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE SET NULL,

    -- Fact content
    fact_type VARCHAR(100) NOT NULL, -- preference, expertise, context, habit, goal
    fact_key VARCHAR(255) NOT NULL, -- e.g., "preferred_communication_style", "timezone"
    fact_value TEXT NOT NULL,
    confidence_score FLOAT DEFAULT 1.0, -- 0.0 to 1.0

    -- Source tracking
    source VARCHAR(100) DEFAULT 'manual', -- manual, conversation, onboarding, email_analysis
    learned_at TIMESTAMPTZ DEFAULT NOW(),
    last_confirmed_at TIMESTAMPTZ,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Constraints
    UNIQUE(user_id, workspace_id, fact_key)
);

-- Indexes for performance
CREATE INDEX idx_voice_sessions_user ON voice_sessions(user_id);
CREATE INDEX idx_voice_sessions_workspace ON voice_sessions(workspace_id);
CREATE INDEX idx_voice_sessions_state ON voice_sessions(state);
CREATE INDEX idx_voice_sessions_created ON voice_sessions(created_at DESC);
CREATE INDEX idx_voice_sessions_session_id ON voice_sessions(session_id);

CREATE INDEX idx_voice_session_events_session ON voice_session_events(session_id);
CREATE INDEX idx_voice_session_events_type ON voice_session_events(event_type);
CREATE INDEX idx_voice_session_events_created ON voice_session_events(created_at DESC);

CREATE INDEX idx_user_facts_user ON user_facts(user_id);
CREATE INDEX idx_user_facts_workspace ON user_facts(workspace_id);
CREATE INDEX idx_user_facts_type ON user_facts(fact_type);
CREATE INDEX idx_user_facts_key ON user_facts(fact_key);

-- Triggers for updated_at
CREATE OR REPLACE FUNCTION update_voice_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER voice_sessions_updated_at
    BEFORE UPDATE ON voice_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_voice_sessions_updated_at();

CREATE OR REPLACE FUNCTION update_user_facts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_facts_updated_at
    BEFORE UPDATE ON user_facts
    FOR EACH ROW
    EXECUTE FUNCTION update_user_facts_updated_at();

-- Comments for documentation
COMMENT ON TABLE voice_sessions IS 'Stores metadata about voice conversation sessions for analytics and context';
COMMENT ON TABLE voice_session_events IS 'Detailed event log for voice sessions (analytics, debugging, replay)';
COMMENT ON TABLE user_facts IS 'Learned facts about users for personalization (preferences, expertise, habits)';

COMMENT ON COLUMN voice_sessions.session_id IS 'External session ID provided by client (e.g., UUID from frontend)';
COMMENT ON COLUMN voice_session_events.event_data IS 'Flexible JSONB for event-specific data (transcript, response, error details)';
COMMENT ON COLUMN user_facts.confidence_score IS 'Confidence in fact accuracy (0.0-1.0), decreases if contradicted';
COMMENT ON COLUMN user_facts.source IS 'How this fact was learned (manual, conversation, onboarding, email_analysis)';
