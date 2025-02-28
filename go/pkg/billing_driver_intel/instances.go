// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing_driver_intel

import (
	"context"
	"time"

	billing "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IntelBillingInstancesService struct {
	pb.UnimplementedBillingInstancesServiceServer
	meteringServiceClient *billing.MeteringClient
	productServiceClient  *billing.ProductClient
	config                *Config
}

func (svc *IntelBillingInstancesService) Read(ctx context.Context, in *pb.BillingPaidInstanceFilter) (*pb.BillingPaidInstanceResponse, error) {
	ctx, log, span := obs.LogAndSpanFromContext(ctx).WithName("IntelBillingInstancesService.Read").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	log.Info("intel billing read paid instancetypes: enter")
	defer log.Info("intel billing read paid instancetypes: return")
	defer log.Info("END")
	log.Info("intel billing driver read paid instancetypes", "input filters", in)

	duration := svc.config.CommonConfig.InstanceSearchWindow
	log.Info("intel billing InstanceSearchWindow", "InstanceSearchWindow", duration)
	start, end, err := billing.DetermineInstanceSearchWindow(ctx, time.Duration(duration))
	if err != nil {
		log.Error(err, "error in billing.DetermineInstanceSearchWindow")
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	instFilter := pb.UsageFilter{
		StartTime:      start,
		EndTime:        end,
		CloudAccountId: &in.CloudAccountId,
	}

	paidInstanceTypes, err := billing.GetPaidInstanceTypes(ctx, svc.meteringServiceClient, svc.productServiceClient, &instFilter, pb.AccountType_ACCOUNT_TYPE_INTEL)
	if err != nil {
		log.Error(err, "error in billing.GetPaidInstanceTypes")
		return nil, err
	}

	out := &pb.BillingPaidInstanceResponse{InstanceTypes: paidInstanceTypes}
	return out, nil
}
