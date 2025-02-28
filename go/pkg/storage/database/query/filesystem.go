// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
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
	insertFilesystemQuery = `
		insert into filesystem 
			(resource_id, cloud_account_id, name, value) 
		values ($1, $2, $3, $4)
	`

	updateFilesystemQuery = `
		update filesystem
		set    resource_version = nextval('filesystem_resource_version_seq'),
		value = $1
		where resource_id = $2
	`

	getAllFilesystemsByCloudAccount = `
		select value
		from filesystem
		where cloud_account_id = $1 
		and deleted_timestamp = $2 FOR UPDATE
	`

	getFilesystemByName = `
		select value
		from filesystem
		where cloud_account_id = $1 
		and name = $2
		and deleted_timestamp = $3
	`

	getFilesystemByResourceId = `
		select value
		from filesystem
		where cloud_account_id = $1 
		and resource_id = $2
		and deleted_timestamp = $3
	`

	getAllFSRequests = `
		select resource_version, value
		from filesystem
		where deleted_timestamp = $1
	`

	updateFilesystemForDeletion = `
		update filesystem
		set    resource_version = nextval('filesystem_resource_version_seq'),
		value = $1
		where cloud_account_id = $2
		and name = $3
		and deleted_timestamp = $4
	`
	updateFilesystemDeletionTS = `
		update filesystem
		set    resource_version = nextval('filesystem_resource_version_seq'),
		deleted_timestamp = $1
		where cloud_account_id = $2
		and name = $3
		and deleted_timestamp = $4
	`
	updateFilesystemStatus = `
		update filesystem
		set    resource_version = nextval('filesystem_resource_version_seq'),
		value = $1
		where cloud_account_id = $2
		and name = $3 
		and deleted_timestamp = $4
	`
)

const (
	timestampInfinityStr = "infinity"
)

func StoreFilesystemRequest(ctx context.Context, tx *sql.Tx, fs *pb.FilesystemPrivate) error {
	logger := log.FromContext(ctx).WithName("StoreFilesystemRequest").
		WithValues(logkeys.ResourceId, fs.Metadata.ResourceId, logkeys.CloudAccountId, fs.Metadata.CloudAccountId)
	logger.Info("begin filesystem record stored insertion")

	jsonVal, err := json.MarshalIndent(fs, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		insertFilesystemQuery,
		fs.Metadata.ResourceId,
		fs.Metadata.CloudAccountId,
		fs.Metadata.Name,
		string(jsonVal))
	if err != nil {
		return err
	}

	logger.Info("filesystem record stored successfully")
	return nil
}

func UpdateFilesystemRequest(ctx context.Context, tx *sql.Tx, fs *pb.FilesystemPrivate) error {
	logger := log.FromContext(ctx).WithName("UpdateFilesystemRequest").
		WithValues(logkeys.ResourceId, fs.Metadata.ResourceId, logkeys.CloudAccountId, fs.Metadata.CloudAccountId)
	logger.Info("begin filesystem record stored update")

	jsonVal, err := json.MarshalIndent(fs, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		updateFilesystemQuery,
		string(jsonVal),
		fs.Metadata.ResourceId)
	if err != nil {
		logger.Error(err, "error executing update filesystem in db")
		return status.Errorf(codes.Internal, "filesystem db update failed")
	}
	logger.Info("filesystem record updated successfully")
	return nil
}

func GetFilesystemsByCloudaccountId(ctx context.Context, tx *sql.Tx, cloudaccountId string, fsTypeFilter pb.FilesystemType, deletionTs string) ([]*pb.FilesystemPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetFilesystemsByCloudaccountId")
	logger.Info("begin filesystem record retrieve", logkeys.CloudAccountId, cloudaccountId, logkeys.FilesystemTypeFilter, fsTypeFilter)

	rows, err := tx.QueryContext(ctx, getAllFilesystemsByCloudAccount, cloudaccountId, deletionTs)
	if err != nil {
		logger.Error(err, "error searching filesystem record in db")
		return nil, status.Errorf(codes.Internal, "filesystem record search failed")
	}
	defer rows.Close()
	resp := []*pb.FilesystemPrivate{}

	for rows.Next() {
		dataBuf := []byte{}
		fsPrivate := pb.FilesystemPrivate{}
		if err := rows.Scan(&dataBuf); err != nil {
			// FIXME: Handle this failure
			logger.Info("error reading result row, continue...", logkeys.Error, err)
		}
		err := json.Unmarshal([]byte(dataBuf), &fsPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "filesystem record search failed")
		}
		if fsTypeFilter != pb.FilesystemType_Unspecified &&
			fsPrivate.Spec.FilesystemType != fsTypeFilter {
			logger.Info("skipping filesystem from result")
			continue
		}
		// fsPublic := convertFilesystemPrivateToPublic(&fsPrivate)
		resp = append(resp, &fsPrivate)
	}
	logger.Info("filesystem record retrieve completed successfully", logkeys.CloudAccountId, cloudaccountId)
	return resp, nil
}

