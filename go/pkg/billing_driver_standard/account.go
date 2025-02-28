// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package standard

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

type StandardBillingAccountService struct {
	pb.UnimplementedBillingAccountServiceServer
}

// Test scripts use CreateAccountError to force an error return from
// BillingAcountService.Create for a standard account.
var CreateAccountError error

func (svc *StandardBillingAccountService) Create(ctx context.Context, in *pb.BillingAccount) (*empty.Empty, error) {
	_, logger, span := obs.LogAndSpanFromContext(ctx).WithName("StandardBillingAccountService.Create").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("creating billing account", "cloudAccountId", in.GetCloudAccountId())
	if CreateAccountError != nil {
		logger.Info("create billing account forced error", "error", CreateAccountError)
		return nil, CreateAccountError
	}
	// standard service doesn't manage any account information outside
	// of cloudaccount, so account creation is a successful NOOP
	return &emptypb.Empty{}, nil
}
