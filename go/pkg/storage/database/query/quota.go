// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsedQuotaFile struct {
	CloudaccountId     string
	TotalFileInstances int64
	TotalFileSizeInGB  int64
}

type UsedQuotaObj struct {
	TotalBuckets    int64
	TotalPrincipals int64
}

type StorageQuotaByAccount struct {
	CloudAccountId    string
	CloudAccountType  string
	Reason            string
	FilesizeQuotaInTB int64
	FilevolumesQuota  int64
	BucketsQuota      int64
}

type StorageQuotaByType struct {
	CloudAccountType  string
	FilesizeQuotaInTB int64
	FilevolumesQuota  int64
	BucketsQuota      int64
}

const (
	getBucketGroupsByAccountId = `
		select cloud_account_id, count(*)
		from bucket
		where deleted_timestamp = 'infinity'
		group by cloud_account_id
	`
	insertStorageQuotaByAccount = `
		INSERT INTO storage_quota_by_account (cloud_account_id, cloud_account_type, reason, filesize_quota_in_TB, filevolumes_quota, buckets_quota)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING cloud_account_id, cloud_account_type, reason, filesize_quota_in_TB, filevolumes_quota, buckets_quota
	`
	updateStorageQuotaByAccount = `
		UPDATE storage_quota_by_account
		SET cloud_account_type= $2, reason = $3, filesize_quota_in_TB = $4, filevolumes_quota = $5, buckets_quota = $6, deleted_timestamp = 'infinity'
		WHERE cloud_account_id = $1 AND deleted_timestamp = 'infinity'
		RETURNING cloud_account_id, cloud_account_type, reason, filesize_quota_in_TB, filevolumes_quota, buckets_quota
	`
	deleteStorageQuotaByAccount = `
		UPDATE storage_quota_by_account
		SET deleted_timestamp = NOW()
		WHERE cloud_account_id = $1 AND deleted_timestamp = 'infinity'
	`
	getStorageQuotaByAccount = `
		SELECT cloud_account_id, cloud_account_type, reason, filesize_quota_in_TB, filevolumes_quota, buckets_quota
		FROM storage_quota_by_account
		WHERE cloud_account_id = $1 AND deleted_timestamp = 'infinity'
	`
	getAllStorageQuota = `
		SELECT sq.cloud_account_id, sq.cloud_account_type, sq.reason, sq.filesize_quota_in_TB, sq.filevolumes_quota, sq.buckets_quota
		FROM storage_quota_by_account sq
		WHERE sq.deleted_timestamp = 'infinity'
	`
)

func GetUsedQuotaByAccountId(ctx context.Context, tx *sql.Tx, cloudaccountId string) (*UsedQuotaFile, error) {
	logger := log.FromContext(ctx).WithName("GetUsedQuotaByAccountId").
		WithValues(logkeys.CloudAccountId, cloudaccountId)
	logger.Info("begin fiesystem quota retrieve")

	fsSearchResp, err := GetFilesystemsByCloudaccountId(ctx, tx, cloudaccountId,
		pb.FilesystemType_Unspecified, timestampInfinityStr)
	if err != nil {
		logger.Info("error reading filesystem used quota")
		return nil, fmt.Errorf("error checking quota")
	}

	usedQuota := UsedQuotaFile{
		CloudaccountId: cloudaccountId,
	}
	failed := false
	for idx := 0; idx < len(fsSearchResp); idx++ {
		fs := fsSearchResp[idx]
		if fs.Status.Phase != pb.FilesystemPhase_FSReady {
			continue
		}
		usedQuota.TotalFileInstances++
		fsSize := utils.ParseFileSize(fs.Spec.Request.Storage)
		if fsSize == -1 {
			logger.Info("failed to read file size quota for fs", logkeys.FilesystemName, fs.Metadata.Name)
			failed = true
			break
		} else {
			usedQuota.TotalFileSizeInGB += fsSize
		}
	}
	if failed {
		return nil, fmt.Errorf("error checking quota")
	}
	return &usedQuota, nil
}

