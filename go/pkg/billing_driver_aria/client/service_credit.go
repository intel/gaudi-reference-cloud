// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"encoding/json"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

// TODO: service_code_option needs to be sorted out
var createAriaCloudCreditsRequestDefaults = []byte(`
rest_call: create_advanced_service_credit_m
output_format: json
service_code_option: 2
credit_expiry_type_ind: D
`)

// ServiceCreditClient - Credit client API for billing to create cloud credits for premium and enterprise users
type ServiceCreditClient struct {
	ariaClient              *AriaClient
	ariaCredentials         *AriaCredentials
	ariaCreateCreditRequest *request.CreateAdvancedServiceCreditMRequest
}

func NewServiceCreditClient(ariaClient *AriaClient, ariaCredentials *AriaCredentials) *ServiceCreditClient {
	var createCreditRequest = request.CreateAdvancedServiceCreditMRequest{}
	return &ServiceCreditClient{
		ariaClient:              ariaClient,
		ariaCredentials:         ariaCredentials,
		ariaCreateCreditRequest: &createCreditRequest,
	}
}

func (serviceCreditClient *ServiceCreditClient) CreateAndAddDefaultsToCreditRequest(ctx context.Context, clientAccountId string, amount float64, reasonCode int64, creditExpiryDate string, comments string) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ServiceCreditClient.CreateAndAddDefaultsToCreditRequest").Start()
	defer span.End()
	logger.V(2).Info("BEGIN")
	defer logger.V(2).Info("END")
	var createCreditRequest = request.CreateAdvancedServiceCreditMRequest{}
	createAriaCreditRequestJson, err := ConvertYamlToJson(ctx, createAriaCloudCreditsRequestDefaults, "credit")
	if err != nil {
		logger.Error(err, "failed to convert for create cloud credits request", "createAriaCreditRequestJson", createAriaCloudCreditsRequestDefaults, "context", "ConvertYamlToJson")
		return err
	}
	logger.V(2).Info("Create credits request json ", string(createAriaCreditRequestJson))
	if err = json.Unmarshal(createAriaCreditRequestJson, &createCreditRequest); err != nil {
		logger.Error(err, "failed to JSON unmarshal default values for create cloud credits request")
		return err
	}
	createCreditRequest.ClientNo = serviceCreditClient.ariaCredentials.clientNo
	createCreditRequest.AuthKey = serviceCreditClient.ariaCredentials.authKey
	createCreditRequest.AltCallerId = AriaClientId
	createCreditRequest.ClientAcctId = clientAccountId
	createCreditRequest.Amount = amount
	createCreditRequest.ReasonCode = reasonCode
	createCreditRequest.CreditExpiryDate = creditExpiryDate
	createCreditRequest.Comments = comments
	serviceCreditClient.ariaCreateCreditRequest = &createCreditRequest
	return nil
}

func (serviceCreditClient *ServiceCreditClient) CreateServiceCredits(ctx context.Context, clientAccountId string, amount float64, reasonCode int64, creditExpiryDate string, comments string) (*response.CreateAdvancedServiceCreditMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ServiceCreditClient.CreateServiceCredits").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	if err := serviceCreditClient.CreateAndAddDefaultsToCreditRequest(ctx, clientAccountId, amount, reasonCode, creditExpiryDate, comments); err != nil {
		logger.Error(err, "failed to create aria credit request", "clientAccountId", clientAccountId, "amount", amount, "reasonCode", reasonCode, "creditExpiryDate", creditExpiryDate, "comments", comments, "context", "CreateAndAddDefaultsToCreditRequest")
		return nil, err
	}

	return CallAria[response.CreateAdvancedServiceCreditMResponse](ctx, serviceCreditClient.ariaClient, serviceCreditClient.ariaCreateCreditRequest, FailedToCreateCloudCreditError)
}

func (serviceCreditClient *ServiceCreditClient) GetUnappliedServiceCredits(ctx context.Context, clientAccountId string) (*response.GetUnappliedServiceCreditsMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ServiceCreditClient.GetUnappliedServiceCredits").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getCreditDetailsAllRequest := request.GetUnappliedServiceCreditsMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_unapplied_service_credits_m"},
		OutputFormat: "json",
		ClientNo:     serviceCreditClient.ariaCredentials.clientNo,
		AuthKey:      serviceCreditClient.ariaCredentials.authKey,
		ClientAcctId: clientAccountId,
	}

	return CallAria[response.GetUnappliedServiceCreditsMResponse](ctx, serviceCreditClient.ariaClient, &getCreditDetailsAllRequest, FailedToGetUnappliedServiceCreditsError)
}

// Function to call Aria API (apply_service_credit_m) to apply credits to aria account
func (serviceCreditClient *ServiceCreditClient) ApplyCreditService(ctx context.Context, clientAcctId string, creditAmount float64) (*response.ApplyServiceCreditResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("ServiceCreditClient.ApplyCreditService").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	applyCreditServiceRequest := request.ApplyServiceCredit{
		AriaRequest: request.AriaRequest{
			RestCall: "apply_service_credit_m"},
		OutputFormat: "json",
		ClientNo:     serviceCreditClient.ariaCredentials.clientNo,
		AuthKey:      serviceCreditClient.ariaCredentials.authKey,
		ClientAcctId: clientAcctId,
		CreditAmount: creditAmount,
		//	CreditReasonCode: 1 --> General Credit/others
		CreditReasonCode: 1,
	}
	return CallAria[response.ApplyServiceCreditResponse](ctx, serviceCreditClient.ariaClient, &applyCreditServiceRequest, FailedToApplyServiceCredit)
}
