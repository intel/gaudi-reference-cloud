// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"database/sql"
	fmt "fmt"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetVPCHandler(db *sql.DB, r *v1.GetVPCRequest) (*v1.GetVPCResponse, error) {
	var err error = nil
	uuid := r.GetVpcId().GetUuid()
	Logger := Logger.WithName("GetVPCHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("get vpc %s", r.VpcId.Uuid))

	tx, err := db.Begin()
	if err != nil {
		Logger.Error(err, "cannot begin SQL transaction")
		return nil, GrpcErrorFromSql(err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				Logger.Error(err, "failed to rollback transaction")
				return
			}
		}
	}()

	var vpcName, tenantId, regionId string
	err = tx.QueryRow(sqlQueryGetVpc, uuid).Scan(&vpcName, &tenantId, &regionId)
	if err == sql.ErrNoRows {
		err = status.Errorf(codes.NotFound, "cannot find requested vpc")
		return nil, err
	}
	if err != nil {
		Logger.Error(err, "SQL select error for vpc")
		return nil, GrpcErrorFromSql(err)
	}

	// Gather information from related network objects
	var routerIds []*v1.RouterId
	rows, err := tx.Query(sqlQueryGetVpcRouters, uuid)
	if err != nil {
		Logger.Error(err, "SQL select error for router")
		return nil, GrpcErrorFromSql(err)
	}
	for rows.Next() {
		var routerId string
		err = rows.Scan(&routerId)
		if err != nil {
			Logger.Error(err, "cannot scan SQL rows for router")
			return nil, GrpcErrorFromSql(err)
		}
		routerIds = append(routerIds, &v1.RouterId{
			Uuid: routerId,
		})
	}

	var subnetIds []*v1.SubnetId
	rows, err = tx.Query(sqlQueryGetVpcSubnets, uuid)
	if err != nil {
		Logger.Error(err, "SQL select error for subnet")
		return nil, GrpcErrorFromSql(err)
	}
	for rows.Next() {
		var subnetId string
		err = rows.Scan(&subnetId)
		if err != nil {
			Logger.Error(err, "cannot scan SQL rows for subnet")
			return nil, GrpcErrorFromSql(err)
		}
		subnetIds = append(subnetIds, &v1.SubnetId{
			Uuid: subnetId,
		})
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error for vpc")
		return nil, GrpcErrorFromSql(err)
	}
	return &v1.GetVPCResponse{
		Vpc: &v1.VPC{
			Id:       &v1.VPCId{Uuid: string(uuid)},
			Name:     vpcName,
			TenantId: string(tenantId),
			RegionId: string(regionId),
			Routers:  routerIds,
			Subnets:  subnetIds,
		}}, nil
}

func ListVPCsHandler(db *sql.DB, r *v1.ListVPCsRequest) (*v1.ListVPCsResponse, error) {
	var err error = nil
	var vpcs []*v1.VPCId
	Logger := Logger.WithName("ListVPCsHandler")
	Logger.V(DebugLevel).Info("list vpc")

	rows, err := db.Query(sqlQueryListVpc)
	if err != nil {
		Logger.Error(err, "SQL select error for vpc")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	var vpcId string
	for rows.Next() {
		err = rows.Scan(&vpcId)
		if err != nil {
			Logger.Error(err, "cannot scan SQL rows for vpc")
			return nil, GrpcErrorFromSql(err)
		}
		vpcs = append(vpcs, &v1.VPCId{
			Uuid: vpcId,
		})
	}

	return &v1.ListVPCsResponse{VpcIds: vpcs}, nil
}

func CreateVPCHandler(db *sql.DB, r *v1.CreateVPCRequest) (*v1.CreateVPCResponse, error) {
	var err error = nil
	uuid := r.GetVpcId().GetUuid()
	resp := &v1.CreateVPCResponse{VpcId: &v1.VPCId{Uuid: uuid}}
	Logger := Logger.WithName("CreateVPCHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("create vpc %s", uuid))

	tx, err := db.Begin()
	if err != nil {
		Logger.Error(err, "cannot begin SQL transaction")
		return nil, GrpcErrorFromSql(err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				Logger.Error(err, "failed to rollback transaction")
				return
			}
		}
	}()

	_, err = tx.Exec(sqlQueryCreateVpc, uuid, r.Name, r.TenantId, r.RegionId)
	if err != nil {
		Logger.Error(err, "SQL insert error for vpc")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error for vpc")
		return nil, GrpcErrorFromSql(err)
	}
	// Create a lock for all future write operations within this VPC
	GetVpcMutex(uuid)
	return resp, nil
}

func DeleteVPCHandler(db *sql.DB, r *v1.DeleteVPCRequest) (*v1.DeleteVPCResponse, error) {
	var err error = nil
	uuid := r.GetVpcId().GetUuid()
	Logger := Logger.WithName("DeleteVPCHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("delete vpc %s", uuid))
	resp := &v1.DeleteVPCResponse{VpcId: &v1.VPCId{Uuid: uuid}}

	mu := GetVpcMutex(uuid)
	LockVpcMutex(mu, uuid)
	defer UnlockVpcMutex(mu, uuid)

	tx, err := db.Begin()
	if err != nil {
		Logger.Error(err, "cannot begin SQL transaction")
		return nil, GrpcErrorFromSql(err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				Logger.Error(err, "failed to rollback transaction")
				return
			}
		}
	}()

	res, err := tx.Exec(sqlQueryDeleteVpc, uuid)
	if err != nil {
		Logger.Error(err, "SQL delete error for vpc")
		return nil, GrpcErrorFromSql(err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		Logger.Error(err, "SQL delete error for vpc")
		return nil, GrpcErrorFromSql(err)
	}
	if count < 1 {
		errMsg := fmt.Sprintf("No vpc found with UUID: %s", uuid)
		Logger.Error(err, errMsg)
		return nil, status.Errorf(codes.NotFound, errMsg)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error for vpc")
		return nil, GrpcErrorFromSql(err)
	}

	// If deletion is success, remove the VPC lock from the sync map
	RemoveVpcMutex(uuid)
	return resp, nil
}
