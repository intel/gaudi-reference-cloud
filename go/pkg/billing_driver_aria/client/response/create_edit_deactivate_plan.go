// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type PlanResponse struct {
	AriaResponse
	PlanNo int `json:"plan_no"`
}