func UpdateUsedFileQuotaForAllAccounts(ctx context.Context, dbconn *sql.DB, accQuota *map[string]UsedQuotaFile, deletionTs string) error {
	logger := log.FromContext(ctx).WithName("UpdateUsedFileQuotaForAllAccounts")
	logger.Info("retrieving used file quota for all accounts")

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error startind db tx: %w", err)
	}
	defer tx.Commit()

	query := getAllFSRequests
	params := []interface{}{}
	params = append(params, deletionTs)
	rows, err := tx.QueryContext(ctx, query, params...)
	if err != nil {
		logger.Error(err, "error searching filesystem record in db")
		return status.Errorf(codes.Internal, "filesystem record search failed")
	}
	defer rows.Close()
	nextVersion := "0"
	for rows.Next() {
		dataBuf := []byte{}

		filesystemPrivate := pb.FilesystemPrivate{}
		if err := rows.Scan(&nextVersion, &dataBuf); err != nil {
			// FIXME: Handle this failure
			logger.Info("error reading result row, continue...", logkeys.Error, err)
		}
		err := json.Unmarshal([]byte(dataBuf), &filesystemPrivate)
		if err != nil {
			logger.Error(err, "Error Unmarshalling JSON")
			return status.Errorf(codes.Internal, "filesystem record search failed")
		}

		// parse volume size
		size := utils.ParseFileSize(filesystemPrivate.Spec.Request.Storage)
		if size == -1 {
			logger.Info("invalid input size", "fsSize", filesystemPrivate.Spec.Request.Storage)
			return status.Error(codes.InvalidArgument, "invalid storage size")
		}

		// there are no quota checks for compute IKS currently.
		if filesystemPrivate.Spec.FilesystemType == pb.FilesystemType_ComputeGeneral {
			if currQuota, found := (*accQuota)[filesystemPrivate.Metadata.CloudAccountId]; found {
				currQuota.TotalFileInstances++
				currQuota.TotalFileSizeInGB += size
				(*accQuota)[filesystemPrivate.Metadata.CloudAccountId] = currQuota
			} else {
				(*accQuota)[filesystemPrivate.Metadata.CloudAccountId] = UsedQuotaFile{
					TotalFileInstances: 1,
					TotalFileSizeInGB:  size,
				}
			}
		}
	}

	logger.Info("total used filesystem quota (computeGeneral)", logkeys.TotalQuota, fmt.Sprintf("%v", *accQuota))
	return nil
}

func UpdateUsedObjQuotaForAllAccounts(ctx context.Context, dbconn *sql.DB, accObjQuota *map[string]UsedQuotaObj) error {
	logger := log.FromContext(ctx).WithName("UpdateUsedObjQuotaForAllAccounts")
	logger.Info("retrieving used object quota for all accounts")

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error startind db tx")
		return status.Errorf(codes.FailedPrecondition, "failed to start db transaction")
	}
	defer tx.Commit()

	rows, err := tx.QueryContext(ctx, getBucketGroupsByAccountId)
	if err != nil {
		logger.Error(err, "error searching bucket record in db")
		return status.Errorf(codes.Internal, "bucket record group search failed")
	}
	defer rows.Close()
	var cloudaccountId string
	var activeBuckets int64
	for rows.Next() {
		if err := rows.Scan(&cloudaccountId, &activeBuckets); err != nil {
			// FIXME: Handle this failure
			logger.Info("error reading result row, continue...", logkeys.Error, err)
		}
		(*accObjQuota)[cloudaccountId] = UsedQuotaObj{
			TotalBuckets: activeBuckets,
		}
	}

	logger.Info("total used buckets quota", logkeys.TotalQuota, fmt.Sprintf("%v", *accObjQuota))
	return nil
}
func InsertStorageQuotaByAccount(ctx context.Context, tx *sql.Tx, cloudAccountId, cloudAccountType, reason string, filesizeQuotaInTB, filevolumesQuota, bucketsQuota int64) (*StorageQuotaByAccount, error) {
	logger := log.FromContext(ctx).WithName("InsertStorageQuotaByAccount")
	logger.Info("insert storage quota by account")

	var quota StorageQuotaByAccount
	err := tx.QueryRowContext(ctx, insertStorageQuotaByAccount, cloudAccountId, cloudAccountType, reason, filesizeQuotaInTB, filevolumesQuota, bucketsQuota).Scan(&quota.CloudAccountId, &quota.CloudAccountType, &quota.Reason, &quota.FilesizeQuotaInTB, &quota.FilevolumesQuota, &quota.BucketsQuota)
	if err != nil {
		logger.Error(err, "error inserting bucket record in db")
		return nil, status.Errorf(codes.Internal, "bucket record insertion failed")
	}
	return &quota, nil
}

