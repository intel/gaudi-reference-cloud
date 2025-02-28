// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: postgres_size.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPostgresSize = `-- name: CreatePostgresSize :one


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
RETURNING id, name, description, number_of_instances_default, number_of_instances_min, number_of_instances_max, resource_cpu_limit, resource_cpu_request, resource_memory_limit, resource_memory_request, number_of_pgpool_instances_default, number_of_pgpool_instances_min, number_of_pgpool_instances_max, resource_pgpool_cpu_limit, resource_pgpool_cpu_request, resource_pgpool_memory_limit, resource_pgpool_memory_request, disk_size_in_gb_default, disk_size_in_gb_min, disk_size_in_gb_max, storage_class_name, is_active, created_at, created_by, updated_at, updated_by
`

type CreatePostgresSizeParams struct {
	ID                             string
	Name                           string
	Description                    pgtype.Text
	NumberOfInstancesDefault       int32
	NumberOfInstancesMin           pgtype.Int4
	NumberOfInstancesMax           pgtype.Int4
	ResourceCpuLimit               pgtype.Text
	ResourceCpuRequest             pgtype.Text
	ResourceMemoryLimit            pgtype.Text
	ResourceMemoryRequest          pgtype.Text
	NumberOfPgpoolInstancesDefault int32
	NumberOfPgpoolInstancesMin     pgtype.Int4
	NumberOfPgpoolInstancesMax     pgtype.Int4
	ResourcePgpoolCpuLimit         pgtype.Text
	ResourcePgpoolCpuRequest       pgtype.Text
	ResourcePgpoolMemoryLimit      pgtype.Text
	ResourcePgpoolMemoryRequest    pgtype.Text
	DiskSizeInGbDefault            int32
	DiskSizeInGbMin                pgtype.Int4
	DiskSizeInGbMax                pgtype.Int4
	StorageClassName               pgtype.Text
	CreatedBy                      string
}

func (q *Queries) CreatePostgresSize(ctx context.Context, arg CreatePostgresSizeParams) (PostgresSize, error) {
	row := q.db.QueryRow(ctx, createPostgresSize,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.NumberOfInstancesDefault,
		arg.NumberOfInstancesMin,
		arg.NumberOfInstancesMax,
		arg.ResourceCpuLimit,
		arg.ResourceCpuRequest,
		arg.ResourceMemoryLimit,
		arg.ResourceMemoryRequest,
		arg.NumberOfPgpoolInstancesDefault,
		arg.NumberOfPgpoolInstancesMin,
		arg.NumberOfPgpoolInstancesMax,
		arg.ResourcePgpoolCpuLimit,
		arg.ResourcePgpoolCpuRequest,
		arg.ResourcePgpoolMemoryLimit,
		arg.ResourcePgpoolMemoryRequest,
		arg.DiskSizeInGbDefault,
		arg.DiskSizeInGbMin,
		arg.DiskSizeInGbMax,
		arg.StorageClassName,
		arg.CreatedBy,
	)
	var i PostgresSize
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.NumberOfInstancesDefault,
		&i.NumberOfInstancesMin,
		&i.NumberOfInstancesMax,
		&i.ResourceCpuLimit,
		&i.ResourceCpuRequest,
		&i.ResourceMemoryLimit,
		&i.ResourceMemoryRequest,
		&i.NumberOfPgpoolInstancesDefault,
		&i.NumberOfPgpoolInstancesMin,
		&i.NumberOfPgpoolInstancesMax,
		&i.ResourcePgpoolCpuLimit,
		&i.ResourcePgpoolCpuRequest,
		&i.ResourcePgpoolMemoryLimit,
		&i.ResourcePgpoolMemoryRequest,
		&i.DiskSizeInGbDefault,
		&i.DiskSizeInGbMin,
		&i.DiskSizeInGbMax,
		&i.StorageClassName,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const deletePostgresSize = `-- name: DeletePostgresSize :exec
update postgres_size
set is_active=false,
  updated_at = now()
