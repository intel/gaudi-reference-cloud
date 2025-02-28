// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type CancelUnappliedServiceCreditsMResponse struct {
	ErrorCode int64  `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}
