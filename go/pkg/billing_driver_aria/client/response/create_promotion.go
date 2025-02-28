// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type CreatePromotionMResponse struct {
	AriaResponse
	PromoCd string `json:"promo_cd"`
}
