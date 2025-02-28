// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PromotionalPlanSet struct {
	PromoSetNo       int64              `json:"promo_set_no,omitempty"`
	PromoSetName     string             `json:"promo_set_name,omitempty"`
	PromoSetDesc     string             `json:"promo_set_desc,omitempty"`
	PromotionsForSet []PromotionsForSet `json:"promotions_for_set,omitempty"`
	ClientPromoSetId string             `json:"client_promo_set_id,omitempty"`
}
