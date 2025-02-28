// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PlanServiceRate struct {
	RateSeqNo            int64   `json:"rate_seq_no,omitempty"`
	FromUnit             float32 `json:"from_unit,omitempty"`
	RatePerUnit          float64 `json:"rate_per_unit,omitempty"`
	MonthlyFee           float32 `json:"monthly_fee,omitempty"`
	ClientRateScheduleId string  `json:"client_rate_schedule_id,omitempty"`
}
