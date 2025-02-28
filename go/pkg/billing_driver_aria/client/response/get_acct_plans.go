// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetAcctPlansMResponse struct {
	AriaResponse
	AcctPlansM []data.AcctPlansM `json:"acct_plans_m,omitempty"`
}
