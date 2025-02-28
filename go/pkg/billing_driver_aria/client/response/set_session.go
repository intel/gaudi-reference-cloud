// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type SetSessionMResponse struct {
	AriaResponse
	SessionId string `json:"session_id"`
}
