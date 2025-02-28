// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package instance_type

import (
	"io/fs"

	gittogrpcsynchronizer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/git_to_grpc_synchronizer"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Return a GitToGrpcSynchronizer for Instance Types.
func NewInstanceTypeSynchronizer(fsys fs.FS, clientConn *grpc.ClientConn) (*gittogrpcsynchronizer.GitToGrpcSynchronizer, error) {
	synchronizer := &gittogrpcsynchronizer.GitToGrpcSynchronizer{
		Fsys:                             fsys,
		ClientConn:                       clientConn,
		EmptyMessage:                     &pb.InstanceType{},
		EmptySearchMethodResponseMessage: &pb.InstanceType{},
		SearchMethod:                     "/proto.InstanceTypeService/SearchStream",
		CreateMethod:                     "/proto.InstanceTypeService/Put",
		UpdateMethod:                     "/proto.InstanceTypeService/Put",
		DeleteMethod:                     "/proto.InstanceTypeService/Delete",
		MessageKeyFunc: func(m proto.Message) (any, error) {
			return m.(*pb.InstanceType).Metadata.Name, nil
		},
		MessageToDeleteRequestFunc: func(m proto.Message) (proto.Message, error) {
			return &pb.InstanceTypeDeleteRequest{
				Metadata: &pb.InstanceTypeDeleteRequest_Metadata{
					Name: m.(*pb.InstanceType).Metadata.Name,
				},
			}, nil
		},
		MessageToUpdateRequestFunc: func(fileMessage proto.Message, serverMessage proto.Message) (proto.Message, error) {
			return fileMessage, nil
		},
	}
	return synchronizer, nil
}
