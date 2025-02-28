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

func TestCreateRouterInterfaceHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	routerId1 := "c1e6fdc0-9c7d-4920-a112-1392c9b8c448"
	subnetId1 := "751eb7e1-ab22-4baf-9773-321f95289270"

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
	_, _ = db.Exec(sqlQueryCreateRouter, routerId1, "lr1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateSubnet, subnetId1, "ls1", "10.1.2.0/24", "az1", vpcId1)

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalRouter{
				UUID: routerId1,
				Name: "lr1",
			},
			&nbdb.LogicalSwitch{
				UUID: subnetId1,
				Name: "ls1",
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Start testing the handler
	uuid := "79487137-324f-46fc-9b24-70c96e88db2e"
	request := &v1.CreateRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{Uuid: uuid},
		RouterId:          &v1.RouterId{Uuid: routerId1},
		SubnetId:          &v1.SubnetId{Uuid: subnetId1},
		Interface_IP:      "10.1.2.254/24",
		Interface_MAC:     "11:22:33:44:55:66",
	}
	_, err = CreateRouterInterfaceHandler(db, nbClient, request)
	if err != nil {
		t.Fatalf("Failed to create router interface: %v", err)
	}

	// Check port UUIDs are returned to SQL
	var lrpUuid, lspUuid string
	err = db.QueryRow("SELECT router_port_id, switch_port_id FROM router_interface WHERE router_interface_id=$1", uuid).Scan(&lrpUuid, &lspUuid)
	if err != nil {
		t.Fatalf("Failed to setup router interface SQL row: %v", err)
	}

	// Check ports are created in OVSDB
	switchPorts := []nbdb.LogicalSwitchPort{}
	err = nbClient.Where(&nbdb.LogicalSwitchPort{UUID: lspUuid}).List(context.Background(), &switchPorts)
	if err != nil || len(switchPorts) < 1 {
		t.Fatalf("Switch-router interface not created in OVSDB: %v", err)
	}
	routerPorts := []nbdb.LogicalRouterPort{}
	err = nbClient.Where(&nbdb.LogicalRouterPort{UUID: lrpUuid}).List(context.Background(), &routerPorts)
	if err != nil || len(routerPorts) < 1 {
		t.Fatalf("Router-switch interface not created in OVSDB: %v", err)
	}
	if routerPorts[0].MAC != "11:22:33:44:55:66" ||
		len(routerPorts[0].Networks) < 1 ||
		routerPorts[0].Networks[0] != "10.1.2.254/24" {
		t.Fatalf("Wrong interface information in OVSDB")
	}

	// Reject invalid requests: router must exist
	bad_request := &v1.CreateRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{Uuid: "52b34f67-ad5a-4a92-add8-32243a45815d"},
		RouterId:          &v1.RouterId{Uuid: "cac94d75-b340-4c7a-99e6-9c98c49207bb"},
		SubnetId:          &v1.SubnetId{Uuid: subnetId1},
		Interface_IP:      "10.2.1.254/16",
		Interface_MAC:     "aa:bb:cc:dd:ee:ff",
	}
	_, err = CreateRouterInterfaceHandler(db, nbClient, bad_request)
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
	// One subnet can only have one router interface
	bad_request = &v1.CreateRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{Uuid: "4a387aa1-035d-4cbb-b20c-eaeed6be74e2"},
		RouterId:          &v1.RouterId{Uuid: routerId1},
		SubnetId:          &v1.SubnetId{Uuid: subnetId1},
		Interface_IP:      "10.2.3.254/16",
		Interface_MAC:     "aa:bb:cc:dd:00:11",
	}
	_, err = CreateRouterInterfaceHandler(db, nbClient, bad_request)
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}

func TestDeleteRouterInterfaceHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	routerId1 := "c1e6fdc0-9c7d-4920-a112-1392c9b8c448"
	subnetId1 := "751eb7e1-ab22-4baf-9773-321f95289270"
	lrpId := "9bff618d-f4d5-441c-ac26-4d8896cb10c5"
	lspId := "566c5c0d-22f4-4d4c-bbc5-bdf59414f5f0"
	uuid := "f7fd0fda-df6d-48fd-af10-554fbd8fb21b"

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
	_, _ = db.Exec(sqlQueryCreateRouter, routerId1, "lr1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateSubnet, subnetId1, "ls1", "10.1.2.0/24", "az1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateRouterInterface, uuid, subnetId1, routerId1, lrpId, lspId, "10.1.2.254/24", "11:22:33:aa:bb:cc")

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalRouter{UUID: routerId1, Ports: []string{lrpId}},
			&nbdb.LogicalSwitch{UUID: subnetId1, Ports: []string{lspId}},
			&nbdb.LogicalSwitchPort{UUID: lspId},
			&nbdb.LogicalRouterPort{UUID: lrpId},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Start testing the handler
	request := &v1.DeleteRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{Uuid: uuid},
	}
	_, err = DeleteRouterInterfaceHandler(db, nbClient, request)
	if err != nil {
		t.Fatalf("failed to delete router interface: %v", err)
	}

	// Check row removed from SQL
	row := db.QueryRow("SELECT * FROM router_interface WHERE router_interface_id=$1", uuid)
	if row.Err() != nil {
		t.Fatalf("failed to delete router interface SQL row: %v", err)
	}
	// Check ports are no longer in OVSDB
	switchPorts := []nbdb.LogicalSwitchPort{}
	err = nbClient.Where(&nbdb.LogicalSwitchPort{UUID: lspId}).List(context.Background(), &switchPorts)
	if err != nil || len(switchPorts) > 0 {
		t.Fatalf("Switch-router interface not removed in OVSDB: %v", err)
	}
	routerPorts := []nbdb.LogicalRouterPort{}
	err = nbClient.Where(&nbdb.LogicalRouterPort{UUID: lrpId}).List(context.Background(), &routerPorts)
	if err != nil || len(routerPorts) > 0 {
		t.Fatalf("Router-switch interface not removed in OVSDB: %v", err)
	}

	// Reject invalid requests
	bad_request := &v1.DeleteRouterInterfaceRequest{
		RouterInterfaceId: &v1.RouterInterfaceId{Uuid: "911f2c14-bf63-44ac-bd59-2aa93397b40d"},
	}
	_, err = DeleteRouterInterfaceHandler(db, nbClient, bad_request)
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}

func TestListRouterInterfaceHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	routerId1 := "c1e6fdc0-9c7d-4920-a112-1392c9b8c448"
	subnetId1 := "751eb7e1-ab22-4baf-9773-321f95289270"
	lrpId := "9bff618d-f4d5-441c-ac26-4d8896cb10c5"
	lspId := "566c5c0d-22f4-4d4c-bbc5-bdf59414f5f0"
	uuid := "f7fd0fda-df6d-48fd-af10-554fbd8fb21b"

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
	_, _ = db.Exec(sqlQueryCreateRouter, routerId1, "lr1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateSubnet, subnetId1, "ls1", "10.1.2.0/24", "az1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateRouterInterface, uuid, subnetId1, routerId1, lrpId, lspId, "10.1.2.254/24", "11:22:33:aa:bb:cc")

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalRouter{UUID: routerId1, Ports: []string{lrpId}},
			&nbdb.LogicalSwitch{UUID: subnetId1, Ports: []string{lspId}},
			&nbdb.LogicalSwitchPort{UUID: lspId},
			&nbdb.LogicalRouterPort{UUID: lrpId},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Start testing the handler
	res, err := ListRouterInterfacesHandler(db, nbClient,
		&v1.ListRouterInterfacesRequest{})
	if err != nil {
		t.Fatalf("failed to list router interface: %v", err)
	}
	// Check gRPC response
	if len(res.RouterInterfaceIds) < 1 || res.RouterInterfaceIds[0].Uuid != uuid {
		t.Fatalf("incomplete results of listing router interfaces")
	}
}

func TestGetRouterInterfaceHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	routerId1 := "c1e6fdc0-9c7d-4920-a112-1392c9b8c448"
	subnetId1 := "751eb7e1-ab22-4baf-9773-321f95289270"
	lrpId := "9bff618d-f4d5-441c-ac26-4d8896cb10c5"
	lspId := "566c5c0d-22f4-4d4c-bbc5-bdf59414f5f0"
	uuid := "f7fd0fda-df6d-48fd-af10-554fbd8fb21b"

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
	_, _ = db.Exec(sqlQueryCreateRouter, routerId1, "lr1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateSubnet, subnetId1, "ls1", "10.1.2.0/24", "az1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateRouterInterface, uuid, subnetId1, routerId1, lrpId, lspId, "10.1.2.254/24", "11:22:33:aa:bb:cc")

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalRouter{UUID: routerId1, Ports: []string{lrpId}},
			&nbdb.LogicalSwitch{UUID: subnetId1, Ports: []string{lspId}},
			&nbdb.LogicalSwitchPort{UUID: lspId},
			&nbdb.LogicalRouterPort{UUID: lrpId},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Start testing the handler
	res, err := GetRouterInterfaceHandler(db, nbClient,
		&v1.GetRouterInterfaceRequest{RouterInterfaceId: &v1.RouterInterfaceId{Uuid: uuid}})
	if err != nil {
		t.Fatalf("failed to get router interface: %v", err)
	}
	// Check gRPC response
	if res.RouterInterface.RouterId.Uuid != routerId1 || res.RouterInterface.SubnetId.Uuid != subnetId1 {
		t.Fatalf("wrong router interface information")
	}

	// Reject invalid requests
	bad_request := &v1.GetRouterInterfaceRequest{RouterInterfaceId: &v1.RouterInterfaceId{Uuid: "d5235933-0a7e-41b1-bab0-36c0d4e3f075"}}
	_, err = GetRouterInterfaceHandler(db, nbClient, bad_request)
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}
