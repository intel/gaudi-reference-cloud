// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	insertFilesystemAccountQuery = `
		insert into filesystem_namespace_user_account 
			(cloud_account_id, cluster_name, cluster_addr, cluster_uuid,  namespace_name) 
		values ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING
	`

	getExistingFilesystemAccounts = `
		select cluster_name, cluster_addr, cluster_uuid, namespace_name
		from filesystem_namespace_user_account
		where cloud_account_id = $1 
		and deleted_timestamp = $2
	`
	deleteFilesystemAccounts = `
		delete from filesystem_namespace_user_account
		where cloud_account_id = $1
		and cluster_uuid = $2 
		and deleted_timestamp = $3
	`
)

func StoreFilesystemAccount(ctx context.Context, tx *sql.Tx, cloudaccountId string, status *pb.FilesystemSchedule) error {
	logger := log.FromContext(ctx).WithName("StoreFilesystemAccount").WithValues(logkeys.CloudAccountId, cloudaccountId)

	logger.Info("begin fiesystem account insertion")

	if status != nil {
		if status.Cluster == nil || status.Cluster.ClusterName == "" {
			logger.Info("filesystem volume scheduled is empty")
			return nil
		}
		_, err := tx.ExecContext(ctx,
			insertFilesystemAccountQuery,
			cloudaccountId,
			status.Cluster.ClusterName,
			status.Cluster.ClusterAddr,
			status.Cluster.ClusterUUID,
			status.Namespace.Name,
		)

		if err != nil {
			return err
		}
	}
	logger.Info("fiesystem account record stored successfully")

	return nil
}

// GetFilesystemAccounts: returns a list of already assigned/alloted clusters and namespaces
func GetFilesystemAccounts(ctx context.Context, tx *sql.Tx, cloudaccountId, deletionTs string) (*pb.FilesystemScheduleRequest, error) {
	logger := log.FromContext(ctx).WithName("GetFilesystemAccounts").WithValues(logkeys.CloudAccountId, cloudaccountId)

	logger.Info("begin fiesystem account retrieve")

	rows, err := tx.QueryContext(ctx, getExistingFilesystemAccounts, cloudaccountId, deletionTs)
	if err != nil {
		logger.Error(err, "error searching filesystem record in db")
		return nil, status.Errorf(codes.Internal, "filesystem record search failed")
	}
	defer rows.Close()
	resp := pb.FilesystemScheduleRequest{
		CloudaccountId: cloudaccountId,
	}
	for rows.Next() {
		currRes := pb.ResourceSchedule{}
		if err := rows.Scan(&currRes.ClusterName, &currRes.ClusterAddr, &currRes.ClusterUUID, &currRes.Namespace); err != nil {
			return nil, fmt.Errorf("db read: %w", err)
		}
		resp.Assignments = append(resp.Assignments, &currRes)
	}
	logger.Info("fiesystem account records retrieve completed successfully")
	return &resp, nil
}

// DeleteFilesystemAccount: Deletes the entry from the filesystem_namespace_user_account if the last fs has been deleted.
func DeleteFilesystemAccount(ctx context.Context, tx *sql.Tx, cloudaccountId, clusterId string, deletionTs string) error {
	logger := log.FromContext(ctx).WithName("DeleteFilesystemAccount").WithValues(logkeys.CloudAccountId, cloudaccountId)

	logger.Info("begin filesystem account delete")

	rows, err := tx.QueryContext(ctx, deleteFilesystemAccounts, cloudaccountId, clusterId, deletionTs)
	if err != nil {
		logger.Error(err, "error deleting filesystem record in db")
		return status.Errorf(codes.Internal, "filesystem record delete failed")
	}
	defer rows.Close()
	logger.Info("filesystem account record delete completed successfully")
	return nil
}
