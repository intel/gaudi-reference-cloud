// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/credits"
)

type CancelUnappliedServiceCreditsMRequest struct {
	RestCall     string `json:"rest_call"`
	OutputFormat string `json:"output_format"`
	ClientNo     int64  `json:"client_no"`
	AuthKey      string `json:"auth_key"`
	AltCallerId  string `json:"alt_caller_id"`
	// Aria-assigned account identifier. This value is unique across all Aria-managed accounts.
	AcctNo    int64              `json:"acct_no,omitempty"`
	CreditIds []credits.CreditId `json:"credit_ids"`
}
