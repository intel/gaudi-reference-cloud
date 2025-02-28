// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type AcctHierarchyDtl struct {
	AcctNo             int64                `json:"acct_no,omitempty"`
	ClientAcctId       string               `json:"client_acct_id,omitempty"`
	Userid             string               `json:"userid,omitempty"`
	SeniorAcctNo       int64                `json:"senior_acct_no,omitempty"`
	SeniorAcctUserId   string               `json:"senior_acct_user_id,omitempty"`
	SeniorClientAcctId string               `json:"senior_client_acct_id,omitempty"`
	TestAcctInd        int64                `json:"test_acct_ind,omitempty"`
	BillingGroupsInfo  []BillingGroupsInfo  `json:"billing_groups_info,omitempty"`
	PaymentMethodsInfo []PaymentMethodsInfo `json:"payment_methods_info,omitempty"`
	MasterPlansInfo    []MasterPlansInfo    `json:"master_plans_info,omitempty"`
	ChildAcctNo        []ChildAcctNo        `json:"child_acct_no,omitempty"`
}
