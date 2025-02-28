// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package request

type Request interface {
	GetRestCall() string
}

type AriaRequest struct {
	RestCall string `json:"rest_call"`
}

func (ar *AriaRequest) GetRestCall() string {
	return ar.RestCall
}
