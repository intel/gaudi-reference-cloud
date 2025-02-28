package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	insertBucketLifecycleRuleQuery = `
		insert into bucket_lifecycle_rule 
			(resource_id, cloud_account_id, name, bucket_id, value) 
		values ($1, $2, $3, $4, $5)
	`

	updateBucketLifecycleRuleQuery = `
		update bucket_lifecycle_rule
		set    resource_version = nextval('bucket_lifecycle_rule_seq'),
		value = $1
		where cloud_account_id = $2
		and name = $3
		and deleted_timestamp = $4
	`

	getBucketLifecycleRuleById = `
		select value
		from bucket_lifecycle_rule
		where cloud_account_id = $1 
		and resource_id = $2
		and bucket_id = $3
		and deleted_timestamp = $4
	`
	getBucketLifecycleRuleByName = `
		select value
		from bucket_lifecycle_rule
		where cloud_account_id = $1 
		and name = $2
		and bucket_id = $3
		and deleted_timestamp = $4
	`
	getAllBucketLifecycleRulesByCloudAccount = `
		select value
		from bucket_lifecycle_rule
		where cloud_account_id = $1
		and bucket_id = $2 
		and deleted_timestamp = $3
	`

	updateBucketLifecycleRuleForDeletion = `
		update bucket_lifecycle_rule
		set    resource_version = nextval('bucket_lifecycle_rule_seq'),
		value = $1, 
		deleted_timestamp = $2
		where cloud_account_id = $3
		and name = $4
		and bucket_id = $5 
		and deleted_timestamp = $6
	`
)

func InsertBucketLifecycleRequest(ctx context.Context, tx *sql.Tx, lfRule *pb.BucketLifecycleRulePrivate) error {
	logger := log.FromContext(ctx).WithName("InsertBucketLifecycleRequest").
		WithValues(logkeys.CloudAccountId, lfRule.Metadata.CloudAccountId, "ruleName", lfRule.Metadata.RuleName)

	logger.Info("begin bucket lifecycle rule record insertion")

	jsonVal, err := json.MarshalIndent(lfRule, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		insertBucketLifecycleRuleQuery,
		lfRule.Metadata.ResourceId,
		lfRule.Metadata.CloudAccountId,
		lfRule.Metadata.RuleName,
		lfRule.Metadata.BucketId,
		string(jsonVal))

	if err != nil {
		return err
	}
	logger.Info("bucket lifecycle rule record stored successfully")

	return nil
}

func UpdateBucketLifecycleRequest(ctx context.Context, tx *sql.Tx, lfRule *pb.BucketLifecycleRulePrivate) error {
	logger := log.FromContext(ctx).WithName("UpdateBucketLifecycleRequest").
		WithValues(logkeys.CloudAccountId, lfRule.Metadata.CloudAccountId, "ruleName", lfRule.Metadata.RuleName)

	logger.Info("begin bucket lifecycle rule record stored update")

	jsonVal, err := json.MarshalIndent(lfRule, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		updateBucketLifecycleRuleQuery,
		string(jsonVal),
		lfRule.Metadata.CloudAccountId,
		lfRule.Metadata.RuleName,
		timestampInfinityStr,
	)

	if err != nil {
		return err
	}
	logger.Info("bucket lifecycle rule record updated successfully")

	return nil
}

func GetBucketLifecycleRuleById(ctx context.Context, tx *sql.Tx, cloudaccountId, resourceId, bucketId string) (*pb.BucketLifecycleRulePrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketLifecycleRuleById").
		WithValues(logkeys.CloudAccountId, cloudaccountId, logkeys.ResourceId, resourceId)

	logger.Info("begin bucket lifecycle rule record retrieve")
	rulePrivate, err := readLifecycleRuleById(ctx, tx, cloudaccountId, resourceId, bucketId)
	if err != nil {
		return nil, err
	}

	return rulePrivate, nil
}

func readLifecycleRuleById(ctx context.Context, tx *sql.Tx, cloudaccountId, ruleId, bucketId string) (*pb.BucketLifecycleRulePrivate, error) {
	logger := log.FromContext(ctx).WithName("readLifecycleRuleById")
	dataBuf := []byte{}
	rulePrivate := pb.BucketLifecycleRulePrivate{}
	row := tx.QueryRowContext(ctx, getBucketLifecycleRuleById, cloudaccountId, ruleId, bucketId, timestampInfinityStr)
	switch err := row.Scan(&dataBuf); err {
	case sql.ErrNoRows:
		logger.Info("no records found", logkeys.ResourceId, ruleId, logkeys.BucketId, bucketId)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		err := json.Unmarshal([]byte(dataBuf), &rulePrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket record search failed")
		}
	default:
		logger.Error(err, "error searching bucket lifecycle rule record in db")
		return nil, status.Errorf(codes.Internal, "bucket lifecycle rule record find failed")
	}

	return &rulePrivate, nil
}

