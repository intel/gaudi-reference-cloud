// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package handlers

import (
	"database/sql"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	pq "github.com/lib/pq"
	libovsdbclient "github.com/ovn-org/libovsdb/client"
	libovsdbops "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/libovsdb/ops"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Protocol uint32

const (
	Protocol_TCP Protocol = iota
	Protocol_UDP
	Protocol_ICMP
)

type Direction uint32

const (
	Direction_UNSPECIFIED Direction = iota
	Direction_INGRESS
	Direction_EGRESS
)

type SecurityAction uint32

const (
	SecurityAction_UNSPECIFIED SecurityAction = iota
	SecurityAction_ALLOW
	SecurityAction_DENY
)

type PortRange struct {
	Min uint32
	Max uint32
}

type SecurityOutputType uint32

const (
	SecurityOutputType_PORT SecurityOutputType = 0
	SecurityOutputType_SG   SecurityOutputType = 1
)

type SecurityOutput struct {
	Type SecurityOutputType
	Uuid UUID
}

type SecurityRule struct {
	Uuid                 UUID
	Name                 string
	Priority             uint32
	Direction            Direction
	SourceIPs            []string
	DestinationIPs       []string
	SrcAddressSetUuid    *UUID
	DstAddressSetUuid    *UUID
	Protocol             *Protocol
	SourcePortRange      *PortRange
	DestinationPortRange *PortRange
	Action               SecurityAction
	VpcId                UUID
	securityGroupId      UUID
	sr_type              int32
	portGroupUUID        *UUID
	portGroupName        *string
	subnets              []string
}

func CreateSecurityRuleHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.CreateSecurityRuleRequest) (*v1.CreateSecurityRuleResponse, error) {
	Logger := Logger.WithName("CreateSecurityRuleHandler")

	// Validate vpc_id and fetch the lock
	vpcId := r.GetVpcId().GetUuid()
	var vpcExists bool
	err := db.QueryRow(sqlQueryCheckVpcExists, vpcId).Scan(&vpcExists)
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

	Logger.V(DebugLevel).Info(fmt.Sprintf("create Security rule %s ", r.Name))
	var destinationPortRange *PortRange
	if r.DestinationPortRange != nil {
		destinationPortRange = &PortRange{
			Min: r.DestinationPortRange.Min,
			Max: r.DestinationPortRange.Max,
		}
	}

	var sourcePortRange *PortRange
	if r.SourcePortRange != nil {
		sourcePortRange = &PortRange{
			Min: r.SourcePortRange.Min,
			Max: r.SourcePortRange.Max,
		}
	}

	var protocol Protocol
	if r.Protocol != nil {
		protocol = Protocol(*r.Protocol)
	}

	securityRule := SecurityRule{
		Uuid:                 UUID(r.GetSecurityRuleId().Uuid),
		Name:                 r.Name,
		Priority:             r.Priority,
		Direction:            Direction(r.Direction),
		SourceIPs:            r.Source_IPAddresses,
		DestinationIPs:       r.Destination_IPAddresses,
		Protocol:             &protocol,
		SourcePortRange:      sourcePortRange,
		DestinationPortRange: destinationPortRange,
		Action:               SecurityAction(r.Action),
		VpcId:                UUID(r.GetVpcId().GetUuid()),
		securityGroupId:      UUID(r.SecurityGroupId.Uuid),
	}

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

	var securityGroupName, vpc_id string

	err = tx.QueryRow(sqlQueryGetSecurityGroup, securityRule.securityGroupId).Scan(&securityGroupName, &vpc_id,
		&securityRule.sr_type, &securityRule.portGroupUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			errMsg := fmt.Sprintf("No group found with UUID: %s", securityRule.securityGroupId)
			Logger.Error(err, errMsg)
			return nil, status.Errorf(codes.NotFound, errMsg)
		}
		Logger.Error(err, "Failed to fetch security group")
		return nil, GrpcErrorFromSql(err)
	}

	if securityRule.sr_type == int32(v1.SecurityGroupType_PORT) {
		securityRule.portGroupName = &securityGroupName
	} else {
		securityRule.subnets, err = getGroupDeps(sqlQueryGetSecurityGroupSubnetUUIDs, tx, string(securityRule.securityGroupId))
		if err != nil {
			return nil, err
		}
	}
	aclUUID := uuid.NewString()
	_, err = tx.Exec(sqlQueryCreateSecurityRule,
		securityRule.Uuid,
		securityRule.Name,
		securityRule.Priority,
		securityRule.Direction,
		securityRule.Protocol,
		pq.Array(securityRule.SourceIPs),
		toInt4Range(securityRule.SourcePortRange),
		pq.Array(securityRule.DestinationIPs),
		toInt4Range(securityRule.DestinationPortRange),
		securityRule.Action,
		nil,
		nil,
		securityRule.VpcId,
		aclUUID,
		"AZ1",
		securityRule.securityGroupId,
	)

	if err != nil {
		Logger.Error(err, "SQL insert error")
		return nil, GrpcErrorFromSql(err)
	}

	var direction string
	var action string
	if securityRule.Direction == Direction_INGRESS {
		direction = nbdb.ACLDirectionFromLport
	} else {
		direction = nbdb.ACLDirectionToLport
	}
	if securityRule.Action == SecurityAction_ALLOW {
		action = nbdb.ACLActionAllowRelated
	} else {
		action = nbdb.ACLActionDrop
	}

	if len(securityRule.SourceIPs) > 1 {
		addrSet := nbdb.AddressSet{
			UUID:      uuid.NewString(),
			Name:      HashForOVN("as_src_" + securityRule.Name),
			Addresses: securityRule.SourceIPs,
		}
		//Create address set
		addrSet.Addresses = securityRule.SourceIPs
		err := libovsdbops.CreateAddressSets(ovnClient, &addrSet)
		if err != nil {
			err = fmt.Errorf("could not create address set for rule %s: %w ",
				securityRule.Name, err)
			return nil, err
		}
		Logger.V(DebugLevel).Info(fmt.Sprintf("created address set with UUID: %s", addrSet.UUID))
		securityRule.SrcAddressSetUuid = (*UUID)(&addrSet.UUID)
		_, err = tx.Exec(sqlQueryUpdateSrcAddressSetUUID,
			securityRule.SrcAddressSetUuid, securityRule.Uuid)
		if err != nil {
			Logger.Error(err, "SQL update error when updating src_address_set_uuid")
			return nil, GrpcErrorFromSql(err)
		}
	}

	if len(securityRule.DestinationIPs) > 1 {
		addrSet := nbdb.AddressSet{
			UUID:      uuid.NewString(),
			Name:      HashForOVN("as_dst_" + securityRule.Name),
			Addresses: securityRule.DestinationIPs,
		}
		//Create address set
		addrSet.Addresses = securityRule.DestinationIPs
		err := libovsdbops.CreateAddressSets(ovnClient, &addrSet)
		if err != nil {
			err = fmt.Errorf("could not create address set for rule %s: %w ",
				securityRule.Name, err)
			return nil, err
		}
		Logger.V(DebugLevel).Info(fmt.Sprintf("created address set with UUID: %s", addrSet.UUID))
		securityRule.DstAddressSetUuid = (*UUID)(&addrSet.UUID)
		_, err = tx.Exec(sqlQueryUpdateDstAddressSetUUID,
			securityRule.DstAddressSetUuid, securityRule.Uuid)
		if err != nil {
			Logger.Error(err, "SQL update error when updating dst_address_set_uuid")
			return nil, GrpcErrorFromSql(err)
		}
	}

	//Create Match
	match := createMatchString(&securityRule)
	objIDs := libovsdbops.NewDbObjectIDs(libovsdbops.PortGroupCluster, "sdn-controller",
		map[libovsdbops.ExternalIDKey]string{
			libovsdbops.ObjectNameKey: securityRule.Name,
		})
	var options map[string]string
	acl := libovsdbops.BuildACL(
		securityRule.Name,
		direction,
		int(securityRule.Priority),
		match,
		action,
		"",
		"",
		false,
		objIDs.GetExternalIDs(),
		options,
		types.DefaultACLTier,
	)

	acl.UUID = aclUUID
	Logger.V(DebugLevel).Info(fmt.Sprintf("create ACL with match %s and UUID %s for rule %s ", match, acl.UUID,
		securityRule.Uuid))
	ops, err := libovsdbops.CreateOrUpdateACLsOps(ovnClient, nil, acl)
	if err != nil {
		err = fmt.Errorf("could not create ACL for rule %s :%w",
			securityRule.Uuid, err)
		return nil, err
	}

	Logger.V(DebugLevel).Info(fmt.Sprintf("Attach security rule %s to Group %s", aclUUID, securityGroupName))

	if securityRule.sr_type == int32(v1.SecurityGroupType_PORT) {
		ops, err = libovsdbops.AddACLsToPortGroupOps(ovnClient, ops, HashForOVN(*securityRule.portGroupName), acl)
		if err != nil {
			err = fmt.Errorf("could not create ACL for rule %s :%w",
				securityRule.Uuid, err)
			return nil, err
		}

	} else {
		// Add ACL to the associated subnets
		for _, subnetUUID := range securityRule.subnets {
			swselect := nbdb.LogicalSwitch{
				UUID: string(subnetUUID),
			}
			sw, err := libovsdbops.GetLogicalSwitch(ovnClient, &swselect)
			if err != nil {
				err = fmt.Errorf("could not fetch subnet %s :%w", subnetUUID, err)
				return nil, err
			}
			ops, err = libovsdbops.AddACLsToLogicalSwitchOps(ovnClient, ops, sw.Name, acl)
			if err != nil {
				err = fmt.Errorf("could not assign ACLs to subnet %s of Group %s :%w", subnetUUID,
					securityRule.securityGroupId, err)
				return nil, err
			}
		}
	}
	_, err = libovsdbops.TransactAndCheck(ovnClient, ops)
	if err != nil {
		err = fmt.Errorf("could not create ACL to rule %s :%w", securityRule.Uuid, err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	Logger.V(DebugLevel).Info(fmt.Sprintf("created security rule UUID: %s with ACL %s",
		securityRule.Uuid, acl.UUID))

	return &v1.CreateSecurityRuleResponse{
		SecurityRuleId: &v1.SecurityRuleId{Uuid: string(securityRule.Uuid)}}, nil
}

func DeleteSecurityRuleHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.DeleteSecurityRuleRequest) (*v1.DeleteSecurityRuleResponse, error) {
	Logger := Logger.WithName("DeleteSecurityRuleHandler")
	securityRuleUuid := r.GetSecurityRuleId().GetUuid()

	var vpcId string
	err := db.QueryRow(sqlQueryVpcIdFromSecurityRule, securityRuleUuid).Scan(&vpcId)
	if err != nil {
		Logger.Error(err, "cannot find VPC")
		return nil, GrpcErrorFromSql(err)
	}
	mu := GetVpcMutex(vpcId)
	LockVpcMutex(mu, vpcId)
	defer UnlockVpcMutex(mu, vpcId)

	Logger.V(DebugLevel).Info(fmt.Sprintf("delete Security rule %s", securityRuleUuid))

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

	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM security_rule WHERE security_rule_id = $1)", securityRuleUuid).Scan(&exists)
	if !exists {
		errMsg := fmt.Sprintf("No security rule found with UUID: %s", securityRuleUuid)
		Logger.Error(err, errMsg)
		return nil, status.Errorf(codes.NotFound, errMsg)
	}

	var aclUUID, securityGroupUuid UUID
	var portGroupUuid *UUID
	var srcAddressSetUuid, dstAddressSetUuid *string
	var securityGroupName string
	var sr_type int32

	err = tx.QueryRow(sqlQueryGetSecurityRuleData, securityRuleUuid).Scan(&aclUUID, &securityGroupUuid, &portGroupUuid,
		&securityGroupName, &sr_type, &srcAddressSetUuid, &dstAddressSetUuid)
	if err != nil {
		Logger.Error(err, "SQL query error when fetching associated security group")
		return nil, GrpcErrorFromSql(err)
	}

	if sr_type == int32(v1.SecurityGroupType_PORT) {
		pgselect := nbdb.PortGroup{UUID: string(*portGroupUuid)}
		pg, err := libovsdbops.GetPortGroup(ovnClient, &pgselect)
		if err != nil {
			return nil, fmt.Errorf("could not find port Group %s: %w", string(*portGroupUuid), err)
		}
		var acls []string
		for _, acl := range pg.ACLs {
			if acl != string(aclUUID) {
				acls = append(acls, acl)
			}
		}
		pg.ACLs = acls
		err = libovsdbops.CreateOrUpdatePortGroups(ovnClient, pg)
		if err != nil {
			return nil, fmt.Errorf("could not remove Port ACLs from Group %s: %w", securityGroupUuid, err)
		}

	} else {
		subnetIDs, err := getGroupDeps(sqlQueryGetSecurityGroupSubnetUUIDs, tx, string(securityGroupUuid))
		if err != nil {
			return nil, err
		}
		// Remove ACL from the associated subnets
		for _, subnetUUID := range subnetIDs {
			p := func(item *nbdb.LogicalSwitch) bool { return item.UUID == string(subnetUUID) }
			acl := nbdb.ACL{UUID: string(aclUUID)}
			err := libovsdbops.RemoveACLsFromLogicalSwitchesWithPredicate(ovnClient, p, &acl)
			if err != nil {
				return nil, fmt.Errorf("could not delete ACL for subnet %s, rule %s of Group %s: %w",
					subnetUUID, aclUUID, securityGroupUuid, err)
			}
		}
	}
	// If the security rule has an associated address set, delete it
	if srcAddressSetUuid != nil {
		addrSet := nbdb.AddressSet{
			UUID: *srcAddressSetUuid,
		}
		err := libovsdbops.DeleteAddressSets(ovnClient, &addrSet)
		if err != nil {
			err = fmt.Errorf("could not delete src address set for rule %s: %w",
				securityRuleUuid, err)
			return nil, err
		}
		Logger.V(DebugLevel).Info(fmt.Sprintf("deleted address set with UUID: %s", addrSet.UUID))
	}

	if dstAddressSetUuid != nil {
		addrSet := nbdb.AddressSet{
			UUID: *dstAddressSetUuid,
		}
		err := libovsdbops.DeleteAddressSets(ovnClient, &addrSet)
		if err != nil {
			err = fmt.Errorf("could not delete dst address set for rule %s: %w",
				securityRuleUuid, err)
			return nil, err
		}
		Logger.V(DebugLevel).Info(fmt.Sprintf("deleted address set with UUID: %s", addrSet.UUID))
	}

	// Delete the security rule refs
	_, err = tx.Exec(sqlQueryDeleteSecurityRuleAndRefs, securityRuleUuid)
	if err != nil {
		Logger.Error(err, "SQL delete error when deleting security rule and its associations")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	Logger.V(DebugLevel).Info(fmt.Sprintf("deleted security rule UUID: %s", r.SecurityRuleId.Uuid))

	return &v1.DeleteSecurityRuleResponse{
		SecurityRuleId: &v1.SecurityRuleId{Uuid: r.SecurityRuleId.Uuid},
	}, nil
}

