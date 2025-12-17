-- name: ListNodes :many
SELECT * FROM nodes
WHERE user_id = $1
  AND is_archived = FALSE
ORDER BY sort_order ASC, name ASC;

-- name: GetNodeTree :many
SELECT * FROM nodes
WHERE user_id = $1
  AND is_archived = FALSE
ORDER BY parent_id NULLS FIRST, sort_order ASC, name ASC;

-- name: GetActiveNode :one
SELECT * FROM nodes
WHERE user_id = $1 AND is_active = TRUE
LIMIT 1;

-- name: GetNode :one
SELECT * FROM nodes
WHERE id = $1 AND user_id = $2;

-- name: CreateNode :one
INSERT INTO nodes (user_id, parent_id, context_id, name, type, health, purpose, current_status, this_week_focus, decision_queue, delegation_ready, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpdateNode :one
UPDATE nodes
SET name = $2, type = $3, health = $4, purpose = $5, current_status = $6, this_week_focus = $7, decision_queue = $8, delegation_ready = $9, context_id = $10, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ActivateNode :exec
UPDATE nodes
SET is_active = FALSE
WHERE user_id = $1;

-- name: SetNodeActive :one
UPDATE nodes
SET is_active = TRUE, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeactivateNode :one
UPDATE nodes
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ArchiveNode :one
UPDATE nodes
SET is_archived = TRUE, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteNode :exec
DELETE FROM nodes
WHERE id = $1 AND user_id = $2;

-- name: GetNodeChildren :many
SELECT * FROM nodes
WHERE parent_id = $1 AND user_id = $2 AND is_archived = FALSE
ORDER BY sort_order ASC, name ASC;

-- name: UpdateNodeSortOrder :exec
UPDATE nodes
SET sort_order = $2, updated_at = NOW()
WHERE id = $1;
