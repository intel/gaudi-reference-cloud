// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"testing"

	//"github.com/DATA-DOG/go-sqlmock"
	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/query"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	clusterUuidPut   string
	nodeGroupUuidPut string
)

var newClusterName = "new_cluster_name"
var newDesc = "new desc"
var annotations = pb.Annotations{}
var updateClusterRequest = pb.UpdateClusterRequest{
	CloudAccountId: "iks_user",
	Description:    &newDesc,
	Name:           &newClusterName,
	Annotations: []*pb.Annotations{
		{
			Key:   "akey1",
			Value: "avalue1",
		},
		{
			Key:   "akey2",
			Value: "avalue2",
		},
	},
	Tags: []*pb.KeyValuePair{
		{
			Key:   "tkey1",
			Value: "tvalue1",
		},
		{
			Key:   "tkey2",
			Value: "tvalue2",
		},
	},
}

var updateSecRuleRequest = pb.UpdateFirewallRuleRequest{
	CloudAccountId: "iks_user",
	Sourceip:       []string{"10.50.60.80"},
	Internalip:     "10.20.30.40",
	Port:           443,
}

func setVipTable(vipid int32, clusteruuid string) error {
	// set vip entry for cluster id
	// set vip details for cluster id

	sqlDB, err := managedDb.Open(context.Background())

	_, err = sqlDB.Exec(`
			Update vip set vip_ip = '10.20.30.40', firewall_status= '' where vip_id = $1
		`, vipid)
	if err != nil {
		return err
	}

	_, err = sqlDB.Exec(`
		Update cluster set clusterstate_name = 'Pending' where unique_id = $1`, clusteruuid)
	if err != nil {
		return err
	}

	var vipi int32
	var port int32

	err = sqlDB.QueryRow(`select vip_id , port from vipdetails`).Scan(&vipi, &port)
	if err != nil {
		return err
	}
	fmt.Println("fromdetails", vipi, port)
	_, err = sqlDB.Exec(`
		Update vipdetails set port = '443' where vip_id = $1`, vipid)
	if err != nil {
		return err
	}

	/*
		_,err = sqlDB.Exec(`INSERT INTO public.vipdetails (vip_id, vip_name, description, port, pool_name, pool_port)
		VALUES ($1, 'test-vip', 'tst-vip', 80, 'vip-pool', 80)`,vipid)
		if err != nil {
			fmt.Println("error port",err)
			return err
		}*/

	defer sqlDB.Close()
	return nil
}

func TestSetupForClusterPutTests(t *testing.T) {
	var err error
	// Insert Cluster
	clusterUuidPut, err = insertClusterIntoDb(t)
	updateClusterRequest.Clusteruuid = clusterUuidPut
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
}

func TestPutClusterClusterExistance(t *testing.T) {
	// Set up
	fake := "fake_id"
	original := updateClusterRequest.Clusteruuid
	updateClusterRequest.Clusteruuid = fake
	// Test
	_, err := client.PutCluster(context.Background(), &updateClusterRequest)
	updateClusterRequest.Clusteruuid = original
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + fake
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}
func TestPutClusterClusterPermissions(t *testing.T) {
	// Set up
	original := updateClusterRequest.CloudAccountId
	updateClusterRequest.CloudAccountId = "fake_user"
	// Test
	_, err := client.PutCluster(context.Background(), &updateClusterRequest)
	updateClusterRequest.CloudAccountId = original
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + updateClusterRequest.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}

}
func TestPutClusterClusterActionable(t *testing.T) {
	// Set up
	// Test
	_, err := client.PutCluster(context.Background(), &updateClusterRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Cluster not in actionable state"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}

}
func TestPutClusterSuccess(t *testing.T) {
	// set up
	updateClusterRequest.Clusteruuid = clusterUuidPut
	err := setActionableState(updateClusterRequest.Clusteruuid)
	// test
	res, err := client.PutCluster(context.Background(), &updateClusterRequest)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		t.Fatalf("Could not open DB connection")
	}
	var clusterName string
	err = sqlDB.QueryRow(`SELECT name FROM public.cluster WHERE unique_id = $1`, updateClusterRequest.Clusteruuid).Scan(&clusterName)
	if err != nil {
		t.Fatalf("Could not query DB: %v", err.Error())
	}
	if res.Name != *updateClusterRequest.Name {
		t.Fatalf("Cluster response name not expected")
	}
	if clusterName != *updateClusterRequest.Name {
		t.Fatalf("DB name not expected")
	}
}

