-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation
-- name: GetAirflowById :one
SELECT * FROM airflow 
WHERE id = $1 AND cloud_account_id=$2 AND is_active = true 
LIMIT 1;

-- name: GetAirflowByName :one
SELECT * FROM airflow 
WHERE cloud_account_id = $1 and name = $2 AND is_active = true 
LIMIT 1;

-- name: CheckUniqueAirflow :one
SELECT count(1) as cnt FROM airflow 
WHERE cloud_account_id = $1 
and (name = $2 or workspace_name = $3)
AND is_active = true;

-- name: ListAirflow :many
SELECT * FROM airflow 
WHERE cloud_account_id = sqlc.arg('cloud_account_id')
AND is_active = true 
AND lower(name) like lower('%' || coalesce(sqlc.narg('name'), name) || '%')
and (lower(workspace_id) = lower(coalesce(sqlc.narg('workspace_id'), workspace_id)) or workspace_id is null)
ORDER BY name
LIMIT sqlc.arg('limit') 
OFFSET sqlc.arg('offset');


-- name: CountListAirflow :one
SELECT count(1) FROM airflow 
WHERE cloud_account_id = sqlc.arg('cloud_account_id')
AND is_active = true 
AND lower(name) like lower('%' || coalesce(sqlc.narg('name'), name) || '%')
and (lower(workspace_id) = lower(coalesce(sqlc.narg('workspace_id'), workspace_id)) or workspace_id is null)
;


-- name: CreateAirflow :one
INSERT INTO airflow (
    cloud_account_id,
    id ,
    name ,
    description ,
    version,
    tags,
    bucket_id,
    bucket_principal,
    dag_folder_path,
    plugin_folder_path ,
    requirement_path ,
    log_folder ,
    endpoint,
    webserver_admin_username ,
    webserver_admin_password_secret_id,
    size,
    number_of_nodes ,
    number_of_workers ,
    number_of_schedulers ,
    backend_database_id,
    iks_cluster_id,
    workspace_id,
    workspace_name,
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
        $22,
        $23,
        $24,
        $25,
        $26,
        $27,
        $28,
        $29
    ) RETURNING *;

-- name: GetUpdateAirflow :one
SELECT id, description, tags FROM airflow 
WHERE id = $1 AND is_active = true LIMIT 1;

-- name: UpdateAirflow :one
UPDATE airflow
set
    description = $2,
    tags = $3,
    updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING *;

-- name: CommitAirflowCreate :one
update airflow
set 
    workspace_id = coalesce(sqlc.narg('workspace_id'), workspace_id),
    bucket_principal = coalesce(sqlc.narg('bucket_principal'), bucket_principal),
    endpoint = coalesce(sqlc.narg('endpoint'), endpoint),
    backend_database_id = coalesce(sqlc.narg('backend_database_id'), backend_database_id),
    dag_folder_path = coalesce(sqlc.narg('dag_folder_path'), dag_folder_path),
    plugin_folder_path = coalesce(sqlc.narg('plugin_folder_path'), plugin_folder_path),
    requirement_path = coalesce(sqlc.narg('requirement_path'), requirement_path),
    log_folder = coalesce(sqlc.narg('log_folder'), log_folder),
    iks_cluster_id = coalesce(sqlc.narg('iks_cluster_id'), iks_cluster_id),
    node_group_id = coalesce(sqlc.narg('node_group_id'), node_group_id),
    deployment_status_state = coalesce(sqlc.narg('deployment_status_state'), deployment_status_state),
    deployment_status_display_name = coalesce(sqlc.narg('deployment_status_display_name'), deployment_status_display_name),
    deployment_status_message = coalesce(sqlc.narg('deployment_status_message'), deployment_status_message),   
    updated_at = now()
where id = $1
RETURNING *;

-- name: UpgradeAirflow :one
UPDATE airflow
set
    version = $2,
    updated_at = now()
WHERE id = $1 
AND is_active = true
RETURNING *;

-- name: ResizeAirflow :one
UPDATE airflow
set
    size = $2,
    number_of_nodes = $3,
    number_of_workers = $4,
    number_of_schedulers = $5,
    updated_at = now()
WHERE id = $1 
AND is_active = true
RETURNING *;

-- name: DeleteAirflow :exec
UPDATE airflow
set is_active=false,
  updated_at = now()
WHERE id = $1
AND is_active = true;

-- name: RestartAirflow :exec
UPDATE airflow
set updated_at = now()
WHERE id = $1
AND is_active = true;

-- name: CheckUniqueAirflowId :one
select count(*)  as cnt
from airflow
where id like $1 || '%';