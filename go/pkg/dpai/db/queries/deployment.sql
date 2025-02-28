-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetDeployment :one
SELECT * FROM deployment
WHERE cloud_account_id= coalesce(sqlc.narg('cloud_account_id'), cloud_account_id)
and id = sqlc.arg('id') LIMIT 1;


-- name: ListDeploymentsByAccountId :many
SELECT * FROM deployment
WHERE cloud_account_id=$1
ORDER BY service_type, created_at;

-- name: ListDeploymentsByWorkspaceId :many
SELECT * FROM deployment
WHERE cloud_account_id=$1
and workspace_id=$2
ORDER BY service_type, created_at;

-- name: ListDeployments :many
SELECT * FROM deployment
WHERE cloud_account_id=$1
and (workspace_id= coalesce(sqlc.narg('workspace_id'), workspace_id) or workspace_id is null)
and (service_id = coalesce(sqlc.narg('service_id'), service_id) or service_id is null)
and (service_type = coalesce(sqlc.narg('service_type'), service_type) or service_type is null)
and change_indicator = coalesce(sqlc.narg('change_indicator'), change_indicator)
and created_by = coalesce(sqlc.narg('created_by'), created_by)
and (parent_deployment_id = coalesce(sqlc.narg('parent_deployment_id'), parent_deployment_id) or parent_deployment_id is null)
and (status_state = coalesce(sqlc.narg('status_state'), status_state) or status_state is null)
ORDER BY created_at desc
LIMIT sqlc.arg('limit') 
OFFSET sqlc.arg('offset')
;

-- name: CountListDeployments :one
SELECT count(1) FROM deployment
WHERE cloud_account_id=$1
and (workspace_id= coalesce(sqlc.narg('workspace_id'), workspace_id) or workspace_id is null)
and (service_id = coalesce(sqlc.narg('service_id'), service_id) or service_id is null)
and (service_type = coalesce(sqlc.narg('service_type'), service_type) or service_type is null)
and change_indicator = coalesce(sqlc.narg('change_indicator'), change_indicator)
and created_by = coalesce(sqlc.narg('created_by'), created_by)
and (parent_deployment_id = coalesce(sqlc.narg('parent_deployment_id'), parent_deployment_id) or parent_deployment_id is null)
and (status_state = coalesce(sqlc.narg('status_state'), status_state) or status_state is null)
;
-- name: CreateDeployment :one
INSERT INTO deployment (
  cloud_account_id, workspace_id, id, service_id, service_type, change_indicator, input_payload, parent_deployment_id, created_by
) VALUES (
  $1, $2, $3,  $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- -- name: CreateDeploymentForCreate :one
-- INSERT INTO deployment (
--   cloud_account_id, id, request_id, service_type, change_indicator, input_payload, created_by
-- ) VALUES (
--   $1, $2, $3,  $4, 'CREATE', $5, $6
-- )
-- RETURNING *;

-- -- name: CreateDeploymentForUpdateOrDelete :one
-- INSERT INTO deployment (
--   cloud_account_id, id, service_id, service_type, change_indicator, input_payload, created_by
-- ) VALUES (
--   $1, $2, $3, $4, $5, $6, $7
-- )
-- RETURNING *;

-- name: GetUpdateDeployment :one
SELECT id, status_state, error_message FROM deployment
WHERE  id = $1 LIMIT 1;

-- name: UpdateDeploymentNodeGroupId :one
UPDATE deployment
  set node_group_id = $2,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateDeployment :one
UPDATE deployment
  set status_state = $2,
  status_display_name = $3,
  status_message = $4,
  error_message = $5,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateDeploymentStatusAsRunning :one
UPDATE deployment
  set status_state = 'DPAI_RUNNING',
    status_display_name = $2,
  status_message = $3,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateDeploymentStatusAsFailed :one
UPDATE deployment
  set status_state = 'DPAI_FAILED',
  status_display_name = $2,
  status_message = $3,
  error_message = $4,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateDeploymentStatusAsSuccess :one
UPDATE deployment
  set status_state = 'DPAI_SUCCESS',
    status_display_name = $3,
  status_message = $4,
  output_payload = $2,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDeployment :exec
DELETE FROM deployment
WHERE id = $1;

-- name: CheckUniqueDeploymentId :one
select count(*)  as cnt
from deployment
where id like $1 || '%';