// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetPlanDetailsMRequest struct {
	AriaRequest
	OutputFormat string `json:"output_format"`
	ClientNo     int64  `json:"client_no"`
	AuthKey      string `json:"auth_key"`
	ClientPlanId string `json:"client_plan_id,omitempty"`
	// Other fields omitted
}
