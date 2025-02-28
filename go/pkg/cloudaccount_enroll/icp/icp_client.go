// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package icp

import (
	"context"
	"fmt"
)

type ICPClient interface {
	IsEnterprisePending(ctx context.Context, email string, oid string) (bool, string, error)
	GetPersonId(ctx context.Context, email string, oid string) (string, error)
}

// ICPClient with a cached token.
type ICPClientImpl struct{}

func CreateICPClient(cfg *ICPConfig) ICPClient {
	return &ICPClientImpl{}
}

// Function to return PersonID.
func (icp *ICPClientImpl) GetPersonId(ctx context.Context, email string, oid string) (string, error) {
	fmt.Println("Default getPersonId called")
	return "", nil
}

// Function to verify if the account waiting on enterprise approval.
func (icp *ICPClientImpl) IsEnterprisePending(ctx context.Context, email string, oid string) (bool, string, error) {
	fmt.Println("Default IsEnterprisePending called")
	return false, "", nil
}
