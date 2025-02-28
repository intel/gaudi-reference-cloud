// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type PendingInvoice struct {
	InvoiceNo                  int64  `json:"invoice_no,omitempty"`
	AcctNo                     int64  `json:"acct_no,omitempty"`
	ClientMasterPlanInstanceId string `json:"client_master_plan_instance_id,omitempty"`
	BillingGroupNo             int64  `json:"billing_group_no,omitempty"`
	ClientBillingGroupId       string `json:"client_billing_group_id,omitempty"`
}
