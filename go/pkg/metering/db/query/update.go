// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/protodb"
)

// SQL query is limited to 64K bytes so we need to do this in chunks
const MAX_UPDATE_IDS = 5000

func UpdateUsageRecord(ctx context.Context, dbconn *sql.DB, updates *v1.UsageUpdate) error {

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for ids := updates.GetId(); len(ids) > 0; {
		amount := len(ids)
		if amount > MAX_UPDATE_IDS {
			amount = MAX_UPDATE_IDS
		}
		args := []any{updates.GetReported()}
		args, argStr := protodb.AddArrayArgs(args, ids[:amount])
		query := "UPDATE usage_report SET reported=$1 WHERE id IN (" + argStr + ")"
		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
		ids = ids[amount:]
	}

	return tx.Commit()
}
