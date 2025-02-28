// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type CreateAcctCompleteMResponse struct {
	AriaResponse
	OutAcct       []data.OutAcct       `json:"out_acct,omitempty"`
	ChiefAcctInfo []data.ChiefAcctInfo `json:"chief_acct_info,omitempty"`
}