func protocolToString(protocol Protocol) string {
	switch protocol {
	case Protocol_TCP:
		return "tcp"
	case Protocol_UDP:
		return "udp"
	case Protocol_ICMP:
		return "icmp"
	default:
		return ""
	}
}

func createMatchString(rule *SecurityRule) string {
	var matchConditions []string
	if rule.sr_type == int32(v1.SecurityGroupType_PORT) {
		if rule.Direction == Direction_INGRESS {
			matchConditions = append(matchConditions, fmt.Sprintf("inport == @%s", HashForOVN(*rule.portGroupName)))
		} else if rule.Direction == Direction_EGRESS {
			matchConditions = append(matchConditions, fmt.Sprintf("outport == @%s", HashForOVN(*rule.portGroupName)))
		}
	}

	if rule.SrcAddressSetUuid != nil {
		addesseSetName := HashForOVN("as_src_" + rule.Name)
		matchConditions = append(matchConditions, fmt.Sprintf("ip4.src == $%s", addesseSetName))
	} else if len(rule.SourceIPs) == 1 {
		matchConditions = append(matchConditions, fmt.Sprintf("ip4.src == %s", rule.SourceIPs[0]))
	}
	if rule.DstAddressSetUuid != nil {
		addesseSetName := HashForOVN("as_dst_" + rule.Name)
		matchConditions = append(matchConditions, fmt.Sprintf("ip4.dst == $%s", addesseSetName))
	} else if len(rule.DestinationIPs) == 1 {
		matchConditions = append(matchConditions, fmt.Sprintf("ip4.dst == %s", rule.DestinationIPs[0]))
	}
	/* If not protocol is defined and a port range is present use tcp by default */
	protocolStr := "tcp"
	if rule.Protocol != nil {
		protocolStr = protocolToString(*rule.Protocol)
		matchConditions = append(matchConditions, protocolStr)
	}

	// TODO: Add support for udp ports in the match string.
	if rule.SourcePortRange != nil {
		if rule.SourcePortRange.Min == rule.SourcePortRange.Max {
			matchConditions = append(matchConditions,
				fmt.Sprintf(protocolStr+".src == %d", rule.SourcePortRange.Min))
		} else {
			matchConditions = append(matchConditions,
				fmt.Sprintf(protocolStr+".src >= %d && "+protocolStr+".src <= %d", rule.SourcePortRange.Min, rule.SourcePortRange.Max))
		}
	}
	if rule.DestinationPortRange != nil {
		if rule.DestinationPortRange.Min == rule.DestinationPortRange.Max {
			matchConditions = append(matchConditions,
				fmt.Sprintf(protocolStr+".dst == %d", rule.DestinationPortRange.Min))
		} else {
			matchConditions = append(matchConditions,
				fmt.Sprintf(protocolStr+".dst >= %d && "+protocolStr+".dst <= %d", rule.DestinationPortRange.Min, rule.DestinationPortRange.Max))
		}
	}
	matchString := strings.Join(matchConditions, " && ")
	return matchString
}

func GetSecurityRuleHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.GetSecurityRuleRequest) (*v1.GetSecurityRuleResponse, error) {
	var err error = nil
	uuid := r.GetSecurityRuleId().GetUuid()
	Logger := Logger.WithName("GetSecurityRuleHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("get vpc %s", uuid))

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

	securityRule := &v1.SecurityRule{
		Id:              &v1.SecurityRuleId{Uuid: uuid},
		VpcId:           &v1.VPCId{},
		SecurityGroupId: &v1.SecurityGroupId{},
	}
	var sourcePortRange, destinationPortRange sql.NullString // To hold raw range from DB, handles NULL values
	var src_address_set_uuid, dst_address_set_uuid *string
	var acl_uuid string

	err = db.QueryRow(sqlQueryGetSecurityRule, uuid).Scan(
		&securityRule.Id.Uuid,
		&securityRule.Name,
		&securityRule.Priority,
		&securityRule.Direction,
		&securityRule.Protocol,
		pq.Array(&securityRule.Source_IPs),
		pq.Array(&securityRule.Destination_IPs),
		&sourcePortRange,      // Scan the raw source port range
		&destinationPortRange, // Scan the raw destination port range
		&securityRule.Action,
		&securityRule.VpcId.Uuid,
		&src_address_set_uuid,
		&dst_address_set_uuid,
		&acl_uuid,
		&securityRule.SecurityGroupId.Uuid,
	)

	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}

	// Convert the raw port ranges into PortRange structs, handling NULL values
	if sourcePortRange.Valid {
		securityRule.SourcePortRange, err = parsePortRange(sourcePortRange.String)
		if err != nil {
			Logger.Error(err, "Error parsing source port range")
			return nil, fmt.Errorf("error parsing source port range: %v", err)
		}
	} else {
		securityRule.SourcePortRange = nil
	}

	if destinationPortRange.Valid {
		securityRule.DestinationPortRange, err = parsePortRange(destinationPortRange.String)
		if err != nil {
			Logger.Error(err, "Error parsing destination port range")
			return nil, fmt.Errorf("error parsing destination port range: %v", err)
		}
	} else {
		securityRule.DestinationPortRange = nil
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}
	return &v1.GetSecurityRuleResponse{
		SecurityRule: securityRule}, nil
}

