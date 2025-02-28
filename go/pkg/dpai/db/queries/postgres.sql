-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetPostgresById :one
SELECT * FROM postgres WHERE id = $1 AND is_active = true LIMIT 1;

-- name: GetPostgresByName :one
SELECT * FROM postgres WHERE workspace_id = $1 and name = $2 AND is_active = true LIMIT 1;

-- name: ListPostgres :many
SELECT * FROM postgres 
WHERE 
cloud_account_id = sqlc.arg('cloud_account_id')
and (lower(workspace_id) = lower(coalesce(sqlc.narg('workspace_id'), workspace_id)) or workspace_id is null)
AND is_active = true 
ORDER BY name
LIMIT sqlc.arg('limit') 
OFFSET sqlc.arg('offset');

-- name: CreatePostgres :one

INSERT INTO postgres (
        id,
        cloud_account_id,
        workspace_id,
        name,
        version_id,
        size_id,
        description,
        number_of_instances,
        number_of_pgpool_instances,
        disk_size_in_gb,
        initial_database_name,
        admin_username,
        admin_password_secret_reference,
        advance_configuration,
        tags,
        server_url,
        node_group_id,
        deployment_id,
        deployment_status_state,
        deployment_status_display_name,
        deployment_status_message,
        created_by
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17,
        $18,
        $19,
        $20,
        $21,
        $22
    ) RETURNING *;

-- name: CommitPostgresCreate :one
UPDATE postgres
set number_of_instances = coalesce(sqlc.narg('number_of_instances'), number_of_instances),
    number_of_pgpool_instances = coalesce(sqlc.narg('number_of_pgpool_instances'), number_of_pgpool_instances),
    disk_size_in_gb = coalesce(sqlc.narg('disk_size_in_gb'), disk_size_in_gb),
    server_url = coalesce(sqlc.narg('server_url'), server_url),
    deployment_status_state = coalesce(sqlc.narg('deployment_status_state'), deployment_status_state),
    deployment_status_display_name = coalesce(sqlc.narg('deployment_status_display_name'), deployment_status_display_name),
    deployment_status_message = coalesce(sqlc.narg('deployment_status_message'), deployment_status_message),
  updated_at = now()
WHERE id = $1
and is_active = true
RETURNING *;

-- name: GetUpdatePostgres :one
SELECT id, description, tags FROM postgres WHERE id = $1 AND is_active = true LIMIT 1;

-- name: UpdatePostgres :one
UPDATE postgres
set
    description = $2,
    tags = $3,
    updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: UpgradePostgres :one
UPDATE postgres
set
    version_id = $2,
    updated_at = now()
WHERE id = $1 
AND is_active = true
RETURNING *;

-- name: ResizePostgres :one
UPDATE postgres
set
    size_id = $2,
    number_of_instances = $3,
    number_of_pgpool_instances = $4,
    disk_size_in_gb = $5,
    updated_at = now()
WHERE id = $1 
AND is_active = true
RETURNING *;

-- name: DeletePostgres :exec
UPDATE postgres
set is_active=false,
  updated_at = now()
WHERE id = $1
AND is_active = true;

-- name: RestartPostgres :exec
UPDATE postgres
set updated_at = now()
WHERE id = $1
AND is_active = true;

-- name: CheckUniquePostgresId :one
select count(*)  as cnt
from postgres
where id like $1 || '%';