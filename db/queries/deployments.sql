-- Register a deployment
-- name: RegisterDeployment :exec
INSERT INTO deployments (id, instance_id, version, deployed_at)
VALUES ($1, $2, $3, $4);

-- name: ListDeployments :many
SELECT
  instance_id,
  version,
  deployed_at
FROM deployments;