func parsePortRange(rangeStr string) (*v1.PortRange, error) {
	// Handle empty or NULL range string
	if rangeStr == "" {
		return nil, nil
	}

	// Check if the upper bound is exclusive (ending with ')')
	exclusiveUpper := strings.HasSuffix(rangeStr, ")")

	// Trim the starting "[" or "(" and the ending "]" or ")"
	rangeStr = strings.TrimLeft(rangeStr, "[(")
	rangeStr = strings.TrimRight(rangeStr, "])")

	// Split the string into start and end values
	parts := strings.Split(rangeStr, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format: %s", rangeStr)
	}

	// Parse the start of the range
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid start of range: %v", err)
	}

	// Parse the end of the range
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid end of range: %v", err)
	}

	// If the upper bound is exclusive, we add 1 to convert it to an inclusive range
	if exclusiveUpper {
		end -= 1
	}

	return &v1.PortRange{
		Min: uint32(start),
		Max: uint32(end),
	}, nil
}

/* A hash function for the names of address_sets to be accepted by OVN. Special characters are not accepted */
func HashForOVN(s string) string {
	h := fnv.New64a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return ""
	}
	hashString := strconv.FormatUint(h.Sum64(), 10)
	return fmt.Sprintf("a%s", hashString)
}

func ListSecurityRulesHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.ListSecurityRulesRequest) (*v1.ListSecurityRulesResponse, error) {
	var err error = nil
	Logger := Logger.WithName("ListSecurityRulesHandler")
	Logger.V(DebugLevel).Info("listing all security rules")

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
	rows, err := db.Query(sqlQueryListSecurityRules)
	if err != nil {
		Logger.Error(err, "SQL query error")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	var securityRuleUUIDs []*v1.SecurityRuleId

	for rows.Next() {
		securityRuleId := &v1.SecurityRuleId{}
		err := rows.Scan(&securityRuleId.Uuid)
		if err != nil {
			Logger.Error(err, "SQL scan error")
			return nil, GrpcErrorFromSql(err)
		}

		securityRuleUUIDs = append(securityRuleUUIDs, securityRuleId)
	}

	if err := rows.Err(); err != nil {
		Logger.Error(err, "SQL rows iteration error")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	return &v1.ListSecurityRulesResponse{
		SecurityRuleIds: securityRuleUUIDs,
	}, nil
}

func UpdateSecurityRuleHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.UpdateSecurityRuleRequest) (*v1.UpdateSecurityRuleResponse, error) {
	Logger := Logger.WithName("UpdateSecurityRuleHandler")
	securityRuleUuid := r.GetSecurityRuleId().GetUuid()
	Logger.V(DebugLevel).Info(fmt.Sprintf("update Security rule %s", securityRuleUuid))

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

	// Check if security rule exists
	var exists bool
	err = tx.QueryRow(sqlQueryCheckSecurityRuleExists, securityRuleUuid).Scan(&exists)
	if err != nil {
		Logger.Error(err, "SQL query error when checking security rule existence")
		return nil, GrpcErrorFromSql(err)
	}
	if !exists {
		errMsg := fmt.Sprintf("No security rule found with UUID: %s", securityRuleUuid)
		Logger.Error(err, errMsg)
		return nil, status.Errorf(codes.NotFound, errMsg)
	}

	// Fetch existing security rule from the database
	var securityRule SecurityRule
	var srcPortRange sql.NullString
	var dstPortRange sql.NullString
	var protocol sql.NullInt64
	var srcAddressSetUuid sql.NullString
	var dstAddressSetUuid sql.NullString
	var aclUUID string

	err = tx.QueryRow(sqlQueryGetSecurityRule, securityRuleUuid).Scan(
		&securityRule.Uuid,
		&securityRule.Name,
		&securityRule.Priority,
		&securityRule.Direction,
		&protocol,
		pq.Array(&securityRule.SourceIPs),
		pq.Array(&securityRule.DestinationIPs),
		&srcPortRange,
		&dstPortRange,
		&securityRule.Action,
		&securityRule.VpcId,
		&srcAddressSetUuid,
		&dstAddressSetUuid,
		&aclUUID,
		&securityRule.securityGroupId,
	)
	if err != nil {
		Logger.Error(err, "SQL query error when fetching security rule")
		return nil, GrpcErrorFromSql(err)
	}

	var securityGroupName, vpc_id string

	err = tx.QueryRow(sqlQueryGetSecurityGroup, securityRule.securityGroupId).Scan(&securityGroupName, &vpc_id,
		&securityRule.sr_type, &securityRule.portGroupUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			errMsg := fmt.Sprintf("No group found with UUID: %s", securityRule.securityGroupId)
			Logger.Error(err, errMsg)
			return nil, status.Errorf(codes.NotFound, errMsg)
		}
		Logger.Error(err, "Failed to fetch security group")
		return nil, GrpcErrorFromSql(err)
	}

	if securityRule.sr_type == int32(v1.SecurityGroupType_PORT) {
		securityRule.portGroupName = &securityGroupName
	}

	var destinationPortRange *PortRange
	if r.DestinationPortRange != nil {
		destinationPortRange = &PortRange{
			Min: r.DestinationPortRange.Min,
			Max: r.DestinationPortRange.Max,
		}
	}
	securityRule.DestinationPortRange = destinationPortRange

	var sourcePortRange *PortRange
	if r.SourcePortRange != nil {
		sourcePortRange = &PortRange{
			Min: r.SourcePortRange.Min,
			Max: r.SourcePortRange.Max,
		}
	}
	securityRule.SourcePortRange = sourcePortRange

	if srcAddressSetUuid.Valid {
		uuid := UUID(srcAddressSetUuid.String)
		securityRule.SrcAddressSetUuid = &uuid
	} else {
		securityRule.SrcAddressSetUuid = nil
	}

	if dstAddressSetUuid.Valid {
		uuid := UUID(dstAddressSetUuid.String)
		securityRule.DstAddressSetUuid = &uuid
	} else {
		securityRule.DstAddressSetUuid = nil
	}

	// Update fields based on the request (excluding vpcId and name)
	securityRule.Priority = r.Priority
	securityRule.Direction = Direction(r.Direction)
	securityRule.SourceIPs = r.Source_IPAddresses
	securityRule.DestinationIPs = r.Destination_IPAddresses

	if r.Protocol != nil {
		p := Protocol(*r.Protocol)
		securityRule.Protocol = &p
	} else {
		securityRule.Protocol = nil
	}

	securityRule.Action = SecurityAction(r.Action)
	var direction string
	if securityRule.Direction == Direction_INGRESS {
		direction = nbdb.ACLDirectionFromLport
	} else {
		direction = nbdb.ACLDirectionToLport
	}
	if len(securityRule.SourceIPs) > 1 {
		if securityRule.SrcAddressSetUuid != nil {
			// Update existing address set
			addrSet := &nbdb.AddressSet{
				UUID:      string(*securityRule.SrcAddressSetUuid),
				Name:      HashForOVN("as_src_" + securityRule.Name),
				Addresses: securityRule.SourceIPs,
			}

			err := libovsdbops.UpdateAddressSetsAddresses(ovnClient, addrSet)
			if err != nil {
				err = fmt.Errorf("could not update source address set for rule %s: %w", securityRule.Name, err)
				return nil, err
			}
		} else {
			// Create new address set
			addrSet := &nbdb.AddressSet{
				UUID:      uuid.NewString(),
				Name:      HashForOVN("as_src_" + securityRule.Name),
				Addresses: securityRule.SourceIPs,
			}
			err := libovsdbops.CreateAddressSets(ovnClient, addrSet)
			if err != nil {
				err = fmt.Errorf("could not create source address set for rule %s: %w", securityRule.Name, err)
				return nil, err
			}
			newUuid := UUID(addrSet.UUID)
			securityRule.SrcAddressSetUuid = &newUuid
		}
	} else {
		// Delete existing address set if necessary
		if securityRule.SrcAddressSetUuid != nil {
			addrSet := &nbdb.AddressSet{
				UUID: string(*securityRule.SrcAddressSetUuid),
			}
			err := libovsdbops.DeleteAddressSets(ovnClient, addrSet)
			if err != nil {
				err = fmt.Errorf("could not delete src address set for rule %s: %w", securityRule.Name, err)
				return nil, err
			}
		}
		securityRule.SrcAddressSetUuid = nil
	}

	// Handle destination address set
	if len(securityRule.DestinationIPs) > 1 {
		if securityRule.DstAddressSetUuid != nil {
			// Update existing address set
			addrSet := &nbdb.AddressSet{
				UUID:      string(*securityRule.DstAddressSetUuid),
				Name:      HashForOVN("as_dst_" + securityRule.Name),
				Addresses: securityRule.DestinationIPs,
			}
			err := libovsdbops.UpdateAddressSetsAddresses(ovnClient, addrSet)
			if err != nil {
				err = fmt.Errorf("could not update destination address set for rule %s: %w", securityRule.Name, err)
				return nil, err
			}
		} else {
			// Create new address set
			addrSet := &nbdb.AddressSet{
				UUID:      uuid.NewString(),
				Name:      HashForOVN("as_dst_" + securityRule.Name),
				Addresses: securityRule.DestinationIPs,
			}
			err := libovsdbops.CreateAddressSets(ovnClient, addrSet)
			if err != nil {
				err = fmt.Errorf("could not create destination address set for rule %s: %w", securityRule.Name, err)
				return nil, err
			}
			newUuid := UUID(addrSet.UUID)
			securityRule.DstAddressSetUuid = &newUuid
		}
	} else {
		// Delete existing address set if necessary
		if securityRule.DstAddressSetUuid != nil {
			addrSet := &nbdb.AddressSet{
				UUID: string(*securityRule.DstAddressSetUuid),
			}
			err := libovsdbops.DeleteAddressSets(ovnClient, addrSet)
			if err != nil {
				err = fmt.Errorf("could not delete dst address set for rule %s: %w", securityRule.Name, err)
				return nil, err
			}
		}
		securityRule.DstAddressSetUuid = nil
	}

	// Update the ACL in OVN
	match := createMatchString(&securityRule)

	// Fetch the ACL from OVN using FindACLs
	acl := &nbdb.ACL{
		UUID: aclUUID,
	}
	acls, err := libovsdbops.FindACLs(ovnClient, []*nbdb.ACL{acl})
	if err != nil {
		err = fmt.Errorf("could not fetch ACL with UUID %s: %w", aclUUID, err)
		return nil, err
	}
	if len(acls) == 0 {
		err = fmt.Errorf("ACL with UUID %s not found", aclUUID)
		return nil, err
	}
	acl = acls[0]

	// Update ACL fields
	acl.Priority = int(securityRule.Priority)
	acl.Match = match
	acl.Direction = direction

	if securityRule.Action == SecurityAction_ALLOW {
		acl.Action = nbdb.ACLActionAllowRelated
	} else {
		acl.Action = nbdb.ACLActionDrop
	}

	// Update the ACL in OVN
	ops, err := libovsdbops.UpdateACLsOps(ovnClient, nil, acl)
	if err != nil {
		err = fmt.Errorf("could not update ACL for rule %s: %w", securityRule.Uuid, err)
		return nil, err
	}
	_, err = libovsdbops.TransactAndCheck(ovnClient, ops)
	if err != nil {
		err = fmt.Errorf("could not update ACL for rule %s: %w", securityRule.Uuid, err)
		return nil, err
	}

	var protocolValue sql.NullInt64
	if securityRule.Protocol != nil {
		protocolValue = sql.NullInt64{Int64: int64(*securityRule.Protocol), Valid: true}
	} else {
		protocolValue = sql.NullInt64{Valid: false}
	}

	_, err = tx.Exec(sqlQueryUpdateSecurityRule,
		securityRule.Priority,
		securityRule.Direction,
		protocolValue,
		pq.Array(securityRule.SourceIPs),
		pq.Array(securityRule.DestinationIPs),
		toInt4Range(securityRule.SourcePortRange),
		toInt4Range(securityRule.DestinationPortRange),
		securityRule.Action,
		securityRule.SrcAddressSetUuid,
		securityRule.DstAddressSetUuid,
		securityRule.Uuid,
	)
	if err != nil {
		Logger.Error(err, "SQL update error when updating security rule")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	Logger.V(DebugLevel).Info(fmt.Sprintf("updated security rule UUID: %s", securityRule.Uuid))

	return &v1.UpdateSecurityRuleResponse{
		SecurityRuleId: &v1.SecurityRuleId{Uuid: string(securityRule.Uuid)},
	}, nil
}
