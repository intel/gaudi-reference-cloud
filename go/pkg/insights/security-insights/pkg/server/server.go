// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/security-insights/pkg/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Service struct {
	Mdb *manageddb.ManagedDb
}

type Config struct {
	ListenPort uint16           `koanf:"listenPort"`
	Database   manageddb.Config `koanf:"database"`
	GithubKey  string           `koanf:"githubKey"`
	TestMode   bool
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

	log.Info("initializing IDC security insights service...")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

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

	insightsSrv, err := NewSecurityInsightsService(sqlDB)
	if err != nil {
		log.Error(err, "error initializing metering service")
		return err
	}

	v1.RegisterSecurityInsightsServer(grpcServer, insightsSrv)

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	log.Info("security insights service started", "listening on port ", cfg.ListenPort)

	return nil
}

func (*Service) Name() string {
	return "security-insights"
}

func StartSecurityInsightsService() {
	ctx := context.Background()
	err := grpcutil.Run[*Config](ctx, &Service{}, &Config{})
	if err != nil {
		logger := log.FromContext(ctx).WithName("StartSecurityInsightsService")
		logger.Error(err, "init err")
	}
}
