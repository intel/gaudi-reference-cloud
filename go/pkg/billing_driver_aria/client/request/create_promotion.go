// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type CreatePromotionMRequest struct {
	AriaRequest
	OutputFormat   string `json:"output_format"`
	ClientNo       int64  `json:"client_no"`
	AuthKey        string `json:"auth_key"`
	PromoCd        string `json:"promo_cd"`
	PromoDesc      string `json:"promo_desc"`
	PromoPlanSetNo int64  `json:"promo_plan_set_no"`
	StartDate      string `json:"start_date,omitempty"`
	ExpDate        string `json:"exp_date,omitempty"`
}
