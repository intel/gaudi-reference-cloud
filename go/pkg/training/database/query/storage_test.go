// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// -- Test Cases --
func TestDeleteStorageNodePrivate(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	db, cleanup := setupDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Delete storage node from cluster", func(t *testing.T) {
		// Create slurm cluster entry
		expected := createSlurmCluster(ctx, db, t, nil)
		assert.NotNil(t, expected)

		// Check if the new cluster were added
		result, err := getCluster(db, expected.Cluster.ClusterId, expected.CloudAccountId)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Check if the new cluster storage node were added
		storageResults, err := getClusterStorage(ctx, db, expected.Cluster.ClusterId)
		assert.NoError(t, err)
		assert.NotNil(t, storageResults)

		assert.Equal(t, len(expected.Cluster.StorageNodes), len(storageResults))
		assert.Equal(t, expected.Cluster.StorageNodes[0].Name, storageResults[0].Name)
		assert.Equal(t, expected.Cluster.StorageNodes[0].Description, storageResults[0].Description)
		assert.Equal(t, expected.Cluster.StorageNodes[0].Capacity, storageResults[0].Capacity)
		assert.Equal(t, expected.Cluster.StorageNodes[0].AccessMode, storageResults[0].AccessMode)
		assert.Equal(t, expected.Cluster.StorageNodes[0].Mount, storageResults[0].Mount)
		assert.Equal(t, expected.Cluster.StorageNodes[0].LocalMountDir, storageResults[0].LocalMountDir)
		assert.Equal(t, expected.Cluster.StorageNodes[0].RemoteMountDir, storageResults[0].RemoteMountDir)

		// Call function under test
		err = DeleteStorageNodePrivate(ctx, db, expected.Cluster.StorageNodes[0].Name, expected.CloudAccountId)
		assert.NoError(t, err)

		// Perform validation on the database tables
		storageMappingQuery := `
			SELECT fs_resource_id FROM cluster_storage_mapping
			WHERE cluster_id IN (SELECT id FROM cluster WHERE cluster_id = $1 and cloud_account_id = $2)
			LIMIT 1
		`

		rows := db.QueryRowContext(ctx, storageMappingQuery, expected.Cluster.ClusterId, expected.CloudAccountId)
		assert.NotNil(t, rows)

		var secondFsResourceId string
		err = rows.Scan(&secondFsResourceId)
		assert.NoError(t, err)
		assert.NotNil(t, expected.Cluster.StorageNodes[1].FsResourceId, secondFsResourceId)

		// Call function under test
		err = DeleteStorageNodePrivate(ctx, db, expected.Cluster.StorageNodes[1].Name, expected.CloudAccountId)
		assert.NoError(t, err)

		rows = db.QueryRowContext(ctx, storageMappingQuery, expected.Cluster.ClusterId, expected.CloudAccountId)
		assert.NotNil(t, rows)

		var fsResourceId string
		err = rows.Scan(&fsResourceId)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Equal(t, "", fsResourceId)

		// Check if the storage is removed from the table
		storageResults, err = getClusterStorage(ctx, db, expected.Cluster.ClusterId)
		assert.NoError(t, err)
		assert.Equal(t, storageResults, []*pb.StorageNode{})

		// Test case teardown
		err = clearAllTrainingDatabase(db)
		assert.NoError(t, err)
	})
}
