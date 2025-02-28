// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudcreditsworker

import (
	"context"
	"os"

	billingCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	cloudCredits "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloud_credits"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	CloudAccountClient        *billingCommon.CloudAccountSvcClient
	notificationGwClient      *billingCommon.NotificationGatewayClient
	CloudCreditsClient        *cloudCredits.CloudCreditsServiceClient
	CloudCreditUsageScheduler *cloudCredits.CloudCreditUsageEventScheduler
)

type Worker struct {
}

type CloudAccountService struct {
	pb.UnimplementedCloudAccountServiceServer // Used for forward compatability
}

func (w *Worker) Init(ctx context.Context, cfg *config.Config, resolver grpcutil.Resolver) error {
	config.Cfg = cfg
	logger := log.FromContext(ctx).WithName("Worker.Init")
	logger.Info("worker initialization")

	if err := w.initClients(ctx, cfg, resolver); err != nil {
		return err
	}
	if err := w.initAWS(ctx); err != nil {
		return err
	}
	stopChan := make(chan os.Signal, 1)
	cloudCreditWorker := NewCloudCreditsWorker(&stopChan, notificationGwClient, CloudAccountClient, CloudCreditsClient, CloudCreditUsageScheduler)
	if !config.Cfg.TestProfile {
		cloudCreditWorker.StartSQSConsumerProcess(ctx)
	}
	return nil
}

func (w *Worker) initAWS(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("worker.initAWSSession")
	logger.Info("AWSSession initialization")
	// Queue
	awsQueueName := config.Cfg.GetAWSSQSQueueName()
	logger.Info("AWS SQS", "awsQueueName", awsQueueName)
	return nil
}

func (w *Worker) initClients(ctx context.Context, cfg *config.Config, resolver grpcutil.Resolver) error {
	logger := log.FromContext(ctx).WithName("worker.initClients")
	logger.Info("clients initialization")
	config.Cfg = cfg
	var err error
	CloudAccountClient, err = billingCommon.NewCloudAccountClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize cloud account client")
		return err
	}

	notificationGwClient, err = billingCommon.NewNotificationGatewayClient(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize cloud account client")
		return err
	}
	CloudCreditsClient, err = cloudCredits.NewCloudCreditsService(ctx, resolver)
	if err != nil {
		logger.Error(err, "failed to initialize cloud credit client")
		return err
	}
	schedulerCloudAccountState := &cloudCredits.SchedulerCloudAccountState{
		AccessTimestamp: "",
	}
	CloudCreditUsageScheduler = cloudCredits.NewCloudCreditUsageEventScheduler(notificationGwClient, schedulerCloudAccountState, CloudAccountClient)
	if err != nil {
		logger.Error(err, "failed to initialize cloud credit client")
		return err
	}
	if !config.Cfg.TestProfile {
		err = billingCommon.InitDrivers(ctx, CloudAccountClient.CloudAccountClient, resolver)
		if err != nil {
			logger.Error(err, "failed to initialize drivers")
			return err
		}
	}

	return nil
}

func (*Worker) Name() string {
	return "cloudcredits-worker"
}
