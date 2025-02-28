// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudaccount"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

// -- Teardown Functions --
func clearAllTrainingDatabase(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM cluster_node_mapping")
	_, err = db.Exec("DELETE FROM node")

	_, err = db.Exec("DELETE FROM cluster_storage_mapping")
	_, err = db.Exec("DELETE FROM storage")
	_, err = db.Exec("DELETE FROM vnet_spec")

	_, err = db.Exec("DELETE FROM cluster")
	return err
}

// -- Utility Functions --
func getCluster(db *sql.DB, clusterId, cloudAccountId string) (*pb.Cluster, error) {
	cl := pb.Cluster{}
	sshKeyNames := json.RawMessage{}

	err := db.QueryRow(getClusterById, clusterId, cloudAccountId).Scan(&cl.Name, &cl.Description, &sshKeyNames)
	json.Unmarshal(sshKeyNames, &cl.SSHKeyName)

	return &cl, err
}

func getClusterVnet(ctx context.Context, db *sql.DB, clusterId string) (*pb.VNetSpec, error) {
	clusterVnet := &pb.VNetSpec{}

	specRows, err := db.QueryContext(ctx, getClusterVnetSpec, clusterId)
	if err != nil {
		return nil, err
	}

	for specRows.Next() {
		spec := pb.VNetSpec{}

		if err := specRows.Scan(&spec.Region, &spec.AvailabilityZone, &spec.PrefixLength); err != nil {
			return nil, err
		}

		clusterVnet = &spec
	}

	return clusterVnet, err
}

func getClusterStorage(ctx context.Context, db *sql.DB, clusterId string) ([]*pb.StorageNode, error) {
	clusterStorages := []*pb.StorageNode{}

	storageRows, err := db.QueryContext(ctx, getClusterStorageNodes, clusterId)
	if err != nil {
		return nil, err
	}

	for storageRows.Next() {
		storageFS := pb.StorageNode{}
		var access, mount string

		if err := storageRows.Scan(&storageFS.FsResourceId, &storageFS.Name, &storageFS.Description, &storageFS.Capacity, &access, &mount, &storageFS.LocalMountDir, &storageFS.RemoteMountDir); err != nil {
			return nil, err
		}

		storageFS.AccessMode = marshalAccessModeToPb(access)
		storageFS.Mount = marshalMountProtocolToPb(mount)
		clusterStorages = append(clusterStorages, &storageFS)
	}

	return clusterStorages, err
}

func getClusterNode(ctx context.Context, db *sql.DB, clusterId string) ([]*pb.ClusterNode, error) {
	clusterNodes := []*pb.ClusterNode{}

	nodeRows, err := db.QueryContext(ctx, getClusterNodesWithCount, clusterId)

	for nodeRows.Next() {
		node := pb.ClusterNode{}
		var role, machineType, nodeStatus string
		labels := json.RawMessage{}

		err := nodeRows.Scan(&node.NodeId, &node.Name, &role, &labels, &machineType, &node.ImageType, &nodeStatus, &node.Count)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(labels, &node.Labels)

		node.Role = marshalRoleToPb(role)
		node.MachineType = marshalMachineTypeToPb(machineType)
		clusterNodes = append(clusterNodes, &node)
	}

	return clusterNodes, err
}

