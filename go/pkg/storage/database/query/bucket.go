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
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	insertBucketQuery = `
		insert into bucket 
			(resource_id, cloud_account_id, name, value) 
		values ($1, $2, $3, $4)
	`

	getBucketById = `
		select value
		from bucket
		where cloud_account_id = $1 
		and resource_id = $2
		and deleted_timestamp = $3
	`

	getBucketByName = `
		select value
		from bucket
		where cloud_account_id = $1 
		and name = $2
		and deleted_timestamp = $3
	`
	getAllBucketsByCloudAccount = `
		select value
		from bucket
		where cloud_account_id = $1 
		and deleted_timestamp = $2
	`

	updateBucketForDeletion = `
		update bucket
		set    resource_version = nextval('bucket_resource_version_seq'),
		value = $1
		where cloud_account_id = $2
		and name = $3
		and deleted_timestamp = $4
	`

	updateBucketDeletionTS = `
		update bucket
		set    resource_version = nextval('bucket_resource_version_seq'),
		deleted_timestamp = $1
		where cloud_account_id = $2
		and name = $3
		and deleted_timestamp = $4
	`

	getAllBucketsRequests = `
		select resource_version, value
		from bucket
		where deleted_timestamp = $1
	`

	updateBucketStatus = `
		update bucket
		set    resource_version = nextval('bucket_resource_version_seq'),
		value = $1
		where cloud_account_id = $2
		and name = $3 
		and deleted_timestamp = $4
	`

	getActiveBucket = `
    select value
    from bucket
    where cloud_account_id = $1
    and deleted_timestamp = $2
    LIMIT 1
	`
)

func StoreBucketRequest(ctx context.Context, tx *sql.Tx, bucketPrivate *pb.ObjectBucketPrivate) error {
	logger := log.FromContext(ctx).WithName("StoreBucketRequest")

	logger.Info("begin bucket record stored insertion", logkeys.ResourceId, bucketPrivate.Metadata.ResourceId,
		logkeys.CloudAccountId, bucketPrivate.Metadata.CloudAccountId)
	jsonVal, err := json.MarshalIndent(bucketPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		insertBucketQuery,
		bucketPrivate.Metadata.ResourceId,
		bucketPrivate.Metadata.CloudAccountId,
		bucketPrivate.Metadata.Name,
		string(jsonVal))
	if err != nil {
		return err
	}
	logger.Info("bucket record stored successfully", logkeys.ResourceId, bucketPrivate.Metadata.ResourceId,
		logkeys.CloudAccountId, bucketPrivate.Metadata.CloudAccountId)

	return nil
}

func GetBucketById(ctx context.Context, tx *sql.Tx, cloudaccountId, resourceId, deletionTs string) (*pb.ObjectBucketPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketByResourceId")

	logger.Info("begin bucket record retrieve",
		logkeys.CloudAccountId, cloudaccountId, logkeys.ResourceId, resourceId)
	bucketPrivate, err := readBucketById(ctx, tx, cloudaccountId, resourceId, deletionTs)
	if err != nil {
		return nil, err
	}

	return bucketPrivate, nil
}

func GetBucketByName(ctx context.Context, tx *sql.Tx, cloudaccountId, name, deletionTs string) (*pb.ObjectBucketPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketByName")

	logger.Info("begin bucket record retrieve",
		logkeys.CloudAccountId, cloudaccountId, logkeys.Name, name)
	bucketPrivate, err := readBucketByName(ctx, tx, cloudaccountId, name, deletionTs)
	if err != nil {
		return nil, err
	}
	return bucketPrivate, nil
}

func readBucketByName(ctx context.Context, tx *sql.Tx, cloudaccountId, bucketName, deletionTs string) (*pb.ObjectBucketPrivate, error) {
	logger := log.FromContext(ctx).WithName("readBucketByName")
	dataBuf := []byte{}
	bucketPrivate := pb.ObjectBucketPrivate{}
	row := tx.QueryRowContext(ctx, getBucketByName, cloudaccountId, bucketName, deletionTs)
	switch err := row.Scan(&dataBuf); err {
	case sql.ErrNoRows:
		logger.Info("no records found ", logkeys.BucketName, bucketName)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		err := json.Unmarshal([]byte(dataBuf), &bucketPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket record search failed")
		}
	default:
		logger.Error(err, "error searching bucket record in db")
		return nil, status.Errorf(codes.Internal, "bucket record find failed")
	}

	return &bucketPrivate, nil
}

func readBucketById(ctx context.Context, tx *sql.Tx, cloudaccountId, bucketId, deletionTs string) (*pb.ObjectBucketPrivate, error) {
	logger := log.FromContext(ctx).WithName("readBucketById")
	dataBuf := []byte{}
	bucketPrivate := pb.ObjectBucketPrivate{}
	row := tx.QueryRowContext(ctx, getBucketById, cloudaccountId, bucketId, deletionTs)
	switch err := row.Scan(&dataBuf); err {
	case sql.ErrNoRows:
		logger.Info("no records found ", logkeys.BucketId, bucketId)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		err := json.Unmarshal([]byte(dataBuf), &bucketPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket record search failed")
		}
	default:
		logger.Error(err, "error searching bucket record in db")
		return nil, status.Errorf(codes.Internal, "bucket record find failed")
	}

	return &bucketPrivate, nil
}

