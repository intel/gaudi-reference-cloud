// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PlanSuppField struct {
	PlanSuppFieldNo    int64  `json:"plan_supp_field_no,omitempty"`
	PlanSuppFieldName  string `json:"plan_supp_field_name,omitempty"`
	PlanSuppFieldDesc  string `json:"plan_supp_field_desc,omitempty"`
	PlanSuppFieldValue string `json:"plan_supp_field_value,omitempty"`
}
