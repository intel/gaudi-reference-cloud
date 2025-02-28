// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	SERVICE_CREDITS = "S"
	CASH_CREDITS    = "C"
)

type AriaBillingCreditService struct {
	ariaController *AriaController
	pb.UnimplementedBillingCreditServiceServer
}

func (ariaBillingCreditService *AriaBillingCreditService) Create(ctx context.Context, in *pb.BillingCredit) (*emptypb.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingCreditService.Create").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	expiry := in.Expiration.AsTime()
	expirationDate := fmt.Sprintf("%04d-%02d-%02d", expiry.Year(), expiry.Month(), expiry.Day())
	logger.Info("expirationDate", "expirationDate", expirationDate)
	if err := ariaBillingCreditService.ariaController.CreateServiceCredit(ctx, in.CloudAccountId, in.OriginalAmount, expirationDate, in.CouponCode); err != nil {
		logger.Error(err, "error in creating service credit", "request", in)
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToCreateBillingCredit, err).Error())
	}
	return &emptypb.Empty{}, nil
}

// Aria Driver API to apply the credits to Aria via aria controller
func (ariaBillingCreditService *AriaBillingCreditService) CreditMigrate(ctx context.Context, in *pb.BillingUnappliedCredit) (*empty.Empty, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingCreditService.CreditMigrate").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	logger.Info("credit migration account to Aria", "id", client.GetAccountClientId(in.GetCloudAccountId()))

	// futureTime is 31 days validity check
	futureTime := timestamppb.New(timestamppb.Now().AsTime().AddDate(0, 0, 31))

	expiry := in.Expiration.AsTime()
	// Check if the creditDetails.ExpirationDate is less than 31 days from today
	if expiry.Before(futureTime.AsTime()) {
		// defaulting it to 31 days - premiumDefaultCreditExpirationDays
		expiry = timestamppb.New(timestamppb.Now().AsTime().AddDate(0, 0, config.Cfg.PremiumDefaultCreditExpirationDays)).AsTime()
	}

	expirationDate := fmt.Sprintf("%04d-%02d-%02d", expiry.Year(), expiry.Month(), expiry.Day())
	logger.Info("expirationDate", "expirationDate", expirationDate)
	if err := ariaBillingCreditService.ariaController.CreateServiceCredit(ctx, in.CloudAccountId, in.RemainingAmount, expirationDate, "Migration of credits from standard"); err != nil {
		logger.Error(err, "error in creating service credit", "request", in)
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToCreateBillingCredit, err).Error())
	}
	return &emptypb.Empty{}, nil
}

func (ariaBillingCreditService *AriaBillingCreditService) GetBillingCredit(ctx context.Context, creditData *data.AllCredit, cloudAccountId string) (*pb.BillingCredit, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaBillingCreditService.ReadBillingCredit").WithValues("cloudAccountId", cloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	clientAcctId := client.GetAccountClientId(cloudAccountId)
	creditDetails, err := ariaBillingCreditService.ariaController.GetCreditDetails(ctx, clientAcctId, creditData.CreditNo)
	if err != nil {
		logger.Error(err, "aria get credit detail api error", "readBillingCredit", creditData)
		return nil, status.Errorf(codes.Internal, client.GetDriverError(InvalidDataReturnedForCredits, err).Error())
	}
	var expiryTimestamp *timestamppb.Timestamp
	if creditData.CreditType != CASH_CREDITS {
		expiryTimestamp, err = parseTime(creditDetails.CreditExpiryDate)
		if err != nil {
			logger.Error(err, "aria credit expiry date parse error", "readBillingCredit", creditData)
			return nil, status.Errorf(codes.Internal, client.GetDriverError(InvalidDataReturnedForCredits, err).Error())
		}
	}
	createdDateTimeStamp, err := parseTime(creditData.CreatedDate)
	if err != nil {
		logger.Error(err, "aria create date parse error", "readBillingCredit", creditData)
		return nil, status.Errorf(codes.Internal, "aria created date parse error %v", creditData)
	}
	billingCredit := &pb.BillingCredit{
		OriginalAmount:  float64(creditData.Amount),
		RemainingAmount: float64(creditData.UnappliedAmount),
		Created:         createdDateTimeStamp,
		Expiration:      expiryTimestamp,
		CloudAccountId:  cloudAccountId,
		// todo: we need to support different reasons - this needs to change.
		Reason:     pb.BillingCreditReason_CREDIT_INITIAL,
		CouponCode: creditDetails.Comments,
		AmountUsed: float64(creditData.AppliedAmount),
	}
	return billingCredit, nil
}

func (ariaBillingCreditService *AriaBillingCreditService) ReadInternal(billingAccount *pb.BillingAccount, outStream pb.BillingCreditService_ReadInternalServer) error {
	ctx := outStream.Context()
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaBillingCreditService.ReadInternal").WithValues("cloudAccountId", billingAccount.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	accountCredits, err := ariaBillingCreditService.ariaController.GetAccountServiceCredits(ctx, billingAccount.CloudAccountId)

	if err != nil {
		logger.Error(err, "aria get account credits api error", "billingAccount", billingAccount)
		return status.Errorf(codes.Internal, client.GetDriverError(FailedToReadBillingCredit, err).Error())
	}

	for _, accountCredit := range accountCredits {
		billingCredit, err := ariaBillingCreditService.GetBillingCredit(ctx, &accountCredit, billingAccount.CloudAccountId)
		if err != nil {
			logger.Error(err, "aria read billing credits error", "billingAccount", billingAccount)
			return status.Errorf(codes.Internal, client.GetDriverError(FailedToReadBillingCredit, err).Error())
		}
		if err := outStream.Send(billingCredit); err != nil {
			logger.Error(err, "error sending read details")
			return err
		}
	}
	return nil
}

