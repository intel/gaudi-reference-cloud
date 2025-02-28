// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

package server

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/iks/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type GrpcService struct {
	// Format should be ":30002"
	ListenAddr                   string
	ManagedDb                    *manageddb.ManagedDb
	grpcServer                   *grpc.Server
	ComputeInstanceTypeService   v1.InstanceTypeServiceClient
	ComputeInstanceServiceClient v1.InstanceServiceClient
	SshKeyService                v1.SshPublicKeyServiceClient
	VnetClient                   v1.VNetServiceClient
	ProductcatalogServiceClient  v1.ProductCatalogServiceClient
	db                           *sql.DB
	errc                         chan error
	cfg                          config.Config
}

func New(ctx context.Context,
	cfg *config.Config,
	managedDb *manageddb.ManagedDb,
	computeClient v1.InstanceTypeServiceClient,
	sshclient v1.SshPublicKeyServiceClient,
	vnetClient v1.VNetServiceClient,
	productcatalogServiceClient v1.ProductCatalogServiceClient,
	computeInstanceServiceClient v1.InstanceServiceClient) (*GrpcService, error) {
	if cfg.ListenPort <= 0 {
		return nil, fmt.Errorf("ListenPort must be greater than 0")
	}

	return &GrpcService{
		ListenAddr:                   fmt.Sprintf(":%d", cfg.ListenPort),
		ManagedDb:                    managedDb,
		ComputeInstanceTypeService:   computeClient,
		ComputeInstanceServiceClient: computeInstanceServiceClient,
		SshKeyService:                sshclient,
		VnetClient:                   vnetClient,
		ProductcatalogServiceClient:  productcatalogServiceClient,
		cfg:                          *cfg,
	}, nil
}

// Run service, blocking until an error occurs.
func (s *GrpcService) Run(ctx context.Context, managedDbSO *manageddb.ManagedDb) error {
	if err := s.Start(ctx, managedDbSO); err != nil {
		return err
	}
	// Wait for ListenAndServe to return, return error.
	return <-s.errc
}

func (s *GrpcService) Start(ctx context.Context, managedDbSO *manageddb.ManagedDb) error {
	log := log.FromContext(ctx).WithName("GrpcService.Start")
	log.Info("BEGIN", logkeys.ListenAddr, s.ListenAddr)
	defer log.Info("END")

	// get db connection, execute db migrations
	// and other db related boot logic
	err := s.bootstrapDb(ctx, managedDbSO)
	if err != nil {
		return err
	}

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	}

	s.grpcServer, err = grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return err
	}

	iksSrv, err := NewIksService(s.db,
		s.ComputeInstanceTypeService,
		s.SshKeyService,
		s.VnetClient,
		s.ProductcatalogServiceClient,
		s.ComputeInstanceServiceClient,
		s.cfg)
	if err != nil {
		log.Error(err, "error initializing iks service")
		return err
	}
	iksPrivateReconcilerSrv, err := NewIksPrivateReconcilerService(s.db, s.cfg)
	if err != nil {
		log.Error(err, "error initializing service")
		return err
	}
	iksPrivateAdminSrv, err := NewIksPrivateAdminService(s.db, s.cfg, s.ComputeInstanceTypeService)
	if err != nil {
		log.Error(err, "error initializing service")
		return err
	}
	iksSupercomputeSrv, err := NewIksSuperComputeService(s.db, s.ComputeInstanceTypeService, s.SshKeyService, s.VnetClient, s.cfg)
	if err != nil {
		log.Error(err, "error initializing iks service")
		return err
	}

	v1.RegisterIksServer(s.grpcServer, iksSrv)
	v1.RegisterIksPrivateReconcilerServer(s.grpcServer, iksPrivateReconcilerSrv)
	v1.RegisterIksAdminServer(s.grpcServer, iksPrivateAdminSrv)
	v1.RegisterIksSuperComputeServer(s.grpcServer, iksSupercomputeSrv)

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

func (s *GrpcService) bootstrapDb(ctx context.Context, managedDbSO *manageddb.ManagedDb) error {
	l := log.FromContext(ctx).WithName("GrpcService.bootstrapDb")
	l.Info("BEGIN", logkeys.ListenAddr, s.ListenAddr)
	defer l.Info("END")

	if s.ManagedDb == nil {
		return fmt.Errorf("ManagedDb not provided")
	}

	if managedDbSO == nil {
		return fmt.Errorf("ManagedDbSO not provided")
	}

	// generate all the data for the db seed process
	dbSeedTmplData, err := s.cfg.DbSeed.TemplateData()
	if err != nil {
		l.Error(err, "error creating db seed data")
		return err
	}
	// idempotent function for executing db migration
	if err := managedDbSO.Migrate(
		ctx,
		db.NewTemplateFs(db.MigrationsFs, dbSeedTmplData),
		db.MigrationsDir); err != nil {
		l.Error(err, "error running db migration")
		return err
	}
	// open db connection
	s.db, err = s.ManagedDb.Open(ctx)
	if err != nil {
		return err
	}
	return nil
}
