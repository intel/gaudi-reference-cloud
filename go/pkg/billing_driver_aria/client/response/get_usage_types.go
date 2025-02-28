// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetUsageTypes struct {
	AriaResponse
	UsageTypes []data.UsageType `json:"usage_types"`
}
