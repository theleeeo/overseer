-- Create an instance
-- name: CreateInstance :one
INSERT INTO instances (environment_id, application_id, name)
VALUES ($1, $2, $3)
RETURNING id;

-- name: UpdateInstance :exec
UPDATE instances
SET name = $2
WHERE id = $1;

-- name: ListInstances :many
SELECT
  id,
  environment_id,
  application_id,
  name
FROM instances
WHERE name = $1 OR $1 IS NULL; -- filter by name if provided

