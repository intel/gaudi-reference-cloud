// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	libovsdbclient "github.com/ovn-org/libovsdb/client"
	libovsdbops "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/libovsdb/ops"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
	log "github.com/sirupsen/logrus"
)

type GwService int

const (
	INTERNET GwService = iota
	STORAGE
)

/* Gateway structure stores gateway informations parsed from the config file
 * A gateway is assigned to a single chassis
* Access network is the localnet subnet connecting the gatewat to the TOR.
*/
type Gateway struct {
	Uuid                UUID   `koanf:"Uuid"`
	ChassisId           int    `koanf:"ChassisId"`
	Nexthop             string `koanf:"Nexthop"`
	TargetNetwork       string `koanf:"TargetNetwork"`
	AccessNetwork       string `koanf:"AccessNetwork"`
	Service             string `koanf:"Service"`
	MaxNATs             int    `koanf:"MaxNATs"`
	Name                string
	SubnetId            UUID
	Svc                 GwService
	NATipManager        *IPManager
	AccessNetworkPrefix string
	ExtPortName         string
	ExtNetworkName      string
}

/*
func (s *GwService) UnmarshalJSON(data []byte) error {
	var serviceStr string
	if err := json.Unmarshal(data, &serviceStr); err != nil {
		return err
	}

	switch serviceStr {
	case "INTERNET":
		*s = INTERNET
	case "STORAGE":
		*s = STORAGE
	default:
		return fmt.Errorf("unknown service type: %s", serviceStr)
	}
	return nil
}
*/

var (
	internet_gw_RR_Index = 0
	gateways             = make(map[GwService](map[UUID]*Gateway))
	gatewayKeys          = make(map[GwService][]UUID)
)

func generateMACFromUUID(uuidStr UUID) (string, error) {
	hasher := sha256.New()
	if _, err := hasher.Write([]byte(uuidStr)); err != nil {
		return "", err
	}
	hash := hasher.Sum(nil)

	mac := make([]byte, 6)
	copy(mac, hash[:6])
	mac[0] |= 0x02
	mac[0] &= 0xfe
	// TODO: Add OUI prefix
	macStr := hex.EncodeToString(mac)
	return formatMAC(macStr), nil
}

func formatMAC(macStr string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s",
		macStr[0:2], macStr[2:4], macStr[4:6],
		macStr[6:8], macStr[8:10], macStr[10:12])
}

func selectGW(service GwService) (*Gateway, error) {
	keys, exists := gatewayKeys[service]
	if !exists || len(keys) == 0 {
		return nil, fmt.Errorf("no gateways available for service: %d", service)
	}

	// Select the gateway using round-robin
	selectedKey := keys[internet_gw_RR_Index]
	internet_gw_RR_Index = (internet_gw_RR_Index + 1) % len(keys)

	return gateways[service][selectedKey], nil
}

