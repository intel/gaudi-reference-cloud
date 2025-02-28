// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package accts

type MasterPlansDetail struct {
	ClientPlanId       string  `json:"client_plan_id,omitempty"`
	PlanInstanceIdx    int64   `json:"plan_instance_idx,omitempty"`
	PlanInstanceUnits  float32 `json:"plan_instance_units,omitempty"`
	PlanInstanceStatus int64   `json:"plan_instance_status,omitempty"`
	BillingGroupIdx    int64   `json:"billing_group_idx,omitempty"`
	DunningGroupIdx    int64   `json:"dunning_group_idx,omitempty"`
}
