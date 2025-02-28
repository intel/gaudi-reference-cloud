// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"context"
	"math/rand"

	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/tests/utils"

	"sync"
	"testing"
	"time"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"

	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
	libovsdbtest "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/testing/libovsdb"
)

// Scenario: Send each request to a different VPC. Ack the creation before request to delete
// Expect: All requests succeed; no racing; OVSDB and SQL return empty after the test
func TestMultiVpcSync(t *testing.T) {

	// Set up SQL
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})

	count := 50
	// Create a new VPC for each group of requests
	vpcIds := make([]string, 0)
	for i := 0; i < count; i++ {
		vpcId := uuid.NewString()
		vpcIds = append(vpcIds, vpcId)
		_, _ = db.Exec(sqlQueryCreateVpc, vpcId, vpcId, vpcId, vpcId)
	}

	// Set up an in-memory OVSDB
	setup := libovsdbtest.TestSetup{
		NBData: []libovsdbtest.TestData{},
	}
	nbClient, cleanup, err := libovsdbtest.NewNBTestHarness(setup, nil)
	if err != nil {
		t.Fatalf("Failed to set up test harness: %v", err)
	}
	t.Cleanup(cleanup.Cleanup)

	// Generate a pattern of concurrent requests
	chErr := make(chan bool, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		routerId := uuid.NewString()

		createRouterReq := &v1.CreateRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
			Name:     routerId,
			VpcId:    &v1.VPCId{Uuid: vpcIds[i]},
		}
		deleteRouterReq := &v1.DeleteRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
		}

		// Send requests and count failure rate
		go func(req1 *v1.CreateRouterRequest, req2 *v1.DeleteRouterRequest) {
			defer wg.Done()
			_, err1 := CreateRouterHandler(db, nbClient, req1)
			_, err2 := DeleteRouterHandler(db, nbClient, req2)
			chErr <- (err1 != nil && err2 != nil)
		}(createRouterReq, deleteRouterReq)
	}

	// Count failed requests: expect no failures
	wg.Wait()
	close(chErr)
	failures := 0
	for i := range chErr {
		if i {
			failures += 1
		}
	}
	if failures > 0 {
		t.Fatalf("Tests failed: %v/%v", failures, count)
	}

	// Check OVSDB and SQL: both should not have routers
	routers := []nbdb.LogicalRouter{}
	err = nbClient.List(context.Background(), &routers)
	if err != nil {
		t.Fatal("Fail to fetch routers in OVSDB")
	}
	if len(routers) > 0 {
		t.Fatal("OVSDB not consistent")
	}

	rows, err := db.Query(sqlQueryListRouter)
	if err != nil {
		t.Fatal("Fail to fetch routers in SQL")
	}
	if rows.Next() {
		t.Fatal("SQL not consistent")
	}
}

// Scenario: Send all requests to the same VPC. Ack the creation before request to delete
// Expect: All requests succeed; no racing; OVSDB and SQL return empty after the test
func TestSingleVpcSync(t *testing.T) {

	// Set up SQL
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})

	vpcId1 := "fc0789dc-4513-48f3-b763-6e4285d68ff3"
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

	// Generate a pattern of concurrent requests
	count := 50
	chErr := make(chan bool, count)
	var wg sync.WaitGroup
	wg.Add(count)
	for i := 0; i < count; i++ {
		routerId := uuid.NewString()

		createRouterReq := &v1.CreateRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
			Name:     routerId,
			VpcId:    &v1.VPCId{Uuid: vpcId1},
		}
		deleteRouterReq := &v1.DeleteRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
		}

		// Send requests and count failure rate
		go func(req1 *v1.CreateRouterRequest, req2 *v1.DeleteRouterRequest) {
			defer wg.Done()
			_, err1 := CreateRouterHandler(db, nbClient, req1)
			_, err2 := DeleteRouterHandler(db, nbClient, req2)
			chErr <- (err1 != nil && err2 != nil)
		}(createRouterReq, deleteRouterReq)
	}

	// Count failed requests: expect no failures
	wg.Wait()
	close(chErr)
	failures := 0
	for i := range chErr {
		if i {
			failures += 1
		}
	}
	if failures > 0 {
		t.Fatalf("Tests failed: %v/%v", failures, count)
	}

	// Check OVSDB and SQL: both should not have routers
	routers := []nbdb.LogicalRouter{}
	err = nbClient.List(context.Background(), &routers)
	if err != nil {
		t.Fatal("Fail to fetch routers in OVSDB")
	}
	if len(routers) > 0 {
		t.Fatal("OVSDB not consistent")
	}

	rows, err := db.Query(sqlQueryListRouter)
	if err != nil {
		t.Fatal("Fail to fetch routers in SQL")
	}
	if rows.Next() {
		t.Fatal("SQL not consistent")
	}
}

