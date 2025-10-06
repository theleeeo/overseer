-- Upsert an instance
-- name: UpsertInstance :exec
INSERT INTO instances (environment_id, application_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (instance_id)
DO UPDATE
SET 
  name = EXCLUDED.name;

-- name: ListInstances :many
SELECT
  id,
  environment_id,
  application_id,
  name
FROM instances;

