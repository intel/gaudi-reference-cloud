// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetClientPlansAllMResponse struct {
	AriaResponse
	AllClientPlanDtls []data.AllClientPlanDtl `json:"all_client_plan_dtls,omitempty"`
}