WHERE id = $1
`

func (q *Queries) DeletePostgresSize(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deletePostgresSize, id)
	return err
}

const getPostgresSizeById = `-- name: GetPostgresSizeById :one
SELECT id, name, description, number_of_instances_default, number_of_instances_min, number_of_instances_max, resource_cpu_limit, resource_cpu_request, resource_memory_limit, resource_memory_request, number_of_pgpool_instances_default, number_of_pgpool_instances_min, number_of_pgpool_instances_max, resource_pgpool_cpu_limit, resource_pgpool_cpu_request, resource_pgpool_memory_limit, resource_pgpool_memory_request, disk_size_in_gb_default, disk_size_in_gb_min, disk_size_in_gb_max, storage_class_name, is_active, created_at, created_by, updated_at, updated_by FROM postgres_size
WHERE id = $1 
AND is_active = true
LIMIT 1
`

// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
func (q *Queries) GetPostgresSizeById(ctx context.Context, id string) (PostgresSize, error) {
	row := q.db.QueryRow(ctx, getPostgresSizeById, id)
	var i PostgresSize
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.NumberOfInstancesDefault,
		&i.NumberOfInstancesMin,
		&i.NumberOfInstancesMax,
		&i.ResourceCpuLimit,
		&i.ResourceCpuRequest,
		&i.ResourceMemoryLimit,
		&i.ResourceMemoryRequest,
		&i.NumberOfPgpoolInstancesDefault,
		&i.NumberOfPgpoolInstancesMin,
		&i.NumberOfPgpoolInstancesMax,
		&i.ResourcePgpoolCpuLimit,
		&i.ResourcePgpoolCpuRequest,
		&i.ResourcePgpoolMemoryLimit,
		&i.ResourcePgpoolMemoryRequest,
		&i.DiskSizeInGbDefault,
		&i.DiskSizeInGbMin,
		&i.DiskSizeInGbMax,
		&i.StorageClassName,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const getPostgresSizeByName = `-- name: GetPostgresSizeByName :one
SELECT id, name, description, number_of_instances_default, number_of_instances_min, number_of_instances_max, resource_cpu_limit, resource_cpu_request, resource_memory_limit, resource_memory_request, number_of_pgpool_instances_default, number_of_pgpool_instances_min, number_of_pgpool_instances_max, resource_pgpool_cpu_limit, resource_pgpool_cpu_request, resource_pgpool_memory_limit, resource_pgpool_memory_request, disk_size_in_gb_default, disk_size_in_gb_min, disk_size_in_gb_max, storage_class_name, is_active, created_at, created_by, updated_at, updated_by FROM postgres_size
WHERE name = $1 
AND is_active = true
LIMIT 1
`

func (q *Queries) GetPostgresSizeByName(ctx context.Context, name string) (PostgresSize, error) {
	row := q.db.QueryRow(ctx, getPostgresSizeByName, name)
	var i PostgresSize
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.NumberOfInstancesDefault,
		&i.NumberOfInstancesMin,
		&i.NumberOfInstancesMax,
		&i.ResourceCpuLimit,
		&i.ResourceCpuRequest,
		&i.ResourceMemoryLimit,
		&i.ResourceMemoryRequest,
		&i.NumberOfPgpoolInstancesDefault,
		&i.NumberOfPgpoolInstancesMin,
		&i.NumberOfPgpoolInstancesMax,
		&i.ResourcePgpoolCpuLimit,
		&i.ResourcePgpoolCpuRequest,
		&i.ResourcePgpoolMemoryLimit,
		&i.ResourcePgpoolMemoryRequest,
		&i.DiskSizeInGbDefault,
		&i.DiskSizeInGbMin,
		&i.DiskSizeInGbMax,
		&i.StorageClassName,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const listPostgresSize = `-- name: ListPostgresSize :many
