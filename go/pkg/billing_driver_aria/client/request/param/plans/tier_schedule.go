// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package plans

type TierSchedule struct {
	From int `json:"from,omitempty" url:"from,omitempty"`
	To   int `json:"to,omitempty" url:"to"`
	// No omitempty on Amount, because if 0 it still must be specified
	Amount float64 `json:"amount" url:"amount"`
	// TODO: Remove omitempty for Amount when billing handles IntelRates
}
