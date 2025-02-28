// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/metering/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var MeteringDb *sql.DB

type Service struct {
	Mdb *manageddb.ManagedDb
}

type Config struct {
	ListenPort uint16           `koanf:"listenPort"`
	Database   manageddb.Config `koanf:"database"`
	// productcatalog and billing enabled serviceTypes
	ValidServiceTypes []string `koanf:"validServiceTypes"`
	TestMode          bool
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (svc *Service) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	log.Info("initializing IDC metering service...")

	if svc.Mdb == nil {
		var err error
		svc.Mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			log.Error(err, "error connecting to database ")
			return err
		}
	}

	log.Info("successfully connected to the database")

	if err := db.AutoMigrateDB(ctx, svc.Mdb); err != nil {
		log.Error(err, "error migrating database")
		return err
	}
	log.Info("successfully migrated database model")

	sqlDB, err := svc.Mdb.Open(ctx)
	if err != nil {
		return err
	}

	MeteringDb = sqlDB

	// Create a metrics registry.
	reg := prometheus.NewRegistry()
	// Create some standard server metrics.
	grpcMetrics := grpc_prometheus.NewServerMetrics()
	reg.MustRegister(grpcMetrics)

	meteringSrv, err := NewMeteringService(sqlDB, cfg.ValidServiceTypes)
	if err != nil {
		log.Error(err, "error initializing metering service")
		return err
	}

	v1.RegisterMeteringServiceServer(grpcServer, meteringSrv)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	log.Info("metering service started", "listening on port:", cfg.ListenPort, "valid service types:", cfg.ValidServiceTypes)

	// Disable prometheus in test mode because multiple tests with
	// metering service run in parallel and we get failures due to
	// port conflicts.
	//
	// Need to use a random port if we want to test prometheus with
	// embedded metering server
	if !cfg.TestMode {
		// Create a HTTP server for prometheus.
		httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: fmt.Sprintf("0.0.0.0:%d", 9092)}

		// Start your http server for prometheus.
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				log.Error(err, "Unable to start a http server.")
				os.Exit(1)
			}
		}()
		log.Info("prometheus service started", "listening on port ", 9092)
	}

	return nil
}

func (*Service) Name() string {
	return "metering"
}

func StartMeteringService() {
	ctx := context.Background()

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &Service{}, &Config{})
	if err != nil {
		logger := log.FromContext(ctx).WithName("StartMeteringService")
		logger.Error(err, "init err")
	}
}
