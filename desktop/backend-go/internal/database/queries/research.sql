-- Research System SQLC Queries
-- Generated: 2026-01-19

-- ============================================================================
-- RESEARCH TASKS
-- ============================================================================

-- name: CreateResearchTask :one
INSERT INTO research_tasks (
    user_id,
    workspace_id,
    conversation_id,
    query,
    status
) VALUES (
    $1, $2, $3, $4, 'pending'
) RETURNING *;

-- name: GetResearchTask :one
SELECT * FROM research_tasks
WHERE id = $1;

-- name: ListResearchTasksByUser :many
SELECT * FROM research_tasks
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListResearchTasksByWorkspace :many
SELECT * FROM research_tasks
WHERE workspace_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListResearchTasksByConversation :many
SELECT * FROM research_tasks
WHERE conversation_id = $1
ORDER BY created_at DESC;

-- name: UpdateResearchTaskStatus :exec
UPDATE research_tasks
SET status = $2,
    started_at = CASE WHEN $2 = 'planning' AND started_at IS NULL THEN NOW() ELSE started_at END,
    completed_at = CASE WHEN $2 IN ('completed', 'failed') THEN NOW() ELSE completed_at END
WHERE id = $1;

-- name: UpdateResearchTaskMetrics :exec
UPDATE research_tasks
SET duration_ms = $2,
    llm_tokens_used = $3,
    search_api_calls = $4,
    search_cost_usd = $5,
    total_sources = $6,
    cited_sources = $7,
    word_count = $8,
    quality_score = $9,
    source_diversity_score = $10
WHERE id = $1;

-- name: UpdateResearchTaskError :exec
UPDATE research_tasks
SET status = 'failed',
    error_message = $2,
    error_phase = $3,
    completed_at = NOW()
WHERE id = $1;

-- name: SetResearchTaskReport :exec
UPDATE research_tasks
SET report_content = $2,
    report_format = $3,
    status = 'completed',
    completed_at = NOW()
WHERE id = $1;

-- name: DeleteResearchTask :exec
DELETE FROM research_tasks
WHERE id = $1;

-- ============================================================================
-- RESEARCH QUERIES (Sub-questions)
-- ============================================================================

-- name: CreateResearchQuery :one
INSERT INTO research_queries (
    task_id,
    question,
    search_type,
    weight,
    order_num,
    depends_on
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetResearchQuery :one
SELECT * FROM research_queries
WHERE id = $1;

-- name: ListResearchQueriesByTask :many
SELECT * FROM research_queries
WHERE task_id = $1
ORDER BY order_num ASC;

-- name: ListPendingResearchQueries :many
SELECT * FROM research_queries
WHERE task_id = $1
  AND completed = FALSE
ORDER BY order_num ASC;

-- name: MarkResearchQueryCompleted :exec
UPDATE research_queries
SET completed = TRUE,
    results_count = $2,
    duration_ms = $3,
    completed_at = NOW()
WHERE id = $1;

-- name: MarkResearchQueryFailed :exec
UPDATE research_queries
SET completed = FALSE,
    error_message = $2
WHERE id = $1;

-- name: GetResearchQueryDependencies :many
SELECT * FROM research_queries
WHERE task_id = $1
  AND id = ANY($2::uuid[]);

-- ============================================================================
-- RESEARCH SOURCES
-- ============================================================================

-- name: CreateResearchSource :one
INSERT INTO research_sources (
    task_id,
    query_id,
    source_type,
    url,
    domain,
    title,
    content,
    snippet,
    relevance_score,
    author,
    published_at,
    content_hash
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetResearchSource :one
SELECT * FROM research_sources
WHERE id = $1;

-- name: ListResearchSourcesByTask :many
SELECT * FROM research_sources
WHERE task_id = $1
ORDER BY relevance_score DESC;

-- name: ListResearchSourcesByQuery :many
SELECT * FROM research_sources
WHERE query_id = $1
ORDER BY relevance_score DESC;

-- name: ListCitedResearchSources :many
SELECT * FROM research_sources
WHERE task_id = $1
  AND cited = TRUE
ORDER BY final_rank ASC;

-- name: ListTopResearchSources :many
SELECT * FROM research_sources
WHERE task_id = $1
  AND final_rank IS NOT NULL
ORDER BY final_rank ASC
LIMIT $2;

-- name: UpdateResearchSourceRank :exec
UPDATE research_sources
SET final_rank = $2
WHERE id = $1;

-- name: MarkResearchSourceCited :exec
UPDATE research_sources
SET cited = TRUE
WHERE id = $1;

-- name: UpdateResearchSourceEmbedding :exec
UPDATE research_sources
SET embedding = $2::vector
WHERE id = $1;

-- name: FindSimilarResearchSources :many
SELECT id, task_id, title, url,
       1 - (embedding <=> $2::vector) AS similarity
FROM research_sources
WHERE task_id = $1
  AND embedding IS NOT NULL
ORDER BY embedding <=> $2::vector
LIMIT $3;

-- name: CheckDuplicateSourceByHash :one
SELECT id, task_id, title, url
FROM research_sources
WHERE task_id = $1
  AND content_hash = $2
LIMIT 1;

-- name: GetSourceDiversityCount :one
SELECT COUNT(DISTINCT domain) AS unique_domains
FROM research_sources
WHERE task_id = $1
  AND cited = TRUE;

-- ============================================================================
-- RESEARCH REPORTS
-- ============================================================================

-- name: CreateResearchReport :one
INSERT INTO research_reports (
    task_id,
    content,
    format,
    citations,
    sections,
    word_count,
    citation_count,
    section_count,
    quality_score
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetResearchReport :one
SELECT * FROM research_reports
WHERE id = $1;

-- name: GetResearchReportByTask :one
SELECT * FROM research_reports
WHERE task_id = $1;

-- name: UpdateResearchReport :exec
UPDATE research_reports
SET content = $2,
    format = $3,
    citations = $4,
    sections = $5,
    word_count = $6,
    citation_count = $7,
    section_count = $8,
    quality_score = $9,
    updated_at = NOW()
WHERE task_id = $1;

-- name: ListRecentResearchReports :many
SELECT r.*, t.query, t.user_id, t.workspace_id
FROM research_reports r
JOIN research_tasks t ON r.task_id = t.id
WHERE t.user_id = $1
ORDER BY r.created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteResearchReport :exec
DELETE FROM research_reports
WHERE task_id = $1;

-- ============================================================================
-- ANALYTICS & METRICS
-- ============================================================================

-- name: GetResearchTaskStats :one
SELECT
    COUNT(*) AS total_tasks,
    COUNT(*) FILTER (WHERE status = 'completed') AS completed_tasks,
    COUNT(*) FILTER (WHERE status = 'failed') AS failed_tasks,
    AVG(duration_ms) FILTER (WHERE status = 'completed') AS avg_duration_ms,
    AVG(quality_score) FILTER (WHERE status = 'completed') AS avg_quality_score,
    SUM(llm_tokens_used) AS total_llm_tokens,
    SUM(search_cost_usd) AS total_search_cost
FROM research_tasks
WHERE user_id = $1
  AND created_at > $2;

-- name: GetResearchTaskPerformance :many
SELECT
    DATE_TRUNC('day', created_at) AS day,
    COUNT(*) AS tasks_count,
    AVG(duration_ms) AS avg_duration_ms,
    AVG(quality_score) AS avg_quality,
    SUM(search_cost_usd) AS total_cost
FROM research_tasks
WHERE workspace_id = $1
  AND created_at > $2
  AND status = 'completed'
GROUP BY DATE_TRUNC('day', created_at)
ORDER BY day DESC;
