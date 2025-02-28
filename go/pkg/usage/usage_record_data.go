// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const insertProductUsageRecordQuery = `
	INSERT INTO product_usage_records (id, transaction_id, cloud_account_id, product_name, region, quantity, timestamp, 
		start_time, end_time, properties)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

const searchProductUsageRecordQueryBase = `
		SELECT id, transaction_id, cloud_account_id, product_name, region, quantity, timestamp, reported, start_time, end_time, properties
		FROM product_usage_records
	`

const searchInvalidProductUsageRecordQueryBase = `
		SELECT id, record_id, transaction_id, cloud_account_id, product_name, region, quantity, timestamp, start_time,
		end_time, invalidity_reason, properties
		FROM invalid_product_usage_records
	`

const insertInvalidProductUsageRecordQuery = `
	INSERT INTO invalid_product_usage_records (id, record_id, transaction_id, cloud_account_id, product_name, region, quantity, timestamp, 
		start_time, end_time, invalidity_reason, properties) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
`

type UsageRecordData struct {
	session *sql.DB
}

func NewUsageRecordData(session *sql.DB) *UsageRecordData {
	return &UsageRecordData{session: session}
}

func (urd UsageRecordData) StoreProductUsageRecord(ctx context.Context, productUsageRecordCreate *pb.ProductUsageRecordCreate) error {
	logger := log.FromContext(ctx).WithName("UsageRecordData.StoreProductUsageRecord")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	_, err := urd.session.ExecContext(ctx,
		insertProductUsageRecordQuery,
		uuid.NewString(),
		productUsageRecordCreate.TransactionId,
		productUsageRecordCreate.CloudAccountId,
		productUsageRecordCreate.ProductName,
		productUsageRecordCreate.Region,
		productUsageRecordCreate.Quantity,
		productUsageRecordCreate.Timestamp.AsTime(),
		productUsageRecordCreate.StartTime.AsTime(),
		productUsageRecordCreate.EndTime.AsTime(),
		productUsageRecordCreate.Properties,
	)

	if err != nil {
		return err
	}

	return nil
}

func (urd UsageRecordData) StoreInvalidProductUsageRecord(ctx context.Context, invalidProductUsageRecordCreate *pb.InvalidProductUsageRecordCreate) error {
	logger := log.FromContext(ctx).WithName("UsageRecordData.StoreInvalidProductUsageRecord")
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	_, err := urd.session.ExecContext(ctx,
		insertInvalidProductUsageRecordQuery,
		uuid.NewString(),
		invalidProductUsageRecordCreate.RecordId,
		invalidProductUsageRecordCreate.TransactionId,
		invalidProductUsageRecordCreate.CloudAccountId,
		invalidProductUsageRecordCreate.ProductName,
		invalidProductUsageRecordCreate.Region,
		invalidProductUsageRecordCreate.Quantity,
		invalidProductUsageRecordCreate.Timestamp.AsTime(),
		invalidProductUsageRecordCreate.StartTime.AsTime(),
		invalidProductUsageRecordCreate.EndTime.AsTime(),
		invalidProductUsageRecordCreate.ProductUsageRecordInvalidityReason,
		invalidProductUsageRecordCreate.Properties,
	)

	if err != nil {
		return err
	}

	return nil
}

