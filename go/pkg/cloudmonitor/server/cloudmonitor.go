// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"google.golang.org/grpc"
	"net"
	"net/http"
	// "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	// "google.golang.org/grpc"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudmonitor/config"
	"google.golang.org/grpc/reflection"
)

type GrpcService struct {
	// Format should be ":30002"
	ListenAddr      string
	ManagedDb       *manageddb.ManagedDb
	grpcServer      *grpc.Server
	db              *sql.DB
	errc            chan error
	cfg             config.Config
	InstanceService v1.InstanceServiceClient
}

func New(ctx context.Context, cfg *config.Config, managedDb *manageddb.ManagedDb) (*GrpcService, error) {
	if cfg.ListenPort <= 0 {
		return nil, fmt.Errorf("ListenPort must be greater than 0")
	}

	return &GrpcService{
		ListenAddr: fmt.Sprintf(":%d", cfg.ListenPort),
		ManagedDb:  managedDb,
		cfg:        *cfg,
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

var db *sql.DB

//go:embed sql/*.sql
var fs embed.FS

func (s *GrpcService) Start(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("GrpcService.Start")
	log.Info("BEGIN", "ListenAddr", s.ListenAddr)
	defer log.Info("END")
	if s.ManagedDb == nil {
		return fmt.Errorf("Managed Db not provided")
	}

	if err := s.ManagedDb.Migrate(ctx, fs, "sql"); err != nil {
		return err
	}

	db, err := s.ManagedDb.Open(ctx)
	if err != nil {
		return err
	}
	s.db = db

	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor(), obs.CustomUnaryServerInterceptor("CloudAccountId"), grpc_prometheus.UnaryServerInterceptor),
		grpc.ChainStreamInterceptor(otelgrpc.StreamServerInterceptor(), grpc_prometheus.StreamServerInterceptor),
	}

	s.grpcServer, err = grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return err
	}
	grpc_prometheus.Register(s.grpcServer)
	lensSrv, err := NewCloudMonitorService(db, s.cfg)
	if err != nil {
		log.Error(err, "error initializing metering service")
		return err
	}
	v1.RegisterCloudMonitorServiceServer(s.grpcServer, lensSrv)

	reflection.Register(s.grpcServer)
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}
	log.Info("Service running")

	//httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: fmt.Sprintf("0.0.0.0:%d", 9090)}

	s.errc = make(chan error, 1)
	go func() {
		// http.Handle("/metrics", promhttp.Handler())
		// log.Info("Prometheus metrics available at :9090/metrics")
		// if err := http.ListenAndServe(":9090", nil); err != nil {
		// 	log.Error(err, "Failed to start metrics HTTP server:")
		// }
		s.errc <- s.grpcServer.Serve(listener)
		close(s.errc)

	}()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Info("Prometheus metrics available at :9090/metrics")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Error(err, "Failed to start metrics HTTP server:")
		}
		// s.errc <- s.grpcServer.Serve(listener)
		// close(s.errc)

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
