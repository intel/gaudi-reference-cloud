// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PlanNsoItem struct {
	ItemNo               int64                  `json:"item_no,omitempty"`
	ActiveInd            int64                  `json:"active_ind,omitempty"`
	MinQty               string                 `json:"min_qty,omitempty"`
	MaxQty               string                 `json:"max_qty,omitempty"`
	ItemScope            int64                  `json:"item_scope,omitempty"`
	PlanNsoPriceOverride []PlanNsoPriceOverride `json:"plan_nso_price_override,omitempty"`
}
