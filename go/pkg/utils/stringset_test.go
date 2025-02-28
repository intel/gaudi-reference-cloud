// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Intersect", func() {
	Context("test Intersect function", func() {
		It("when the input lists contain both common and unique elements", func() {
			list1 := []string{"pool1", "pool2", "pool3"}
			list2 := []string{"pool2", "pool3", "pool5"}
			expected := []string{"pool2", "pool3"}
			Expect(Intersect(list1, list2)).To(ConsistOf(expected))
		})

		It("when the input lists contain common and a few duplicate elements", func() {
			list1 := []string{"pool1", "pool2", "pool3", "pool2", "pool3"}
			list2 := []string{"pool2", "pool3", "pool5", "pool5", "pool2"}
			expected := []string{"pool2", "pool3"}
			Expect(Intersect(list1, list2)).To(ConsistOf(expected))
		})

		It("when the input lists have no common elements", func() {
			list1 := []string{"pool1", "pool2", "pool3", "pool2", "pool3"}
			list2 := []string{"pool9", "pool8", "pool5", "pool5", "pool9"}
			Expect(Intersect(list1, list2)).To(BeEmpty())
		})

		It("when one of the input slices is empty", func() {
			list1 := []string{}
			list2 := []string{"pool2", "pool3", "pool4"}
			Expect(Intersect(list1, list2)).To(BeEmpty())
		})
	})
})
