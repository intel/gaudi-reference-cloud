// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package handlers

import (
	"database/sql"
	"fmt"
	"net"

	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sdn-vn-controller/api/sdn/v1"

	libovsdbclient "github.com/ovn-org/libovsdb/client"
	libovsdbops "github.com/ovn-org/ovn-kubernetes/go-controller/pkg/libovsdb/ops"
	"github.com/ovn-org/ovn-kubernetes/go-controller/pkg/nbdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Deprecated: would be replaced by DB.
type Subnet struct {
	Uuid             UUID
	Name             string
	Cidr             string
	AvailabilityZone string
	VpcId            UUID
}

func ListSubnetsHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.ListSubnetsRequest) (*v1.ListSubnetsResponse, error) {
	var err error = nil
	var subnets []*v1.SubnetId
	Logger := Logger.WithName("ListSubnetsHandler")
	Logger.V(DebugLevel).Info("list subnets")

	rows, err := db.Query(sqlQueryListSubnet)
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}
	defer rows.Close()

	var subnetId string
	for rows.Next() {
		err = rows.Scan(&subnetId)
		if err != nil {
			Logger.Error(err, "cannot scan SQL rows for subnet")
			return nil, GrpcErrorFromSql(err)
		}
		subnets = append(subnets, &v1.SubnetId{
			Uuid: subnetId,
		})
	}

	return &v1.ListSubnetsResponse{SubnetIds: subnets}, err
}

func GetSubnetHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.GetSubnetRequest) (*v1.GetSubnetResponse, error) {
	var err error = nil
	uuid := r.GetSubnetId().GetUuid()
	Logger := Logger.WithName("GetSubnetHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("get subnet %s", r.SubnetId.Uuid))

	// Locate the subnet using the Id
	var name, cidr, az, vpcId string
	err = db.QueryRow(sqlQueryGetSubnet, uuid).Scan(&name, &cidr, &az, &vpcId)
	if err == sql.ErrNoRows {
		err = status.Errorf(codes.NotFound, "cannot find requested subnet")
		return nil, err
	}
	if err != nil {
		Logger.Error(err, "SQL select error")
		return nil, GrpcErrorFromSql(err)
	}

	return &v1.GetSubnetResponse{Subnet: &v1.Subnet{
		Id:               &v1.SubnetId{Uuid: uuid},
		Name:             name,
		Cidr:             cidr,
		AvailabilityZone: az,
		VpcId:            &v1.VPCId{Uuid: vpcId},
	}}, err
}

// isOverlap checks if two CIDRs overlap
func isOverlap(cidr1, cidr2 string) (bool, error) {
	_, ipNet1, err := net.ParseCIDR(cidr1)
	if err != nil {
		return false, err
	}

	_, ipNet2, err := net.ParseCIDR(cidr2)
	if err != nil {
		return false, err
	}

	return ipNet1.Contains(ipNet2.IP) || ipNet2.Contains(ipNet1.IP), nil
}

func CreateSubnetHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.CreateSubnetRequest) (*v1.CreateSubnetResponse, error) {
	var err error
	uuid := r.GetSubnetId().GetUuid()
	vpcId := r.GetVpcId().GetUuid()
	Logger := Logger.WithName("CreateSubnetHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("create subnet: %s", r.Name))

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

	// Check for CIDR overlap
	for _, gw := range gateways[INTERNET] {
		overlap, err := isOverlap(r.Cidr, gw.AccessNetwork)
		if err != nil {
			return nil, err
		}
		if overlap {
			return nil, fmt.Errorf("CIDR %s overlaps with access network %s of gateway %d",
				r.Cidr, gw.AccessNetwork, gw.ChassisId)
		}
	}

	// Logical switch information for OVSDB
	logicalSwitch := nbdb.LogicalSwitch{
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

	_, err = tx.Exec(sqlQueryCreateSubnet, uuid, r.Name, r.Cidr, r.AvailabilityZone, vpcId)
	if err != nil {
		Logger.Error(err, "SQL insert error")
		return nil, GrpcErrorFromSql(err)
	}

	// invoke backend to program the object
	err = libovsdbops.CreateOrUpdateLogicalSwitch(ovnClient, &logicalSwitch)
	if err != nil {
		Logger.Error(err, "failed to create subnet in OVN")
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	// return the UUID
	Logger.V(DebugLevel).Info(fmt.Sprintln("The UUID of the created subnet is ", uuid))
	return &v1.CreateSubnetResponse{SubnetId: &v1.SubnetId{Uuid: uuid}}, err
}

func DeleteSubnetHandler(db *sql.DB, ovnClient libovsdbclient.Client, r *v1.DeleteSubnetRequest) (*v1.DeleteSubnetResponse, error) {
	var err error = nil
	uuid := r.GetSubnetId().GetUuid()
	Logger := Logger.WithName("DeleteSubnetHandler")
	Logger.V(DebugLevel).Info(fmt.Sprintf("subnet delete: %s", r.SubnetId.Uuid))

	var vpcId string
	err = db.QueryRow(sqlQueryVpcIdFromSubnet, uuid).Scan(&vpcId)
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
	err = tx.QueryRow(sqlQueryDeleteSubnetAndRefs, uuid).Scan(&subnetName)
	if err != nil {
		if err == sql.ErrNoRows {
			Logger.Error(err, fmt.Sprintf("No subnet found with UUID: %s", uuid))
			return nil, fmt.Errorf("No subnet found with UUID: %s", uuid)
		}
		Logger.Error(err, fmt.Sprintf("Error querying subnet with UUID %s: %v", uuid, err))
		return nil, GrpcErrorFromSql(err)
	}

	// invoke backend to program the object
	err = libovsdbops.DeleteLogicalSwitch(ovnClient, subnetName)
	if err != nil {
		Logger.Error(err, "failed to delete subnet in OVN")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		Logger.Error(err, "SQL commit error")
		return nil, GrpcErrorFromSql(err)
	}

	// return the UUID and other information
	return &v1.DeleteSubnetResponse{SubnetId: &v1.SubnetId{Uuid: r.SubnetId.Uuid}}, err
}