// Scenario: Send all requests to the same VPC. DO NOT ack the creation before request to delete
// Expect: No racing; OVSDB and SQL records are consistent
// Cannot guarantee: All requests succeed
func TestSingleVpcAsync(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"

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

	// Generate a pattern of concurrent requests
	count := 50
	chErr := make(chan bool, count*2)
	var wg sync.WaitGroup
	wg.Add(count * 2)
	for i := 0; i < count; i++ {
		routerId := uuid.NewString()

		createRouterReq := &v1.CreateRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
			Name:     routerId,
			VpcId:    &v1.VPCId{Uuid: vpcId1},
		}
		deleteRouterReq := &v1.DeleteRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
		}

		// Send async requests and count failure rate
		go func(req *v1.CreateRouterRequest) {
			defer wg.Done()
			_, err := CreateRouterHandler(db, nbClient, req)
			chErr <- (err != nil)
		}(createRouterReq)

		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

		go func(req *v1.DeleteRouterRequest) {
			defer wg.Done()
			_, err := DeleteRouterHandler(db, nbClient, req)
			chErr <- (err != nil)
		}(deleteRouterReq)
	}

	// Check how many requests failed
	// The current implementation does not preserve the serving order, therefore failure does not count as an error
	wg.Wait()
	close(chErr)
	failures := 0
	for i := range chErr {
		if i {
			failures += 1
		}
	}
	t.Logf("Request Error Rate: %v / %v", failures, count*2)

	// Check consistency between OVSDB and SQL
	routers := []nbdb.LogicalRouter{}
	routerIds := make(map[string]int)
	err = nbClient.List(context.Background(), &routers)
	if err != nil {
		t.Fatal("Fail to fetch routers in OVSDB")
	}
	t.Logf("Routers remained in OVSDB: %v", len(routers))
	for _, router := range routers {
		routerIds[router.UUID] = 1
	}

	rows, err := db.Query(sqlQueryListRouter)
	if err != nil {
		t.Fatal("Fail to fetch routers in SQL")
	}
	matchedRows := 0
	var pendingId string
	for rows.Next() {
		rows.Scan(&pendingId)
		if _, ok := routerIds[pendingId]; !ok {
			t.Fatal("Inconsistency router ID between SQL and OVSDB")
		}
		matchedRows += 1
	}
	if matchedRows != len(routers) {
		t.Fatal("Inconsistency router amount between SQL and OVSDB")
	}
}

