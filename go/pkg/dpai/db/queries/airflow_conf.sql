-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetAirflowConfById :one
SELECT * FROM airflow_conf
WHERE id = $1
AND is_active = true
LIMIT 1;

-- name: ListAirflowConf :many
SELECT * FROM airflow_conf
WHERE airflow_id = $1 
AND is_active = true
order by airflow_id, key;


-- name: CreateAirflowConf :one
INSERT INTO airflow_conf (
  id, airflow_id, cloud_account_id, key, value, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateAirflowConf :one
UPDATE airflow_conf
  set key = $2,
  value = $3,
  updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: DeleteAirflowConf :exec
UPDATE airflow_conf
set is_active=false,
  updated_at = now()
WHERE id = $1;

-- name: DeleteAirflowConfByAirflowId :exec
UPDATE airflow_conf
set is_active=false,
  updated_at = now()
WHERE airflow_id = $1;