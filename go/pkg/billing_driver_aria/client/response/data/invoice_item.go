// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type InvoiceItem struct {
	InvoiceLineNo        int64   `json:"invoice_line_no,omitempty"`
	PlanNo               int64   `json:"plan_no,omitempty"`
	ClientPlanId         string  `json:"client_plan_id,omitempty"`
	PlanInstanceNo       int64   `json:"plan_instance_no,omitempty"`
	ClientPlanInstanceId string  `json:"client_plan_instance_id,omitempty"`
	PlanName             string  `json:"plan_name,omitempty"`
	ServiceNo            int64   `json:"service_no,omitempty"`
	ClientServiceId      string  `json:"client_service_id,omitempty"`
	ServiceName          string  `json:"service_name,omitempty"`
	ServiceCoaId         int64   `json:"service_coa_id,omitempty"`
	Units                float32 `json:"units,omitempty"`
	RatePerUnit          float32 `json:"rate_per_unit,omitempty"`
	LineAmount           float32 `json:"line_amount,omitempty"`
	LineDescription      string  `json:"line_description,omitempty"`
	StartDateRange       string  `json:"start_date_range,omitempty"`
	EndDateRange         string  `json:"end_date_range,omitempty"`
	LineType             int64   `json:"line_type,omitempty"`
	BasePlanUnits        float32 `json:"base_plan_units,omitempty"`
	ProrationFactor      float32 `json:"proration_factor,omitempty"`
	RateScheduleNo       int64   `json:"rate_schedule_no,omitempty"`
	RateScheduleTierNo   int64   `json:"rate_schedule_tier_no,omitempty"`
	CustomRateInd        int64   `json:"custom_rate_ind,omitempty"`
}
