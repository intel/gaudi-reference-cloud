// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var defaultLabels = map[string]string{
	"label1": "val1",
	"label2": "val2",
}

var updatedLabels = map[string]string{
	"label1": "val1",
	"label2": "val2",
	"label3": "val3",
}

func NewCreateVPCRequest(cloudAccountId, name, cidr string) *pb.VPCCreateRequest {
	req := &pb.VPCCreateRequest{
		Metadata: &pb.VPCMetadataCreate{
			Name:           name,
			CloudAccountId: cloudAccountId,
			Labels:         defaultLabels,
		},
		Spec: &pb.VPCSpec{
			CidrBlock: cidr,
		},
	}
	return req
}

func NewUpdateVPCByNameRequest(cloudAccountId, name string) *pb.VPCUpdateRequest {
	req := &pb.VPCUpdateRequest{
		Metadata: &pb.VPCMetadataUpdate{
			CloudAccountId: cloudAccountId,
			NameOrId: &pb.VPCMetadataUpdate_Name{
				Name: name,
			},
			Labels: updatedLabels,
		},
	}
	return req
}

func NewUpdateVPCByIdRequest(cloudAccountId, id string) *pb.VPCUpdateRequest {
	req := &pb.VPCUpdateRequest{
		Metadata: &pb.VPCMetadataUpdate{
			CloudAccountId: cloudAccountId,
			NameOrId: &pb.VPCMetadataUpdate_ResourceId{
				ResourceId: id,
			},
			Labels: updatedLabels,
		},
	}
	return req
}

func NewUpdateVPCByIdRequest_MissingMetadata(cloudAccountId, id string) *pb.VPCUpdateRequest {
	return &pb.VPCUpdateRequest{}
}

func NewGetVPCByIdRequest(cloudAccountId, id string) *pb.VPCGetRequest {
	req := &pb.VPCGetRequest{
		Metadata: &pb.VPCMetadataReference{
			CloudAccountId: cloudAccountId,
			NameOrId: &pb.VPCMetadataReference_ResourceId{
				ResourceId: id,
			},
		},
	}
	return req
}

func NewGetVPCByNameRequest(cloudAccountId, name string) *pb.VPCGetRequest {
	req := &pb.VPCGetRequest{
		Metadata: &pb.VPCMetadataReference{
			CloudAccountId: cloudAccountId,
			NameOrId: &pb.VPCMetadataReference_Name{
				Name: name,
			},
		},
	}
	return req
}

func NewCreateVpcAndSubnet(ctx context.Context, cloudAccountId string, vpcServiceClient pb.VPCServiceClient, subnetServiceClient pb.SubnetServiceClient) (*pb.VPC, *pb.VPCSubnet, error) {
	// Create a VPC
	cidr := "10.0.0.0/16"
	name := "default"
	availabilityZone := "us-dev-1a"

	createVPCReq := NewCreateVPCRequest(cloudAccountId, name, cidr)
	gotVPC, err := vpcServiceClient.Create(ctx, createVPCReq)
	if err != nil {
		return nil, nil, err
	}

	// Create a Subnet for the vpc
	createReq := &pb.SubnetCreateRequest{
		Metadata: &pb.SubnetMetadataCreate{
			CloudAccountId: cloudAccountId,
		},
		Spec: &pb.SubnetSpec{
			CidrBlock:        "10.0.0.1/24",
			AvailabilityZone: availabilityZone,
			VpcId:            gotVPC.Metadata.ResourceId,
		},
	}
	gotSubnet, err := subnetServiceClient.Create(ctx, createReq)
	if err != nil {
		return nil, nil, err
	}

	return gotVPC, gotSubnet, nil
}
