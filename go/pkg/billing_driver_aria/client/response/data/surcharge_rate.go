// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type SurchargeRate struct {
	RateSeqNo              int64   `json:"rate_seq_no,omitempty"`
	FromUnit               float32 `json:"from_unit,omitempty"`
	ToUnit                 float32 `json:"to_unit,omitempty"`
	RatePerUnit            float32 `json:"rate_per_unit,omitempty"`
	IncludeZero            int64   `json:"include_zero,omitempty"`
	RateSchedIsAssignedInd int64   `json:"rate_sched_is_assigned_ind,omitempty"`
	RateTierDescription    string  `json:"rate_tier_description,omitempty"`
}