func GetBucketLifecycleRuleByName(ctx context.Context, tx *sql.Tx, cloudaccountId, ruleName, bucketId string) (*pb.BucketLifecycleRulePrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketLifecycleRuleByName").
		WithValues(logkeys.CloudAccountId, cloudaccountId, "resourceName", ruleName)
	logger.Info("begin bucket lifecycle rule record retrieve")
	rulePrivate, err := readLifecycleRuleByName(ctx, tx, cloudaccountId, ruleName, bucketId)
	if err != nil {
		return nil, err
	}

	return rulePrivate, nil
}

func readLifecycleRuleByName(ctx context.Context, tx *sql.Tx, cloudaccountId, ruleName, bucketId string) (*pb.BucketLifecycleRulePrivate, error) {
	logger := log.FromContext(ctx).WithName("readLifecycleRuleByName")
	dataBuf := []byte{}
	rulePrivate := pb.BucketLifecycleRulePrivate{}
	row := tx.QueryRowContext(ctx, getBucketLifecycleRuleByName, cloudaccountId, ruleName, bucketId, timestampInfinityStr)
	switch err := row.Scan(&dataBuf); err {
	case sql.ErrNoRows:
		logger.Info("no records found", logkeys.ResourceId, ruleName, logkeys.BucketId, bucketId)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		err := json.Unmarshal([]byte(dataBuf), &rulePrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket record search failed")
		}
	default:
		logger.Error(err, "error searching bucket lifecycle rule record in db")
		return nil, status.Errorf(codes.Internal, "bucket lifecycle rule record find failed")
	}

	return &rulePrivate, nil
}

func GetBucketLifecycleRulesByCloudaccountId(ctx context.Context, tx *sql.Tx, cloudaccountId, bucketId string) ([]*pb.BucketLifecycleRulePrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketLifecycleRulesByCloudaccountId").
		WithValues(logkeys.CloudAccountId, cloudaccountId, logkeys.BucketId, bucketId)
	logger.Info("begin bucket lifecycle rules record retrieve")

	rows, err := tx.QueryContext(ctx, getAllBucketLifecycleRulesByCloudAccount,
		cloudaccountId, bucketId, timestampInfinityStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resp := []*pb.BucketLifecycleRulePrivate{}

	for rows.Next() {
		dataBuf := []byte{}
		rulePrivate := pb.BucketLifecycleRulePrivate{}
		if err := rows.Scan(&dataBuf); err != nil {
			// FIXME: Handle this failure
			logger.Info("error reading result row, continue...", logkeys.Error, err)
		}
		err := json.Unmarshal([]byte(dataBuf), &rulePrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket lifecycle rule record search failed")
		}
		resp = append(resp, &rulePrivate)
	}
	logger.Info("bucket lifecycle rule record retrieve completed successfully")
	return resp, nil
}

func UpdateBucketLifecycleRulesForDeletion(ctx context.Context, tx *sql.Tx, in *pb.BucketLifecycleRuleMetadataRef) error {
	logger := log.FromContext(ctx).WithName("UpdateBucketLifecycleRulesForDeletion").
		WithValues(logkeys.CloudAccountId, in.CloudAccountId, logkeys.BucketId, in.BucketId, logkeys.RuleId, in.RuleId)

	logger.Info("begin bucket lifecycle record update for deletion")
	var err error
	rulePrivate := &pb.BucketLifecycleRulePrivate{}
	rulePrivate, err = readLifecycleRuleById(ctx, tx, in.CloudAccountId, in.RuleId, in.BucketId)
	if err != nil {
		logger.Info("failed to update bucket user record for deletion")
		return err
	}
	rulePrivate.Metadata.DeletionTimestamp = timestamppb.New(time.Now())
	return updateBucketLifecycleRuleForDelete(ctx, tx, rulePrivate)
}

func updateBucketLifecycleRuleForDelete(ctx context.Context, tx *sql.Tx, userPrivate *pb.BucketLifecycleRulePrivate) error {
	logger := log.FromContext(ctx).WithName("updateBucketLifecycleRuleForDelete")
	jsonVal, err := json.MarshalIndent(userPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	logger.Info("updating object lifecycle rule bucket state")
	result, err := tx.ExecContext(ctx,
		updateBucketLifecycleRuleForDeletion,
		string(jsonVal),
		time.Now(),
		userPrivate.Metadata.CloudAccountId,
		userPrivate.Metadata.RuleName,
		userPrivate.Metadata.BucketId,
		timestampInfinityStr,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	logger.Info("debug", logkeys.NumAffectedRows, rowsAffected)

	if rowsAffected < 1 {
		return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
	}
	return nil
}
