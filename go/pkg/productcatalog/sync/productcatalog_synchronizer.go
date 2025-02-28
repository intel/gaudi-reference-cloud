// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package sync

import (
	"fmt"
	"io/fs"

	gittogrpcsynchronizer "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/git_to_grpc_synchronizer"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// NewProductSynchronizer creates a new GitToGrpcSynchronizer for Product messages.
func NewProductSynchronizer(fsys fs.FS, clientConn *grpc.ClientConn) (*gittogrpcsynchronizer.GitToGrpcSynchronizer, error) {
	synchronizer := &gittogrpcsynchronizer.GitToGrpcSynchronizer{
		Fsys:                             fsys,
		ClientConn:                       clientConn,
		EmptyMessage:                     &pb.DefaultProduct{}, // An empty instance of the Product message
		EmptySearchMethodResponseMessage: &pb.DefaultProduct{},
		SearchMethod:                     "/proto.ProductSyncService/SearchStream",
		CreateMethod:                     "/proto.ProductSyncService/Put",
		UpdateMethod:                     "/proto.ProductSyncService/Put",
		DeleteMethod:                     "/proto.ProductSyncService/Delete", // Keep the delete method
		MessageKeyFunc: func(m proto.Message) (any, error) {
			return m.(*pb.DefaultProduct).Metadata.Name, nil
		},
		MessageToDeleteRequestFunc: func(message proto.Message) (proto.Message, error) {
			// Return an empty product to avoid actual deletion
			return &pb.DefaultProductDeleteRequest{Id: ""}, nil
		},
		MessageToUpdateRequestFunc: func(fileMessage proto.Message, serverMessage proto.Message) (proto.Message, error) {
			// Assuming the PutProduct method requires the entire Product message
			// and that the fileMessage is already the updated version of the product
			return fileMessage, nil
		},
		MessageComparatorFunc: productComparator, // Use the custom comparator function
	}

	return synchronizer, nil
}

func productComparator(fileMessage proto.Message, serverMessage proto.Message) (bool, error) {
	fileProduct, ok1 := fileMessage.(*pb.DefaultProduct)
	serverProduct, ok2 := serverMessage.(*pb.DefaultProduct)
	if !ok1 || !ok2 {
		return false, fmt.Errorf("messages are not of type *pb.DefaultProduct")
	}

	// Compare Metadata fields
	if fileProduct.Metadata.Name != serverProduct.Metadata.Name {
		return false, nil
	}

	return true, nil
}
