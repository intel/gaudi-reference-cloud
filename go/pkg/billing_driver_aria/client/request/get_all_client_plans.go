// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetAllClientPlans struct {
	AriaRequest
	OutputFormat string `json:"output_format"`
	ClientNo     int64  `json:"client_no"`
	AuthKey      string `json:"auth_key"`
	AltCallerId  string `json:"alt_caller_id"`
	PromoCode    string `json:"promo_code,omitempty"`
	ClientPlanId string `json:"client_plan_id"`
}
