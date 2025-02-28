// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

import (
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_driver_aria/client/request/param/usages"
)

type BulkRecordUsageMRequest struct {
	AriaRequest
	RestCall     string            `json:"rest_call"`
	OutputFormat string            `json:"output_format"`
	ClientNo     int64             `json:"client_no"`
	AuthKey      string            `json:"auth_key"`
	AltCallerId  string            `json:"alt_caller_id"`
	UsageRecs    []usages.UsageRec `json:"usage_recs"`
}
