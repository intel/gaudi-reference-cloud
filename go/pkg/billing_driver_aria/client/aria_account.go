// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"errors"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/accts"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

/*
Account client API for billing to create account for premium and enterprise user.
*/
type AriaAccountClient struct {
	ariaClient      *AriaClient
	ariaCredentials *AriaCredentials
}

func NewAriaAccountClient(ariaClient *AriaClient, ariaCredentials *AriaCredentials) *AriaAccountClient {
	return &AriaAccountClient{
		ariaClient:      ariaClient,
		ariaCredentials: ariaCredentials,
	}
}

// Todo: Identify and get billing group, Master plan details and Account request fields
func (ariaAccountClient *AriaAccountClient) CreateAndAddDefaultsToAccountRequest(clientAccountId string, clientPlanId string, accountType string) request.CreateAcctCompleteMRequest {
	acct := accts.Acct{
		ClientAcctId:         clientAccountId,
		NotifyMethod:         ACCOUNT_NOTIFY_METHOD,
		Userid:               clientAccountId,
		AcctCurrency:         ACCOUNT_CURRENCY,
		InvoicingOption:      ACCOUNT_INVOICE_OPTION,
		ClientSeqFuncGroupId: ACCOUNT_FUNC_GROUP_ID,
		FunctionalAcctGroup:  []accts.FunctionalAcctGroup{GetFunctionalAccountGroup()},
		MasterPlansDetail:    []accts.MasterPlansDetail{GetMasterPlanDetail(clientPlanId)},
		BillingGroup:         []accts.BillingGroup{GetBillingGroup(clientAccountId, clientPlanId)},
		DunningGroup:         []accts.DunningGroup{GetDunningGroup(clientAccountId, clientPlanId)},
		SuppField:            GetSuppFiled(),
		//TODO: replace with country ISO code from cloud account
		Country: ACCOUNT_COUNTRY,
	}
	var accountTypeValue string

	if accountType == ACCOUNT_TYPE_PREMIUM {
		accountTypeValue = SUPP_FIELD_VALUE_ACCOUNT_TYPE_PREMIUM
	} else if accountType == ACCOUNT_TYPE_ENTERPRISE {
		accountTypeValue = SUPP_FIELD_VALUE_ACCOUNT_TYPE_ENTERPRISE
	} else if accountType == ACCOUNT_TYPE_ENTERPRISE_PENDING {
		accountTypeValue = SUPP_FIELD_VALUE_ACCOUNT_TYPE_ENT_PENDING
	} else {
		accountTypeValue = SUPP_FIELD_VALUE_ACCOUNT_TYPE_UNKNOWN
	}

	suppFieldAccountType := accts.SuppField{
		SuppFieldName:  SUPP_FIELD_NAME_ACCOUNT_TYPE,
		SuppFieldValue: accountTypeValue,
	}

	acct.SuppField = append(acct.SuppField, suppFieldAccountType)

	accounts := []accts.Acct{acct}
	createAccountRequest := request.CreateAcctCompleteMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "create_acct_complete_m"},
		OutputFormat: "json",
		Acct:         accounts,
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey}

	return createAccountRequest
}

func (ariaAccountClient *AriaAccountClient) CreateAriaAccount(ctx context.Context, clientAccountId string, clientPlanId string, acctType string) (*response.CreateAcctCompleteMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.CreateAriaAccount").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	ariaCreateAccountRequest := ariaAccountClient.CreateAndAddDefaultsToAccountRequest(clientAccountId, clientPlanId, acctType)
	return CallAria[response.CreateAcctCompleteMResponse](ctx, ariaAccountClient.ariaClient, &ariaCreateAccountRequest, FailedToCreateAriaAccountError)
}

func (ariaAccountClient *AriaAccountClient) UpdateAriaAccount(ctx context.Context, accountNo int64, seniorAccountNo int64) (*response.UpdateAcctCompleteMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.UpdateAriaAccount").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	updateAcctRequest := request.UpdateAcctCompleteMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "update_acct_complete_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		AcctNo:       accountNo,
		SeniorAcctNo: seniorAccountNo,
	}
	return CallAria[response.UpdateAcctCompleteMResponse](ctx, ariaAccountClient.ariaClient, &updateAcctRequest, FailedToUpdateAccount)
}

func (ariaAccountClient *AriaAccountClient) UpdateAriaAccountStatus(ctx context.Context, accountNo int64, statuscd int64) (*response.UpdateAcctStatusResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.UpdateAriaAccountStatus").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	updateAriaAccountStatusRequest := request.UpdateAcctStatusMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "update_acct_status_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		AccountNo:    accountNo,
		StatusCd:     statuscd,
	}
	return CallAria[response.UpdateAcctStatusResponse](ctx, ariaAccountClient.ariaClient, &updateAriaAccountStatusRequest, FailedToUpdateAccountStatus)
}

