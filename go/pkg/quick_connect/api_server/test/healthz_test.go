// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Healthz Tests", func() {
	It("Readyz should succeed", func() {
		request, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		Expect(response.Code).Should(Equal(http.StatusOK))
		Expect(response.Body.String()).Should(Equal("ok"))
	})
})
