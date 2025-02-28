// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type AllClientPlanDtl struct {
	PlanNo                   int64                `json:"plan_no,omitempty"`
	PlanName                 string               `json:"plan_name,omitempty"`
	PlanDesc                 string               `json:"plan_desc,omitempty"`
	SuppPlanInd              int64                `json:"supp_plan_ind,omitempty"`
	BillingInd               int64                `json:"billing_ind,omitempty"`
	DisplayInd               int64                `json:"display_ind,omitempty"`
	NewAcctStatus            int64                `json:"new_acct_status,omitempty"`
	RolloverAcctStatus       int64                `json:"rollover_acct_status,omitempty"`
	RolloverAcctStatusDays   int64                `json:"rollover_acct_status_days,omitempty"`
	PrepaidInd               int64                `json:"prepaid_ind,omitempty"`
	CurrencyCd               string               `json:"currency_cd,omitempty"`
	ClientPlanId             string               `json:"client_plan_id,omitempty"`
	ProrationInvoiceTimingCd string               `json:"proration_invoice_timing_cd,omitempty"`
	RolloverPlanUomCd        int64                `json:"rollover_plan_uom_cd,omitempty"`
	InitFreePeriodUomCd      string               `json:"init_free_period_uom_cd,omitempty"`
	InitialPlanStatusCd      int64                `json:"initial_plan_status_cd,omitempty"`
	RolloverPlanStatusUomCd  int64                `json:"rollover_plan_status_uom_cd,omitempty"`
	RolloverPlanStatusCd     int64                `json:"rollover_plan_status_cd,omitempty"`
	NsoInclListScope         int64                `json:"nso_incl_list_scope,omitempty"`
	PlanServices             []PlanService        `json:"plan_services,omitempty"`
	PromotionalPlanSets      []PromotionalPlanSet `json:"promotional_plan_sets,omitempty"`
	PlanSuppFields           []PlanSuppField      `json:"plan_supp_fields,omitempty"`
}
