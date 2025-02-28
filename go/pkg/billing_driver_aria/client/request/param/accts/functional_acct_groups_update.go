// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package accts

type FunctionalAcctGroupsUpdate struct {
	FunctionalAcctGroupNo       int64  `json:"functional_acct_group_no,omitempty"`
	ClientFunctionalAcctGroupId string `json:"client_functional_acct_group_id,omitempty"`
	FunctionalAcctGrpDirective  int64  `json:"functional_acct_grp_directive,omitempty"`
}
