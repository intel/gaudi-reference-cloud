-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetHmsSizeById :one
SELECT * FROM hms_size
WHERE id = $1 
AND is_active = true
LIMIT 1;

-- name: GetHmsSizeByName :one
SELECT * FROM hms_size
WHERE name = $1 
AND is_active = true
LIMIT 1;

-- name: ListHmsSize :many
SELECT * FROM hms_size
WHERE is_active = true
ORDER BY name;

-- name: CreateHmsSize :one

INSERT INTO hms_size (
id
,name
,description
,instance_type_id
,number_of_instances_default
,number_of_instances_min
,number_of_instances_max
,resource_cpu_limit
,resource_cpu_request
,resource_memory_limit
,resource_memory_request
,backend_database_size_id
,created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: UpdateHmsSize :one
UPDATE hms_size
set  description = $2
,number_of_instances_default = $3 
,number_of_instances_min = $4
,number_of_instances_max = $5 
,resource_cpu_limit = $6
,resource_cpu_request = $7 
,resource_memory_limit = $8 
,resource_memory_request = $9 
,backend_database_size_id = $10
,instance_type_id = $11
,updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: DeleteHmsSize :exec
update hms_size
set is_active=false,
  updated_at = now()
WHERE id = $1;