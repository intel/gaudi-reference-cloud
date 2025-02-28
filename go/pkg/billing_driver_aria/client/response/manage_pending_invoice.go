// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type ManagePendingInvoice struct {
	AriaResponse
	NewInvoiceNo         int64                  `json:"new_invoice_no,omitempty"`
	AcctNo               int64                  `json:"acct_no,omitempty"`
	ClientAcctId         string                 `json:"client_acct_id,omitempty"`
	BillingGroupNo       int64                  `json:"billing_group_no,omitempty"`
	ClientBillingGroupId string                 `json:"client_billing_group_id,omitempty"`
	CollectionErrorCode  int64                  `json:"collection_error_code,omitempty"`
	CollectionErrorMsg   string                 `json:"collection_error_msg,omitempty"`
	StatementErrorCode   int64                  `json:"statement_error_code,omitempty"`
	StatementErrorMsg    string                 `json:"statement_error_msg,omitempty"`
	ProcCvvResponse      string                 `json:"proc_cvv_response,omitempty"`
	ProcAvsResponse      string                 `json:"proc_avs_response,omitempty"`
	ProcCavvResponse     string                 `json:"proc_cavv_response,omitempty"`
	ProcStatusCode       string                 `json:"proc_status_code,omitempty"`
	ProcStatusText       string                 `json:"proc_status_text,omitempty"`
	ProcPaymentId        string                 `json:"proc_payment_id,omitempty"`
	ProcAuthCode         string                 `json:"proc_auth_code,omitempty"`
	ProcMerchComments    string                 `json:"proc_merch_comments,omitempty"`
	ProcInitialAuthTxnId string                 `json:"proc_initial_auth_txn_id,omitempty"`
	CollectionErrors     []data.CollectionError `json:"collection_errors,omitempty"`
}
