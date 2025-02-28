// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"errors"
	"io"

	"github.com/golang/protobuf/ptypes/empty"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type BillingCreditService struct {
	pb.UnimplementedBillingCreditServiceServer
}

func (s *BillingCreditService) Create(ctx context.Context, in *pb.BillingCredit) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCreditService.Create").WithValues("req", in).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	driver, err := s.getDriverByCloudAccountId(ctx, in.CloudAccountId)
	if err != nil {
		return nil, err
	}
	if in.Reason != pb.BillingCreditReason_CREDIT_INITIAL {
		logger.Error(errors.New(InvalidBillingCreditReason), "invalid reason code", "reason code", in.GetReason())
		return nil, status.Errorf(codes.InvalidArgument, InvalidBillingCreditReason)
	}
	if in.OriginalAmount <= 0 {
		logger.Error(errors.New(InvalidBillingCreditAmount), "invalid billing credit amount", "originalAmount", in.GetOriginalAmount())
		return nil, status.Errorf(codes.InvalidArgument, InvalidBillingCreditAmount)
	}
	if in.Expiration == nil {
		logger.Error(errors.New(InvalidBillingCreditExpiration), "invalid billing credit expiration", "expiration", in.GetExpiration())
		return nil, status.Errorf(codes.InvalidArgument, InvalidBillingCreditExpiration)
	}
	currentTime := timestamppb.Now().AsTime()
	if in.Expiration.AsTime().Sub(currentTime) < 0 {
		logger.Error(errors.New(InvalidBillingCreditExpiration), "expiration time cannot be lesser than current", "expiration time", in.Expiration.AsTime())
		return nil, status.Errorf(codes.InvalidArgument, InvalidBillingCreditExpiration)
	}
	// enable this after all tests have changed to pass the coupon code
	//isCouponValid := ValidateCouponCode(in.CouponCode)
	//if !isCouponValid {
	//	logger.Error(errors.New(InvalidCouponCode), "invalid coupon code", "couponcode", in.CouponCode)
	//	return nil, status.Errorf(codes.InvalidArgument, InvalidCouponCode)
	//}
	res, err := driver.billingCredit.Create(ctx, in)
	if err != nil {
		logger.Error(err, "failed to create billing credit")
		return nil, status.Errorf(codes.Internal, GetBillingError(FailedToCreateBillingCredit, err).Error())
	}
	return res, err
}

func (s *BillingCreditService) ReadInternal(in *pb.BillingAccount, outStream pb.BillingCreditService_ReadInternalServer) error {
	ctx := outStream.Context()
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCreditService.ReadInternal").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	driver, err := s.getDriverByCloudAccountId(ctx, in.CloudAccountId)
	if err != nil {
		return err
	}
	res, err := driver.billingCredit.ReadInternal(ctx, in)
	if err != nil {
		logger.Error(err, "failed to read billing credit")
		return status.Errorf(codes.Internal, GetBillingError(FailedToReadBillingCredit, err).Error())
	}
	for {
		out, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			logger.Error(err, "failed to read billing credit")
			return status.Errorf(codes.Internal, GetBillingError(FailedToReadBillingCredit, err).Error())
		}
		if err := outStream.Send(out); err != nil {
			logger.Error(err, "error sending read details")
			return err
		}
	}
}

func (s *BillingCreditService) Read(ctx context.Context, in *pb.BillingCreditFilter) (*pb.BillingCreditResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCreditService.Read").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	driver, err := s.getDriverByCloudAccountId(ctx, in.CloudAccountId)
	if err != nil {
		return nil, err
	}
	res, err := driver.billingCredit.Read(ctx, in)
	if err != nil {
		logger.Error(err, "failed to read billing credit")
		return nil, status.Errorf(codes.Internal, GetBillingError(FailedToReadBillingCredit, err).Error())
	}
	return res, err
}

func (s *BillingCreditService) ReadUnappliedCreditBalance(ctx context.Context, in *pb.BillingAccount) (*pb.BillingUnappliedCreditBalance, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCreditService.ReadUnappliedCreditBalance").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	driver, err := s.getDriverByCloudAccountId(ctx, in.CloudAccountId)
	if err != nil {
		return nil, err
	}
	res, err := driver.billingCredit.ReadUnappliedCreditBalance(ctx, in)
	if err != nil {
		logger.Error(err, "failed to read unapplied cloud credit balance")
		return nil, status.Errorf(codes.Internal, GetBillingError(FailedToReadUnappliedCreditBalance, err).Error())
	}
	return res, err
}

func (s *BillingCreditService) getDriverByCloudAccountId(ctx context.Context, cloudAcctId string) (*BillingDriverClients, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCreditService.getDriverByCloudAccountId").WithValues("cloudAccountId", cloudAcctId).Start()
	defer span.End()
	isCloudAcctValid := ValidateCloudAcctId(cloudAcctId)
	if !isCloudAcctValid {
		logger.Error(errors.New(InvalidCloudAccountId), "invalid cloud account id", "cloudAccountId", cloudAcctId)
		return nil, status.Errorf(codes.InvalidArgument, InvalidCloudAccountId)
	}
	driver, err := GetDriver(ctx, cloudAcctId)
	if err != nil {
		logger.Error(err, "unable to find driver", "cloudAccountId", cloudAcctId)
		return nil, status.Errorf(codes.InvalidArgument, GetBillingError(InvalidCloudAcct, err).Error())
	}
	return driver, nil
}

