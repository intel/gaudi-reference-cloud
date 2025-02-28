// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcredits

import (
	"context"
	"database/sql"
	"embed"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	db                 *sql.DB
	cloudacctClient    pb.CloudAccountServiceClient
	usageServiceClient pb.UsageServiceClient
	conf               *config.Config
)
var (
	cloudAccountLocks *CloudAccountLocks
)

type Service struct {
	Mdb                   *manageddb.ManagedDb
	cloudAccountSvcClient *billingCommon.CloudAccountSvcClient
	usageServiceClient    *billingCommon.UsageSvcClient
}

func grpcConnect(ctx context.Context, addr string) (*grpc.ClientConn, error) {

	dialOptions := []grpc.DialOption{
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}
	conn, err := grpcutil.NewClient(ctx, addr, dialOptions...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (svc *Service) Init(ctx context.Context, cfg *config.Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("Service.Init").Start()
	defer span.End()
	logger.Info("service initialization")
	addr, err := resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		logger.Error(err, "failed to resolve cloud account service")
		return err
	}
	conn, err := grpcConnect(ctx, addr)
	if err != nil {
		logger.Error(err, "failed to connect to cloud account service")
		return err
	}
	cloudacctClient = pb.NewCloudAccountServiceClient(conn)

	config.Cfg = cfg
	if svc.Mdb == nil {
		var err error
		svc.Mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			logger.Error(err, "error connecting to database ")
			return err
		}
	}
	if err := openDb(ctx, svc.Mdb); err != nil {
		logger.Error(err, "error connecting to database ")
		return err
	}

	// cloud account client
	cloudAccountClient, err := billingCommon.NewCloudAccountClient(ctx, resolver)
	svc.cloudAccountSvcClient = cloudAccountClient
	if err != nil {
		logger.Error(err, "failed to initialize cloud account client")
		return err
	}

	usageServiceClient, err := billingCommon.NewUsageServiceClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize usage service client")
		return err
	}
	svc.usageServiceClient = usageServiceClient

	if !cfg.TestProfile && cfg.RunCreditEventSchedulers {
		notificationGatewayClient, err := billingCommon.NewNotificationGatewayClient(ctx, resolver)
		if err != nil {
			logger.Error(err, "failed to initialize notification client")
			return err
		}
		schedulerCloudAccountState := &SchedulerCloudAccountState{
			AccessTimestamp: "",
		}

		logger.Info("credit usage scheduler enabled", "CreditUsageScheduler", cfg.Features.CreditUsageEventScheduler)
		if cfg.Features.CreditUsageEventScheduler {
			cloudCreditUsageScheduler := NewCloudCreditUsageEventScheduler(notificationGatewayClient, schedulerCloudAccountState, svc.cloudAccountSvcClient)
			startCloudCreditUsageEventScheduler(ctx, *cloudCreditUsageScheduler)
		}
		logger.Info("credit expiry scheduler enabled", "CreditExpiryScheduler", cfg.Features.CreditExpiryEventScheduler)
		if cfg.Features.CreditExpiryEventScheduler {
			cloudCreditExpiryScheduler := NewCloudCreditExpiryEventScheduler(notificationGatewayClient, schedulerCloudAccountState, svc.cloudAccountSvcClient)
			startCloudCreditExpiryEventScheduler(ctx, *cloudCreditExpiryScheduler)
		}

		logger.Info("credit install scheduler enabled", "CreditInstallScheduler", cfg.Features.CreditUsageReportScheduler)
		if cfg.Features.CreditUsageReportScheduler {
			cloudCreditUsageReport := NewCreditUsageReportScheduler(db, svc.cloudAccountSvcClient, usageServiceClient)
			StartCloudCreditUsageReportScheduler(ctx, *cloudCreditUsageReport)
		}
	}
	cloudacctClient = svc.cloudAccountSvcClient.CloudAccountClient
	if err := billingCommon.InitDrivers(ctx, cloudacctClient, resolver); err != nil {
		logger.Error(err, "failed to initialize drivers")
		return err
	}

	cloudAccountLocks = NewCloudAccountLocks(ctx)
	pb.RegisterCloudCreditsCouponServiceServer(grpcServer, &CloudCreditsCouponService{})
	pb.RegisterCloudCreditsCreditServiceServer(grpcServer, &CloudCreditsCreditService{})
	reflection.Register(grpcServer)
	conf = cfg
	return nil
}

func (*Service) Name() string {
	return "cloudcredits"
}

//go:embed sql/*.sql
var fs embed.FS

func openDb(ctx context.Context, mdb *manageddb.ManagedDb) error {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("CloudCreditsCreditService.openDb").Start()
	defer span.End()
	if err := mdb.Migrate(ctx, fs, "sql"); err != nil {
		log.Error(err, "migrate:")
		return err
	}
	var err error
	db, err = mdb.Open(ctx)
	if err != nil {
		log.Error(err, "mdb.Open failed")
		return err
	}
	return nil
}