func createSlurmCluster(ctx context.Context, db *sql.DB, t *testing.T, clusterReq *pb.SlurmClusterCreateRequest) *pb.SlurmClusterCreateRequest {
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("transaction init failed %v", err)
	}
	defer tx.Rollback()

	var newCluster *pb.SlurmClusterCreateRequest
	if clusterReq != nil {
		newCluster = clusterReq
	} else {
		expectedCloudAccountId := cloudaccount.MustNewId()
		expectedClusterId := uuid.NewString()
		expectedStorageDescription := "Test storage node database insertion"

		newCluster = &pb.SlurmClusterCreateRequest{
			CloudAccountId: expectedCloudAccountId,
			Cluster: &pb.Cluster{
				CloudAccountId: expectedCloudAccountId,
				Name:           "test-static-slurm-cluster-creation",
				ClusterId:      expectedClusterId,
				Description:    "Test creating a static slurm cluster",
				SSHKeyName:     []string{"user1@acme.com", "user2@acme.com"},
				Spec: &pb.VNetSpec{
					Region:           "us-staging-1",
					AvailabilityZone: "us-staging-1a",
					PrefixLength:     24,
				},
				Nodes: []*pb.ClusterNode{
					{
						Count:       uint32(2),
						ImageType:   "ubuntu-2204-jammy-v20230122",
						MachineType: pb.MachineType_MED_VM_TYPE,
						Role:        pb.NodeRole_JUPYTERHUB_NODE,
						Labels: map[string]string{
							"oneapi-instance-role": "slurm-jupyterhub-node",
						},
					},
					{
						Count:       uint32(2),
						ImageType:   "ubuntu-2204-jammy-v20230122",
						MachineType: pb.MachineType_LRG_VM_TYPE,
						Role:        pb.NodeRole_LOGIN_NODE,
						Labels: map[string]string{
							"oneapi-instance-role": "slurm-login-node",
						},
					},
					{
						Count:       uint32(1),
						ImageType:   "ubuntu-2204-jammy-v20230122",
						MachineType: pb.MachineType_SML_VM_TYPE,
						Role:        pb.NodeRole_CONTROLLER_NODE,
						Labels: map[string]string{
							"oneapi-instance-role": "slurm-controller-node",
						},
					},
					{
						Count:       uint32(2),
						ImageType:   "ubuntu-2204-jammy-v20230122",
						MachineType: pb.MachineType_GAUDI_BM_TYPE,
						Role:        pb.NodeRole_COMPUTE_NODE,
						Labels: map[string]string{
							"oneapi-instance-role": "slurm-compute-node",
						},
					},
				},
				StorageNodes: []*pb.StorageNode{
					{
						Name:           "test-storage-1",
						Description:    &expectedStorageDescription,
						Capacity:       "5GB",
						AccessMode:     pb.StorageAccessModeType_STORAGE_READ_WRITE,
						Mount:          pb.StorageMountType_STORAGE_WEKA,
						LocalMountDir:  "/test",
						RemoteMountDir: "/export/test",
					},
					{
						Name:           "test-storage-2",
						Description:    &expectedStorageDescription,
						Capacity:       "10GB",
						AccessMode:     pb.StorageAccessModeType_STORAGE_READ_WRITE_ONCE,
						Mount:          pb.StorageMountType_STORAGE_WEKA,
						LocalMountDir:  "/test",
						RemoteMountDir: "/export/test",
					},
				},
			},
		}
	}

	// Create entry in database
	err = CreateClusterState(ctx, db, newCluster.Cluster.ClusterId, newCluster)
	assert.NoError(t, err)
	err = tx.Commit()
	assert.NoError(t, err)

	return newCluster
}

func getTotalNodeCount(clusterDetails []*pb.ClusterNode) int {
	totalCount := uint32(0)

	for _, node := range clusterDetails {
		totalCount += node.Count
	}

	return int(totalCount)
}

