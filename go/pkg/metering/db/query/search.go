// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	FixedPaginationSize = 2
	findPreviousQuery   = `
		SELECT DISTINCT ON (resource_id)
		id, transaction_id, resource_id, cloud_account_id, timestamp, properties, reported 
		FROM usage_report
		WHERE resource_id = $1 AND id < $2
		ORDER BY resource_id, timestamp DESC;
	`

	searchQueryBase = `
		SELECT id, transaction_id, resource_id, cloud_account_id, timestamp, properties, reported
		FROM usage_report
	`

	searchInvalidQueryBase = `
		SELECT id, record_id, transaction_id, resource_id, cloud_account_id, region, timestamp, invalidity_reason, properties
		FROM invalidated_metering_records
	`
)

func FindPreviousRecord(ctx context.Context, dbconn *sql.DB, resourceId string, id int64) (*v1.Usage, error) {
	log := log.FromContext(ctx).WithName("FindPreviousRecord()")
	res := v1.Usage{}
	row := dbconn.QueryRowContext(ctx,
		findPreviousQuery,
		resourceId, id)
	ts := time.Time{}
	props := []byte{}

	switch err := row.Scan(&res.Id, &res.TransactionId, &res.ResourceId,
		&res.CloudAccountId, &ts, &props,
		&res.Reported); err {
	case sql.ErrNoRows:
		log.Info("no records found ", " resourceID ", resourceId)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		res.Timestamp = timestamppb.New(ts)
		err := json.Unmarshal([]byte(props), &res.Properties)
		if err != nil {
			log.Error(err, "Error Unmarshalling JSON:")
		}
	default:
		log.Error(err, "error searching metering record in db")
		return nil, status.Errorf(codes.Internal, "metering record find failed")
	}

	return &res, nil
}

func SearchInvalid(rs v1.MeteringService_SearchInvalidServer, dbconn *sql.DB, filter *v1.InvalidMeteringRecordFilter) error {

	log := log.FromContext(rs.Context()).WithName("SearchInvalid()")
	var q strings.Builder

	_, err := q.WriteString(searchInvalidQueryBase)
	if err != nil {
		log.Error(err, "failed to write search invalid query base")
		return err
	}
	// following clause is essentially a no-op
	// used to complete the below concatanated where clauses
	if _, err := q.WriteString(" WHERE 1=1"); err != nil {
		log.Error(err, "failed to write search invalid query base")
	}

	params := []interface{}{}

	if filter.Id != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND id=$%d ", len(params)+1))
		params = append(params, filter.GetId())
		if err != nil {
			log.Error(err, "failed to append id to search query for searching invalid metering records")

		}
	}
	if filter.RecordId != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND record_id=$%d ", len(params)+1))
		params = append(params, filter.GetRecordId())
		if err != nil {
			log.Error(err, "failed to append record id to search query for searching invalid metering records")

		}
	}
	if filter.ResourceId != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND resource_id=$%d ", len(params)+1))
		params = append(params, filter.GetResourceId())
		if err != nil {
			log.Error(err, "failed to append resource id to search query for searching invalid metering records")
		}
	}
	if filter.CloudAccountId != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND cloud_account_id=$%d ", len(params)+1))
		params = append(params, filter.GetCloudAccountId())
		if err != nil {
			log.Error(err, "failed to append cloud account id to search query for searching invalid metering records")
		}
	}
	if filter.TransactionId != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND transaction_id=$%d ", len(params)+1))
		params = append(params, filter.GetTransactionId())
		if err != nil {
			log.Error(err, "failed to append transaction id to search query for searching invalid metering records")
		}
	}
	if filter.Region != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND region=$%d ", len(params)+1))
		params = append(params, filter.GetRegion())
		if err != nil {
			log.Error(err, "failed to append region to search query for searching invalid metering records")
		}
	}
	if filter.StartTime != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND timestamp>=$%d ", len(params)+1))
		params = append(params, (filter.GetStartTime()).AsTime())
		if err != nil {
			log.Error(err, "failed to append start time to search query for searching invalid metering records")
		}
	}
	if filter.EndTime != nil {
		_, err := q.WriteString(fmt.Sprintf("   AND timestamp<=$%d ", len(params)+1))
		params = append(params, (filter.GetEndTime()).AsTime())
		if err != nil {
			log.Error(err, "failed to append end time to search query for searching invalid metering records")
		}
	}
	_, err = q.WriteString(" ORDER BY id")
	if err != nil {
		log.Error(err, "failed to append order by to search query for searching invalid metering records")
	}

	rows, err := dbconn.QueryContext(rs.Context(), q.String(), params...)
	if err != nil {
		log.Error(err, "error searching invalid metering record in db")
		return status.Errorf(codes.Internal, "invalid metering record find failed")
	}

	defer rows.Close()
	for rows.Next() {
		u := v1.InvalidMeteringRecord{}
		props := []byte{}
		// todo: add the read for invalidity reasons
		ts := time.Time{}
		if err := rows.Scan(&u.Id, &u.RecordId, &u.TransactionId, &u.ResourceId, &u.CloudAccountId,
			&u.Region, &ts, &u.MeteringRecordInvalidityReason, &props); err != nil {
			log.Error(err, "error reading result row, continue...")
		}
		u.Timestamp = timestamppb.New(ts)
		err := json.Unmarshal([]byte(props), &u.Properties)
		if err != nil {
			log.Error(err, "Error Unmarshalling JSON")
		}
		if err := rs.Send(&u); err != nil {
			log.Error(err, "error sending usage record")
		}
	}
	return nil
}