func UpdateStorageQuotaByAccount(ctx context.Context, tx *sql.Tx, cloudAccountId, cloudAcountType, reason string, filesizeQuotaInTB, filevolumesQuota, bucketsQuota int64) (*StorageQuotaByAccount, error) {
	logger := log.FromContext(ctx).WithName("UpdateStorageQuotaByAccount")
	logger.Info("update storage quota by account")
	// Fetch the current quota values
	var currentQuota StorageQuotaByAccount
	err := tx.QueryRowContext(ctx, getStorageQuotaByAccount, cloudAccountId).Scan(&currentQuota.CloudAccountId, &currentQuota.CloudAccountType, &currentQuota.Reason, &currentQuota.FilesizeQuotaInTB, &currentQuota.FilevolumesQuota, &currentQuota.BucketsQuota)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle no rows error specifically
			logger.Info("No current quota found for the given account ID")
		}
		return nil, err
	}
	logger.Info("quota values :",
		"currentFilesizeQuota", currentQuota.FilesizeQuotaInTB,
		"newFilesizeQuota", filesizeQuotaInTB,
		"currentFilevolumesQuota", currentQuota.FilevolumesQuota,
		"newFilevolumesQuota", filevolumesQuota,
		"currentBucketsQuota", currentQuota.BucketsQuota,
		"newBucketsQuota", bucketsQuota)

	// Validate that the new quota values are not less than the current values
	if filesizeQuotaInTB < currentQuota.FilesizeQuotaInTB || filevolumesQuota < currentQuota.FilevolumesQuota || bucketsQuota < currentQuota.BucketsQuota {
		return nil, errors.New("new quota values cannot be less than the current values")
	}

	var quota StorageQuotaByAccount
	err = tx.QueryRowContext(ctx, updateStorageQuotaByAccount, cloudAccountId, cloudAcountType, reason, filesizeQuotaInTB, filevolumesQuota, bucketsQuota).Scan(&quota.CloudAccountId, &quota.CloudAccountType, &quota.Reason, &quota.FilesizeQuotaInTB, &quota.FilevolumesQuota, &quota.BucketsQuota)
	if err != nil {
		logger.Error(err, "error updating storage account quota record in db")
		return nil, status.Errorf(codes.Internal, "storage account quota record update failed")
	}
	logger.Info("quota update values :", "quotas", quota)
	return &quota, nil
}

func DeleteStorageQuotaByAccount(ctx context.Context, tx *sql.Tx, cloudAccountId string) error {
	logger := log.FromContext(ctx).WithName("DeleteStorageQuotaByAccount")
	logger.Info("delete storage quota by account")

	row := tx.QueryRowContext(ctx, deleteStorageQuotaByAccount, cloudAccountId)

	// Assuming the query does not return any rows
	var dummyVar string
	err := row.Scan(&dummyVar)

	if err != nil {
		if err == sql.ErrNoRows {
			// No rows were returned by the query, but this is not necessarily an error
			return nil
		}
		// An actual error occurred
		return err
	}
	return nil
}

func GetStorageQuotaByAccount(ctx context.Context, tx *sql.Tx, cloudAccountId string) (*StorageQuotaByAccount, error) {
	logger := log.FromContext(ctx).WithName("GetStorageQuotaByAccount")
	logger.Info("get storage quota by account")

	row := tx.QueryRowContext(ctx, getStorageQuotaByAccount, cloudAccountId)

	var quota StorageQuotaByAccount
	err := row.Scan(&quota.CloudAccountId, &quota.CloudAccountType, &quota.Reason, &quota.FilesizeQuotaInTB, &quota.FilevolumesQuota, &quota.BucketsQuota)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &quota, nil
}

func GetAllStorageQuota(ctx context.Context, tx *sql.Tx) ([]StorageQuotaByAccount, error) {
	logger := log.FromContext(ctx).WithName("GetAllStorageQuota")
	logger.Info("get all storage quota by account")

	rows, err := tx.QueryContext(ctx, getAllStorageQuota)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var quotas []StorageQuotaByAccount
	for rows.Next() {
		var quota StorageQuotaByAccount
		if err := rows.Scan(&quota.CloudAccountId, &quota.CloudAccountType, &quota.Reason, &quota.FilesizeQuotaInTB, &quota.FilevolumesQuota, &quota.BucketsQuota); err != nil {
			return nil, err
		}
		quotas = append(quotas, quota)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return quotas, nil
}
