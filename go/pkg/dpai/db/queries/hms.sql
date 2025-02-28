-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetHmsById :one
SELECT * FROM hms 
WHERE id = $1 AND is_active = true 
LIMIT 1;

-- name: GetHmsByName :one
SELECT * FROM hms 
WHERE workspace_id = $1 and name = $2 AND is_active = true 
LIMIT 1;

-- name: ListHms :many
SELECT * FROM hms 
WHERE workspace_id = $1 AND is_active = true 
ORDER BY name;

-- name: CreateHms :one

INSERT INTO hms (
        id,
        workspace_id,
        name,
        description,
        version_id,
        size_id,
        number_of_instances,
        tags,
        endpoint,
        object_store_storage_endpoint,
        object_store_warehouse_directory,
        object_store_storage_access_key_secret_reference,
        object_store_storage_access_secret_secret_reference,
        backend_database_id,
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
        $20
    ) RETURNING *;

-- name: GetUpdateHms :one
SELECT id, description, tags FROM hms WHERE id = $1 AND is_active = true LIMIT 1;

-- name: UpdateHms :one
UPDATE hms
set
    description = $2,
    tags = $3,
    updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: UpgradeHms :one
UPDATE hms
set
    version_id = $2,
    updated_at = now()
WHERE id = $1 
AND is_active = true
RETURNING *;

-- name: ResizeHms :one
UPDATE hms
set
    size_id = $2,
    number_of_instances = $3,
    updated_at = now()
WHERE id = $1 
AND is_active = true
RETURNING *;

-- name: DeleteHms :exec
UPDATE hms
set is_active=false,
  updated_at = now()
WHERE id = $1
AND is_active = true;

-- name: RestartHms :exec
UPDATE hms
set updated_at = now()
WHERE id = $1
AND is_active = true;