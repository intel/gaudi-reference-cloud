// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetPlanDetailResponse struct {
	AriaResponse
	PlanNo                     int                `json:"plan_no,omitempty"`
	PlanName                   string             `json:"plan_name,omitempty"`
	PlanDesc                   string             `json:"plan_desc,omitempty"`
	PlanLevel                  int                `json:"plan_level,omitempty"`
	PlanType                   string             `json:"plan_type,omitempty"`
	PlanGroups                 []data.PlanGroup   `json:"plan_groups,omitempty"`
	PlanGroupsIds              []data.PlanGroupId `json:"plan_group_ids,omitempty"`
	CurrencyCd                 string             `json:"currency_cd,omitempty"`
	ClientPlanId               string             `json:"client_plan_id,omitempty"`
	ActiveInd                  int                `json:"active_ind,omitempty"`
	RolloverPlanDuration       int                `json:"rollover_plan_duration,omitempty"`
	RolloverPlanUomCd          int                `json:"rollover_plan_uom_cd,omitempty"`
	InitFreePeriodDuration     int                `json:"init_free_period_duration,omitempty"`
	InitFreePeriodUomCd        int                `json:"init_free_period_uom_cd,omitempty"`
	InitialPlanStatusCd        int                `json:"initial_plan_status_cd,omitempty"`
	RolloverPlanStatusDuration int                `json:"rollover_plan_status_duration,omitempty"`
	RolloverPlanStatusUomCd    int                `json:"rollover_plan_status_uom_cd,omitempty"`
	RolloverPlanStatusCd       int                `json:"rollover_plan_status_cd,omitempty"`
	AllowChildAccounts         string             `json:"allow_child_accounts,omitempty"`
	DunningPlanNo              int                `json:"dunning_plan_no,omitempty"`
	DunningClientPlanId        string             `json:"dunning_client_plan_id,omitempty"`
	AcctStatusCd               string             `json:"acct_status_cd,omitempty"`
	RolloverAcctStatusDays     int                `json:"rollover_acct_status_days,omitempty"`
	RolloverAcctStatusCd       string             `json:"rollover_acct_status_cd,omitempty"`
	//TemplateNo                   int                           		`json:"template_no,omitempty"`
	TemplateId                   string                           `json:"template_id,omitempty"`
	PlanCancelMinMonths          string                           `json:"plan_cancel_min_months,omitempty"`
	HowToApplyMinFee             string                           `json:"how_to_apply_min_fee,omitempty"`
	IsDeletable                  string                           `json:"is_deletable,omitempty"`
	Services                     []data.Service                   `json:"services,omitempty"`
	ParentPlans                  []data.ParentPlan                `json:"parent_plans,omitempty"`
	ParentPlanIds                []data.ParentPlanId              `json:"parent_plan_ids,omitempty"`
	ExclusionPlans               []data.ExclusionPlan             `json:"exclusion_plans,omitempty"`
	Resources                    []data.Resource                  `json:"resources,omitempty"`
	PromotionalPlanSets          []data.PromotionalPlanSet        `json:"promotional_plan_sets,omitempty"`
	PlanSuppFields               []data.PlanSuppField             `json:"supplemental_obj_fields,omitempty"`
	Surcharges                   []data.Surcharge                 `json:"surcharges,omitempty"`
	ProrationInvoiceTimingCd     string                           `json:"proration_invoice_timing_cd,omitempty"`
	RateSched                    []data.RateSchedule              `json:"rate_sched,omitempty"`
	ContractRolloverPlanNo       int                              `json:"contract_rollover_plan_no,omitempty"`
	ContractRolloverClientPlanId string                           `json:"contract_rollover_client_plan_id,omitempty"`
	ContractRolloverRateSched    []data.ContractRolloverRateSched `json:"contract_rollover_rate_sched,omitempty"`
	PlanNsoItems                 []data.PlanNsoItem               `json:"plan_nso_items,omitempty"`
	PlanNsoGroup                 []data.PlanNsoGroup              `json:"plan_nso_group,omitempty"`
	NsoInclListScope             int                              `json:"nso_incl_list_scope,omitempty"`
	PlanNsoInclList              []data.PlanNsoInclList           `json:"plan_nso_incl_list,omitempty"`
	PlanTranslationDetails       []data.PlanTranslationInfo       `json:"plan_translation_info,omitempty"`
	ItemNo                       int                              `json:"item_no,omitempty"`
	PlanNsoGroupPriceOverride    []data.PlanNsoPriceOverride      `json:"plan_nso_group_price_override,omitempty"`
	NsoGroupMinQty               string                           `json:"nso_group_min_qty,omitempty"`
	NsoGroupMaxQty               string                           `json:"nso_group_max_qty,omitempty"`
	NsoGroupItemScope            string                           `json:"nso_group_item_scope,omitempty"`
}
