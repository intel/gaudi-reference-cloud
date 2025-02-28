// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/config"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

type PromoClient struct {
	adminClient *AriaAdminClient
	creds       *AriaCredentials
}

func NewPromoClient(adminClient *AriaAdminClient, creds *AriaCredentials) *PromoClient {
	return &PromoClient{adminClient: adminClient, creds: creds}
}

var (
	PromoPlanSetNo int64
)

// When sync creates new plans, their ids are passed to AddPlansToPromo
// at the end of sync to include the plans in the promo's plan set. This
// sets up the plans to be included in subsequent calls to
// GetAllClientPlansForPromoCode
func (pc *PromoClient) AddPlansToPromo(ctx context.Context, planClientIds []string) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.AddPlansToPromo").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	planSetId := GetPlanSetId()
	details, err := pc.getPromoPlanSetDetails(ctx, planSetId)
	if err != nil {
		return err
	}
	planMap := map[string]bool{}
	builder := strings.Builder{}
	for _, plan := range details.Plan {
		if builder.Len() > 0 {
			if err := builder.WriteByte('|'); err != nil {
				logger.Error(err, "Error executing WriteByte")
			}
		}
		_, err := builder.WriteString(plan.ClientPlanId)
		if err != nil {
			logger.Error(err, "Error executing WriteString")
		}
		planMap[plan.ClientPlanId] = true
	}

	for _, id := range planClientIds {
		if planMap[id] {
			continue
		}
		if builder.Len() > 0 {
			err := builder.WriteByte('|')
			if err != nil {
				return err
			}
		}
		_, err := builder.WriteString(id)
		if err != nil {
			logger.Error(err, "Error executing WriteString", id)
			return err
		}
	}

	_, err = pc.updatePlanSet(ctx, details.PromoPlanSetName, details.PromoPlanSetDesc, planSetId, builder.String())
	return err
}

func (pc *PromoClient) EnsurePlanSet(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.EnsurePlanSet").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	sets, err := pc.getPromoPlanSets(ctx)

	if err != nil {
		return err
	}

	setClientId := GetPlanSetId()
	for _, set := range sets.PromoPlanSet {
		if set.ClientPlanTypeId == setClientId {
			// Got it!
			PromoPlanSetNo = set.PromoPlanSetNo
			return nil
		}
	}

	logger.Info("plan set not found, creating")

	name := GetPlanSetName()
	res, err := pc.createPromoPlanSet(ctx, name, name, setClientId)
	if err != nil {
		return err
	}
	num, err := strconv.ParseInt(res.PromoPlanSetNo, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid promo_plan_set_no: %v", err)
	}
	PromoPlanSetNo = num
	return nil
}

func (pc *PromoClient) EnsurePromo(ctx context.Context) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.EnsurePromo").Start()
	defer span.End()
	logger.Info("BEGIN")
	defer logger.Info("END")
	promos, err := pc.getPromotions(ctx)
	if err != nil {
		return err
	}

	code := GetPromoCode()
	for _, promo := range promos.Promotions {
		if promo.PromoCd == code {
			// Already exists
			return nil
		}
	}

	logger.Info("IDC promotion not found, creating")

	desc := strings.ToUpper(config.Cfg.ClientIdPrefix) + " Plans"
	_, err = pc.createPromotion(ctx, code, desc, PromoPlanSetNo)
	if err != nil {
		return err
	}
	//bug fix in case no plan plan set shows all plan
	err = pc.addDefaultPlanToPromo(ctx, GetDefaultPlanClientId())
	return err
}

func (pc *PromoClient) getPromoPlanSets(ctx context.Context) (*response.GetPromoPlanSetsMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.getPromoPlanSets").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	req := request.GetPromoPlanSetsMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_promo_plan_sets_m"},
		OutputFormat: "json",
		ClientNo:     pc.creds.clientNo,
		AuthKey:      pc.creds.authKey,
	}

	return CallAriaAdmin[response.GetPromoPlanSetsMResponse](ctx, pc.adminClient, &req, FailedToGetPromoPlanSetsError)
}

