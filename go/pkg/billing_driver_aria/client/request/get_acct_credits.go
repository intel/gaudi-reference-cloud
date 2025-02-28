// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetAccountCredits struct {
	AriaRequest
	OutputFormat               string `json:"output_format"`
	ClientNo                   int64  `json:"client_no"`
	ClientAcctId               string `json:"client_acct_id,omitempty"`
	AuthKey                    string `json:"auth_key"`
	AltCallerId                string `json:"alt_caller_id"`
	AcctNo                     int64  `json:"acct_no,omitempty"`
	MasterPlanInstanceNo       int64  `json:"master_plan_instance_no,omitempty"`
	ClientMasterPlanInstanceId string `json:"client_master_plan_instance_id,omitempty"`
	LimitRecords               int64  `json:"limit_records,omitempty"`
}
