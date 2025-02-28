// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetDunningGroupResponse struct {
	AriaResponse
	DunningGroupDetails []data.DunningGroupsDetails `json:"dunning_group_details,omitempty"`
	DunningGroups       []data.DunningGroupsDetails `json:"dunning_groups,omitempty"`
}