func Search(rs v1.MeteringService_SearchServer, dbconn *sql.DB, filter *v1.UsageFilter) error {

	log := log.FromContext(rs.Context()).WithName("Search()")
	var q strings.Builder

	_, err := q.WriteString(searchQueryBase)
	if err != nil {
		log.Error(err, "failed to write search query base")
		return err
	}
	// following clause is essentially a no-op
	// used to complete the below concatanated where clauses
	if _, err := q.WriteString(" WHERE 1=1"); err != nil {
		log.Error(err, "failed to write search query base")
	}

	params := []interface{}{}

	if filter.Id != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND id=$%d ", len(params)+1))
		params = append(params, filter.GetId())
		if err != nil {
			log.Error(err, "failed to write search query base")

		}
	}
	if filter.ResourceId != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND resource_id=$%d ", len(params)+1))
		params = append(params, filter.GetResourceId())
		if err != nil {
			log.Error(err, "failed to write search query base")
		}
	}
	if filter.CloudAccountId != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND cloud_account_id=$%d ", len(params)+1))
		params = append(params, filter.GetCloudAccountId())
		if err != nil {
			log.Error(err, "failed to write search query base")
		}
	}
	if filter.TransactionId != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND transaction_id=$%d ", len(params)+1))
		params = append(params, filter.GetTransactionId())
		if err != nil {
			log.Error(err, "failed to write search query base")
		}
	}
	if filter.StartTime != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND timestamp>=$%d ", len(params)+1))
		params = append(params, (filter.GetStartTime()).AsTime())
		if err != nil {
			log.Error(err, "failed to write search query base")
		}
	}
	if filter.EndTime != nil {
		_, err := q.WriteString(fmt.Sprintf("   AND timestamp<=$%d ", len(params)+1))
		params = append(params, (filter.GetEndTime()).AsTime())
		if err != nil {
			log.Error(err, "failed to write search query base")
		}
	}
	if filter.Reported != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND reported=$%d ", len(params)+1))
		params = append(params, (filter.GetReported()))
		if err != nil {
			log.Error(err, "failed to write search query base")
		}
	}

	_, err = q.WriteString(" ORDER BY id")
	if err != nil {
		log.Error(err, "failed to write search query base")
	}

	rows, err := dbconn.QueryContext(rs.Context(), q.String(), params...)
	if err != nil {
		log.Error(err, "error searching metering record in db")
		return status.Errorf(codes.Internal, "metering record find failed")
	}

	defer rows.Close()
	for rows.Next() {
		u := v1.Usage{}
		props := []byte{}
		ts := time.Time{}
		if err := rows.Scan(&u.Id, &u.TransactionId, &u.ResourceId, &u.CloudAccountId, &ts, &props, &u.Reported); err != nil {
			log.Error(err, "error reading result row, continue...")
		}
		u.Timestamp = timestamppb.New(ts)
		err := json.Unmarshal([]byte(props), &u.Properties)
		if err != nil {
			log.Error(err, "Error Unmarshalling JSON")
		}
		if err := rs.Send(&u); err != nil {
			log.Error(err, "error sending usage record")
		}
	}
	return nil
}

