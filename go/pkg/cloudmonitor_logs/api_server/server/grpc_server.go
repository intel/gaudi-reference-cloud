// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudmonitor_logs/api_server/cloudmonitor_logs"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudmonitor_logs/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcService struct {
	listener   net.Listener
	grpcServer *grpc.Server
	errc       chan error
	cfg        config.Config
}

func New(ctx context.Context, cfg *config.Config, listener net.Listener) (*GrpcService, error) {
	if cfg.ListenPort <= 0 {
		return nil, fmt.Errorf("ListenPort must be greater than 0")
	}

	return &GrpcService{
		listener: listener,
		cfg:      *cfg,
	}, nil
}

// Run service, blocking until an error occurs.
func (s *GrpcService) Run(ctx context.Context) error {
	if err := s.Start(ctx); err != nil {
		return err
	}
	// Wait for ListenAndServe to return, return error.
	return <-s.errc
}

func (s *GrpcService) Start(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("GrpcService.Start")
	log.Info("BEGIN", logkeys.ListenAddr, s.listener.Addr().(*net.TCPAddr).Port)
	defer log.Info("END")

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	}

	err := errors.New("")
	s.grpcServer, err = grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return err
	}

	cloudMonitorLogsService, err := cloudmonitor_logs.NewCloudMonitorLogsService(s.cfg)
	if err != nil {
		return err
	}
	// s.CloudMonitorLogsService = cloudMonitorLogsService

	pb.RegisterCloudMonitorLogsServiceServer(s.grpcServer, cloudMonitorLogsService)
	reflection.Register(s.grpcServer)
	log.Info("Service running")
	s.errc = make(chan error, 1)
	go func() {
		s.errc <- s.grpcServer.Serve(s.listener)
		close(s.errc)
	}()

	return nil
}

func (s *GrpcService) Stop(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("GrpcService.Stop")
	log.Info("BEGIN")
	defer log.Info("END")
	// Stop immediately. Do not wait for existing RPCs because streaming RPCs may never end.
	if s != nil && s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	return nil
}
