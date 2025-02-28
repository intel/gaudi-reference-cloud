// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: airflow.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const checkUniqueAirflow = `-- name: CheckUniqueAirflow :one
SELECT count(1) as cnt FROM airflow 
WHERE cloud_account_id = $1 
and (name = $2 or workspace_name = $3)
AND is_active = true
`

type CheckUniqueAirflowParams struct {
	CloudAccountID string
	Name           string
	WorkspaceName  string
}

func (q *Queries) CheckUniqueAirflow(ctx context.Context, arg CheckUniqueAirflowParams) (int64, error) {
	row := q.db.QueryRow(ctx, checkUniqueAirflow, arg.CloudAccountID, arg.Name, arg.WorkspaceName)
	var cnt int64
	err := row.Scan(&cnt)
	return cnt, err
}

const checkUniqueAirflowId = `-- name: CheckUniqueAirflowId :one
select count(*)  as cnt
from airflow
where id like $1 || '%'
`

func (q *Queries) CheckUniqueAirflowId(ctx context.Context, dollar_1 pgtype.Text) (int64, error) {
	row := q.db.QueryRow(ctx, checkUniqueAirflowId, dollar_1)
	var cnt int64
	err := row.Scan(&cnt)
	return cnt, err
}

const commitAirflowCreate = `-- name: CommitAirflowCreate :one
update airflow
set 
    workspace_id = coalesce($2, workspace_id),
    bucket_principal = coalesce($3, bucket_principal),
    endpoint = coalesce($4, endpoint),
    backend_database_id = coalesce($5, backend_database_id),
    dag_folder_path = coalesce($6, dag_folder_path),
    plugin_folder_path = coalesce($7, plugin_folder_path),
    requirement_path = coalesce($8, requirement_path),
    log_folder = coalesce($9, log_folder),
    iks_cluster_id = coalesce($10, iks_cluster_id),
    node_group_id = coalesce($11, node_group_id),
    deployment_status_state = coalesce($12, deployment_status_state),
    deployment_status_display_name = coalesce($13, deployment_status_display_name),
    deployment_status_message = coalesce($14, deployment_status_message),   
    updated_at = now()
where id = $1
RETURNING cloud_account_id, id, name, description, version, tags, bucket_id, bucket_principal, dag_folder_path, plugin_folder_path, requirement_path, log_folder, endpoint, webserver_admin_username, webserver_admin_password_secret_id, size, number_of_nodes, number_of_workers, number_of_schedulers, deployment_id, backend_database_id, iks_cluster_id, workspace_id, workspace_name, node_group_id, deployment_status_state, deployment_status_display_name, deployment_status_message, is_active, created_at, created_by, updated_at, updated_by
`

type CommitAirflowCreateParams struct {
	ID                          string
	WorkspaceID                 pgtype.Text
	BucketPrincipal             pgtype.Text
	Endpoint                    pgtype.Text
	BackendDatabaseID           pgtype.Text
	DagFolderPath               pgtype.Text
	PluginFolderPath            pgtype.Text
	RequirementPath             pgtype.Text
	LogFolder                   pgtype.Text
	IksClusterID                pgtype.Text
	NodeGroupID                 pgtype.Text
	DeploymentStatusState       pgtype.Text
	DeploymentStatusDisplayName pgtype.Text
	DeploymentStatusMessage     pgtype.Text
}

func (q *Queries) CommitAirflowCreate(ctx context.Context, arg CommitAirflowCreateParams) (Airflow, error) {
	row := q.db.QueryRow(ctx, commitAirflowCreate,
		arg.ID,
		arg.WorkspaceID,
		arg.BucketPrincipal,
		arg.Endpoint,
		arg.BackendDatabaseID,
		arg.DagFolderPath,
		arg.PluginFolderPath,
		arg.RequirementPath,
		arg.LogFolder,
		arg.IksClusterID,
		arg.NodeGroupID,
		arg.DeploymentStatusState,
		arg.DeploymentStatusDisplayName,
		arg.DeploymentStatusMessage,
	)
	var i Airflow
	err := row.Scan(
		&i.CloudAccountID,
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Version,
		&i.Tags,
		&i.BucketID,
		&i.BucketPrincipal,
		&i.DagFolderPath,
		&i.PluginFolderPath,
		&i.RequirementPath,
		&i.LogFolder,
		&i.Endpoint,
		&i.WebserverAdminUsername,
		&i.WebserverAdminPasswordSecretID,
		&i.Size,
		&i.NumberOfNodes,
		&i.NumberOfWorkers,
		&i.NumberOfSchedulers,
		&i.DeploymentID,
		&i.BackendDatabaseID,
		&i.IksClusterID,
		&i.WorkspaceID,
		&i.WorkspaceName,
		&i.NodeGroupID,
		&i.DeploymentStatusState,
		&i.DeploymentStatusDisplayName,
		&i.DeploymentStatusMessage,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const countListAirflow = `-- name: CountListAirflow :one
