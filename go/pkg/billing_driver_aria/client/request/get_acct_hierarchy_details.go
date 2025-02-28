// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetAccountHierachyDetailsRequest struct {
	AriaRequest
	OutputFormat    string `json:"output_format"`
	ClientNo        int64  `json:"client_no"`
	ClientAcctId    string `json:"client_acct_id,omitempty"`
	AuthKey         string `json:"auth_key"`
	AltCallerId     string `json:"alt_caller_id"`
	AcctNo          int64  `json:"acct_no,omitempty"`
	HierarchyFilter int64  `json:"hierarchy_filter,omitempty"`
}
