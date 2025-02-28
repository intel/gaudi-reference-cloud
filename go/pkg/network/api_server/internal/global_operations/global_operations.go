// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package global_operations

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/subnet"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/vpc"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultCidrForVPC    = "172.31.0.0/16"
	defaultCidrForSubnet = "172.31.0.0/20"
)

type GlobalOperationsService struct {
	pb.UnimplementedGlobalOperationsServiceServer
	vpcService        *vpc.VPCService
	subnetService     *subnet.SubnetService
	availabilityZones []string
}

func NewGlobalOperationsService(
	vpcService *vpc.VPCService,
	subnetService *subnet.SubnetService,
	availabilityZones []string,
) (*GlobalOperationsService, error) {
	return &GlobalOperationsService{
		vpcService:        vpcService,
		subnetService:     subnetService,
		availabilityZones: availabilityZones,
	}, nil
}

func (s *GlobalOperationsService) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	log := log.FromContext(ctx).WithName("GlobalOperationsService.Ping")
	log.Info("Ping")
	return &emptypb.Empty{}, nil
}

// Public API: Create a new network (vpc, subnet, router, etc) for a cloud account.
func (s *GlobalOperationsService) CreateDefault(ctx context.Context, req *pb.CreateDefaultRequest) (*pb.VPC, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("GlobalOperationsService.CreateDefault").WithValues(logkeys.CloudAccountId, req.Metadata.CloudAccountId).Start()
	defer span.End()
	logger.Info("Request", logkeys.Request, req)

	cloudAccountId := req.Metadata.CloudAccountId

	resp, err := s.vpcService.Create(ctx, &pb.VPCCreateRequest{
		Metadata: &pb.VPCMetadataCreate{
			CloudAccountId: cloudAccountId,
		},
		Spec: &pb.VPCSpec{
			CidrBlock: defaultCidrForVPC,
		},
	})
	if err != nil {
		return nil, err
	}

	vpcId := resp.Metadata.ResourceId

	// Create subnet for each AZ.
	for _, availabilityZone := range s.availabilityZones {
		_, err := s.subnetService.Create(ctx, &pb.SubnetCreateRequest{
			Metadata: &pb.SubnetMetadataCreate{
				CloudAccountId: cloudAccountId,
			},
			Spec: &pb.SubnetSpec{
				CidrBlock:        defaultCidrForSubnet,
				AvailabilityZone: availabilityZone,
				VpcId:            vpcId,
			},
		})
		if err != nil {
			logger.Error(err, "Failed to create subnet for AvailabilityZone", "az", availabilityZone)
		}
	}

	log.LogResponseOrError(logger, req, resp, err)
	return resp, err
}
