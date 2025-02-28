// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type RecordUsage struct {
	AriaResponse
	UsageRecNo int64 `json:"usage_rec_no,omitempty"`
}