SELECT count(1) FROM airflow 
WHERE cloud_account_id = $1
AND is_active = true 
AND lower(name) like lower('%' || coalesce($2, name) || '%')
and (lower(workspace_id) = lower(coalesce($3, workspace_id)) or workspace_id is null)
`

type CountListAirflowParams struct {
	CloudAccountID string
	Name           pgtype.Text
	WorkspaceID    pgtype.Text
}

func (q *Queries) CountListAirflow(ctx context.Context, arg CountListAirflowParams) (int64, error) {
	row := q.db.QueryRow(ctx, countListAirflow, arg.CloudAccountID, arg.Name, arg.WorkspaceID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createAirflow = `-- name: CreateAirflow :one
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
    ) RETURNING cloud_account_id, id, name, description, version, tags, bucket_id, bucket_principal, dag_folder_path, plugin_folder_path, requirement_path, log_folder, endpoint, webserver_admin_username, webserver_admin_password_secret_id, size, number_of_nodes, number_of_workers, number_of_schedulers, deployment_id, backend_database_id, iks_cluster_id, workspace_id, workspace_name, node_group_id, deployment_status_state, deployment_status_display_name, deployment_status_message, is_active, created_at, created_by, updated_at, updated_by
`

type CreateAirflowParams struct {
	CloudAccountID                 string
	ID                             string
	Name                           string
	Description                    pgtype.Text
	Version                        string
	Tags                           []byte
	BucketID                       pgtype.Text
	BucketPrincipal                pgtype.Text
	DagFolderPath                  pgtype.Text
	PluginFolderPath               pgtype.Text
	RequirementPath                pgtype.Text
	LogFolder                      pgtype.Text
	Endpoint                       pgtype.Text
	WebserverAdminUsername         pgtype.Text
	WebserverAdminPasswordSecretID pgtype.Int4
	Size                           string
	NumberOfNodes                  pgtype.Int4
	NumberOfWorkers                pgtype.Int4
	NumberOfSchedulers             pgtype.Int4
	BackendDatabaseID              string
	IksClusterID                   string
	WorkspaceID                    string
	WorkspaceName                  string
	NodeGroupID                    string
	DeploymentID                   string
	DeploymentStatusState          string
	DeploymentStatusDisplayName    pgtype.Text
	DeploymentStatusMessage        pgtype.Text
	CreatedBy                      string
}

func (q *Queries) CreateAirflow(ctx context.Context, arg CreateAirflowParams) (Airflow, error) {
	row := q.db.QueryRow(ctx, createAirflow,
		arg.CloudAccountID,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.Version,
		arg.Tags,
		arg.BucketID,
		arg.BucketPrincipal,
		arg.DagFolderPath,
		arg.PluginFolderPath,
		arg.RequirementPath,
		arg.LogFolder,
		arg.Endpoint,
		arg.WebserverAdminUsername,
		arg.WebserverAdminPasswordSecretID,
		arg.Size,
		arg.NumberOfNodes,
		arg.NumberOfWorkers,
		arg.NumberOfSchedulers,
		arg.BackendDatabaseID,
		arg.IksClusterID,
		arg.WorkspaceID,
		arg.WorkspaceName,
		arg.NodeGroupID,
		arg.DeploymentID,
		arg.DeploymentStatusState,
		arg.DeploymentStatusDisplayName,
		arg.DeploymentStatusMessage,
		arg.CreatedBy,
	)
	var i Airflow
	err := row.Scan(
		&i.CloudAccountID,
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Version,
		&i.Tags,
		&i.BucketID,
		&i.BucketPrincipal,
		&i.DagFolderPath,
		&i.PluginFolderPath,
		&i.RequirementPath,
		&i.LogFolder,
		&i.Endpoint,
		&i.WebserverAdminUsername,
		&i.WebserverAdminPasswordSecretID,
		&i.Size,
		&i.NumberOfNodes,
		&i.NumberOfWorkers,
		&i.NumberOfSchedulers,
		&i.DeploymentID,
		&i.BackendDatabaseID,
		&i.IksClusterID,
		&i.WorkspaceID,
		&i.WorkspaceName,
		&i.NodeGroupID,
		&i.DeploymentStatusState,
		&i.DeploymentStatusDisplayName,
		&i.DeploymentStatusMessage,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const deleteAirflow = `-- name: DeleteAirflow :exec
