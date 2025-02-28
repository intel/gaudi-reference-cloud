// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetUnbilledUsageSummaryMRequest struct {
	AriaRequest
	OutputFormat               string `json:"output_format"`
	ClientNo                   int64  `json:"client_no,omitempty"`
	AuthKey                    string `json:"auth_key,omitempty"`
	ClientAcctId               string `json:"client_acct_id,omitempty"`
	ClientMasterPlanInstanceId string `json:"client_master_plan_instance_id,omitempty"`
}