func GetFilesystemByResourceId(ctx context.Context, tx *sql.Tx, cloudaccountId, resourceId, deletionTs string) (*pb.FilesystemPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetFilesystemByResourceId")
	logger.Info("begin filesystem record retrieve", logkeys.CloudAccountId, cloudaccountId, logkeys.ResourceId, resourceId)
	fsPrivate, err := readFilesystemByResourceId(ctx, tx, cloudaccountId, resourceId, deletionTs)
	if err != nil {
		return nil, err
	}

	return fsPrivate, nil
	// fsPublic := convertFilesystemPrivateToPublic(fsPrivate)
	// return fsPublic, nil
}

func GetFilesystemByName(ctx context.Context, tx *sql.Tx, cloudaccountId, name, deletionTs string) (*pb.FilesystemPrivate, error) {
	logger := log.FromContext(ctx).WithName("GetFilesystemByName")
	logger.Info("begin filesystem record retrieve", logkeys.CloudAccountId, cloudaccountId, logkeys.Name, name)
	fsPrivate, err := readFilesystemByName(ctx, tx, cloudaccountId, name, deletionTs)
	if err != nil {
		return nil, err
	}
	return fsPrivate, nil
	// fsPublic := convertFilesystemPrivateToPublic(fsPrivate)
	// return fsPublic, nil
}

func GetFilesystemsRequests(tx *sql.Tx, resourceVersion, deletionTs string, rs pb.FilesystemPrivateService_SearchFilesystemRequestsServer) error {
	ctx := rs.Context()
	logger := log.FromContext(ctx).WithName("GetFilesystemsRequests")
	query := getAllFSRequests
	params := []interface{}{}
	params = append(params, deletionTs)
	if resourceVersion != "0" {
		query += "  and resource_version > $2"
		params = append(params, resourceVersion)
	}
	rows, err := tx.QueryContext(ctx, query, params...)
	if err != nil {
		logger.Error(err, "error searching filesystem record in db")
		return status.Errorf(codes.Internal, "filesystem record search failed")
	}
	defer rows.Close()
	nextVersion := "0"
	for rows.Next() {
		dataBuf := []byte{}

		resp := pb.FilesystemRequestResponse{}
		resp.Filesystem = &pb.FilesystemPrivate{}
		if err := rows.Scan(&nextVersion, &dataBuf); err != nil {
			// FIXME: Handle this failure
			logger.Info("error reading result row, continue...", logkeys.Error, err)
		}
		err := json.Unmarshal([]byte(dataBuf), &resp.Filesystem)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return status.Errorf(codes.Internal, "filesystem record search failed")
		}
		resp.Filesystem.Metadata.ResourceVersion = nextVersion
		if err := rs.Send(&resp); err != nil {
			logger.Error(err, "error sending filesystem record")
			return status.Errorf(codes.Internal, "filesystem record search failed")
		}
	}
	return nil
}

func UpdateFilesystemForDeletion(ctx context.Context, tx *sql.Tx, in *pb.FilesystemMetadataReference) (int64, error) {
	logger := log.FromContext(ctx).WithName("UpdateFilesystemForDeletion()")
	logger.Info("begin filesystem record update for deletion", logkeys.CloudAccountId, in.CloudAccountId, logkeys.NameOrId, in.NameOrId)

	var reClaimedSize int64
	fsPrivate := &pb.FilesystemPrivate{}
	var err error
	if in.GetName() != "" {
		fsPrivate, err = readFilesystemByName(ctx, tx, in.CloudAccountId, in.GetName(), timestampInfinityStr)
		if err != nil {
			logger.Info("failed to update filesystem record for deletion")
			return reClaimedSize, err
		}
	} else if in.GetResourceId() != "" {
		fsPrivate, err = readFilesystemByResourceId(ctx, tx, in.CloudAccountId, in.GetResourceId(), timestampInfinityStr)
		if err != nil {
			logger.Info("failed to update filesystem record for deletion")
			return reClaimedSize, err
		}
	}

	fsPrivate.Metadata.DeletionTimestamp = timestamppb.New(time.Now())
	size := utils.ParseFileSize(fsPrivate.Spec.Request.Storage)
	if size == -1 {
		logger.Info("invalid input size arguments", "fsSize", fsPrivate.Spec.Request.Storage)
		return -1, status.Error(codes.InvalidArgument, "invalid storage size")
	}
	if fsPrivate.Status.Phase == pb.FilesystemPhase_FSProvisioning {
		return size, UpdateFilesystemDeletionTime(ctx, tx, fsPrivate.Metadata.CloudAccountId, fsPrivate.Metadata.Name)
	}
	fsPrivate.Status.Phase = pb.FilesystemPhase_FSDeleting

	return size, updateFilesystemForDelete(ctx, tx, fsPrivate)
}