func (s *BillingCreditService) CreditMigrate(ctx context.Context, in *pb.BillingUnappliedCredit) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("BillingCreditService.CreditMigrate").WithValues("cloudAccountId", in.CloudAccountId).Start()
	res := &emptypb.Empty{}
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	cloudAccountLocks.Lock(ctx, in.CloudAccountId)
	defer cloudAccountLocks.Unlock(ctx, in.CloudAccountId)
	driver, err := GetDriverAll(ctx, in.CloudAccountId)
	if err != nil {
		logger.Error(err, "unable to find driver", "cloudAccountId", in.CloudAccountId)
		return nil, status.Errorf(codes.InvalidArgument, GetBillingError(InvalidCloudAcct, err).Error())
	}

	standardDriver, err := GetDriverAllByType(pb.AccountType_ACCOUNT_TYPE_STANDARD)
	if err != nil {
		logger.Error(err, "unable to find standard driver", "cloudAccountId", in.CloudAccountId)
		return nil, status.Errorf(codes.InvalidArgument, GetBillingError(InvalidCloudAcct, err).Error())
	}

	credHistory := true
	creditDetails, err := standardDriver.billingCredit.Read(ctx, &pb.BillingCreditFilter{CloudAccountId: in.CloudAccountId, History: &credHistory})
	if err != nil {
		logger.Error(err, "unable to read standard accounts credit details", "cloudAccountId", in.CloudAccountId)
		return nil, status.Errorf(codes.InvalidArgument, GetBillingError(FailedToReadBillingCredit, err).Error())
	}

	currAccountObj, err := cloudacctClient.GetById(ctx, &pb.CloudAccountId{Id: in.CloudAccountId})
	if err != nil {
		logger.Error(err, "failed to get cloudaccount details.")
		return nil, status.Error(codes.Internal, "failed to get cloudaccount details.")
	}

	if currAccountObj.UpgradedToPremium == pb.UpgradeStatus_UPGRADE_PENDING_CC {
		// TODO: Handle this to Enterprise as well when Upgrade to enterprise is supported
		logger.Info("Cannot complete upgrade since credit card not added successfully", "cloudAccountId", in.CloudAccountId)
		return nil, status.Error(codes.Internal, "credit card addition not successful")
	}

	if currAccountObj.UpgradedToPremium == pb.UpgradeStatus_UPGRADE_PENDING || currAccountObj.UpgradedToPremium == pb.UpgradeStatus_UPGRADE_PENDING_CC_VERIFIED {
		if creditDetails.GetTotalRemainingAmount() > 0 {

			creditPremDetailsOne, err := driver.billingCredit.Read(ctx, &pb.BillingCreditFilter{CloudAccountId: in.CloudAccountId, History: &credHistory})
			if err != nil {
				logger.Error(err, "unable to read premium account credit details", "cloudAccountId", in.CloudAccountId)
				return nil, status.Errorf(codes.InvalidArgument, GetBillingError(FailedToReadBillingCredit, err).Error())
			}

			res, err = driver.billingCredit.CreditMigrate(ctx, &pb.BillingUnappliedCredit{
				CloudAccountId:  in.CloudAccountId,
				RemainingAmount: creditDetails.TotalRemainingAmount,
				Expiration:      creditDetails.ExpirationDate,
			})
			if err != nil {
				logger.Error(err, "failed to migrate credits")

				creditPremDetailsTwo, err := driver.billingCredit.Read(ctx, &pb.BillingCreditFilter{CloudAccountId: in.CloudAccountId, History: &credHistory})
				if err != nil {
					logger.Error(err, "unable to read premium account credit details", "cloudAccountId", in.CloudAccountId)
					return nil, status.Errorf(codes.InvalidArgument, GetBillingError(FailedToReadBillingCredit, err).Error())
				}

				// In case credit service in aria was created with standard unapplied amount and failed in the subsequent
				// steps, we need to take care of deletion of those migrated credits.
				creditValue := creditPremDetailsTwo.TotalRemainingAmount - creditPremDetailsOne.TotalRemainingAmount
				if creditValue == creditDetails.TotalRemainingAmount {
					_, err = standardDriver.billingCredit.DeleteMigratedCredit(ctx, &pb.BillingMigratedCredit{
						CloudAccountId: in.CloudAccountId,
					})
					if err != nil {
						logger.Error(err, "failed to delete migrated credits")
						return nil, status.Errorf(codes.Internal, GetBillingError(FailedToDeleteMigratedCredits, err).Error())
					}
				}

				return nil, status.Errorf(codes.Internal, GetBillingError(FailedToMigrateCredits, err).Error())
			}

			// Deletion of credits in standard CA which is upgraded to premium
			_, err = standardDriver.billingCredit.DeleteMigratedCredit(ctx, &pb.BillingMigratedCredit{
				CloudAccountId: in.CloudAccountId,
			})
			if err != nil {
				logger.Error(err, "failed to delete migrated credits")
				return nil, status.Errorf(codes.Internal, GetBillingError(FailedToDeleteMigratedCredits, err).Error())
			}
		}
		// TODO: Handle this to Enterprise as well when Upgrade to enterprise is supported
		_, err = cloudacctClient.Update(ctx, &pb.CloudAccountUpdate{
			Id:                in.CloudAccountId,
			UpgradedToPremium: pb.UpgradeStatus_UPGRADE_COMPLETE.Enum(),
		})
		if err != nil {
			logger.Error(err, "cloudaccount update failed", "cloudAccountId", in.CloudAccountId)
			return nil, err
		}
	}

	return res, nil
}
