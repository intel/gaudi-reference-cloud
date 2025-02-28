// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import (
	"context"
	"time"

	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StandardBillingInstancesService struct {
	pb.UnimplementedBillingInstancesServiceServer
	meteringServiceClient *billing.MeteringClient
	productServiceClient  *billing.ProductClient
	config                *Config
}

func (svc *StandardBillingInstancesService) Read(ctx context.Context, in *pb.BillingPaidInstanceFilter) (*pb.BillingPaidInstanceResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("StandardBillingInstancesService.Read").WithValues("cloudAcccountId", in.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")
	log.Info("standard billing read paid instancetypes: enter")
	defer log.Info("standard billing read paid instancetypes: return")

	log.Info("standard billing driver read paid instancetypes", "input filters", in)

	duration := svc.config.CommonConfig.InstanceSearchWindow
	log.Info("standard billing InstanceSearchWindow", "InstanceSearchWindow", duration)
	start, end, err := billing.DetermineInstanceSearchWindow(ctx, time.Duration(duration))
	if err != nil {
		log.Error(err, "error in billing.InstanceSearchWindow")
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	instFilter := pb.UsageFilter{
		StartTime:      start,
		EndTime:        end,
		CloudAccountId: &in.CloudAccountId,
	}

	paidInstanceTypes, err := billing.GetPaidInstanceTypes(ctx, svc.meteringServiceClient, svc.productServiceClient, &instFilter, pb.AccountType_ACCOUNT_TYPE_STANDARD)
	if err != nil {
		log.Error(err, "error in billing.GetPaidInstanceTypes")
		return nil, err
	}

	out := &pb.BillingPaidInstanceResponse{InstanceTypes: paidInstanceTypes}
	return out, nil
}
