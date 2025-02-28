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

func GetStaticRouteHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.GetStaticRouteRequest) (*v1.GetStaticRouteResponse, error) {
	var err error = nil
	uuid := r.GetStaticRouteId().GetUuid()
	Logger := Logger.WithName("GetStaticRouteHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("get static route %s", uuid))

	// Locate the static route using the Id
	var prefix, nexthop, routerId string
	err = db.QueryRow(sqlQueryGetStaticRoute, uuid).Scan(&prefix, &nexthop, &routerId)
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}
	staticRoute := &v1.StaticRoute{
		Id:       r.StaticRouteId,
		Prefix:   prefix,
		Nexthop:  nexthop,
		RouterId: &v1.RouterId{Uuid: routerId},
	}
	return &v1.GetStaticRouteResponse{StaticRoute: staticRoute}, err
}

func ListStaticRoutesHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.ListStaticRoutesRequest) (*v1.ListStaticRoutesResponse, error) {
	var err error = nil
	var staticRoutes []*v1.StaticRouteId
	Logger := Logger.WithName("ListStaticRoutesHandler")
	Logger.V(DebugLevel).Info("list static routes")

	rows, err := db.Query(sqlQueryListStaticRoute)
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	var staticRouteId string
	for rows.Next() {
		err = rows.Scan(&staticRouteId)
		if err != nil {
			Logger.Error(err, "cannot scan SQL rows for static route")
			return nil, GrpcErrorFromSql(err)
		}
		staticRoutes = append(staticRoutes, &v1.StaticRouteId{
			Uuid: staticRouteId,
		})
	}
	return &v1.ListStaticRoutesResponse{StaticRouteIds: staticRoutes}, err
}

func CreateStaticRouteHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.CreateStaticRouteRequest) (*v1.CreateStaticRouteResponse, error) {
	var err error = nil
	Logger := Logger.WithName("CreateStaticRouteHandler")

	staticRouteUuid := r.GetStaticRouteId().GetUuid()
	routerUuid := r.GetRouterId().GetUuid()

	var vpcId string
	err = db.QueryRow(sqlQueryVpcIdFromRouter, routerUuid).Scan(&vpcId)
	if err != nil {
		Logger.Error(err, "cannot find VPC")
		return nil, GrpcErrorFromSql(err)
	}
	mu := GetVpcMutex(vpcId)
	LockVpcMutex(mu, vpcId)
	defer UnlockVpcMutex(mu, vpcId)

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

	_, err = tx.Exec(sqlQueryCreateStaticRoute, staticRouteUuid, r.Prefix, r.Nexthop, routerUuid)
	if err != nil {
		Logger.Error(err, "SQL insert error")
		return nil, GrpcErrorFromSql(err)
	}

	// Get the router name string from UUID
	var routerName string
	err = tx.QueryRow(sqlQueryGetRouter, routerUuid).Scan(&routerName, &vpcId)
	if err != nil {
		Logger.Error(err, "Cannot find the associated router")
		return nil, GrpcErrorFromSql(err)
	}

	// Create the static route in the backend
	staticRoute := nbdb.LogicalRouterStaticRoute{
		IPPrefix: r.Prefix,
		Nexthop:  r.Nexthop,
		UUID:     staticRouteUuid,
	}
	predicate := func(item *nbdb.LogicalRouterStaticRoute) bool { return item.IPPrefix == r.Prefix }
	err = libovsdbops.CreateOrReplaceLogicalRouterStaticRouteWithPredicate(ovnClient, routerName, &staticRoute, predicate)

	if err != nil {
		err = fmt.Errorf("could not create static route: %w ", err)

		err = status.Errorf(
			codes.InvalidArgument,
			err.Error(),
		)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}
	Logger.V(DebugLevel).Info(fmt.Sprintln("The UUID of the static route is ", staticRouteUuid))
	return &v1.CreateStaticRouteResponse{StaticRouteId: &v1.StaticRouteId{Uuid: staticRouteUuid}}, err
}

func DeleteStaticRouteHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.DeleteStaticRouteRequest) (*v1.DeleteStaticRouteResponse, error) {
	var err error = nil
	staticRouteUuid := r.GetStaticRouteId().GetUuid()
	Logger := Logger.WithName("DeleteStaticRouteHandler")

	var vpcId string
	err = db.QueryRow(sqlQueryVpcIdFromStaticRoute, staticRouteUuid).Scan(&vpcId)
	if err != nil {
		Logger.Error(err, "cannot find VPC")
		return nil, GrpcErrorFromSql(err)
	}
	mu := GetVpcMutex(vpcId)
	LockVpcMutex(mu, vpcId)
	defer UnlockVpcMutex(mu, vpcId)

	Logger.V(DebugLevel).Info(fmt.Sprintf("static route delete: %s", staticRouteUuid))

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

	var routerUuid string
	err = tx.QueryRow(sqlQueryDeleteStaticRoute, staticRouteUuid).Scan(&routerUuid)
	if err != nil {
		Logger.Error(err, "SQL delete error")
		return nil, GrpcErrorFromSql(err)
	}

	// Get the router name string from UUID
	var routerName string
	err = tx.QueryRow(sqlQueryGetRouter, routerUuid).Scan(&routerName, &vpcId)
	if err != nil {
		Logger.Error(err, "Cannot find the associated router")
		return nil, GrpcErrorFromSql(err)
	}

	// invoke backend to program the object
	staticRoute := nbdb.LogicalRouterStaticRoute{
		UUID: staticRouteUuid,
	}
	err = libovsdbops.DeleteLogicalRouterStaticRoutes(ovnClient, routerName, &staticRoute)
	if err != nil {
		Logger.Error(err, "failed to delete static route in OVN")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}
	return &v1.DeleteStaticRouteResponse{StaticRouteId: &v1.StaticRouteId{Uuid: r.StaticRouteId.Uuid}}, err
}