func GetResourceMeteringRecordsMap(ctx context.Context, dbconn *sql.DB, filter *v1.MeteringFilter) (map[string]*v1.ResourceMeteringRecords, error) {
	log := log.FromContext(ctx).WithName("GetResourceMeteringRecordsMap")
	var q strings.Builder

	_, err := q.WriteString(searchQueryBase)
	if err != nil {
		log.Error(err, "failed to build query for searching resource metering records")
		return nil, err
	}

	if _, err := q.WriteString(" WHERE 1=1"); err != nil {
		log.Error(err, "failed to build query for searching resource metering records")
		return nil, err
	}

	params := []interface{}{}

	if filter.CloudAccountId != nil {
		_, err := q.WriteString(fmt.Sprintf(" AND cloud_account_id=$%d ", len(params)+1))
		params = append(params, filter.CloudAccountId)
		if err != nil {
			log.Error(err, "failed to append cloud account id to search query for searching resource metering records")
			return nil, err
		}
	}

	if filter.StartTime != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND timestamp>=$%d ", len(params)+1))
		params = append(params, (filter.GetStartTime()).AsTime())
		if err != nil {
			log.Error(err, "failed to append start time to search query for searching resource metering records")
			return nil, err
		}
	}

	if filter.EndTime != nil {
		_, err := q.WriteString(fmt.Sprintf("   AND timestamp<=$%d ", len(params)+1))
		params = append(params, (filter.GetEndTime()).AsTime())
		if err != nil {
			log.Error(err, "failed to append end time to search query for searching resource metering records")
			return nil, err
		}
	}

	if filter.Reported != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND reported=$%d ", len(params)+1))
		params = append(params, (filter.GetReported()))
		if err != nil {
			log.Error(err, "failed to append reported to search query for searching resource metering records")
			return nil, err
		}
	}

	_, err = q.WriteString(" ORDER BY timestamp DESC")
	if err != nil {
		log.Error(err, "failed to append order by time desc to search query for searching resource metering records")
		return nil, err
	}

	rows, err := dbconn.QueryContext(ctx, q.String(), params...)

	if err != nil {
		log.Error(err, "failed to get unreported metering records for cloud acct", "id", filter.CloudAccountId)
		return nil, err
	}

	defer rows.Close()

	resourceMeteringRecordsMap := map[string]*v1.ResourceMeteringRecords{}

	for rows.Next() {
		meteringRecord := &v1.MeteringRecord{}
		props := []byte{}
		ts := time.Time{}
		if err := rows.Scan(&meteringRecord.Id, &meteringRecord.TransactionId, &meteringRecord.ResourceId, &meteringRecord.CloudAccountId, &ts, &props, &meteringRecord.Reported); err != nil {
			log.Error(err, "error reading a metering record")
			continue
		}

		meteringRecord.Timestamp = timestamppb.New(ts)

		err := json.Unmarshal([]byte(props), &meteringRecord.Properties)
		if err != nil {
			log.Error(err, "error unmarshalling json for the row read", "id", meteringRecord.Id)
			continue
		}

		const regionKey = "region"
		if _, ok := meteringRecord.Properties[regionKey]; !ok {
			log.Error(err, "missing region in the metering record", "id", meteringRecord.Id)
			continue
		}

		meteringRecord.Region = meteringRecord.Properties[regionKey]

		if _, ok := resourceMeteringRecordsMap[meteringRecord.ResourceId]; ok {
			log.V(2).Info("appending to metering records for", "resource", meteringRecord.ResourceId)
			resourceMeteringRecords := resourceMeteringRecordsMap[meteringRecord.ResourceId]
			resourceMeteringRecords.MeteringRecords = append(resourceMeteringRecords.MeteringRecords, meteringRecord)
		} else {
			log.V(2).Info("adding metering records for", "resource", meteringRecord.ResourceId)
			meteringRecords := []*v1.MeteringRecord{}
			meteringRecords = append(meteringRecords, meteringRecord)
			resourceMeteringRecordsMap[meteringRecord.ResourceId] = &v1.ResourceMeteringRecords{
				ResourceId:      meteringRecord.ResourceId,
				CloudAccountId:  meteringRecord.CloudAccountId,
				Region:          meteringRecord.Region,
				MeteringRecords: meteringRecords,
			}
		}
	}

	return resourceMeteringRecordsMap, nil
}

/*
Get resource metering records.
This is a internal method that does exactly what a given consumption of the API needs - calculate usages.
One could have multiple such internal methods if needed.
*/
func GetResourceMeteringRecords(ctx context.Context, dbconn *sql.DB, filter *v1.MeteringFilter) (*v1.ResourceMeteringRecordsList, error) {

	log := log.FromContext(ctx).WithName("GetResourceMeteringRecords")

	resourceMeteringRecordsList := &v1.ResourceMeteringRecordsList{}
	resourceMeteringRecordsMap, err := GetResourceMeteringRecordsMap(ctx, dbconn, filter)

	if err != nil {
		log.Error(err, "failed to build resource metering records list")
		return nil, err
	}

	for resourceId, resMeteringRecords := range resourceMeteringRecordsMap {
		resourceMeteringRecords := &v1.ResourceMeteringRecords{
			ResourceId:      resourceId,
			CloudAccountId:  resMeteringRecords.CloudAccountId,
			Region:          resMeteringRecords.Region,
			MeteringRecords: resMeteringRecords.MeteringRecords,
		}
		resourceMeteringRecordsList.ResourceMeteringRecordsList = append(resourceMeteringRecordsList.ResourceMeteringRecordsList, resourceMeteringRecords)
	}

	return resourceMeteringRecordsList, nil
}

