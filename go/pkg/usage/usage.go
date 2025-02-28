// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package usage

import (
	"context"
	"database/sql"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/usage/db"

	//"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/usage/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	usageDb            *sql.DB
	meteringClient     *billingCommon.MeteringClient
	productClient      *billingCommon.ProductClient
	cloudAccountClient *billingCommon.CloudAccountSvcClient
)

func GetUsageDb() *sql.DB {
	return usageDb
}

type Service struct {
	Mdb *manageddb.ManagedDb
	Sql *sql.DB
}

func (svc *Service) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	logger := log.FromContext(ctx).WithName("Usage Service.Init")

	//establish database connection
	if svc.Mdb == nil {
		var err error
		svc.Mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			logger.Error(err, "error connecting to database ")
			return err
		}
	}
	logger.Info("successfully opened database connection")

	// migrate the database
	if err := svc.Mdb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir); err != nil {
		logger.Error(err, "error migrating database")
		return err
	}
	logger.Info("successfully migrated database model")

	var err error
	svc.Sql, err = svc.Mdb.Open(ctx)
	if err != nil {
		return err
	}
	logger.Info("successfully connected to the database")

	usageDb = svc.Sql

	logger.Info("initializing metering client")

	// metering client
	meteringClient, err = billingCommon.NewMeteringClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize metering client")
		return err
	}

	// product catalog client
	productClient, err = billingCommon.NewProductClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize product catalog client")
		return err
	}

	// cloud account client
	cloudAccountClient, err = billingCommon.NewCloudAccountClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize cloud account client")
		return err
	}

	usageData := NewUsageData(usageDb)
	usageRecordData := NewUsageRecordData(usageDb)

	usageService := NewUsageService(cfg, usageDb)
	pb.RegisterUsageServiceServer(grpcServer, usageService)

	usageRecordService := NewUsageRecordService(cfg, usageDb)
	pb.RegisterUsageRecordServiceServer(grpcServer, usageRecordService)

	usageController := NewUsageController(*cloudAccountClient, productClient, meteringClient, usageData, usageRecordData)
	usageScheduler := NewUsageScheduler(ctx, usageController, cfg)
	productUsageScheduler := NewProductUsageScheduler(ctx, usageController, cfg)

	reflection.Register(grpcServer)

	if !cfg.TestProfile {
		//cfg.Features.UsageScheduler = true
		logger.Info("usage scheduler enablement", "UsageScheduler", cfg.Features.UsageScheduler)
		if cfg.Features.UsageScheduler {
			usageScheduler.StartCalculatingUsage(ctx)
		}

		logger.Info("product usage scheduler enablement", "ProductUsageScheduler", cfg.Features.ProductUsageScheduler)
		if cfg.Features.ProductUsageScheduler {
			productUsageScheduler.StartCalculatingProductUsage(ctx)
		}
	}
	return nil
}

func (*Service) Name() string {
	return "usage"
}

func GetDriverDb() *sql.DB {
	return usageDb
}