UPDATE airflow
set is_active=false,
  updated_at = now()
WHERE id = $1
AND is_active = true
`

func (q *Queries) DeleteAirflow(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteAirflow, id)
	return err
}

const getAirflowById = `-- name: GetAirflowById :one
SELECT cloud_account_id, id, name, description, version, tags, bucket_id, bucket_principal, dag_folder_path, plugin_folder_path, requirement_path, log_folder, endpoint, webserver_admin_username, webserver_admin_password_secret_id, size, number_of_nodes, number_of_workers, number_of_schedulers, deployment_id, backend_database_id, iks_cluster_id, workspace_id, workspace_name, node_group_id, deployment_status_state, deployment_status_display_name, deployment_status_message, is_active, created_at, created_by, updated_at, updated_by FROM airflow 
WHERE id = $1 AND cloud_account_id=$2 AND is_active = true 
LIMIT 1
`

type GetAirflowByIdParams struct {
	ID             string
	CloudAccountID string
}

// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
func (q *Queries) GetAirflowById(ctx context.Context, arg GetAirflowByIdParams) (Airflow, error) {
	row := q.db.QueryRow(ctx, getAirflowById, arg.ID, arg.CloudAccountID)
	var i Airflow
	err := row.Scan(
		&i.CloudAccountID,
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Version,
		&i.Tags,
		&i.BucketID,
		&i.BucketPrincipal,
		&i.DagFolderPath,
		&i.PluginFolderPath,
		&i.RequirementPath,
		&i.LogFolder,
		&i.Endpoint,
		&i.WebserverAdminUsername,
		&i.WebserverAdminPasswordSecretID,
		&i.Size,
		&i.NumberOfNodes,
		&i.NumberOfWorkers,
		&i.NumberOfSchedulers,
		&i.DeploymentID,
		&i.BackendDatabaseID,
		&i.IksClusterID,
		&i.WorkspaceID,
		&i.WorkspaceName,
		&i.NodeGroupID,
		&i.DeploymentStatusState,
		&i.DeploymentStatusDisplayName,
		&i.DeploymentStatusMessage,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const getAirflowByName = `-- name: GetAirflowByName :one
SELECT cloud_account_id, id, name, description, version, tags, bucket_id, bucket_principal, dag_folder_path, plugin_folder_path, requirement_path, log_folder, endpoint, webserver_admin_username, webserver_admin_password_secret_id, size, number_of_nodes, number_of_workers, number_of_schedulers, deployment_id, backend_database_id, iks_cluster_id, workspace_id, workspace_name, node_group_id, deployment_status_state, deployment_status_display_name, deployment_status_message, is_active, created_at, created_by, updated_at, updated_by FROM airflow 
WHERE cloud_account_id = $1 and name = $2 AND is_active = true 
LIMIT 1
`

type GetAirflowByNameParams struct {
	CloudAccountID string
	Name           string
}

func (q *Queries) GetAirflowByName(ctx context.Context, arg GetAirflowByNameParams) (Airflow, error) {
	row := q.db.QueryRow(ctx, getAirflowByName, arg.CloudAccountID, arg.Name)
	var i Airflow
	err := row.Scan(
		&i.CloudAccountID,
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Version,
		&i.Tags,
		&i.BucketID,
		&i.BucketPrincipal,
		&i.DagFolderPath,
		&i.PluginFolderPath,
		&i.RequirementPath,
		&i.LogFolder,
		&i.Endpoint,
		&i.WebserverAdminUsername,
		&i.WebserverAdminPasswordSecretID,
		&i.Size,
		&i.NumberOfNodes,
		&i.NumberOfWorkers,
		&i.NumberOfSchedulers,
		&i.DeploymentID,
		&i.BackendDatabaseID,
		&i.IksClusterID,
		&i.WorkspaceID,
		&i.WorkspaceName,
		&i.NodeGroupID,
		&i.DeploymentStatusState,
		&i.DeploymentStatusDisplayName,
		&i.DeploymentStatusMessage,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const getUpdateAirflow = `-- name: GetUpdateAirflow :one
