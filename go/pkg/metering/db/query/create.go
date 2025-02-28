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

	insertInvalidatedMeteringRecordQuery = `
		INSERT INTO invalidated_metering_records (record_id,transaction_id,resource_id,region,cloud_account_id,timestamp,invalidity_reason,properties) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (transaction_id,resource_id,cloud_account_id) DO NOTHING
	`

	insertInvalidUsageRecordsQuery = `
		INSERT INTO invalid_metering_records (transaction_id,resource_id,cloud_account_id,timestamp,properties) 
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

func CreateInvalidMeteringRec(ctx context.Context, dbconn *sql.DB, record *v1.InvalidMeteringRecordCreate) error {
	log := log.FromContext(ctx).WithName("CreateInvalidMeteringRec()")

	_, err := dbconn.ExecContext(ctx,
		insertInvalidatedMeteringRecordQuery,
		record.RecordId,
		record.TransactionId,
		record.ResourceId,
		record.Region,
		record.CloudAccountId,
		record.Timestamp.AsTime(),
		record.MeteringRecordInvalidityReason,
		record.Properties,
	)

	if err != nil {
		log.Error(err, "error inserting invalid metering record in db")
		return status.Errorf(codes.Internal, "invalid metering record insertion failed")
	}

	return nil
}

// this function needs to be renamed later.
func CreateInvalidUsageRecord(ctx context.Context, dbconn *sql.DB, record *v1.UsageCreate) error {

	log := log.FromContext(ctx).WithName("CreateInvalidUsageRecord()")

	_, err := dbconn.ExecContext(ctx,
		insertInvalidUsageRecordsQuery,
		record.TransactionId,
		record.ResourceId,
		record.CloudAccountId,
		record.Timestamp.AsTime(),
		record.Properties,
	)

	if err != nil {
		log.Error(err, "inserting invalid_metering_record in db", err.Error())
		return status.Errorf(codes.Internal, "invalid metering record insertion failed")
	}

	return nil
}
