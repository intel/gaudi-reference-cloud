-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetPostgresSizeById :one
SELECT * FROM postgres_size
WHERE id = $1 
AND is_active = true
LIMIT 1;

-- name: GetPostgresSizeByName :one
SELECT * FROM postgres_size
WHERE name = $1 
AND is_active = true
LIMIT 1;

-- name: ListPostgresSize :many
SELECT * FROM postgres_size
WHERE is_active = true
ORDER BY name;

-- name: CreatePostgresSize :one


INSERT INTO postgres_size (
id
,name
,description
,number_of_instances_default
,number_of_instances_min
,number_of_instances_max
,resource_cpu_limit
,resource_cpu_request
,resource_memory_limit
,resource_memory_request
,number_of_pgpool_instances_default
,number_of_pgpool_instances_min
,number_of_pgpool_instances_max
,resource_pgpool_cpu_limit
,resource_pgpool_cpu_request
,resource_pgpool_memory_limit
,resource_pgpool_memory_request
,disk_size_in_gb_default
,disk_size_in_gb_min
,disk_size_in_gb_max
,storage_class_name
,created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
)
RETURNING *;

-- name: UpdatePostgresSize :one
UPDATE postgres_size
set  description = $2
,number_of_instances_default = $3 
,number_of_instances_min = $4
,number_of_instances_max = $5 
,resource_cpu_limit = $6
,resource_cpu_request = $7 
,resource_memory_limit = $8 
,resource_memory_request = $9 
,number_of_pgpool_instances_default = $10
,number_of_pgpool_instances_min = $11
,number_of_pgpool_instances_max = $12
,resource_pgpool_cpu_limit = $13
,resource_pgpool_cpu_request = $14 
,resource_pgpool_memory_limit = $15
,resource_pgpool_memory_request = $16 
,disk_size_in_gb_default = $17
,disk_size_in_gb_min = $18
,disk_size_in_gb_max = $19
,storage_class_name = $20
,updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: DeletePostgresSize :exec
update postgres_size
set is_active=false,
  updated_at = now()
WHERE id = $1;