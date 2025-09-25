-- name: ListEnvironments :many
SELECT id, name, sort_order
FROM environments
ORDER BY sort_order, id;

-- name: GetEnvironment :one
SELECT id, name, sort_order
FROM environments
WHERE id = $1
ORDER BY sort_order;

-- name: CreateEnvironment :one
INSERT INTO environments (name, sort_order)
VALUES ($1,
        COALESCE((SELECT MAX(sort_order) + 1 FROM environments), 1))
RETURNING id, name, sort_order;

-- name: UpdateEnvironment :one
UPDATE environments
SET name = $2,
    sort_order = $3
WHERE id = $1
RETURNING id, name, sort_order;

-- Bulk reorder: pass IDs in desired order
-- name: ReorderEnvironments :exec
UPDATE environments AS e
SET sort_order = u.ord
FROM UNNEST($1::int[]) WITH ORDINALITY AS u(id, ord)
WHERE e.id = u.id;

-- name: DeleteEnvironment :exec
DELETE FROM environments
WHERE id = $1;
