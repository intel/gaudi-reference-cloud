// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetPromoPlanSetDetailsMRequest struct {
	AriaRequest
	OutputFormat     string `json:"output_format"`
	ClientNo         int64  `json:"client_no"`
	AuthKey          string `json:"auth_key"`
	ClientPlanTypeId string `json:"client_plan_type_id"`
}
