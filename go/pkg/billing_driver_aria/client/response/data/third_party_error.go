// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type ThirdPartyError struct {
	ErrorClass string `json:"error_class,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
	ErrorMsg   string `json:"error_msg,omitempty"`
}
