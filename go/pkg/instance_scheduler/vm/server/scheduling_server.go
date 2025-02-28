// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"net"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	ctrl "sigs.k8s.io/controller-runtime"
)

// A composite server that consists of the GRPC server and the SchedulingService.
type SchedulingServer struct {
	GrpcServer        *GrpcServer
	SchedulingService *SchedulingService
}

func NewSchedulingServer(ctx context.Context, cfg *privatecloudv1alpha1.VmInstanceSchedulerConfig, mgr ctrl.Manager, listener net.Listener, instanceTypeServiceClient pb.InstanceTypeServiceClient) (*SchedulingServer, error) {
	schedulingService, err := NewSchedulingService(ctx, cfg, mgr, instanceTypeServiceClient)
	if err != nil {
		return nil, err
	}
	grpcServer, err := NewGrpcServer(ctx, cfg, mgr, schedulingService, listener)
	if err != nil {
		return nil, err
	}
	return &SchedulingServer{
		GrpcServer:        grpcServer,
		SchedulingService: schedulingService,
	}, nil
}
