// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloudaccount_enroll

import (
	"context"

	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FinancialFlags struct {
	LowCredits             bool
	CreditsDepleted        *timestamppb.Timestamp
	TerminatePaidInstances bool
	TerminateMsgQueued     bool
	Delinquent             bool
	PaidServiceAllowed     bool
	UpgradedToPremium      pb.UpgradeStatus
	UpgradedToEnterprise   pb.UpgradeStatus
}

func DefaultFinFlags() *FinancialFlags {
	return &FinancialFlags{
		LowCredits:             false,
		CreditsDepleted:        &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
		TerminatePaidInstances: false,
		TerminateMsgQueued:     false,
		Delinquent:             false,
		PaidServiceAllowed:     false,
		UpgradedToPremium:      pb.UpgradeStatus_UPGRADE_NOT_INITIATED,
		UpgradedToEnterprise:   pb.UpgradeStatus_UPGRADE_NOT_INITIATED,
	}
}

func (ces *CloudAccountEnrollService) ValidateCoupon(ctx context.Context, cloudAccountId string, couponResponse *pb.BillingCouponResponse) (bool, error) {
	couponRes := couponResponse.Coupons[0]
	if couponRes.Disabled != nil {
		if couponRes.Disabled.AsTime().Sub(timestamppb.Now().AsTime()) < 0 {
			return false, status.Errorf(codes.Internal, "coupon %v has been disabled", couponRes.Code)
		}
	}
	if couponRes.Start.AsTime().Sub(timestamppb.Now().AsTime()) > 0 {
		return false, status.Errorf(codes.Internal, "coupon %v cannot be redeemed before %v", couponRes.Code, couponRes.Start.AsTime())
	}
	if couponRes.Expires.AsTime().Sub(timestamppb.Now().AsTime()) < 0 {
		return false, status.Errorf(codes.Internal, "cannot redeem coupon %v as it has expired on %v", couponRes.Code, couponRes.Expires.AsTime())
	}
	if couponRes.NumUses == couponRes.NumRedeemed {
		return false, status.Errorf(codes.Internal, "cannot redeem coupon as it has exceeded number of uses %v redemptions %v", couponRes.NumUses, couponRes.NumRedeemed)
	}
	if couponRes.Redemptions != nil {
		for _, redeemed := range couponRes.Redemptions {
			if redeemed.CloudAccountId == cloudAccountId {
				return false, status.Errorf(codes.AlreadyExists, "the coupon code (%v) has already been redeemed", couponRes.Code)
			}
		}
	}
	if *couponRes.IsStandard {
		return false, status.Errorf(codes.Internal, "coupon %v can be redeemed only for standard customers", couponRes.Code)
	}
	return true, nil
}

func (ces *CloudAccountEnrollService) ValidateCouponCredits(ctx context.Context, cloudAccountId string, couponResponse *pb.CouponResponse) (bool, error) {

	couponRes := couponResponse.Coupons[0]
	if couponRes.Disabled != nil {
		if couponRes.Disabled.AsTime().Sub(timestamppb.Now().AsTime()) < 0 {
			return false, status.Errorf(codes.Internal, "coupon %v has been disabled", couponRes.Code)
		}
	}
	if couponRes.Start.AsTime().Sub(timestamppb.Now().AsTime()) > 0 {
		return false, status.Errorf(codes.Internal, "coupon %v cannot be redeemed before %v", couponRes.Code, couponRes.Start.AsTime())
	}
	if couponRes.Expires.AsTime().Sub(timestamppb.Now().AsTime()) < 0 {
		return false, status.Errorf(codes.Internal, "cannot redeem coupon %v as it has expired on %v", couponRes.Code, couponRes.Expires.AsTime())
	}
	if couponRes.NumUses == couponRes.NumRedeemed {
		return false, status.Errorf(codes.Internal, "cannot redeem coupon as it has exceeded number of uses %v redemptions %v", couponRes.NumUses, couponRes.NumRedeemed)
	}
	if couponRes.Redemptions != nil {
		for _, redeemed := range couponRes.Redemptions {
			if redeemed.CloudAccountId == cloudAccountId {
				return false, status.Errorf(codes.AlreadyExists, "the coupon code (%v) has already been redeemed", couponRes.Code)
			}
		}
	}
	if *couponRes.IsStandard {
		return false, status.Errorf(codes.Internal, "coupon %v can be redeemed only for standard customers", couponRes.Code)
	}
	return true, nil
}
func (ces *CloudAccountEnrollService) Upgrade(ctx context.Context, req *pb.CloudAccountUpgradeRequest) (*pb.CloudAccountUpgradeResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountEnrollService.Upgrade").WithValues("cloudAccountId", req.CloudAccountId).Start()
	logger.Info("upgrade API invoked", "cloudAccountId", req.CloudAccountId)
	defer logger.Info("upgrade API return", "cloudAccountId", req.CloudAccountId)
	defer span.End()

	resp := &pb.CloudAccountUpgradeResponse{
		CloudAccountId: req.CloudAccountId,
		Action:         pb.CloudAccountUpgradeAction_UPGRADE_ACTION_RETRY,
	}

	if req.Code == nil || *req.Code == "" {
		logger.Info("Coupon code cannot be empty")
		return resp, status.Error(codes.InvalidArgument, "coupon code needs to be provided")
	}

	// check for cloudaccount
	currAccountObj, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: req.CloudAccountId})
	if err != nil {
		logger.Error(err, "failed to get cloudaccount details", "cloudAccountId", req.CloudAccountId, "context", "GetById")
		return resp, status.Error(codes.Internal, "failed to get cloudaccount details.")
	}
	crrAccntId := currAccountObj.Id
	crrAccntType := currAccountObj.Type
	initAccntType := crrAccntType
	resp.CloudAccountType = crrAccntType

	if currAccountObj.UpgradedToPremium == pb.UpgradeStatus_UPGRADE_NOT_INITIATED && req.CloudAccountUpgradeToType == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		if crrAccntType == pb.AccountType_ACCOUNT_TYPE_STANDARD {
			validCoupon := false
			// Checks for valid coupon for upgrade
			logger.Info("validating coupon")

			couponResponse, err := cloudCreditsCouponServiceClient.Read(ctx, &pb.CouponFilter{
				Code: req.Code,
			})
			if err != nil {
				logger.Error(err, "unable to get coupon", "coupon", req.Code)
				return resp, err
			}
			validCoupon, err = ces.ValidateCouponCredits(ctx, crrAccntId, couponResponse)
			if err != nil || !validCoupon {
				logger.Error(err, "invalid", "coupon", req.Code)
				return resp, err
			}

			// update cloudaccount type and reset flags
			dflags := DefaultFinFlags()
			dflags.UpgradedToPremium = pb.UpgradeStatus_UPGRADE_PENDING
			updatedAccountObj, stdAccntFlags, err := ces.UpdateCloudAccount(ctx, currAccountObj, &req.CloudAccountUpgradeToType, dflags)
			if err != nil {
				logger.Error(err, "cloudaccount reset failed", "cloudAccountId", currAccountObj.Id, "cloudAccountUpgradeToType", &req.CloudAccountUpgradeToType, "dflags", dflags, "context", "UpdateCloudAccount")
				return resp, status.Error(codes.Internal, "cloudaccount reset failed.")
			}
			logger.Info("updated cloud account details", "currAccountObj", updatedAccountObj)
			currAccountObj = updatedAccountObj
			resp.CloudAccountType = currAccountObj.Type

			// billing account creation
			// if already exist account is not re-created
			_, err = billingAcctClient.Create(ctx, &pb.BillingAccount{
				CloudAccountId: crrAccntId,
			})
			if err != nil {
				logger.Error(err, "billing account create failed", "cloudAccountId", req.CloudAccountId)
				// revert cloudaccount flags
				updatedAccountObj, _, err := ces.UpdateCloudAccount(ctx, currAccountObj, &initAccntType, stdAccntFlags)
				if err != nil {
					logger.Error(err, "cloudaccount reset failed", "cloudAccountId", currAccountObj.Id, "initAccntType", &initAccntType, "stdAccntFlags", stdAccntFlags, "context", "UpdateCloudAccount")
				}
				logger.Error(err, "cloud account revert upgrade", "cloudAccountId", updatedAccountObj.Id, "type", updatedAccountObj.Type)
				resp.Action = pb.CloudAccountUpgradeAction_UPGRADE_ACTION_RETRY
				return resp, status.Error(codes.Internal, "failed to create billing account.")
			}

			if validCoupon {
				_, err = cloudCreditsCouponServiceClient.Redeem(ctx, &pb.CloudCreditsCouponRedeem{
					CloudAccountId: crrAccntId,
					Code:           *req.Code,
				})
				if err != nil {
					logger.Error(err, "failed to redeem coupon", "cloudAccountId", updatedAccountObj.Id, "type", updatedAccountObj.Type)
					resp.Action = pb.CloudAccountUpgradeAction_UPGRADE_ACTION_RETRY
					return resp, status.Error(codes.Internal, "failed to redeem coupon, try redeeming with valid coupon")
				}
			}

			resp.Action = pb.CloudAccountUpgradeAction_UPGRADE_ACTION_NONE
		} else {
			logger.Info("Upgrade path not supported.")
			resp.Action = pb.CloudAccountUpgradeAction_UPGRADE_ACTION_UNSUPPORTED
		}
	} else {
		logger.Info("upgrade path not supported.")
		resp.Action = pb.CloudAccountUpgradeAction_UPGRADE_ACTION_UNSUPPORTED
	}
	return resp, nil
}

