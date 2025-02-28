// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"database/sql"
	"fmt"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	libovsdbclient "github.com/ovn-org/libovsdb/client"
	libovsdbops "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/libovsdb/ops"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deprecated: would be replaced by DB.
type LogicalRouter struct {
	Uuid           UUID
	Name           string
	GatewayUuid    *UUID
	Snat_IPAddress *string
	VpcId          UUID
	NATruleCount   int
}

func ListRoutersHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.ListRoutersRequest) (*v1.ListRoutersResponse, error) {
	var err error = nil
	var routers []*v1.RouterId
	Logger := Logger.WithName("ListRoutersHandler")
	Logger.V(DebugLevel).Info("list routers")

	rows, err := db.Query(sqlQueryListRouter)
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	var routerId string
	for rows.Next() {
		err = rows.Scan(&routerId)
		if err != nil {
			Logger.Error(err, "cannot scan SQL rows for router")
			return nil, GrpcErrorFromSql(err)
		}
		routers = append(routers, &v1.RouterId{
			Uuid: routerId,
		})
	}

	return &v1.ListRoutersResponse{RouterIds: routers}, err
}

func GetRouterHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.GetRouterRequest) (*v1.GetRouterResponse, error) {
	var err error = nil
	uuid := r.GetRouterId().GetUuid()
	Logger := Logger.WithName("GetRouterHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("get router %s", uuid))

	// Locate the router using the Id
	var name, vpcId string
	err = db.QueryRow(sqlQueryGetRouter, uuid).Scan(&name, &vpcId)
	if err == sql.ErrNoRows {
		err = status.Errorf(codes.NotFound, "cannot find requested router")
		return nil, err
	}
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}
	router := &v1.Router{
		Id:    r.RouterId,
		Name:  name,
		VpcId: &v1.VPCId{Uuid: vpcId},
	}
	return &v1.GetRouterResponse{Router: router}, err
}

func CreateRouterHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.CreateRouterRequest) (*v1.CreateRouterResponse, error) {
	var err error = nil
	Logger := Logger.WithName("CreateRouterHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("create router: %s", r.Name))

	uuid := r.GetRouterId().GetUuid()
	vpcId := r.GetVpcId().GetUuid()

	// Validate vpc_id and fetch the lock
	var vpcExists bool
	err = db.QueryRow(sqlQueryCheckVpcExists, vpcId).Scan(&vpcExists)
	if err != nil {
		Logger.Error(err, "SQL query error when checking VPC existence")
		return nil, GrpcErrorFromSql(err)
	}
	if !vpcExists {
		Logger.Error(err, "Attempts to fetch a non-existing VPC lock.")
		return nil, status.Errorf(codes.NotFound, "VPC not found.")
	}
	mu := GetVpcMutex(vpcId)
	LockVpcMutex(mu, vpcId)
	defer UnlockVpcMutex(mu, vpcId)

	ovnlogicalRouter := nbdb.LogicalRouter{
		Name: r.Name,
		UUID: uuid,
	}

	// Try inserting data in SQL
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

	_, err = tx.Exec(sqlQueryCreateRouter, uuid, r.Name, vpcId)
	if err != nil {
		Logger.Error(err, "SQL insert error")
		return nil, GrpcErrorFromSql(err)
	}

	// Invoke backend to program the object
	err = libovsdbops.CreateOrUpdateLogicalRouter(ovnClient, &ovnlogicalRouter)
	if err != nil {
		err = fmt.Errorf("could not create router %s: %w ", r.Name, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	// Deprecated: remove this after SQL is fully integrated into all related handlers
	logicalRouter := LogicalRouter{
		Uuid:  UUID(uuid),
		Name:  r.Name,
		VpcId: UUID(vpcId),
	}
	Logger.V(DebugLevel).Info(fmt.Sprintf("The UUID of the router is %s", logicalRouter.Uuid))

	Logger.V(DebugLevel).Info(fmt.Sprintf("The UUID of the router is %s", logicalRouter.Uuid))
	return &v1.CreateRouterResponse{RouterId: &v1.RouterId{Uuid: string(logicalRouter.Uuid)}}, err
}

func DeleteRouterHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.DeleteRouterRequest) (*v1.DeleteRouterResponse, error) {
	var err error = nil
	uuid := r.GetRouterId().GetUuid()
	Logger := Logger.WithName("DeleteRouterHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("router delete: %s\n", r.RouterId.Uuid))

	var vpcId string
	err = db.QueryRow(sqlQueryVpcIdFromRouter, uuid).Scan(&vpcId)
	if err != nil {
		Logger.Error(err, "cannot find VPC")
		return nil, GrpcErrorFromSql(err)
	}
	mu := GetVpcMutex(vpcId)
	LockVpcMutex(mu, vpcId)
	defer UnlockVpcMutex(mu, vpcId)

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

	var routerName string
	err = tx.QueryRow(sqlQueryDeleteRouter, uuid).Scan(&routerName)
	if err != nil {
		Logger.Error(err, "SQL delete error")
		return nil, GrpcErrorFromSql(err)
	}

	// invoke backend to program the object
	logicalRouter := nbdb.LogicalRouter{
		UUID: string(uuid),
	}
	err = libovsdbops.DeleteLogicalRouter(ovnClient, &logicalRouter)
	if err != nil {
		Logger.Error(err, "failed to delete router in OVN")
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	return &v1.DeleteRouterResponse{RouterId: &v1.RouterId{Uuid: r.RouterId.Uuid}}, err
}
