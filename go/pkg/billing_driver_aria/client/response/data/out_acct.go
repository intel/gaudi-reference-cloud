// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type OutAcct struct {
	AcctNo                  int64                    `json:"acct_no,omitempty"`
	Userid                  string                   `json:"userid,omitempty"`
	ClientAcctId            string                   `json:"client_acct_id,omitempty"`
	AcctLocaleNo            int64                    `json:"acct_locale_no,omitempty"`
	AcctLocaleName          string                   `json:"acct_locale_name,omitempty"`
	AcctContactNo           int64                    `json:"acct_contact_no,omitempty"`
	StatementContactDetails []StatementContactDetail `json:"statement_contact_details,omitempty"`
	BillingErrors           []BillingError           `json:"billing_errors,omitempty"`
	MasterPlansAssigned     []MasterPlansAssigned    `json:"master_plans_assigned,omitempty"`
	InvoiceInfo             []InvoiceInfo            `json:"invoice_info,omitempty"`
}
