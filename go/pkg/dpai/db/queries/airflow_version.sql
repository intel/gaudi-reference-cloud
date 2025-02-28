-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetAirflowVersionById :one
SELECT * FROM airflow_version
WHERE id = $1 
AND is_active = true
LIMIT 1;

-- name: GetAirflowVersionByName :one
SELECT * FROM airflow_version
WHERE name = $1 
AND is_active = true
LIMIT 1;

-- name: ListAirflowVersion :many
SELECT * FROM airflow_version
WHERE is_active = true
ORDER BY name;

-- name: CreateAirflowVersion :one
INSERT INTO airflow_version (
    id,
    name,
    version,
    airflow_version,
    python_version ,
    postgres_version ,
    redis_version ,
    executor_type ,
    image_reference ,
    chart_reference ,
    description ,
    backward_compatible_from ,
    backend_database_version_id,
    created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,$14
)
RETURNING *;

-- name: UpdateAirflowVersion :one
UPDATE airflow_version
set 
    version = $2,
    description = $3,
    image_reference = $4,
    chart_reference = $5,
    backward_compatible_from = $6,
    updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: DeleteAirflowVersion :exec
update airflow_version
set is_active=false,
  updated_at = now()
WHERE id = $1;

-- name: ListAirflowVersionUpgrades :many
SELECT *
FROM airflow_version
WHERE version > (
        SELECT version
        FROM airflow_version
        WHERE id = $1
        AND is_active = true
    )
and backward_compatible_from <= (
        SELECT version
        FROM airflow_version
        WHERE id = $1
        AND is_active = true
    )
AND is_active = true
ORDER BY version DESC;