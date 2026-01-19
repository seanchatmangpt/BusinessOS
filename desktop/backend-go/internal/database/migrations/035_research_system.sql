-- Migration 035: Deep Research Agent System
-- Description: Tables for autonomous research agent (planner, executor, aggregator, writer)
-- Author: Claude Code
-- Date: 2026-01-19

-- Enable pgvector extension if not already enabled (for source deduplication)
CREATE EXTENSION IF NOT EXISTS vector;

-- Research Tasks (top-level tracking)
CREATE TABLE IF NOT EXISTS research_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    workspace_id UUID NOT NULL,
    conversation_id UUID,

    -- Query
    query TEXT NOT NULL,

    -- Status tracking
    status TEXT NOT NULL DEFAULT 'pending',
    -- Statuses: pending, planning, searching, aggregating, writing, completed, failed

    -- Results summary
    total_sources INT DEFAULT 0,
    cited_sources INT DEFAULT 0,
    word_count INT DEFAULT 0,

    -- Report content (may be null if failed)
    report_content TEXT,
    report_format TEXT DEFAULT 'markdown',

    -- Performance & cost metrics
    duration_ms INT,
    llm_tokens_used INT,
    search_api_calls INT,
    search_cost_usd NUMERIC(10,4),

    -- Quality metrics
    quality_score NUMERIC(3,2), -- 0.00-1.00
    source_diversity_score INT, -- unique domains count

    -- Error tracking
    error_message TEXT,
    error_phase TEXT, -- which phase failed: planning, searching, etc.

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,

    -- Foreign keys
    CONSTRAINT fk_research_workspace FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
    CONSTRAINT fk_research_conversation FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE SET NULL
);

-- Indexes for research_tasks
CREATE INDEX idx_research_tasks_user_created ON research_tasks(user_id, created_at DESC);
CREATE INDEX idx_research_tasks_workspace_created ON research_tasks(workspace_id, created_at DESC);
CREATE INDEX idx_research_tasks_status ON research_tasks(status, created_at DESC);
CREATE INDEX idx_research_tasks_conversation ON research_tasks(conversation_id) WHERE conversation_id IS NOT NULL;

-- Research Queries (sub-questions from planning phase)
CREATE TABLE IF NOT EXISTS research_queries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL,

    -- Query details
    question TEXT NOT NULL,
    search_type TEXT NOT NULL, -- web, rag, memory, hybrid
    weight NUMERIC(3,2) DEFAULT 0.5, -- importance 0.00-1.00
    order_num INT NOT NULL, -- execution order

    -- Dependencies (which queries must complete first)
    depends_on UUID[], -- array of query IDs

    -- Results tracking
    results_count INT DEFAULT 0,
    completed BOOLEAN DEFAULT FALSE,
    duration_ms INT,

    -- Error tracking
    error_message TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,

    -- Foreign keys
    CONSTRAINT fk_research_query_task FOREIGN KEY (task_id) REFERENCES research_tasks(id) ON DELETE CASCADE
);

-- Indexes for research_queries
CREATE INDEX idx_research_queries_task_order ON research_queries(task_id, order_num);
CREATE INDEX idx_research_queries_completed ON research_queries(task_id, completed);

-- Research Sources (results from search execution)
CREATE TABLE IF NOT EXISTS research_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL,
    query_id UUID, -- which sub-question found this

    -- Source metadata
    source_type TEXT NOT NULL, -- web, rag, memory
    url TEXT, -- may be null for memory/rag sources
    domain TEXT, -- extracted domain for diversity calculation
    title TEXT NOT NULL,
    content TEXT, -- full content if available
    snippet TEXT, -- preview snippet

    -- Ranking & relevance
    relevance_score NUMERIC(5,4) NOT NULL, -- 0.0000-1.0000
    final_rank INT, -- after RRF aggregation
    cited BOOLEAN DEFAULT FALSE, -- included in final report?

    -- Deduplication
    embedding VECTOR(1536), -- for similarity-based deduplication
    content_hash TEXT, -- for exact match deduplication

    -- Metadata
    author TEXT,
    published_at TIMESTAMPTZ,
    fetched_at TIMESTAMPTZ DEFAULT NOW(),

    -- Foreign keys
    CONSTRAINT fk_research_source_task FOREIGN KEY (task_id) REFERENCES research_tasks(id) ON DELETE CASCADE,
    CONSTRAINT fk_research_source_query FOREIGN KEY (query_id) REFERENCES research_queries(id) ON DELETE SET NULL
);

