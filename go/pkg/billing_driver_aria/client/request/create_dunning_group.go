// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type CreateDunningGroupRequest struct {
	AriaRequest
	OutputFormat            string `json:"output_format"`
	ClientNo                int64  `json:"client_no"`
	AuthKey                 string `json:"auth_key"`
	AcctNo                  int64  `json:"acct_no,omitempty"`
	DunningGroupName        int64  `json:"dunning_group_name,omitempty"`
	DunningGroupDescription int64  `json:"dunning_group_description,omitempty"`
	ClientDunningGroupId    string `json:"client_dunning_group_id,omitempty"`
	ClientDunningProcessId  string `json:"client_dunning_process_id,omitempty"`
	ClientAcctId            string `json:"client_acct_id,omitempty"`
	AltCallerId             string `json:"alt_caller_id,omitempty"`
}
