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

func TestCreatePortHandler(t *testing.T) {
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

	// Test 1: Send a normal request
	portId1 := "4fbff1a4-b581-4842-8dae-acbf80ad63c5"
	_, err = CreatePortHandler(db, nbClient,
		&v1.CreatePortRequest{
			SubnetId:           &v1.SubnetId{Uuid: subnetId1},
			PortId:             &v1.PortId{Uuid: portId1},
			ChassisId:          "1",
			DeviceId:           2,
			Internal_IPAddress: "10.0.0.1",
			IsNAT:              false,
		})
	if err != nil {
		t.Fatalf("Failed to create port: %v", err)
	}
	// TODO: NAT-related functions are now covered in scenario tests. Consider whether to create unit test cases.

	// Check OVSDB cache
	subnets := []nbdb.LogicalSwitch{}
	err = nbClient.Where(&nbdb.LogicalSwitch{UUID: subnetId1}).List(context.Background(), &subnets)
	if err != nil || len(subnets) < 1 {
		t.Fatalf("Subnet not found in OVSDB: %v", err)
	}
	if len(subnets[0].Ports) < 1 {
		t.Fatalf("Port not attached to subnet in OVSDB")
	}
	ports := []nbdb.LogicalSwitchPort{}
	err = nbClient.Where(&nbdb.LogicalSwitchPort{UUID: portId1}).List(context.Background(), &ports)
	if err != nil || len(ports) < 1 {
		t.Fatalf("Port not created in OVSDB: %v", err)
	}

	// Check SQL
	var retSubnetId string
	var retDeviceId, retChassisId int
	err = db.QueryRow("SELECT subnet_id, device_id, chassis_id FROM port WHERE port_id = $1", portId1).Scan(&retSubnetId, &retDeviceId, &retChassisId)
	if err != nil {
		t.Fatalf("Port not created in SQL: %v", err)
	}
	if retSubnetId != subnetId1 || retDeviceId != 2 || retChassisId != 1 {
		t.Fatalf("Wrong port information in SQL")
	}

	// Test 2: Attempt to create port in a non-existing subnet. Should fail.
	_, err = CreatePortHandler(db, nbClient,
		&v1.CreatePortRequest{
			SubnetId:           &v1.SubnetId{Uuid: "d473e307-1945-4d61-95cd-5629eb115893"},
			PortId:             &v1.PortId{Uuid: "b3013417-4dac-47ab-b364-a792ff8cd225"},
			ChassisId:          "1",
			DeviceId:           2,
			Internal_IPAddress: "10.0.0.2",
			IsNAT:              false,
		})
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}

func TestDeletePortHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	subnetId1 := "a792b75b-afa5-41a4-97d8-92a5cb80cfdc"
	portId1 := "4fbff1a4-b581-4842-8dae-acbf80ad63c5"
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
	_, _ = db.Exec(sqlQueryCreatePort, portId1, "1@port2", subnetId1, "1", 2, true, nil, "10.0.0.1", false, nil, nil, nil)

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalSwitchPort{
				UUID: portId1,
				Name: "1@port2",
			},
			&nbdb.LogicalSwitch{
				UUID:  subnetId1,
				Name:  "subnet1",
				Ports: []string{portId1},
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Test 1: Normal delete request
	_, err = DeletePortHandler(db, nbClient, &v1.DeletePortRequest{PortId: &v1.PortId{Uuid: portId1}})
	if err != nil {
		t.Fatalf("Failed to delete port: %v", err)
	}
	// Check OVSDB
	subnets := []nbdb.LogicalSwitch{}
	err = nbClient.Where(&nbdb.LogicalSwitch{UUID: subnetId1}).List(context.Background(), &subnets)
	if err != nil || len(subnets) < 1 {
		t.Fatalf("Subnet not found in OVSDB: %v", err)
	}
	if len(subnets[0].Ports) > 0 {
		t.Fatalf("Port not detached from subnet in OVSDB")
	}
	ports := []nbdb.LogicalSwitchPort{}
	err = nbClient.Where(&nbdb.LogicalSwitchPort{UUID: portId1}).List(context.Background(), &ports)
	if err != nil || len(ports) > 0 {
		t.Fatalf("Port not deleted in OVSDB: %v", err)
	}
	// Check SQL
	rows, err := db.Query(sqlQueryListPort)
	if err != nil || rows.Next() {
		t.Fatalf("Port not deleted in SQL: %v", err)
	}

	// Test 2: Negative case
	_, err = DeletePortHandler(db, nbClient, &v1.DeletePortRequest{PortId: &v1.PortId{Uuid: "b058961c-54a1-4682-95c7-5a92a0284913"}})
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}

func TestListPortHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	subnetId1 := "a792b75b-afa5-41a4-97d8-92a5cb80cfdc"
	portId1 := "4fbff1a4-b581-4842-8dae-acbf80ad63c5"
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
	_, _ = db.Exec(sqlQueryCreatePort, portId1, "1@port2", subnetId1, "1", 2, true, nil, "10.0.0.1", false, nil, nil, nil)

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalSwitchPort{
				UUID: portId1,
				Name: "1@port2",
			},
			&nbdb.LogicalSwitch{
				UUID:  subnetId1,
				Name:  "subnet1",
				Ports: []string{portId1},
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Send a request
	res, err := ListPortsHandler(db, nbClient, &v1.ListPortsRequest{})
	if err != nil {
		t.Fatalf("Failed to list ports: %v", err)
	}
	// Check gRPC response
	if len(res.PortIds) < 1 || res.PortIds[0].Uuid != portId1 {
		t.Fatalf("incomplete results of listing ports")
	}
}

func TestGetPortHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	subnetId1 := "a792b75b-afa5-41a4-97d8-92a5cb80cfdc"
	portId1 := "4fbff1a4-b581-4842-8dae-acbf80ad63c5"
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
	_, _ = db.Exec(sqlQueryCreatePort, portId1, "1@port2", subnetId1, "1", 2, true, nil, "10.0.0.1", false, nil, nil, nil)

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalSwitchPort{
				UUID: portId1,
				Name: "1@port2",
			},
			&nbdb.LogicalSwitch{
				UUID:  subnetId1,
				Name:  "subnet1",
				Ports: []string{portId1},
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Send a request
	res, err := GetPortHandler(db, nbClient, &v1.GetPortRequest{PortId: &v1.PortId{Uuid: portId1}})
	if err != nil {
		t.Fatalf("Failed to get port: %v", err)
	}
	// Check gRPC response
	if res.Port.Id.Uuid != portId1 || res.Port.ChassisId != "1" || res.Port.DeviceId != 2 {
		t.Fatalf("Get wrong port information")
	}

	// Negative case
	_, err = GetPortHandler(db, nbClient, &v1.GetPortRequest{PortId: &v1.PortId{Uuid: "56280447-14e8-4523-addd-dd87b6896b54"}})
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}
