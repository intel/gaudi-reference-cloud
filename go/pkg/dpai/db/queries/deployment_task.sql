-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetDeploymentTask :one
SELECT * FROM deployment_task
WHERE id = $1 LIMIT 1;


-- name: ListDeploymentTasks :many
SELECT * FROM deployment_task
WHERE deployment_id = $1
ORDER BY created_at;

-- name: CreateDeploymentTask :one
INSERT INTO deployment_task (
  id, deployment_id, name, description, status_state, status_display_name, status_message
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetUpdateDeploymentTask :one
SELECT id, status_state, status_display_name, status_message, error_message FROM deployment_task
WHERE  id = $1 LIMIT 1;

-- name: UpdateDeploymentTask :one
UPDATE deployment_task
  set status_state = $2,
  status_display_name = $3,
  status_message = $4,
  error_message = $5,
  ended_at = now(),
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateDeploymentTaskStatusAsRunning :exec
UPDATE deployment_task
  set status_state = 'DPAI_RUNNING',
  status_display_name = $2,
  status_message = $3,
  started_at = now(),
  updated_at = now()
WHERE id = $1;

-- name: UpdateDeploymentTaskStatusAsFailed :exec
UPDATE deployment_task
  set status_state = 'DPAI_FAILED',
  status_display_name = $2,
  status_message = $3,
  error_message = $4,
  ended_at = now(),
  updated_at = now()
WHERE id = $1;

-- name: UpdateDeploymentTaskStatusAsSuccess :exec
UPDATE deployment_task
  set status_state = 'DPAI_SUCCESS',
  status_display_name = $2,
  status_message = $3,
  output_payload = $4,
  ended_at = now(),
  updated_at = now()
WHERE id = $1;

-- name: UpdateDeploymentTaskStatusAsWaitingForUpstream :exec
UPDATE deployment_task
  set status_state = 'DPAI_WAITING_FOR_UPSTREAM',
  status_display_name = $2,
  status_message = $3,
  updated_at = now()
WHERE id = $1;

-- name: UpdateDeploymentTaskStatusAsUpstreamFailed :exec
UPDATE deployment_task
  set status_state = 'DPAI_UPSTREAM_FAILED',
    status_display_name = $2,
  status_message = $3,
  error_message = $4,
  updated_at = now(),
  ended_at = now()
WHERE id = $1;


-- name: DeleteDeploymentTask :exec
DELETE FROM deployment_task
WHERE id = $1;

-- name: DeleteDeploymentTaskByDeploymentId :exec
DELETE FROM deployment_task
WHERE deployment_id = $1;

-- name: CheckUniqueDeploymentTaskId :one
select count(*)  as cnt
from deployment_task
where id like $1 || '%';