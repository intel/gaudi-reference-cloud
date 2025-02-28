// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package client

// Todo: This is a very initial implementation

type AriaCredentials struct {
	clientNo int64
	authKey  string
}

func NewAriaCredentials(clientNo int64, authKey string) *AriaCredentials {
	return &AriaCredentials{
		clientNo: clientNo,
		authKey:  authKey,
	}
}
