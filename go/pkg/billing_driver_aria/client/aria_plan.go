// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"encoding/json"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/plans"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"gopkg.in/yaml.v3"
)

var createPlanYaml = []byte(`
output_format: json
plan_type: Master Recurring Plan
active: 1
currency: usd
service:
  - service_type: Usage-Based
    gl_cd: "1"

`)

var deactivatePlanYaml = []byte(`
output_format: json
rest_call: edit_plan_m
plan_type: Master Recurring Plan
active: 0
currency: usd
edit_directives: 2
`)

type AriaPlanClient struct {
	config                    *config.Config
	createPlanRequestJson     []byte
	deactivatePlanRequestJson []byte
	ariaAdminClient           *AriaAdminClient
	ariaClient                *AriaClient
	ariaCredentials           *AriaCredentials
}

func NewAriaPlanClient(config *config.Config, ariaAdminClient *AriaAdminClient, ariaClient *AriaClient, ariaCredentials *AriaCredentials) *AriaPlanClient {
	logger := log.FromContext(context.Background()).WithName("AriaPlanClient.NewAriaPlanClient")
	createPlan := map[any]any{}
	if err := yaml.Unmarshal(createPlanYaml, &createPlan); err != nil {
		logger.Error(err, "failed to unmarshal YAML default values for create plan request")
		return nil
	}
	createPlanRequestJson, err := json.Marshal(mapKeysToString(createPlan))
	if err != nil {
		logger.Error(err, "failed to JSON marshal default values for create plan request")
		return nil
	}
	deactivatePlan := map[any]any{}
	if err := yaml.Unmarshal(deactivatePlanYaml, &deactivatePlan); err != nil {
		logger.Error(err, "failed to unmarshal YAML default values for deactivate plan request")
		return nil
	}
	deactivatePlanRequestJson, err := json.Marshal(mapKeysToString(deactivatePlan))
	if err != nil {
		logger.Error(err, "failed to JSON marshal default values for deactivate plan request")
		return nil
	}
	return &AriaPlanClient{
		config:                    config,
		createPlanRequestJson:     createPlanRequestJson,
		deactivatePlanRequestJson: deactivatePlanRequestJson,
		ariaAdminClient:           ariaAdminClient,
		ariaClient:                ariaClient,
		ariaCredentials:           ariaCredentials,
	}
}

func (aPC *AriaPlanClient) generateRequest(ctx context.Context, restCall string, product *pb.Product, productFamily *pb.ProductFamily, usageType *data.UsageType) (*request.PlanRequest, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.NewAriaPlanClient").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	planRequest, err := aPC.newPlanRequest(ctx)
	if err != nil {
		return nil, err
	}
	err = GetPlanRequestData(ctx, planRequest, product, productFamily, usageType, aPC.config)
	if err != nil {
		return nil, err
	}
	planRequest.RestCall = restCall
	return planRequest, err

}

func (aPC *AriaPlanClient) CreatePlan(ctx context.Context, product *pb.Product, productFamily *pb.ProductFamily, usageType *data.UsageType) (*response.PlanResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.CreatePlan").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	createPlanRequest, err := aPC.generateRequest(ctx, "create_new_plan_m", product, productFamily, usageType)
	suppOpjFields := GetSupplementalObjField(product)
	suppFieldSync := plans.SupplementalObjectField{
		FieldName:  PLANSUPPFIELDSYNC,
		FieldValue: []string{PLANSUPPFIELDSYNCVALUE},
	}
	suppOpjFields = append(suppOpjFields, suppFieldSync)
	createPlanRequest.SupplementalObjField = suppOpjFields
	if err != nil {
		logger.Error(err, "failed to generate request for create plan")
		return nil, err
	}
	return CallAriaAdmin[response.PlanResponse](ctx, aPC.ariaAdminClient, createPlanRequest, FailedToCreatePlanError)
}

func (aPC *AriaPlanClient) EditPlan(ctx context.Context, product *pb.Product, productFamily *pb.ProductFamily, usageType *data.UsageType) (*response.PlanResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.EditPlan").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	editPlanRequest, err := aPC.generateRequest(ctx, "edit_plan_m", product, productFamily, usageType)
	editPlanRequest.EditDirectives = PLAN_EDIT_DIRECTIVE
	editPlanRequest.SupplementalObjField = GetSupplementalObjField(product)
	if err != nil {
		logger.Error(err, "failed to generate request for edit plan")
		return nil, err
	}
	return CallAriaAdmin[response.PlanResponse](ctx, aPC.ariaAdminClient, editPlanRequest, FailedToEditPlanError)
}

