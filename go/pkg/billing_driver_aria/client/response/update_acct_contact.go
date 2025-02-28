// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type UpdateAccountContact struct {
	AriaResponse
	ContactNo  string `json:"contact_no,omitempty"`
	ContactNo2 int64  `json:"contact_no_2,omitempty"`
}