-- Indexes for research_sources
CREATE INDEX idx_research_sources_task_rank ON research_sources(task_id, final_rank) WHERE final_rank IS NOT NULL;
CREATE INDEX idx_research_sources_cited ON research_sources(task_id, cited) WHERE cited = TRUE;
CREATE INDEX idx_research_sources_type ON research_sources(task_id, source_type);
CREATE INDEX idx_research_sources_domain ON research_sources(task_id, domain);
CREATE INDEX idx_research_sources_content_hash ON research_sources(content_hash) WHERE content_hash IS NOT NULL;

-- Vector similarity index for deduplication (using ivfflat)
-- This will be created after embeddings are added (lazy index creation)
-- CREATE INDEX idx_research_sources_embedding ON research_sources USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Research Reports (final output storage)
CREATE TABLE IF NOT EXISTS research_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL,

    -- Report content
    content TEXT NOT NULL,
    format TEXT NOT NULL DEFAULT 'markdown', -- markdown, html, pdf

    -- Citations (stored as JSONB for flexibility)
    citations JSONB, -- [{source_id, citation_text, inline_refs: [1,2,3]}]

    -- Structure (for navigation)
    sections JSONB, -- [{title, content, source_ids: [uuid, uuid]}]

    -- Metrics
    word_count INT,
    citation_count INT,
    section_count INT,

    -- Quality indicators
    quality_score NUMERIC(3,2), -- 0.00-1.00

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Foreign keys
    CONSTRAINT fk_research_report_task FOREIGN KEY (task_id) REFERENCES research_tasks(id) ON DELETE CASCADE,

    -- Ensure one report per task
    CONSTRAINT unique_task_report UNIQUE (task_id)
);

-- Indexes for research_reports
CREATE INDEX idx_research_reports_task ON research_reports(task_id);
CREATE INDEX idx_research_reports_created ON research_reports(created_at DESC);

-- GIN index on citations JSONB for fast citation lookups
CREATE INDEX idx_research_reports_citations ON research_reports USING gin(citations);

-- Comments for documentation
COMMENT ON TABLE research_tasks IS 'Top-level tracking for research agent tasks';
COMMENT ON TABLE research_queries IS 'Sub-questions generated during planning phase';
COMMENT ON TABLE research_sources IS 'Individual sources found during execution phase';
COMMENT ON TABLE research_reports IS 'Final research reports with citations';

COMMENT ON COLUMN research_tasks.status IS 'Workflow status: pending, planning, searching, aggregating, writing, completed, failed';
COMMENT ON COLUMN research_tasks.quality_score IS 'Overall quality score (0.00-1.00) based on source diversity, recency, relevance';
COMMENT ON COLUMN research_queries.search_type IS 'Search strategy: web (external), rag (internal docs), memory (workspace), hybrid (all)';
COMMENT ON COLUMN research_queries.depends_on IS 'Array of query IDs that must complete before this one (for sequential dependencies)';
COMMENT ON COLUMN research_sources.source_type IS 'Origin of source: web (search API), rag (document search), memory (workspace knowledge)';
COMMENT ON COLUMN research_sources.embedding IS 'Vector embedding for similarity-based deduplication (cosine distance)';
COMMENT ON COLUMN research_reports.citations IS 'JSONB array of citations with format: [{source_id, citation_text, inline_refs}]';
COMMENT ON COLUMN research_reports.sections IS 'JSONB array of report sections: [{title, content, source_ids}]';
