// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type AssignCollectionsAccountGroupRequest struct {
	AriaRequest
	OutputFormat      string `json:"output_format"`
	ClientNo          int64  `json:"client_no"`
	AuthKey           string `json:"auth_key"`
	AltCallerId       string `json:"alt_caller_id"`
	ClientAcctGroupId string `json:"client_acct_group_id,omitempty"`
	ClientAccountId   string `json:"client_acct_id"`
}
