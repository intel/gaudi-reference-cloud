// INTEL CONFIDENTIAL
// Copyright (C) 2024 Intel Corporation
package mocks

import (
	"github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/vast/client"
)

type MockVastClient struct {
	client.ClientWithResponsesInterface
}

func w[T ~string | ~int | ~uint64 | ~bool](value T) *T {
	return &value
}
