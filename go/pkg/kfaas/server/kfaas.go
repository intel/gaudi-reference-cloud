// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"

	// "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kfaas/config"
	"google.golang.org/grpc/reflection"
)

type GrpcService struct {
	// Format should be ":30002"
	ListenAddr string
	ManagedDb  *manageddb.ManagedDb
	grpcServer *grpc.Server
	db         *sql.DB
	errc       chan error
	cfg        config.Config
	IKSService v1.IksClient
}

func New(ctx context.Context, cfg *config.Config, managedDb *manageddb.ManagedDb, iksClient v1.IksClient) (*GrpcService, error) {
	if cfg.ListenPort <= 0 {
		return nil, fmt.Errorf("ListenPort must be greater than 0")
	}

	return &GrpcService{
		ListenAddr: fmt.Sprintf(":%d", cfg.ListenPort),
		ManagedDb:  managedDb,
		IKSService: iksClient,
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
	log.Info("BEGIN", logkeys.ListenAddr, s.ListenAddr)
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
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	}

	s.grpcServer, err = grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return err
	}

	kfaasSrv, err := NewKFService(db, s.IKSService, s.cfg)
	if err != nil {
		log.Error(err, "error initializing metering service")
		return err
	}
	v1.RegisterKFServiceServer(s.grpcServer, kfaasSrv)

	reflection.Register(s.grpcServer)
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}
	log.Info("Service running")
	s.errc = make(chan error, 1)
	go func() {
		s.errc <- s.grpcServer.Serve(listener)
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