SELECT id, description, tags FROM airflow 
WHERE id = $1 AND is_active = true LIMIT 1
`

type GetUpdateAirflowRow struct {
	ID          string
	Description pgtype.Text
	Tags        []byte
}

func (q *Queries) GetUpdateAirflow(ctx context.Context, id string) (GetUpdateAirflowRow, error) {
	row := q.db.QueryRow(ctx, getUpdateAirflow, id)
	var i GetUpdateAirflowRow
	err := row.Scan(&i.ID, &i.Description, &i.Tags)
	return i, err
}

const listAirflow = `-- name: ListAirflow :many
SELECT cloud_account_id, id, name, description, version, tags, bucket_id, bucket_principal, dag_folder_path, plugin_folder_path, requirement_path, log_folder, endpoint, webserver_admin_username, webserver_admin_password_secret_id, size, number_of_nodes, number_of_workers, number_of_schedulers, deployment_id, backend_database_id, iks_cluster_id, workspace_id, workspace_name, node_group_id, deployment_status_state, deployment_status_display_name, deployment_status_message, is_active, created_at, created_by, updated_at, updated_by FROM airflow 
WHERE cloud_account_id = $1
AND is_active = true 
AND lower(name) like lower('%' || coalesce($2, name) || '%')
and (lower(workspace_id) = lower(coalesce($3, workspace_id)) or workspace_id is null)
ORDER BY name
LIMIT $5 
OFFSET $4
`

type ListAirflowParams struct {
	CloudAccountID string
	Name           pgtype.Text
	WorkspaceID    pgtype.Text
	Offset         int64
	Limit          int64
}

func (q *Queries) ListAirflow(ctx context.Context, arg ListAirflowParams) ([]Airflow, error) {
	rows, err := q.db.Query(ctx, listAirflow,
		arg.CloudAccountID,
		arg.Name,
		arg.WorkspaceID,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Airflow
	for rows.Next() {
		var i Airflow
		if err := rows.Scan(
			&i.CloudAccountID,
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Version,
			&i.Tags,
			&i.BucketID,
			&i.BucketPrincipal,
			&i.DagFolderPath,
			&i.PluginFolderPath,
			&i.RequirementPath,
			&i.LogFolder,
			&i.Endpoint,
			&i.WebserverAdminUsername,
			&i.WebserverAdminPasswordSecretID,
			&i.Size,
			&i.NumberOfNodes,
			&i.NumberOfWorkers,
			&i.NumberOfSchedulers,
			&i.DeploymentID,
			&i.BackendDatabaseID,
			&i.IksClusterID,
			&i.WorkspaceID,
			&i.WorkspaceName,
			&i.NodeGroupID,
			&i.DeploymentStatusState,
			&i.DeploymentStatusDisplayName,
			&i.DeploymentStatusMessage,
			&i.IsActive,
			&i.CreatedAt,
			&i.CreatedBy,
			&i.UpdatedAt,
			&i.UpdatedBy,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const resizeAirflow = `-- name: ResizeAirflow :one
UPDATE airflow
set
    size = $2,
    number_of_nodes = $3,
    number_of_workers = $4,
    number_of_schedulers = $5,
    updated_at = now()
WHERE id = $1 
AND is_active = true
RETURNING cloud_account_id, id, name, description, version, tags, bucket_id, bucket_principal, dag_folder_path, plugin_folder_path, requirement_path, log_folder, endpoint, webserver_admin_username, webserver_admin_password_secret_id, size, number_of_nodes, number_of_workers, number_of_schedulers, deployment_id, backend_database_id, iks_cluster_id, workspace_id, workspace_name, node_group_id, deployment_status_state, deployment_status_display_name, deployment_status_message, is_active, created_at, created_by, updated_at, updated_by
`

type ResizeAirflowParams struct {
	ID                 string
	Size               string
	NumberOfNodes      pgtype.Int4
	NumberOfWorkers    pgtype.Int4
	NumberOfSchedulers pgtype.Int4
}

func (q *Queries) ResizeAirflow(ctx context.Context, arg ResizeAirflowParams) (Airflow, error) {
	row := q.db.QueryRow(ctx, resizeAirflow,
		arg.ID,
		arg.Size,
		arg.NumberOfNodes,
		arg.NumberOfWorkers,
		arg.NumberOfSchedulers,
	)
	var i Airflow
	err := row.Scan(
		&i.CloudAccountID,
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Version,
		&i.Tags,
		&i.BucketID,
		&i.BucketPrincipal,
		&i.DagFolderPath,
		&i.PluginFolderPath,
		&i.RequirementPath,
		&i.LogFolder,
		&i.Endpoint,
		&i.WebserverAdminUsername,
		&i.WebserverAdminPasswordSecretID,
		&i.Size,
		&i.NumberOfNodes,
		&i.NumberOfWorkers,
		&i.NumberOfSchedulers,
		&i.DeploymentID,
		&i.BackendDatabaseID,
		&i.IksClusterID,
		&i.WorkspaceID,
		&i.WorkspaceName,
		&i.NodeGroupID,
		&i.DeploymentStatusState,
		&i.DeploymentStatusDisplayName,
		&i.DeploymentStatusMessage,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const restartAirflow = `-- name: RestartAirflow :exec
