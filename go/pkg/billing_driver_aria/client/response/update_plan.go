// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type UpdatePlan struct {
	ErrorCode int64  `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	PlanNo    int    `json:"plan_no"`
}