// Used to update the templated group id for account
func (ariaAccountClient *AriaAccountClient) SetAccountNotifyTemplateGroup(ctx context.Context, clientAccountId string, notificationTemplateGroupId string) (*response.AriaResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.SetAccountNotifyTemplateGroup").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	setAccoutNotifyTemplateGroupId := request.AccountNotifyTemplateGroup{
		AriaRequest: request.AriaRequest{
			RestCall: "set_acct_notify_tmplt_grp_m"},
		OutputFormat:                "json",
		ClientNo:                    ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:                     ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId:                clientAccountId,
		NotificationTemplateGroupId: notificationTemplateGroupId,
	}
	return CallAria[response.AriaResponse](ctx, ariaAccountClient.ariaClient, &setAccoutNotifyTemplateGroupId, FailedToSetAccountNotifyTemplateGroupError)
}

func (ariaAccountClient *AriaAccountClient) GetAriaAccountDetailsAllForClientId(ctx context.Context, clientAccountID string) (*response.GetAcctDetailsAllMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetAriaAccountDetailsAllForClientId").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getAccountDetailsAllRequest := request.GetAcctDetailsAllMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_details_all_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId: clientAccountID,
	}

	return CallAria[response.GetAcctDetailsAllMResponse](ctx, ariaAccountClient.ariaClient, &getAccountDetailsAllRequest, FailedToGetAriaAccountDetailsAllError)
}

func (ariaAccountClient *AriaAccountClient) CreateBillingGroup(ctx context.Context, clientAcctId string, clientBillingGroupId string) (*response.CreateBillingGroupResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.CreateBillingGroup").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	createBillingGroupRequest := request.CreateBillingGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "create_acct_billing_group_m"},
		OutputFormat:         "json",
		ClientNo:             ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:              ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId:         clientAcctId,
		ClientBillingGroupId: clientBillingGroupId,
		NotifyMethod:         ACCOUNT_NOTIFY_METHOD,
		AltCallerId:          ALT_CALLER_ID,
	}

	return CallAria[response.CreateBillingGroupResponse](ctx, ariaAccountClient.ariaClient, &createBillingGroupRequest, FailedToCreateBillingGroup)
}

func (ariaAccountClient *AriaAccountClient) CreateDunningGroup(ctx context.Context, clientAcctId string, clientDunningGroupId string) (*response.CreateDunningGroupResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.CreateDunningGroup").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	createDunningGroupRequest := request.CreateDunningGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "create_acct_dunning_group_m"},
		OutputFormat:         "json",
		ClientNo:             ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:              ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId:         clientAcctId,
		ClientDunningGroupId: clientDunningGroupId,
		AltCallerId:          ALT_CALLER_ID,
	}

	return CallAria[response.CreateDunningGroupResponse](ctx, ariaAccountClient.ariaClient, &createDunningGroupRequest, FailedToCreateDunningGroup)
}

func (ariaAccountClient *AriaAccountClient) GetBillingGroup(ctx context.Context, clientAcctId string) (*response.GetBillingGroupResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetBillingGroup").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getBillingGroupRequest := request.GetGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_billing_group_details_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId: clientAcctId,
	}

	return CallAria[response.GetBillingGroupResponse](ctx, ariaAccountClient.ariaClient, &getBillingGroupRequest, FailedToGetBillingGroup)
}

func (ariaAccountClient *AriaAccountClient) GetDunningGroup(ctx context.Context, clientAcctId string) (*response.GetDunningGroupResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetDunningGroup").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getDunningGroupRequest := request.GetGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_dunning_group_details_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId: clientAcctId,
	}

	return CallAria[response.GetDunningGroupResponse](ctx, ariaAccountClient.ariaClient, &getDunningGroupRequest, FailedToGetDunningGroup)
}

// For Test
func (ariaAccountClient *AriaAccountClient) GetAndSetAccountRequest(ctx context.Context, clientPlanId string) request.CreateAcctCompleteMRequest {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetAndSetAccountRequest").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	var createAccountRequest = request.CreateAcctCompleteMRequest{}
	createAccountRequest.Acct[0].MasterPlansDetail[0].ClientPlanId = clientPlanId
	return createAccountRequest
}

