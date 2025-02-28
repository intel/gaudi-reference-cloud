// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type UpdateAccountBillingGroupMResponse struct {
	AriaResponse
	ProcStatusCode       string                    `json:"proc_status_code,omitempty"`
	ProcStatusText       string                    `json:"proc_status_text,omitempty"`
	ProcpaymentId        string                    `json:"proc_payment_id,omitempty"`
	ProcAuthCode         string                    `json:"proc_auth_code,omitempty"`
	ProcMerchantComments string                    `json:"proc_merch_comments,omitempty"`
	BillingGroupNo       string                    `json:"billing_group_no,omitempty"`
	BillingGroupNo2      int64                     `json:"billing_group_no_2,omitempty"`
	StmtContactNo        int64                     `json:"stmt_contact_no,omitempty"`
	StmtContactNo2       string                    `json:"stmt_contact_no_2,omitempty"`
	BillingContactInfo   []data.BillingContactInfo `json:"billing_contact_info,omitempty"`
}