func CreateGateways(db *sql.DB, ovnClient libovsdbclient.Client, gws []Gateway) error {
	Logger := Logger.WithName("CreateGateways")
	Logger.V(DebugLevel).Info("Create Gateways")
	var err error

	gateways[INTERNET] = make(map[UUID]*Gateway)

	for _, gw_iter := range gws {
		gw := gw_iter
		gw.Name = "gw-" + strconv.Itoa(gw.ChassisId)
		gw.NATipManager, err = NewIPManager(db, gw.Uuid, gw.AccessNetwork, 0,
			gw.MaxNATs, gw.Nexthop)
		if err != nil {
			Logger.Error(err, "Error initializing NAT IP manager")
			// TODO: Need to figure out what to do with the log.Fatalf
			log.Fatalf("Error initializing NAT IP manager: %v", err)
		}

		prefixParts := strings.Split(gw.AccessNetwork, "/")
		if len(prefixParts) != 2 {
			return fmt.Errorf("invalid AccessNetwork format: %s", gw.AccessNetwork)
		}
		gw.AccessNetworkPrefix = "/" + prefixParts[1]
		switch gw.Service {
		case "INTERNET":
			gw.Svc = INTERNET
		case "STORAGE":
			gw.Svc = STORAGE
		default:
			return fmt.Errorf("unknown service type: %s", gw.Service)
		}
		serviceType := gw.Svc
		if serviceType == INTERNET {
			gateways[INTERNET][gw.Uuid] = &gw
		}
	}

	tx, err := db.Begin()
	if err != nil {
		Logger.Error(err, "cannot begin SQL transaction")
		return GrpcErrorFromSql(err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				Logger.Error(err, "failed to rollback transaction")
				return
			}
		}
	}()

	for k, gw := range gateways[INTERNET] {
		Logger.V(DebugLevel).Info(fmt.Sprintf("Processing gateway for ChassisId: %d", gw.ChassisId))

		// Check if the gateway already exists in the database
		var existingGatewayID UUID
		err := tx.QueryRow("SELECT gateway_id FROM gateway WHERE gateway_id = $1", gw.Uuid).Scan(&existingGatewayID)

		if err != nil {
			if err == sql.ErrNoRows {
				if err := createGatewayOvnLogicalComponents(gw, ovnClient, tx); err != nil {
					return err
				}
			} else {
				// Handle other SQL errors
				Logger.Error(err, "SQL query error when checking gateway existence")
				return GrpcErrorFromSql(err)
			}
		} else {
			// Gateway exists, fetch existing data and populate the gw structure
			Logger.V(DebugLevel).Info(fmt.Sprintf("Gateway already exists in DB, fetching existing data for ChassisId: %d", gw.ChassisId))

			err = tx.QueryRow("SELECT subnet_id, ext_port_name, ext_network_name FROM gateway WHERE gateway_id = $1", gw.Uuid).Scan(
				&gw.SubnetId,
				&gw.ExtPortName,
				&gw.ExtNetworkName,
			)
			if err != nil {
				Logger.Error(err, "SQL query error when fetching existing gateway data")
				return GrpcErrorFromSql(err)
			}
		}

		gatewayKeys[INTERNET] = append(gatewayKeys[INTERNET], k)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return GrpcErrorFromSql(err)
	}
	return nil
}

func createGatewayOvnLogicalComponents(gw *Gateway, ovnClient libovsdbclient.Client, tx *sql.Tx) error {
	Logger.V(DebugLevel).Info(fmt.Sprintf("Gateway not found in DB, creating new gateway for ChassisId: %d", gw.ChassisId))

	// Gateway does not exist, create the logical switch and port
	subnetName := gw.Name
	logicalSwitch := nbdb.LogicalSwitch{
		Name: subnetName,
	}

	err := libovsdbops.CreateOrUpdateLogicalSwitch(ovnClient, &logicalSwitch)
	if err != nil {
		return fmt.Errorf("could not create subnet for GW %d: %w", gw.ChassisId, err)
	}

	gw.SubnetId = UUID(logicalSwitch.UUID)
	Logger.V(DebugLevel).Info(fmt.Sprintf("The UUID of the created subnet is %s", string(gw.SubnetId)))

	gw.ExtPortName = subnetName + "-port"
	gw.ExtNetworkName = "ext-" + subnetName

	ovnlogicalSwitchPort := nbdb.LogicalSwitchPort{
		Name:    gw.ExtPortName,
		Type:    "localnet",
		Options: map[string]string{"network_name": gw.ExtNetworkName},
	}

	sw := nbdb.LogicalSwitch{UUID: string(gw.SubnetId)}

	err = libovsdbops.CreateOrUpdateLogicalSwitchPortsOnSwitch(ovnClient, &sw, &ovnlogicalSwitchPort)
	if err != nil {
		return fmt.Errorf("failed to create port %+v on subnet %s: %v", ovnlogicalSwitchPort, gw.SubnetId, err)
	}

	// Insert the new gateway into the database
	_, err = tx.Exec(sqlQueryCreateGateway,
		gw.Uuid,
		gw.ChassisId,
		gw.Name,
		gw.SubnetId,
		gw.ExtPortName,
		gw.ExtNetworkName,
		gw.AccessNetwork,
		gw.MaxNATs,
	)

	if err != nil {
		Logger.Error(err, "SQL insert error for gateway")
		return GrpcErrorFromSql(err)
	}

	return nil
}

