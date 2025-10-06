-- Upsert a deployment (create or update version/timestamp)
-- name: UpsertDeployment :exec
INSERT INTO deployments (instance_id, version, deployed_at)
VALUES ($1, $2, $3)
ON CONFLICT (environment_id, application_id)
DO UPDATE
SET version = EXCLUDED.version,
    deployed_at = EXCLUDED.deployed_at;

-- name: ListDeployments :many
SELECT
  instance_id,
  version,
  deployed_at
FROM deployments;

