// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin_ui_server/api_server/config"
	fleet_admin_ui_server "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin_ui_server/api_server/fleet_admin_ui_server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type GrpcService struct {
	ManagedDb           *manageddb.ManagedDb
	FleetAdminUIService *fleet_admin_ui_server.FleetAdminUIService
	listener            net.Listener
	grpcServer          *grpc.Server
	db                  *sql.DB
	errc                chan error
	cfg                 config.Config
}

func New(ctx context.Context, cfg *config.Config, managedDb *manageddb.ManagedDb, listener net.Listener) (*GrpcService, error) {
	return &GrpcService{
		ManagedDb: managedDb,
		listener:  listener,
		cfg:       *cfg,
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
	if s.ManagedDb == nil {
		return fmt.Errorf("ManagedDb not provided")
	}
	//TODO: Uncomment once the multi-service-single-db solution becomes available.
	// if err := s.ManagedDb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir); err != nil {
	// 	return err
	// }
	db, err := s.ManagedDb.Open(ctx)
	if err != nil {
		return err
	}
	s.db = db

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    120 * time.Second, // The time a connection is kept alive without any activity.
			Timeout: 20 * time.Second,  // Maximum time the server waits for activity before closing the connection.
		}),
	}

	s.grpcServer, err = grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return err
	}

	fleetAdminUIService, err := fleet_admin_ui_server.NewFleetAdminUIService(db, s.cfg)
	if err != nil {
		return err
	}
	s.FleetAdminUIService = fleetAdminUIService

	pb.RegisterFleetAdminUIServiceServer(s.grpcServer, fleetAdminUIService)
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