func updateFilesystemForDelete(ctx context.Context, tx *sql.Tx, fsPrivate *pb.FilesystemPrivate) error {
	logger := log.FromContext(ctx).WithName("updateFilesystemForDelete")
	jsonVal, err := json.MarshalIndent(fsPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	logger.Info("updating filesystem state", logkeys.Query, updateFilesystemForDeletion)
	result, err := tx.ExecContext(ctx,
		updateFilesystemForDeletion,
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
	logger.Info("debug", logkeys.NumAffectedRows, rowsAffected)

	if rowsAffected < 1 {
		return status.Error(codes.FailedPrecondition, "no records updated; possible update conflict")
	}

	return nil
}

func UpdateFilesystemDeletionTime(ctx context.Context, tx *sql.Tx, cloudaccountId, resurceName string) error {
	logger := log.FromContext(ctx).WithName("UpdateFilesystemDeletionTime")
	logger.Info("updating filesystem deletion timestamp")
	result, err := tx.ExecContext(ctx,
		updateFilesystemDeletionTS,
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

func UpdateFilesystemState(ctx context.Context, tx *sql.Tx, in *pb.FilesystemUpdateStatusRequest) error {
	logger := log.FromContext(ctx).WithName("UpdateFilesystemState")
	logger.Info("updating filesystem state")
	fsPrivate, err := readFilesystemByName(ctx, tx, in.Metadata.CloudAccountId, in.Metadata.ResourceId, timestampInfinityStr)
	if err != nil {
		logger.Info("failed to update filesystem record for deletion")
		return err
	}
	// When filesystem resource is deleted, and operator takes long to delete the state,
	// the replicator replicates the state as "ready" from "deleting".
	// Following check will avoid updating state to "ready" after "deleting"

	// TODO: Verify this in the CRD status, why the state is not getting
	// set to deleting.
	if in.Status.Phase == pb.FilesystemPhase_FSReady && fsPrivate.Status.Phase == pb.FilesystemPhase_FSDeleting {
		logger.Info("skipping filesystem state update from `deleting` -> `ready'")
		return nil
	}

	fsPrivate.Status = in.Status
	jsonVal, err := json.MarshalIndent(fsPrivate, "", "    ")
	if err != nil {
		return fmt.Errorf("json marshaling: %w", err)
	}

	result, err := tx.ExecContext(ctx,
		updateFilesystemStatus,
		jsonVal,
		in.Metadata.CloudAccountId,
		in.Metadata.ResourceId,
		timestampInfinityStr,
	)
	if err != nil {
		logger.Error(err, "error updating filesystem state")
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

func readFilesystemByName(ctx context.Context, tx *sql.Tx, cloudaccountId, fsname, deletionTs string) (*pb.FilesystemPrivate, error) {
	logger := log.FromContext(ctx).WithName("readFilesystemByName")
	dataBuf := []byte{}
	fsPrivate := pb.FilesystemPrivate{}
	row := tx.QueryRowContext(ctx, getFilesystemByName, cloudaccountId, fsname, deletionTs)
	switch err := row.Scan(&dataBuf); err {
	case sql.ErrNoRows:
		logger.Info("no records found ", logkeys.ResourceName, fsname)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		err := json.Unmarshal([]byte(dataBuf), &fsPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "filesystem record search failed")
		}
	default:
		logger.Error(err, "error searching filesystem record in db")
		return nil, status.Errorf(codes.Internal, "filesystem record find failed")
	}

	return &fsPrivate, nil
}

func readFilesystemByResourceId(ctx context.Context, tx *sql.Tx, cloudaccountId, resourceId, deletionTs string) (*pb.FilesystemPrivate, error) {
	logger := log.FromContext(ctx).WithName("readFilesystemByResourceId")
	dataBuf := []byte{}
	fsPrivate := pb.FilesystemPrivate{}
	row := tx.QueryRowContext(ctx, getFilesystemByResourceId, cloudaccountId, resourceId, deletionTs)
	switch err := row.Scan(&dataBuf); err {
	case sql.ErrNoRows:
		logger.Info("no records found ", logkeys.ResourceId, resourceId)
		return nil, status.Errorf(codes.NotFound, "no matching records found")
	case nil:
		err := json.Unmarshal([]byte(dataBuf), &fsPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return nil, status.Errorf(codes.Internal, "filesystem record search failed")
		}
	default:
		logger.Error(err, "error searching filesystem record in db")
		return nil, status.Errorf(codes.Internal, "filesystem record find failed")
	}

	return &fsPrivate, nil
}