// -- Test Cases --
func TestCreateClusterState(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	db, cleanup := setupDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Insert new slurm cluster - 1 storage, 1 node", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("transaction init failed %v", err)
		}
		defer tx.Rollback()

		expectedCloudAccountId := cloudaccount.MustNewId()
		expectedClusterId := uuid.NewString()
		expectedStorageDescription := "Test storage node database insertion"

		cluster := &pb.SlurmClusterCreateRequest{
			CloudAccountId: expectedCloudAccountId,
			Cluster: &pb.Cluster{
				CloudAccountId: expectedCloudAccountId,
				Name:           "test-static-slurm-cluster-creation",
				ClusterId:      expectedClusterId,
				Description:    "Test creating a static slurm cluster",
				SSHKeyName:     []string{"user1@acme.com", "user2@acme.com"},
				Spec: &pb.VNetSpec{
					Region:           "us-staging-1",
					AvailabilityZone: "us-staging-1a",
					PrefixLength:     24,
				},
				Nodes: []*pb.ClusterNode{
					{
						Count:       uint32(1),
						ImageType:   "ubuntu-2204-jammy-v20230122",
						MachineType: pb.MachineType_PVC_BM_1100_4,
						Role:        pb.NodeRole_JUPYTERHUB_NODE,
						Labels: map[string]string{
							"oneapi-instance-role": "slurm-jupyterhub-node",
						},
					},
				},
				StorageNodes: []*pb.StorageNode{
					{
						Name:           "test-storage",
						Description:    &expectedStorageDescription,
						Capacity:       "5GB",
						AccessMode:     pb.StorageAccessModeType_STORAGE_READ_ONLY,
						Mount:          pb.StorageMountType_STORAGE_WEKA,
						LocalMountDir:  "/test",
						RemoteMountDir: "/export/test",
					},
				},
			},
		}

		// Call the function under test
		err = CreateClusterState(ctx, db, expectedClusterId, cluster)
		assert.NoError(t, err)
		err = tx.Commit()
		assert.NoError(t, err)

		// Check if the new cluster were added
		result, err := getCluster(db, cluster.Cluster.ClusterId, cluster.CloudAccountId)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, cluster.Cluster.Name, result.Name)
		assert.Equal(t, cluster.Cluster.Description, result.Description)
		assert.Equal(t, cluster.Cluster.SSHKeyName, result.SSHKeyName)

		// Check if the new cluster vnet spec were added
		vnetSpecResults, err := getClusterVnet(ctx, db, cluster.Cluster.ClusterId)
		assert.NoError(t, err)
		assert.NotNil(t, vnetSpecResults)

		assert.Equal(t, cluster.Cluster.Spec.Region, vnetSpecResults.Region)
		assert.Equal(t, cluster.Cluster.Spec.AvailabilityZone, vnetSpecResults.AvailabilityZone)
		assert.Equal(t, cluster.Cluster.Spec.PrefixLength, vnetSpecResults.PrefixLength)

		// Check if the new cluster storage node were added
		storageResults, err := getClusterStorage(ctx, db, cluster.Cluster.ClusterId)
		assert.NoError(t, err)
		assert.NotNil(t, storageResults)

		assert.Equal(t, cluster.Cluster.StorageNodes[0].Name, storageResults[0].Name)
		assert.Equal(t, cluster.Cluster.StorageNodes[0].Description, storageResults[0].Description)
		assert.Equal(t, cluster.Cluster.StorageNodes[0].Capacity, storageResults[0].Capacity)
		assert.Equal(t, cluster.Cluster.StorageNodes[0].AccessMode, storageResults[0].AccessMode)
		assert.Equal(t, cluster.Cluster.StorageNodes[0].Mount, storageResults[0].Mount)
		assert.Equal(t, cluster.Cluster.StorageNodes[0].LocalMountDir, storageResults[0].LocalMountDir)
		assert.Equal(t, cluster.Cluster.StorageNodes[0].RemoteMountDir, storageResults[0].RemoteMountDir)

		// Check if the new cluster compute nodes were added
		nodeResults, err := getClusterNode(ctx, db, cluster.Cluster.ClusterId)
		assert.NoError(t, err)
		assert.NotNil(t, nodeResults)

		assert.Equal(t, cluster.Cluster.Nodes[0].Count, nodeResults[0].Count)
		assert.Equal(t, cluster.Cluster.Nodes[0].ImageType, nodeResults[0].ImageType)
		assert.Equal(t, cluster.Cluster.Nodes[0].MachineType, nodeResults[0].MachineType)
		assert.Equal(t, cluster.Cluster.Nodes[0].Role, nodeResults[0].Role)
		assert.Equal(t, cluster.Cluster.Nodes[0].Labels, nodeResults[0].Labels)

		// Test case teardown
		err = clearAllTrainingDatabase(db)
		assert.NoError(t, err)
	})

	t.Run("Insert static slurm cluster - Complete", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("transaction init failed %v", err)
		}
		defer tx.Rollback()

		// Call the function under test
		expectedStaticSlurmCluster := createSlurmCluster(ctx, db, t, nil)
		assert.NotNil(t, expectedStaticSlurmCluster)

		expectedClusterId := expectedStaticSlurmCluster.Cluster.ClusterId

		// Check for the new cluster
		result, err := getCluster(db, expectedClusterId, expectedStaticSlurmCluster.CloudAccountId)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, expectedStaticSlurmCluster.Cluster.Name, result.Name)
		assert.Equal(t, expectedStaticSlurmCluster.Cluster.Description, result.Description)
		assert.Equal(t, expectedStaticSlurmCluster.Cluster.SSHKeyName, result.SSHKeyName)

		vnetspecs, err := getClusterVnet(ctx, db, expectedClusterId)
		assert.NoError(t, err)
		assert.NotNil(t, vnetspecs)
		assert.Equal(t, expectedStaticSlurmCluster.Cluster.Spec, vnetspecs)

		storages, err := getClusterStorage(ctx, db, expectedClusterId)
		assert.NoError(t, err)
		assert.NotNil(t, storages)
		assert.Equal(t, len(expectedStaticSlurmCluster.Cluster.StorageNodes), len(storages))

		nodes, err := getClusterNode(ctx, db, expectedClusterId)
		assert.NoError(t, err)
		assert.NotNil(t, nodes)
		assert.Equal(t, getTotalNodeCount(expectedStaticSlurmCluster.Cluster.Nodes), len(nodes))

		// Test case teardown
		err = clearAllTrainingDatabase(db)
		assert.NoError(t, err)
	})

	t.Run("Insert static slurm cluster error - Duplicate storage name entries", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("transaction init failed %v", err)
		}
		defer tx.Rollback()

		expectedCloudAccountId := cloudaccount.MustNewId()
		expectedClusterId := uuid.NewString()
		expectedStorageDescription := "Test storage node database insertion"
		expectedStorageDescDuplicate := "Different description same storage name"

		cluster := &pb.SlurmClusterCreateRequest{
			CloudAccountId: expectedCloudAccountId,
			Cluster: &pb.Cluster{
				CloudAccountId: expectedCloudAccountId,
				Name:           "test-static-slurm-cluster-creation",
				ClusterId:      expectedClusterId,
				Description:    "Test creating a static slurm cluster",
				SSHKeyName:     []string{"user1@acme.com", "user2@acme.com"},
				Spec: &pb.VNetSpec{
					Region:           "us-staging-1",
					AvailabilityZone: "us-staging-1a",
					PrefixLength:     24,
				},
				Nodes: []*pb.ClusterNode{
					{
						Count:       uint32(1),
						ImageType:   "ubuntu-2204-jammy-v20230122",
						MachineType: pb.MachineType_MED_VM_TYPE,
						Role:        pb.NodeRole_JUPYTERHUB_NODE,
						Labels: map[string]string{
							"oneapi-instance-role": "slurm-jupyterhub-node",
						},
					},
				},
				StorageNodes: []*pb.StorageNode{
					{
						Name:           "test-storage-1",
						Description:    &expectedStorageDescription,
						Capacity:       "5GB",
						AccessMode:     pb.StorageAccessModeType_STORAGE_READ_ONLY,
						Mount:          pb.StorageMountType_STORAGE_WEKA,
						LocalMountDir:  "/test",
						RemoteMountDir: "/export/test",
					},
					{
						Name:           "test-storage-1",
						Description:    &expectedStorageDescDuplicate,
						Capacity:       "10GB",
						AccessMode:     pb.StorageAccessModeType_STORAGE_READ_ONLY,
						Mount:          pb.StorageMountType_STORAGE_WEKA,
						LocalMountDir:  "/test",
						RemoteMountDir: "/export/test",
					},
				},
			},
		}

		// Call the function under test
		err = CreateClusterState(ctx, db, expectedClusterId, cluster)
		assert.Error(t, err)
		assert.Equal(t, err, errors.New("error duplicate storage names are not allowed"))

		// Test case teardown
		err = clearAllTrainingDatabase(db)
		assert.NoError(t, err)
	})
}