func (ariaAccountClient *AriaAccountClient) GetAccountNoFromUserId(ctx context.Context, clientAccountId string) (*response.GetAccountNoFromUserIdMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetAccountNoFromUserId").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getAccountNoFromUserIdRequest := request.GetAccountNoFromUserIdMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_no_from_user_id_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		AltCallerId:  AriaClientId,
		UserId:       clientAccountId,
	}

	return CallAria[response.GetAccountNoFromUserIdMResponse](ctx, ariaAccountClient.ariaClient, &getAccountNoFromUserIdRequest, FailedToGetAcctNoFromUserId)
}

func (ariaAccountClient *AriaAccountClient) GetAccountHierarchyDetails(ctx context.Context, accountNo int64) (*response.GetAcctHierarchyDetailsMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetAccountHierarchyDetails").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	const hierarchyFilter int64 = 2
	getAccountHierarchyDetails := request.GetAccountHierachyDetailsRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_hierarchy_details_m"},
		OutputFormat:    "json",
		ClientNo:        ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:         ariaAccountClient.ariaCredentials.authKey,
		AltCallerId:     AriaClientId,
		AcctNo:          accountNo,
		HierarchyFilter: hierarchyFilter,
	}

	return CallAria[response.GetAcctHierarchyDetailsMResponse](ctx, ariaAccountClient.ariaClient, &getAccountHierarchyDetails, FailedToGetAcctHierarchyDetails)
}

func (ariaAccountClient *AriaAccountClient) GetAcctPlansAll(ctx context.Context, accountNo int64, clientPlanId string) (*response.GetAcctPlansAllMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetAcctPlansAll").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	productCatalogPlanFilter := accts.ProductCatalogPlanFilter{ClientPlanId: clientPlanId}
	var productCatalogPlanFilters []accts.ProductCatalogPlanFilter
	productCatalogPlanFilters = append(productCatalogPlanFilters, productCatalogPlanFilter)
	getAcctPlansAllRequest := request.GetAcctPlansAllMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_plans_all_m"},
		OutputFormat:                   "json",
		ClientNo:                       ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:                        ariaAccountClient.ariaCredentials.authKey,
		AcctNo:                         accountNo,
		IncludeServiceSuppFields:       "false",
		IncludeProductFields:           "false",
		IncludePlanServices:            "false",
		IncludeSurcharges:              "false",
		IncludeRateSchedule:            "false",
		IncludeContractAndRolloverInfo: "false",
		IncludeDunningInfo:             "false",
		ProductCatalogPlanFilter:       productCatalogPlanFilters,
	}

	return CallAria[response.GetAcctPlansAllMResponse](ctx, ariaAccountClient.ariaClient, &getAcctPlansAllRequest, FailedToGetAllAcctPlans)
}

func (ariaAccountClient *AriaAccountClient) GetAcctPlans(ctx context.Context, clientAccountId string) (*response.GetAcctPlansMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetAcctPlans").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getAcctPlansRequest := request.GetAcctPlansMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_plans_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId: clientAccountId,
	}

	return CallAria[response.GetAcctPlansMResponse](ctx, ariaAccountClient.ariaClient, &getAcctPlansRequest, FailedToGetAcctPlans)
}

func (ariaAccountClient *AriaAccountClient) AssignPlanToAccountWithBillingAndDunningGroup(ctx context.Context, clientAccountId string, newClientPlanId string,
	clientPlanInstanceId string, clientAltRateScheduleId string, billLagDays int64, altBillDay int64) (*response.AssignAcctPlanMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.AssignPlanToAccountWithBillingAndDunningGroup").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	assignPlanToAccountRequest := ariaAccountClient.getAssignPlanToAccountRequest(clientAccountId, newClientPlanId, clientPlanInstanceId, clientAltRateScheduleId, billLagDays, altBillDay)
	billingGrp := GetBillingGroup(clientAccountId, newClientPlanId)
	dunningGrp := GetDunningGroup(clientAccountId, newClientPlanId)
	assignPlanToAccountRequest.ExistingClientBillingGroupId = billingGrp.ClientBillingGroupId
	assignPlanToAccountRequest.ExistingClientDefDunningGroupId = dunningGrp.ClientDunningGroupId
	return CallAria[response.AssignAcctPlanMResponse](ctx, ariaAccountClient.ariaClient, &assignPlanToAccountRequest, FailedToAssignPlanToAccount)
}

