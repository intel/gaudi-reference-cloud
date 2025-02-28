// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"

	"github.com/golang/protobuf/ptypes/empty"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BillingAccountService struct {
	pb.UnimplementedBillingAccountServiceServer
}

func (s *BillingAccountService) Create(ctx context.Context, in *pb.BillingAccount) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingAccountService.Create").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	isCloudAcctValid := ValidateCloudAcctId(in.CloudAccountId)
	if !isCloudAcctValid {
		logger.Error(errors.New(InvalidCloudAccountId), "invalid cloud account id", "cloudAccountId", in.GetCloudAccountId())
		return nil, status.Errorf(codes.InvalidArgument, InvalidCloudAccountId)
	}
	driver, err := GetDriverAll(ctx, in.CloudAccountId)
	if err != nil {
		logger.Error(err, "unable to find driver", "cloudAccountId", in.GetCloudAccountId())
		return nil, status.Errorf(codes.InvalidArgument, GetBillingError(InvalidCloudAcct, err).Error())
	}
	res, err := driver.billingAcct.Create(ctx, in)
	if err != nil {
		logger.Error(err, "failed to create acct")
		return nil, status.Errorf(codes.Internal, GetBillingError(FailedToCreateBillingAccount, err).Error())
	}
	return res, err
}

func (s *BillingAccountService) DowngradePremiumtoStandard(ctx context.Context, in *pb.BillingAccountDowngrade) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("BillingAccountService.DowngradePremiumtoStandard").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	isCloudAcctValid := ValidateCloudAcctId(in.CloudAccountId)
	if !isCloudAcctValid {
		logger.Error(errors.New(InvalidCloudAccountId), "invalid cloud account id", "cloudAccountId", in.GetCloudAccountId())
		return nil, status.Errorf(codes.InvalidArgument, InvalidCloudAccountId)
	}

	cloudAccount, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: in.CloudAccountId})
	if err != nil {
		logger.Error(err, "failed to get cloud account to downgrade from premium to standard")
		return nil, status.Errorf(codes.NotFound, "failed to get cloud account to downgrade from premium to standard")
	}
	if cloudAccount.Type != pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		logger.Error(err, "provided cloudAccountid is not premium type not supported")
		return nil, status.Errorf(codes.InvalidArgument, "provided cloudAccountid is not premium type not supported")
	}

	res, err := ariaDriver.billingAcct.DowngradePremiumtoStandard(ctx, in)
	if err != nil {
		logger.Error(err, "failed to downgrade account from premium to standard.")
		return nil, status.Errorf(codes.Internal, GetBillingError(FailedToDowngradeBillingAccountPremiumToStandard, err).Error())
	}

	//Update cloud account table to change the user type from premium to standard
	_, err = cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{Id: in.CloudAccountId, Type: pb.AccountType_ACCOUNT_TYPE_STANDARD.Enum()})
	if err != nil {
		logger.Error(err, "failed to update cloud account type from on downgrading premium to standard")
		return nil, status.Errorf(codes.Internal, "failed to update cloud account type from on downgrading premium to standard")
	}

	logger.V(9).Info("cloud account was successfully downgraded from premium to standard")

	return res, err
}
