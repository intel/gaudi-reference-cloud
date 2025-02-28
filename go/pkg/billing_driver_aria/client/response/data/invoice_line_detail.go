// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type InvoiceLineDetail struct {
	LineNo                      int64   `json:"line_no,omitempty"`
	ServiceNo                   int64   `json:"service_no,omitempty"`
	ServiceName                 string  `json:"service_name,omitempty"`
	Units                       float32 `json:"units,omitempty"`
	RatePerUnit                 float32 `json:"rate_per_unit,omitempty"`
	Amount                      float32 `json:"amount,omitempty"`
	DateRangeStart              string  `json:"date_range_start,omitempty"`
	DateRangeEnd                string  `json:"date_range_end,omitempty"`
	UsageTypeNo                 int64   `json:"usage_type_no,omitempty"`
	PlanNo                      int64   `json:"plan_no,omitempty"`
	PlanName                    string  `json:"plan_name,omitempty"`
	CreditReasonCd              int64   `json:"credit_reason_cd,omitempty"`
	CreditReasonCodeDescription string  `json:"credit_reason_code_description,omitempty"`
	UsageTypeCd                 string  `json:"usage_type_cd,omitempty"`
	ClientPlanId                string  `json:"client_plan_id,omitempty"`
}