func GetBucketsByCloudaccountId(ctx context.Context, tx *sql.Tx, cloudaccountId, deletionTs string) ([]*pb.ObjectBucketPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetBucketsByCloudaccountId")

	logger.Info("begin bucket record retrieve", logkeys.CloudAccountId, cloudaccountId)

	rows, err := tx.QueryContext(ctx, getAllBucketsByCloudAccount, cloudaccountId, deletionTs)
	if err != nil {
		logger.Error(err, "error searching bucket record in db")
		return nil, status.Errorf(codes.Internal, "bucket record search failed")
	}
	defer rows.Close()
	resp := []*pb.ObjectBucketPrivate{}

	for rows.Next() {
		dataBuf := []byte{}
		bucketPrivate := pb.ObjectBucketPrivate{}
		if err := rows.Scan(&dataBuf); err != nil {
			// FIXME: Handle this failure
			logger.Info("error reading result row, continue...", logkeys.Error, err)
		}
		err := json.Unmarshal([]byte(dataBuf), &bucketPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "bucket record search failed")
		}
		// fsPublic := convertFilesystemPrivateToPublic(&fsPrivate)
		resp = append(resp, &bucketPrivate)
	}
	logger.Info("bucket record retrieve completed successfully", logkeys.CloudAccountId, cloudaccountId)
	return resp, nil
}

func UpdateBucketForDeletion(ctx context.Context, tx *sql.Tx, in *pb.ObjectBucketMetadataRef) (int64, error) {
	logger := log.FromContext(ctx).WithName("UpdateBucketForDeletion")
	logger.Info("begin bucket record update for deletion", logkeys.CloudAccountId, in.CloudAccountId, logkeys.NameOrId, in.NameOrId)

	var reClaimedSize int64
	bucketPrivate := &pb.ObjectBucketPrivate{}
	var err error
	if in.GetBucketName() != "" {
		bucketPrivate, err = readBucketByName(ctx, tx, in.CloudAccountId, in.GetBucketName(), timestampInfinityStr)
		if err != nil {
			logger.Info("failed to update bucket record for deletion")
			return reClaimedSize, err
		}
	} else if in.GetBucketId() != "" {
		bucketPrivate, err = readBucketById(ctx, tx, in.CloudAccountId, in.GetBucketId(), timestampInfinityStr)
		if err != nil {
			logger.Info("failed to update bucket record for deletion")
			return reClaimedSize, err
		}
	}

	bucketPrivate.Metadata.DeletionTimestamp = timestamppb.New(time.Now())
	size := utils.ParseFileSize(bucketPrivate.Spec.Request.Size)
	if size == -1 {
		logger.Info("invalid input size arguments", "bucketSize", bucketPrivate.Spec.Request.Size)
		return -1, status.Error(codes.InvalidArgument, "invalid bucket size")
	}
	if bucketPrivate.Status.Phase == pb.BucketPhase_BucketProvisioning {
		return size, UpdateBucketDeletionTime(ctx, tx, bucketPrivate.Metadata.CloudAccountId, bucketPrivate.Metadata.Name)
	}
	bucketPrivate.Status.Phase = pb.BucketPhase_BucketDeleting
	return size, updateBucketForDelete(ctx, tx, bucketPrivate)
}

