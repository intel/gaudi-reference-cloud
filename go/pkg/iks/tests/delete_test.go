// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package tests

import (
	"context"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	db_query "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/db_query_constants"
	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/iks_utils"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db/query"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

var (
	clusterUuidDelete   string
	nodegroupUuidDelete string
	vipIdDelete         int32
)

/* CLUSTER UNIT TESTS */
func TestSetupForClusterDeletionTests(t *testing.T) {
	var err error
	// Insert Cluster
	clusterUuidDelete, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodegroupUuidDelete, err = insertNodeGroupIntoDB(t, clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
}

func TestFailureClusterNotFound(t *testing.T) {
	// Set Up
	clusterId := pb.ClusterID{
		Clusteruuid:    "fake_uuid",
		CloudAccountId: "iks_user",
	}
	// Test
	_, err := client.DeleteCluster(context.Background(), &clusterId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + clusterId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected existance error, received: %v", err.Error())
	}
}

func TestFailureOnPermissions(t *testing.T) {
	// Set Up
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "fake_user",
	}
	// Test
	_, err := client.DeleteCluster(context.Background(), &clusterId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + clusterId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}

func TestFailureOnAlreadyDeleting(t *testing.T) {
	// Set Up
	err := setDeletingState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up deleting state: %v", err.Error())
	}
	defer setActionableState(clusterUuidDelete) // setting back to actionable state for next tests

	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
	}

	// Test
	_, err = client.DeleteCluster(context.Background(), &clusterId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Cluster can not be deleted , it is currently in deleting state"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected already deleting error, received: %v", err.Error())
	}
}

func TestDeleteSuccess(t *testing.T) {
	// Set Up
	clusterId := &pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: clusterCreate.CloudAccountId,
	}
	// Test
	_, err := client.DeleteCluster(context.Background(), clusterId)
	if err != nil {
		t.Fatalf("Was not expecting error: %v", err.Error())
	}
	// Check
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		t.Fatalf("Could not open DB connection")
	}
	var (
		clusterState   string
		nodeGroupState string
		tempClusterId  int
	)
	err = sqlDB.QueryRow(`SELECT clusterstate_name, cluster_id FROM public.cluster WHERE unique_id = $1`,
		clusterId.Clusteruuid,
	).Scan(&clusterState, &tempClusterId)
	if err != nil {
		t.Fatalf("Should not get error: %v", err.Error())
	}
	if clusterState != "DeletePending" {
		t.Fatalf("Cluster not set to correct state")
	}
	err = sqlDB.QueryRow(`SELECT nodegroupstate_name FROM public.nodegroup WHERE cluster_id = $1`,
		tempClusterId,
	).Scan(&nodeGroupState)
	if err != nil {
		t.Fatalf("Was not expecting error: %v", err.Error())
	}
	if nodeGroupState != "Deleting" {
		t.Fatalf("Nodegroup not set to correct state")
	}
}

func TestDeleteDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

/* NODEGROUP UNIT TESTS */

func TestSetupForNodeGroupTests(t *testing.T) {
	var err error
	// Insert Cluster
	clusterUuidDelete, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodegroupUuidDelete, err = insertNodeGroupIntoDB(t, clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
}

func TestDeleteNodegroupExistence(t *testing.T) {
	// Set Up
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    "fake_uuid",
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}
	// Test
	_, err := client.DeleteNodeGroup(context.Background(), &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "NodeGroup not found in Cluster: " + nodeGroupId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}

}

func TestDeleteNodeGroupActionable(t *testing.T) {
	// Set Up
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}
	// Test
	_, err := client.DeleteNodeGroup(context.Background(), &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Cluster not in actionable state"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected preconfition error, received: %v", err.Error())
	}
}

func TestDeleteNodeGroupPermissions(t *testing.T) {
	// Set Up
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "fake_user",
	}
	// Test
	_, err := client.DeleteNodeGroup(context.Background(), &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + nodeGroupId.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found error, received: %v", err.Error())
	}
}

func TestDeleteNodeGroupAlreadyDeleting(t *testing.T) {
	// Set Up
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}

	err := setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}

	err = setNodeGroupsDeletingState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up nodegroup to deleting state: %v", err.Error())
	}

	defer setNodeGroupsActionableState(clusterUuidDelete) // setting back to actionable state for next tests

	// Test
	_, err = client.DeleteNodeGroup(context.Background(), &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Cannot delete nodegroup in deleting state"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected already deleting error: %s, received: %v", expectedError, err.Error())
	}
}

func TestDeleteNodeGroupSuccess(t *testing.T) {
	// Set up
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}

	// Test
	_, err := client.DeleteNodeGroup(context.Background(), &nodeGroupId)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		t.Fatalf("Could not open DB connection")
	}
	var (
		clusterState   string
		nodeGroupState string
		tempClusterId  int
	)
	err = sqlDB.QueryRow(`SELECT clusterstate_name, cluster_id FROM public.cluster WHERE unique_id = $1`,
		nodeGroupId.Clusteruuid,
	).Scan(&clusterState, &tempClusterId)
	if err != nil {
		t.Fatalf("Should not get error: %v", err.Error())
	}
	if clusterState != "Pending" {
		t.Fatalf("Cluster not set to correct state")
	}
	err = sqlDB.QueryRow(`SELECT nodegroupstate_name FROM public.nodegroup WHERE unique_id = $1`,
		nodeGroupId.Nodegroupuuid,
	).Scan(&nodeGroupState)
	if err != nil {
		t.Fatalf("Was not expecting error: %v", err.Error())
	}
	if nodeGroupState != "Deleting" {
		t.Fatalf("Nodegroup not set to correct state")
	}
}

func TestDeleteNodeGroupDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

/* VIP UNIT TESTS */
func TestSetupFoVipTests(t *testing.T) {
	var err error
	// Insert Cluster
	clusterUuidDelete, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodegroupUuidDelete, err = insertNodeGroupIntoDB(t, clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
	// Insert SetActionable
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	vipIdDelete, err = insertVipIntoDB(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up cluster vip: %v", err)
	}
}

func TestDeleteVipExistance(t *testing.T) {
	// Set Up
	vipCreate := pb.VipId{
		Clusteruuid:    "fake_uuid",
		Vipid:          vipIdDelete,
		CloudAccountId: "iks_user",
	}
	// Test
	_, err := client.DeleteVip(context.Background(), &vipCreate)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + vipCreate.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected permission error, received: %v", err.Error())
	}
}

func TestDeleteVipPermissions(t *testing.T) {
	// Set up
	vipCreate := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		Vipid:          vipIdDelete,
		CloudAccountId: "fake_user",
	}
	// Test
	_, err := client.DeleteVip(context.Background(), &vipCreate)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseNotFound + "Cluster not found: " + vipCreate.Clusteruuid
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected not found, received: %v", err.Error())
	}
}

func TestDeleteVipActionable(t *testing.T) {
	// Set Up
	vipCreate := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		Vipid:          vipIdDelete,
		CloudAccountId: "iks_user",
	}
	defer setActionableState(clusterUuidDelete) // setting back to actionable state for next tests

	// Test
	_, err := client.DeleteVip(context.Background(), &vipCreate)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Cluster not in actionable state"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected actionable error: %s, received: %v", expectedError, err.Error())
	}

}

func TestDeleteVipAlreadyDeleting(t *testing.T) {
	// Set Up
	vipCreate := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		Vipid:          vipIdDelete,
		CloudAccountId: "iks_user",
	}
	err := setVipDeletingState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up deleting state: %v", err.Error())
	}
	defer setVipActiveState(clusterUuidDelete) // setting back to actionable state for next tests
	// Test
	_, err = client.DeleteVip(context.Background(), &vipCreate)
	if err == nil {
		t.Fatalf("Should get error, none was received")
	}
	// Check
	expectedError := errorBaseFailedPrecondition + "Vip can not be deleted, it is currently in deleting state"
	if expectedError != err.Error() {
		t.Fatalf("Delete Cluster: Expected deleting error: %s, received: %v", expectedError, err.Error())
	}

}

func TestDeleteVipSuccess(t *testing.T) {
	/*
		// MOCK SETUP
		db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
		if err != nil {
			t.Fatalf("Could not mock database connection")
		}
		defer db.Close()*/

	// Set up
	vipCreate := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		Vipid:          vipIdDelete,
		CloudAccountId: "iks_user",
	}

	err := setVipTable(vipIdDelete, clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up vip for firewall: %v", err.Error())
	}
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Test
	_, err = client.DeleteVip(context.Background(), &vipCreate)
	if err != nil {
		t.Fatalf("Should not get error, received: %v", err.Error())
	}
	// Check
	sqlDB, err := managedDb.Open(context.Background())
	if err != nil {
		t.Fatalf("Could not open DB connection")
	}
	var (
		clusterState  string
		tempClusterId int
	)
	err = sqlDB.QueryRow(`SELECT clusterstate_name, cluster_id FROM public.cluster WHERE unique_id = $1`,
		vipCreate.Clusteruuid,
	).Scan(&clusterState, &tempClusterId)
	if err != nil {
		t.Fatalf("Should not get error: %v", err.Error())
	}

}

func TestDeleteVipDBCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

// MOCKS DELETE

func TestSetupForMockClusterDeletionTests(t *testing.T) {
	var err error
	// Insert Cluster
	clusterUuidDelete, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
}

func mockClusterDeletionValues(mock sqlmock.Sqlmock, stopLocation int) {
	rows := sqlmock.NewRows([]string{"cluster_id"}).AddRow(1)
	mock.ExpectQuery(utils.ClusterExistanceQuery).WillReturnRows(rows)
	if stopLocation == 1 {
		return
	}
	rows = sqlmock.NewRows([]string{"cloudaccount_id"}).AddRow("iks_user")
	mock.ExpectQuery(utils.GetClusterCloudAccountQuery).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"clusterstate_name"}).AddRow("updating")
	mock.ExpectQuery(utils.GetClusterStateQuery).WillReturnRows(rows)
	if stopLocation == 2 {
		return
	}
	mock.ExpectBegin()
	rows = sqlmock.NewRows([]string{})
	mock.ExpectQuery(query.SetAllClusterNodegroupsState).WillReturnRows(rows)
	if stopLocation == 3 {
		return
	}
	mock.ExpectQuery(query.SetAllClusterVipsState).WillReturnRows(rows)
	if stopLocation == 4 {
		return
	}
	mock.ExpectQuery(query.SetAllClusterAddonsState).WillReturnRows(rows)
	if stopLocation == 5 {
		return
	}
	mock.ExpectQuery(query.SetClusterState).WillReturnRows(rows)
	if stopLocation == 6 {
		return
	}
	mock.ExpectQuery(query.InsertProvisioningQuery).WillReturnRows(rows)
	mock.ExpectCommit()
}

