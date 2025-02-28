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

func TestCreateRouterHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	routerId1 := "c1e6fdc0-9c7d-4920-a112-1392c9b8c448"

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
	_, err = CreateRouterHandler(db, nbClient,
		&v1.CreateRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId1},
			Name:     "lr1",
			VpcId:    &v1.VPCId{Uuid: vpcId1},
		})
	if err != nil {
		t.Fatalf("Failed to create router: %v", err)
	}

	// Check OVSDB cache
	routers := []nbdb.LogicalRouter{}
	err = nbClient.Where(&nbdb.LogicalRouter{UUID: routerId1}).List(context.Background(), &routers)
	if err != nil || len(routers) < 1 {
		t.Fatalf("Router not created in OVSDB: %v", err)
	}
	if routers[0].UUID != routerId1 || routers[0].Name != "lr1" {
		t.Fatalf("Wrong router information in OVSDB: %v", err)
	}

	// Check SQL
	var resName, resVpc string
	err = db.QueryRow(sqlQueryGetRouter, routerId1).Scan(&resName, &resVpc)
	if err != nil {
		t.Fatalf("Router not created in SQL: %v", err)
	}
	if resName != "lr1" || resVpc != vpcId1 {
		t.Fatalf("Wrong router information in SQL: %v", err)
	}

	// Test 2: Attempt to create router in a non-existing VPC. Should fail.
	_, err = CreateRouterHandler(db, nbClient,
		&v1.CreateRouterRequest{
			RouterId: &v1.RouterId{Uuid: "6fa4aede-6a82-4700-9191-820255991436"},
			Name:     "lr2",
			VpcId:    &v1.VPCId{Uuid: "bd8e18ab-1b69-4628-80f6-acfb379739bf"},
		})
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}

func TestDeleteRouterHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	routerId1 := "c1e6fdc0-9c7d-4920-a112-1392c9b8c448"

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

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalRouter{
				UUID: routerId1,
				Name: "lr1",
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Test 1: A normal request
	_, err = DeleteRouterHandler(db, nbClient, &v1.DeleteRouterRequest{RouterId: &v1.RouterId{Uuid: routerId1}})
	if err != nil {
		t.Fatalf("Failed to delete router: %v", err)
	}

	// Check OVSDB cache
	routers := []nbdb.LogicalRouter{}
	err = nbClient.Where(&nbdb.LogicalRouter{UUID: routerId1}).List(context.Background(), &routers)
	if err != nil || len(routers) > 0 {
		t.Fatalf("Router not deleted in OVSDB: %v", err)
	}
	// Check SQL
	rows, err := db.Query(sqlQueryListRouter)
	if rows.Next() {
		t.Fatalf("Router not deleted in SQL: %v", err)
	}

	// Test 2: Attempt to delete a non-existing router. Should fail.
	_, err = DeleteRouterHandler(db, nbClient, &v1.DeleteRouterRequest{RouterId: &v1.RouterId{Uuid: "3245f1c7-6db9-4bc5-ba8c-996cda0f2872"}})
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}

func TestListRoutersHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	routerId1 := "c1e6fdc0-9c7d-4920-a112-1392c9b8c448"

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

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalRouter{
				UUID: routerId1,
				Name: "lr1",
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Start testing the handler
	res, err := ListRoutersHandler(db, nbClient, &v1.ListRoutersRequest{})
	if err != nil {
		t.Fatalf("failed to list routers: %v", err)
	}

	// Check gRPC response
	if len(res.RouterIds) < 1 || res.RouterIds[0].Uuid != routerId1 {
		t.Fatalf("incomplete results of listing routers: %v", err)
	}
}

func TestGetRouterHandler(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"
	routerId1 := "c1e6fdc0-9c7d-4920-a112-1392c9b8c448"

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

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			&nbdb.LogicalRouter{
				UUID: routerId1,
				Name: "lr1",
			},
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Start testing the handler
	res, err := GetRouterHandler(db, nbClient, &v1.GetRouterRequest{RouterId: &v1.RouterId{Uuid: routerId1}})
	if err != nil {
		t.Fatalf("failed to get router: %v", err)
	}

	// Check gRPC response
	if res.Router.Name != "lr1" || res.Router.VpcId.Uuid != vpcId1 {
		t.Fatalf("wrong router information: %v", err)
	}
}
