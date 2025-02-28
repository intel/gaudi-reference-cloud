// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type ApplyServiceCreditResponse struct {
	AriaResponse
	CreditId int64 `json:"credit_id"`
}