func setNAT(db *sql.DB, ovnClient libovsdbclient.Client,
	port *LogicalSwitchPort, router *LogicalRouter, NATflag bool) (*v1.UpdatePortResponse, error) {

	if router.GatewayUuid == nil {
		Logger.V(DebugLevel).Info(fmt.Sprintln("Connect router ", router.Uuid, " to a GW"))
		gw, err := selectGW(INTERNET)
		if err != nil {
			return &v1.UpdatePortResponse{PortId: &v1.PortId{}},
				fmt.Errorf("no GW selected for Router %s", string(router.Uuid))
		}
		err = attachRouterToGW(db, ovnClient, router, gw)
		if err != nil {
			return &v1.UpdatePortResponse{PortId: &v1.PortId{}},
				fmt.Errorf("could not attach Router %s to GW %d : %w",
					router.Uuid, gw.ChassisId, err)
		}
	}
	err := addNATRule(db, ovnClient, port, router, NATflag)
	if err != nil {
		return &v1.UpdatePortResponse{PortId: &v1.PortId{}},
			fmt.Errorf("could not add NAT rule to Router %s: %w",
				router.Uuid, err)
	}
	return &v1.UpdatePortResponse{
		PortId:         &v1.PortId{},
		Snat_IPAddress: router.Snat_IPAddress,
	}, err
}

func resetNAT(db *sql.DB, ovnClient libovsdbclient.Client,
	port *LogicalSwitchPort, router *LogicalRouter, NATflag bool) (*v1.UpdatePortResponse, error) {

	err := deleteNATRule(db, ovnClient, port, router, NATflag)
	if err != nil {
		Logger.Info(fmt.Sprintln("could not delete NAT rule from Router ", router.Uuid))
	}

	Logger.Info(fmt.Sprintln("New NAT count of router ", router.Uuid,
		" is ", router.NATruleCount, " (GW ", router.GatewayUuid, ")"))

	if router.NATruleCount == 1 && router.GatewayUuid != nil {
		gw := gateways[INTERNET][UUID(*router.GatewayUuid)]
		err = detachRouterFromGW(db, ovnClient, router, gw)
		if err != nil {
			return &v1.UpdatePortResponse{PortId: &v1.PortId{}},
				fmt.Errorf("could not detach Router %s from GW %s",
					router.Uuid, string(*router.GatewayUuid))
		}
	}

	return &v1.UpdatePortResponse{
		PortId: &v1.PortId{},
	}, err
}

