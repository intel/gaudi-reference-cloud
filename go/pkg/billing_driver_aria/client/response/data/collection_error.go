// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package data

type CollectionError struct {
	CollectionErrorCode int64  `json:"collection_error_code,omitempty"`
	CollectionErrorMsg  string `json:"collection_error_msg,omitempty"`
}
