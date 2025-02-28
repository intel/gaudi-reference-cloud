// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

import "fmt"

type AriaClientError struct {
	AriaRequestType  string
	AriaErrorCode    int64
	AriaErrorMessage string
	Err              error
}

func (ariaClientError *AriaClientError) Error() string {
	return fmt.Sprintf("aria_request: %s, aria_error code: %d, aria_error msg: %s, err: %v",
		ariaClientError.AriaRequestType, ariaClientError.AriaErrorCode,
		ariaClientError.AriaErrorMessage, ariaClientError.Err)
}
