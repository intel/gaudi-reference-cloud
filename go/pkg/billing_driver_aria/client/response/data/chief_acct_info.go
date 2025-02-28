// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type ChiefAcctInfo struct {
	ChiefAcctNo       int64  `json:"chief_acct_no,omitempty"`
	ChiefAcctUserId   string `json:"chief_acct_user_id,omitempty"`
	ChiefClientAcctId string `json:"chief_client_acct_id,omitempty"`
}
