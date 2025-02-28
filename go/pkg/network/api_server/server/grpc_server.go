// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/address_translation"
	"net"
	"net/http"
	"os"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/global_operations"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/iprm"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/subnet"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/api_server/internal/vpc"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/db"
	sdnv1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/network/sdn"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type GrpcService struct {
	ManagedDb                 *manageddb.ManagedDb
	VPCService                *vpc.VPCService
	SubnetService             *subnet.SubnetService
	IPRMService               *iprm.IPRMService
	GlobalOperationsService   *global_operations.GlobalOperationsService
	AddressTranslationService *address_translation.AddressTranslationPrivateService
	listener                  net.Listener
	grpcServer                *grpc.Server
	db                        *sql.DB
	errc                      chan error
	cfg                       config.Config
	cloudAccountServiceClient pb.CloudAccountServiceClient
	sdnClient                 sdnv1.OvnnetClient
	availabilityZones         []string
	prometheusPort            int
}

func New(
	ctx context.Context,
	cfg *config.Config,
	managedDb *manageddb.ManagedDb,
	listener net.Listener,
	cloudAccountServiceClient pb.CloudAccountServiceClient,
	availabilityZones []string,
) (*GrpcService, error) {
	return &GrpcService{
		ManagedDb:                 managedDb,
		listener:                  listener,
		cfg:                       *cfg,
		cloudAccountServiceClient: cloudAccountServiceClient,
		availabilityZones:         availabilityZones,
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
	if err := s.ManagedDb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir); err != nil {
		return err
	}
	db, err := s.ManagedDb.Open(ctx)
	if err != nil {
		return err
	}
	s.db = db

	// Create a metrics registry.
	reg := prometheus.NewRegistry()
	// Create some standard server metrics.
	srvMetrics := grpc_prometheus.NewServerMetrics()
	reg.MustRegister(srvMetrics)

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.ChainUnaryInterceptor(srvMetrics.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(srvMetrics.StreamServerInterceptor()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    120 * time.Second, // The time a connection is kept alive without any activity.
			Timeout: 20 * time.Second,  // Maximum time the server waits for activity before closing the connection.
		}),
	}

	s.grpcServer, err = grpcutil.NewServer(ctx, serverOptions...)
	if err != nil {
		return err
	}

	vpcService, err := vpc.NewVPCService(db, s.cfg, s.cloudAccountServiceClient)
	if err != nil {
		return err
	}
	s.VPCService = vpcService

	subnetService, err := subnet.NewSubnetService(db, s.cfg, s.cloudAccountServiceClient, vpcService, s.availabilityZones)
	if err != nil {
		return err
	}
	s.SubnetService = subnetService

	iprmService, err := iprm.NewIPRMService(db, s.cfg, s.cloudAccountServiceClient, subnetService)
	if err != nil {
		return err
	}
	s.IPRMService = iprmService

	GlobalOperationsService, err := global_operations.NewGlobalOperationsService(vpcService, subnetService, s.availabilityZones)
	if err != nil {
		return err
	}
	s.GlobalOperationsService = GlobalOperationsService

	addressTranslationService, err := address_translation.NewAddressTranslationPrivateService(db, s.cfg, s.cloudAccountServiceClient)
	if err != nil {
		return err
	}
	s.AddressTranslationService = addressTranslationService

	pb.RegisterVPCServiceServer(s.grpcServer, vpcService)
	pb.RegisterVPCPrivateServiceServer(s.grpcServer, vpcService)
	pb.RegisterSubnetServiceServer(s.grpcServer, subnetService)
	pb.RegisterSubnetPrivateServiceServer(s.grpcServer, subnetService)
	pb.RegisterIPRMPrivateServiceServer(s.grpcServer, iprmService)

	pb.RegisterGlobalOperationsServiceServer(s.grpcServer, GlobalOperationsService)
	pb.RegisterAddressTranslationPrivateServiceServer(s.grpcServer, addressTranslationService)

	srvMetrics.InitializeMetrics(s.grpcServer)
	reflection.Register(s.grpcServer)
	log.Info("Service running")
	s.errc = make(chan error, 1)
	go func() {
		s.errc <- s.grpcServer.Serve(s.listener)
		close(s.errc)
	}()

	if s.cfg.PrometheusListenPort != 0 {
		httpSrv := &http.Server{Addr: fmt.Sprintf("0.0.0.0:%d", s.cfg.PrometheusListenPort)}
		m := http.NewServeMux()
		// Create HTTP handler for Prometheus metrics.
		m.Handle("/metrics", promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{
				// Opt into OpenMetrics e.g. to support exemplars.
				EnableOpenMetrics: true,
			},
		))

		httpSrv.Handler = m
		log.Info("starting HTTP server", "addr", httpSrv.Addr)
		// Start your http server for prometheus.
		go func() {
			if err := httpSrv.ListenAndServe(); err != nil {
				log.Error(err, "Unable to start http server.")
				os.Exit(1)
			}
		}()
		log.Info("prometheus service started", logkeys.ListenPort, s.cfg.PrometheusListenPort)
	}
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
