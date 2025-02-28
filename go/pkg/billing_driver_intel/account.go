// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package billing_driver_intel

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IntelBillingAccountService struct {
	pb.UnimplementedBillingAccountServiceServer
}

func (svc *IntelBillingAccountService) Create(ctx context.Context, in *pb.BillingAccount) (*empty.Empty, error) {
	_, log, span := obs.LogAndSpanFromContext(ctx).WithName("IntelBillingAccountService.Create").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	log.Info("BEGIN")
	defer log.Info("END")
	log.Info("creating billing account", "cloudAccountId", in.GetCloudAccountId())

	// intel billing driver doesn't manage any account information outside
	// of cloudaccount, so account creation is a successful NOOP
	return &emptypb.Empty{}, nil
}
