// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type Response interface {
	GetErrorCode() int64
	GetErrorMsg() string
}

type AriaResponse struct {
	ErrorCode int64  `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

func (ar AriaResponse) GetErrorCode() int64 {
	return ar.ErrorCode
}

func (ar AriaResponse) GetErrorMsg() string {
	return ar.ErrorMsg
}
