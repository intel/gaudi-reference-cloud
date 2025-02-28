// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type CreateService struct {
	AriaResponse
	ServiceNo int `json:"service_no"`
}
