-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetHmsConfById :one
SELECT * FROM hms_conf
WHERE id = $1
AND is_active = true
LIMIT 1;

-- name: ListHmsConf :many
SELECT * FROM hms_conf
WHERE hms_id = $1 
AND is_active = true
order by group_id, key;

-- name: ListHmsConfByGroupId :many
SELECT * FROM hms_conf
WHERE hms_id = $1 
and group_id = $2
AND is_active = true
ORDER BY key;

-- name: CreateHmsConf :one
INSERT INTO hms_conf (
  id, hms_id, group_id, key, value, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateHmsConf :one
UPDATE hms_conf
  set key = $2,
  value = $3,
  updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: DeleteHmsConf :exec
UPDATE hms_conf
set is_active=false,
  updated_at = now()
WHERE id = $1;

-- name: DeleteHmsConfByHmsId :exec
UPDATE hms_conf
set is_active=false,
  updated_at = now()
WHERE hms_id = $1;