func parseTime(s string) (*timestamppb.Timestamp, error) {
	layout := "2006-01-02 15:04:05"
	s = strings.TrimSuffix(s, " 00:00:00")
	t, err := time.Parse(layout, s+" 00:00:00")
	if err != nil {
		return nil, err
	}
	ts := timestamppb.New(t)
	return ts, nil
}

func (ariaBillingCreditService *AriaBillingCreditService) ReadUnappliedCreditBalance(ctx context.Context, in *pb.BillingAccount) (*pb.BillingUnappliedCreditBalance, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaBillingCreditService.ReadUnappliedCreditBalance").WithValues("cloudAccountId", in.CloudAccountId).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	unAppliedServiceCredit, err := ariaBillingCreditService.ariaController.GetUnAppliedServiceCredits(ctx, in.CloudAccountId)
	if err != nil {
		logger.Error(err, "error in get unapplied service credit", "request", in)
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToReadUnappliedCreditBalance, err).Error())
	}
	res := &pb.BillingUnappliedCreditBalance{UnappliedAmount: float64(unAppliedServiceCredit)}
	return res, nil
}

type SortedCredits []*pb.BillingCredit

func (credit SortedCredits) Len() int      { return len(credit) }
func (credit SortedCredits) Swap(i, j int) { credit[i], credit[j] = credit[j], credit[i] }

// sort based on expiration time
func (credit SortedCredits) Less(i, j int) bool {
	return credit[i].Created.AsTime().Before(credit[j].Created.AsTime())
}

func (ariaBillingCreditService *AriaBillingCreditService) Read(ctx context.Context, in *pb.BillingCreditFilter) (*pb.BillingCreditResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingCreditService.Read").WithValues("cloudAccountId", in.GetCloudAccountId()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	acctCredits, err := ariaBillingCreditService.ariaController.GetAccountServiceCredits(ctx, in.CloudAccountId)
	if err != nil {
		logger.Error(err, "error in get cloud credits", "billingCreditFilter", in)
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToReadBillingCredits, err).Error())
	}
	var totalUsedAmount, totalRemainingAmount float32 = 0.0, 0.0
	expireTime := time.Now()
	timeNow := time.Now()
	lastUpdated, err := parseTime(expireTime.Format("2006-01-02 00:00:00"))
	if err != nil {
		logger.Error(err, "failed to parse last updated")
		return nil, status.Errorf(codes.Internal, client.GetDriverError(InvalidDataReturnedForCredits, err).Error())
	}
	expirationDate := lastUpdated
	//Store BillingCredits
	billingCredits := []*pb.BillingCredit{}
	for _, val := range acctCredits {
		billingCredit, err := ariaBillingCreditService.GetBillingCredit(ctx, &val, in.CloudAccountId)
		if err != nil {
			logger.Error(err, "aria read billing credits error", "billingCreditFilter", in)
			return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToReadBillingCredits, err).Error())
		}
		creditExpiryTime := billingCredit.Expiration.AsTime()
		//If expiry time of this credit is ahead of current time
		if creditExpiryTime.After(timeNow) {
			totalRemainingAmount += val.UnappliedAmount
			//If expiry time of this credit is ahead of the stored expiry time
			if creditExpiryTime.After(expireTime) {
				expirationDate = timestamppb.New(creditExpiryTime)
				expireTime = creditExpiryTime //Updating the previous time with the new time
			}
		}

		totalUsedAmount += val.AppliedAmount
		billingCredits = append(billingCredits, billingCredit)
	}
	unbilledUsage, err := ariaBillingCreditService.ariaController.GetUnbilledUsage(ctx, in.CloudAccountId)
	if err != nil {
		logger.Error(err, "failed to get unbilled usage")
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToGetUnbilledUsage, err).Error())
	}
	var unAppliedUnbilledUsage = unbilledUsage
	sort.Sort(SortedCredits(billingCredits))
	for _, billingCredit := range billingCredits {
		// only apply unused amount to not expired credits.
		if billingCredit.Expiration.AsTime().After(time.Now()) {
			if billingCredit.RemainingAmount > 0 {
				if unAppliedUnbilledUsage > billingCredit.RemainingAmount {
					unAppliedUnbilledUsage = unAppliedUnbilledUsage - billingCredit.RemainingAmount
					billingCredit.AmountUsed += billingCredit.RemainingAmount
					billingCredit.RemainingAmount = 0

				} else {
					billingCredit.RemainingAmount = billingCredit.RemainingAmount - unAppliedUnbilledUsage
					billingCredit.AmountUsed += unAppliedUnbilledUsage
					unAppliedUnbilledUsage = 0
				}
			}
		}
	}
	// we only subtract what we could deduct from credits.
	totalRemainingAmount = totalRemainingAmount - float32(unbilledUsage-unAppliedUnbilledUsage)
	// we only add what we could deduct from credits.
	totalUsedAmount += float32(unbilledUsage - unAppliedUnbilledUsage)
	res := &pb.BillingCreditResponse{
		TotalRemainingAmount: float64(totalRemainingAmount),
		TotalUsedAmount:      float64(totalUsedAmount),
		LastUpdated:          lastUpdated,
		ExpirationDate:       expirationDate,
		Credits:              billingCredits,
	}
	return res, nil
}
