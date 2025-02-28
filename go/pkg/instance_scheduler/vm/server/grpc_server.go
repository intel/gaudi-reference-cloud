// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"net"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	cloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type GrpcServer struct {
	cfg               *cloudv1alpha1.VmInstanceSchedulerConfig
	listener          net.Listener
	grpcServer        *grpc.Server
	schedulingService *SchedulingService
}

func NewGrpcServer(ctx context.Context, cfg *cloudv1alpha1.VmInstanceSchedulerConfig, mgr ctrl.Manager, schedServer *SchedulingService, listener net.Listener) (*GrpcServer, error) {
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    120 * time.Second, // The time a connection is kept alive without any activity.
			Timeout: 20 * time.Second,  // Maximum time the server waits for activity before closing the connection.
		}),
	}
	grpcServer, err := grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return nil, err
	}
	grpcService := &GrpcServer{
		cfg:               cfg,
		listener:          listener,
		grpcServer:        grpcServer,
		schedulingService: schedServer,
	}
	if err := mgr.Add(manager.RunnableFunc(grpcService.Run)); err != nil {
		return nil, err
	}
	return grpcService, nil
}

// Run GRPC server, blocking until an error occurs.
func (s *GrpcServer) Run(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("GrpcServer.Run")
	log.Info("BEGIN", logkeys.ListenPort, s.listener.Addr().(*net.TCPAddr).Port)
	defer log.Info("END")
	defer utilruntime.HandleCrash()
	err := func() error {
		// We don't want to answer GRPC requests until the scheduler cache is up-to-date.
		if err := s.schedulingService.Sched.WaitForCacheSync(ctx); err != nil {
			return err
		}
		pb.RegisterInstanceSchedulingServiceServer(s.grpcServer, s.schedulingService)
		reflection.Register(s.grpcServer)
		// Stop GRPC server when context is cancelled.
		go func() {
			<-ctx.Done()
			s.grpcServer.GracefulStop()
		}()
		log.Info("Service running")
		return s.grpcServer.Serve(s.listener)
	}()
	if err != nil {
		// Log the error because the caller (manager) does not.
		log.Error(err, logkeys.Error)
	}
	return err
}
