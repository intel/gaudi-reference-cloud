// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type MasterPlansAssigned struct {
	PlanInstanceNo       int64  `json:"plan_instance_no,omitempty"`
	ClientPlanInstanceId string `json:"client_plan_instance_id,omitempty"`
}
