// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package accts

type ProductCatalogPlanFilter struct {
	PlanNo       int64  `json:"plan_no,omitempty"`
	ClientPlanId string `json:"client_plan_id,omitempty"`
}
