// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package ip_resource_manager

import (
	"fmt"
	"io/fs"

	gittogrpcsynchronizer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/git_to_grpc_synchronizer"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Return a GitToGrpcSynchronizer for subnets.
func NewSubnetSynchronizer(fsys fs.FS, clientConn *grpc.ClientConn) (*gittogrpcsynchronizer.GitToGrpcSynchronizer, error) {
	synchronizer := &gittogrpcsynchronizer.GitToGrpcSynchronizer{
		Fsys:                             fsys,
		ClientConn:                       clientConn,
		EmptyMessage:                     &pb.CreateSubnetRequest{},
		EmptySearchMethodResponseMessage: &pb.Subnet{},
		SearchMethod:                     "/proto.IpResourceManagerService/SearchSubnetStream",
		CreateMethod:                     "/proto.IpResourceManagerService/PutSubnet",
		UpdateMethod:                     "/proto.IpResourceManagerService/PutSubnet",
		DeleteMethod:                     "/proto.IpResourceManagerService/DeleteSubnet",
		MessageKeyFunc: func(m proto.Message) (any, error) {
			var region string
			var availabilityZone string
			var addressSpace string
			var subnet string
			var vlanDomain string
			switch msg := m.(type) {
			case *pb.Subnet:
				region = msg.Region
				availabilityZone = msg.AvailabilityZone
				addressSpace = msg.AddressSpace
				subnet = msg.Subnet
				vlanDomain = msg.VlanDomain
			case *pb.CreateSubnetRequest:
				region = msg.Region
				availabilityZone = msg.AvailabilityZone
				addressSpace = msg.AddressSpace
				subnet = msg.Subnet
				vlanDomain = msg.VlanDomain
			default:
				return nil, fmt.Errorf("unsupported type")
			}
			return struct {
				Region           string
				AvailabilityZone string
				AddressSpace     string
				Subnet           string
				VlanDomain       string
			}{
				Region:           region,
				AvailabilityZone: availabilityZone,
				AddressSpace:     addressSpace,
				Subnet:           subnet,
				VlanDomain:       vlanDomain,
			}, nil

		},
		MessageToDeleteRequestFunc: func(m proto.Message) (proto.Message, error) {
			msg := m.(*pb.Subnet)
			return &pb.DeleteSubnetRequest{
				Region:           msg.Region,
				AvailabilityZone: msg.AvailabilityZone,
				AddressSpace:     msg.AddressSpace,
				Subnet:           msg.Subnet,
				PrefixLength:     msg.PrefixLength,
			}, nil
		},
		MessageToUpdateRequestFunc: func(fileMessage proto.Message, serverMessage proto.Message) (proto.Message, error) {
			return fileMessage, nil
		},
	}
	return synchronizer, nil
}
