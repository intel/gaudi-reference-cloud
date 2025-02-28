// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	insertUsageRecordQuery = `
		INSERT INTO usage_report (transaction_id,resource_id,cloud_account_id,timestamp,properties) 
		VALUES ($1, $2, $3, $4,$5)
		ON CONFLICT (transaction_id,resource_id,cloud_account_id) DO NOTHING
	`
)

func CreateUsageRecord(ctx context.Context, dbconn *sql.DB, record *v1.UsageCreate) error {

	log := log.FromContext(ctx).WithName("CreateUsageRecord()")

	_, err := dbconn.ExecContext(ctx,
		insertUsageRecordQuery,
		record.TransactionId,
		record.ResourceId,
		record.CloudAccountId,
		record.Timestamp.AsTime(),
		record.Properties,
	)

	if err != nil {
		log.Error(err, "error updating usage_report state in db")
		return status.Errorf(codes.Internal, "metering record insertion failed")
	}

	return nil
}