// we could have a common methdo to have the request generated but that's a overkill.
func (ariaAccountClient *AriaAccountClient) AssignPlanToPremiumAcct(ctx context.Context, clientAccountId string, newClientPlanId string,
	clientPlanInstanceId string, clientAltRateScheduleId string, billLagDays int64, altBillDay int64, matchingPlanInstanceId string) (*response.AssignAcctPlanMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.AssignPlanToPremiumAcct").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	assignPlanToAccountRequest := ariaAccountClient.getAssignPlanToAccountRequest(clientAccountId, newClientPlanId, clientPlanInstanceId, clientAltRateScheduleId, billLagDays, altBillDay)
	billingGrp := GetBillingGroup(clientAccountId, newClientPlanId)
	dunningGrp := GetDunningGroup(clientAccountId, newClientPlanId)
	assignPlanToAccountRequest.ExistingClientBillingGroupId = billingGrp.ClientBillingGroupId
	assignPlanToAccountRequest.ExistingClientDefDunningGroupId = dunningGrp.ClientDunningGroupId
	assignPlanToAccountRequest.OverrideDatesClientMpInstanceId = matchingPlanInstanceId
	assignPlanToAccountRequest.AssignmentDirective = PREMIUM_ACCOUNT_ASSIGNMENT_DIRECTIVE
	assignPlanToAccountRequest.InvoicingOption = PREMIUM_ACCOUNT_PLAN_INVOICING_OPTION
	assignPlanToAccountRequest.StatusUntilAltStart = 0
	return CallAria[response.AssignAcctPlanMResponse](ctx, ariaAccountClient.ariaClient, &assignPlanToAccountRequest, FailedToAssignPlanToAccount)
}

func (ariaAccountClient *AriaAccountClient) getAssignPlanToAccountRequest(clientAccountId string, newClientPlanId string, clientPlanInstanceId string, clientAltRateScheduleId string, billLagDays int64, altBillDay int64) request.AssignAcctPlanMRequest {
	assignPlanToAccountRequest := request.AssignAcctPlanMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "assign_acct_plan_m"},
		OutputFormat:            "json",
		ClientNo:                ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:                 ariaAccountClient.ariaCredentials.authKey,
		AltCallerId:             AriaClientId,
		ClientAcctId:            clientAccountId,
		NewClientPlanId:         newClientPlanId,
		ClientPlanInstanceId:    clientPlanInstanceId,
		ClientAltRateScheduleId: clientAltRateScheduleId,
		BillLagDays:             billLagDays,
		StatusUntilAltStart:     STATUS_UNTIL_ALT_START,
	}
	if altBillDay != 0 {
		assignPlanToAccountRequest.AltBillDay = altBillDay
	}
	return assignPlanToAccountRequest
}

func (ariaAccountClient *AriaAccountClient) GetAccountCredits(ctx context.Context, accountNo int64) (*response.GetAccountCredits, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetAccountCredits").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getAccountCreditsRequest := request.GetAccountCredits{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_credits_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		AcctNo:       accountNo,
		AltCallerId:  AriaClientId,
	}
	return CallAria[response.GetAccountCredits](ctx, ariaAccountClient.ariaClient, &getAccountCreditsRequest, FailedToGetAccountCreditsError)
}

func (ariaAccountClient *AriaAccountClient) GetAccountCreditDetails(ctx context.Context, clientAcctId string, creditNo int64) (*response.GetCreditDetails, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetAccountCreditDetails").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getAccountCreditDetailsRequest := request.GetCreditDetails{
		AriaRequest: request.AriaRequest{
			RestCall: "get_credit_details_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId: clientAcctId,
		CreditNo:     creditNo,
		AltCallerId:  AriaClientId,
	}
	return CallAria[response.GetCreditDetails](ctx, ariaAccountClient.ariaClient, &getAccountCreditDetailsRequest, FailedToGetCreditDetailsError)
}

func (ariaAccountClient *AriaAccountClient) GetAccountNotificationDetails(ctx context.Context, clientAcctId string) (*response.GetAccountNotificationDetails, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.GetAccountNotificationDetails").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getAccountNotificationDetailsRequest := request.GetAccountNotificationDetails{
		AriaRequest: request.AriaRequest{
			RestCall: "get_acct_notification_details_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId: clientAcctId,
		AltCallerId:  AriaClientId,
	}
	return CallAria[response.GetAccountNotificationDetails](ctx, ariaAccountClient.ariaClient, &getAccountNotificationDetailsRequest, FailedToGetAccountNotificationDetailsError)
}

