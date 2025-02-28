// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package utils

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func createLabels(count int) map[string]string {
	labels := map[string]string{}
	for i := 0; i < count; i++ {
		labels[fmt.Sprintf("key%d", i)] = "value"
	}
	return labels
}

var _ = Describe("Labels Tests", Serial, func() {
	It("Single simple key/value should succeed", func() {
		Expect(ValidateLabels(map[string]string{"key": "value"})).Should(Succeed())
	})

	It("Single complex key/value should succeed", func() {
		Expect(ValidateLabels(map[string]string{"my-Label_key.com": "my-Label_value.com"})).Should(Succeed())
	})

	It("Single complex key/value with a hostname prefix should succeed", func() {
		Expect(ValidateLabels(map[string]string{"cloud.intel.com/name": "value"})).Should(Succeed())
	})

	It("Empty value should succeed", func() {
		Expect(ValidateLabels(map[string]string{"key": ""})).Should(Succeed())
	})

	It("Empty key should fail", func() {
		Expect(ValidateLabels(map[string]string{"": "value"})).ShouldNot(Succeed())
	})

	It("Key with invalid character should fail", func() {
		Expect(ValidateLabels(map[string]string{"invalid!!!.label.com/name": "value"})).ShouldNot(Succeed())
	})

	It("Value with invalid character should fail", func() {
		Expect(ValidateLabels(map[string]string{"key": "invalid.*value"})).ShouldNot(Succeed())
	})

	It("Max key/value length should succeed", func() {
		Expect(ValidateLabels(map[string]string{strings.Repeat("k", 63): strings.Repeat("v", 63)})).Should(Succeed())
	})

	It("Long key should fail", func() {
		Expect(ValidateLabels(map[string]string{strings.Repeat("k", 64): "value"})).ShouldNot(Succeed())
	})

	It("Long value should fail", func() {
		Expect(ValidateLabels(map[string]string{"key": strings.Repeat("v", 64)})).ShouldNot(Succeed())
	})

	It("Maximum number of labels should succeed", func() {
		Expect(ValidateLabels(createLabels(MaxNumberOfLabels))).Should(Succeed())
	})

	It("Too many labels should fail", func() {
		Expect(ValidateLabels(createLabels(MaxNumberOfLabels + 1))).ShouldNot(Succeed())
	})
})
