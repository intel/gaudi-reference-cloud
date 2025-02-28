// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package response

type GetAccountNoFromUserIdMResponse struct {
	AriaResponse
	AcctNo int64 `json:"acct_no"`
}
