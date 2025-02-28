// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/invoice"
)

type Invoice struct {
	AriaRequest
	OutputFormat                  string                                 `json:"output_format"`
	ClientNo                      int64                                  `json:"client_no"`
	AuthKey                       string                                 `json:"auth_key"`
	AltCallerId                   string                                 `json:"alt_caller_id"`
	AcctNo                        int64                                  `json:"acct_no,omitempty"`
	InvoiceNo                     int64                                  `json:"invoice_no"`
	ClientAcctId                  string                                 `json:"client_acct_id,omitempty"`
	MasterPlanInstanceNo          int64                                  `json:"master_plan_instance_no,omitempty"`
	ClientMasterPlanInstanceId    string                                 `json:"client_master_plan_instance_id,omitempty"`
	BillingGroupNo                int64                                  `json:"billing_group_no,omitempty"`
	ClientBillingGroupId          string                                 `json:"client_billing_group_id,omitempty"`
	ForcePending                  string                                 `json:"force_pending,omitempty"`
	ClientReceiptId               string                                 `json:"client_receipt_id,omitempty"`
	AltBillDay                    int64                                  `json:"alt_bill_day,omitempty"`
	InvoiceMode                   int64                                  `json:"invoice_mode,omitempty"`
	CombineInvoices               int64                                  `json:"combine_invoices,omitempty"`
	OptionalTransactionQualifiers []invoice.OptionalTransactionQualifier `json:"optional_transaction_qualifiers,omitempty"`
}
