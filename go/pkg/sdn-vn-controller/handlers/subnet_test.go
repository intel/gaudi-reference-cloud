// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"context"
	"testing"

	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/tests/utils"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
	libovsdbtest "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/testing/libovsdb"
)

func TestCreateSubnetHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	subnetId1 := "a792b75b-afa5-41a4-97d8-92a5cb80cfdc"
	cidr1 := "10.0.0.0/24"
	az1 := "az1"

	// Set up SQL
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})
	_, _ = db.Exec(sqlQueryCreateVpc, vpcId1, "vpc1", "tenant1", "region1")

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Test 1: Send a normal request
	_, err = CreateSubnetHandler(db, nbClient,
		&v1.CreateSubnetRequest{
			SubnetId:         &v1.SubnetId{Uuid: subnetId1},
			Name:             "subnet1",
			Cidr:             cidr1,
			AvailabilityZone: az1,
			VpcId:            &v1.VPCId{Uuid: vpcId1},
		})
	if err != nil {
		t.Fatalf("Failed to create subnet: %v", err)
	}

	// Check OVSDB cache
	subnets := []nbdb.LogicalSwitch{}
	err = nbClient.Where(&nbdb.LogicalSwitch{UUID: subnetId1}).List(context.Background(), &subnets)
	if err != nil || len(subnets) < 1 {
		t.Fatalf("Subnet not created in OVSDB: %v", err)
	}
	if subnets[0].UUID != subnetId1 || subnets[0].Name != "subnet1" {
		t.Fatalf("Wrong subnet information in OVSDB: %v", err)
	}

	// Check SQL
	var resName, resCidr, resAz, resVpc string
	err = db.QueryRow(sqlQueryGetSubnet, subnetId1).Scan(&resName, &resCidr, &resAz, &resVpc)
	if err != nil {
		t.Fatalf("Subnet not created in SQL: %v", err)
	}
	if resName != "subnet1" || resVpc != vpcId1 {
		t.Fatalf("Wrong subnet information in SQL: %v", err)
	}

	// Test 2: Attempt to create subnet in a non-existing VPC. Should fail.
	_, err = CreateSubnetHandler(db, nbClient,
		&v1.CreateSubnetRequest{
			SubnetId:         &v1.SubnetId{Uuid: "bf6471b1-68a8-4b46-bdc4-669be6d6647f"},
			Name:             "subnet2",
			Cidr:             "172.17.0.0/28",
			AvailabilityZone: "az2",
			VpcId:            &v1.VPCId{Uuid: "ed81871d-3f2e-4a2f-91f8-c578ee7c7197"},
		})
	if err == nil {
		t.Fatalf("Failed to block invalid request: %v", err)
	}
}

func TestDeleteSubnetHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	subnetId1 := "a792b75b-afa5-41a4-97d8-92a5cb80cfdc"
	cidr1 := "10.0.0.0/24"
	az1 := "az1"

	// Set up SQL
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})
	_, _ = db.Exec(sqlQueryCreateVpc, vpcId1, "vpc1", "tenant1", "region1")
	_, _ = db.Exec(sqlQueryCreateSubnet, subnetId1, "subnet1", cidr1, az1, vpcId1)

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalSwitch{
				UUID: subnetId1,
				Name: "subnet1",
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Test 1: A normal subnet deletion
	_, err = DeleteSubnetHandler(db, nbClient, &v1.DeleteSubnetRequest{SubnetId: &v1.SubnetId{Uuid: subnetId1}})
	if err != nil {
		t.Fatalf("Failed to delete subnet: %v", err)
	}

	// Check OVSDB cache
	subnets := []nbdb.LogicalSwitch{}
	err = nbClient.Where(&nbdb.LogicalSwitch{UUID: subnetId1}).List(context.Background(), &subnets)
	if err != nil || len(subnets) > 0 {
		t.Fatalf("Subnet not deleted in OVSDB: %v", err)
	}
	// Check SQL
	rows, err := db.Query(sqlQueryListSubnet)
	if rows.Next() {
		t.Fatalf("Subnet not deleted in SQL: %v", err)
	}

	// Test 2: Attempt to delete a non-existing subnet. Should fail.
	_, err = DeleteSubnetHandler(db, nbClient, &v1.DeleteSubnetRequest{SubnetId: &v1.SubnetId{Uuid: "bb55d86b-a4c3-4843-939b-e7976f1da81d"}})
	if err == nil {
		t.Fatalf("Failed to block invalid request: %v", err)
	}
}

func TestListSubnetsHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	subnetId1 := "a792b75b-afa5-41a4-97d8-92a5cb80cfdc"
	cidr1 := "10.0.0.0/24"
	az1 := "az1"

	// Set up SQL
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})
	_, _ = db.Exec(sqlQueryCreateVpc, vpcId1, "vpc1", "tenant1", "region1")
	_, _ = db.Exec(sqlQueryCreateSubnet, subnetId1, "subnet1", cidr1, az1, vpcId1)

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalSwitch{
				UUID: subnetId1,
				Name: "subnet1",
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Start testing the handler
	res, err := ListSubnetsHandler(db, nbClient, &v1.ListSubnetsRequest{})
	if err != nil {
		t.Fatalf("failed to list subnet: %v", err)
	}

	// Check gRPC response
	if len(res.SubnetIds) < 1 || res.SubnetIds[0].Uuid != subnetId1 {
		t.Fatalf("incomplete results of listing subnets: %v", err)
	}
}

func TestGetSubnetHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	subnetId1 := "a792b75b-afa5-41a4-97d8-92a5cb80cfdc"
	cidr1 := "10.0.0.0/24"
	az1 := "az1"

	// Set up SQL
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})
	_, _ = db.Exec(sqlQueryCreateVpc, vpcId1, "vpc1", "tenant1", "region1")
	_, _ = db.Exec(sqlQueryCreateSubnet, subnetId1, "subnet1", cidr1, az1, vpcId1)

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalSwitch{
				UUID: subnetId1,
				Name: "subnet1",
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Start testing the handler
	res, err := GetSubnetHandler(db, nbClient, &v1.GetSubnetRequest{SubnetId: &v1.SubnetId{Uuid: subnetId1}})
	if err != nil {
		t.Fatalf("failed to get subnet: %v", err)
	}

	// Check gRPC response
	if res.Subnet.Name != "subnet1" || res.Subnet.VpcId.Uuid != vpcId1 {
		t.Fatalf("wrong subnet information: %v", err)
	}
}
