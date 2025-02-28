// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/accts"
)

type CreateAcctCompleteMRequest struct {
	AriaRequest
	OutputFormat    string       `json:"output_format"`
	ClientNo        int64        `json:"client_no"`
	AuthKey         string       `json:"auth_key"`
	AltCallerId     string       `json:"alt_caller_id"`
	ClientReceiptId string       `json:"client_receipt_id,omitempty"`
	Acct            []accts.Acct `json:"acct,omitempty"`
	ApplicationId   string       `json:"application_id,omitempty"`
	ApplicationDate string       `json:"application_date,omitempty"`
}
