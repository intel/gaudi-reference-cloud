// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type InvoiceHist struct {
	InvoiceNo            int64   `json:"invoice_no,omitempty"`
	OutacctNo            int64   `json:"outacct_no,omitempty"`
	OutclientAcctId      string  `json:"outclient_acct_id,omitempty"`
	BillingGroupNo       int64   `json:"billing_group_no,omitempty"`
	ClientBillingGroupId string  `json:"client_billing_group_id,omitempty"`
	CurrencyCd           string  `json:"currency_cd,omitempty"`
	BillDate             string  `json:"bill_date,omitempty"`
	PaidDate             string  `json:"paid_date,omitempty"`
	Amount               float32 `json:"amount,omitempty"`
	Credit               float32 `json:"credit,omitempty"`
	RecurringBillFrom    string  `json:"recurring_bill_from,omitempty"`
	RecurringBillThru    string  `json:"recurring_bill_thru,omitempty"`
	UsageBillFrom        string  `json:"usage_bill_from,omitempty"`
	UsageBillThru        string  `json:"usage_bill_thru,omitempty"`
	IsVoidedInd          int64   `json:"is_voided_ind,omitempty"`
	InvoiceTypeCd        string  `json:"invoice_type_cd,omitempty"`
	OverallBillFromDate  string  `json:"overall_bill_from_date,omitempty"`
	OverallBillThruDate  string  `json:"overall_bill_thru_date,omitempty"`
}
