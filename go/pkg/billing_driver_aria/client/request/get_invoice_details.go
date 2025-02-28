// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type GetInvoiceDetails struct {
	AriaRequest
	OutputFormat               string `json:"output_format"`
	ClientNo                   int64  `json:"client_no"`
	AuthKey                    string `json:"auth_key"`
	AltCallerId                string `json:"alt_caller_id"`
	AcctNo                     int64  `json:"acct_no,omitempty"`
	InvoiceNo                  int64  `json:"invoice_no"`
	ClientAcctId               string `json:"client_acct_id,omitempty"`
	MasterPlanInstanceNo       int64  `json:"master_plan_instance_no,omitempty"`
	ClientMasterPlanInstanceId string `json:"client_master_plan_instance_id,omitempty"`
	LocaleNo                   int64  `json:"locale_no,omitempty"`
	LocaleName                 string `json:"locale_name,omitempty"`
}
