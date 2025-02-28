// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

const (
	checkDuplicateStorageName = `
		SELECT fs_resource_id, name, description, capacity FROM storage
		WHERE name = $1 AND cloud_account_id = $2
	`

	deleteStorageFSByName = `
		DELETE FROM storage WHERE name = $1 AND cloud_account_id = $2;
	`

	deleteStorageFSMappingByName = `
		DELETE FROM cluster_storage_mapping
		WHERE fs_resource_id IN (SELECT id FROM storage WHERE name = $1 AND cloud_account_id = $2);
	`
)

// Private cluster storage delete function that returns an error and not a status code.
func DeleteStorageNodePrivate(ctx context.Context, dbconn *sql.DB, storageName, cloudAccountId string) error {
	logger := log.FromContext(ctx).WithName("DeleteStorageNodePrivate")

	tx, err := dbconn.BeginTx(ctx, nil)
	if err != nil {
		logger.Error(err, "error begin transaction data")
		return fmt.Errorf("error deleting storage request")
	}
	defer tx.Rollback()

	logger.Info("start the process to delete storage node", "name", storageName)

	result, err := tx.ExecContext(ctx, deleteStorageFSMappingByName, storageName, cloudAccountId)
	if err != nil {
		logger.Error(err, "error performing delete operations on cluster storage mapper")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(err, "error reading rows affected by delete in cluster storage mapper")
		return err
	}

	logger.Info("cluster storage mapping", "number of affected rows", rowsAffected)

	result, err = tx.ExecContext(ctx, deleteStorageFSByName, storageName, cloudAccountId)
	if err != nil {
		logger.Error(err, "error performing delete operations on storage database")
		return err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		logger.Error(err, "error reading rows affected by delete in storage database")
		return err
	}

	// tx.Commit()
	err = tx.Commit()
	if err != nil {
		logger.Error(err, "error committing transaction")
		return err
	}
	logger.Info("storage database", "number of affected rows", rowsAffected)

	return nil
}