func (ces *CloudAccountEnrollService) UpgradeWithCreditCard(ctx context.Context, req *pb.CloudAccountUpgradeWithCreditCardRequest) (*pb.CloudAccountUpgradeResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountEnrollService.UpgradeWithCreditCard").WithValues("cloudAccountId", req.CloudAccountId).Start()
	logger.Info("upgrade with credit card API invoked", "cloudAccountId", req.CloudAccountId)
	defer logger.Info("upgrade with credit card API return", "cloudAccountId", req.CloudAccountId)
	defer span.End()

	resp := &pb.CloudAccountUpgradeResponse{
		CloudAccountId: req.CloudAccountId,
		Action:         pb.CloudAccountUpgradeAction_UPGRADE_ACTION_RETRY,
	}

	// check for cloudaccount
	currAccountObj, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: req.CloudAccountId})
	if err != nil {
		logger.Error(err, "failed to get cloudaccount details", "cloudAccountId", req.CloudAccountId, "context", "GetById")
		return resp, status.Error(codes.Internal, "failed to get cloudaccount details.")
	}
	crrAccntId := currAccountObj.Id
	crrAccntType := currAccountObj.Type
	initAccntType := crrAccntType
	resp.CloudAccountType = crrAccntType

	if currAccountObj.UpgradedToPremium == pb.UpgradeStatus_UPGRADE_NOT_INITIATED && req.CloudAccountUpgradeToType == pb.AccountType_ACCOUNT_TYPE_PREMIUM {
		if crrAccntType == pb.AccountType_ACCOUNT_TYPE_STANDARD {
			// update cloudaccount type and reset flags
			dflags := DefaultFinFlags()
			dflags.UpgradedToPremium = pb.UpgradeStatus_UPGRADE_PENDING_CC
			updatedAccountObj, stdAccntFlags, err := ces.UpdateCloudAccount(ctx, currAccountObj, &req.CloudAccountUpgradeToType, dflags)
			if err != nil {
				logger.Error(err, "cloudaccount reset failed", "cloudAccountId", currAccountObj.Id, "cloudAccountUpgradeToType", &req.CloudAccountUpgradeToType, "dflags", dflags, "context", "UpdateCloudAccount")
				return resp, status.Error(codes.Internal, "cloudaccount reset failed.")
			}
			logger.Info("updated cloud account details", "currAccountObj", updatedAccountObj)
			currAccountObj = updatedAccountObj
			resp.CloudAccountType = currAccountObj.Type

			// billing account creation
			// if already exist account is not re-created
			_, err = billingAcctClient.Create(ctx, &pb.BillingAccount{
				CloudAccountId: crrAccntId,
			})
			if err != nil {
				logger.Error(err, "billing account create failed", "cloudAccountId", req.CloudAccountId)
				// revert cloudaccount flags
				updatedAccountObj, _, err := ces.UpdateCloudAccount(ctx, currAccountObj, &initAccntType, stdAccntFlags)
				if err != nil {
					logger.Error(err, "cloudaccount reset failed", "cloudAccountId", currAccountObj.Id, "initAccntType", &initAccntType, "stdAccntFlags", stdAccntFlags, "context", "UpdateCloudAccount")
				}
				logger.Error(err, "cloud account revert upgrade", "cloudAccountId", updatedAccountObj.Id, "type", updatedAccountObj.Type)
				resp.Action = pb.CloudAccountUpgradeAction_UPGRADE_ACTION_RETRY
				return resp, status.Error(codes.Internal, "failed to create billing account.")
			}
			resp.Action = pb.CloudAccountUpgradeAction_UPGRADE_ACTION_NONE
		} else {
			logger.Info("Upgrade path not supported.")
			resp.Action = pb.CloudAccountUpgradeAction_UPGRADE_ACTION_UNSUPPORTED
		}
	} else {
		logger.Info("upgrade path not supported.")
		resp.Action = pb.CloudAccountUpgradeAction_UPGRADE_ACTION_UNSUPPORTED
	}
	return resp, nil
}