func TestMockDeleteClusterValidateClusterExistence(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mock.ExpectQuery(utils.ClusterExistanceQuery)
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
	}
	// RUN
	_, err = query.DeleteRecord(context.Background(), db, &clusterId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not delete Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockDeleteClusterValidateClusterCloudAccount(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterDeletionValues(mock, 1)
	mock.ExpectQuery(utils.GetClusterCloudAccountQuery)
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
	}
	// RUN
	_, err = query.DeleteRecord(context.Background(), db, &clusterId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not delete Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockDeleteClusterSetAllClusterNodegroupsState(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterDeletionValues(mock, 2)
	mock.ExpectBegin()
	mock.ExpectQuery(query.SetAllClusterNodegroupsState)
	mock.ExpectRollback()
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
	}
	// RUN
	_, err = query.DeleteRecord(context.Background(), db, &clusterId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not delete Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockDeleteClusterSetAllClusterVipsState(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterDeletionValues(mock, 3)
	mock.ExpectQuery(query.SetAllClusterVipsState)
	mock.ExpectRollback()
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
	}
	// RUN
	_, err = query.DeleteRecord(context.Background(), db, &clusterId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not delete Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockDeleteClusterSetAllClusterAddonsState(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterDeletionValues(mock, 4)
	mock.ExpectQuery(query.SetAllClusterAddonsState)
	mock.ExpectRollback()
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
	}
	// RUN
	_, err = query.DeleteRecord(context.Background(), db, &clusterId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not delete Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockDeleteClusterSetClusterState(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterDeletionValues(mock, 5)
	mock.ExpectQuery(query.SetClusterState)
	mock.ExpectRollback()
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
	}
	// RUN
	_, err = query.DeleteRecord(context.Background(), db, &clusterId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not delete Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}

func TestMockDeleteClusterInsertProvisioning(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterDeletionValues(mock, 6)
	mock.ExpectQuery(query.InsertProvisioningQuery)
	mock.ExpectRollback()
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
	}
	// RUN
	_, err = query.DeleteRecord(context.Background(), db, &clusterId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	// CHECK
	expectedErrorMessage := "Could not delete Cluster. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockDeleteClusterCommit(t *testing.T) {
	// MOCK SETUP
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterDeletionValues(mock, -1)
	clusterId := pb.ClusterID{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
	}
	// RUN
	_, err = query.DeleteRecord(context.Background(), db, &clusterId)
	if err != nil {
		t.Fatalf("Should not get Error, recieved: %v", err.Error())
	}
	// CHECK
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}
}
func TestMockDeleteClusterCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

func TestSetupForMockNodeGroupDeletionTests(t *testing.T) {
	var err error
	// Insert Cluster
	clusterUuidDelete, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodegroupUuidDelete, err = insertNodeGroupIntoDB(t, clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
}

func mockClusterNodeGroupValues(mock sqlmock.Sqlmock, stopLocation int) {
	rows := sqlmock.NewRows([]string{"cluster_id", "nodegroup_id"}).AddRow(1, 1)
	mock.ExpectQuery(utils.NodeGroupExistanceQuery).WillReturnRows(rows)
	if stopLocation == 1 {
		return
	}
	rows = sqlmock.NewRows([]string{"cloudaccount_id"}).AddRow("iks_user")
	mock.ExpectQuery(utils.GetClusterCloudAccountQuery).WillReturnRows(rows)
	if stopLocation == 2 {
		return
	}
	rows = sqlmock.NewRows([]string{"clusterstate_name"}).AddRow("Active")
	mock.ExpectQuery(utils.GetClusterStateQuery).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"change_applied"}).AddRow(true)
	mock.ExpectQuery(utils.GetClusterRevChangeAppliedQuery).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"cluster_id", "clusterType"}).AddRow(1, "generalpurpose")
	mock.ExpectQuery(db_query.GetClusterTypeQuery).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"nodegroupstate_name"}).AddRow("Updating")
	mock.ExpectQuery(utils.GetNodeGroupStateQuery).WillReturnRows(rows)

	if stopLocation == 3 {
		return
	}
	rows = sqlmock.NewRows([]string{}).AddRow()
	mock.ExpectBegin()
	mock.ExpectQuery(query.UpdateNodegroupStateQuery).WillReturnRows(rows)
	if stopLocation == 4 {
		return
	}
	mock.ExpectQuery(query.UpdateClusterStateQuery).WillReturnRows(rows)
	if stopLocation == 5 {
		return
	}
	desiredSpecJson := fmt.Sprintf(`{"metadata":{"name": "%v","annotations":{}},"spec":{"kubernetesVersion":"1.27.1","instanceType":"small","instanceIMI":"image-2vzjd","runtime":"Containerd","sshKey":["test-key"],"runtimeArgs":{},"kubernetesProvider":"iks","kubernetesProviderConfig":{},"nodeProviderConfig":{"cloudaccountid":"013635516463"},"nodeProvider":"Harvester","network":{"serviceCIDR":"100.66.0.0/16","podCIDR":"100.68.0.0/16","clusterDNS":"100.66.0.10","region":"us-region-1"},"nodegroups":[{"name":"%v","kubernetesVersion":"1.27.1","instanceType":"small","instanceIMI":"image-2vzjd","sshKey":["test-key"],"count":1,"upgradeStrategy":{"maxUnavailablePercent":10,"drainBefore":true},"taints":{},"labels":{},"annotations":{},"vnets":null}],"addons":[{"name":"kube-proxy","type":"kubectl-apply","artifact":"http://10.11.183.185:8081/kube-proxy-k1271-1.template"},{"name":"coredns","type":"kubectl-apply","artifact":"http://10.11.183.185:8081/coredns-171-k1271-1.template"},{"name":"calico-operator","type":"kubectl-replace","artifact":"http://10.11.183.185:8081/calico-operator-3260-k1271-1.template"},{"name":"calico-config","type":"kubectl-apply","artifact":"http://10.11.183.185:8081/calico-config-3260-k1271-1.template"},{"name":"konnectivity-agent","type":"kubectl-apply","artifact":"http://10.11.183.185:8081/konnectivity-agent.template"}],"ilbs":[{"name":"etcd","description":"Etcd members loadbalancer","port":443,"iptype":"private","persist":"","ipprotocol":"tcp","environment":8,"usergroup":1545,"pool":{"name":"etcd","description":"","port":2379,"loadBalancingMode":"least-connections-member","minActiveMembers":1,"monitor":"i_tcp","memberConnectionLimit":0,"memberPriorityGroup":0,"memberRatio":1,"memberAdminState":"enabled"},"owner":"system"},{"name":"apiserver","description":"Kube-apiserver members loadbalancer","port":443,"iptype":"private","persist":"","ipprotocol":"tcp","environment":8,"usergroup":1545,"pool":{"name":"apiserver","description":"","port":6443,"loadBalancingMode":"least-connections-member","minActiveMembers":1,"monitor":"i_tcp","memberConnectionLimit":0,"memberPriorityGroup":0,"memberRatio":1,"memberAdminState":"enabled"},"owner":"system"},{"name":"konnectivity","description":"Konnectivity members loadbalancer","port":443,"iptype":"private","persist":"","ipprotocol":"tcp","environment":8,"usergroup":1545,"pool":{"name":"konnectivity","description":"","port":8132,"loadBalancingMode":"least-connections-member","minActiveMembers":1,"monitor":"i_tcp","memberConnectionLimit":0,"memberPriorityGroup":0,"memberRatio":1,"memberAdminState":"enabled"},"owner":"system"}],"backup":"","advancedConfig":{"kubeApiServerArgs":"","kubeControllerManagerArgs":"","kubeSchedulerArgs":"","kubeProxyArgs":"","kubeletArgs":""},"vnets":[{"availabilityzone":"us-region-1a-default","networkvnet":"us-region-1a"}]}}`, clusterUuidDelete, nodegroupUuidDelete)
	rows = sqlmock.NewRows([]string{"desiredspec_json"}).AddRow(desiredSpecJson)
	mock.ExpectQuery(utils.GetLatestClusterRevQuery).WillReturnRows(rows)
	if stopLocation == 6 {
		return
	}
	rows = sqlmock.NewRows([]string{"clusterrev_id"}).AddRow(1)
	mock.ExpectQuery(query.InsertRevQuery).WillReturnRows(rows)
	mock.ExpectCommit()
}
func TestMockDeleteNodeGroupExistance(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}
	mock.ExpectQuery(utils.NodeGroupExistanceQuery)
	// RUN
	_, err = query.DeleteNodeGroupRecord(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	expectedErrorMessage := "Could not delete Cluster Node Group. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteNodeGroupCloudAccount(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterNodeGroupValues(mock, 1)
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}
	mock.ExpectQuery(utils.GetClusterCloudAccountQuery)
	// RUN
	_, err = query.DeleteNodeGroupRecord(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	expectedErrorMessage := "Could not delete Cluster Node Group. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteNodeGroupValidateActionable(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterNodeGroupValues(mock, 2)
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}
	mock.ExpectQuery(utils.GetClusterStateQuery)
	// RUN
	_, err = query.DeleteNodeGroupRecord(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	expectedErrorMessage := "Could not delete Cluster Node Group. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteNodeGroupUpdateNodegroupStateQuery(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterNodeGroupValues(mock, 3)
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(query.UpdateNodegroupStateQuery)
	mock.ExpectRollback()

	// RUN
	_, err = query.DeleteNodeGroupRecord(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	expectedErrorMessage := "Could not delete Cluster Node Group. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteNodeGroupUpdateClusterStateQuery(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterNodeGroupValues(mock, 4)
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}
	mock.ExpectQuery(query.UpdateClusterStateQuery)
	mock.ExpectRollback()
	// RUN
	_, err = query.DeleteNodeGroupRecord(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	expectedErrorMessage := "Could not delete Cluster Node Group. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteNodeGroupGetLatestClusterRevQuery(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterNodeGroupValues(mock, 5)
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}
	mock.ExpectQuery(utils.GetLatestClusterRevQuery)
	mock.ExpectRollback()
	// RUN
	_, err = query.DeleteNodeGroupRecord(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	expectedErrorMessage := "Could not delete Cluster Node Group. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteNodeGroupInsertRevQuery(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterNodeGroupValues(mock, 6)
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}
	mock.ExpectQuery(query.InsertRevQuery)
	mock.ExpectRollback()
	// RUN
	_, err = query.DeleteNodeGroupRecord(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	expectedErrorMessage := "Could not delete Cluster Node Group. Please try again."
	if expectedErrorMessage != err.Error() {
		t.Fatalf("Expected a generic mismatch error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}

func TestMockDeleteNodeGroupCommit(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockClusterNodeGroupValues(mock, -1)
	nodeGroupId := pb.NodeGroupid{
		Clusteruuid:    clusterUuidDelete,
		Nodegroupuuid:  nodegroupUuidDelete,
		CloudAccountId: "iks_user",
	}
	// RUN
	_, err = query.DeleteNodeGroupRecord(context.Background(), db, &nodeGroupId)
	if err != nil {
		t.Fatalf("Should not get Error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}

func TestMockDeleteNodegroupCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

func TestSetupForMockVipDeletionTests(t *testing.T) {
	var err error
	// Insert Cluster
	clusterUuidDelete, err = insertClusterIntoDb(t)
	if err != nil {
		t.Fatalf("Could not set up cluster: %v", err)
	}
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert Nodegroup
	nodegroupUuidDelete, err = insertNodeGroupIntoDB(t, clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up cluster nodegroup: %v", err)
	}
	// Insert SetActionable
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
	// Insert VIP
	vipIdDelete, err = insertVipIntoDB(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up cluster vip: %v", err)
	}
	err = setActionableState(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not set up actionable state: %v", err.Error())
	}
}

func mockDeleteClusterVipValues(mock sqlmock.Sqlmock, stopLocation int) {
	rows := sqlmock.NewRows([]string{"cluster_id"}).AddRow(1)
	mock.ExpectQuery(utils.ClusterExistanceQuery).WillReturnRows(rows)
	if stopLocation == 1 {
		return
	}
	rows = sqlmock.NewRows([]string{"cloudaccount_id"}).AddRow("iks_user")
	mock.ExpectQuery(utils.GetClusterCloudAccountQuery).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"cluster_id", "vip_id"}).AddRow(1, 1)
	mock.ExpectQuery(utils.VipExistanceQuery + "AND v.owner = $3").WillReturnRows(rows)

	if stopLocation == 2 {
		return
	}
	rows = sqlmock.NewRows([]string{"clusterstate_name"}).AddRow("Active")
	mock.ExpectQuery(utils.GetClusterStateQuery).WillReturnRows(rows)
	rows = sqlmock.NewRows([]string{"change_applied"}).AddRow(true)
	mock.ExpectQuery(utils.GetClusterRevChangeAppliedQuery).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"vipstate_name"}).AddRow("Active")
	mock.ExpectQuery(utils.GetVipStateQuery).WillReturnRows(rows)

	rows = sqlmock.NewRows([]string{"firewall_status"}).AddRow("Active")
	mock.ExpectQuery(utils.GetSecStateQuery).WillReturnRows(rows)

	if stopLocation == 3 {
		return
	}
	rows = sqlmock.NewRows([]string{}).AddRow()
	mock.ExpectBegin()
	mock.ExpectQuery(query.UpdateClusterStateQuery).WillReturnRows(rows)
	if stopLocation == 4 {
		return
	}
	mock.ExpectQuery(query.DeleteVipRecordQuery).WillReturnRows(rows)
	if stopLocation == 5 {
		return
	}
	desiredSpecJson := fmt.Sprintf(`{"metadata":{"name": "%v","annotations":{}},"spec":{"kubernetesVersion":"1.27.1","instanceType":"small","instanceIMI":"image-2vzjd","runtime":"Containerd","sshKey":["test-key"],"runtimeArgs":{},"kubernetesProvider":"iks","kubernetesProviderConfig":{},"nodeProviderConfig":{"cloudaccountid":"013635516463"},"nodeProvider":"Harvester","network":{"serviceCIDR":"100.66.0.0/16","podCIDR":"100.68.0.0/16","clusterDNS":"100.66.0.10","region":"us-region-1"},"nodegroups":[{"name":"%v","kubernetesVersion":"1.27.1","instanceType":"small","instanceIMI":"image-2vzjd","sshKey":["test-key"],"count":1,"upgradeStrategy":{"maxUnavailablePercent":10,"drainBefore":true},"taints":{},"labels":{},"annotations":{},"vnets":null}],"addons":[{"name":"kube-proxy","type":"kubectl-apply","artifact":"http://10.11.183.185:8081/kube-proxy-k1271-1.template"},{"name":"coredns","type":"kubectl-apply","artifact":"http://10.11.183.185:8081/coredns-171-k1271-1.template"},{"name":"calico-operator","type":"kubectl-replace","artifact":"http://10.11.183.185:8081/calico-operator-3260-k1271-1.template"},{"name":"calico-config","type":"kubectl-apply","artifact":"http://10.11.183.185:8081/calico-config-3260-k1271-1.template"},{"name":"konnectivity-agent","type":"kubectl-apply","artifact":"http://10.11.183.185:8081/konnectivity-agent.template"}],"ilbs":[{"name":"%v","description":"Etcd members loadbalancer","port":443,"iptype":"private","persist":"","ipprotocol":"tcp","environment":8,"usergroup":1545,"pool":{"name":"etcd","description":"","port":2379,"loadBalancingMode":"least-connections-member","minActiveMembers":1,"monitor":"i_tcp","memberConnectionLimit":0,"memberPriorityGroup":0,"memberRatio":1,"memberAdminState":"enabled"},"owner":"system"},{"name":"apiserver","description":"Kube-apiserver members loadbalancer","port":443,"iptype":"private","persist":"","ipprotocol":"tcp","environment":8,"usergroup":1545,"pool":{"name":"apiserver","description":"","port":6443,"loadBalancingMode":"least-connections-member","minActiveMembers":1,"monitor":"i_tcp","memberConnectionLimit":0,"memberPriorityGroup":0,"memberRatio":1,"memberAdminState":"enabled"},"owner":"system"},{"name":"konnectivity","description":"Konnectivity members loadbalancer","port":443,"iptype":"private","persist":"","ipprotocol":"tcp","environment":8,"usergroup":1545,"pool":{"name":"konnectivity","description":"","port":8132,"loadBalancingMode":"least-connections-member","minActiveMembers":1,"monitor":"i_tcp","memberConnectionLimit":0,"memberPriorityGroup":0,"memberRatio":1,"memberAdminState":"enabled"},"owner":"system"}],"backup":"","advancedConfig":{"kubeApiServerArgs":"","kubeControllerManagerArgs":"","kubeSchedulerArgs":"","kubeProxyArgs":"","kubeletArgs":""},"vnets":[{"availabilityzone":"us-region-1a-default","networkvnet":"us-region-1a"}],"firewall":[{"destinationIp":"146.152.227.74","port":80,"protocol":"TCP","sourceips":["100.67.89.90","120.56.65.78","190.90.87.10"]}]}}`, clusterUuidDelete, nodegroupUuidDelete, vipCreate.Name)
	rows = sqlmock.NewRows([]string{"desiredspec_json"}).AddRow(desiredSpecJson)
	mock.ExpectQuery(utils.GetLatestClusterRevQuery).WillReturnRows(rows)
	if stopLocation == 6 {
		return
	}
	rows = sqlmock.NewRows([]string{"vip_name"}).AddRow(vipCreate.Name)
	mock.ExpectQuery(query.GetVipNameQuery).WillReturnRows(rows)

	if stopLocation == 7 {
		return
	}
	rows = sqlmock.NewRows([]string{"vip_ip"}).AddRow("10.20.30.40")
	mock.ExpectQuery(query.GetdestipQuery).WillReturnRows(rows)

	if stopLocation == 8 {
		return
	}

	mock.ExpectQuery(query.DeleteFirewallRuleQuery).WillReturnRows(rows)
	if stopLocation == 9 {
		return
	}

	rows = sqlmock.NewRows([]string{"clusterrev_id"}).AddRow(1)
	mock.ExpectQuery(query.InsertRevQuery).WillReturnRows(rows)
	mock.ExpectCommit()
}
func TestMockDeleteVipClusterExistance(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mock.ExpectQuery(utils.ClusterExistanceQuery)
	nodeGroupId := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
		Vipid:          vipIdDelete,
	}
	// RUN
	_, err = query.DeleteVip(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteVipCloudAccount(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockDeleteClusterVipValues(mock, 1)
	mock.ExpectQuery(utils.GetClusterCloudAccountQuery)
	nodeGroupId := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
		Vipid:          vipIdDelete,
	}
	// RUN
	_, err = query.DeleteVip(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteVipActionable(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockDeleteClusterVipValues(mock, 2)
	mock.ExpectQuery(utils.GetClusterStateQuery)
	nodeGroupId := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
		Vipid:          vipIdDelete,
	}
	// RUN
	_, err = query.DeleteVip(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteVipUpdateClusterStateQuery(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockDeleteClusterVipValues(mock, 3)
	mock.ExpectBegin()
	mock.ExpectQuery(query.UpdateClusterStateQuery)
	nodeGroupId := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
		Vipid:          vipIdDelete,
	}
	// RUN
	_, err = query.DeleteVip(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteVipDeleteVipRecordQuery(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockDeleteClusterVipValues(mock, 4)
	mock.ExpectQuery(query.DeleteVipRecordQuery)
	mock.ExpectRollback()
	nodeGroupId := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
		Vipid:          vipIdDelete,
	}
	// RUN
	_, err = query.DeleteVip(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteVipGetLatesClusterRev(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockDeleteClusterVipValues(mock, 5)
	mock.ExpectQuery(utils.GetLatestClusterRevQuery)
	mock.ExpectRollback()
	nodeGroupId := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
		Vipid:          vipIdDelete,
	}
	// RUN
	_, err = query.DeleteVip(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}
func TestMockDeleteVipGetVipNameQuery(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockDeleteClusterVipValues(mock, 6)
	mock.ExpectQuery(query.GetVipNameQuery)
	mock.ExpectRollback()
	nodeGroupId := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
		Vipid:          vipIdDelete,
	}
	// RUN
	_, err = query.DeleteVip(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}

/*
func TestMockDeleteVipInsertRevQuery(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()

	mock.ExpectQuery(utils.ClusterExistanceQuery)

	mockDeleteClusterVipValues(mock, 9)
	mock.ExpectQuery(query.InsertRevQuery)
	mock.ExpectRollback()
	nodeGroupId := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
		Vipid:          vipIdDelete,
	}

	// RUN
	_, err = query.DeleteVip(context.Background(), db, &nodeGroupId)
	if err == nil {
		t.Fatalf("Should get Error, none was recieved")
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}*/

func TestMockDeleteVipCommit(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Could not mock database connection")
	}
	defer db.Close()
	mockDeleteClusterVipValues(mock, -1)
	nodeGroupId := pb.VipId{
		Clusteruuid:    clusterUuidDelete,
		CloudAccountId: "iks_user",
		Vipid:          vipIdDelete,
	}
	// RUN
	_, err = query.DeleteVip(context.Background(), db, &nodeGroupId)
	if err != nil {
		t.Fatalf("Should not get Error, recieved: %v", err.Error())
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatalf("Mock expectations were not met: %v", err)
	}

}

func TestMockDeleteVipsCleanup(t *testing.T) {
	err := removeClusterFromDB(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}

func TestDeleteNodeGroupInstanceWithoutDownsize(t *testing.T) {
	_ = setRestrictedToFalse()
	// Insert Cluster
	clusterId, err := insertNewClusterIntoDb(t)
	assert.Nil(t, err)
	// Set Actionable
	assert.Nil(t, setActionableState(clusterId))
	// Insert Nodegroup
	nodeGroupId, err := insertNewNodeGroupIntoDB(t, clusterId)
	assert.Nil(t, err)
	// Set Actionable after nodegroup insert
	assert.Nil(t, setActionableState(clusterId))
	downsize := false
	// Initialize compute client mocking
	req := &pb.DeleteNodeGroupInstanceRequest{
		Clusteruuid:    clusterId,
		Nodegroupuuid:  nodeGroupId,
		Downsize:       &downsize,
		CloudAccountId: "iks_user",
		InstanceName:   "nodegroup1",
	}
	ctx := context.Background()
	iksSrv, err := InitInstanceServiceMockClient(t, true, codes.OK, "all is good!")
	assert.Nil(t, err)
	nodeGroupStatus, err := iksSrv.DeleteNodeGroupInstance(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, nodeGroupStatus.Count, nodeGroupCreate.Count)
	assert.Nil(t, deleteClusterFromDb(t, clusterId))
}

func TestDeleteNodeGroupInstanceWithWrongNodeGroup(t *testing.T) {
	_ = setRestrictedToFalse()
	// Insert Cluster
	clusterId, err := insertNewClusterIntoDb(t)
	assert.Nil(t, err)
	// Set Actionable
	assert.Nil(t, setActionableState(clusterId))
	// Insert Nodegroup
	_, err = insertNewNodeGroupIntoDB(t, clusterId)
	assert.Nil(t, err)
	// Set Actionable after nodegroup insert
	assert.Nil(t, setActionableState(clusterId))
	// Initialize compute client mocking
	req := &pb.DeleteNodeGroupInstanceRequest{
		Clusteruuid:    clusterId,
		Nodegroupuuid:  "non-existing-nodegroup",
		CloudAccountId: "iks_user",
		InstanceName:   "test-name",
	}
	ctx := context.Background()
	iksSrv, err := InitInstanceServiceMockClient(t, false, codes.OK, "")
	assert.Nil(t, err)
	_, err = iksSrv.DeleteNodeGroupInstance(ctx, req)
	assert.Equal(t, status.Code(err), codes.NotFound)
	assert.Nil(t, deleteClusterFromDb(t, clusterId))
}

func TestDeleteNodeGroupInstanceWithDownsize(t *testing.T) {
	_ = setRestrictedToFalse()
	// Insert Cluster
	clusterId, err := insertNewClusterIntoDb(t)
	if err != nil {
		fmt.Println(err)
	}
	assert.Nil(t, err)
	// Set Actionable
	assert.Nil(t, setActionableState(clusterId))
	// Insert Nodegroup
	nodeGroupId, err := insertNewNodeGroupIntoDB(t, clusterId)
	assert.Nil(t, err)
	// Set Actionable after nodegroup insert
	assert.Nil(t, setActionableState(clusterId))
	downsize := true
	// Initialize compute client mocking
	req := &pb.DeleteNodeGroupInstanceRequest{
		Clusteruuid:    clusterId,
		Nodegroupuuid:  nodeGroupId,
		Downsize:       &downsize,
		CloudAccountId: "iks_user",
		InstanceName:   "test-name",
	}
	ctx := context.Background()
	iksSrv, err := InitInstanceServiceMockClient(t, true, codes.OK, "all is good!")
	assert.Nil(t, err)
	nodeGroupStatus, err := iksSrv.DeleteNodeGroupInstance(ctx, req)
	assert.Nil(t, err)
	assert.Equal(t, nodeGroupStatus.Count, nodeGroupCreate.Count-1)
	assert.Nil(t, deleteClusterFromDb(t, clusterId))
}

func insertNewClusterIntoDb(t *testing.T) (string, error) {
	// Setup mock
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with ssh mocking failed")
	}
	clusterCreate = pb.ClusterRequest{
		CloudAccountId: "iks_user",
		Name:           uuid.New().String()[0:11],
		Description:    &desc,
		K8Sversionname: "1.28",
		Runtimename:    "Containerd",
	}
	// THIS FUNCTION ASSUMES THAT CREATE CLUSTER WORKS. REMOVE ONCE WE HAVE DB MIGRATE FOR TESTS SET UP
	res, err := iksSrv.CreateNewCluster(context.Background(), &clusterCreate)
	if err != nil {
		return "", err
	}
	return res.Uuid, nil
}

func insertNewNodeGroupIntoDB(t *testing.T, clusterUuid string) (string, error) {
	// THIS FUNCTION ASSUMES THAT CREATE CLUSTER WORKS. REMOVE ONCE WE HAVE DB MIGRATE FOR TESTS SET UP
	// Initialize compute client mocking
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNodegroup").WithValues(logkeys.NodeGroupName, nodeGroupCreate.Name, logkeys.CloudAccountId, nodeGroupCreate.CloudAccountId).Start()
	defer span.End()

	iksSrv, err := InitVnetMockClient(ctx, t)
	if err != nil {
		t.Fatalf("iks Server with compute mocking failed")
	}

	//Set up
	nodeGroupCreate = pb.CreateNodeGroupRequest{
		Clusteruuid:    clusterUuid,
		Name:           "nodegroup1",
		Description:    &desc,
		Instancetypeid: "vm-spr-tny",
		Count:          2,
		CloudAccountId: "iks_user",
		Vnets: []*pb.Vnet{ // Must matche return value in NewMockVnetServiceClient return function
			{
				Availabilityzonename:     "us-dev-1a",
				Networkinterfacevnetname: "us-dev-1a-default",
			},
		},
		Userdataurl: &uri,
	}

	res, err := iksSrv.CreateNodeGroup(context.Background(), &nodeGroupCreate)
	if err != nil {
		return "", err
	}
	return res.Nodegroupuuid, nil
}

func deleteClusterFromDb(t *testing.T, clusterUuid string) error {
	// Setup mock
	ctx, _, span := obs.LogAndSpanFromContextOrGlobal(context.Background()).WithName("Server.CreateNewCluster").WithValues(logkeys.ClusterName, clusterCreate.Name, logkeys.CloudAccountId, clusterCreate.CloudAccountId).Start()
	defer span.End()
	iksSrv, err := InitSshMockClient(ctx, t)
	if err != nil {
		return err
	}
	clusterDelete := pb.ClusterID{Clusteruuid: clusterUuid, CloudAccountId: "iks_user"}
	_, err = iksSrv.DeleteCluster(ctx, &clusterDelete)
	return err
}

// THIS FUNCTION ASSUMES THAT CREATE CLUSTER WORKS. REMOVE ONCE WE HAVE DB MIGRATE FOR TESTS SET UP
//res, err := iksSrv.DeleteCluster(context.Background(), &clusterCreate)
//if err != nil {
//return "", err
//}
//return res.Uuid, nil
//sqlDB, _ := managedDb.Open(context.Background())
//
//res, err := sqlDB.Exec(`
//	delete from clusterrev
//	       where cluster_id =
//	             (select cluster_id from cluster where unique_id = $1)`,
//	clusterUniqueId,
//)
//fmt.Println(res.RowsAffected())
//res, err = sqlDB.Exec(`
//	DELETE from cluster
//	WHERE unique_id = $1`,
//	clusterUniqueId,
//)
//fmt.Println(res.RowsAffected())
//return err
//}

/*

func TestDeleteSecurityRule(t *testing.T) {
	// RUN
	_, err = query.DeleteFirewallRule(context.Background(), db, &nodeGroupId)
	if err != nil {
		t.Fatalf("Should not get Error, recieved: %v", err.Error())
	}
	err := removeFWFromDb(clusterUuidDelete)
	if err != nil {
		t.Fatalf("Could not clean up entry %v", err.Error())
	}
}*/
