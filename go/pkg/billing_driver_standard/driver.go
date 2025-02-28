// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import (
	"context"
	"database/sql"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var standardDriverDb *sql.DB

type Service struct {
	cloudAccountClient *billingCommon.CloudAccountSvcClient
	usageClient        *billingCommon.MeteringClient
	productClient      *billingCommon.ProductClient
	Mdb                *manageddb.ManagedDb
	Sql                *sql.DB
}

var StandardService *Service

func (svc *Service) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {

	logger := log.FromContext(ctx).WithName("BillingStandardDriver Service.Init")
	StandardService = svc
	//database connection
	if svc.Mdb == nil {
		var err error
		svc.Mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			logger.Error(err, "error connecting to database ")
			return err
		}
	}
	var err error
	svc.Sql, err = svc.Mdb.Open(ctx)
	if err != nil {
		return err
	}
	logger.Info("successfully connected to the database")

	standardDriverDb = svc.Sql

	logger.Info("initializing metering usage client")

	// metering usage client
	svc.usageClient, err = billingCommon.NewMeteringClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize metering usage client")
		return err
	}

	// product catalog client
	svc.productClient, err = billingCommon.NewProductClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize product catalog client")
		return err
	}

	// cloud account client
	svc.cloudAccountClient, err = billingCommon.NewCloudAccountClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize cloud account client")
		return err
	}

	pb.RegisterBillingAccountServiceServer(grpcServer, &StandardBillingAccountService{})
	pb.RegisterBillingOptionServiceServer(grpcServer, &StandardBillingOptionService{})
	pb.RegisterBillingRateServiceServer(grpcServer, &StandardBillingRateService{})
	pb.RegisterBillingCreditServiceServer(grpcServer, &StandardBillingCreditService{session: svc.Sql})
	pb.RegisterBillingInstancesServiceServer(grpcServer, &StandardBillingInstancesService{
		meteringServiceClient: svc.usageClient,
		productServiceClient:  svc.productClient,
		config:                cfg,
	})
	pb.RegisterBillingProductCatalogSyncServiceServer(grpcServer, &StandardBillingProductCatalogSyncService{})
	reflection.Register(grpcServer)
	return nil
}

func (*Service) Name() string {
	return "billing-standard"
}

func GetDriverDb() *sql.DB {
	return standardDriverDb
}