func (aPC *AriaPlanClient) CreateDefaultPlan(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.CreateDefaultPlan").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	usageTypeClient := NewAriaUsageTypeClient(aPC.ariaAdminClient, aPC.ariaCredentials)
	usageType, err := usageTypeClient.GetMinutesUsageType(ctx)
	if err != nil {
		return err
	}
	req, err := aPC.newPlanRequest(ctx)
	if err != nil {
		return err
	}
	service := plans.Service{
		Name:        "Master Service",
		ServiceType: USAGE_BASED,
		GlCd:        GLCD,
		UsageType:   usageType.UsageTypeNo,
		RateType:    TIERED_PRICING,
		PricingRule: "Standard",
	}
	tierSched := plans.TierSchedule{
		From:   1,
		Amount: 0,
	}
	tier := plans.Tier{
		Schedule: []plans.TierSchedule{tierSched},
	}
	service.Tier = []plans.Tier{tier}
	req.Service = []plans.Service{service}

	sched := plans.Schedule{
		ScheduleName: "default",
		CurrencyCd:   CURRENCY_CD,
		IsDefault:    ISDEFAULT,
	}
	req.PlanName = "IDC Master Plan"
	req.Schedule = []plans.Schedule{sched}
	req.ClientPlanId = GetDefaultPlanClientId()
	req.SupplementalObjField = []plans.SupplementalObjectField{
		{FieldName: "SYNC TO SAP GTS", FieldValue: []string{"No"}},
	}
	_, err = CallAriaAdmin[response.PlanResponse](ctx, aPC.ariaAdminClient, req, FailedToCreatePlanError)
	return err
}

func (aPC *AriaPlanClient) newPlanRequest(ctx context.Context) (*request.PlanRequest, error) {
	createPlanRequest := request.PlanRequest{}
	createPlanRequest.RestCall = "create_new_plan_m"
	logger := log.FromContext(ctx).WithName("AriaPlanClient.newPlanRequest")
	err := json.Unmarshal(aPC.createPlanRequestJson, &createPlanRequest)
	if err != nil {
		logger.Error(err, "failed to JSON unmarshal default values for create plan request")
		return nil, err
	}
	createPlanRequest.ClientNo = aPC.ariaCredentials.clientNo
	createPlanRequest.AuthKey = aPC.ariaCredentials.authKey

	return &createPlanRequest, err
}

// DeactivatePlan accept product and productFamily proto structure and map deactivatePlan's parameters
// Aria API requires these parameters to be set in order to deactivate the plan, just passing `active = 0` doesnt suffice
func (aPC *AriaPlanClient) DeactivatePlan(ctx context.Context, product *pb.Product, productFamily *pb.ProductFamily, usageType *data.UsageType) (*response.PlanResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.DeactivatePlan").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	deactivatePlanRequest := request.PlanRequest{}
	err := json.Unmarshal(aPC.deactivatePlanRequestJson, &deactivatePlanRequest)
	if err != nil {
		logger.Error(err, "failed to JSON unmarshal default values for deactivate plan request")
		return nil, err
	}
	deactivatePlanRequest.ClientNo = aPC.ariaCredentials.clientNo
	deactivatePlanRequest.AuthKey = aPC.ariaCredentials.authKey
	err = GetPlanRequestData(ctx, &deactivatePlanRequest, product, productFamily, usageType, aPC.config)
	if err != nil {
		logger.Error(err, "failed to get service and schedule data")
		return nil, err
	}
	return CallAriaAdmin[response.PlanResponse](ctx, aPC.ariaAdminClient, &deactivatePlanRequest, FailedToDeactivatePlanError)
}

func (aPC *AriaPlanClient) DeactivatePlanFromClientPlanDetail(ctx context.Context, clientPlanDetail *data.AllClientPlanDtl, usageType *data.UsageType) (*response.PlanResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.DeactivatePlanFromClientPlanDetail").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	deactivatePlanRequest := request.PlanRequest{}
	err := json.Unmarshal(aPC.deactivatePlanRequestJson, &deactivatePlanRequest)
	if err != nil {
		logger.Error(err, "failed to JSON unmarshal default values for deactivate plan request")
		return nil, err
	}
	deactivatePlanRequest.ClientNo = aPC.ariaCredentials.clientNo
	deactivatePlanRequest.AuthKey = aPC.ariaCredentials.authKey
	GetPlanRequestDataFromClientPlanDetail(ctx, &deactivatePlanRequest, clientPlanDetail, usageType, aPC.config)
	return CallAriaAdmin[response.PlanResponse](ctx, aPC.ariaAdminClient, &deactivatePlanRequest, FailedToDeactivatePlanError)
}

// DeletePlans accept array of PlanNos on which those PlanNos will be deleted from AriaSystem
func (aPC *AriaPlanClient) DeletePlans(ctx context.Context, planNos []int) (*response.DeletePlansResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.DeletePlans").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	deletePlansRequest := request.DeletePlans{
		AriaRequest: request.AriaRequest{
			RestCall: "delete_plans_m"}, // "delete_plans_m" is the AriaSystem's API
		OutputFormat: "json",
	}
	deletePlansRequest.ClientNo = aPC.ariaCredentials.clientNo
	deletePlansRequest.AuthKey = aPC.ariaCredentials.authKey
	deletePlansRequest.PlanNos = make([]int, len(planNos))
	copy(deletePlansRequest.PlanNos, planNos)
	return CallAriaAdmin[response.DeletePlansResponse](ctx, aPC.ariaAdminClient, &deletePlansRequest, FailedToDeletePlansError)
}