func GetResourceMeteringRecordsAsStream(rs v1.MeteringService_SearchResourceMeteringRecordsAsStreamServer,
	dbconn *sql.DB, filter *v1.MeteringFilter) error {

	ctx := rs.Context()
	log := log.FromContext(ctx).WithName("GetResourceMeteringRecordsAsStream")

	resourceMeteringRecordsList := &v1.ResourceMeteringRecordsList{}
	resourceMeteringRecordsMap, err := GetResourceMeteringRecordsMap(ctx, dbconn, filter)

	if err != nil {
		log.Error(err, "failed to build resource metering records list for streaming")
		return nil
	}

	// we could have done db pagination but because of it being time series, doing pagination in code instead.
	countOfResources := 0
	lengthOfResourceMeteringRecordsMap := len(resourceMeteringRecordsMap)
	lengthOfLeftOverResources := lengthOfResourceMeteringRecordsMap
	for resourceId, resMeteringRecords := range resourceMeteringRecordsMap {

		resourceMeteringRecords := &v1.ResourceMeteringRecords{
			ResourceId:      resourceId,
			CloudAccountId:  resMeteringRecords.CloudAccountId,
			Region:          resMeteringRecords.Region,
			MeteringRecords: resMeteringRecords.MeteringRecords,
		}

		resourceMeteringRecordsList.ResourceMeteringRecordsList = append(resourceMeteringRecordsList.ResourceMeteringRecordsList,
			resourceMeteringRecords)
		countOfResources++

		if (countOfResources == FixedPaginationSize) || (countOfResources == lengthOfLeftOverResources) {
			if err := rs.Send(resourceMeteringRecordsList); err != nil {
				log.Error(err, "error sending resource metering records list")
				return err
			}
			// reinitialize the resource.
			resourceMeteringRecordsList = &v1.ResourceMeteringRecordsList{}
			countOfResources = 0
			if lengthOfLeftOverResources > FixedPaginationSize {
				lengthOfLeftOverResources = lengthOfLeftOverResources - FixedPaginationSize
			}
		}

	}

	return nil
}

func MeteringRecordAvailable(ctx context.Context, dbconn *sql.DB, cloudAccountId string, startTime *timestamppb.Timestamp) (*v1.MeteringRecord, error) {
	log := log.FromContext(ctx).WithName("MeteringRecordAvailable")
	log.V(9).Info("BEGIN")
	defer log.V(9).Info("END")

	var q strings.Builder

	_, err := q.WriteString(searchQueryBase)
	if err != nil {
		log.Error(err, "failed to build query for searching available resource metering records")
		return nil, err
	}

	if _, err := q.WriteString(" WHERE 1=1"); err != nil {
		log.Error(err, "failed to build query for searching available resource metering records")
		return nil, err
	}

	params := []interface{}{}

	if cloudAccountId != "" {
		_, err := q.WriteString(fmt.Sprintf(" AND cloud_account_id=$%d ", len(params)+1))
		params = append(params, cloudAccountId)
		if err != nil {
			log.Error(err, "failed to append cloud account id to search query for searching available resource metering records")
			return nil, err
		}
	}

	if startTime != nil {
		_, err := q.WriteString(fmt.Sprintf("  AND timestamp>=$%d ", len(params)+1))
		params = append(params, (startTime).AsTime())
		if err != nil {
			log.Error(err, "failed to append start time to search query for searching available resource metering records")
			return nil, err
		}
	}

	_, err = q.WriteString(" ORDER BY timestamp DESC LIMIT 1")
	if err != nil {
		log.Error(err, "failed to append order by time desc to search query for searching available resource metering records")
		return nil, err
	}

	rows, err := dbconn.QueryContext(ctx, q.String(), params...)

	if err != nil {
		log.Error(err, "failed to get metering records for cloud acct", "id", cloudAccountId)
		return nil, err
	}

	defer rows.Close()

	meteringRecord := &v1.MeteringRecord{}
	for rows.Next() {

		props := []byte{}
		ts := time.Time{}
		if err := rows.Scan(&meteringRecord.Id, &meteringRecord.TransactionId, &meteringRecord.ResourceId, &meteringRecord.CloudAccountId, &ts, &props, &meteringRecord.Reported); err != nil {
			log.Error(err, "error reading a metering record")
			continue
		}

		meteringRecord.Timestamp = timestamppb.New(ts)
		err := json.Unmarshal([]byte(props), &meteringRecord.Properties)
		if err != nil {
			log.Error(err, "error unmarshalling json for the row read", "id", meteringRecord.Id)
			continue
		}

		log.V(2).Info("adding available metering record for", "resource", meteringRecord.ResourceId)
	}

	return meteringRecord, nil
}