func attachRouterToGW(db *sql.DB, ovnClient libovsdbclient.Client,
	router *LogicalRouter, gw *Gateway) error {
	Logger.V(DebugLevel).Info(fmt.Sprintln("Attach router ", router.Uuid, " to GW ", gw.ChassisId))

	tx, err := db.Begin()
	if err != nil {
		Logger.Error(err, "cannot begin SQL transaction")
		return GrpcErrorFromSql(err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				Logger.Error(err, "failed to rollback transaction")
				return
			}
		}
	}()

	gw_name := "gw-" + strconv.Itoa(gw.ChassisId)
	lrpName := router.Name + "-" + gw_name
	ipAddr, err := gw.NATipManager.AllocateIP(tx)
	if err != nil {
		err = fmt.Errorf("could not allocate IP for GW interface of router %s to gw %s: %w ",
			router.Uuid, gw.Uuid, err)
		return err
	}

	snat_IPAddress, err := gw.NATipManager.AllocateIP(tx)
	if err != nil {
		err = fmt.Errorf("could not allocate IP for GW interface of router %s to gw %s: %w ",
			router.Uuid, gw.Uuid, err)
		return err
	}
	networks := []string{ipAddr + gw.AccessNetworkPrefix}
	MAC, err := generateMACFromUUID(router.Uuid)
	if err != nil {
		return fmt.Errorf("could not generate MAC for router uuid %q: %w ", router.Uuid, err)
	}
	gatewayChassis := nbdb.GatewayChassis{
		Name:        gw_name,
		ChassisName: gw_name,
		Priority:    1,
	}
	logicalRouterPort := nbdb.LogicalRouterPort{
		Name:     lrpName,
		Networks: networks,
		MAC:      MAC,
	}
	Logger.V(DebugLevel).Info(fmt.Sprintln("Generated MAC for router interface ", lrpName,
		"  to GW ", gw.ChassisId, " : ", MAC))
	logicalRouter := nbdb.LogicalRouter{UUID: string(router.Uuid)}
	err = libovsdbops.CreateOrUpdateLogicalRouterPort(ovnClient, &logicalRouter, &logicalRouterPort,
		&gatewayChassis, &logicalRouterPort.MAC, &logicalRouterPort.Networks)

	if err != nil {
		err = fmt.Errorf("could not create interface for router %s to gw %s: %w ",
			router.Uuid, gw.Uuid, err)
		return err
	}
	lspName := gw_name + "-" + router.Name
	logicalSwitchPort := nbdb.LogicalSwitchPort{
		Name:      lspName,
		Type:      "router",
		Addresses: []string{"router"},
		Options: map[string]string{
			"router-port": lrpName,
		},
	}
	logicalSwitch := nbdb.LogicalSwitch{
		UUID: string(gw.SubnetId),
	}
	err = libovsdbops.CreateOrUpdateLogicalSwitchPortsOnSwitch(ovnClient, &logicalSwitch, &logicalSwitchPort)
	if err != nil {
		err = fmt.Errorf("could not create GW subnet port for router %s on GW: %d: %w",
			router.Uuid, gw.ChassisId, err)
		return err
	}
	// Add Route
	ovnlogicalRouter, err := libovsdbops.GetLogicalRouter(ovnClient, &nbdb.LogicalRouter{UUID: string(router.Uuid)})
	if err != nil {
		err = fmt.Errorf("error finding router with UUID %s: %w ", router.Uuid, err)
		return err
	}
	staticRoute := nbdb.LogicalRouterStaticRoute{
		UUID:     uuid.NewString(),
		IPPrefix: gw.TargetNetwork,
		Nexthop:  gw.Nexthop,
	}
	predicate := func(item *nbdb.LogicalRouterStaticRoute) bool {
		return item.IPPrefix == gw.TargetNetwork
	}

	err = libovsdbops.CreateOrReplaceLogicalRouterStaticRouteWithPredicate(ovnClient,
		ovnlogicalRouter.GetName(), &staticRoute, predicate)
	if err != nil {
		err = fmt.Errorf("could not create static route: %w ", err)
		return err
	}

	_, err = tx.Exec(sqlQueryUpdateRouterGwFields, gw.Uuid, logicalRouterPort.UUID,
		logicalSwitchPort.UUID, ipAddr, staticRoute.UUID, snat_IPAddress, router.Uuid)
	if err != nil {
		Logger.Error(err, "Error updating router fields")
		return GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return GrpcErrorFromSql(err)
	}
	router.Snat_IPAddress = &snat_IPAddress
	router.GatewayUuid = &gw.Uuid
	Logger.V(DebugLevel).Info(fmt.Sprintf("Router %s attached to GW %s successfully", router.Uuid, gw.Uuid))
	return nil
}

func detachRouterFromGW(db *sql.DB, ovnClient libovsdbclient.Client,
	router *LogicalRouter, gw *Gateway) error {
	Logger.V(DebugLevel).Info(fmt.Sprintln("Detach router ", router.Uuid, " from GW ", gw.ChassisId))
	var gwInterfaceUuid, gwRouteUuid, gwSwitchPortId string
	var gwPortIpAddress, snatIpAddress string

	tx, err := db.Begin()
	if err != nil {
		Logger.Error(err, "cannot begin SQL transaction")
		return GrpcErrorFromSql(err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				Logger.Error(err, "failed to rollback transaction")
				return
			}
		}
	}()
	err = tx.QueryRow(sqlQueryFetchAndClearRouterGWFields, router.Uuid).Scan(

		&gwInterfaceUuid,
		&gwSwitchPortId,
		&gwPortIpAddress,
		&gwRouteUuid,
		&snatIpAddress,
	)
	if err != nil {
		Logger.Error(err, "Error updating router fields")
		return GrpcErrorFromSql(err)
	}

	Logger.V(DebugLevel).Info(fmt.Sprintln("Detach router ", router.Uuid, " from GW ", gw.ChassisId))
	logicalRouterPort := nbdb.LogicalRouterPort{
		UUID: gwInterfaceUuid,
	}
	logicalRouter := nbdb.LogicalRouter{UUID: string(router.Uuid)}
	err = libovsdbops.DeleteLogicalRouterPorts(ovnClient, &logicalRouter, &logicalRouterPort)
	if err != nil {
		err = fmt.Errorf("could not delete interface for router %s to gw %s: %w ",
			router.Uuid, *router.GatewayUuid, err)
		return err
	}
	logicalSwitchPort := nbdb.LogicalSwitchPort{
		UUID: gwSwitchPortId,
	}
	logicalSwitch := nbdb.LogicalSwitch{
		UUID: string(gw.SubnetId),
	}
	err = libovsdbops.DeleteLogicalSwitchPorts(ovnClient, &logicalSwitch, &logicalSwitchPort)
	if err != nil {
		err = fmt.Errorf("could not delete GW subnet port for router %s on GW: %d: %w",
			router.Uuid, gw.ChassisId, err)
		return err
	}

	err = gw.NATipManager.FreeIP(tx, gwPortIpAddress)
	if err != nil {
		Logger.Info(fmt.Sprintln("could not free IP from GW interface of router ", router.Uuid, " to GW ",
			gw.Uuid))
	}

	err = gw.NATipManager.FreeIP(tx, snatIpAddress)
	if err != nil {
		Logger.Info(fmt.Sprintln("could not free SNAT IP from GW interface of router ", router.Uuid, " to GW ",
			gw.Uuid))
	}
	// Delete Route
	ovnlogicalRouter, err := libovsdbops.GetLogicalRouter(ovnClient, &nbdb.LogicalRouter{UUID: string(router.Uuid)})
	if err != nil {
		err = fmt.Errorf("error finding router with UUID %s: %w ", router.Uuid, err)
		return err
	}
	staticRoute := nbdb.LogicalRouterStaticRoute{
		UUID: gwRouteUuid,
	}

	err = libovsdbops.DeleteLogicalRouterStaticRoutes(ovnClient,
		ovnlogicalRouter.GetName(), &staticRoute)
	if err != nil {
		err = fmt.Errorf("could not create static route: %w ", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return GrpcErrorFromSql(err)
	}
	router.GatewayUuid = nil
	return nil
}

