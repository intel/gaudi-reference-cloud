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
	insertBucketUserQuery = `
		insert into object_user 
			(resource_id, cloud_account_id, name, value) 
		values ($1, $2, $3, $4)
	`

	updateBucketUserQuery = `
		update object_user
		set    resource_version = nextval('object_user_resource_version_seq'),
		value = $1  
		where cloud_account_id = $2
		and name = $3
		and deleted_timestamp = $4
	`

	getBucketUserById = `
		select value
		from object_user
		where cloud_account_id = $1 
		and resource_id = $2
		and deleted_timestamp = $3
	`

	getBucketUserByName = `
		select value
		from object_user
		where cloud_account_id = $1 
		and name = $2
		and deleted_timestamp = $3
	`
	getAllBucketUsersByCloudAccount = `
		select value
		from object_user
		where cloud_account_id = $1 
		and deleted_timestamp = $2
	`

	updateBucketUserForDeletion = `
		update object_user
		set    resource_version = nextval('object_user_resource_version_seq'),
		value = $1, 
		deleted_timestamp = $2
		where cloud_account_id = $3
		and name = $4
		and deleted_timestamp = $5
	`

	getBucketPrincipal = `
    select value
    from object_user
    where cloud_account_id = $1
    and deleted_timestamp = $2
    LIMIT 1`
)

func StoreBucketUserRequest(ctx context.Context, tx *sql.Tx, bucketUserPrivate *pb.ObjectUserPrivate) error {
	logger := log.FromContext(ctx).WithName("StoreBucketUserRequest()").
		WithValues("userId", bucketUserPrivate.Metadata.UserId, logkeys.CloudAccountId, bucketUserPrivate.Metadata.CloudAccountId)
	logger.Info("begin bucket user record stored insertion")

	// remove credentials before storing
	jsonVal, err := json.MarshalIndent(bucketUserPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		insertBucketUserQuery,
		bucketUserPrivate.Metadata.UserId,
		bucketUserPrivate.Metadata.CloudAccountId,
		bucketUserPrivate.Metadata.Name,
		string(jsonVal))
	if err != nil {
		return err
	}
	logger.Info("bucket user record stored successfully")

	return nil
}

func UpdateBucketUserPolicy(ctx context.Context, tx *sql.Tx, bucketUserPrivate *pb.ObjectUserPrivate, deletionTs string) error {
	logger := log.FromContext(ctx).WithName("UpdateBucketUserRequest").
		WithValues("userId", bucketUserPrivate.Metadata.UserId, logkeys.CloudAccountId, bucketUserPrivate.Metadata.CloudAccountId)

	logger.Info("begin bucket user record update ")

	// remove credentials before storing
	jsonVal, err := json.MarshalIndent(bucketUserPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		updateBucketUserQuery,
		string(jsonVal),
		bucketUserPrivate.Metadata.CloudAccountId,
		bucketUserPrivate.Metadata.Name,
		deletionTs,
	)
	if err != nil {
		return err
	}
	logger.Info("bucket user record updated successfully")

	return nil
}

func GetBucketUserById(ctx context.Context, tx *sql.Tx, cloudaccountId, resourceId, deletionTs string) (*pb.ObjectUserPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketUserById")

	logger.Info("begin bucket user record retrieve", logkeys.CloudAccountId, cloudaccountId, logkeys.ResourceId, resourceId)
	userPrivate, err := readBucketUserById(ctx, tx, cloudaccountId, resourceId, deletionTs)
	if err != nil {
		return nil, err
	}

	return userPrivate, nil
}

func GetBucketUserByName(ctx context.Context, tx *sql.Tx, cloudaccountId, name, deletionTs string) (*pb.ObjectUserPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketUserByName")

	logger.Info("begin bucket user record retrieve", logkeys.CloudAccountId, cloudaccountId, logkeys.Name, name)
	userPrivate, err := readBucketUserByName(ctx, tx, cloudaccountId, name, deletionTs)
	if err != nil {
		return nil, err
	}
	return userPrivate, nil
}

func readBucketUserByName(ctx context.Context, tx *sql.Tx, cloudaccountId, userName, deletionTs string) (*pb.ObjectUserPrivate, error) {
	logger := log.FromContext(ctx).WithName("readBucketUserByName")
	dataBuf := []byte{}
	userPrivate := pb.ObjectUserPrivate{}
	row := tx.QueryRowContext(ctx, getBucketUserByName, cloudaccountId, userName, deletionTs)
	switch err := row.Scan(&dataBuf); err {
	case sql.ErrNoRows:
		logger.Info("no records found ", logkeys.UserName, userName)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		err := json.Unmarshal([]byte(dataBuf), &userPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket record search failed")
		}
	default:
		logger.Error(err, "error searching bucket user record in db")
		return nil, status.Errorf(codes.Internal, "bucket record find failed")
	}

	return &userPrivate, nil
}