func TestGetNextClusterRequest(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	db, cleanup := setupDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Get next cluster request - no records found", func(t *testing.T) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("transaction init failed %v", err)
		}
		defer tx.Rollback()

		// Ensure the training database is completly empty
		clearAllTrainingDatabase(db)

		err = tx.Commit()
		assert.NoError(t, err)

		// Call the function under test
		nextClusterResults, err := GetNextClusterRequest(ctx, db)
		assert.NoError(t, err)
		assert.Nil(t, nextClusterResults)
	})

	t.Run("Get next cluster request - found complete cluster", func(t *testing.T) {
		// Create slurm cluster entry
		expected := createSlurmCluster(ctx, db, t, nil)
		assert.NotNil(t, expected)

		// Call the function under test
		actualGetClusterResults, err := GetNextClusterRequest(ctx, db)
		assert.NoError(t, err)
		assert.NotNil(t, actualGetClusterResults)

		// Check for the next cluster request returns
		assert.Equal(t, expected.Cluster.Name, actualGetClusterResults.Name)
		assert.Equal(t, expected.Cluster.SSHKeyName, actualGetClusterResults.SSHKeyName)
		assert.Equal(t, expected.Cluster.Spec, actualGetClusterResults.Spec)
		assert.Equal(t, len(expected.Cluster.StorageNodes), len(actualGetClusterResults.StorageNodes))
		assert.Equal(t, getTotalNodeCount(expected.Cluster.Nodes), len(actualGetClusterResults.Nodes))

		// Test case teardown
		err = clearAllTrainingDatabase(db)
		assert.NoError(t, err)
	})
}

