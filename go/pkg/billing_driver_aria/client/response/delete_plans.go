// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type DeletePlansResponse struct {
	AriaResponse
	PlanNos []int `json:"plan_nos"`
}
