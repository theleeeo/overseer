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
WHERE name = $1 OR $1=''; -- filter by name if provided

-- List instances along with their latest deployments
-- name: ListInstancesAndDeployment :many
SELECT
  i.id,
  i.environment_id,
  i.application_id,
  i.name,
  d.version,
  d.deployed_at
FROM instances i
LEFT JOIN deployments d ON i.id = d.instance_id
  AND d.deployed_at = (SELECT MAX(deployed_at) FROM deployments WHERE instance_id = i.id);


