// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"time"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type SchedulerOpsService struct {
	creditsInstallScheduler      *CreditsInstallScheduler
	cloudCreditUsageScheduler    *CloudCreditUsageScheduler
	cloudCreditExpiryScheduler   *CloudCreditExpiryScheduler
	servicesTerminationScheduler *ServicesTerminationScheduler
	pb.UnimplementedBillingOpsActionServiceServer
}

func NewSchedulerOpsService(creditsInstallScheduler *CreditsInstallScheduler,
	cloudCreditUsageScheduler *CloudCreditUsageScheduler,
	cloudCreditExpiryScheduler *CloudCreditExpiryScheduler, servicesTerminationScheduler *ServicesTerminationScheduler) *SchedulerOpsService {
	return &SchedulerOpsService{
		creditsInstallScheduler:      creditsInstallScheduler,
		cloudCreditUsageScheduler:    cloudCreditUsageScheduler,
		cloudCreditExpiryScheduler:   cloudCreditExpiryScheduler,
		servicesTerminationScheduler: servicesTerminationScheduler,
	}
}

type SchedulerOpsContext struct {
	ctx context.Context
}

func (d SchedulerOpsContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (d SchedulerOpsContext) Done() <-chan struct{} {
	return nil
}

func (d SchedulerOpsContext) Err() error {
	return nil
}

func (d SchedulerOpsContext) Value(key any) any {
	return d.ctx.Value(key)
}

func (s *SchedulerOpsService) Create(ctx context.Context, schedulerAction *pb.SchedulerAction) (*emptypb.Empty, error) {
	ctx = SchedulerOpsContext{ctx: ctx}
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("SchedulerOpsService.Create").Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")

	switch schedulerAction.Action {
	case pb.SchedulerActionType_START_CREDIT_INSTALL_SCHEDULER:
		s.creditsInstallScheduler.StopCreditsInstallScheduler()
		s.creditsInstallScheduler.StartCreditsInstallScheduler(ctx)
	case pb.SchedulerActionType_START_CREDIT_USAGE_SCHEDULER:
		stopCloudCreditUsageScheduler()
		startCloudCreditUsageScheduler(ctx, *s.cloudCreditUsageScheduler)
	case pb.SchedulerActionType_START_CREDIT_EXPIRY_SCHEDULER:
		stopCloudCreditExpiryScheduler()
		startCloudCreditExpiryScheduler(ctx, *s.cloudCreditExpiryScheduler)
	case pb.SchedulerActionType_START_SERVICE_TERMINATION_SCHEDULER:
		stopServicesTerminationScheduler()
		startServicesTerminationScheduler(ctx, *s.servicesTerminationScheduler)
	default:
		return nil, status.Errorf(codes.Internal, InvalidSchedulerOpsAction)
	}
	return &emptypb.Empty{}, nil
}
