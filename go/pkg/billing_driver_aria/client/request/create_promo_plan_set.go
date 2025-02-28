// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type CreatePromoPlanSetMRequest struct {
	AriaRequest
	OutputFormat     string `json:"output_format"`
	ClientNo         int64  `json:"client_no"`
	AuthKey          string `json:"auth_key"`
	PromoPlanSetName string `json:"promo_plan_set_name"`
	PromoPlanSetDesc string `json:"promo_plan_set_desc"`
	ClientPlanTypeId string `json:"client_plan_type_id,omitempty"`
	PlanNo           int64  `json:"plan_no,omitempty"`
	ClientPlanId     string `json:"client_plan_id,omitempty"`
}
