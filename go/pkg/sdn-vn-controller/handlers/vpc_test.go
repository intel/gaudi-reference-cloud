// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"context"
	"testing"

	utils "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/tests/utils"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"
)

func TestCreateVPCHandler(t *testing.T) {
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})

	fakeUuid := "f132dc6b-195c-4816-aeed-a3e2017e0f6e"
	fakeName := "fake vpc"
	fakeTenant := "12345"
	fakeRegion := "54321"
	fakeRequest := v1.CreateVPCRequest{
		VpcId:    &v1.VPCId{Uuid: fakeUuid},
		Name:     fakeName,
		TenantId: fakeTenant,
		RegionId: fakeRegion,
	}
	_, err = CreateVPCHandler(db, &fakeRequest)
	if err != nil {
		t.Fatal("Failed to create VPC", err)
	}
	// Check db for the new row
	var name, tenantId, regionId string
	err = db.QueryRow(sqlQueryGetVpc, fakeUuid).Scan(&name, &tenantId, &regionId)
	if err != nil {
		t.Fatal("SQL row not found", err)
	}
	if tenantId != fakeTenant || name != fakeName || regionId != fakeRegion {
		t.Fatal("Inserted data not correct", err)
	}
}

func TestGetVPCHandler(t *testing.T) {
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})

	fakeUuid := "f132dc6b-195c-4816-aeed-a3e2017e0f6e"
	fakeName := "fake vpc"
	fakeTenant := "12345"
	fakeRegion := "54321"
	_, err = db.Exec(sqlQueryCreateVpc, fakeUuid, fakeName, fakeTenant, fakeRegion)
	if err != nil {
		t.Fatal("Failed to insert fake data to SQL", err)
	}

	res, err := GetVPCHandler(db, &v1.GetVPCRequest{
		VpcId: &v1.VPCId{Uuid: string(fakeUuid)},
	})
	if err != nil {
		t.Fatal("Errors during getting VPCs", err)
	}
	if res.Vpc.Name != fakeName || res.Vpc.TenantId != fakeTenant || res.Vpc.RegionId != fakeRegion {
		t.Fatal("Got incorrect VPC data", err)
	}
}

func TestListVPCsHandler(t *testing.T) {
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})

	fakeUuid := "f132dc6b-195c-4816-aeed-a3e2017e0f6e"
	fakeName := "fake vpc"
	fakeTenant := "12345"
	fakeRegion := "54321"
	_, err = db.Exec(sqlQueryCreateVpc, fakeUuid, fakeName, fakeTenant, fakeRegion)
	if err != nil {
		t.Fatal("Failed to insert fake data to SQL", err)
	}

	res, err := ListVPCsHandler(db, &v1.ListVPCsRequest{})
	if err != nil || len(res.VpcIds) < 1 || res.VpcIds[0].GetUuid() != fakeUuid {
		t.Fatal("Errors during listing VPCs", err)
	}
}

func TestDeleteVPCHandler(t *testing.T) {
	container, db, err := utils.NewTestDbClient("../db/migrations/frdb.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		container.Terminate(context.Background())
		db.Close()
	})

	// Test Case 1
	fakeUuid := "f132dc6b-195c-4816-aeed-a3e2017e0f6e"
	fakeName := "fake vpc"
	fakeTenant := "12345"
	fakeRegion := "54321"
	_, err = db.Exec(sqlQueryCreateVpc, fakeUuid, fakeName, fakeTenant, fakeRegion)
	if err != nil {
		t.Fatal("Failed to insert fake data to SQL", err)
	}

	_, err = DeleteVPCHandler(db, &v1.DeleteVPCRequest{
		VpcId: &v1.VPCId{Uuid: string(fakeUuid)},
	})
	if err != nil {
		t.Fatal("Errors during deleting VPCs", err)
	}
	rows, err := db.Query(sqlQueryGetVpc, fakeUuid)
	if err != nil || rows.Next() {
		t.Fatal("Failed to remove VPC from SQL", err)
	}

	// Test Case 2: VPC is referred and should not be deleted
	_, err = db.Exec(sqlQueryCreateVpc, fakeUuid, fakeName, fakeTenant, fakeRegion)
	if err != nil {
		t.Fatal("Failed to insert fake data to SQL", err)
	}
	_, err = db.Exec("INSERT INTO router (router_id, router_name, vpc_id) VALUES ($1, $2, $3)", "fe87e017-2fa1-4099-9028-a372d23ecdd6", "r1", fakeUuid)
	if err != nil {
		t.Fatal("Failed to insert fake data to SQL", err)
	}
	_, err = DeleteVPCHandler(db, &v1.DeleteVPCRequest{
		VpcId: &v1.VPCId{Uuid: string(fakeUuid)},
	})
	if err == nil {
		t.Fatal("Failed to reject deleting referred VPC")
	}

}
