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

func TestCreateStaticRoutesHandler(t *testing.T) {
	vpcId1 := "a7a82472-459c-4c14-a8c0-e02e37ae4081"
	routeUuid1 := "0f438af4-1b4c-4b9d-8a6f-44b9ba1ea195"
	routerUuid1 := "c0b619bb-812b-4b0b-b139-d4955426b4d7"
	fakeIpPrefix := "10.0.0.0/8"
	fakeNextHop := "10.0.0.254"

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
	_, _ = db.Exec(sqlQueryCreateRouter, routerUuid1, "FakeLogicalRouter1", vpcId1)

	// Set up in-memory OVSDB with 1 router and no static routes
	fakeRouter1 := nbdb.LogicalRouter{
		UUID: routerUuid1,
		Name: "FakeLogicalRouter1",
	}
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			fakeRouter1.DeepCopy(),
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Use the handler to add 1 static route to the fake router
	_, err = CreateStaticRouteHandler(db, nbClient,
		&v1.CreateStaticRouteRequest{
			StaticRouteId: &v1.StaticRouteId{Uuid: routeUuid1},
			RouterId:      &v1.RouterId{Uuid: routerUuid1},
			Prefix:        fakeIpPrefix,
			Nexthop:       fakeNextHop,
		})
	if err != nil {
		t.Fatalf("Failed to create static routes: %v", err)
	}

	// Check OVSDB cache if the route is correctly created
	staticRoutes := []nbdb.LogicalRouterStaticRoute{}
	err = nbClient.Where(&nbdb.LogicalRouterStaticRoute{
		UUID:     routeUuid1,
		IPPrefix: fakeIpPrefix,
		Nexthop:  fakeNextHop,
	}).List(context.Background(), &staticRoutes)

	if err != nil {
		t.Fatalf("Failed to list static routes: %v", err)
	}
	if len(staticRoutes) < 1 {
		t.Fatalf("Target static route not created")
	}

	// Check OVSDB cache if the route is associated to the logical router
	routers := []nbdb.LogicalRouter{}
	err = nbClient.List(context.Background(), &routers)
	if err != nil {
		t.Fatalf("Failed to get logical router: %v", err)
	}
	if len(routers[0].StaticRoutes) < 1 || routers[0].StaticRoutes[0] != routeUuid1 {
		t.Fatalf("Target static route not added to logical router")
	}
	// Check SQL
	var resPrefix, resNexthop, resRouter string
	err = db.QueryRow(sqlQueryGetStaticRoute, routeUuid1).Scan(&resPrefix, &resNexthop, &resRouter)
	if err != nil {
		t.Fatalf("Static route not created in SQL: %v", err)
	}
	if resPrefix != fakeIpPrefix || resNexthop != fakeNextHop || resRouter != routerUuid1 {
		t.Fatalf("Wrong static route information in SQL: %v", err)
	}

	// Negative case: invalid router_id
	_, err = CreateStaticRouteHandler(db, nbClient,
		&v1.CreateStaticRouteRequest{
			StaticRouteId: &v1.StaticRouteId{Uuid: "6123d0eb-8b2e-4866-b78a-c91edd218689"},
			RouterId:      &v1.RouterId{Uuid: "a8f673cf-5772-4f6b-a785-f33b3ca8271e"},
			Prefix:        "20.1.2.0/24",
			Nexthop:       "20.1.2.3",
		})
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}

