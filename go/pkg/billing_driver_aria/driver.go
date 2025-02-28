// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"database/sql"
	"fmt"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/sqsutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var sched *Scheduler
var billingDb *sql.DB
var usageConn *grpc.ClientConn
var meteringClientConn *grpc.ClientConn
var cloudAccountConn *grpc.ClientConn
var productClient *billingCommon.ProductClient

type Service struct {
	meteringClient     pb.MeteringServiceClient
	cloudAccountClient pb.CloudAccountServiceClient
	usageServiceClient pb.UsageServiceClient
	Sql                *sql.DB
	Mdb                *manageddb.ManagedDb
	testDB             manageddb.TestDb
}

var AriaService Service

func (svc *Service) Init(ctx context.Context, cfg *config.Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("Service.Init").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	config.Cfg = cfg
	if config.Cfg.ClientIdPrefix == "" {
		return fmt.Errorf("must specify clientIdPrefix in config. For testing, use your IDSID")
	}
	var err error
	if cfg.TestProfile {
		svc.Mdb, err = svc.testDB.Start(ctx)
		if err != nil {
			return fmt.Errorf("testDb.Start: %m", err)
		}
	}

	if svc.Mdb == nil {
		var err error
		svc.Mdb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			logger.Error(err, "error connecting to database ")
			return err
		}
	}

	svc.Sql, err = svc.Mdb.Open(ctx)
	if err != nil {
		return err
	}
	logger.Info("successfully connected to the database")
	billingDb = svc.Sql

	// Product client
	productClient, err = billingCommon.NewProductClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize product client")
		return err
	}

	usageClient, err := billingCommon.NewMeteringClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize usages client")
		return err
	}

	addr, err := resolver.Resolve(ctx, "metering")
	if err != nil {
		return err
	}
	meteringClientConn, err = grpcutil.NewClient(ctx, addr)
	if err != nil {
		return err
	}
	svc.meteringClient = pb.NewMeteringServiceClient(meteringClientConn)

	addr, err = resolver.Resolve(ctx, "cloudaccount")
	if err != nil {
		return err
	}
	cloudAccountConn, err = grpcutil.NewClient(ctx, addr)
	if err != nil {
		return err
	}
	svc.cloudAccountClient = pb.NewCloudAccountServiceClient(cloudAccountConn)

	addr, err = resolver.Resolve(ctx, "usage")
	if err != nil {
		return err
	}
	usageConn, err = grpcutil.NewClient(ctx, addr)
	if err != nil {
		return err
	}
	svc.usageServiceClient = pb.NewUsageServiceClient(usageConn)

	// Aria client
	// The values for configuring the Aria admin client will be loaded from configuration.
	ariaAdminClient := client.NewAriaAdminClient(config.Cfg.GetAriaSystemServerUrlAdminToolsApi(), config.Cfg.GetAriaSystemInsecureSsl()) // ssl Aria flag
	// The values for configuring the Aria client will be loaded from configuration.
	ariaClient := client.NewAriaClient(config.Cfg.GetAriaSystemServerUrlCoreApi(), config.Cfg.GetAriaSystemCoreApiSuffix(), config.Cfg.GetAriaSystemInsecureSsl()) // ssl Aria flag
	// Credentials will come from the secrets service.
	ariaCredentials := client.NewAriaCredentials(config.Cfg.GetAriaSystemClientNo(), config.Cfg.GetAriaSystemAuthKey())

	ariaController := NewAriaController(ariaClient, ariaAdminClient, ariaCredentials)
	productController := NewProductController(ariaClient, ariaAdminClient, ariaCredentials, productClient, ariaController)

	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(svc.Sql)
	eventManager := events.NewEventManager(eventData, eventDispatcher)

	pb.RegisterBillingAccountServiceServer(grpcServer, &AriaBillingAccountService{ariaController: ariaController})
	pb.RegisterBillingOptionServiceServer(grpcServer, &AriaBillingOptionService{ariaController: ariaController})
	pb.RegisterBillingRateServiceServer(grpcServer, &AriaBillingRateService{})
	pb.RegisterBillingCreditServiceServer(grpcServer, &AriaBillingCreditService{ariaController: ariaController})
	invoiceController := NewInvoiceController(ariaClient, ariaCredentials, svc.cloudAccountClient)
	pb.RegisterBillingInvoiceServiceServer(grpcServer, &AriaBillingInvoiceService{invoiceController: invoiceController})
	pb.RegisterBillingProductCatalogSyncServiceServer(grpcServer, &AriaBillingProductCatalogSyncService{})
	pb.RegisterPaymentServiceServer(grpcServer, &AriaPaymentService{ariaController: ariaController})
	pb.RegisterBillingInstancesServiceServer(grpcServer, &AriaBillingInstancesService{meteringServiceClient: usageClient, productServiceClient: productClient})
	ariaAccountClient := client.NewAriaAccountClient(ariaClient, ariaCredentials)
	ariaServiceCreditClient := client.NewServiceCreditClient(ariaClient, ariaCredentials)
	ariaUsageClient := client.NewAriaUsageClient(ariaClient, ariaCredentials)
	reflection.Register(grpcServer)
	enterpriseAcctLinkScheduler := NewEnterpriseAcctLinkScheduler(ctx, eventManager, svc.cloudAccountClient, ariaAccountClient, ariaServiceCreditClient)
	sched = NewScheduler(productController)
	// do not start the schedulers in testing mode.
	if !cfg.TestProfile {
		StartEnterpriseAcctLinkScheduler(ctx, enterpriseAcctLinkScheduler)
		sched.StartSync(ctx)
	}

	// Aria PaidServicesDeactivationController
	logger.Info("features deactivation scheduler", "DeactivationScheduler", cfg.GetFeaturesDeactivationScheduler())
	if cfg.GetFeaturesDeactivationScheduler() {
		ariaSQSUtil := sqsutil.SQSUtil{}
		awsQueueURL := config.Cfg.GetAWSSQSQueueUrl()
		awsQueueRegion := config.Cfg.GetAWSSQSRegion()
		awsCredentialsFile := config.Cfg.GetAWSSQSCredentialsFile()
		logger.Info("AWS SQS", "awsQueueURL", awsQueueURL, "awsQueueRegion", awsQueueRegion, "awsCredentialsFile", awsCredentialsFile)
		if err := ariaSQSUtil.Init(ctx, awsQueueRegion, awsQueueURL, awsCredentialsFile); err != nil {
			logger.Error(err, "couldn't init sqsProcessor")
			return err
		}
		paidServicesDeactivationController := NewPaidServicesDeactivationController(&ariaSQSUtil, eventManager, svc.cloudAccountClient)
		if !cfg.TestProfile {
			StartPaidServicesDeactivationController(ctx, paidServicesDeactivationController)
			sched.StartSync(ctx)
		}
		logger.Info("paidServicesDeactivationController", "created", paidServicesDeactivationController)
	}

	if !cfg.TestProfile && cfg.BillingDriverAria.Features.ReportUsageScheduler {
		logger.Info("report resource usage scheduler starting")
		usageControler := NewUsageController(ariaCredentials, svc.cloudAccountClient,
			svc.usageServiceClient, ariaUsageClient, ariaAccountClient)
		StartReportUsageScheduler(ctx, usageControler)
	}
	if !cfg.TestProfile && cfg.BillingDriverAria.Features.ReportProductUsageScheduler {
		logger.Info("report product usage scheduler starting")
		usageControler := NewProductUsageController(ariaCredentials, svc.cloudAccountClient,
			svc.usageServiceClient, ariaUsageClient, ariaAccountClient)
		StartProductUsageReportScheduler(ctx, usageControler)
	}
	return nil
}

func (*Service) Name() string {
	return "billing-aria"
}

func (*Service) Done() {
	sched.StopSync()
}
