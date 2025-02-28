// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package handlers

import (
	"database/sql"
	"fmt"
	"strings"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	"github.com/google/uuid"
	pq "github.com/lib/pq"
	libovsdbclient "github.com/ovn-org/libovsdb/client"
	libovsdbops "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/libovsdb/ops"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SecurityGroup struct {
	Uuid    UUID
	Name    string
	VpcId   UUID
	sg_type int32
}

func CreateSecurityGroupHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.CreateSecurityGroupRequest) (*v1.CreateSecurityGroupResponse, error) {
	Logger := Logger.WithName("CreateSecurityGroupHandler")

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

	Logger.V(DebugLevel).Info(fmt.Sprintf("create Security Group %s ", r.Name))

	securityGroup := SecurityGroup{
		Uuid:    UUID(uuid.NewString()),
		Name:    r.Name,
		VpcId:   UUID(r.VpcId.GetUuid()),
		sg_type: int32(r.Type),
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

	if r.GetSecurityGroupId().Uuid != "" {
		securityGroup.Uuid = UUID(r.GetSecurityGroupId().Uuid)
	}
	_, err = tx.Exec(sqlQueryCreateSecurityGroup,
		securityGroup.Uuid,
		securityGroup.Name,
		securityGroup.VpcId,
		securityGroup.sg_type,
	)
	if err != nil {
		Logger.Error(err, "SQL insert")
		return nil, GrpcErrorFromSql(err)
	}

	if len(r.PortIds) > 0 && r.Type == v1.SecurityGroupType_PORT {
		return CreatePortSecurityGroup(&securityGroup, db, tx, ovnClient, r)
	} else if len(r.SubnetIds) > 0 && r.Type == v1.SecurityGroupType_SUBNET {
		return CreateSubnetSecurityGroup(&securityGroup, db, tx, ovnClient, r)
	} else {
		err := fmt.Errorf("type not matching with the list of ports/subnets provided")
		return nil, err
	}
}

func CreatePortSecurityGroup(securityGroup *SecurityGroup, db *sql.DB, tx *sql.Tx, ovnClient libovsdbclient.Client, r *v1.CreateSecurityGroupRequest) (*v1.CreateSecurityGroupResponse, error) {

	var portValues []string
	var portArgs []interface{}

	for i, portId := range r.PortIds {
		portValues = append(portValues, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		portArgs = append(portArgs, portId.GetUuid(), securityGroup.Uuid)
	}
	query := `
	INSERT INTO port_security_group (port_id, security_group_id)
	VALUES ` + strings.Join(portValues, ", ")

	_, err := tx.Exec(query, portArgs...)
	if err != nil {
		Logger.Error(err, "SQL insert error when associating ports")
		return nil, GrpcErrorFromSql(err)
	}

	//Create portgroup
	var pgUUID *UUID
	pg := nbdb.PortGroup{
		Name: HashForOVN(r.Name),
	}
	for _, portId := range r.PortIds {
		Logger.V(DebugLevel).Info(fmt.Sprintf("Add port %s to Group %s ", portId.Uuid, securityGroup.Name))
		pg.Ports = append(pg.Ports, portId.Uuid)
	}

	err = libovsdbops.CreatePortGroup(ovnClient, &pg)
	if err != nil {
		err = fmt.Errorf("could not create port group %s: %w ", r.Name, err)
		return nil, err
	}
	pgUUID = (*UUID)(&pg.UUID)
	/* TODO: update AZ when the info is in the API */
	_, err = tx.Exec(sqlQueryCreatePortGroup, pgUUID, securityGroup.Uuid, "AZ1")
	if err != nil {
		Logger.Error(err, "SQL update error when updating port_group_uuid")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}
	return &v1.CreateSecurityGroupResponse{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: string(securityGroup.Uuid)}}, nil
}

func CreateSubnetSecurityGroup(securityGroup *SecurityGroup, db *sql.DB, tx *sql.Tx, ovnClient libovsdbclient.Client, r *v1.CreateSecurityGroupRequest) (*v1.CreateSecurityGroupResponse, error) {

	var subnetValues []string
	var subnetArgs []interface{}

	for i, subnetId := range r.SubnetIds {
		subnetValues = append(subnetValues, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		subnetArgs = append(subnetArgs, subnetId.GetUuid(), securityGroup.Uuid)
	}

	query := `
		INSERT INTO subnet_security_group (subnet_id, security_group_id)
		VALUES ` + strings.Join(subnetValues, ", ")

	_, err := tx.Exec(query, subnetArgs...)
	if err != nil {
		Logger.Error(err, "SQL insert error when associating subnets")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}
	return &v1.CreateSecurityGroupResponse{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: string(securityGroup.Uuid)}}, nil
}

func DeleteSecurityGroupHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.DeleteSecurityGroupRequest) (*v1.DeleteSecurityGroupResponse, error) {
	Logger := Logger.WithName("DeleteSecurityGroupHandler")
	securityGroupUuid := r.GetSecurityGroupId().GetUuid()

	var vpcId string
	err := db.QueryRow(sqlQueryVpcIdFromSecurityGroup, securityGroupUuid).Scan(&vpcId)
	if err != nil {
		Logger.Error(err, "cannot find VPC")
		return nil, GrpcErrorFromSql(err)
	}
	mu := GetVpcMutex(vpcId)
	LockVpcMutex(mu, vpcId)
	defer UnlockVpcMutex(mu, vpcId)

	Logger.V(DebugLevel).Info(fmt.Sprintf("delete Security Group %s", securityGroupUuid))

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
	var portGroupUUID *string
	var sg_type int32
	err = tx.QueryRow(sqlQueryGetSecurityGroup, securityGroupUuid).Scan(&securityGroupName, &vpc_id, &sg_type, &portGroupUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			errMsg := fmt.Sprintf("No group found with UUID: %s", securityGroupUuid)
			Logger.Error(err, errMsg)
			return nil, status.Errorf(codes.NotFound, errMsg)
		}
		Logger.Error(err, "Failed to check security group existence")
		return nil, GrpcErrorFromSql(err)
	}

	// Delete security group dependencies
	_, err = tx.Exec(sqlQueryDeleteSecurityGroupDependencies, securityGroupUuid)
	if err != nil {
		Logger.Error(err, "SQL delete error when deleting security group and its dependencies")
		return nil, GrpcErrorFromSql(err)
	}

	if portGroupUUID != nil {
		err := libovsdbops.DeletePortGroups(ovnClient, HashForOVN(securityGroupName))
		if err != nil {
			err = fmt.Errorf("could not delete port group %s: %w", securityGroupName, err)
			return nil, err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	Logger.V(DebugLevel).Info("Successfully deleted security group and its dependencies")

	return &v1.DeleteSecurityGroupResponse{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: r.SecurityGroupId.Uuid},
	}, nil
}

func GetSecurityGroupHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.GetSecurityGroupRequest) (*v1.GetSecurityGroupResponse, error) {
	var err error = nil
	uuid := r.GetSecurityGroupId().GetUuid()
	Logger := Logger.WithName("GetSecurityGroupHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("get security group %s", uuid))

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

	securityGroup := &v1.SecurityGroup{
		Id:    &v1.SecurityGroupId{Uuid: uuid},
		VpcId: &v1.VPCId{},
	}

	// Query to fetch security group data
	err = db.QueryRow(sqlQueryGetSecurityGroupInfo, uuid).Scan(
		&securityGroup.Name,
		&securityGroup.VpcId.Uuid,
		&securityGroup.Type,
	)
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}

	// Query to fetch security rule IDs associated with the security group
	rows, err := db.Query(sqlQueryGetSecurityGroupRules, uuid)
	if err != nil {
		Logger.Error(err, "SQL query error when fetching security group rules")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	for rows.Next() {
		securityRuleId := &v1.SecurityRuleId{}
		err := rows.Scan(&securityRuleId.Uuid)
		if err != nil {
			Logger.Error(err, "SQL scan error for security rule")
			return nil, GrpcErrorFromSql(err)
		}
		securityGroup.SecurityRuleIds = append(securityGroup.SecurityRuleIds, securityRuleId)
	}

	if err := rows.Err(); err != nil {
		Logger.Error(err, "SQL rows error for security rules")
		return nil, GrpcErrorFromSql(err)
	}

	// Query to fetch port IDs associated with the security group
	rows, err = db.Query(sqlQueryGetSecurityGroupPorts, uuid)
	if err != nil {
		Logger.Error(err, "SQL query error when fetching security group ports")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	for rows.Next() {
		portId := &v1.PortId{}
		err := rows.Scan(&portId.Uuid)
		if err != nil {
			Logger.Error(err, "SQL scan error for port")
			return nil, GrpcErrorFromSql(err)
		}
		securityGroup.PortIds = append(securityGroup.PortIds, portId)
	}

	if err := rows.Err(); err != nil {
		Logger.Error(err, "SQL rows error for ports")
		return nil, GrpcErrorFromSql(err)
	}

	// Query to fetch subnet IDs associated with the security group
	rows, err = db.Query(sqlQueryGetSecurityGroupSubnets, uuid)
	if err != nil {
		Logger.Error(err, "SQL query error when fetching security group subnets")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	for rows.Next() {
		subnetId := &v1.SubnetId{}
		err := rows.Scan(&subnetId.Uuid)
		if err != nil {
			Logger.Error(err, "SQL scan error for subnet")
			return nil, GrpcErrorFromSql(err)
		}
		securityGroup.SubnetIds = append(securityGroup.SubnetIds, subnetId)
	}

	if err := rows.Err(); err != nil {
		Logger.Error(err, "SQL rows error for subnets")
		return nil, GrpcErrorFromSql(err)
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	return &v1.GetSecurityGroupResponse{
		SecurityGroup: securityGroup,
	}, nil
}

func ListSecurityGroupsHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.ListSecurityGroupsRequest) (*v1.ListSecurityGroupsResponse, error) {
	var err error = nil
	Logger := Logger.WithName("ListSecurityGroupsHandler")
	Logger.V(DebugLevel).Info("listing all security groups")

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

	// Query to get all security group UUIDs
	rows, err := db.Query(sqlQueryListSecurityGroups)
	if err != nil {
		Logger.Error(err, "SQL query error")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	var securityGroupUUIDs []*v1.SecurityGroupId

	for rows.Next() {
		securityGroupId := &v1.SecurityGroupId{}
		err := rows.Scan(&securityGroupId.Uuid)
		if err != nil {
			Logger.Error(err, "SQL scan error")
			return nil, GrpcErrorFromSql(err)
		}
		securityGroupUUIDs = append(securityGroupUUIDs, securityGroupId)
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

	return &v1.ListSecurityGroupsResponse{
		SecurityGroupIds: securityGroupUUIDs,
	}, nil
}
func UpdateSecurityGroupHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.UpdateSecurityGroupRequest) (*v1.UpdateSecurityGroupResponse, error) {
	Logger := Logger.WithName("UpdateSecurityGroupHandler")
	securityGroupUuid := r.GetSecurityGroupId().GetUuid()
	Logger.V(DebugLevel).Info(fmt.Sprintf("update Security Group %s", securityGroupUuid))

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

	// Fetch existing security group name and port group UUID
	var securityGroupName, vpc_id string
	var portGroupUUID *string
	var sg_type int32
	err = tx.QueryRow(sqlQueryGetSecurityGroup, securityGroupUuid).Scan(&securityGroupName, &vpc_id, &sg_type, &portGroupUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			errMsg := fmt.Sprintf("No group found with UUID: %s", securityGroupUuid)
			Logger.Error(err, errMsg)
			return nil, status.Errorf(codes.NotFound, errMsg)
		}
		Logger.Error(err, "Failed to fetch security group")
		return nil, GrpcErrorFromSql(err)
	}

	// Fetch existing security rule, subnet, and port UUIDs associated with the security group
	RuleUUIDs, err := getGroupDeps(sqlQueryGetSecurityGroupSecurityRuleUUIDs, tx, securityGroupUuid)
	if err != nil {
		return nil, err
	}
	var aclUUIDs []*nbdb.ACL
	var aclUUIDs_str []string
	aclRows, err := tx.Query(sqlQueryGetSecurityRulesACLs, pq.Array(RuleUUIDs))
	if err != nil {
		Logger.Error(err, "SQL query error when fetching ACL UUIDs")
		return nil, GrpcErrorFromSql(err)
	}
	defer aclRows.Close()

	for aclRows.Next() {
		var aclUUID string
		if err := aclRows.Scan(&aclUUID); err != nil {
			return nil, fmt.Errorf("error scanning ACL UUID: %v", err)
		}
		aclUUIDs = append(aclUUIDs, &nbdb.ACL{UUID: aclUUID})
		aclUUIDs_str = append(aclUUIDs_str, aclUUID)
	}

	if sg_type == int32(v1.SecurityGroupType_PORT) {
		existingPortUUIDs, err := getGroupDeps(sqlQueryGetSecurityGroupPortUUIDs, tx, securityGroupUuid)
		if err != nil {
			return nil, err
		}
		newPortUUIDs := extractUUIDsFromPorts(r.PortIds)
		toAddPorts, toRemovePorts := getUUIDsToAddAndRemove(existingPortUUIDs, newPortUUIDs)

		if len(toRemovePorts) > 0 {
			_, err = tx.Exec(sqlQueryDetachOldPorts, pq.Array(toRemovePorts), securityGroupUuid)
			if err != nil {
				Logger.Error(err, "Failed to delete old ports")
				return nil, GrpcErrorFromSql(err)
			}

			err := libovsdbops.DeletePortsFromPortGroup(ovnClient, HashForOVN(securityGroupName), toRemovePorts...)
			if err != nil {
				err = fmt.Errorf("could not delete ports from port group %s: %w", securityGroupName, err)
				return nil, err
			}
		}
		if len(toAddPorts) > 0 {
			_, err = tx.Exec(sqlQueryInsertNewPorts, pq.Array(toAddPorts), securityGroupUuid)
			if err != nil {
				Logger.Error(err, "Failed to insert new ports")
				return nil, GrpcErrorFromSql(err)
			}
			err := libovsdbops.AddPortsToPortGroup(ovnClient, HashForOVN(securityGroupName), toAddPorts...)
			if err != nil {
				err = fmt.Errorf("could not add ports to port group %s: %w", securityGroupName, err)
				return nil, err
			}
		}
	} else {
		existingSubnetUUIDs, err := getGroupDeps(sqlQueryGetSecurityGroupSubnetUUIDs, tx, securityGroupUuid)
		if err != nil {
			return nil, err
		}
		newSubnetUUIDs := extractUUIDsFromSubnets(r.SubnetIds)
		toAddSubnets, toRemoveSubnets := getUUIDsToAddAndRemove(existingSubnetUUIDs, newSubnetUUIDs)
		if len(toRemoveSubnets) > 0 {
			_, err = tx.Exec(sqlQueryDetachOldSubnets, pq.Array(toRemoveSubnets), securityGroupUuid)
			if err != nil {
				Logger.Error(err, "Failed to delete old subnets")
				return nil, GrpcErrorFromSql(err)
			}
			for _, subnetId := range toRemoveSubnets {
				p := func(item *nbdb.LogicalSwitch) bool { return item.UUID == string(subnetId) }
				err = libovsdbops.RemoveACLsFromLogicalSwitchesWithPredicate(ovnClient, p, aclUUIDs...)
				if err != nil {
					return nil, fmt.Errorf("could not remove old ACLs from subnet %s for group %s: %w",
						subnetId, securityGroupUuid, err)
				}
			}
		}
		if len(toAddSubnets) > 0 {
			_, err = tx.Exec(sqlQueryInsertNewSubnets, pq.Array(toAddSubnets), securityGroupUuid)
			if err != nil {
				Logger.Error(err, "Failed to insert new subnets")
				return nil, GrpcErrorFromSql(err)
			}
			for _, subnetId := range toAddSubnets {
				swselect := nbdb.LogicalSwitch{
					UUID: string(subnetId),
				}
				sw, err := libovsdbops.GetLogicalSwitch(ovnClient, &swselect)
				if err != nil {
					err = fmt.Errorf("could not fetch subnet %s :%w", subnetId, err)
					return nil, err
				}
				sw.ACLs = aclUUIDs_str
				err = libovsdbops.CreateOrUpdateLogicalSwitch(ovnClient, sw)
				if err != nil {
					err = fmt.Errorf("could not assign ACLs to subnet %s of Group %s :%w", subnetId, securityGroupUuid, err)
					return nil, err
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}
	Logger.V(DebugLevel).Info(fmt.Sprintf("Group %s updated successfully", securityGroupName))
	return &v1.UpdateSecurityGroupResponse{
		SecurityGroupId: &v1.SecurityGroupId{Uuid: securityGroupUuid},
	}, nil
}

// Helper function to fetch existing UUIDs from a table
func getGroupDeps(query string, tx *sql.Tx, groupID string) ([]string, error) {
	rows, err := tx.Query(query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query UUIDs: %v", err)
	}
	defer rows.Close()

	var uuids []string
	for rows.Next() {
		var uuid string
		if err := rows.Scan(&uuid); err != nil {
			return nil, fmt.Errorf("failed to scan UUID: %v", err)
		}
		uuids = append(uuids, uuid)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	return uuids, nil
}

// Extract UUIDs from SecurityRuleId structs
func extractUUIDsFromSecurityRules(securityRules []*v1.SecurityRuleId) []string {
	var uuids []string
	for _, rule := range securityRules {
		uuids = append(uuids, rule.GetUuid())
	}
	return uuids
}

// Extract UUIDs from SubnetId structs
func extractUUIDsFromSubnets(subnetIds []*v1.SubnetId) []string {
	var uuids []string
	for _, subnet := range subnetIds {
		uuids = append(uuids, subnet.GetUuid())
	}
	return uuids
}

// Extract UUIDs from PortId structs
func extractUUIDsFromPorts(portIds []*v1.PortId) []string {
	var uuids []string
	for _, port := range portIds {
		uuids = append(uuids, port.GetUuid())
	}
	return uuids
}

// Get UUIDs to add and remove
func getUUIDsToAddAndRemove(existingUUIDs []string, newUUIDs []string) ([]string, []string) {
	existingSet := make(map[string]struct{}, len(existingUUIDs))
	for _, uuid := range existingUUIDs {
		existingSet[uuid] = struct{}{}
	}

	toAdd := []string{}
	toRemove := []string{}

	// Add new UUIDs if not in the existing set
	for _, newUUID := range newUUIDs {
		if _, found := existingSet[newUUID]; !found {
			toAdd = append(toAdd, newUUID)
		} else {
			delete(existingSet, newUUID) // Remove from existingSet if in both lists
		}
	}

	// What remains in existingSet needs to be removed
	for uuid := range existingSet {
		toRemove = append(toRemove, uuid)
	}

	return toAdd, toRemove
}
