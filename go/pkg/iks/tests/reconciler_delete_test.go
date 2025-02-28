// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"testing"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	clusterUuidReconcilerDelete string
	nodegroupReconcilerUuid     string
)

func TestSetupForReconcilerDelete(t *testing.T) {
	var err error
	clusterUuidReconcilerDelete, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	err = setActionableState(clusterUuidReconcilerDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodegroupReconcilerUuid, err = insertNodeGroupIntoDB(t, clusterUuidReconcilerDelete)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
	clusterId := &pb.ClusterID{
		Clusteruuid:    clusterUuidReconcilerDelete,
		CloudAccountId: clusterCreate.CloudAccountId,
	}
	_, err = client.DeleteCluster(context.Background(), clusterId)
	if err != nil {
		t.Fatalf("Was not expecting error: %v", err.Error())
	}
}

func TestReconcilerDeleteNoClusterFound(t *testing.T) {
	// Set up
	clusterDeletionRequest := &pb.ClusterDeletionRequest{
		Uuid: "fake_uuid",
	}
	// Test
	_, err := reconcilerClient.DeleteClusterReconciler(context.Background(), clusterDeletionRequest)
	if err == nil {
		t.Fatalf("Was expecting error, none was received")
	}
	// Check
	expectedError := errorBaseDefault + "No cluster found"
	if expectedError != err.Error() {
		t.Fatalf("Reconciler Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}

func TestReconcilerDeleteSuccess(t *testing.T) {
	// Set up
	clusterDeletionRequest := &pb.ClusterDeletionRequest{
		Uuid: clusterUuidReconcilerDelete,
	}
	// Test
	_, err := reconcilerClient.DeleteClusterReconciler(context.Background(), clusterDeletionRequest)
	if err != nil {
		t.Fatalf("Was not expecting error: %v", err.Error())
	}
	// Check
	// Check
	sqlDB, err := managedDb.Open(context.Background())
	var (
		clusterState   string
		nodegroupCount int
		tempClusterId  int
	)
	err = sqlDB.QueryRow(`SELECT clusterstate_name, cluster_id FROM public.cluster WHERE unique_id = $1`,
		clusterDeletionRequest.Uuid,
	).Scan(&clusterState, &tempClusterId)
	if err != nil {
		t.Fatalf("Was not expecting error: %v", err.Error())
	}
	if clusterState != "Deleted" {
		t.Fatalf("Cluster not set to correct state")
	}
	err = sqlDB.QueryRow(`SELECT count(nodegroup_id) FROM public.nodegroup WHERE cluster_id = $1`,
		tempClusterId,
	).Scan(&nodegroupCount)
	if err != nil {
		t.Fatalf("Was not expecting error: %v", err.Error())
	}
	if nodegroupCount != 0 {
		t.Fatalf("Nodegroups were not deleted")
	}
}

func TestDeleteDBCleanupForReconcilerDelete(t *testing.T) {
	err := removeClusterFromDB(clusterUuidReconcilerDelete)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}
