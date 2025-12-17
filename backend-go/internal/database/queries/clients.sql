-- name: ListClients :many
SELECT * FROM clients
WHERE user_id = $1
  AND (sqlc.narg(status)::clientstatus IS NULL OR status = sqlc.narg(status))
  AND (sqlc.narg(client_type)::clienttype IS NULL OR type = sqlc.narg(client_type))
  AND (sqlc.narg(search)::text IS NULL OR name ILIKE '%' || sqlc.narg(search) || '%')
ORDER BY updated_at DESC;

-- name: GetClient :one
SELECT * FROM clients
WHERE id = $1 AND user_id = $2;

-- name: CreateClient :one
INSERT INTO clients (user_id, name, type, email, phone, website, industry, company_size, address, city, state, zip_code, country, status, source, assigned_to, tags, custom_fields, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
RETURNING *;

-- name: UpdateClient :one
UPDATE clients
SET name = $2, type = $3, email = $4, phone = $5, website = $6, industry = $7, company_size = $8,
    address = $9, city = $10, state = $11, zip_code = $12, country = $13, status = $14, source = $15,
    assigned_to = $16, tags = $17, custom_fields = $18, notes = $19, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateClientStatus :one
UPDATE clients
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteClient :exec
DELETE FROM clients
WHERE id = $1 AND user_id = $2;

-- name: ListClientContacts :many
SELECT * FROM client_contacts
WHERE client_id = $1
ORDER BY is_primary DESC, name ASC;

-- name: CreateClientContact :one
INSERT INTO client_contacts (client_id, name, email, phone, role, is_primary, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateClientContact :one
UPDATE client_contacts
SET name = $2, email = $3, phone = $4, role = $5, is_primary = $6, notes = $7, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteClientContact :exec
DELETE FROM client_contacts
WHERE id = $1 AND client_id = $2;

-- name: ListClientInteractions :many
SELECT * FROM client_interactions
WHERE client_id = $1
ORDER BY occurred_at DESC;

-- name: CreateClientInteraction :one
INSERT INTO client_interactions (client_id, contact_id, type, subject, description, outcome, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListClientDeals :many
SELECT * FROM client_deals
WHERE client_id = $1
ORDER BY created_at DESC;

-- name: CreateClientDeal :one
INSERT INTO client_deals (client_id, name, value, stage, probability, expected_close_date, notes)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateClientDeal :one
UPDATE client_deals
SET name = $2, value = $3, stage = $4, probability = $5, expected_close_date = $6, notes = $7, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDealStage :one
UPDATE client_deals
SET stage = $2, updated_at = NOW(), closed_at = CASE WHEN $2 IN ('closed_won', 'closed_lost') THEN NOW() ELSE NULL END
WHERE id = $1
RETURNING *;

-- name: ListDeals :many
SELECT d.*, c.name as client_name
FROM client_deals d
JOIN clients c ON c.id = d.client_id
WHERE c.user_id = $1
  AND (sqlc.narg(stage)::dealstage IS NULL OR d.stage = sqlc.narg(stage))
ORDER BY d.updated_at DESC;
