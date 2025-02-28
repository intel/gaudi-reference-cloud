// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	deleteAllRecordsQuery = `DELETE FROM usage_report`
)

func DeleteAllRecords(ctx context.Context, dbconn *sql.DB) error {

	log := log.FromContext(ctx).WithName("DeleteAllRecords()")

	_, err := dbconn.ExecContext(ctx, deleteAllRecordsQuery)

	if err != nil {
		log.Error(err, "error deleting all records in db")
		return status.Errorf(codes.Internal, "metering records deletion failed")
	}

	return nil
}
