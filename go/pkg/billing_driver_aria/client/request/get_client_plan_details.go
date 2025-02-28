// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetPlanDetails struct {
	AriaRequest
	GetAllClientPlans
	IncludeRateScheduleSummary string `json:"include_rs_summary"`
}
