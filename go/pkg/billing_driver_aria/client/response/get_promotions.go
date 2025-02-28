// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type Promotion struct {
	PromoCd      string `json:"promo_cd"`
	PromoDesc    string `json:"promo_desc"`
	ExpiresAfter string `json:"expires_after"`
}

type GetPromotionsMResponse struct {
	AriaResponse
	Promotions []Promotion `json:"promotions"`
}
