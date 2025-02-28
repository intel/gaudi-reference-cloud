// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetInvoiceHistory struct {
	AriaRequest
	OutputFormat               string `json:"output_format"`
	ClientNo                   int64  `json:"client_no"`
	AuthKey                    string `json:"auth_key"`
	AltCallerId                string `json:"alt_caller_id"`
	AcctNo                     int64  `json:"acct_no,omitempty"`
	ClientAcctId               string `json:"client_acct_id,omitempty"`
	MasterPlanInstanceId       int64  `json:"master_plan_instance_id,omitempty"`
	UserId                     string `json:"user_id,omitempty"`
	StartBillDate              string `json:"start_bill_date,omitempty"`
	EndBillDate                string `json:"end_bill_date,omitempty"`
	IncludeVoided              string `json:"include_voided,omitempty"`
	ClientMasterPlanInstanceId string `json:"client_master_plan_instance_id,omitempty"`
	RbOption                   int64  `json:"rb_option,omitempty"`
}
