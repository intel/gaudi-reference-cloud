// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package machine_image

import (
	"io/fs"

	gittogrpcsynchronizer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/git_to_grpc_synchronizer"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// Return a GitToGrpcSynchronizer for Machine Images.
func NewMachineImageSynchronizer(fsys fs.FS, clientConn *grpc.ClientConn) (*gittogrpcsynchronizer.GitToGrpcSynchronizer, error) {
	synchronizer := &gittogrpcsynchronizer.GitToGrpcSynchronizer{
		Fsys:                             fsys,
		ClientConn:                       clientConn,
		EmptyMessage:                     &pb.MachineImage{},
		EmptySearchMethodResponseMessage: &pb.MachineImage{},
		SearchMethod:                     "/proto.MachineImageService/SearchStream",
		CreateMethod:                     "/proto.MachineImageService/Put",
		UpdateMethod:                     "/proto.MachineImageService/Put",
		DeleteMethod:                     "/proto.MachineImageService/Delete",
		MessageKeyFunc: func(m proto.Message) (any, error) {
			return m.(*pb.MachineImage).Metadata.Name, nil
		},
		MessageToDeleteRequestFunc: func(m proto.Message) (proto.Message, error) {
			return &pb.MachineImageDeleteRequest{
				Metadata: &pb.MachineImageDeleteRequest_Metadata{
					Name: m.(*pb.MachineImage).Metadata.Name,
				},
			}, nil
		},
		MessageToUpdateRequestFunc: func(fileMessage proto.Message, serverMessage proto.Message) (proto.Message, error) {
			return fileMessage, nil
		},
	}
	return synchronizer, nil
}
