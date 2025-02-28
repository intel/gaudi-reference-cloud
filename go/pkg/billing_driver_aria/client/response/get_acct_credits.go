// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetAccountCredits struct {
	AriaResponse
	AllCredits []data.AllCredit `json:"all_credits,omitempty"`
}
