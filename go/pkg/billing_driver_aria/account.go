// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AriaBillingAccountService struct {
	ariaController *AriaController
	pb.UnimplementedBillingAccountServiceServer
}

func (ariaBillingAccountService *AriaBillingAccountService) Create(ctx context.Context, in *pb.BillingAccount) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingAccountService.Create").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("creating account in Aria", "id", client.GetAccountClientId(in.GetCloudAccountId()))
	if err := ariaBillingAccountService.ariaController.CreateAriaAccount(ctx, in.CloudAccountId); err != nil {
		logger.Error(err, "error in creating aria account", "request", in, "cloudAccountId", in.GetCloudAccountId())
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToCreateBillingAcct, err).Error())
	}
	return &emptypb.Empty{}, nil
}

func (ariaBillingAccountService *AriaBillingAccountService) DowngradePremiumtoStandard(ctx context.Context, in *pb.BillingAccountDowngrade) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingAccountService.DowngradePremiumtoStandard").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	logger.Info("downgrading account from premium to standard.")

	if err := ariaBillingAccountService.ariaController.DowngradePremiumtoStandard(ctx, in.CloudAccountId, in.Force); err != nil {
		logger.Error(err, "error in downgrading aria account from premium to standard", "request", in, "force", in.Force)
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToDowngradeAcctPremiumToStandard, err).Error())
	}

	return &emptypb.Empty{}, nil
}
