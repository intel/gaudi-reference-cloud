// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

// AriaPaymentClient - Payment client API for billing to manage the payment methods
type AriaPaymentClient struct {
	ariaClient      *AriaClient
	ariaCredentials *AriaCredentials
}

type CreditCardDetails struct {
	CCNumber      int64
	CCExpireMonth int
	CCExpireYear  int
	CCV           int
}

const (
	/*
		kPaymentMethodsAndTerms = 1	 --> The payments returned "1" will fetch only the payment terms associated with the account no.
		kPaymentMethodsAndTerms = 2	 --> The payments returned "2" will fetch only the payment methods associated with the account no.
		kPaymentMethodsAndTerms = 3	 --> The payments returned "3" will fetch all the payment methods and payment terms associated with the account no.
	*/
	kPaymentMethodsAndTerms = 3

	/*
		kfilterstatus = 0	--> The filter status "0" will fetch only the disabled payment methods associated with the account no.
		kfilterstatus = 1	--> The filter status "1" will fetch only the active payment methods associated with the account no.
		kfilterstatus = 2	--> The filter status "2" will fetch all the payment methods associated with the account no.
	*/
	kfilterstatus = 1
)

func NewAriaPaymentClient(ariaClient *AriaClient, ariaCredentials *AriaCredentials) *AriaPaymentClient {
	return &AriaPaymentClient{
		ariaClient:      ariaClient,
		ariaCredentials: ariaCredentials,
	}
}

func (ariaPaymentClient *AriaPaymentClient) GetPaymentMethods(ctx context.Context, clientAccountId string) (*response.GetPaymentMethodsMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPaymentClient.GetPaymentMethods").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getPaymentmethodsRequest := request.GetPaymentMethodsRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_payment_methods_and_terms_m"},
		OutputFormat:     "json",
		ClientNo:         ariaPaymentClient.ariaCredentials.clientNo,
		AuthKey:          ariaPaymentClient.ariaCredentials.authKey,
		AltCallerId:      AriaClientId,
		ClientAccountId:  clientAccountId,
		PaymentsReturned: kPaymentMethodsAndTerms,
		FilterStatus:     kfilterstatus,
	}

	return CallAria[response.GetPaymentMethodsMResponse](ctx, ariaPaymentClient.ariaClient, &getPaymentmethodsRequest, FailedToGetPaymentMethodsAllError)
}

// Assigning a specified account to a collections account group required to add the payment method
func (ariaPaymentClient *AriaPaymentClient) AssignCollectionsAccountGroup(ctx context.Context, clientAccountId string, clientAcctGroupId string) (*response.AssignCollectionsAccountGroupMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPaymentClient.AssignCollectionsAccountGroup").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	assignCollectionsAccountGroupRequest := request.AssignCollectionsAccountGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "assign_collections_acct_group_m"},
		OutputFormat:      "json",
		ClientNo:          ariaPaymentClient.ariaCredentials.clientNo,
		AuthKey:           ariaPaymentClient.ariaCredentials.authKey,
		AltCallerId:       AriaClientId,
		ClientAccountId:   clientAccountId,
		ClientAcctGroupId: clientAcctGroupId,
	}

	return CallAria[response.AssignCollectionsAccountGroupMResponse](ctx, ariaPaymentClient.ariaClient, &assignCollectionsAccountGroupRequest, FailedToAssignCollectionsAccountGroup)
}

func (ariaPaymentClient *AriaPaymentClient) AddAccountPaymentMethod(ctx context.Context, clientAccountId string, clientPaymentMethodId string, clientBillingGroupId string, payMethodType int, creditCardDetails CreditCardDetails) (*response.UpdateAccountBillingGroupMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPaymentClient.AddAccountPaymentMethod").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	updateAccountBillingGroupRequest := request.UpdateAccountBillingGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "update_acct_billing_group_m"},
		OutputFormat:                 "json",
		ClientNo:                     ariaPaymentClient.ariaCredentials.clientNo,
		AuthKey:                      ariaPaymentClient.ariaCredentials.authKey,
		ClientAccountId:              clientAccountId,
		ClientBillingGroupId:         clientBillingGroupId,
		ClientPrimaryPaymentMethodId: clientPaymentMethodId,
		ClientPaymentMethodId:        clientPaymentMethodId,
		PayMethodType:                payMethodType,
		CCNumber:                     creditCardDetails.CCNumber,
		CCExpireMonth:                creditCardDetails.CCExpireMonth,
		CCExpireYear:                 creditCardDetails.CCExpireYear,
		CCV:                          creditCardDetails.CCV,
	}

	return CallAria[response.UpdateAccountBillingGroupMResponse](ctx, ariaPaymentClient.ariaClient, &updateAccountBillingGroupRequest, FailedToUpdateAccountBillingGroup)
}

