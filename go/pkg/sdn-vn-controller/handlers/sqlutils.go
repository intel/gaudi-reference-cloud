// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package handlers

import (
	"database/sql"
	"fmt"

	pq "github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var postgresToGrpcErrMap = map[pq.ErrorCode]codes.Code{
	"23505": codes.AlreadyExists,      // Unique Violation
	"23503": codes.FailedPrecondition, // Foreign Key Violation
	"23514": codes.InvalidArgument,    // Check Violation
	"22001": codes.InvalidArgument,    // Data too long for column
	"23502": codes.InvalidArgument,    // Not Null Violation
	"42601": codes.InvalidArgument,    // Syntax error
	"22P02": codes.InvalidArgument,    // invalid_text_representation (e.g. wrong UUID)
	"08001": codes.Unavailable,        // SQLClient unable to establish SQLConnection
	"08006": codes.Unavailable,        // Connection failure
}

var (
	sqlQueryGetVpc        = "SELECT vpc_name, tenant_id, region_id FROM vpc WHERE vpc_id = $1"
	sqlQueryGetVpcRouters = "SELECT router_id FROM router WHERE vpc_id = $1"
	sqlQueryGetVpcSubnets = "SELECT subnet_id FROM subnet WHERE vpc_id = $1"
	sqlQueryListVpc       = "SELECT vpc_id FROM vpc"
	sqlQueryCreateVpc     = "INSERT INTO vpc (vpc_id, vpc_name, tenant_id, region_id) VALUES ($1, $2, $3, $4)"
	sqlQueryDeleteVpc     = "DELETE FROM vpc WHERE vpc_id = $1"

	sqlQueryGetSubnet    = "SELECT subnet_name, subnet_cidr, availability_zone, vpc_id FROM subnet WHERE subnet_id = $1"
	sqlQueryListSubnet   = "SELECT subnet_id FROM subnet"
	sqlQueryCreateSubnet = "INSERT INTO subnet (subnet_id, subnet_name, subnet_cidr, availability_zone, vpc_id) VALUES ($1, $2, $3, $4, $5)"

	sqlQueryGetRouter    = "SELECT router_name, vpc_id FROM router WHERE router_id = $1"
	sqlQueryListRouter   = "SELECT router_id FROM router"
	sqlQueryCreateRouter = "INSERT INTO router (router_id, router_name, vpc_id) VALUES ($1, $2, $3)"
	sqlQueryDeleteRouter = "DELETE FROM router WHERE router_id = $1 RETURNING router_name"

	sqlQueryCreateRouterInterface = `
		INSERT INTO router_interface (
			router_interface_id, subnet_id, router_id, router_port_id, switch_port_id, router_interface_ip_address, router_mac_address
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	sqlQueryGetSubnetName = `
		SELECT subnet_name
		FROM subnet
		WHERE subnet_id = $1
	`

	sqlQueryGetRouterGWData = `
		SELECT router_name, gateway_id, snat_ip_address
		FROM router
		WHERE router_id = $1
	`

	sqlQueryDeleteRouterInterface = `
		DELETE FROM router_interface 
		WHERE router_interface_id = $1 
		RETURNING 
			router_id, 
			subnet_id, 
			router_port_id, 
			switch_port_id,
			(SELECT r.router_name FROM router r WHERE r.router_id = router_interface.router_id) AS router_name,
			(SELECT r.gateway_id FROM router r WHERE r.router_id = router_interface.router_id) AS gateway_id,
			(SELECT r.snat_ip_address FROM router r WHERE r.router_id = router_interface.router_id) AS snat_ip_address
	`

	sqlQueryListRouterInterface = `
		SELECT router_interface_id
		FROM router_interface
	`

	sqlQueryGetRouterInterface = `
		SELECT router_id, subnet_id, router_interface_ip_address, router_mac_address
		FROM router_interface
		WHERE router_interface_id = $1
	`

	sqlQueryGetStaticRoute    = "SELECT prefix, nexthop, router_id FROM static_route where static_route_id = $1"
	sqlQueryListStaticRoute   = "SELECT static_route_id FROM static_route"
	sqlQueryCreateStaticRoute = "INSERT INTO static_route (static_route_id, prefix, nexthop, router_id) VALUES ($1, $2, $3, $4)"
	sqlQueryDeleteStaticRoute = "DELETE FROM static_route WHERE static_route_id = $1 RETURNING router_id"

	sqlQueryGetNATPortsInSubnet = `
		SELECT 
			p.port_id,
			p.port_name,
			p.admin_state_up,
			p.internal_ip_address,
			p.snat_rule_id,
			p.external_ip_address
		FROM port p
		WHERE p.subnet_id = $1 AND p.isnat = TRUE
	`
	sqlQueryDeleteSubnetAndRefs = `
		WITH deleted_security_group_refs AS (
			DELETE FROM subnet_security_group 
			WHERE subnet_id = $1
		)
		DELETE FROM subnet
		WHERE subnet_id = $1
		RETURNING subnet_name
	`

	sqlQueryListPort = `
	    SELECT port_id
		FROM port
	`

	sqlQueryCreatePort = `
		INSERT INTO port (
			port_id,
			port_name,
			subnet_id,
			chassis_id,
			device_id,
			admin_state_up,
			mac_address,
			internal_ip_address,
			isNat,
			snat_rule_id,
			external_ip_address,
			external_mac_address,
			availability_zone,
			vpc_id
		) 
		SELECT 
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			s.availability_zone, s.vpc_id
		FROM subnet s
		WHERE s.subnet_id = $3
	`

	sqlQueryGetSwitchRouterData = `
		SELECT 
			r.router_id, 
			r.router_name, 
			r.gateway_id, 
			r.snat_ip_address
		FROM router_interface ri
		JOIN router r ON ri.router_id = r.router_id
		WHERE ri.subnet_id = $1
	`

	sqlQueryDeletePortAndRefs = `
		WITH deleted_security_group_refs AS (
			DELETE FROM port_security_group 
			WHERE port_id = $1
			RETURNING port_id
		),
		deleted_port AS (
			DELETE FROM port
			WHERE port_id = $1
			RETURNING port_name, subnet_id, snat_rule_id
		)
		SELECT port_name, subnet_id, snat_rule_id FROM deleted_port
	`

	sqlQueryGetPort = `
		SELECT
			port_name,
			subnet_id,
			internal_ip_address,
			mac_address,
			chassis_id,
			device_id,
			admin_state_up,
			isNAT,
			external_ip_address,
			external_mac_address
		FROM port
		WHERE port_id = $1
	`

	sqlQueryGetPortAndRouterData = `
        SELECT 
            p.port_name, 
            p.subnet_id, 
            p.admin_state_up, 
            p.internal_ip_address,
            p.snat_rule_id,
            p.external_ip_address,
            r.router_id, 
            r.router_name, 
            r.gateway_id, 
            r.snat_ip_address
        FROM port p
        LEFT JOIN router_interface ri ON p.subnet_id = ri.subnet_id
        LEFT JOIN router r ON ri.router_id = r.router_id
        WHERE p.port_id = $1
    `

	sqlQueryUpdateAdminState = `
		UPDATE port
		SET admin_state_up = $1
		WHERE port_id = $2
	`

	sqlQueryCreateSecurityRule = `
		WITH new_security_rule AS (
			INSERT INTO security_rule (
				security_rule_id, 
				security_rule_name, 
				security_rule_priority, 
				direction, 
				protocol, 
				source_ip_addresses, 
				source_port, 
				destination_ip_addresses, 
				destination_port, 
				security_rule_action, 
				src_address_set_uuid,
				dst_address_set_uuid,
				vpc_id,
				security_group_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $16)
			RETURNING security_rule_id
		)
		INSERT INTO acl (
			ovn_acl_id, 
			security_rule_id, 
			availability_zone
		) VALUES ($14, $1, $15)
	`

	sqlQueryUpdateSrcAddressSetUUID = `
		UPDATE security_rule
		SET src_address_set_uuid = $1
		WHERE security_rule_id = $2
	`

	sqlQueryUpdateDstAddressSetUUID = `
		UPDATE security_rule
		SET dst_address_set_uuid = $1
		WHERE security_rule_id = $2
	`

	sqlQueryGetSecurityRuleData = `
		SELECT a.ovn_acl_id, sg.security_group_id, pg.port_group_id, sg.security_group_name, sg.security_group_type, sr.src_address_set_uuid, sr.dst_address_set_uuid 
		FROM acl a
		INNER JOIN security_rule sr ON a.security_rule_id = sr.security_rule_id
		INNER JOIN security_group sg ON sg.security_group_id = sr.security_group_id
		LEFT JOIN port_group pg ON pg.security_group_id = sg.security_group_id
		WHERE sr.security_rule_id = $1
	`

	sqlQueryDeleteSecurityRuleAndRefs = `
		WITH deleted_acls AS (
			DELETE FROM acl 
			WHERE security_rule_id = $1
			RETURNING security_rule_id
		)
		DELETE FROM security_rule
		WHERE security_rule_id = $1
	`

	sqlQueryGetSecurityRule = `
		SELECT 
			sr.security_rule_id,
			sr.security_rule_name,
			sr.security_rule_priority,
			sr.direction,
			sr.protocol,
			sr.source_ip_addresses,
			sr.destination_ip_addresses,
			sr.source_port,
			sr.destination_port,
			sr.security_rule_action,
			sr.vpc_id,
			sr.src_address_set_uuid,
			sr.dst_address_set_uuid,
			acl.ovn_acl_id as acl_uuid,
			sr.security_group_id
		FROM security_rule sr
		LEFT JOIN acl ON sr.security_rule_id = acl.security_rule_id
		WHERE sr.security_rule_id = $1;
	`

	sqlQueryUpdateSecurityRule = `
		UPDATE security_rule
		SET
			security_rule_priority = $1,
			direction = $2,
			protocol = $3,
			source_ip_addresses = $4,
			destination_ip_addresses = $5,
			source_port = $6,
			destination_port = $7,
			security_rule_action = $8,
			src_address_set_uuid = $9,
			dst_address_set_uuid = $10
		WHERE
			security_rule_id = $11;
	`

	sqlQueryListSecurityRules = `
		SELECT 
			security_rule_id
		FROM security_rule
	`

	sqlQueryCheckSecurityRuleExists = `
		SELECT EXISTS(SELECT 1 FROM security_rule WHERE security_rule_id = $1)
	`

	sqlQueryCreateSecurityGroup = `
		INSERT INTO security_group (
			security_group_id, 
			security_group_name, 
			vpc_id,
			security_group_type
		) VALUES ($1, $2, $3, $4)
	`

	sqlQueryCreatePortGroup = `
		INSERT INTO port_group (
			port_group_id,
			security_group_id,
			availability_zone
		) VALUES ($1, $2, $3)
	`

	sqlQueryGetSecurityGroup = `
		SELECT sg.security_group_name, sg.vpc_id, sg.security_group_type, pg.port_group_id
		FROM security_group sg
		LEFT JOIN port_group pg ON sg.security_group_id = pg.security_group_id
		WHERE sg.security_group_id = $1
	`

	sqlQueryDeleteSecurityGroupDependencies = `
		WITH deleted_port_security_groups AS (
			DELETE FROM port_security_group WHERE security_group_id = $1
		),
		deleted_subnet_security_groups AS (
			DELETE FROM subnet_security_group WHERE security_group_id = $1
		),
		deleted_port_groups AS (
			DELETE FROM port_group WHERE security_group_id = $1
		)
		DELETE FROM security_group WHERE security_group_id = $1;
	`

	sqlQueryGetSecurityGroupSecurityRuleUUIDs = `
    	SELECT security_rule_id FROM security_rule WHERE security_group_id = $1
	`

	sqlQueryGetSecurityGroupSubnetUUIDs = `
		SELECT subnet_id FROM subnet_security_group WHERE security_group_id = $1
	`

	sqlQueryGetSecurityGroupPortUUIDs = `
		SELECT port_id FROM port_security_group WHERE security_group_id = $1
	`

	sqlQueryDetachOldSubnets = `
		DELETE FROM subnet_security_group
		WHERE subnet_id = ANY($1) AND security_group_id = $2
	`

	sqlQueryDetachOldPorts = `
		DELETE FROM port_security_group
		WHERE port_id = ANY($1) AND security_group_id = $2
	`

	sqlQueryInsertNewSubnets = `
		INSERT INTO subnet_security_group (subnet_id, security_group_id)
		VALUES (unnest($1::uuid[]), $2)
	`

	sqlQueryInsertNewPorts = `
		INSERT INTO port_security_group (port_id, security_group_id)
		VALUES (unnest($1::uuid[]), $2)
	`

	sqlQueryGetSecurityRulesACLs = `
		SELECT a.ovn_acl_id
		FROM acl a
		INNER JOIN security_rule sr ON a.security_rule_id = sr.security_rule_id
		WHERE sr.security_rule_id = ANY($1)
	`
	sqlQueryUpdateRouterGwFields = `
		UPDATE router
		SET 
			gateway_id = $1,
			gw_interface_uuid = $2,
			gw_switch_port_id = $3,
			gw_port_ip_address = $4,
			gw_route_uuid = $5,
			snat_ip_address = $6
		WHERE 
			router_id = $7
	`

	sqlQueryGetSecurityGroupInfo = `
		SELECT security_group_name, vpc_id, security_group_type
		FROM security_group
		WHERE security_group_id = $1
	`

	sqlQueryGetSecurityGroupRules = `
		SELECT security_rule_id
		FROM security_rule
		WHERE security_group_id = $1
	`

	sqlQueryGetSecurityGroupPorts = `
		SELECT port_id
		FROM port_security_group
		WHERE security_group_id = $1
	`

	sqlQueryGetSecurityGroupSubnets = `
		SELECT subnet_id
		FROM subnet_security_group
		WHERE security_group_id = $1
	`

	sqlQueryListSecurityGroups = `
		SELECT security_group_id
		FROM security_group
	`
	sqlQueryUpdatePortNATRuleId = `
		UPDATE port
		SET snat_rule_id = $1,
			isNAT = $2
		WHERE 
			port_id = $3
	`
	/* set the values for isNAT and snat_rule_id calculate the number of ports whithn all the subnets connected to the router */
	/* That is needed to detach the router from the gateway when no more ports have NAT enabled */

	sqlQueryUpdatePortAndCountNAT = `
		WITH subnet_router AS (
			SELECT ri.subnet_id
			FROM router_interface ri
			WHERE ri.router_id = $1
		),
		nat_port_count AS (
			SELECT COUNT(*)
			FROM port p
			WHERE p.snat_rule_id IS NOT NULL AND p.subnet_id IN (
				SELECT subnet_id FROM subnet_router
			)
		),
		updated_port AS (
			UPDATE port
			SET 
				snat_rule_id = $2,
				isNat = $3
			WHERE 
				port_id = $4
			RETURNING subnet_id
		)
		SELECT * FROM nat_port_count;
	`

	sqlQueryFetchAndClearRouterGWFields = `
		WITH fetched AS (
			SELECT 
				gw_interface_uuid,
				gw_switch_port_id,
				gw_port_ip_address,
				gw_route_uuid,
				snat_ip_address
			FROM 
				router
			WHERE 
				router_id = $1
			FOR UPDATE
		),
		updated AS (
			UPDATE router
			SET
				gateway_id = NULL,
				gw_interface_uuid = NULL,
				gw_switch_port_id = NULL,
				gw_port_ip_address = NULL,
				gw_route_uuid = NULL,
				snat_ip_address = NULL
			WHERE 
				router_id = $1
			RETURNING router_id
		)
		SELECT 
			f.gw_interface_uuid,
			f.gw_switch_port_id,
			f.gw_port_ip_address,
			f.gw_route_uuid,
			f.snat_ip_address
		FROM fetched f
		JOIN updated u ON u.router_id = $1	
	`
	sqlQueryCreateGateway = `
		INSERT INTO gateway (
			gateway_id,
			chassis_id,
			gateway_name,
			subnet_id,
			ext_port_name,
			ext_network_name,
			access_network,
			maxNATs
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	// VPC thread locks requiring fetching vpc_id given other types of id
	sqlQueryCheckVpcExists = `
		SELECT EXISTS(SELECT 1 FROM vpc WHERE vpc_id = $1)
	`
	sqlQueryVpcIdFromRouter = `
		SELECT vpc_id
		FROM router
		WHERE router_id = $1
	`
	sqlQueryVpcIdFromSubnet = `
		SELECT vpc_id
		FROM subnet
		WHERE subnet_id = $1
	`
	sqlQueryVpcIdFromPort = `
		SELECT vpc_id
		FROM port
		WHERE port_id = $1
	`
	sqlQueryVpcIdFromSecurityRule = `
		SELECT vpc_id
		FROM security_rule
		WHERE security_rule_id = $1
	`
	sqlQueryVpcIdFromSecurityGroup = `
		SELECT vpc_id
		FROM security_group
		WHERE security_group_id = $1
	`
	sqlQueryVpcIdFromRouterInterface = `
		SELECT s.vpc_id
		FROM router_interface ri
		JOIN subnet s ON s.subnet_id = ri.subnet_id
		WHERE ri.router_interface_id = $1
	`
	sqlQueryVpcIdFromStaticRoute = `
		SELECT r.vpc_id
		FROM static_route sr
		JOIN router r ON r.router_id = sr.router_id
		WHERE sr.static_route_id = $1
	`
)

// Convert an error from PostgreSQL to gRPC message
func GrpcErrorFromSql(err error) error {
	// Translate SQL error codes
	if err == sql.ErrNoRows {
		return status.Errorf(codes.NotFound, err.Error())
	}
	// Translate Postgres error codes
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return status.Errorf(codes.Unknown, err.Error())
	}
	grpcCode, ok := postgresToGrpcErrMap[pqErr.Code]
	if !ok {
		return status.Errorf(codes.Unknown, pqErr.Detail+" PostgreSQL Error code: "+string(pqErr.Code))
	}
	return status.Errorf(grpcCode, pqErr.Detail)
}

func toInt4Range(pr *PortRange) interface{} {
	if pr == nil {
		return nil
	}
	return fmt.Sprintf("[%d,%d]", pr.Min, pr.Max)
}
