// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: airflow_conf.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createAirflowConf = `-- name: CreateAirflowConf :one
INSERT INTO airflow_conf (
  id, airflow_id, cloud_account_id, key, value, created_by
) VALUES (
  $1, $2, $3, $4, $5, $6
)
RETURNING id, cloud_account_id, airflow_id, key, value, is_active, created_at, created_by, updated_at, updated_by
`

type CreateAirflowConfParams struct {
	ID             string
	AirflowID      string
	CloudAccountID string
	Key            string
	Value          pgtype.Text
	CreatedBy      string
}

func (q *Queries) CreateAirflowConf(ctx context.Context, arg CreateAirflowConfParams) (AirflowConf, error) {
	row := q.db.QueryRow(ctx, createAirflowConf,
		arg.ID,
		arg.AirflowID,
		arg.CloudAccountID,
		arg.Key,
		arg.Value,
		arg.CreatedBy,
	)
	var i AirflowConf
	err := row.Scan(
		&i.ID,
		&i.CloudAccountID,
		&i.AirflowID,
		&i.Key,
		&i.Value,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const deleteAirflowConf = `-- name: DeleteAirflowConf :exec
UPDATE airflow_conf
set is_active=false,
  updated_at = now()
WHERE id = $1
`

func (q *Queries) DeleteAirflowConf(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteAirflowConf, id)
	return err
}

const deleteAirflowConfByAirflowId = `-- name: DeleteAirflowConfByAirflowId :exec
UPDATE airflow_conf
set is_active=false,
  updated_at = now()
WHERE airflow_id = $1
`

func (q *Queries) DeleteAirflowConfByAirflowId(ctx context.Context, airflowID string) error {
	_, err := q.db.Exec(ctx, deleteAirflowConfByAirflowId, airflowID)
	return err
}

const getAirflowConfById = `-- name: GetAirflowConfById :one
SELECT id, cloud_account_id, airflow_id, key, value, is_active, created_at, created_by, updated_at, updated_by FROM airflow_conf
WHERE id = $1
AND is_active = true
LIMIT 1
`

// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
func (q *Queries) GetAirflowConfById(ctx context.Context, id string) (AirflowConf, error) {
	row := q.db.QueryRow(ctx, getAirflowConfById, id)
	var i AirflowConf
	err := row.Scan(
		&i.ID,
		&i.CloudAccountID,
		&i.AirflowID,
		&i.Key,
		&i.Value,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}

const listAirflowConf = `-- name: ListAirflowConf :many
SELECT id, cloud_account_id, airflow_id, key, value, is_active, created_at, created_by, updated_at, updated_by FROM airflow_conf
WHERE airflow_id = $1 
AND is_active = true
order by airflow_id, key
`

func (q *Queries) ListAirflowConf(ctx context.Context, airflowID string) ([]AirflowConf, error) {
	rows, err := q.db.Query(ctx, listAirflowConf, airflowID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AirflowConf
	for rows.Next() {
		var i AirflowConf
		if err := rows.Scan(
			&i.ID,
			&i.CloudAccountID,
			&i.AirflowID,
			&i.Key,
			&i.Value,
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

const updateAirflowConf = `-- name: UpdateAirflowConf :one
UPDATE airflow_conf
  set key = $2,
  value = $3,
  updated_at = now()
WHERE id = $1
AND is_active = true
RETURNING id, cloud_account_id, airflow_id, key, value, is_active, created_at, created_by, updated_at, updated_by
`

type UpdateAirflowConfParams struct {
	ID    string
	Key   string
	Value pgtype.Text
}

func (q *Queries) UpdateAirflowConf(ctx context.Context, arg UpdateAirflowConfParams) (AirflowConf, error) {
	row := q.db.QueryRow(ctx, updateAirflowConf, arg.ID, arg.Key, arg.Value)
	var i AirflowConf
	err := row.Scan(
		&i.ID,
		&i.CloudAccountID,
		&i.AirflowID,
		&i.Key,
		&i.Value,
		&i.IsActive,
		&i.CreatedAt,
		&i.CreatedBy,
		&i.UpdatedAt,
		&i.UpdatedBy,
	)
	return i, err
}
