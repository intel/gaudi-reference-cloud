// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mock

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type DispatcherServer struct {
	pb.UnimplementedDispatcherServer
}

func NewDispatcherServer() *DispatcherServer {
	return &DispatcherServer{}
}

func (m *DispatcherServer) GenerateStream(req *pb.DispatcherRequest, stream pb.Dispatcher_GenerateStreamServer) error {
	// Implement your mock behavior here
	return stream.Send(&pb.DispatcherResponse{
		Model: req.Model,
		Response: &pb.GenerateStreamResponse{
			Token: &pb.GenerateAPIToken{
				Text: "Hello World From Mock Server",
			},
		},
		RequestID: req.RequestID,
	})
}