func (ces *CloudAccountEnrollService) UpdateCloudAccount(ctx context.Context, currAccountObj *pb.CloudAccount, upgradeType *pb.AccountType, flags *FinancialFlags) (*pb.CloudAccount, *FinancialFlags, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountEnrollService.UpdateCloudAccount").WithValues("cloudAccountId", currAccountObj.Id).Start()
	defer span.End()
	logger.Info("Attempting to update cloud account ", "currAccntType", currAccountObj)

	stdAccntFlags := ces.getFlags(ctx, currAccountObj)
	stdAccntFlags.UpgradedToPremium = flags.UpgradedToPremium

	_, err := cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{
		Id:                     currAccountObj.Id,
		Type:                   upgradeType,
		LowCredits:             &flags.LowCredits,
		CreditsDepleted:        flags.CreditsDepleted,
		TerminatePaidServices:  &flags.TerminatePaidInstances,
		TerminateMessageQueued: &flags.TerminateMsgQueued,
		Delinquent:             &flags.Delinquent,
		PaidServicesAllowed:    &flags.PaidServiceAllowed,
		UpgradedToPremium:      &flags.UpgradedToPremium,
		UpgradedToEnterprise:   &flags.UpgradedToEnterprise,
	})
	if err != nil {
		logger.Error(err, "cloudaccount update failed", "cloudAccountId", currAccountObj.Id)
		return nil, nil, err
	}

	// new cloudaccount details
	updatedAccountObj, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: currAccountObj.Id})
	if err != nil {
		logger.Error(err, "failed to get cloudaccount details", "cloudAccountId", currAccountObj.Id, "context", "GetById")
		return nil, nil, err
	}
	return updatedAccountObj, stdAccntFlags, nil
}