UPDATE airflow
set updated_at = now()
WHERE id = $1
AND is_active = true
`

func (q *Queries) RestartAirflow(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, restartAirflow, id)
	return err
}

const updateAirflow = `-- name: UpdateAirflow :one
UPDATE airflow
set
    description = $2,
    tags = $3,
    updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING cloud_account_id, id, name, description, version, tags, bucket_id, bucket_principal, dag_folder_path, plugin_folder_path, requirement_path, log_folder, endpoint, webserver_admin_username, webserver_admin_password_secret_id, size, number_of_nodes, number_of_workers, number_of_schedulers, deployment_id, backend_database_id, iks_cluster_id, workspace_id, workspace_name, node_group_id, deployment_status_state, deployment_status_display_name, deployment_status_message, is_active, created_at, created_by, updated_at, updated_by
`

type UpdateAirflowParams struct {
	ID          string
	Description pgtype.Text
	Tags        []byte
}

func (q *Queries) UpdateAirflow(ctx context.Context, arg UpdateAirflowParams) (Airflow, error) {
	row := q.db.QueryRow(ctx, updateAirflow, arg.ID, arg.Description, arg.Tags)
	var i Airflow
	err := row.Scan(
		&i.CloudAccountID,
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Version,
		&i.Tags,
		&i.BucketID,
		&i.BucketPrincipal,
		&i.DagFolderPath,
		&i.PluginFolderPath,
		&i.RequirementPath,
		&i.LogFolder,
		&i.Endpoint,
		&i.WebserverAdminUsername,
		&i.WebserverAdminPasswordSecretID,
		&i.Size,
		&i.NumberOfNodes,
		&i.NumberOfWorkers,
		&i.NumberOfSchedulers,
		&i.DeploymentID,
		&i.BackendDatabaseID,
		&i.IksClusterID,
		&i.WorkspaceID,
		&i.WorkspaceName,
		&i.NodeGroupID,
		&i.DeploymentStatusState,
		&i.DeploymentStatusDisplayName,
		&i.DeploymentStatusMessage,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const upgradeAirflow = `-- name: UpgradeAirflow :one
UPDATE airflow
set
    version = $2,
    updated_at = now()
WHERE id = $1 
AND is_active = true
RETURNING cloud_account_id, id, name, description, version, tags, bucket_id, bucket_principal, dag_folder_path, plugin_folder_path, requirement_path, log_folder, endpoint, webserver_admin_username, webserver_admin_password_secret_id, size, number_of_nodes, number_of_workers, number_of_schedulers, deployment_id, backend_database_id, iks_cluster_id, workspace_id, workspace_name, node_group_id, deployment_status_state, deployment_status_display_name, deployment_status_message, is_active, created_at, created_by, updated_at, updated_by
`

type UpgradeAirflowParams struct {
	ID      string
	Version string
}

func (q *Queries) UpgradeAirflow(ctx context.Context, arg UpgradeAirflowParams) (Airflow, error) {
	row := q.db.QueryRow(ctx, upgradeAirflow, arg.ID, arg.Version)
	var i Airflow
	err := row.Scan(
		&i.CloudAccountID,
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Version,
		&i.Tags,
		&i.BucketID,
		&i.BucketPrincipal,
		&i.DagFolderPath,
		&i.PluginFolderPath,
		&i.RequirementPath,
		&i.LogFolder,
		&i.Endpoint,
		&i.WebserverAdminUsername,
		&i.WebserverAdminPasswordSecretID,
		&i.Size,
		&i.NumberOfNodes,
		&i.NumberOfWorkers,
		&i.NumberOfSchedulers,
		&i.DeploymentID,
		&i.BackendDatabaseID,
		&i.IksClusterID,
		&i.WorkspaceID,
		&i.WorkspaceName,
		&i.NodeGroupID,
		&i.DeploymentStatusState,
		&i.DeploymentStatusDisplayName,
		&i.DeploymentStatusMessage,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}