func addNATRule(db *sql.DB, ovnClient libovsdbclient.Client, port *LogicalSwitchPort,
	router *LogicalRouter, NATflag bool) error {
	NATrule := nbdb.NAT{
		UUID:       uuid.NewString(),
		LogicalIP:  port.IpAddr,
		ExternalIP: *router.Snat_IPAddress,
		Type:       nbdb.NATTypeSNAT,
		Options: map[string]string{
			"stateless": "false",
		},
	}
	NATs := []*nbdb.NAT{
		&NATrule,
	}
	ovnRouter := &nbdb.LogicalRouter{UUID: string(router.Uuid)}
	err := libovsdbops.CreateOrUpdateNATs(ovnClient, ovnRouter, NATs...)
	if err != nil {
		err = fmt.Errorf("could not create NAT rule on router %s for port %s: %w ",
			router.Uuid, port.Uuid, err)
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		Logger.Error(err, "cannot begin SQL transaction")
		return GrpcErrorFromSql(err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				Logger.Error(err, "failed to rollback transaction")
				return
			}
		}
	}()
	_, err = tx.Exec(sqlQueryUpdatePortNATRuleId, NATrule.UUID, NATflag, port.Uuid)
	if err != nil {
		Logger.Error(err, "Error updating router fields")
		return GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return GrpcErrorFromSql(err)
	}
	return nil
}

func deleteNATRule(db *sql.DB, ovnClient libovsdbclient.Client, port *LogicalSwitchPort,
	router *LogicalRouter, NATflag bool) error {
	NATrule := nbdb.NAT{
		UUID: string(*port.SNATruleUUID),
	}
	NATs := []*nbdb.NAT{
		&NATrule,
	}
	ovnRouter := &nbdb.LogicalRouter{UUID: string(router.Uuid)}
	err := libovsdbops.DeleteNATs(ovnClient, ovnRouter, NATs...)
	if err != nil {
		Logger.Info(fmt.Sprintln("could not delete NAT rule from router ", router.Uuid, " for port ",
			port.Uuid, " :  ", err))
	}
	tx, err := db.Begin()
	if err != nil {
		Logger.Error(err, "cannot begin SQL transaction")
		return GrpcErrorFromSql(err)
	}
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				Logger.Error(err, "failed to rollback transaction")
				return
			}
		}
	}()

	err = tx.QueryRow(sqlQueryUpdatePortAndCountNAT, router.Uuid, nil, NATflag, port.Uuid).Scan(&router.NATruleCount)
	if err != nil {
		Logger.Error(err, "Error updating port SNAT rule ID and counting NAT ports")
		return GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return GrpcErrorFromSql(err)
	}
	return nil
}
