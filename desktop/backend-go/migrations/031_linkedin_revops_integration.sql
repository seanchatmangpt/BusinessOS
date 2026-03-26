-- Migration: LinkedIn RevOps Integration
-- Adds tables for LinkedIn contact management, outreach sequences, and message queuing
-- Date: 2026-03-26

-- LinkedIn contacts table
CREATE TABLE IF NOT EXISTS linkedin_contacts (
    id BIGSERIAL PRIMARY KEY,
    linkedin_id VARCHAR(255) UNIQUE,
    name VARCHAR(255) NOT NULL,
    title VARCHAR(255),
    company VARCHAR(255),
    industry VARCHAR(255),
    connection_date TIMESTAMP WITH TIME ZONE,
    icp_score DECIMAL(3, 2) DEFAULT 0.0,
    icp_scored_at TIMESTAMP WITH TIME ZONE,
    raw_csv TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_linkedin_contacts_icp_score ON linkedin_contacts(icp_score DESC);
CREATE INDEX IF NOT EXISTS idx_linkedin_contacts_created_at ON linkedin_contacts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_linkedin_contacts_company ON linkedin_contacts(company);

-- Outreach sequences (campaign templates)
CREATE TABLE IF NOT EXISTS outreach_sequences (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    target_icp_min_score DECIMAL(3, 2) DEFAULT 0.7,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outreach_sequences_name ON outreach_sequences(name);

-- Sequence steps (email/message templates in sequence)
CREATE TABLE IF NOT EXISTS sequence_steps (
    id BIGSERIAL PRIMARY KEY,
    sequence_id BIGINT NOT NULL REFERENCES outreach_sequences(id) ON DELETE CASCADE,
    step_order INT NOT NULL,
    message_template TEXT NOT NULL,
    delay_days INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT unique_sequence_step UNIQUE(sequence_id, step_order)
);

CREATE INDEX IF NOT EXISTS idx_sequence_steps_sequence_id ON sequence_steps(sequence_id);

-- Sequence enrollments (contact progress through sequence)
CREATE TABLE IF NOT EXISTS sequence_enrollments (
    id BIGSERIAL PRIMARY KEY,
    contact_id BIGINT NOT NULL REFERENCES linkedin_contacts(id) ON DELETE CASCADE,
    sequence_id BIGINT NOT NULL REFERENCES outreach_sequences(id) ON DELETE CASCADE,
    current_step INT DEFAULT 1,
    enrolled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sequence_enrollments_contact_id ON sequence_enrollments(contact_id);
CREATE INDEX IF NOT EXISTS idx_sequence_enrollments_sequence_id ON sequence_enrollments(sequence_id);
CREATE INDEX IF NOT EXISTS idx_sequence_enrollments_completed_at ON sequence_enrollments(completed_at);

-- LinkedIn message queue (rate-limited outreach)
CREATE TABLE IF NOT EXISTS linkedin_message_queue (
    id BIGSERIAL PRIMARY KEY,
    contact_id BIGINT NOT NULL REFERENCES linkedin_contacts(id) ON DELETE CASCADE,
    step_id BIGINT NOT NULL REFERENCES sequence_steps(id) ON DELETE CASCADE,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed', 'skipped')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_linkedin_message_queue_status ON linkedin_message_queue(status);
CREATE INDEX IF NOT EXISTS idx_linkedin_message_queue_scheduled_at ON linkedin_message_queue(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_linkedin_message_queue_contact_id ON linkedin_message_queue(contact_id);

-- Table comments
COMMENT ON TABLE linkedin_contacts IS 'LinkedIn contacts imported from CSV export';
COMMENT ON COLUMN linkedin_contacts.icp_score IS 'Ideal Customer Profile score (0.0-1.0): title 50% + industry 30% + company 20%';
COMMENT ON TABLE outreach_sequences IS 'Campaign templates for LinkedIn outreach (multi-step workflows)';
COMMENT ON TABLE sequence_steps IS 'Individual message steps within a sequence (e.g., step 1: connection request, step 2: intro message)';
COMMENT ON TABLE sequence_enrollments IS 'Tracks contact enrollment and progress through sequences';
COMMENT ON TABLE linkedin_message_queue IS 'Rate-limited message queue: max 5 messages per contact per day (Redis: linkedin:rate:{contact_id}:{date})';
