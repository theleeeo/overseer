-- name: ListApplications :many
SELECT id, name, sort_order
FROM applications
ORDER BY sort_order, id;

-- name: GetApplication :one
SELECT id, name, sort_order
FROM applications
WHERE id = $1
ORDER BY sort_order;

-- name: CreateApplication :one
INSERT INTO applications (name, sort_order)
VALUES ($1,
        COALESCE((SELECT MAX(sort_order) + 1 FROM applications), 1))
RETURNING id, name, sort_order;

-- name: UpdateApplication :one
UPDATE applications
SET name = $2
WHERE id = $1
RETURNING id, name, sort_order;

-- Bulk reorder: pass IDs in desired order (first element gets sort_order=1, etc.)
-- name: ReorderApplications :exec
UPDATE applications AS a
SET sort_order = u.ord
FROM UNNEST($1::int[]) WITH ORDINALITY AS u(id, ord)
WHERE a.id = u.id;

-- name: DeleteApplication :exec
DELETE FROM applications
WHERE id = $1;