func (aPC *AriaPlanClient) GetAriaPlanDetails(ctx context.Context, clientPlanId string) (*response.GetPlanDetailResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.GetAriaPlanDetails").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getPlanDetalsRequest := request.GetPlanDetails{
		AriaRequest: request.AriaRequest{
			RestCall: "get_plan_details_m"},
		GetAllClientPlans: request.GetAllClientPlans{
			OutputFormat: "json",

			ClientNo:     aPC.ariaCredentials.clientNo,
			AuthKey:      aPC.ariaCredentials.authKey,
			ClientPlanId: clientPlanId,
		},
		IncludeRateScheduleSummary: "true",
	}
	return CallAriaAdmin[response.GetPlanDetailResponse](ctx, aPC.ariaAdminClient, &getPlanDetalsRequest, FailedToGetClientPlanDetailsError)
}

func (aPC *AriaPlanClient) GetAllClientPlansForPromoCode(ctx context.Context, promoCode string) (*response.GetClientPlansAllMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.GetAllClientPlansForPromoCode").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getPlansRequest := request.GetAllClientPlans{
		AriaRequest: request.AriaRequest{
			RestCall: "get_client_plans_all_m"},
		OutputFormat: "json",
		ClientNo:     aPC.ariaCredentials.clientNo,
		AuthKey:      aPC.ariaCredentials.authKey,
		AltCallerId:  AriaClientId,
		PromoCode:    promoCode,
	}

	return CallAria[response.GetClientPlansAllMResponse](ctx, aPC.ariaClient, &getPlansRequest, FailedToGetPlansError)
}

// Get the plans for a client defined plan ID.
func (aPC *AriaPlanClient) GetAriaPlanDetailsAllForClientPlanId(ctx context.Context, clientPlanId string) (*response.GetClientPlansAllMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.GetAriaPlanDetailsAllForClientPlanId").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getClientPlansAllRequest := request.GetAllClientPlans{
		AriaRequest: request.AriaRequest{
			RestCall: "get_client_plans_all_m"},
		OutputFormat: "json",
		ClientNo:     aPC.ariaCredentials.clientNo,
		AuthKey:      aPC.ariaCredentials.authKey,
		ClientPlanId: clientPlanId,
	}

	return CallAria[response.GetClientPlansAllMResponse](ctx, aPC.ariaClient, &getClientPlansAllRequest, FailedToGetClientPlansAllError)
}

func (aPC *AriaPlanClient) GetClientPlanServiceRates(ctx context.Context, clientPlanId string, clientServiceId string) (*response.GetClientPlanServiceRates, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.GetClientPlanServiceRates").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	getClientPlanServiceRates := request.GetClientPlanServiceRates{
		AriaRequest: request.AriaRequest{
			RestCall: "get_client_plan_service_rates_m"},
		OutputFormat:    "json",
		ClientNo:        aPC.ariaCredentials.clientNo,
		AuthKey:         aPC.ariaCredentials.authKey,
		ClientPlanId:    clientPlanId,
		ClientServiceId: clientServiceId,
	}
	return CallAria[response.GetClientPlanServiceRates](ctx, aPC.ariaClient, &getClientPlanServiceRates, FailedToGetClientPlanServiceRatesError)
}

// For Testing
func (aPC *AriaPlanClient) CreateTestPlan(ctx context.Context, serviceName string, scheduleName string, planName string, clientPlanId string) (*response.PlanResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("AriaPlanClient.CreateTestPlan").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	usageTypeClient := NewAriaUsageTypeClient(aPC.ariaAdminClient, aPC.ariaCredentials)
	usageType, err := usageTypeClient.GetMinutesUsageType(ctx)
	if err != nil {
		return nil, err
	}
	req, err := aPC.newPlanRequest(ctx)
	if err != nil {
		return nil, err
	}
	service := plans.Service{
		Name:        serviceName,
		ServiceType: USAGE_BASED,
		GlCd:        GLCD,
		UsageType:   usageType.UsageTypeNo,
		RateType:    TIERED_PRICING,
		PricingRule: "Standard",
	}
	tierSched := plans.TierSchedule{
		From:   1,
		Amount: 0,
	}
	tier := plans.Tier{
		Schedule: []plans.TierSchedule{tierSched},
	}
	service.Tier = []plans.Tier{tier}
	req.Service = []plans.Service{service}

	sched := plans.Schedule{
		ScheduleName: scheduleName,
		CurrencyCd:   CURRENCY_CD,
		IsDefault:    ISDEFAULT,
	}
	req.PlanName = planName
	req.Schedule = []plans.Schedule{sched}
	req.ClientPlanId = clientPlanId
	req.SupplementalObjField = []plans.SupplementalObjectField{
		{FieldName: "SYNC TO SAP GTS", FieldValue: []string{"No"}},
	}
	resp, err := CallAriaAdmin[response.PlanResponse](ctx, aPC.ariaAdminClient, req, FailedToCreatePlanError)
	return resp, err
}