func TestUpdateClusterState(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	db, cleanup := setupDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Update cluster states", func(t *testing.T) {
		getClusterStatus := `
			SELECT status FROM cluster
			WHERE cluster_id = $1 AND cloud_account_id = $2
		`

		// Create slurm cluster entry
		expected := createSlurmCluster(ctx, db, t, nil)
		assert.NotNil(t, expected)

		expectedClusterId := expected.Cluster.ClusterId
		expectedCloudAccountId := expected.CloudAccountId

		expectedStatus := []string{
			STATE_PROVISIONING,
			STATE_READY,
			STATE_FAILED,
		}

		for _, status := range expectedStatus {
			expectedClusterStatus := status
			err := UpdateClusterState(ctx, db, expectedClusterId, expectedCloudAccountId, expectedClusterStatus)
			assert.NoError(t, err)

			var actualStatus string
			err = db.QueryRowContext(ctx, getClusterStatus, expectedClusterId, expectedCloudAccountId).Scan(&actualStatus)
			assert.NoError(t, err)
			assert.Equal(t, expectedClusterStatus, actualStatus)
		}

		// Test case teardown
		err := clearAllTrainingDatabase(db)
		assert.NoError(t, err)
	})
}

func TestGetClustersByCloudAccount(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	db, cleanup := setupDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Get cluster ids using cloud account id", func(t *testing.T) {
		expectedClusters := []*pb.SlurmClusterCreateRequest{}

		// Create slurm cluster entry
		expected := createSlurmCluster(ctx, db, t, nil)
		assert.NotNil(t, expected)

		expectedClusters = append(expectedClusters, expected)

		// Call function under test
		result, err := GetClustersByCloudAccount(ctx, db, &pb.ClusterListOption{CloudAccountId: expectedClusters[0].CloudAccountId})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, expectedClusters[0].CloudAccountId, result.Clusters[0].CloudAccountId)
		assert.Equal(t, expectedClusters[0].Cluster.ClusterId, result.Clusters[0].ClusterId)
		assert.Equal(t, expectedClusters[0].Cluster.Name, result.Clusters[0].Name)
		assert.Equal(t, expectedClusters[0].Cluster.Description, result.Clusters[0].Description)

		// Add one more to get multiple clusters per cloud account
		expectedClusterId := uuid.NewString()
		expectedStorageDescription := "Test second cluster per cloud account"

		secondCluster := &pb.SlurmClusterCreateRequest{
			CloudAccountId: expectedClusters[0].CloudAccountId,
			Cluster: &pb.Cluster{
				CloudAccountId: expectedClusters[0].CloudAccountId,
				Name:           "test-slurm-cluster-getters",
				ClusterId:      expectedClusterId,
				Description:    "Test creating a static slurm cluster",
				SSHKeyName:     []string{},
				Spec: &pb.VNetSpec{
					Region:           "us-staging-1",
					AvailabilityZone: "us-staging-1a",
					PrefixLength:     24,
				},
				Nodes: []*pb.ClusterNode{
					{
						Count:       uint32(2),
						ImageType:   "ubuntu-2204-jammy-v20230122",
						MachineType: pb.MachineType_LRG_VM_TYPE,
						Role:        pb.NodeRole_JUPYTERHUB_NODE,
						Labels: map[string]string{
							"oneapi-instance-role": "slurm-jupyterhub-node",
						},
					},
				},
				StorageNodes: []*pb.StorageNode{
					{
						Name:           "new-storage-test-1",
						Description:    &expectedStorageDescription,
						Capacity:       "10GB",
						AccessMode:     pb.StorageAccessModeType_STORAGE_READ_WRITE,
						Mount:          pb.StorageMountType_STORAGE_WEKA,
						LocalMountDir:  "/test",
						RemoteMountDir: "/export/test",
					},
				},
			},
		}

		expected = createSlurmCluster(ctx, db, t, secondCluster)
		assert.NotNil(t, expected)

		expectedClusters = append(expectedClusters, expected)

		// Call function under test
		result, err = GetClustersByCloudAccount(ctx, db, &pb.ClusterListOption{CloudAccountId: expectedClusters[1].CloudAccountId})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, expectedClusters[1].CloudAccountId, result.Clusters[1].CloudAccountId)
		assert.Equal(t, expectedClusters[1].Cluster.ClusterId, result.Clusters[1].ClusterId)
		assert.Equal(t, expectedClusters[1].Cluster.Name, result.Clusters[1].Name)
		assert.Equal(t, expectedClusters[1].Cluster.Description, result.Clusters[1].Description)

		// Test case teardown
		err = clearAllTrainingDatabase(db)
		assert.NoError(t, err)
	})
}

