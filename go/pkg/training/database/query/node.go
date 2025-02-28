// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const (
	getClusterNodeIds = `
		SELECT n.node_id
		FROM node as n, cluster_node_mapping as cn, cluster as c
		WHERE n.id = cn.node_id AND
		cn.cluster_id = c.id AND
		c.cluster_id = $1
	`

	// FUTURE IMPROVEMENT: Delete specific compute instances per cluster
	deleteNodeByClusterId = `
		DELETE FROM node WHERE node_id = $1 AND cloud_account_id = $2;
	`

	deleteClusterNodeMappingByClusterId = `
		DELETE FROM cluster_node_mapping
		WHERE cluster_id IN (SELECT id FROM cluster WHERE cloud_account_id = $1 AND cluster_id = $2);
	`
)

// Private cluster node delete function that returns an error and not a status code.
func DeleteAllClusterNodeInstancesPrivate(ctx context.Context, dbconn *sql.DB, clusterId, cloudAccountId string) error {
	logger := log.FromContext(ctx).WithName("DeleteAllClusterNodeInstancesPrivate")

	allClusterNodeIds := []string{}

	// Get all the associated nodeIds for the specified cluster to delete after removing db constraints
	nodeRows, err := dbconn.QueryContext(ctx, getClusterNodeIds, clusterId)
	if err != nil {
		logger.Error(err, "error reading node ids")
		return errors.New("error querying nodes data")
	}

	for nodeRows.Next() {
		var nodeId string

		if err := nodeRows.Scan(&nodeId); err != nil {
			logger.Error(err, "error scaning node record in db")
			return errors.New("record search failed")
		}

		allClusterNodeIds = append(allClusterNodeIds, nodeId)
	}

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error begin transaction data")
		return fmt.Errorf("error deleting all cluster nodes")
	}
	defer tx.Rollback()

	logger.Info("start the process to delete all nodes from the cluster")

	result, err := tx.ExecContext(ctx, deleteClusterNodeMappingByClusterId, cloudAccountId, clusterId)
	if err != nil {
		logger.Error(err, "error deleting cluster node mapping")
		return fmt.Errorf("error deletig cluster node mapping")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(err, "error reading rows affected by delete in cluster storage mapper")
		return err
	}

	logger.Info("cluster node mapping", "number of affected rows", rowsAffected)

	rowsAffected = 0
	for _, nid := range allClusterNodeIds {
		result, err = tx.ExecContext(ctx, deleteNodeByClusterId, nid, cloudAccountId)
		if err != nil {
			logger.Error(err, "error deleting cluster nodes")
			return fmt.Errorf("error deleting cluster nodes from database")
		}

		rowsChanged, err := result.RowsAffected()
		if err != nil {
			logger.Error(err, "error reading rows affected by delete in node database")
			return err
		}

		rowsAffected += rowsChanged
	}
	// tx.Commit()
	err = tx.Commit()
	if err != nil {
		logger.Error(err, "error committing transaction")
		return err
	}
	logger.Info("node database", "number of affected rows", rowsAffected)

	return nil
}
