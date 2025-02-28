// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type PromoPlanSetDetails struct {
	PlanNo       string `json:"plan_no"`
	PlanName     string `json:"plan_name"`
	PlanDesc     string `json:"plan_desc"`
	ClientPlanId string `json:"client_plan_id"`
}

type GetPromoPlanSetDetailsMResponse struct {
	AriaResponse
	PromoPlanSetNo   int64                 `json:"promo_plan_set_no"`
	PromoPlanSetName string                `json:"promo_plan_set_name"`
	PromoPlanSetDesc string                `json:"promo_plan_set_desc"`
	ClientPlanTypeId string                `json:"client_plan_type_id"`
	Plan             []PromoPlanSetDetails `json:"plan"`
}
