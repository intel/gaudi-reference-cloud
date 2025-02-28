// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type CancelAcctPlanMRequest struct {
	RestCall               string `json:"rest_call"`
	OutputFormat           string `json:"output_format"`
	ClientNo               int64  `json:"client_no"`
	AuthKey                string `json:"auth_key"`
	AltCallerId            string `json:"alt_caller_id"`
	DoWrite                string `json:"do_write,omitempty"`
	AcctNo                 int64  `json:"acct_no,omitempty"`
	PlanInstanceNo         int64  `json:"plan_instance_no,omitempty"`
	AssignmentDirective    int64  `json:"assignment_directive,omitempty"`
	ProrationInvoiceTiming int64  `json:"proration_invoice_timing,omitempty"`
}
