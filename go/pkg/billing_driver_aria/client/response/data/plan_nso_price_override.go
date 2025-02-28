// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PlanNsoPriceOverride struct {
	RateScheduleNo                int    `json:"rate_schedule_no,omitempty"`
	ClientRateScheduleId          string `json:"client_rate_schedule_id,omitempty"`
	CurrencyCd                    string `json:"currency_cd,omitempty"`
	BundleNsoRateScheduleNo       int    `json:"bundle_nso_rate_schedule_no,omitempty"`
	BundleClientNsoRateScheduleId string `json:"bundle_client_nso_rate_schedule_id,omitempty"`
}
