// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type UpdatePromoPlanSetMRequest struct {
	AriaRequest
	OutputFormat     string `json:"output_format"`
	ClientNo         int64  `json:"client_no"`
	AuthKey          string `json:"auth_key"`
	AltCallerId      string `json:"alt_caller_id,omitempty"`
	ClientPlanTypeId string `json:"client_plan_type_id,omitempty"`
	PromoPlanSetName string `json:"promo_plan_set_name,omitempty"`
	PromoPlanSetDesc string `json:"promo_plan_set_desc,omitempty"`
	ClientPlanId     string `json:"client_plan_id,omitempty"`
}
