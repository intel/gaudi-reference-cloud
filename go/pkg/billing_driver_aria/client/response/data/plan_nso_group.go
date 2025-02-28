// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PlanNsoGroup struct {
	ListRateScheduleNo            int64  `json:"list_rate_schedule_no,omitempty"`
	ClientListRateSchedule_id     string `json:"client_list_rate_schedule_id,omitempty"`
	BundleNsoRateScheduleNo       string `json:"bundle_nso_rate_schedule_no,omitempty"`
	BundleClientNsoRateScheduleId string `json:"bundle_client_nso_rate_schedule_id,omitempty"`
}