func TestDeleteStaticRoutesHandler(t *testing.T) {
	vpcId1 := "a7a82472-459c-4c14-a8c0-e02e37ae4081"
	routeUuid1 := "0f438af4-1b4c-4b9d-8a6f-44b9ba1ea195"
	routerUuid1 := "c0b619bb-812b-4b0b-b139-d4955426b4d7"
	fakeIpPrefix := "10.0.0.0/8"
	fakeNextHop := "10.0.0.254"

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
	_, _ = db.Exec(sqlQueryCreateRouter, routerUuid1, "FakeLogicalRouter1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateStaticRoute, routeUuid1, fakeIpPrefix, fakeNextHop, routerUuid1)

	// Set up in-memory OVSDB
	fakeRouter1 := nbdb.LogicalRouter{
		UUID:         routerUuid1,
		Name:         "FakeLogicalRouter1",
		StaticRoutes: []string{routeUuid1},
	}
	fakeStaticRoute1 := nbdb.LogicalRouterStaticRoute{
		UUID:     routeUuid1,
		IPPrefix: fakeIpPrefix,
		Nexthop:  fakeNextHop,
	}

	// Set up an in-memory database with 1 router and no static routes
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			fakeRouter1.DeepCopy(),
			fakeStaticRoute1.DeepCopy(),
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Use the handler to delete the added route
	_, err = DeleteStaticRouteHandler(db, nbClient, &v1.DeleteStaticRouteRequest{
		StaticRouteId: &v1.StaticRouteId{Uuid: routeUuid1},
	})
	if err != nil {
		t.Fatalf("failed to delete static routes: %v", err)
	}

	// Check OVSDB cache if the route no longer exists
	staticRoutes := []nbdb.LogicalRouterStaticRoute{}
	err = nbClient.Where(&nbdb.LogicalRouterStaticRoute{
		UUID: routeUuid1,
	}).List(context.Background(), &staticRoutes)

	if err != nil {
		t.Fatalf("Failed to list static routes: %v", err)
	}
	if len(staticRoutes) > 0 {
		t.Fatalf("Target static route not deleted")
	}

	// Check OVSDB cache if the route is no longer associated to the logical router
	routers := []nbdb.LogicalRouter{}
	err = nbClient.List(context.Background(), &routers)
	if err != nil {
		t.Fatalf("Failed to get logical router: %v", err)
	}
	if len(routers[0].StaticRoutes) > 0 {
		t.Fatalf("Target static route not purged from logical router")
	}

	// Check SQL
	rows, err := db.Query(sqlQueryListStaticRoute)
	if rows.Next() {
		t.Fatalf("Static route not deleted in SQL: %v", err)
	}

	// Negative case: non-existing static route
	_, err = DeleteStaticRouteHandler(db, nbClient, &v1.DeleteStaticRouteRequest{
		StaticRouteId: &v1.StaticRouteId{Uuid: "8ca694e0-20be-4468-89ba-30e8d77e2d21"},
	})
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}

func TestListStaticRoutesHandler(t *testing.T) {
	vpcId1 := "a7a82472-459c-4c14-a8c0-e02e37ae4081"
	routeUuid1 := "0f438af4-1b4c-4b9d-8a6f-44b9ba1ea195"
	routerUuid1 := "c0b619bb-812b-4b0b-b139-d4955426b4d7"
	fakeIpPrefix := "10.0.0.0/8"
	fakeNextHop := "10.0.0.254"

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
	_, _ = db.Exec(sqlQueryCreateRouter, routerUuid1, "FakeLogicalRouter1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateStaticRoute, routeUuid1, fakeIpPrefix, fakeNextHop, routerUuid1)

	// Set up in-memory OVSDB
	fakeRouter1 := nbdb.LogicalRouter{
		UUID:         routerUuid1,
		Name:         "FakeLogicalRouter1",
		StaticRoutes: []string{routeUuid1},
	}
	fakeStaticRoute1 := nbdb.LogicalRouterStaticRoute{
		UUID:     routeUuid1,
		IPPrefix: fakeIpPrefix,
		Nexthop:  fakeNextHop,
	}
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			fakeRouter1.DeepCopy(),
			fakeStaticRoute1.DeepCopy(),
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Try to list the added fake route
	res, err := ListStaticRoutesHandler(db, nbClient, &v1.ListStaticRoutesRequest{})
	if err != nil {
		t.Fatalf("failed to list static routes: %v", err)
	}

	// Check the results of the handler
	if len(res.StaticRouteIds) != 1 {
		t.Fatalf("Returned wrong number of static routes")
	}
	if res.StaticRouteIds[0].Uuid != routeUuid1 {
		t.Fatalf("Returned wrong UUID of static routes")
	}
}

func TestGetStaticRoutesHandler(t *testing.T) {
	vpcId1 := "a7a82472-459c-4c14-a8c0-e02e37ae4081"
	routeUuid1 := "0f438af4-1b4c-4b9d-8a6f-44b9ba1ea195"
	routerUuid1 := "c0b619bb-812b-4b0b-b139-d4955426b4d7"
	fakeIpPrefix := "10.0.0.0/8"
	fakeNextHop := "10.0.0.254"

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
	_, _ = db.Exec(sqlQueryCreateRouter, routerUuid1, "FakeLogicalRouter1", vpcId1)
	_, _ = db.Exec(sqlQueryCreateStaticRoute, routeUuid1, fakeIpPrefix, fakeNextHop, routerUuid1)

	// Set up in-memory OVSDB
	fakeRouter1 := nbdb.LogicalRouter{
		UUID:         routerUuid1,
		Name:         "FakeLogicalRouter1",
		StaticRoutes: []string{routeUuid1},
	}
	fakeStaticRoute1 := nbdb.LogicalRouterStaticRoute{
		UUID:     routeUuid1,
		IPPrefix: fakeIpPrefix,
		Nexthop:  fakeNextHop,
	}
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{
			fakeRouter1.DeepCopy(),
			fakeStaticRoute1.DeepCopy(),
		},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Call the handler to test
	res, err := GetStaticRouteHandler(db, nbClient, &v1.GetStaticRouteRequest{
		StaticRouteId: &v1.StaticRouteId{Uuid: routeUuid1},
	})
	if err != nil {
		t.Fatalf("failed to get static routes: %v", err)
	}

	// Check the results of the handler
	if res.StaticRoute.RouterId.Uuid != routerUuid1 {
		t.Fatalf("Returned wrong router information")
	}
	if res.StaticRoute.Prefix != fakeIpPrefix || res.StaticRoute.Nexthop != fakeNextHop {
		t.Fatalf("Returned wrong IP information")
	}

	// Negative case: non-existing static route
	_, err = GetStaticRouteHandler(db, nbClient, &v1.GetStaticRouteRequest{
		StaticRouteId: &v1.StaticRouteId{Uuid: "8ca694e0-20be-4468-89ba-30e8d77e2d21"},
	})
	if err == nil {
		t.Fatalf("Failed to block invalid request")
	}
}
