// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type AllAcctPlansM struct {
	PlanNo               int64  `json:"plan_no,omitempty"`
	PlanName             string `json:"plan_name,omitempty"`
	PlanDesc             string `json:"plan_desc,omitempty"`
	PlanInstanceNo       int64  `json:"plan_instance_no,omitempty"`
	ClientPlanInstanceId string `json:"client_plan_instance_id,omitempty"`
	BillLagDays          int64  `json:"bill_lag_days,omitempty"`
}
