// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package accts

type FunctionalAcctGroup struct {
	FunctionalAcctGroupNo       int64  `json:"functional_acct_group_no,omitempty"`
	ClientFunctionalAcctGroupId string `json:"client_functional_acct_group_id,omitempty"`
}
