// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package vpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Check given vpc is valid and exists in the db.
// Return the vpc that exists in the db.
func (s *VPCService) ValidateVPC(ctx context.Context, cloudAccountId string, vpcId string) (*pb.VPCPrivate, error) {

	if vpcId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid vpc id")
	}

	if _, err := uuid.Parse(vpcId); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vpc id")
	}

	vpc, err := s.GetPrivate(ctx, &pb.VPCGetPrivateRequest{
		Metadata: &pb.VPCMetadataReference{
			CloudAccountId: cloudAccountId,
			NameOrId: &pb.VPCMetadataReference_ResourceId{
				ResourceId: vpcId,
			},
		},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound || status.Code(err) == codes.InvalidArgument {
			return nil, status.Error(codes.InvalidArgument, "invalid vpcId")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return vpc, nil
}
