// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package mocks

import (
	v4 "github.com/intel-innersource/applications.infrastructure.idcstorage.sds-controller/services/storage_controller/pkg/backend/weka/client/v4"
)

type MockWekaClient struct {
	v4.ClientWithResponsesInterface
}

func w[T ~string | ~int | ~uint64 | ~bool](value T) *T {
	return &value
}
