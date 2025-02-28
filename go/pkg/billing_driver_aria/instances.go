// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"time"

	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AriaBillingInstancesService struct {
	pb.UnimplementedBillingInstancesServiceServer
	meteringServiceClient *billing.MeteringClient
	productServiceClient  *billing.ProductClient
}

func (svc *AriaBillingInstancesService) Read(ctx context.Context, in *pb.BillingPaidInstanceFilter) (*pb.BillingPaidInstanceResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingInstancesService.Read").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	log.Info("BEGIN")
	log.Info("aria billing read paid instancetypes: enter")
	defer log.Info("aria billing read paid instancetypes: return")
	log.Info("aria billing driver read paid instancetypes", "input filters", in)
	defer log.Info("END")

	duration := config.Cfg.InstanceSearchWindow

	log.Info("intel billing driver instance search window", "duration", duration)
	start, end, err := billing.DetermineInstanceSearchWindow(ctx, time.Duration(duration))
	if err != nil {
		log.Error(err, "error in DetermineInstanceSearchWindow")
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	instFilter := pb.UsageFilter{
		StartTime:      start,
		EndTime:        end,
		CloudAccountId: &in.CloudAccountId,
	}

	paidInstanceTypes, err := billing.GetPaidInstanceTypes(ctx, svc.meteringServiceClient, svc.productServiceClient, &instFilter, pb.AccountType_ACCOUNT_TYPE_PREMIUM)
	if err != nil {
		log.Error(err, "error in billing.GetPaidInstanceTypes")
		return nil, err
	}

	out := &pb.BillingPaidInstanceResponse{InstanceTypes: paidInstanceTypes}
	return out, nil
}
