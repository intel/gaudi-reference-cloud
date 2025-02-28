// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package aria

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AriaBillingOptionService struct {
	ariaController *AriaController
	pb.UnimplementedBillingOptionServiceServer
}

const (
	PaymentMethodType_CREDITCARD    int64 = 1
	PaymentMethod_ENABLED           int64 = 1
	PaymentMethodType_UNSPECIFIED   int64 = -1
	PaymentMethodNumber_UNSPECIFIED int64 = -1
)

func (ariaBillingOptionService *AriaBillingOptionService) Read(ctx context.Context, filter *pb.BillingOptionFilter) (*pb.BillingOption, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingOptionService.Read").WithValues("cloudAccountId", filter.GetCloudAccountId()).Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	accountPaymentMethods, primaryPaymentMethodNo, err := ariaBillingOptionService.getAccountPaymentMethods(ctx, filter)
	if err != nil {
		logger.Error(err, "error in getAccountPaymentMethods")
		return nil, status.Errorf(codes.Internal, client.GetDriverError(FailedToReadBillingOptions, err).Error())
	}

	logger.Info("accountPaymentMethods", "total accountPaymentMethods", len(accountPaymentMethods))

	for i := 0; i < len(accountPaymentMethods); i++ {
		accountPaymentMethod := accountPaymentMethods[i]

		// The service should only return information related to the Active primaryPaymentMethodNo.
		if accountPaymentMethod.PaymentMethodType == PaymentMethodType_CREDITCARD && (accountPaymentMethod.Status != PaymentMethod_ENABLED || accountPaymentMethod.PaymentMethodNo != primaryPaymentMethodNo) {
			logger.Info("Skipping", "PaymentMethodNo", accountPaymentMethod.PaymentMethodNo, "Status", accountPaymentMethod.Status)
			continue
		}

		// Service should return first record to caller
		return createPaymentMethodDetails(filter, accountPaymentMethod), nil
	}
	logger.Info("returning empty BillingOption")
	return &pb.BillingOption{}, nil
}

func (ariaBillingOptionService *AriaBillingOptionService) getAccountPaymentMethods(ctx context.Context, filter *pb.BillingOptionFilter) (accountPaymentMethods []data.AccountPaymentMethods, primaryPaymentMethodNo int64, errors error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaBillingOptionService.getAccountPaymentMethods").WithValues("cloudAccountId", filter.GetCloudAccountId()).Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")

	accountPaymentMethods, err := ariaBillingOptionService.ariaController.GetBillingOptions(ctx, *filter.CloudAccountId)

	if err != nil {
		logger.Error(err, "aria read billing options api error")
		return nil, PaymentMethodNumber_UNSPECIFIED, err
	}

	if len(accountPaymentMethods) == 0 {
		logger.Info("failed to get account account payment methods")
		return nil, PaymentMethodNumber_UNSPECIFIED, nil
	}

	clientBillingGroupResponse, err := ariaBillingOptionService.ariaController.ariaAccountClient.GetBillingGroup(ctx, client.GetAccountClientId(*filter.CloudAccountId))
	if err != nil {
		logger.Error(err, "failed to get account billing group")
		return nil, PaymentMethodNumber_UNSPECIFIED, err
	}

	//TODO: support for multiple billing groups
	if clientBillingGroupResponse != nil {
		primaryPaymentMethodNo = clientBillingGroupResponse.BillingGroupDetails[0].PrimaryPaymentMethodNo
		logger.Info("PrimaryPaymentMethodNo", "PrimaryPaymentMethodNo", primaryPaymentMethodNo)
		return accountPaymentMethods, primaryPaymentMethodNo, nil
	}
	return nil, PaymentMethodNumber_UNSPECIFIED, nil
}

func createPaymentMethodDetails(in *pb.BillingOptionFilter, accountPaymentMethod data.AccountPaymentMethods) *pb.BillingOption {
	paymentMethodDetails := &pb.BillingOption{
		Id:             uint64(uuid.New().ID()),
		CloudAccountId: *in.CloudAccountId,
		FirstName:      accountPaymentMethod.BillFirstName,
		MiddleInitial:  accountPaymentMethod.BillMiddelInitial,
		LastName:       accountPaymentMethod.BillLastName,
		Email:          accountPaymentMethod.BillEmail,
	}

	if accountPaymentMethod.PaymentMethodType == PaymentMethodType_CREDITCARD {
		paymentMethodDetails.PaymentType = pb.PaymentType_PAYMENT_CREDIT_CARD
		paymentMethodDetails.PaymentMethod = &pb.BillingOption_CreditCard{CreditCard: &pb.CreditCard{
			Suffix: accountPaymentMethod.Suffix,
			Expiration: strconv.Itoa(int(accountPaymentMethod.CCExpireMonth)) +
				"/" + strconv.Itoa(int(accountPaymentMethod.CCExpireYear)),
			Type: accountPaymentMethod.CCType,
		}}
	} else if accountPaymentMethod.PaymentMethodType == PaymentMethodType_UNSPECIFIED {
		paymentMethodDetails.PaymentType = pb.PaymentType_PAYMENT_UNSPECIFIED
	} else {
		paymentMethodDetails.PaymentType = pb.PaymentType_PAYMENT_OTHER
		paymentMethodDetails.PaymentMethod = &pb.BillingOption_Other{}
	}
	return paymentMethodDetails
}