func TestGetClusterByID(t *testing.T) {
	t.Skip("Tests causing errors elsewhere, skip for now")

	db, cleanup := setupDB(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Get cluster details by ID", func(t *testing.T) {
		// Create slurm cluster entry
		expected := createSlurmCluster(ctx, db, t, nil)
		assert.NotNil(t, expected)

		// Call the function under test
		result, err := GetClusterByID(ctx, db, &pb.SlurmClusterRequest{
			CloudAccountId: expected.CloudAccountId,
			ClusterId:      expected.Cluster.ClusterId,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		assert.Equal(t, expected.CloudAccountId, result.CloudAccountId)
		assert.Equal(t, expected.Cluster.Name, result.Name)
		assert.Equal(t, expected.Cluster.ClusterId, result.ClusterId)
		assert.Equal(t, expected.Cluster.Description, result.Description)
		assert.Equal(t, expected.Cluster.SSHKeyName, result.SSHKeyName)
		assert.Equal(t, expected.Cluster.Spec, result.Spec)
		assert.Equal(t, getTotalNodeCount(expected.Cluster.Nodes), len(result.Nodes))
		assert.Equal(t, len(expected.Cluster.StorageNodes), len(result.StorageNodes))

		// Test case teardown
		err = clearAllTrainingDatabase(db)
		assert.NoError(t, err)
	})

	t.Run("Get cluster details by ID - no records found", func(t *testing.T) {
		// Ensure nothing is in the database
		err := clearAllTrainingDatabase(db)
		assert.NoError(t, err)

		// Call the function under test
		result, err := GetClusterByID(ctx, db, &pb.SlurmClusterRequest{
			CloudAccountId: "",
			ClusterId:      "",
		})
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