SELECT id, name, description, number_of_instances_default, number_of_instances_min, number_of_instances_max, resource_cpu_limit, resource_cpu_request, resource_memory_limit, resource_memory_request, number_of_pgpool_instances_default, number_of_pgpool_instances_min, number_of_pgpool_instances_max, resource_pgpool_cpu_limit, resource_pgpool_cpu_request, resource_pgpool_memory_limit, resource_pgpool_memory_request, disk_size_in_gb_default, disk_size_in_gb_min, disk_size_in_gb_max, storage_class_name, is_active, created_at, created_by, updated_at, updated_by FROM postgres_size
WHERE is_active = true
ORDER BY name
`

func (q *Queries) ListPostgresSize(ctx context.Context) ([]PostgresSize, error) {
	rows, err := q.db.Query(ctx, listPostgresSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PostgresSize
	for rows.Next() {
		var i PostgresSize
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.NumberOfInstancesDefault,
			&i.NumberOfInstancesMin,
			&i.NumberOfInstancesMax,
			&i.ResourceCpuLimit,
			&i.ResourceCpuRequest,
			&i.ResourceMemoryLimit,
			&i.ResourceMemoryRequest,
			&i.NumberOfPgpoolInstancesDefault,
			&i.NumberOfPgpoolInstancesMin,
			&i.NumberOfPgpoolInstancesMax,
			&i.ResourcePgpoolCpuLimit,
			&i.ResourcePgpoolCpuRequest,
			&i.ResourcePgpoolMemoryLimit,
			&i.ResourcePgpoolMemoryRequest,
			&i.DiskSizeInGbDefault,
			&i.DiskSizeInGbMin,
			&i.DiskSizeInGbMax,
			&i.StorageClassName,
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

const updatePostgresSize = `-- name: UpdatePostgresSize :one
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
RETURNING id, name, description, number_of_instances_default, number_of_instances_min, number_of_instances_max, resource_cpu_limit, resource_cpu_request, resource_memory_limit, resource_memory_request, number_of_pgpool_instances_default, number_of_pgpool_instances_min, number_of_pgpool_instances_max, resource_pgpool_cpu_limit, resource_pgpool_cpu_request, resource_pgpool_memory_limit, resource_pgpool_memory_request, disk_size_in_gb_default, disk_size_in_gb_min, disk_size_in_gb_max, storage_class_name, is_active, created_at, created_by, updated_at, updated_by
`

type UpdatePostgresSizeParams struct {
	ID                             string
	Description                    pgtype.Text
	NumberOfInstancesDefault       int32
	NumberOfInstancesMin           pgtype.Int4
	NumberOfInstancesMax           pgtype.Int4
	ResourceCpuLimit               pgtype.Text
	ResourceCpuRequest             pgtype.Text
	ResourceMemoryLimit            pgtype.Text
	ResourceMemoryRequest          pgtype.Text
	NumberOfPgpoolInstancesDefault int32
	NumberOfPgpoolInstancesMin     pgtype.Int4
	NumberOfPgpoolInstancesMax     pgtype.Int4
	ResourcePgpoolCpuLimit         pgtype.Text
	ResourcePgpoolCpuRequest       pgtype.Text
	ResourcePgpoolMemoryLimit      pgtype.Text
	ResourcePgpoolMemoryRequest    pgtype.Text
	DiskSizeInGbDefault            int32
	DiskSizeInGbMin                pgtype.Int4
	DiskSizeInGbMax                pgtype.Int4
	StorageClassName               pgtype.Text
}

func (q *Queries) UpdatePostgresSize(ctx context.Context, arg UpdatePostgresSizeParams) (PostgresSize, error) {
	row := q.db.QueryRow(ctx, updatePostgresSize,
		arg.ID,
		arg.Description,
		arg.NumberOfInstancesDefault,
		arg.NumberOfInstancesMin,
		arg.NumberOfInstancesMax,
		arg.ResourceCpuLimit,
		arg.ResourceCpuRequest,
		arg.ResourceMemoryLimit,
		arg.ResourceMemoryRequest,
		arg.NumberOfPgpoolInstancesDefault,
		arg.NumberOfPgpoolInstancesMin,
		arg.NumberOfPgpoolInstancesMax,
		arg.ResourcePgpoolCpuLimit,
		arg.ResourcePgpoolCpuRequest,
		arg.ResourcePgpoolMemoryLimit,
		arg.ResourcePgpoolMemoryRequest,
		arg.DiskSizeInGbDefault,
		arg.DiskSizeInGbMin,
		arg.DiskSizeInGbMax,
		arg.StorageClassName,
	)
	var i PostgresSize
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.NumberOfInstancesDefault,
		&i.NumberOfInstancesMin,
		&i.NumberOfInstancesMax,
		&i.ResourceCpuLimit,
		&i.ResourceCpuRequest,
		&i.ResourceMemoryLimit,
		&i.ResourceMemoryRequest,
		&i.NumberOfPgpoolInstancesDefault,
		&i.NumberOfPgpoolInstancesMin,
		&i.NumberOfPgpoolInstancesMax,
		&i.ResourcePgpoolCpuLimit,
		&i.ResourcePgpoolCpuRequest,
		&i.ResourcePgpoolMemoryLimit,
		&i.ResourcePgpoolMemoryRequest,
		&i.DiskSizeInGbDefault,
		&i.DiskSizeInGbMin,
		&i.DiskSizeInGbMax,
		&i.StorageClassName,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}
