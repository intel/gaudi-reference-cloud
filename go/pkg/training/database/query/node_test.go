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
func TestDeleteAllClusterNodeInstancesPrivate(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	db, cleanup := setupDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Delete all nodes from a cluster", func(t *testing.T) {
		// Create slurm cluster entry
		expected := createSlurmCluster(ctx, db, t, nil)
		assert.NotNil(t, expected)

		// Check if the new cluster were added
		result, err := getCluster(db, expected.Cluster.ClusterId, expected.CloudAccountId)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, expected.Cluster.Name, result.Name)
		assert.Equal(t, expected.Cluster.Description, result.Description)
		assert.Equal(t, expected.Cluster.SSHKeyName, result.SSHKeyName)

		// Check if the new cluster compute nodes were added
		nodeResults, err := getClusterNode(ctx, db, expected.Cluster.ClusterId)
		assert.NoError(t, err)
		assert.NotNil(t, nodeResults)
		assert.Equal(t, getTotalNodeCount(expected.Cluster.Nodes), len(nodeResults))

		// Call function under test
		err = DeleteAllClusterNodeInstancesPrivate(ctx, db, expected.Cluster.ClusterId, expected.CloudAccountId)
		assert.NoError(t, err)

		// Perform validation on the database tables
		nodeMappingQuery := `
			SELECT node_id FROM cluster_node_mapping
			WHERE cluster_id IN (SELECT id FROM cluster WHERE cluster_id = $1 and cloud_account_id = $2)
			LIMIT 1
		`

		rows := db.QueryRowContext(ctx, nodeMappingQuery, expected.Cluster.ClusterId, expected.CloudAccountId)
		assert.NotNil(t, rows)

		var nodeId string
		err = rows.Scan(&nodeId)
		assert.Error(t, err)
		assert.Equal(t, sql.ErrNoRows, err)
		assert.Equal(t, "", nodeId)

		// Check if the nodes are removed from the node table
		nodeResults, err = getClusterNode(ctx, db, expected.Cluster.ClusterId)
		assert.NoError(t, err)
		assert.Equal(t, nodeResults, []*pb.ClusterNode{})

		// Test case teardown
		err = clearAllTrainingDatabase(db)
		assert.NoError(t, err)
	})

	t.Run("Delete all nodes from a non-existing cluster", func(t *testing.T) {
		// Ensure no entries in the database
		err := clearAllTrainingDatabase(db)
		assert.NoError(t, err)

		fakeCloudAccountId := "fake-cloud-account-id"
		fakeClusterId := "fake-cluster-id"

		// Call function under test
		err = DeleteAllClusterNodeInstancesPrivate(ctx, db, fakeClusterId, fakeCloudAccountId)
		assert.NoError(t, err)
	})
}