func (pc *PromoClient) getPromoPlanSetDetails(ctx context.Context, planSetId string) (*response.GetPromoPlanSetDetailsMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.getPromoPlanSetDetails").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	req := request.GetPromoPlanSetDetailsMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_promo_plan_set_details_m"},
		OutputFormat:     "json",
		ClientNo:         pc.creds.clientNo,
		AuthKey:          pc.creds.authKey,
		ClientPlanTypeId: planSetId,
	}

	return CallAriaAdmin[response.GetPromoPlanSetDetailsMResponse](ctx, pc.adminClient, &req, FailedToGetPromoPlanSetDetails)
}

func (pc *PromoClient) createPromoPlanSet(ctx context.Context,
	name string, desc string, clientId string) (*response.CreatePromoPlanSetMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.createPromoPlanSet").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	req := request.CreatePromoPlanSetMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "create_promo_plan_set_m"},
		OutputFormat:     "json",
		ClientNo:         pc.creds.clientNo,
		AuthKey:          pc.creds.authKey,
		PromoPlanSetName: name,
		PromoPlanSetDesc: desc,
		ClientPlanTypeId: clientId,
		// Initialize the plan set with the default plan id. We do this
		// because a plan set with no explicit plans associated with it
		// implicitly includes every plan in Aria. Keeping the default plan
		// in the plan set enables GetAllClientPlansForPromoCode to return
		// an empty set of plans if there are no IDC plans in Aria, which
		// is what we want.
		ClientPlanId: GetDefaultPlanClientId(),
	}
	return CallAriaAdmin[response.CreatePromoPlanSetMResponse](ctx, pc.adminClient, &req, FailedToCreatePromoPlanSetError)
}

func (pc *PromoClient) updatePlanSet(ctx context.Context, name string, desc string, clientPlanSetId string, clientPlanId string) (*response.UpdatePromoPlanSetMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.updatePlanSet").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	req := request.UpdatePromoPlanSetMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "update_promo_plan_set_m"},
		OutputFormat:     "json",
		ClientNo:         pc.creds.clientNo,
		AuthKey:          pc.creds.authKey,
		AltCallerId:      "IDC Tester",
		PromoPlanSetName: name,
		PromoPlanSetDesc: desc,
		ClientPlanId:     clientPlanId,
		ClientPlanTypeId: clientPlanSetId,
	}
	return CallAriaAdmin[response.UpdatePromoPlanSetMResponse](ctx, pc.adminClient, &req, FailedToUpdatePlanSetError)
}

func (pc *PromoClient) getPromotions(ctx context.Context) (*response.GetPromotionsMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.getPromotions").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	req := request.GetPromotionsMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "get_promotions_m"},
		OutputFormat: "json",
		ClientNo:     pc.creds.clientNo,
		AuthKey:      pc.creds.authKey,
	}
	return CallAriaAdmin[response.GetPromotionsMResponse](ctx, pc.adminClient, &req, FailedToGetPromotionsError)
}

func (pc *PromoClient) createPromotion(ctx context.Context, code string, desc string, planSetNo int64) (*response.CreatePromotionMResponse, error) {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.createPromotion").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	req := request.CreatePromotionMRequest{
		AriaRequest: request.AriaRequest{
			RestCall: "create_promotion_m"},
		OutputFormat:   "json",
		ClientNo:       pc.creds.clientNo,
		AuthKey:        pc.creds.authKey,
		PromoCd:        code,
		PromoDesc:      desc,
		PromoPlanSetNo: planSetNo,
		StartDate:      time.Now().UTC().Format("2006-01-02"),
	}
	return CallAriaAdmin[response.CreatePromotionMResponse](ctx, pc.adminClient, &req, FailedToCreatePromotionError)
}

func (pc *PromoClient) addDefaultPlanToPromo(ctx context.Context, planClientIds string) error {
	ctx, logger, span := obs.LogAndSpanFromContext(ctx).WithName("PromoClient.addDefaultPlanToPromo").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")
	planSetId := GetPlanSetId()
	details, err := pc.getPromoPlanSetDetails(ctx, planSetId)
	if err != nil {
		return err
	}
	_, err = pc.updatePlanSet(ctx, details.PromoPlanSetName, details.PromoPlanSetDesc, planSetId, planClientIds)
	return err
}