func (urd UsageRecordData) MarkProductUsageRecordAsReported(ctx context.Context, productUsageRecordId string) error {

	tx, err := urd.session.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	updateProductUsageRecordQuery := "UPDATE product_usage_records SET reported=$1 WHERE id=$2"

	_, err = tx.ExecContext(ctx, updateProductUsageRecordQuery,
		true, productUsageRecordId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

func (urd UsageRecordData) GetUnreportedProductUsageRecords(ctx context.Context) ([]*pb.ProductUsageRecord, error) {

	logger := log.FromContext(ctx).WithName("UsageRecordData.GetUnreportedProductUsageRecords")

	productUsageRecords := []*pb.ProductUsageRecord{}

	searchUnreportedProductUsageRecords := `
		SELECT id, transaction_id, cloud_account_id, product_name, region, quantity, timestamp, reported, start_time, end_time, properties
		FROM product_usage_records WHERE reported=false ORDER BY timestamp
	`
	rows, err := urd.session.QueryContext(ctx, searchUnreportedProductUsageRecords)

	if err != nil {
		logger.Error(err, "error searching unreported product usage records")
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		u := &pb.ProductUsageRecord{}
		props := []byte{}
		ts := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}

		if err := rows.Scan(&u.Id, &u.TransactionId, &u.CloudAccountId,
			&u.ProductName, &u.Region, &u.Quantity, &ts, &u.Reported, &startTime, &endTime, &props); err != nil {
			logger.Error(err, "error reading result row for unreported product usage records, continuing")
		}

		u.Timestamp = timestamppb.New(ts)
		u.StartTime = timestamppb.New(startTime)
		u.EndTime = timestamppb.New(endTime)

		err := json.Unmarshal([]byte(props), &u.Properties)
		if err != nil {
			logger.Error(err, "error unmarshalling JSON properties for product usage record")
			continue
		}

		productUsageRecords = append(productUsageRecords, u)

	}

	return productUsageRecords, nil
}

func (urd UsageRecordData) SearchProductUsageRecords(rs pb.UsageRecordService_SearchProductUsageRecordsServer,
	filter *pb.ProductUsageRecordsFilter) error {

	logger := log.FromContext(rs.Context()).WithName("UsageRecordData.SearchProductUsageRecords")
	var q strings.Builder

	_, err := q.WriteString(searchProductUsageRecordQueryBase)
	if err != nil {
		logger.Error(err, "failed to write search query base for searching product usage records")
		return err
	}

	// following clause is essentially a no-op
	// used to complete the below concatanated where clauses
	if _, err := q.WriteString(" WHERE 1=1"); err != nil {
		logger.Error(err, "failed to write search query base for searching product usage records")
	}

	params := []interface{}{}

	if filter.Id != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND id=$%d ", len(params)+1))
		params = append(params, filter.GetId())
		if err != nil {
			logger.Error(err, "failed to append id to search query base for searching product usage records")
		}
	}

	if filter.CloudAccountId != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND cloud_account_id=$%d ", len(params)+1))
		params = append(params, filter.GetCloudAccountId())
		if err != nil {
			logger.Error(err, "failed to append cloud account id to search query base for searching product usage records")
		}
	}

	if filter.TransactionId != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND transaction_id=$%d ", len(params)+1))
		params = append(params, filter.GetTransactionId())
		if err != nil {
			logger.Error(err, "failed to append transaction id to search query base for searching product usage records")
		}
	}

	if filter.Region != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND region=$%d ", len(params)+1))
		params = append(params, filter.GetRegion())
		if err != nil {
			logger.Error(err, "failed to append region to search query base for searching product usage records")
		}
	}

	if filter.StartTime != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND start_time>=$%d ", len(params)+1))
		params = append(params, (filter.GetStartTime()).AsTime())
		if err != nil {
			logger.Error(err, "failed to append start time to search query base for searching product usage records")
		}
	}

	if filter.EndTime != nil {
		_, err := q.WriteString(fmt.Sprintf("   AND end_time<=$%d ", len(params)+1))
		params = append(params, (filter.GetEndTime()).AsTime())
		if err != nil {
			logger.Error(err, "failed to append end time to search query base for searching product usage records")
		}
	}

	if filter.Reported != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND reported=$%d ", len(params)+1))
		params = append(params, (filter.GetReported()))
		if err != nil {
			logger.Error(err, "failed to append reported to search query base for searching product usage records")
		}
	}

	_, err = q.WriteString(" ORDER BY timestamp")
	if err != nil {
		logger.Error(err, "failed to append order by to search query base for searching product usage records")
	}

	rows, err := urd.session.QueryContext(rs.Context(), q.String(), params...)
	if err != nil {
		logger.Error(err, "error searching product usage records")
		return status.Errorf(codes.Internal, "error searching product usage records")
	}

	defer rows.Close()

	for rows.Next() {

		productUsageRecord := pb.ProductUsageRecord{}

		props := []byte{}
		ts := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}

		if err := rows.Scan(&productUsageRecord.Id, &productUsageRecord.TransactionId, &productUsageRecord.CloudAccountId,
			&productUsageRecord.ProductName, &productUsageRecord.Region, &productUsageRecord.Quantity, &ts,
			&productUsageRecord.Reported, &startTime, &endTime, &props); err != nil {
			logger.Error(err, "error reading result row for product usage records, continuing")
			continue
		}

		productUsageRecord.Timestamp = timestamppb.New(ts)
		productUsageRecord.StartTime = timestamppb.New(startTime)
		productUsageRecord.EndTime = timestamppb.New(endTime)

		err := json.Unmarshal([]byte(props), &productUsageRecord.Properties)
		if err != nil {
			logger.Error(err, "error unmarshalling JSON properties for product usage record")
			continue
		}

		if err := rs.Send(&productUsageRecord); err != nil {
			logger.Error(err, "error sending product usage records")
		}
	}

	return nil
}