func (ariaPaymentClient *AriaPaymentClient) SetSession(ctx context.Context, clientAccountId string) (*response.SetSessionMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPaymentClient.SetSession").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("start: SetSession")

	setSessionRequest := request.SetSessionMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "set_session_m"},
		OutputFormat:    "json",
		ClientNo:        ariaPaymentClient.ariaCredentials.clientNo,
		AuthKey:         ariaPaymentClient.ariaCredentials.authKey,
		AltCallerId:     AriaClientId,
		ClientAccountId: clientAccountId,
	}
	logger.Info("exit: SetSession")
	return CallAria[response.SetSessionMResponse](context.Background(), ariaPaymentClient.ariaClient, &setSessionRequest, FailedToSetSession)
}

func (ariaPaymentClient *AriaPaymentClient) UpdateAccountBillingGroup(ctx context.Context, clientAccountId string, clientBillingGroupId string, primaryPaymentMethodNo int64) (*response.UpdateAccountBillingGroupMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPaymentClient.UpdateAccountBillingGroup").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("start: UpdateAccountBillingGroup")

	updateAccountBillingGroupRequest := request.UpdateAccountBillingGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "update_acct_billing_group_m"},
		OutputFormat:           "json",
		ClientNo:               ariaPaymentClient.ariaCredentials.clientNo,
		AuthKey:                ariaPaymentClient.ariaCredentials.authKey,
		ClientAccountId:        clientAccountId,
		ClientBillingGroupId:   clientBillingGroupId,
		PrimaryPaymentMethodNo: primaryPaymentMethodNo,
	}
	logger.Info("exit: UpdateAccountBillingGroup")

	return CallAria[response.UpdateAccountBillingGroupMResponse](ctx, ariaPaymentClient.ariaClient, &updateAccountBillingGroupRequest, FailedToUpdateAccountBillingGroup)
}

func (ariaPaymentClient *AriaPaymentClient) RemovePaymentMethod(ctx context.Context, clientAccountId string, paymentMethodNo int64) (*response.RemovePaymentMethodMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPaymentClient.RemovePaymentMethod").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	removePaymentmethodRequest := request.RemovePaymentMethodRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "remove_acct_payment_method_m"},
		OutputFormat:    "json",
		ClientNo:        ariaPaymentClient.ariaCredentials.clientNo,
		AuthKey:         ariaPaymentClient.ariaCredentials.authKey,
		AltCallerId:     AriaClientId,
		ClientAccountId: clientAccountId,
		PaymentMethodNo: paymentMethodNo,
	}

	return CallAria[response.RemovePaymentMethodMResponse](ctx, ariaPaymentClient.ariaClient, &removePaymentmethodRequest, FailedToRemovePaymentMethodAllError)
}

func (ariaPaymentClient *AriaPaymentClient) AuthorizeElectronicPayment(ctx context.Context, clientAccountId string, billingGroupNo int64, amount float64) (*response.AuthorizeElectronicPaymentResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPaymentClient.AuthorizeElectronicPayment").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	logger.Info("start: AuthorizeElectronicPayment")
	authorizeElectronicPaymentRequest := request.AuthorizeElectronicPaymentRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "authorize_electronic_payment_m"},
		OutputFormat:    "json",
		ClientNo:        ariaPaymentClient.ariaCredentials.clientNo,
		AuthKey:         ariaPaymentClient.ariaCredentials.authKey,
		AltCallerId:     AriaClientId,
		ClientAccountId: clientAccountId,
		BillingGroupNo:  billingGroupNo,
		Amount:          amount,
	}
	logger.Info("ariaCall finished: authorize_electronic_payment_m")
	return CallAria[response.AuthorizeElectronicPaymentResponse](ctx, ariaPaymentClient.ariaClient, &authorizeElectronicPaymentRequest, FailedToAuthorizeElectronicPaymentError)
}