func SearchPrincipalClusterInfo(ctx context.Context, tx *sql.Tx, cloudaccountId, deletionTs string) (*pb.ObjectCluster, error) {
	logger := log.FromContext(ctx).WithName("SearchPrincipalClusterInfo")
	row := tx.QueryRowContext(ctx, getBucketPrincipal, cloudaccountId, deletionTs)

	dataBuf := []byte{}
	userPrivate := pb.ObjectUserPrivate{}
	if err := row.Scan(&dataBuf); err != nil {
		if err == sql.ErrNoRows {
			logger.Info("no records found ", logkeys.CloudAccountId, cloudaccountId)
			return nil, nil
		}
		logger.Error(err, "error scanning row")
		return nil, status.Errorf(codes.Internal, "bucket principal search failed")
	}

	if err := json.Unmarshal([]byte(dataBuf), &userPrivate); err != nil {
		logger.Error(err, "Error Unmarshalling JSON")
		return nil, status.Errorf(codes.Internal, "bucket principal search failed")
	}

	return userPrivate.Status.Principal.Cluster, nil
}

func readBucketUserById(ctx context.Context, tx *sql.Tx, cloudaccountId, userId, deletionTs string) (*pb.ObjectUserPrivate, error) {
	logger := log.FromContext(ctx).WithName("readBucketUserById")
	dataBuf := []byte{}
	userPrivate := pb.ObjectUserPrivate{}
	row := tx.QueryRowContext(ctx, getBucketUserById, cloudaccountId, userId, deletionTs)
	switch err := row.Scan(&dataBuf); err {
	case sql.ErrNoRows:
		logger.Info("no records found", logkeys.BucketUserId, userId)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		err := json.Unmarshal([]byte(dataBuf), &userPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket record search failed")
		}
	default:
		logger.Error(err, "error searching bucket user record in db")
		return nil, status.Errorf(codes.Internal, "bucket record find failed")
	}

	return &userPrivate, nil
}

func GetBucketUsersByCloudaccountId(ctx context.Context, tx *sql.Tx, cloudaccountId, deletionTs string) ([]*pb.ObjectUserPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketUsersByCloudaccountId").WithValues(logkeys.CloudAccountId, cloudaccountId)
	logger.Info("begin bucket user record retrieve")

	rows, err := tx.QueryContext(ctx, getAllBucketUsersByCloudAccount, cloudaccountId, deletionTs)
	if err != nil {
		logger.Error(err, "error searching bucket user record in db")
		return nil, status.Errorf(codes.Internal, "bucket user record search failed")
	}
	defer rows.Close()
	resp := []*pb.ObjectUserPrivate{}

	for rows.Next() {
		dataBuf := []byte{}
		userPrivate := pb.ObjectUserPrivate{}
		if err := rows.Scan(&dataBuf); err != nil {
			// FIXME: Handle this failure
			logger.Info("error reading result row, continue...", logkeys.Error, err)
		}
		err := json.Unmarshal([]byte(dataBuf), &userPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket user record search failed")
		}
		resp = append(resp, &userPrivate)
	}
	logger.Info("bucket user record retrieve completed successfully")
	return resp, nil
}

func UpdateBucketUserForDeletion(ctx context.Context, tx *sql.Tx, in *pb.ObjectUserMetadataRef) error {
	logger := log.FromContext(ctx).WithName("UpdateBucketUserForDeletion")
	logger.Info("begin bucket user record update for deletion", logkeys.CloudAccountId, in.CloudAccountId, logkeys.NameOrId, in.NameOrId)

	var err error
	userPrivate := &pb.ObjectUserPrivate{}
	if in.GetUserName() != "" {
		userPrivate, err = readBucketUserByName(ctx, tx, in.CloudAccountId, in.GetUserName(), timestampInfinityStr)
		if err != nil {
			logger.Info("failed to update bucket user record for deletion")
			return err
		}
	} else if in.GetUserId() != "" {
		userPrivate, err = readBucketUserById(ctx, tx, in.CloudAccountId, in.GetUserId(), timestampInfinityStr)
		if err != nil {
			logger.Info("failed to update bucket user record for deletion")
			return err
		}
	}

	userPrivate.Metadata.DeleteTimestamp = timestamppb.New(time.Now())
	userPrivate.Status.Phase = pb.ObjectUserPhase_ObjectUserDeleted

	return updateBucketUserForDelete(ctx, tx, userPrivate)
}

func updateBucketUserForDelete(ctx context.Context, tx *sql.Tx, userPrivate *pb.ObjectUserPrivate) error {
	logger := log.FromContext(ctx).WithName("updateBucketUserForDelete")
	jsonVal, err := json.MarshalIndent(userPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	logger.Info("updating object user  bucket state", logkeys.Query, updateBucketUserForDeletion)
	result, err := tx.ExecContext(ctx,
		updateBucketUserForDeletion,
		string(jsonVal),
		time.Now(),
		userPrivate.Metadata.CloudAccountId,
		userPrivate.Metadata.Name,
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
