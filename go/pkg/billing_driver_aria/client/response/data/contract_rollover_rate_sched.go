// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type ContractRolloverRateSched struct {
	CurrentRateSchedNo        int    `json:"current_rate_sched_no,omitempty"`
	CurrentClientRateSchedId  string `json:"current_client_rate_sched_id,omitempty"`
	RolloverRateSchedNo       int    `json:"rollover_rate_sched_no,omitempty"`
	RolloverClientRateSchedId string `json:"rollover_client_rate_sched_id,omitempty"`
}