func TestPutClusterDuplicateAnnotations(t *testing.T) {
	// set up
	updateClusterRequest.Clusteruuid = clusterUuidPut
	original := updateClusterRequest.Annotations
	updateClusterRequest.Annotations = append(updateClusterRequest.Annotations, updateClusterRequest.Annotations[0])
	err := setActionableState(updateClusterRequest.Clusteruuid)
	// test
	_, err = client.PutCluster(context.Background(), &updateClusterRequest)
	updateClusterRequest.Annotations = original
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseInvalidArgument + "Annotations sent are duplicated. Correct and try again"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}

}
func TestPutClusterDuplicateTags(t *testing.T) {
	// set up
	updateClusterRequest.Clusteruuid = clusterUuidPut
	original := updateClusterRequest.Tags
	updateClusterRequest.Tags = append(updateClusterRequest.Tags, updateClusterRequest.Tags[0])
	err := setActionableState(updateClusterRequest.Clusteruuid)
	// test
	_, err = client.PutCluster(context.Background(), &updateClusterRequest)
	updateClusterRequest.Tags = original
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseInvalidArgument + "Tags sent are duplicated. Correct and try again"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}
func TestPutClusterDuplicateName(t *testing.T) {
	// set up
	tempUuid, err := insertClusterIntoDb(t)
	updateClusterRequest.Clusteruuid = clusterUuidPut
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	updateClusterRequest.Clusteruuid = clusterUuidPut
	err = setActionableState(updateClusterRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	err = setActionableState(tempUuid)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	original := updateClusterRequest.Name
	updateClusterRequest.Name = &clusterCreate.Name
	// test
	_, checkErr := client.PutCluster(context.Background(), &updateClusterRequest)
	err = removeClusterFromDB(tempUuid)
	if err != nil {
		t.Fatalf("Could not clean up entry")
	}
	updateClusterRequest.Name = original
	// Check
	expectedError := errorBaseAlreadyExists + "Cluster name already in use"
	if expectedError != checkErr.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}

func TestPutClustersDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

func TestSetupForClusterUpgradTests(t *testing.T) {
	var err error
	// Insert Cluster
	err = setK8sVersionActiveAndDepricate("1.28.5", "1.28.7")
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	clusterCreate.K8Sversionname = "1.27"
	clusterUuidPut, err = insertClusterIntoDb(t)
	clusterCreate.K8Sversionname = "1.28"
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	updateClusterRequest.Clusteruuid = clusterUuidPut
	// Insert SetActionable
	err = setActionableState(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodeGroupUuidPut, err = insertNodeGroupIntoDB(t, clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
	err = setK8sVersionActiveAndDepricate("1.28.7", "1.28.5")
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
}

var upgradeClusterRequest = pb.UpgradeClusterRequest{
	Clusteruuid:    clusterUuidPut,
	CloudAccountId: "iks_user",
}

func TestUpgradeClustersExistance(t *testing.T) {
	// Set up
	fake := "fake_id"
	original := updateClusterRequest.Clusteruuid
	upgradeClusterRequest.Clusteruuid = fake
	// Test
	_, err := client.UpgradeCluster(context.Background(), &upgradeClusterRequest)
	upgradeClusterRequest.Clusteruuid = original
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + fake
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}
func TestUpgradeClustersPermissions(t *testing.T) {
	// Set up
	original := updateClusterRequest.CloudAccountId
	upgradeClusterRequest.CloudAccountId = "fake_user"
	// Test
	_, err := client.UpgradeCluster(context.Background(), &upgradeClusterRequest)
	upgradeClusterRequest.CloudAccountId = original
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + updateClusterRequest.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}
func TestUpgradeClustersActionable(t *testing.T) {
	// Set up
	// Test
	_, err := client.UpgradeCluster(context.Background(), &upgradeClusterRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Cluster not in actionable state"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}

func TestUpgradeClustersNodeGroupActionable(t *testing.T) {
	// Set up
	updateClusterRequest.Clusteruuid = clusterUuidPut
	err := setActionableState(upgradeClusterRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Failed to set actionable stae: %v", err.Error())
	}
	// Test
	_, err = client.UpgradeCluster(context.Background(), &upgradeClusterRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Unable to upgrade cluster. Nodegroups is updating."
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}
func TestUpgradeClustersSuccess(t *testing.T) {
	// Set up
	updateClusterRequest.Clusteruuid = clusterUuidPut
	err := setActionableState(upgradeClusterRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Failed to set actionable stae: %v", err.Error())
	}
	err = setNodeGroupsActionableState(updateClusterRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Failed to set actionable stae: %v", err.Error())
	}
	// Test
	_, err = client.UpgradeCluster(context.Background(), &upgradeClusterRequest)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		t.Fatalf("Could not open DB connection")
	}
	var k8sVerisonName string
	// Check K8sversion
	err = sqlDB.QueryRow(`SELECT k8sversion_name FROM public.nodegroup WHERE unique_id =  $1`, nodeGroupUuidPut).Scan(&k8sVerisonName)
	if err != nil {
		t.Fatalf("Could not query DB: %v", err.Error())
	}
	if k8sVerisonName != "1.28.7" {
		t.Fatalf("Cluster not updated correctly")
	}
	// Check Values
	_, clusterCrd, err := utils.GetLatestClusterRev(context.Background(), sqlDB, clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not validate cluster CRD")
	}
	// Addon version
	if clusterCrd.Spec.Addons == nil || len(clusterCrd.Spec.Addons) == 0 {
		t.Fatalf("Could not validate cluster CRD Addons")
	}
	for _, addon := range clusterCrd.Spec.Addons {
		if addon.Name == "kube-proxy" && addon.Artifact != "s3://kube-proxy-1287-1.template" {
			t.Fatalf("Cluster Addons have not upgraded successfully")
		}
	}

}

func TestUpgradeClustersNoAvailableUpgrades(t *testing.T) {
	// Set up
	err := setK8sVersionActiveAndDepricate("1.29.2", "1.29.2")
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	updateClusterRequest.Clusteruuid = clusterUuidPut
	err = setActionableState(upgradeClusterRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Failed to set actionable stae: %v", err.Error())
	}
	err = setNodeGroupsActionableState(updateClusterRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Failed to set actionable stae: %v", err.Error())
	}
	// Test
	_, err = client.UpgradeCluster(context.Background(), &upgradeClusterRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "No Upgrades Available"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}

func TestUpgradeClustersDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

func TestNodeGroupForClusterUpgradTests(t *testing.T) {
	var err error
	// Insert Cluster
	clusterUuidPut, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	updateClusterRequest.Clusteruuid = clusterUuidPut
	// Insert SetActionable
	err = setActionableState(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodeGroupUuidPut, err = insertNodeGroupIntoDB(t, clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
	// Insert SetActionable
}

var count = int32(1)
var name = "new_name1"
var updateNodegroupRequest = pb.UpdateNodeGroupRequest{
	Clusteruuid:    "",
	Nodegroupuuid:  "",
	CloudAccountId: "iks_user",
	Count:          &count,
	Name:           &name,
	Upgradestrategy: &pb.UpgradeStrategy{
		Drainnodes:               true,
		Maxunavailablepercentage: 15,
	},
	Annotations: []*pb.Annotations{
		{
			Key:   "akey1",
			Value: "avalue1",
		},
		{
			Key:   "akey2",
			Value: "avalue2",
		},
	},
	Tags: []*pb.KeyValuePair{
		{
			Key:   "tkey1",
			Value: "tvalue1",
		},
		{
			Key:   "tkey2",
			Value: "tvalue2",
		},
	},
}

func TestNodeGroupPutExistance(t *testing.T) {
	// set
	updateNodegroupRequest.Clusteruuid = clusterUuidPut
	updateNodegroupRequest.Nodegroupuuid = "fake_id"
	updateNodegroupRequest.CloudAccountId = "iks_user"
	_, err := client.PutNodeGroup(context.Background(), &updateNodegroupRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	expectedError := errorBaseNotFound + "NodeGroup not found in Cluster: " + updateNodegroupRequest.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}
func TestNodeGroupPutPermissions(t *testing.T) {
	// set
	updateNodegroupRequest.Clusteruuid = clusterUuidPut
	updateNodegroupRequest.Nodegroupuuid = nodeGroupUuidPut
	updateNodegroupRequest.CloudAccountId = "fake_user"
	_, err := client.PutNodeGroup(context.Background(), &updateNodegroupRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	expectedError := errorBaseNotFound + "Cluster not found: " + updateClusterRequest.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}
func TestNodeGroupPutActionable(t *testing.T) {
	// set
	updateNodegroupRequest.Clusteruuid = clusterUuidPut
	updateNodegroupRequest.Nodegroupuuid = nodeGroupUuidPut
	updateNodegroupRequest.CloudAccountId = "iks_user"
	_, err := client.PutNodeGroup(context.Background(), &updateNodegroupRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	expectedError := errorBaseFailedPrecondition + "Cluster not in actionable state"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}

func TestNodeGroupPutNodeCountValidation(t *testing.T) {
	// Set cluster actionable
	err := setActionableState(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// set
	var count = int32(11)
	updateNodegroupRequest.Clusteruuid = clusterUuidPut
	updateNodegroupRequest.Nodegroupuuid = nodeGroupUuidPut
	updateNodegroupRequest.CloudAccountId = "iks_user"
	updateNodegroupRequest.Count = &count
	// db query to update nodecount = 10

	err = setnodescounttomax(nodeGroupUuidPut)
	if err != nil {
		t.Fatalf("Could not set up node count max: %v", err.Error())
	}
	_, err = client.PutNodeGroup(context.Background(), &updateNodegroupRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}

	expectedError := errorBaseFailedPrecondition + "nodegroup count should be between 0 and 10"
	if expectedError != err.Error() {
		t.Fatalf("Update Nodegroup: Expected count validation error, received: %v", err.Error())
	}
}

func TestNodeGroupPutNodeCountValidationAllowScaleToZero(t *testing.T) {
	// Set cluster actionable
	err := setActionableState(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// set
	var count = int32(0) // Set count to 0
	updateNodegroupRequest.Clusteruuid = clusterUuidPut
	updateNodegroupRequest.Nodegroupuuid = nodeGroupUuidPut
	updateNodegroupRequest.CloudAccountId = "iks_user"
	updateNodegroupRequest.Count = &count
	// db query to update nodecount = 0

	_, err = client.PutNodeGroup(context.Background(), &updateNodegroupRequest)
	if err != nil {
		t.Fatalf("Update Nodegroup: Expected to allow sending count=0 to allow scale to zero, received: %v", err.Error())
	}
}

func TestNodeGroupPutSuccess(t *testing.T) {
	// set
	updateNodegroupRequest.Clusteruuid = clusterUuidPut
	updateNodegroupRequest.Nodegroupuuid = nodeGroupUuidPut
	updateNodegroupRequest.CloudAccountId = "iks_user"
	var count = int32(1)
	updateNodegroupRequest.Count = &count

	// Set count back
	err := setnodecountback(nodeGroupUuidPut)
	if err != nil {
		t.Fatalf("Could not set up node count back: %v", err.Error())
	}
	// Test
	err = setActionableState(updateNodegroupRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Failed to set actionable stae: %v", err.Error())
	}
	res, err := client.PutNodeGroup(context.Background(), &updateNodegroupRequest)
	if err != nil {
		fmt.Println(err.Error())
	}
	// check
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		t.Fatalf("Could not open DB connection")
	}
	var nodeGroupName string
	err = sqlDB.QueryRow(`SELECT name FROM public.nodegroup WHERE unique_id = $1`, updateNodegroupRequest.Nodegroupuuid).Scan(&nodeGroupName)
	if err != nil {
		t.Fatalf("Could not query DB: %v", err.Error())
	}
	if res.Name != *updateNodegroupRequest.Name {
		t.Fatalf("Nodegroup response name not expected")
	}
	if nodeGroupName != *updateNodegroupRequest.Name {
		t.Fatalf("DB name not expected")
	}

}

func TestNodeGroupPutDuplicateAnnotations(t *testing.T) {
	// set
	updateNodegroupRequest.Clusteruuid = clusterUuidPut
	updateNodegroupRequest.Nodegroupuuid = nodeGroupUuidPut
	original := updateNodegroupRequest.Annotations
	updateNodegroupRequest.Annotations = append(updateNodegroupRequest.Annotations, updateNodegroupRequest.Annotations[0])
	// Test
	err := setActionableState(updateNodegroupRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Failed to set actionable stae: %v", err.Error())
	}
	_, err = client.PutNodeGroup(context.Background(), &updateNodegroupRequest)
	updateNodegroupRequest.Annotations = original
	if err != nil {
		fmt.Println(err.Error())
	}
	// check
	expectedError := errorBaseInvalidArgument + "Annotations sent are duplicated. Correct and try again"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}

func TestNodeGroupPutDuplicateTags(t *testing.T) {
	// set
	updateNodegroupRequest.Clusteruuid = clusterUuidPut
	updateNodegroupRequest.Nodegroupuuid = nodeGroupUuidPut
	original := updateNodegroupRequest.Tags
	updateNodegroupRequest.Tags = append(updateNodegroupRequest.Tags, updateNodegroupRequest.Tags[0])
	// Test
	err := setActionableState(updateNodegroupRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Failed to set actionable stae: %v", err.Error())
	}
	_, err = client.PutNodeGroup(context.Background(), &updateNodegroupRequest)
	updateNodegroupRequest.Tags = original
	if err != nil {
		fmt.Println(err.Error())
	}
	// check
	expectedError := errorBaseInvalidArgument + "Tags sent are duplicated. Correct and try again"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}

func TestNodeGroupClustersDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

func TestSetupForNodeGroupUpgradTests(t *testing.T) {
	var err error
	err = setK8sVersionActiveAndDepricate("1.28.5", "1.28.7")
	if err != nil {
		t.Fatalf("Could not set K8sVersions: %v", err.Error())
	}
	// Insert Cluster
	clusterUuidPut, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	updateClusterRequest.Clusteruuid = clusterUuidPut
	// Insert SetActionable
	err = setActionableState(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodeGroupUuidPut, err = insertNodeGroupIntoDB(t, clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
	// Upgrade controlplane
	err = setK8sVersionActiveAndDepricate("1.28.7", "1.28.5")
	if err != nil {
		t.Fatalf("Could not set K8sVersions: %v", err.Error())
	}
	err = setControlPlaneToK8sVersion(clusterUuidPut, "1.28.7")
	if err != nil {
		t.Fatalf("Could not set K8sVersions: %v", err.Error())
	}

}

var upgradeNodeGroupRequest = &pb.NodeGroupid{
	Clusteruuid:    clusterUuidPut,
	Nodegroupuuid:  nodeGroupUuidPut,
	CloudAccountId: "iks_user",
}

func TestNodeGroupUpgradeExistance(t *testing.T) {
	// set
	upgradeNodeGroupRequest.Clusteruuid = clusterUuidPut
	upgradeNodeGroupRequest.Nodegroupuuid = "fake_id"
	upgradeNodeGroupRequest.CloudAccountId = "iks_user"
	_, err := client.UpgradeNodeGroup(context.Background(), upgradeNodeGroupRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	expectedError := errorBaseNotFound + "NodeGroup not found in Cluster: " + upgradeNodeGroupRequest.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}
func TestNodeGroupUpgradePermissions(t *testing.T) {
	// set
	upgradeNodeGroupRequest.Clusteruuid = clusterUuidPut
	upgradeNodeGroupRequest.Nodegroupuuid = nodeGroupUuidPut
	upgradeNodeGroupRequest.CloudAccountId = "fake_user"
	_, err := client.UpgradeNodeGroup(context.Background(), upgradeNodeGroupRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	expectedError := errorBaseNotFound + "Cluster not found: " + updateClusterRequest.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}
func TestNodeGroupUpgradeActionable(t *testing.T) {
	// set
	upgradeNodeGroupRequest.Clusteruuid = clusterUuidPut
	upgradeNodeGroupRequest.Nodegroupuuid = nodeGroupUuidPut
	upgradeNodeGroupRequest.CloudAccountId = "iks_user"
	_, err := client.UpgradeNodeGroup(context.Background(), upgradeNodeGroupRequest)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	expectedError := errorBaseFailedPrecondition + "Cluster not in actionable state"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}
func TestNodeGroupUpgradeSuccess(t *testing.T) {
	// set up
	upgradeNodeGroupRequest.Clusteruuid = clusterUuidPut
	upgradeNodeGroupRequest.Nodegroupuuid = nodeGroupUuidPut
	// Insert SetActionable
	err := setActionableState(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	err = setNodeGroupsActionableState(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set nodegroup actionable state: %v", err.Error())
	}
	// Run
	_, err = client.UpgradeNodeGroup(context.Background(), upgradeNodeGroupRequest)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	sqlDB, err := managedDb.Open(context.Background())
	var k8sVersionName string
	if err != nil {
		t.Fatalf("Could not open DB connection")
	}
	err = sqlDB.QueryRow(`SELECT k8sversion_name FROM public.nodegroup WHERE unique_id = $1`, upgradeNodeGroupRequest.Nodegroupuuid).Scan(&k8sVersionName)
	if err != nil {
		t.Fatalf("Could not query DB: %v", err.Error())
	}

	fmt.Println(k8sVersionName)
	if k8sVersionName != "1.28.7" {
		t.Fatalf("Nodegroup response name not expected")
	}
}
func TestNodeGroupUpgradeNoUpgradesAvailable(t *testing.T) {
	// set up
	upgradeNodeGroupRequest.Clusteruuid = clusterUuidPut
	upgradeNodeGroupRequest.Nodegroupuuid = nodeGroupUuidPut
	// Insert SetActionable
	err := setActionableState(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	err = setNodeGroupsActionableState(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not set nodegroup actionable state: %v", err.Error())
	}
	// Run
	_, err = client.UpgradeNodeGroup(context.Background(), upgradeNodeGroupRequest)
	if err == nil {
		t.Fatalf("Should get error, nonw received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "No Upgrades Available"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}

}

/*
func TestUpdateSecurityRules(t *testing.T) {

	/*
	db , mock, err := sqlmock.New()
	if err != nil {
	 	t.Fatalf("An error occurred while creating mock: %s", err)
	}
	defer db.Close()*/
/*
	expectedRowsIp := mock.NewRows([]string{"vip_ip"}).AddRow("10.20.67.80")
	mock.ExpectQuery(query.GetvipidForFirewallQuery).WillReturnRows(expectedRowsIp)

	expectedRows := mock.NewRows([]string{"port"}).AddRow(80)
	mock.ExpectQuery(query.GetVipPortQuery).WillReturnRows(expectedRows)

 	_, err := insertVipIntoDB(clusterUuidPut)
 	if err != nil {
	 	t.Fatalf("Could not set up cluster vip: %v", err)
 	}


	err = setVipTable(63,clusterUuidPut)
	if err != nil {
		fmt.Println(err.Error())
	}


	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockCreateFirewallValues(mock, 1)
	mock.ExpectQuery(query.GetvipidForFirewallQuery)
	mockCreateFirewallValues(mock, 2)
	mock.ExpectQuery(query.GetVipPortQuery)


	updateSecRuleRequest.Clusteruuid = clusterUuidPut

	// Test
	err = setActionableState(updateSecRuleRequest.Clusteruuid)
	if err != nil {
		t.Fatalf("Failed to set actionable stae: %v", err.Error())
		fmt.Println(err.Error())
	}
	_, err = client.UpdateFirewallRule(context.Background(), &updateSecRuleRequest)
	if err != nil {
		fmt.Println(err.Error())
	}
	// check
	expectedError := errorBaseDefault + "Update firewall rule"
	if expectedError != err.Error() {
		t.Fatalf("update security rule, received: %v", err.Error())
	}
}*/

func TestNodeGroupUpgradeDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidPut)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
	err = setK8sVersionActiveAndDepricate("1.27.1", "1.27.4")
	if err != nil {
		t.Fatalf("Could not set K8sVersions: %v", err.Error())
	}
}
