// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetAccountNotificationDetails struct {
	AriaRequest
	OutputFormat   string `json:"output_format"`
	ClientNo       int64  `json:"client_no"`
	AuthKey        string `json:"auth_key"`
	AltCallerId    string `json:"alt_caller_id"`
	AcctNo         int64  `json:"acct_no,omitempty"`
	AcctUserId     string `json:"acct_user_id,omitempty"`
	ClientAcctId   string `json:"client_acct_id,omitempty"`
	ReleaseVersion string `json:"releaseVersion"`
}
