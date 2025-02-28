-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetHmsVersionById :one
SELECT * FROM hms_version
WHERE id = $1
AND is_active = true
LIMIT 1;

-- name: GetHmsVersionByName :one
SELECT * FROM hms_version
WHERE name = $1 
AND is_active = true
LIMIT 1;

-- name: ListHmsVersion :many
SELECT * FROM hms_version
WHERE is_active = true
ORDER BY name;

-- name: CreateHmsVersion :one


INSERT INTO hms_version (
  id, name, description, version, hms_version,
  image_reference, chart_reference,backward_compatible_from, backend_database_version_id, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateHmsVersion :one
UPDATE hms_version
  set description = $2,
  image_reference = $3,
  chart_reference = $4,
  backward_compatible_from = $5,
  updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: DeleteHmsVersion :exec
UPDATE hms_version
set is_active=false,
  updated_at = now()
WHERE id = $1;

-- name: ListHmsVersionUpgrades :many
SELECT *
FROM hms_version
WHERE version > (
        SELECT version
        FROM hms_version
        WHERE id = $1
        AND is_active = true
    )
and backward_compatible_from <= (
        SELECT version
        FROM hms_version
        WHERE id = $1
        AND is_active = true
    )
AND is_active = true
ORDER BY version DESC;