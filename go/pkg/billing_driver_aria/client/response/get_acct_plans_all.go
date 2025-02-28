// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetAcctPlansAllMResponse struct {
	AriaResponse
	RecordCount   int64                `json:"record_count,omitempty"`
	AllAcctPlansM []data.AllAcctPlansM `json:"all_acct_plans_m,omitempty"`
}
