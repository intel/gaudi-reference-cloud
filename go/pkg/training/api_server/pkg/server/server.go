// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	v1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	batchSvc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/api_server/pkg/batch_service"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/api_server/pkg/db"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/config"
	idcComputeSvc "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/training/idc_compute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	// get armada feature flag value
	armadaEnabled, _ = strconv.ParseBool(os.Getenv("ARMADA_ENABLED"))
)

type Service struct {
	ComputeClient *idcComputeSvc.IDCServiceClient
	Mdb           *manageddb.ManagedDb
}

type Config struct {
	ListenPort  uint16           `koanf:"listenPort"`
	Database    manageddb.Config `koanf:"database"`
	TestMode    bool
	CloudConfig config.BatchServiceConfig `koanf:"batchServiceConfig"`
}

func (config *Config) GetListenPort() uint16 {
	return config.ListenPort
}

func (config *Config) SetListenPort(port uint16) {
	config.ListenPort = port
}

func (svc *Service) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	logger := log.FromContext(ctx)

	logger.Info("initializing IDC training batch service...", "config", cfg)

	if svc.Mdb == nil {
		var err error
		svc.Mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			return err
		}
	}

	logger.Info("successfully connected to the database")

	if err := db.AutoMigrateDB(ctx, svc.Mdb); err != nil {
		logger.Error(err, "error migrating database")
		return err
	}

	logger.Info("successfully migrated database model")

	// create client for compute-api-server access in regional(local) cluster
	computeClient, err := idcComputeSvc.NewIDCComputeServiceClient(ctx, &cfg.CloudConfig.IDC)
	if err != nil {
		logger.Error(err, "failed to initialize compute client")
		return fmt.Errorf("error connecting to compute client")
	}

	sqlDB, err := svc.Mdb.Open(ctx)
	if err != nil {
		return err
	}

	// create client for productcatalog access in global cluster
	productClient, err := batchSvc.NewProductClient(ctx, resolver)
	if err != nil {
		return err
	}

	// Ensure that we can ping the Product Catalog service before starting.
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := productClient.ProductCatalogClient.Ping(pingCtx, &emptypb.Empty{}); err != nil {
		logger.Error(err, "batch service: unable to ping Product Catalog service")
		return err
	}
	logger.Info("batch service: Ping to Product Catalog service was successful")

	// create client for cloudaccount access in global cluster
	cloudAccountClient, err := batchSvc.NewCloudAccountClient(ctx, resolver)
	if err != nil {
		return err
	}

	// Ensure that we can ping the Cloud Account service before starting.
	pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := cloudAccountClient.CloudAccountClient.Ping(pingCtx, &emptypb.Empty{}); err != nil {
		logger.Error(err, "batch service: unable to ping Cloud Account service")
		return err
	}
	logger.Info("batch service: Ping to Cloud Account service was successful")

	// create client for billing access in global cluster
	billingClient, err := batchSvc.NewBillingClient(ctx, resolver)
	if err != nil {
		return err
	}

	// Ensure that we can ping the Billing Service service before starting.
	// pingCtx, cancel = context.WithTimeout(ctx, time.Second*10)
	// defer cancel()
	// if _, err := billingClient.BillingCreditServiceClient..Ping(pingCtx, &emptypb.Empty{}); err != nil {
	// 	logger.Error(err, "batch service: unable to ping Billing service")
	// 	return err
	// }
	// logger.Info("batch service: Ping to Billing service was successful")

	trainingBatchSvc, err := batchSvc.NewTrainingBatchUserService(computeClient, cfg.CloudConfig.SlurmBatchServiceEndpoint, cfg.CloudConfig.SlurmJupyterhubServiceEndpoint, cfg.CloudConfig.SlurmSSHServiceEndpoint, sqlDB, productClient, cloudAccountClient, billingClient)
	if err != nil {
		return err
	}

	// Create the new armada service client connecting it to the training database for reconciliation
	trainingClusterSvc, err := NewTrainingClusterService(sqlDB)
	if err != nil {
		return err
	}

	v1.RegisterTrainingBatchUserServiceServer(grpcServer, trainingBatchSvc)

	if armadaEnabled {
		v1.RegisterTrainingClusterServiceServer(grpcServer, trainingClusterSvc)
	}

	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	logger.Info("training api service started", "listening on port ", cfg.ListenPort)

	return nil
}

func (*Service) Name() string {
	return "training-api-service"
}

func StartTrainingAPIService() {
	ctx := context.Background()
	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	err := grpcutil.Run[*Config](ctx, &Service{}, &Config{})
	if err != nil {
		logger := log.FromContext(ctx).WithName("StartTrainingAPIServices")
		logger.Error(err, "init err")
	}
}