func (ces *CloudAccountEnrollService) getFlags(ctx context.Context, currAccountObj *pb.CloudAccount) *FinancialFlags {
	accntflags := DefaultFinFlags()

	// get current account flags
	accntflags.LowCredits = currAccountObj.LowCredits
	accntflags.CreditsDepleted = currAccountObj.CreditsDepleted
	accntflags.TerminatePaidInstances = currAccountObj.TerminatePaidServices
	accntflags.TerminateMsgQueued = currAccountObj.TerminateMessageQueued
	accntflags.Delinquent = currAccountObj.Delinquent
	accntflags.PaidServiceAllowed = currAccountObj.PaidServicesAllowed
	accntflags.UpgradedToPremium = currAccountObj.UpgradedToPremium
	accntflags.UpgradedToEnterprise = currAccountObj.UpgradedToEnterprise

	return accntflags
}

// check if credit card is valid payment method
func (ces *CloudAccountEnrollService) creditCardAvailable(ctx context.Context, currAccountObj *pb.CloudAccount) (bool, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("CloudAccountEnrollService.creditCardAvailable").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	billingOption, err := billingOptionsClient.Read(ctx, &pb.BillingOptionFilter{
		CloudAccountId: &currAccountObj.Id,
	})
	if err != nil {
		logger.Error(err, "read billing options failed", "cloudAccountId", &currAccountObj.Id, "context", "Read")
		return false, err
	}
	logger.Info("Available payment method", "billingOption.PaymentMethod", billingOption.PaymentMethod)
	return billingOption.PaymentType == pb.PaymentType_PAYMENT_CREDIT_CARD, nil
}
