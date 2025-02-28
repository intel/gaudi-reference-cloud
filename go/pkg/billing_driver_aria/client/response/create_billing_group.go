// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type CreateBillingGroupResponse struct {
	AriaResponse
	BillingGroupNo2    int64                     `json:"billing_group_no_2,omitempty"`
	BillingContactInfo []data.BillingContactInfo `json:"billing_contact_info,omitempty"`
}
