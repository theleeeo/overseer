-- Upsert a deployment (create or update version/timestamp)
-- name: UpsertDeployment :one
INSERT INTO deployments (environment_id, application_id, version, deployed_at)
VALUES ($1, $2, $3, COALESCE($4, now()))
ON CONFLICT (environment_id, application_id)
DO UPDATE
SET version = EXCLUDED.version,
    deployed_at = EXCLUDED.deployed_at
RETURNING environment_id, application_id, version, deployed_at;

-- name: ListDeploymentsFlat :many
SELECT
  environment_id,
  application_id,
  version,
  deployed_at
FROM deployments;

-- JSON shaped exactly like your /deployments response:
-- {
--   "prod":  { "auth": {"version":"1.0.0","deployedAt":"..."}, ... },
--   "stage": { ... }
-- }
-- name: DeploymentsJSON :one
-- SELECT COALESCE(
--   jsonb_object_agg(env_name, apps) FILTER (WHERE env_name IS NOT NULL),
--   '{}'::jsonb
-- ) AS deployments
-- FROM (
--   SELECT
--     e.name AS env_name,
--     jsonb_object_agg(
--       a.name,
--       jsonb_build_object(
--         'version', d.version,
--         'deployedAt', d.deployed_at
--       )
--       ORDER BY a.sort_order, a.id
--     ) AS apps
--   FROM deployments d
--   JOIN environments e ON e.id = d.environment_id
--   JOIN applications  a ON a.id = d.application_id
--   GROUP BY e.name
--   ORDER BY MIN(e.sort_order), MIN(e.id)
-- ) s;