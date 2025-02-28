// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QuickConnectProxy Tests", func() {
	DescribeTable("Validating the path parameters",
		func(path string, responseCode int) {
			request, _ := http.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).Should(Equal(responseCode))
		},
		Entry("valid path", "/v1/connect/778464342572/55afcbd4-06bb-4cf9-a199-2bc3ddb2783e/", http.StatusUnauthorized),
		Entry("cloudaccountid is wrong size", "/v1/connect/778464/55afcbd4-06bb-4cf9-a199-2bc3ddb2783e/", http.StatusBadRequest),
		Entry("cloudaccountid has invalid characters", "/v1/connect/A7b4c4342572/55afcbd4-06bb-4cf9-a199-2bc3ddb2783e/", http.StatusBadRequest),
		Entry("cloudaccountid is missing", "/v1/connect//55afcbd4-06bb-4cf9-a199-2bc3ddb2783e/", http.StatusBadRequest),
		Entry("instanceid is wrong size", "/v1/connect/778464342572/55afcbd4-06bb-4cf9-a199/", http.StatusBadRequest),
		Entry("instanceid has invalid characters", "/v1/connect/778464342572/55ghijk4-06bb-4cf9-a199-2bc3ddb2783e/", http.StatusBadRequest),
		Entry("instanceid is missing", "/v1/connect/778464342572//", http.StatusBadRequest),
	)
})
