// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

import "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/usages"

type BulkRecordUsageMResponse struct {
	AriaResponse
	ErrorCode    int64                 `json:"error_code,omitempty"`
	ErrorMsg     string                `json:"error_msg,omitempty"`
	ErrorRecords []usages.ErrorRecords `json:"error_records,omitempty"`
}