func (ariaAccountClient *AriaAccountClient) AssignPlanToEnterpriseChildAccount(ctx context.Context, clientAccountId string, newClientPlanId string,
	clientPlanInstanceId string, clientAltRateScheduleId string, billLagDays int64, altBillDay int64, overrideBillThruDate string, invoicingOption int64, respLevelCd int64, parentClientPlanInstanceId string) (*response.AssignAcctPlanMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.AssignPlanToEnterpriseChildAccount").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	assignPlanToAccountRequest := request.AssignAcctPlanMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "assign_acct_plan_m"},
		OutputFormat:            "json",
		ClientNo:                ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:                 ariaAccountClient.ariaCredentials.authKey,
		AltCallerId:             AriaClientId,
		ClientAcctId:            clientAccountId,
		NewClientPlanId:         newClientPlanId,
		ClientPlanInstanceId:    clientPlanInstanceId,
		ClientAltRateScheduleId: clientAltRateScheduleId,
	}
	assignPlanToAccountRequest.OverrideBillThruDate = overrideBillThruDate
	assignPlanToAccountRequest.InvoicingOption = invoicingOption
	assignPlanToAccountRequest.RespLevelCd = respLevelCd
	assignPlanToAccountRequest.RetroactiveStartDate = overrideBillThruDate
	assignPlanToAccountRequest.AssignmentDirective = ENTERPRISE_ACCOUNT_ASSIGNMENT_DIRECTIVE
	assignPlanToAccountRequest.RespClientMasterPlanInstanceId = parentClientPlanInstanceId
	assignPlanToAccountRequest.BillLagDays = billLagDays
	return CallAria[response.AssignAcctPlanMResponse](ctx, ariaAccountClient.ariaClient, &assignPlanToAccountRequest, FailedToAssignPlanToAccount)
}

func (ariaAccountClient *AriaAccountClient) GetEnterpriseParentAccountPlan(ctx context.Context, clientAccountId string) ([]data.MasterPlansInfo, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.CheckAndGetEnterpriseAccount").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getAcctNoResponse, err := ariaAccountClient.GetAccountNoFromUserId(ctx, clientAccountId)
	if err != nil {
		logger.Error(err, "failed to get account number from client account", "id", clientAccountId)
		return nil, err
	}

	getAcctHierarchyResponse, err := ariaAccountClient.GetAccountHierarchyDetails(ctx, getAcctNoResponse.AcctNo)
	if err != nil {
		logger.Error(err, "failed to get acct hierarchy details")
		return nil, err
	}
	acctHierarchyMap := map[int64][]data.MasterPlansInfo{}
	seniorAcctNo := int64(0)
	const nilSeniorAcctNo = 0
	for _, acctHierarchy := range getAcctHierarchyResponse.AcctHierarchyDtls {
		if len(acctHierarchy.ChildAcctNo) > 0 {
			acctHierarchyMap[acctHierarchy.AcctNo] = acctHierarchy.MasterPlansInfo
		}
		if acctHierarchy.Userid == clientAccountId && acctHierarchy.SeniorAcctNo != nilSeniorAcctNo {
			seniorAcctNo = acctHierarchy.SeniorAcctNo
		}
	}
	if seniorAcctNo != nilSeniorAcctNo {
		return acctHierarchyMap[seniorAcctNo], nil
	}
	return nil, errors.New("warning: enterpise pending client account id" + clientAccountId)
}

func (ariaAccountClient *AriaAccountClient) UpdateAccountDunningGroup(ctx context.Context, clientAcctId string, clientDunningGroupId string, clientDunningProcessId string) (*response.CreateDunningGroupResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.UpdateAccountDunningGroup").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	createDunningGroupRequest := request.CreateDunningGroupRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "update_acct_dunning_group_m"},
		OutputFormat:           "json",
		ClientNo:               ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:                ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId:           clientAcctId,
		ClientDunningGroupId:   clientDunningGroupId,
		ClientDunningProcessId: clientDunningProcessId,
		AltCallerId:            ALT_CALLER_ID,
	}
	return CallAria[response.CreateDunningGroupResponse](ctx, ariaAccountClient.ariaClient, &createDunningGroupRequest, FailedToCreateDunningGroup)
}

func (ariaAccountClient *AriaAccountClient) UpdateAccountContact(ctx context.Context, clientAcctId string, email string) (*response.UpdateAccountContact, error) {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("AriaAccountClient.UpdateAccountContact").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	updateAccountContactRequest := request.UpdateAccountContact{
		AriaRequest: request.AriaRequest{
			RestCall: "update_contact_m"},
		OutputFormat: "json",
		ClientNo:     ariaAccountClient.ariaCredentials.clientNo,
		AuthKey:      ariaAccountClient.ariaCredentials.authKey,
		ClientAcctId: clientAcctId,
		Email:        email,
		ContactInd:   CONTACT_INDICATOR,
		AltCallerId:  ALT_CALLER_ID,
	}
	return CallAria[response.UpdateAccountContact](ctx, ariaAccountClient.ariaClient, &updateAccountContactRequest, FailedToUpdateAccountContact)
}
