// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/response/data"
)

type GetAcctHierarchyDetailsMResponse struct {
	AriaResponse
	AcctHierarchyDtls []data.AcctHierarchyDtl `json:"acct_hierarchy_dtls,omitempty"`
}
