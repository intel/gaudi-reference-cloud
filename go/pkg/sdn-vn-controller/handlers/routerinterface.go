// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"database/sql"
	"errors"
	"fmt"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	libovsdbclient "github.com/ovn-org/libovsdb/client"
	libovsdbops "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/libovsdb/ops"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ListRouterInterfacesHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.ListRouterInterfacesRequest) (*v1.ListRouterInterfacesResponse, error) {
	var err error = nil
	var interfaces []*v1.RouterInterfaceId
	Logger := Logger.WithName("ListRouterInterfacesHandler")
	Logger.V(DebugLevel).Info("list router interfaces")

	rows, err := db.Query(sqlQueryListRouterInterface)
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	var interfaceId string
	for rows.Next() {
		err = rows.Scan(&interfaceId)
		if err != nil {
			Logger.Error(err, "cannot scan SQL rows for router")
			return nil, GrpcErrorFromSql(err)
		}
		interfaces = append(interfaces, &v1.RouterInterfaceId{
			Uuid: interfaceId,
		})
	}

	return &v1.ListRouterInterfacesResponse{RouterInterfaceIds: interfaces}, err
}

func GetRouterInterfaceHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.GetRouterInterfaceRequest) (*v1.GetRouterInterfaceResponse, error) {
	var err error = nil
	uuid := r.GetRouterInterfaceId().GetUuid()
	Logger := Logger.WithName("GetRouterInterfaceHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("get router interface %s", uuid))

	var routerId, subnetId, ipStr, macStr string
	err = db.QueryRow(sqlQueryGetRouterInterface, uuid).Scan(&routerId, &subnetId, &ipStr, &macStr)
	if err == sql.ErrNoRows {
		err = status.Errorf(codes.NotFound, "cannot find requested router interface")
		return nil, err
	}
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}
	// TODO: Compare query results with OVSDB facts
	iface := &v1.RouterInterface{
		Id:            &v1.RouterInterfaceId{Uuid: uuid},
		RouterId:      &v1.RouterId{Uuid: routerId},
		SubnetId:      &v1.SubnetId{Uuid: subnetId},
		Interface_IP:  ipStr,
		Interface_MAC: macStr,
	}

	return &v1.GetRouterInterfaceResponse{RouterInterface: iface}, err
}

func CreateRouterInterfaceHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.CreateRouterInterfaceRequest) (*v1.CreateRouterInterfaceResponse, error) {
	var err error = nil
	ifUuid := r.GetRouterInterfaceId().GetUuid()
	routerUuid := r.GetRouterId().GetUuid()
	subnetUuid := r.GetSubnetId().GetUuid()
	lrpUuid := uuid.NewString()
	lspUuid := uuid.NewString()
	Logger := Logger.WithName("CreateRouterInterfaceHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("create router interface: %s/%s\n", r.RouterId, r.SubnetId))

	var vpcId string
	err = db.QueryRow(sqlQueryVpcIdFromRouter, routerUuid).Scan(&vpcId)
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
	var subnetName string
	router := LogicalRouter{
		Uuid: UUID(routerUuid),
	}

	_, err = tx.Exec(sqlQueryCreateRouterInterface, ifUuid, subnetUuid, router.Uuid, lrpUuid, lspUuid, r.Interface_IP, r.Interface_MAC)
	if err != nil {
		Logger.Error(err, "SQL insert error")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.QueryRow(sqlQueryGetSubnetName, subnetUuid).Scan(&subnetName)
	if err != nil {
		Logger.Error(err, "SQL select error for subnet")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.QueryRow(sqlQueryGetRouterGWData, router.Uuid).Scan(&router.Name, &router.GatewayUuid, &router.Snat_IPAddress)
	if err != nil {
		Logger.Error(err, "SQL select error for router")
		return nil, GrpcErrorFromSql(err)
	}

	// Call OVN backend to create objects
	// TODO: Evaluate if it is necessary to use an atomic OVSDB transaction creating both ports
	lrpName := router.Name + "_" + subnetName
	lrpNetworks := []string{r.Interface_IP}
	logicalRouterPort := nbdb.LogicalRouterPort{
		Name:     lrpName,
		MAC:      r.Interface_MAC,
		Networks: lrpNetworks,
		UUID:     lrpUuid,
	}
	/*
		// TODO: We need the chassis ID to set the gateway-chassis, which in effect pins
		// the logical switch to the current node in OVN. Otherwise, ovn-controller will
		// flood-fill unrelated datapaths unnecessarily, causing scale problems.

				gatewayChassis := nbdb.GatewayChassis{
					Name:        lrpName + "-" + chassisID,
					ChassisName: chassisID,
					Priority:    1,
				}
	*/

	logicalRouter := nbdb.LogicalRouter{UUID: routerUuid}
	err = libovsdbops.CreateOrUpdateLogicalRouterPort(ovnClient, &logicalRouter, &logicalRouterPort,
		nil, // gatewayChassis
		&logicalRouterPort.MAC, &logicalRouterPort.Networks)

	if err != nil {
		err = fmt.Errorf("could not create router to subnet interface %s / %s: %w ", r.RouterId, r.SubnetId, err)
		return nil, err
	}

	// Create the switch port - switch to router
	lspName := subnetName + "_" + router.Name
	logicalSwitchPort := nbdb.LogicalSwitchPort{
		Name:      lspName,
		Type:      "router",
		Addresses: []string{"router"},
		UUID:      lspUuid,
		Options: map[string]string{
			"router-port": lrpName,
		},
	}
	logicalSwitch := nbdb.LogicalSwitch{UUID: subnetUuid}
	err = libovsdbops.CreateOrUpdateLogicalSwitchPortsOnSwitch(ovnClient, &logicalSwitch, &logicalSwitchPort)
	if err != nil {
		err = fmt.Errorf("could not create port on subnet %s / %s: %w", r.SubnetId, r.RouterId, err)
		return nil, err
	}

	// Fetch all NAT-enabled ports in the subnet
	var ports []LogicalSwitchPort
	rows, err := tx.Query(sqlQueryGetNATPortsInSubnet, subnetUuid)
	if err != nil {
		Logger.Error(err, "SQL query error")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	for rows.Next() {
		port := LogicalSwitchPort{}
		err := rows.Scan(&port.Uuid, &port.Name, &port.AdminState, &port.IpAddr, &port.SNATruleUUID, &port.External_IPaddress)
		if err != nil {
			Logger.Error(err, "SQL scan error")
			return nil, GrpcErrorFromSql(err)
		}
		ports = append(ports, port)
	}

	if err = rows.Err(); err != nil {
		Logger.Error(err, "Error during row iteration")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	// Handle NAT settings for each port
	for _, port := range ports {
		if port.SNATruleUUID == nil {
			Logger.V(DebugLevel).Info(fmt.Sprintln("Set NAT for port ", port.Uuid))
			_, err := setNAT(db, ovnClient, &port, &router, true)
			if err != nil {
				return nil, err
			}
		}
	}

	// Return the UUID and other information
	return &v1.CreateRouterInterfaceResponse{RouterInterfaceId: r.RouterInterfaceId}, nil
}

func DeleteRouterInterfaceHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.DeleteRouterInterfaceRequest) (*v1.DeleteRouterInterfaceResponse, error) {
	ifUuid := r.GetRouterInterfaceId().GetUuid()
	Logger := Logger.WithName("DeleteRouterInterfaceHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("delete router interface: %s\n", r.RouterInterfaceId))

	var vpcId string
	err := db.QueryRow(sqlQueryVpcIdFromRouterInterface, ifUuid).Scan(&vpcId)
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

	// Delete in SQL and retrieve information
	var subnetUuid, lrpUuid, lspUuid string
	router := LogicalRouter{}

	err = tx.QueryRow(sqlQueryDeleteRouterInterface, ifUuid).Scan(&router.Uuid, &subnetUuid,
		&lrpUuid, &lspUuid, &router.Name, &router.GatewayUuid, &router.Snat_IPAddress)
	if err != nil {
		Logger.Error(err, "SQL delete error")
		return nil, GrpcErrorFromSql(err)
	}

	// Call backend to delete ports
	logicalSwitch := nbdb.LogicalSwitch{UUID: subnetUuid}
	logicalSwitchPort := nbdb.LogicalSwitchPort{UUID: lspUuid}
	err = libovsdbops.DeleteLogicalSwitchPorts(ovnClient, &logicalSwitch, &logicalSwitchPort)
	if err != nil && !errors.Is(err, libovsdbclient.ErrNotFound) {
		err = fmt.Errorf("failed to delete logical switch port %s from switch %s: %w",
			lspUuid, subnetUuid, err)
		return nil, err
	}
	logicalRouter := nbdb.LogicalRouter{UUID: string(router.Uuid)}
	logicalRouterPort := nbdb.LogicalRouterPort{UUID: lrpUuid}
	err = libovsdbops.DeleteLogicalRouterPorts(ovnClient, &logicalRouter, &logicalRouterPort)
	if err != nil && !errors.Is(err, libovsdbclient.ErrNotFound) {
		err = fmt.Errorf("failed to delete port %s on router %s: %w",
			lrpUuid, router.Uuid, err)
		return nil, err
	}

	// Fetch all NAT-enabled ports in the subnet
	var ports []LogicalSwitchPort
	rows, err := tx.Query(sqlQueryGetNATPortsInSubnet, subnetUuid)
	if err != nil {
		Logger.Error(err, "SQL query error")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	for rows.Next() {
		port := LogicalSwitchPort{}
		err := rows.Scan(&port.Uuid, &port.Name, &port.AdminState, &port.IpAddr, &port.SNATruleUUID, &port.External_IPaddress)
		if err != nil {
			Logger.Error(err, "SQL scan error")
			return nil, GrpcErrorFromSql(err)
		}
		ports = append(ports, port)
	}

	if err = rows.Err(); err != nil {
		Logger.Error(err, "Error during row iteration")
		return nil, GrpcErrorFromSql(err)
	}

	// Handle NAT settings based on isNAT value
	for _, port := range ports {
		if port.SNATruleUUID != nil {
			Logger.V(DebugLevel).Info(fmt.Sprintln("Reset NAT for port ", port.Uuid))
			_, err := resetNAT(db, ovnClient, &port, &router, true)
			if err != nil {
				return nil, err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	return &v1.DeleteRouterInterfaceResponse{RouterInterfaceId: r.RouterInterfaceId}, nil
}
