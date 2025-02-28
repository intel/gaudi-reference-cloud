// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type PromoPlanSet struct {
	PromoPlanSetNo   int64  `json:"promo_plan_set_no"`
	PromoPlanSetName string `json:"promo_plan_set_name,omitempty"`
	PromoPlanSetDesc string `json:"promo_plan_set_desc,omitempty"`
	ClientPlanTypeId string `json:"client_plan_type_id,omitempty"`
}

type GetPromoPlanSetsMResponse struct {
	AriaResponse
	PromoPlanSet []PromoPlanSet `json:"promo_plan_set"`
}
