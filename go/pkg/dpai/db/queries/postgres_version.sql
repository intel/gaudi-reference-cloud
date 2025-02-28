-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetPostgresVersionById :one
SELECT * FROM postgres_version
WHERE id = $1
AND is_active = true
LIMIT 1;

-- name: GetPostgresVersionByName :one
SELECT * FROM postgres_version
WHERE name = $1 
AND is_active = true
LIMIT 1;

-- name: ListPostgresVersion :many
SELECT * FROM postgres_version
WHERE is_active = true
ORDER BY name;

-- name: CreatePostgresVersion :one


INSERT INTO postgres_version (
  id, name, description, version, postgres_version,
  image_reference, chart_reference,backward_compatible_from, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdatePostgresVersion :one
UPDATE postgres_version
  set description = $2,
  image_reference = $3,
  chart_reference = $4,
  backward_compatible_from = $5,
  updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: DeletePostgresVersion :exec
UPDATE postgres_version
set is_active=false,
  updated_at = now()
WHERE id = $1;

-- name: ListPostgresVersionUpgrades :many
SELECT *
FROM postgres_version
WHERE version > (
        SELECT version
        FROM postgres_version
        WHERE id = $1
        AND is_active = true
    )
and backward_compatible_from <= (
        SELECT version
        FROM postgres_version
        WHERE id = $1
        AND is_active = true
    )
AND is_active = true
ORDER BY version DESC;