func updateBucketForDelete(ctx context.Context, tx *sql.Tx, fsPrivate *pb.ObjectBucketPrivate) error {
	logger := log.FromContext(ctx).WithName("updateBucketForDelete")
	jsonVal, err := json.MarshalIndent(fsPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	logger.Info("updating object bucket state", logkeys.Query, updateFilesystemForDeletion)
	result, err := tx.ExecContext(ctx,
		updateBucketForDeletion,
		string(jsonVal),
		fsPrivate.Metadata.CloudAccountId,
		fsPrivate.Metadata.Name,
		timestampInfinityStr,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	logger.Info("debug info", logkeys.NumAffectedRows, rowsAffected)

	if rowsAffected < 1 {
		return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
	}
	return nil
}

func UpdateBucketDeletionTime(ctx context.Context, tx *sql.Tx, cloudaccountId, resurceName string) error {
	logger := log.FromContext(ctx).WithName("UpdateBucketDeletionTime")

	logger.Info("updating bucket deletion timestamp")
	result, err := tx.ExecContext(ctx,
		updateBucketDeletionTS,
		time.Now(),
		cloudaccountId,
		resurceName,
		timestampInfinityStr,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected < 1 {
		return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
	}
	return nil
}

func GetBucketsRequests(tx *sql.Tx, resourceVersion, deletionTs string, rs pb.ObjectStorageServicePrivate_SearchBucketPrivateServer) error {
	ctx := rs.Context()
	logger := log.FromContext(ctx).WithName("GetBucketsRequests")

	query := getAllBucketsRequests
	params := []interface{}{}
	params = append(params, deletionTs)
	if resourceVersion != "0" {
		query += "  and resource_version > $2"
		params = append(params, resourceVersion)
	}
	rows, err := tx.QueryContext(ctx, query, params...)
	if err != nil {
		logger.Error(err, "error searching buckets record in db")
		return status.Errorf(codes.Internal, "buckets record search failed")
	}
	defer rows.Close()
	nextVersion := "0"
	for rows.Next() {
		dataBuf := []byte{}
		resp := pb.ObjectBucketSearchPrivateResponse{}
		resp.Bucket = &pb.ObjectBucketPrivate{}
		if err := rows.Scan(&nextVersion, &dataBuf); err != nil {
			// FIXME: Handle this failure
			logger.Info("error reading result row, continue...", logkeys.Error, err)
		}
		err := json.Unmarshal([]byte(dataBuf), &resp.Bucket)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return status.Errorf(codes.Internal, "bucket record search failed")
		}
		resp.Bucket.Metadata.ResourceVersion = nextVersion
		if err := rs.Send(&resp); err != nil {
			logger.Error(err, "error sending bucket record")
			return status.Errorf(codes.Internal, "bucket record search failed")
		}
	}
	return nil
}

func SearchBucketClusterInfo(ctx context.Context, tx *sql.Tx, cloudaccountId, deletionTs string) (*pb.AssignedCluster, error) {
	logger := log.FromContext(ctx).WithName("SearchBucketClusterInfo")
	row := tx.QueryRowContext(ctx, getActiveBucket, cloudaccountId, deletionTs)

	dataBuf := []byte{}
	bucketPrivate := pb.ObjectBucketPrivate{}
	if err := row.Scan(&dataBuf); err != nil {
		if err == sql.ErrNoRows {
			logger.Info("no records found ", logkeys.CloudAccountId, cloudaccountId)
			return nil, nil
		}
		logger.Error(err, "error scanning row")
		return nil, status.Errorf(codes.Internal, "bucket principal search failed")
	}

	if err := json.Unmarshal([]byte(dataBuf), &bucketPrivate); err != nil {
		logger.Error(err, "Error Unmarshalling JSON")
		return nil, status.Errorf(codes.Internal, "bucket principal search failed")
	}

	return bucketPrivate.Spec.Schedule.Cluster, nil
}

func UpdateBucketState(ctx context.Context, tx *sql.Tx, in *pb.ObjectBucketStatusUpdateRequest) error {
	logger := log.FromContext(ctx).WithName("UpdateBucketState")
	logger.Info("updating bucket state")

	bucketPrivate, err := readBucketByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.ResourceId, timestampInfinityStr)
	if err != nil {
		logger.Info("failed to update bucket record for deletion")
		return err
	}
	// When bucket resource is deleted, and operator takes long to delete the state,
	// the replicator replicates the state as "ready" from "deleting".
	// Following check will avoid updating state to "ready" after "deleting"

	// TODO: Verify this in the CRD status, why the state is not getting
	// set to deleting.
	if in.Status.Phase == pb.BucketPhase_BucketReady && bucketPrivate.Status.Phase == pb.BucketPhase_BucketDeleting {
		logger.Info("skipping bucket state update from `deleting` -> `ready'")
		return nil
	}

	bucketPrivate.Status.Message = in.Status.Message
	bucketPrivate.Status.Phase = in.Status.Phase

	jsonVal, err := json.MarshalIndent(bucketPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	result, err := tx.ExecContext(ctx,
		updateBucketStatus,
		jsonVal,
		in.Metadata.CloudAccountId,
		in.Metadata.ResourceId,
		timestampInfinityStr,
	)
	if err != nil {
		logger.Error(err, "error updating bucket state")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected < 1 {
		return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
	}
	return nil
}

func GetBucketAccessUsers(ctx context.Context, tx *sql.Tx, cloudaccount, bucketId string) ([]pb.BucketUserAccess, error) {
	logger := log.FromContext(ctx).WithName("GetBucketAccessUsers")
	logger.Info("get bucket user access")

	return nil, nil
}

func UpdateSubnetBucketRequest(ctx context.Context, tx *sql.Tx, bucketPrivate *pb.ObjectBucketPrivate) error {
	logger := log.FromContext(ctx).WithName("UpdateSubnetBucketRequest").
		WithValues(logkeys.ResourceId, bucketPrivate.Metadata.ResourceId, logkeys.CloudAccountId, bucketPrivate.Metadata.CloudAccountId)
	logger.Info("begin bucket record stored update")
	jsonVal, err := json.MarshalIndent(bucketPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}
	_, err = tx.ExecContext(ctx,
		updateBucketStatus,
		jsonVal,
		bucketPrivate.Metadata.CloudAccountId,
		bucketPrivate.Metadata.Name,
		timestampInfinityStr,
	)
	if err != nil {
		return err
	}
	logger.Info("bucket record updated successfully")
	return nil
}
