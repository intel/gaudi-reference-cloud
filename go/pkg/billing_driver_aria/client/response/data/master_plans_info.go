// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type MasterPlansInfo struct {
	MasterPlanInstanceNo          int64                    `json:"master_plan_instance_no,omitempty"`
	ClientMasterPlanInstanceId    string                   `json:"client_master_plan_instance_id,omitempty"`
	ClientMasterPlanId            string                   `json:"client_master_plan_id,omitempty"`
	MasterPlanNo                  int64                    `json:"master_plan_no,omitempty"`
	MasterPlanInstanceDescription string                   `json:"master_plan_instance_description,omitempty"`
	DunningGroupNo                int64                    `json:"dunning_group_no,omitempty"`
	ClientDunningGroupId          string                   `json:"client_dunning_group_id,omitempty"`
	DunningGroupName              string                   `json:"dunning_group_name,omitempty"`
	DunningGroupDescription       string                   `json:"dunning_group_description,omitempty"`
	DunningProcessNo              int64                    `json:"dunning_process_no,omitempty"`
	ClientDunningProcessId        string                   `json:"client_dunning_process_id,omitempty"`
	BillingGroupNo                int64                    `json:"billing_group_no,omitempty"`
	ClientBillingGroupId          string                   `json:"client_billing_group_id,omitempty"`
	MasterPlanInstanceStatus      int64                    `json:"master_plan_instance_status,omitempty"`
	MpInstanceStatusLabel         string                   `json:"mp_instance_status_label,omitempty"`
	MasterPlanUnits               int64                    `json:"master_plan_units,omitempty"`
	RespLevelCd                   int64                    `json:"resp_level_cd,omitempty"`
	AltRateScheduleNo             int64                    `json:"alt_rate_schedule_no,omitempty"`
	ClientAltRateScheduleId       string                   `json:"client_alt_rate_schedule_id,omitempty"`
	PromoCd                       string                   `json:"promo_cd,omitempty"`
	BillDay                       int64                    `json:"bill_day,omitempty"`
	LastArrearsBillThruDate       string                   `json:"last_arrears_bill_thru_date,omitempty"`
	LastBillDate                  string                   `json:"last_bill_date,omitempty"`
	LastBillThruDate              string                   `json:"last_bill_thru_date,omitempty"`
	NextBillDate                  string                   `json:"next_bill_date,omitempty"`
	PlanDate                      string                   `json:"plan_date,omitempty"`
	StatusDate                    string                   `json:"status_date,omitempty"`
	RecurringBillingInterval      int64                    `json:"recurring_billing_interval,omitempty"`
	UsageBillingInterval          int64                    `json:"usage_billing_interval,omitempty"`
	RecurringBillingPeriodType    int64                    `json:"recurring_billing_period_type,omitempty"`
	UsageBillingPeriodType        int64                    `json:"usage_billing_period_type,omitempty"`
	InitialPlanStatus             int64                    `json:"initial_plan_status,omitempty"`
	RolloverPlanStatus            int64                    `json:"rollover_plan_status,omitempty"`
	RolloverPlanStatusDuration    int64                    `json:"rollover_plan_status_duration,omitempty"`
	RolloverPlanStatusUomCd       int64                    `json:"rollover_plan_status_uom_cd,omitempty"`
	InitFreePeriodDuration        int64                    `json:"init_free_period_duration,omitempty"`
	InitFreePeriodUomCd           int64                    `json:"init_free_period_uom_cd,omitempty"`
	DunningState                  int64                    `json:"dunning_state,omitempty"`
	DunningStep                   int64                    `json:"dunning_step,omitempty"`
	DunningDegradeDate            string                   `json:"dunning_degrade_date,omitempty"`
	PlanDeprovisionedDate         string                   `json:"plan_deprovisioned_date,omitempty"`
	MasterPlanProductFields       []MasterPlanProductField `json:"master_plan_product_fields,omitempty"`
	MpPlanInstFields              []MpPlanInstField        `json:"mp_plan_inst_fields,omitempty"`
	MasterPlansServices           []MasterPlansService     `json:"master_plans_services,omitempty"`
	BillLagDays                   int64                    `json:"bill_lag_days,omitempty"`
	LastArrRecurBillThruDate      string                   `json:"last_arr_recur_bill_thru_date,omitempty"`
}
