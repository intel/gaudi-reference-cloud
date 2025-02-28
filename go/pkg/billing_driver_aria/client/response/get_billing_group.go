// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetBillingGroupResponse struct {
	AriaResponse
	BillingGroupDetails []data.BillingGroupsInfo `json:"billing_group_details,omitempty"`
	BillingGroups       []data.BillingGroupsInfo `json:"billing_groups,omitempty"`
}