// Scenario: Send async requests involving multiple interacting network objects
// Expect: No racing; OVSDB and SQL records are consistent; same number of switch port and router port
// Cannot guarantee: All requests succeed
func TestSingleVpcAsync2(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"

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

	// Generate a pattern of concurrent requests
	count := 50
	chErr := make(chan bool, count*6)
	var wg sync.WaitGroup
	wg.Add(count * 6)
	for i := 0; i < count; i++ {
		routerId := uuid.NewString()
		subnetId := uuid.NewString()
		interfaceId := uuid.NewString()

		createRouterReq := &v1.CreateRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
			Name:     routerId,
			VpcId:    &v1.VPCId{Uuid: vpcId1},
		}
		deleteRouterReq := &v1.DeleteRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
		}
		createSubnetReq := &v1.CreateSubnetRequest{
			SubnetId:         &v1.SubnetId{Uuid: subnetId},
			Name:             subnetId,
			VpcId:            &v1.VPCId{Uuid: vpcId1},
			Cidr:             "10.0.0.0/24",
			AvailabilityZone: "az1",
		}
		deleteSubnetReq := &v1.DeleteSubnetRequest{
			SubnetId: &v1.SubnetId{Uuid: subnetId},
		}
		createInterfaceReq := &v1.CreateRouterInterfaceRequest{
			RouterInterfaceId: &v1.RouterInterfaceId{Uuid: interfaceId},
			RouterId:          &v1.RouterId{Uuid: routerId},
			SubnetId:          &v1.SubnetId{Uuid: subnetId},
			Interface_IP:      "10.0.0.254",
			Interface_MAC:     "aa:bb:cc:dd:ee:ff",
		}
		deleteInterfaceReq := &v1.DeleteRouterInterfaceRequest{
			RouterInterfaceId: &v1.RouterInterfaceId{Uuid: interfaceId},
		}

		// Send async requests and count failure rate
		go func(req *v1.CreateRouterRequest) {
			defer wg.Done()
			_, err := CreateRouterHandler(db, nbClient, req)
			chErr <- (err != nil)
		}(createRouterReq)

		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

		go func(req *v1.CreateSubnetRequest) {
			defer wg.Done()
			_, err := CreateSubnetHandler(db, nbClient, req)
			chErr <- (err != nil)
		}(createSubnetReq)

		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

		go func(req *v1.CreateRouterInterfaceRequest) {
			defer wg.Done()
			_, err := CreateRouterInterfaceHandler(db, nbClient, req)
			chErr <- (err != nil)
		}(createInterfaceReq)

		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

		go func(req *v1.DeleteRouterInterfaceRequest) {
			defer wg.Done()
			_, err := DeleteRouterInterfaceHandler(db, nbClient, req)
			chErr <- (err != nil)
		}(deleteInterfaceReq)

		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

		go func(req *v1.DeleteSubnetRequest) {
			defer wg.Done()
			_, err := DeleteSubnetHandler(db, nbClient, req)
			chErr <- (err != nil)
		}(deleteSubnetReq)

		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

		go func(req *v1.DeleteRouterRequest) {
			defer wg.Done()
			_, err := DeleteRouterHandler(db, nbClient, req)
			chErr <- (err != nil)
		}(deleteRouterReq)

	}

	// Check how many requests failed
	// The current implementation does not preserve the serving order, therefore failure does not count as an error
	wg.Wait()
	close(chErr)
	failures := 0
	for i := range chErr {
		if i {
			failures += 1
		}
	}
	t.Logf("Request Error Rate: %v / %v", failures, count*6)

	// Check consistency between OVSDB and SQL: Router
	routers := []nbdb.LogicalRouter{}
	routerIds := make(map[string]int)
	err = nbClient.List(context.Background(), &routers)
	if err != nil {
		t.Fatal("Fail to fetch routers in OVSDB")
	}
	t.Logf("Routers remained in OVSDB: %v", len(routers))
	for _, router := range routers {
		routerIds[router.UUID] = 1
	}

	rows, err := db.Query(sqlQueryListRouter)
	if err != nil {
		t.Fatal("Fail to fetch routers in SQL")
	}
	matchedRows := 0
	var pendingId string
	for rows.Next() {
		rows.Scan(&pendingId)
		if _, ok := routerIds[pendingId]; !ok {
			t.Fatal("Inconsistency router ID between SQL and OVSDB")
		}
		matchedRows += 1
	}
	if matchedRows != len(routers) {
		t.Fatal("Inconsistency router amount between SQL and OVSDB")
	}

	// Check consistency between OVSDB and SQL: Subnet
	subnets := []nbdb.LogicalSwitch{}
	subnetIds := make(map[string]int)
	err = nbClient.List(context.Background(), &subnets)
	if err != nil {
		t.Fatal("Fail to fetch subnets in OVSDB")
	}
	t.Logf("Subnets remained in OVSDB: %v", len(subnets))
	for _, subnet := range subnets {
		subnetIds[subnet.UUID] = 1
	}

	rows, err = db.Query(sqlQueryListSubnet)
	if err != nil {
		t.Fatal("Fail to fetch subnets in SQL")
	}
	matchedRows = 0
	for rows.Next() {
		rows.Scan(&pendingId)
		if _, ok := subnetIds[pendingId]; !ok {
			t.Fatal("Inconsistency subnet ID between SQL and OVSDB")
		}
		matchedRows += 1
	}
	if matchedRows != len(subnets) {
		t.Fatal("Inconsistency subnet amount between SQL and OVSDB")
	}

	// Check consistency between OVSDB and SQL: Interface
	switchPorts := []nbdb.LogicalSwitchPort{}
	routerPorts := []nbdb.LogicalRouterPort{}
	_ = nbClient.List(context.Background(), &switchPorts)
	err = nbClient.List(context.Background(), &routerPorts)
	if err != nil {
		t.Fatal("Fail to fetch interfaces in OVSDB")
	}
	if len(switchPorts) != len(routerPorts) {
		t.Fatal("Inconsistency interface information OVSDB")
	}
	var numIfaces int
	err = db.QueryRow("SELECT COUNT(*) FROM router_interface").Scan(&numIfaces)
	if err != nil {
		t.Fatal("Fail to fetch interfaces in SQL")
	}
	if numIfaces != len(switchPorts) {
		t.Fatal("Inconsistency interface amount between SQL and OVSDB")
	}
	t.Logf("Interfaces remained in OVSDB: %v", numIfaces)
}

// Scenario: Verify that requests with invalid VPC field will not crash or block the server
func TestInvalidVpcRequest(t *testing.T) {
	vpcId1 := "5962d6a9-b5af-4f9b-9c82-bd7361f10cb3"

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

	// Construct 10 requests. 7 with good VPC IDs and 3 with bad ones
	chErr := make(chan bool, 10)
	var wg sync.WaitGroup
	wg.Add(10)
	vpcIds := []string{
		vpcId1,
		"64cf5f19-3e6c-45f4-a008-c55d8bd4951c", // does not exist
		vpcId1,
		"abcde", // wrong format
		vpcId1,
		vpcId1,
		"", // empty
		vpcId1,
		vpcId1,
		vpcId1,
	}
	for i := 0; i < 10; i++ {
		routerId := uuid.NewString()

		createRouterReq := &v1.CreateRouterRequest{
			RouterId: &v1.RouterId{Uuid: routerId},
			Name:     routerId,
			VpcId:    &v1.VPCId{Uuid: vpcIds[i]},
		}
		// Send async requests and count failure rate
		go func(req *v1.CreateRouterRequest) {
			defer wg.Done()
			_, err := CreateRouterHandler(db, nbClient, req)
			chErr <- (err != nil)
		}(createRouterReq)

	}

	wg.Wait()
	close(chErr)
	failures := 0
	for i := range chErr {
		if i {
			failures += 1
		}
	}
	if failures != 3 {
		t.Fatalf("Expect 3 failed requests but get %v.", failures)
	}

}
