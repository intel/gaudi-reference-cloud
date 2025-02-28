// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Evaulate the maximum prefix length function", Serial, func() {
	It("GetMaximumPrefixLength should succeed", func() {
		// 27 should be returned as prefix - boundary use cases (similar to staging and prod env)
		Expect(GetMaximumPrefixLength(26)).Should(Equal(int32(27)))
		Expect(GetMaximumPrefixLength(27)).Should(Equal(int32(27)))

		// 26 should be returned as prefix - boundary use cases
		Expect(GetMaximumPrefixLength(28)).Should(Equal(int32(26)))
		Expect(GetMaximumPrefixLength(32)).Should(Equal(int32(26)))
		Expect(GetMaximumPrefixLength(59)).Should(Equal(int32(26)))

		// 24 should be returned as prefix - boundary use cases (similar to staging and prod env)
		Expect(GetMaximumPrefixLength(124)).Should(Equal(int32(24)))
		Expect(GetMaximumPrefixLength(251)).Should(Equal(int32(24)))

		// 22 should be returned as prefix - boundary use cases
		Expect(GetMaximumPrefixLength(1019)).Should(Equal(int32(22)))
		Expect(GetMaximumPrefixLength(508)).Should(Equal(int32(22)))

		// remaining use case validation
		Expect(GetMaximumPrefixLength(507)).Should(Equal(int32(23)))
		Expect(GetMaximumPrefixLength(252)).Should(Equal(int32(23)))
		Expect(GetMaximumPrefixLength(256)).Should(Equal(int32(23)))
		Expect(GetMaximumPrefixLength(16379)).Should(Equal(int32(18)))
		Expect(GetMaximumPrefixLength(0)).Should(Equal(int32(32)))
		Expect(GetMaximumPrefixLength(1)).Should(Equal(int32(32)))
		Expect(GetMaximumPrefixLength(2)).Should(Equal(int32(29)))
		Expect(GetMaximumPrefixLength(-1)).Should(Equal(int32(32)))
	})
})
