// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type ParentPlan struct {
	ParentPlan         int    `json:"parent_plan,omitempty"`
	ParentPlanName     string `json:"parent_plan_name,omitempty"`
	ClientParentPlanId string `json:"client_parent_plan_id,omitempty"`
}
