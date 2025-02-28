-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetAirflowSizeById :one
SELECT * FROM airflow_size
WHERE id = $1 
AND is_active = true
LIMIT 1;

-- name: GetAirflowSizeByName :one
SELECT * FROM airflow_size
WHERE name = $1 
AND is_active = true
LIMIT 1;

-- name: ListAirflowSize :many
SELECT * FROM airflow_size
WHERE is_active = true
ORDER BY name;

-- name: CreateAirflowSize :one
INSERT INTO airflow_size (
    id,
    name,
    description,
    number_of_nodes_default,
    node_size_id,
    backend_database_size_id,
    webserver_count,
    webserver_cpu_limit,
    webserver_memory_limit,
    webserver_cpu_request,
    webserver_memory_request,
    log_directory_disk_size,
    redis_disk_size,
    schedular_count_default,
    scheduler_count_min,
    scheduler_count_max,
    scheduler_cpu_limit,
    scheduler_memory_limit,
    scheduler_memory_request,
    scheduler_cpu_request,
    worker_count_default,
    worker_count_min,
    worker_count_max,
    worker_memory_limit,
    worker_memory_request,
    worker_cpu_limit,
    worker_cpu_request,
    trigger_count,
    trigger_memory_limit,
    trigger_memory_request,
    trigger_cpu_limit,
    trigger_cpu_request,
    created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33
)
RETURNING *;

-- name: UpdateAirflowSize :one
UPDATE airflow_size
set 
    description = $2
    ,number_of_nodes_default = $3
    ,node_size_id = $4
    ,backend_database_size_id = $5
    ,webserver_count = $6
    ,webserver_cpu_limit = $7
    ,webserver_memory_limit = $8
    ,webserver_cpu_request = $9
    ,webserver_memory_request = $10
    ,log_directory_disk_size = $11
    ,redis_disk_size = $12
    ,schedular_count_default = $13
    ,scheduler_count_min = $14
    ,scheduler_count_max = $15
    ,scheduler_cpu_limit = $16
    ,scheduler_memory_limit = $17
    ,scheduler_memory_request = $18
    ,scheduler_cpu_request = $19
    ,worker_count_default = $20
    ,worker_count_min = $21
    ,worker_count_max = $22
    ,worker_memory_limit = $23
    ,worker_memory_request = $24
    ,worker_cpu_limit = $25
    ,worker_cpu_request = $26
    ,trigger_count = $27
    ,trigger_memory_limit = $28
    ,trigger_memory_request = $29
    ,trigger_cpu_limit = $30
    ,trigger_cpu_request = $31
    ,updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: DeleteAirflowSize :exec
update airflow_size
set is_active=false,
  updated_at = now()
WHERE id = $1;