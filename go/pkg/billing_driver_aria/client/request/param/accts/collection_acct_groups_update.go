// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package accts

type CollectionAcctGroupsUpdate struct {
	CollectionAcctGroupNo       int64  `json:"collection_acct_group_no,omitempty"`
	ClientCollectionAcctGroupId string `json:"client_collection_acct_group_id,omitempty"`
	CollectionAcctGrpDirective  int64  `json:"collection_acct_grp_directive,omitempty"`
}
