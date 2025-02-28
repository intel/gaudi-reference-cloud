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

type LogicalSwitchPort struct {
	Uuid               UUID
	Name               string
	SwitchId           string
	AdminState         bool
	IpAddr             string
	External_IPaddress *string
	MAC                string
	SNATruleUUID       *UUID
}

func ListPortsHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.ListPortsRequest) (*v1.ListPortsResponse, error) {
	var err error = nil
	var ports []*v1.PortId
	Logger := Logger.WithName("ListPortsHandler")
	Logger.V(DebugLevel).Info("list ports")

	rows, err := db.Query(sqlQueryListPort)
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	var portId string
	for rows.Next() {
		err = rows.Scan(&portId)
		if err != nil {
			Logger.Error(err, "cannot scan SQL rows for port")
			return nil, GrpcErrorFromSql(err)
		}
		ports = append(ports, &v1.PortId{
			Uuid: portId,
		})
	}

	return &v1.ListPortsResponse{PortIds: ports}, err
}

func GetPortHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.GetPortRequest) (*v1.GetPortResponse, error) {
	var err error = nil
	uuid := r.GetPortId().GetUuid()
	Logger := Logger.WithName("GetPortHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("get port %s", uuid))

	// Locate the port using the Id
	var name, subnetId, chassisId, ipAddr string
	var mac, extIpAddr, extMac, isNat sql.NullString
	var deviceId uint32
	var enabled bool
	err = db.QueryRow(sqlQueryGetPort, uuid).Scan(&name, &subnetId, &ipAddr, &mac,
		&chassisId, &deviceId, &enabled, &isNat, &extIpAddr, &extMac)
	if err == sql.ErrNoRows {
		err = status.Errorf(codes.NotFound, "cannot find requested port")
		return nil, err
	}
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}
	// TODO: Compare query results with OVSDB facts
	port := &v1.Port{
		Id:        r.PortId,
		Name:      name,
		SubnetId:  &v1.SubnetId{Uuid: subnetId},
		IPAddress: ipAddr,
		ChassisId: chassisId,
		DeviceId:  deviceId,
		IsEnabled: enabled,
	}
	if mac.Valid {
		port.MACAddress = &mac.String
	}
	if isNat.Valid {
		port.IsNAT = &isNat.String
	}
	if extIpAddr.Valid {
		port.External_IPAddress = &extIpAddr.String
	}
	if extMac.Valid {
		port.External_MAC = &extMac.String
	}
	return &v1.GetPortResponse{Port: port}, err
}

func CreatePortHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.CreatePortRequest) (*v1.CreatePortResponse, error) {

	var err error = nil
	SwitchId := r.GetSubnetId().GetUuid()
	Logger := Logger.WithName("CreatePortHandler")

	var vpcId string
	err = db.QueryRow(sqlQueryVpcIdFromSubnet, SwitchId).Scan(&vpcId)
	if err != nil {
		Logger.Error(err, "cannot find VPC")
		return nil, GrpcErrorFromSql(err)
	}
	mu := GetVpcMutex(vpcId)
	LockVpcMutex(mu, vpcId)
	defer UnlockVpcMutex(mu, vpcId)

	Logger.V(DebugLevel).Info(
		fmt.Sprintf("create port for subnet %s on chassis %s dev %d", SwitchId, r.ChassisId, r.DeviceId))

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

	port := LogicalSwitchPort{
		Uuid:       UUID(r.GetPortId().GetUuid()),
		Name:       r.ChassisId + fmt.Sprintf("@port%d", r.DeviceId),
		SwitchId:   SwitchId,
		AdminState: r.IsEnabled,
		IpAddr:     r.Internal_IPAddress,
		MAC:        r.MACAddress,
	}

	router := LogicalRouter{}
	var routerUuid *UUID

	// Write the port to the database
	_, err = tx.Exec(sqlQueryCreatePort,
		port.Uuid,
		port.Name,
		SwitchId,
		r.ChassisId,
		r.DeviceId,
		port.AdminState,
		nil, // MAC address is currently generated with a pattern defined in IPU boot script
		port.IpAddr,
		r.IsNAT,
		nil, // snat_rule_id
		nil, // external_ip_address
		r.External_MAC,
	)
	if err != nil {
		Logger.Error(err, "SQL insert error")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.QueryRow(sqlQueryGetSwitchRouterData, SwitchId).Scan(
		&routerUuid,
		&router.Name,
		&router.GatewayUuid,
		&router.Snat_IPAddress,
	)

	// Check if the query returned no rows
	if err == sql.ErrNoRows {
		// No router found for the subnet, this is not an error, continue without router data
		Logger.V(DebugLevel).Info(fmt.Sprintf("No router associated with subnet %s", SwitchId))
	} else if err != nil {
		// Handle other SQL errors
		Logger.Error(err, "SQL insert or query error")
		return nil, GrpcErrorFromSql(err)
	}

	ovnlogicalSwitchPort := nbdb.LogicalSwitchPort{
		Name:      port.Name,
		UUID:      string(port.Uuid),
		Enabled:   &r.IsEnabled,
		Addresses: []string{fmt.Sprintf("%s %s", r.MACAddress, port.IpAddr)},
	}
	if port.MAC == "unknown" {
		ovnlogicalSwitchPort.Addresses = []string{r.MACAddress, port.IpAddr}
	} else {
		ovnlogicalSwitchPort.Addresses = []string{fmt.Sprintf("%s %s", r.MACAddress, port.IpAddr)}
	}

	sw := nbdb.LogicalSwitch{UUID: SwitchId}
	err = libovsdbops.CreateOrUpdateLogicalSwitchPortsOnSwitch(ovnClient, &sw, &ovnlogicalSwitchPort)
	if err != nil {
		return &v1.CreatePortResponse{PortId: &v1.PortId{}}, fmt.Errorf("failed to create port %+v on subnet %s: %v",
			ovnlogicalSwitchPort, SwitchId, err)
	}
	Logger.V(DebugLevel).Info(fmt.Sprintln("The UUID of the created port is ", port.Uuid))

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	if r.IsNAT && routerUuid != nil {
		router.Uuid = *routerUuid
		Logger.V(DebugLevel).Info(fmt.Sprintln("Set NAT for port ", port.Uuid))
		_, err := setNAT(db, ovnClient, &port, &router, true)
		if err != nil {
			return &v1.CreatePortResponse{PortId: &v1.PortId{Uuid: string(port.Uuid)}}, err
		}
	}

	return &v1.CreatePortResponse{PortId: &v1.PortId{Uuid: string(port.Uuid)}}, nil
}
func DeletePortHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.DeletePortRequest) (*v1.DeletePortResponse, error) {
	Logger := Logger.WithName("DeletePortHandler")
	uuid := r.GetPortId().GetUuid()

	var vpcId string
	err := db.QueryRow(sqlQueryVpcIdFromPort, uuid).Scan(&vpcId)
	if err != nil {
		Logger.Error(err, "cannot find VPC")
		return nil, GrpcErrorFromSql(err)
	}
	mu := GetVpcMutex(vpcId)
	LockVpcMutex(mu, vpcId)
	defer UnlockVpcMutex(mu, vpcId)

	Logger.V(DebugLevel).Info(fmt.Sprintf("Delete port %s", uuid))

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

	port := LogicalSwitchPort{
		Uuid: UUID(uuid),
	}
	var routerUuid, GatewayUuid *UUID
	var routerName, Snat_IPAddress *string

	// Execute the query and populate the port struct
	err = tx.QueryRow(sqlQueryGetPortAndRouterData, uuid).Scan(
		&port.Name,
		&port.SwitchId,
		&port.AdminState,
		&port.IpAddr,
		&port.SNATruleUUID,
		&port.External_IPaddress,
		&routerUuid,
		&routerName,
		&GatewayUuid,
		&Snat_IPAddress,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			Logger.Error(err, fmt.Sprintf("No port found with UUID: %s", uuid))
			return nil, fmt.Errorf("No port found with UUID: %s", uuid)
		}
		Logger.Error(err, fmt.Sprintf("Error querying port with UUID %s: %v", uuid, err))
		return nil, GrpcErrorFromSql(err)
	}

	/* If the port has an SNAT rule assigned then reset NAT config */
	if port.SNATruleUUID != nil && routerUuid != nil {
		router := LogicalRouter{
			Uuid:           *routerUuid,
			Name:           *routerName,
			GatewayUuid:    GatewayUuid,
			Snat_IPAddress: Snat_IPAddress,
		}
		Logger.V(DebugLevel).Info(fmt.Sprintf("Reset NAT for port %s", port.Uuid))
		_, err := resetNAT(db, ovnClient, &port, &router, false)
		if err != nil {
			return &v1.DeletePortResponse{PortId: &v1.PortId{Uuid: uuid}}, err
		}
	}

	_, err = tx.Exec(sqlQueryDeletePortAndRefs, uuid)
	if err != nil {
		Logger.Error(err, "Error deleting port and its references")
		return nil, GrpcErrorFromSql(err)
	}

	ovnlogicalSwitchPort := nbdb.LogicalSwitchPort{
		Name: port.Name,
		UUID: uuid,
	}

	sw := nbdb.LogicalSwitch{UUID: port.SwitchId}
	err = libovsdbops.DeleteLogicalSwitchPorts(ovnClient, &sw, &ovnlogicalSwitchPort)
	if err != nil {
		return &v1.DeletePortResponse{PortId: &v1.PortId{}}, fmt.Errorf("failed to delete port %+v from subnet %s: %v",
			ovnlogicalSwitchPort, port.SwitchId, err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	Logger.V(DebugLevel).Info(fmt.Sprintf("Successfully deleted port %s", uuid))
	return &v1.DeletePortResponse{PortId: &v1.PortId{Uuid: uuid}}, nil
}

func UpdatePortHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.UpdatePortRequest) (*v1.UpdatePortResponse, error) {
	Logger := Logger.WithName("UpdatePortHandler")
	uuid := r.GetPortId().GetUuid()

	Logger.V(DebugLevel).Info(fmt.Sprintf("Update port %s", uuid))

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

	port := LogicalSwitchPort{
		Uuid: UUID(uuid),
	}

	var routerUuid, GatewayUuid *UUID
	var routerName, Snat_IPAddress *string
	// Execute the query and populate the port struct
	err = tx.QueryRow(sqlQueryGetPortAndRouterData, uuid).Scan(
		&port.Name,
		&port.SwitchId,
		&port.AdminState,
		&port.IpAddr,
		&port.SNATruleUUID,
		&port.External_IPaddress,
		&routerUuid,
		&routerName,
		&GatewayUuid,
		&Snat_IPAddress,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			Logger.Error(err, fmt.Sprintf("No port found with UUID: %s", uuid))
			return nil, fmt.Errorf("No port found with UUID: %s", uuid)
		}
		Logger.Error(err, fmt.Sprintf("Error querying port with UUID %s: %v", uuid, err))
		return nil, GrpcErrorFromSql(err)
	}

	// Handle NAT settings
	if r.IsNAT != nil && routerUuid != nil {
		router := LogicalRouter{
			Uuid:           *routerUuid,
			Name:           *routerName,
			GatewayUuid:    GatewayUuid,
			Snat_IPAddress: Snat_IPAddress,
		}

		if *r.IsNAT && port.SNATruleUUID == nil {
			// The request enables NAT and if is not already enabled and the subnet is connected to router.
			Logger.V(DebugLevel).Info(fmt.Sprintln("Set NAT for port ", uuid))
			resp, err := setNAT(db, ovnClient, &port, &router, true)
			if err != nil {
				return resp, err
			}
		} else if !(*r.IsNAT) && port.SNATruleUUID != nil {
			// The request disables NAT and it is enabled.
			Logger.V(DebugLevel).Info(fmt.Sprintln("Reset NAT for port ", uuid))
			resp, err := resetNAT(db, ovnClient, &port, &router, false)
			if err != nil {
				return resp, err
			}
		}
		// Do nothing if the request enables NAT and it is already enabled
		// or if the request disables NAT while it is already disabled
	}

	// Handle enabling or disabling the port
	if r.IsEnabled != nil {
		Logger.V(DebugLevel).Info(fmt.Sprintf("%s port %s", map[bool]string{true: "Enable", false: "Disable"}[*r.IsEnabled], uuid))

		// Update the admin_state_up field in the database

		_, err = tx.Exec(sqlQueryUpdateAdminState, *r.IsEnabled, uuid)
		if err != nil {
			Logger.Error(err, "SQL update error when updating admin_state_up")
			return nil, GrpcErrorFromSql(err)
		}

		ovnlogicalSwitchPort := nbdb.LogicalSwitchPort{
			UUID:    uuid,
			Name:    port.Name,
			Enabled: r.IsEnabled,
		}

		sw := nbdb.LogicalSwitch{UUID: port.SwitchId}
		err := libovsdbops.CreateOrUpdateLogicalSwitchPortsOnSwitch(ovnClient, &sw, &ovnlogicalSwitchPort)
		if err != nil {
			return &v1.UpdatePortResponse{PortId: &v1.PortId{}}, fmt.Errorf("failed to update port %+v on subnet %s: %v",
				ovnlogicalSwitchPort, port.SwitchId, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	Logger.V(DebugLevel).Info(fmt.Sprintf("Port %s updated successfully", uuid))
	return &v1.UpdatePortResponse{PortId: &v1.PortId{Uuid: uuid}}, nil
}