func (urd UsageRecordData) SearchInvalidProductUsageRecords(rs pb.UsageRecordService_SearchInvalidProductUsageRecordsServer,
	filter *pb.InvalidProductUsageRecordsFilter) error {

	logger := log.FromContext(rs.Context()).WithName("UsageRecordData.SearchInvalidProductUsageRecords")
	var q strings.Builder

	_, err := q.WriteString(searchInvalidProductUsageRecordQueryBase)
	if err != nil {
		logger.Error(err, "failed to write search query base for searching invalid product usage records")
		return err
	}

	// following clause is essentially a no-op
	// used to complete the below concatanated where clauses
	if _, err := q.WriteString(" WHERE 1=1"); err != nil {
		logger.Error(err, "failed to write search query base for searching invalid product usage records")
	}

	params := []interface{}{}

	if filter.Id != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND id=$%d ", len(params)+1))
		params = append(params, filter.GetId())
		if err != nil {
			logger.Error(err, "failed to append id to search query base for searching invalid product usage records")
		}
	}

	if filter.CloudAccountId != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND cloud_account_id=$%d ", len(params)+1))
		params = append(params, filter.GetCloudAccountId())
		if err != nil {
			logger.Error(err, "failed to append cloud account id to search query base for searching invalid product usage records")
		}
	}

	if filter.TransactionId != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND transaction_id=$%d ", len(params)+1))
		params = append(params, filter.GetTransactionId())
		if err != nil {
			logger.Error(err, "failed to append transaction id to search query base for searching invalid product usage records")
		}
	}

	if filter.Region != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND region=$%d ", len(params)+1))
		params = append(params, filter.GetRegion())
		if err != nil {
			logger.Error(err, "failed to append region to search query base for searching invalid product usage records")
		}
	}

	if filter.StartTime != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND start_time>=$%d ", len(params)+1))
		params = append(params, (filter.GetStartTime()).AsTime())
		if err != nil {
			logger.Error(err, "failed to append start time to search query base for searching invalid product usage records")
		}
	}

	if filter.EndTime != nil {
		_, err := q.WriteString(fmt.Sprintf("   AND end_time<=$%d ", len(params)+1))
		params = append(params, (filter.GetEndTime()).AsTime())
		if err != nil {
			logger.Error(err, "failed to append end time to search query base for searching invalid product usage records")
		}
	}

	_, err = q.WriteString(" ORDER BY timestamp")
	if err != nil {
		logger.Error(err, "failed to append order by to search query base for searching invalid product usage records")
	}

	rows, err := urd.session.QueryContext(rs.Context(), q.String(), params...)
	if err != nil {
		logger.Error(err, "error searching invalid product usage records")
		return status.Errorf(codes.Internal, "error searching invalid product usage records")
	}

	defer rows.Close()

	for rows.Next() {

		invalidProductUsageRecord := pb.InvalidProductUsageRecord{}

		props := []byte{}
		ts := time.Time{}
		startTime := time.Time{}
		endTime := time.Time{}

		if err := rows.Scan(&invalidProductUsageRecord.Id, &invalidProductUsageRecord.RecordId,
			&invalidProductUsageRecord.TransactionId, &invalidProductUsageRecord.CloudAccountId,
			&invalidProductUsageRecord.ProductName, &invalidProductUsageRecord.Region, &invalidProductUsageRecord.Quantity, &ts,
			&startTime, &endTime, &invalidProductUsageRecord.ProductUsageRecordInvalidityReason, &props); err != nil {
			logger.Error(err, "error reading result row for invalid product usage records, continuing")
			continue
		}

		invalidProductUsageRecord.Timestamp = timestamppb.New(ts)
		invalidProductUsageRecord.StartTime = timestamppb.New(startTime)
		invalidProductUsageRecord.EndTime = timestamppb.New(endTime)

		err := json.Unmarshal([]byte(props), &invalidProductUsageRecord.Properties)
		if err != nil {
			logger.Error(err, "error unmarshalling JSON properties for invalid product usage record")
			continue
		}

		if err := rs.Send(&invalidProductUsageRecord); err != nil {
			logger.Error(err, "error sending invalid product usage records")
		}
	}

	return nil
}

func (urd UsageRecordData) DeleteAllUsageRecords(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("UsageRecordData.DeleteAllUsageRecords")

	const deleteProductUsageRecords = `DELETE FROM product_usage_records`
	const deleteInvalidProductUsageRecords = `DELETE FROM invalid_product_usage_records`

	_, err := urd.session.ExecContext(ctx, deleteProductUsageRecords)

	if err != nil {
		logger.Error(err, "error deleting product_usage_records")
		return err
	}

	_, err = urd.session.ExecContext(ctx, deleteInvalidProductUsageRecords)

	if err != nil {
		logger.Error(err, "error deleting invalid_product_usage_records")
		return err
	}

	return nil
}
