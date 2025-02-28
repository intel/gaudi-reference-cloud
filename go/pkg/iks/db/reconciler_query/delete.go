// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package reconciler_query

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang/protobuf/ptypes/empty"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	pb "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

const (
	deleteAllClusterNodes = `
		DELETE FROM public.k8snode n
		WHERE n.cluster_id = $1
	`
	deleteAllClusterNodegroups = `
		DELETE FROM public.nodegroup n
		WHERE n.cluster_id = $1
	`
	deleteAllClusterVips = `
		DELETE FROM public.vip v
		WHERE v.cluster_id = $1
	`
	deleteAllClusterAddons = `
		DELETE FROM public.clusteraddonversion a
		WHERE a.cluster_id = $1
	`
	deleteAllClusterRevs = `
		DELETE FROM public.clusterrev cr
		WHERE cr.cluster_id = $1
	`
	deleteAllClusterSnapshots = `
		DELETE FROM public.snapshot s
		WHERE s.cluster_id = $1
	`
	getClusterVips = `
		SELECT vip_id
		FROM public.vip v
		WHERE v.cluster_id = $1
	`
	deleteVipDetails = `
		DELETE FROM public.vipdetails v
		WHERE v.vip_id = $1
	`
)

func DeleteClusterReconcilerQuery(ctx context.Context, dbconn *sql.DB, req *pb.ClusterDeletionRequest) (*empty.Empty, error) {
	friendlyMessage := "DeleteImi.UnexpectedError"
	failedFunction := "DeleteImi."
	emptyReturn := &empty.Empty{}

	//Check if cluster exists
	var clusterId string
	err := dbconn.QueryRowContext(ctx, CheckUuid, req.Uuid).Scan(&clusterId)
	if err != nil {
		if err == sql.ErrNoRows {
			return emptyReturn, errors.New("No cluster found")
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"CheckUuid", friendlyMessage+err.Error())
	}

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"dbconn.BeginTx", friendlyMessage+err.Error())
	}
	/* DELETE OBJECTS FROM DB */
	// Delete Nodes
	_, err = tx.Exec(deleteAllClusterNodes, clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteAllClusterNodes", friendlyMessage+err.Error())
	}
	// Delete Nodegroups
	_, err = tx.Exec(deleteAllClusterNodegroups, clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteAllClusterNodegroups", friendlyMessage+err.Error())
	}

	// Delete Addons
	_, err = tx.Exec(deleteAllClusterAddons, clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteAllClusterAddons", friendlyMessage+err.Error())
	}

	// Delete Snapshots
	_, err = tx.Exec(deleteAllClusterSnapshots, clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteAllClusterSnapshots", friendlyMessage+err.Error())
	}

	// Delete Cluster Revs
	_, err = tx.Exec(deleteAllClusterRevs, clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteAllClusterRevs", friendlyMessage+err.Error())
	}

	/* DELETE CLUSTER VIPS */
	// Get vips
	rows, err := dbconn.QueryContext(ctx, getClusterVips, clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"getClusterVips", friendlyMessage+err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var vipId int32
		err = rows.Scan(&vipId)
		// Delete Vip Details
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"getClusterVips.rows.scan", friendlyMessage+err.Error())
		}
		_, err = tx.Exec(deleteVipDetails, vipId)
		if err != nil {
			if errtx := tx.Rollback(); errtx != nil {
				return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
			}
			return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteVipDetails", friendlyMessage+err.Error())
		}
	}
	// Delete Vips
	_, err = tx.Exec(deleteAllClusterVips, clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"deleteAllClusterVips", friendlyMessage+err.Error())
	}

	// Set Cluster to Deleted
	_, err = tx.Exec(PutClusterState, "Deleted", clusterId)
	if err != nil {
		if errtx := tx.Rollback(); errtx != nil {
			return emptyReturn, utils.ErrorHandler(ctx, errtx, failedFunction+"TransactionRollbackError", friendlyMessage+errtx.Error())
		}
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"PutClusterState", friendlyMessage+err.Error())
	}

	// commit Transaction
	err = tx.Commit()
	if err != nil {
		return emptyReturn, utils.ErrorHandler(ctx, err, failedFunction+"tx.Commit", friendlyMessage+err.Error())
	}

	return emptyReturn, nil
}
