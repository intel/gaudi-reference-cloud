// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"database/sql"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	events "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/notification_gateway"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var creditsInstallSched *CreditsInstallScheduler
var billingDb *sql.DB
var (
	cloudAccountLocks *CloudAccountLocks
)

type Service struct {
	ManagedDb             *manageddb.ManagedDb
	Sql                   *sql.DB
	CloudAccountSvcClient *billingCommon.CloudAccountSvcClient
	UsageServiceClient    pb.UsageServiceClient
	MeteringServiceClient *billingCommon.MeteringClient
}

func (svc *Service) Init(ctx context.Context, cfg *Config, resolver grpcutil.Resolver, grpcServer *grpc.Server) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("BillingService.Init")
	Cfg = cfg
	log.Info("initializing IDC billing service...")
	if svc.ManagedDb == nil {
		var err error
		svc.ManagedDb, err = manageddb.New(ctx, &cfg.Database)
		if err != nil {
			log.Error(err, "error connecting to database")
			return err
		}
	}

	log.Info("successfully connected to the database")

	// if err := svc.ManagedDb.Migrate(ctx, db.MigrationsFs, db.MigrationsDir); err != nil {
	// 	log.Error(err, "error migrating database")
	// 	return err
	// }
	// log.Info("successfully migrated database model")
	var err error
	svc.Sql, err = svc.ManagedDb.Open(ctx)
	if err != nil {
		return err
	}
	billingDb = svc.Sql
	notificationClient, err := NewNotificationClient(ctx, resolver)
	if err != nil {
		log.Error(err, "failed to initialize notification client")
		return err
	}
	billingCouponService, err := NewBillingCouponService(svc.Sql, cfg.CreditsExpiryMinimumInterval)
	if err != nil {
		log.Error(err, "error initializing billing coupon service")
		return err
	}

	if err := InitDrivers(ctx, resolver); err != nil {
		return err
	}
	// cloud account client
	svc.CloudAccountSvcClient, err = billingCommon.NewCloudAccountClient(ctx, resolver)
	if err != nil {
		log.Error(err, "failed to initialize cloud account client")
		return err
	}

	// metering client
	svc.MeteringServiceClient, err = billingCommon.NewMeteringClient(ctx, resolver)

	log.Info("Connecting to usage")
	addr, err := resolver.Resolve(ctx, "usage")
	if err != nil {
		return err
	}
	conn, err := grpcutil.NewClient(ctx, addr)
	if err != nil {
		return err
	}
	svc.UsageServiceClient = pb.NewUsageServiceClient(conn)
	cloudAccountLocks = NewCloudAccountLocks(ctx)
	RegisterProxies(grpcServer)
	pb.RegisterBillingProductCatalogSyncServiceServer(grpcServer, &BillingProductCatalogSyncService{})
	pb.RegisterBillingCouponServiceServer(grpcServer, billingCouponService)
	pb.RegisterBillingAccountServiceServer(grpcServer, &BillingAccountService{})
	pb.RegisterBillingCreditServiceServer(grpcServer, &BillingCreditService{})
	pb.RegisterBillingDeactivateInstancesServiceServer(grpcServer, &BillingDeactivateInstancesService{cloudAccountClient: svc.CloudAccountSvcClient})
	pb.RegisterBillingUsageServiceServer(grpcServer, &BillingUsageService{usageServiceClient: svc.UsageServiceClient})

	// this will move to a new micro service after 1.0
	//eventPoll := events.NewEventApiSubscriber()
	eventDispatcher := events.NewEventDispatcher()
	eventData := events.NewEventData(svc.Sql)
	eventManager := events.NewEventManager(eventData, eventDispatcher)
	reflection.Register(grpcServer)
	log.Info("RunSchedulers", "RunSchedulers", cfg.RunSchedulers, "TestProfile", cfg.TestProfile)

	// start schedulers exclusively for the scheduler pod within the non-test profile.
	if !cfg.TestProfile && cfg.RunSchedulers {
		schedulerCloudAccountState := &SchedulerCloudAccountState{
			AccessTimestamp: "",
		}

		creditsInstallSched, err := NewCreditsInstallScheduler(svc.Sql, cfg.CreditsInstallSchedulerInterval, cfg.CreditsExpiryMinimumInterval)
		if err != nil {
			log.Error(err, "error starting credits Install scheduler")
			return err
		}

		cloudCreditUsageScheduler := NewCloudCreditUsageScheduler(eventManager, notificationClient, schedulerCloudAccountState, svc.CloudAccountSvcClient)

		cloudCreditExpiryScheduler := NewCloudCreditExpiryScheduler(eventManager, notificationClient, schedulerCloudAccountState, svc.CloudAccountSvcClient)

		//eventExpiryScheduler := events.NewEventExpiryScheduler(eventData)

		servicesTerminationScheduler := NewServicesTerminationScheduler(schedulerCloudAccountState, svc.CloudAccountSvcClient, svc.MeteringServiceClient)

		cloudCreditUsageReport := NewCloudCreditUsageReport(svc.Sql, svc.CloudAccountSvcClient.CloudAccountClient, svc.UsageServiceClient)

		schedulerOpsService := NewSchedulerOpsService(creditsInstallSched, cloudCreditUsageScheduler, cloudCreditExpiryScheduler, servicesTerminationScheduler)
		pb.RegisterBillingOpsActionServiceServer(grpcServer, schedulerOpsService)
		log.Info("credit install scheduler enabled", "CreditInstallScheduler", cfg.Features.CreditInstallScheduler)
		if cfg.Features.CreditInstallScheduler {
			creditsInstallSched.StartCreditsInstallScheduler(ctx)
		}
		log.Info("credit usage scheduler enabled", "CreditUsageScheduler", cfg.Features.CreditUsageScheduler)
		if cfg.Features.CreditUsageScheduler {
			startCloudCreditUsageScheduler(ctx, *cloudCreditUsageScheduler)
		}
		log.Info("credit expiry scheduler enabled", "CreditExpiryScheduler", cfg.Features.CreditExpiryScheduler)
		if cfg.Features.CreditExpiryScheduler {
			startCloudCreditExpiryScheduler(ctx, *cloudCreditExpiryScheduler)
		}
		log.Info("services termination scheduler enabled", "ServicesTerminationScheduler", cfg.Features.ServicesTerminationScheduler)
		if cfg.Features.ServicesTerminationScheduler {
			startServicesTerminationScheduler(ctx, *servicesTerminationScheduler)
		}
		log.Info("event expiry scheduler enabled", "EventExpiryScheduler", cfg.Features.EventExpiryScheduler)
		//if cfg.Features.EventExpiryScheduler {
		//	events.StartEventExpiryScheduler(ctx, eventExpiryScheduler, cfg.EventExpirySchedulerInterval)
		//}
		log.Info("cloud credit usage report scheduler enabled", "EventExpiryScheduler", cfg.Features.CloudCreditUsageReportScheduler)
		if cfg.Features.CloudCreditUsageReportScheduler {
			StartCloudCreditUsageReportScheduler(ctx, cloudCreditUsageReport)
		}
	}
	return nil
}

func (svc *Service) Name() string {
	return "billing"
}

func (*Service) Done() {
	creditsInstallSched.StopCreditsInstallScheduler()
}
