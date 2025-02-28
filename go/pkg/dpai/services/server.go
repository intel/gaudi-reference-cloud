// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package services

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db"
	models "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/db/models"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"

	// "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils/k8s"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

type DpaiServer struct {
	Conn     *pgxpool.Conn
	SqlPool  *pgxpool.Pool
	Sql      *models.Queries
	SqlModel *models.Queries
	// K8sClient *k8s.K8sClient
	Config         *config.Config
	IksClient      *pb.IksClient
	IKSAzClientSet *k8s.K8sAzClient
	// pb.UnimplementedAdminWorkspaceSizeServiceServer
	pb.UnimplementedDpaiDeploymentServiceServer
	pb.UnimplementedDpaiDeploymentTaskServiceServer
	pb.UnimplementedDpaiWorkspaceServiceServer
	pb.UnimplementedDpaiPostgresSizeServiceServer
	pb.UnimplementedDpaiPostgresVersionServiceServer
	pb.UnimplementedDpaiPostgresServiceServer
	pb.UnimplementedDpaiAirflowServiceServer
	pb.UnimplementedDpaiAirflowConfServiceServer
	pb.UnimplementedDpaiAirflowSizeServiceServer
	pb.UnimplementedDpaiAirflowVersionServiceServer
	// pb.UnimplementedDpaiHmsConfGroupServiceServer
	// pb.UnimplementedDpaiHmsConfServiceServer
	// pb.UnimplementedDpaiHmsSizeServiceServer
	// pb.UnimplementedDpaiHmsVersionServiceServer
	// pb.UnimplementedDpaiHmsServiceServer
}

func NewDpaiServer() *DpaiServer {
	return &DpaiServer{}
}

type GrpcService struct {
	// Format should be ":30002"
	ctx        context.Context
	ListenAddr string
	ManagedDb  *manageddb.ManagedDb
	grpcServer *grpc.Server
	db         *sql.DB
	errc       chan error
	cfg        config.Config

	SqlConn  *pgxpool.Conn
	SqlPool  *pgxpool.Pool
	SqlModel *models.Queries
	// init the Az clientset for the IKS control plane environemtnt to interact with IKS operators
	// K8sClient *k8s.K8sClient
	// IksClient      *pb.IksClient
	// GrpcClientConn *grpc.ClientConn
	Config *config.Config
}

func New(ctx context.Context, cfg *config.Config, managedDb *manageddb.ManagedDb) (*GrpcService, error) {
	if cfg.ListenPort <= 0 {
		return nil, fmt.Errorf("ListenPort must be greater than 0")
	}
	log := log.FromContext(ctx).WithName("GrpcService.New")
	// Initialize your DB Connection
	pool, err := db.InitDB(cfg)
	if err != nil {
		log.Error(err, "Failed to initialize database.")
		return nil, err
	}
	// defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Error(err, "Failed to get postgres connection.")
		return nil, err
	}

	return &GrpcService{
		ListenAddr: fmt.Sprintf(":%d", cfg.ListenPort),
		ManagedDb:  managedDb,
		cfg:        *cfg,
		SqlModel:   models.New(pool),
		SqlPool:    pool,
		SqlConn:    conn,
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
	log.Info("BEGIN", "ListenAddr", s.ListenAddr)
	defer log.Info("END")
	if s.ManagedDb == nil {
		return fmt.Errorf("ManagedDb not provided")
	}

	if err := s.ManagedDb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir); err != nil {
		return err
	}

	// mdb, err := s.ManagedDb.Open(ctx)
	// if err != nil {
	// 	return err
	// }
	// s.db = mdb

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	}

	var err error
	s.grpcServer, err = grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return err
	}
	var K8sAzClient k8s.K8sAzClient = k8s.K8sAzClient{}
	azK8sClient, err := K8sAzClient.ConfigureK8sClient()
	if err != nil {
		return err
	}
	var dpai *DpaiServer = NewDpaiServer()

	dpai.Config = &s.cfg
	dpai.SqlPool = s.SqlPool
	dpai.Sql = s.SqlModel
	dpai.SqlModel = s.SqlModel
	dpai.IKSAzClientSet = azK8sClient
	// dpai.IksClient = s.IksClient

	// register the services to the server
	// pb.RegisterAdminWorkspaceSizeServiceServer(s, server)
	pb.RegisterDpaiDeploymentServiceServer(s.grpcServer, dpai)
	pb.RegisterDpaiDeploymentTaskServiceServer(s.grpcServer, dpai)
	pb.RegisterDpaiWorkspaceServiceServer(s.grpcServer, dpai)
	pb.RegisterDpaiAirflowServiceServer(s.grpcServer, dpai)
	pb.RegisterDpaiAirflowConfServiceServer(s.grpcServer, dpai)
	pb.RegisterDpaiAirflowSizeServiceServer(s.grpcServer, dpai)
	pb.RegisterDpaiAirflowVersionServiceServer(s.grpcServer, dpai)
	pb.RegisterDpaiPostgresSizeServiceServer(s.grpcServer, dpai)
	pb.RegisterDpaiPostgresVersionServiceServer(s.grpcServer, dpai)
	pb.RegisterDpaiPostgresServiceServer(s.grpcServer, dpai)
	// pb.RegisterDpaiHmsConfGroupServiceServer(s, server)
	// pb.RegisterDpaiHmsConfServiceServer(s, server)
	// pb.RegisterDpaiHmsSizeServiceServer(s, server)
	// pb.RegisterDpaiHmsVersionServiceServer(s, server)
	// pb.RegisterDpaiHmsServiceServer(s, server)

	reflection.Register(s.grpcServer)
	listener, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}
	log.Info("Service running")
	s.errc = make(chan error, 1)
	go func() {
		s.errc <- s.grpcServer.Serve(listener)
		defer s.SqlPool.Close()
		defer s.SqlConn.Release()
		// defer s.GrpcClientConn.Close()
		close(s.errc)
	}()
	fmt.Println("Closed the start function")
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
