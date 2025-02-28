// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package hasher

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hasher Tests", func() {
	It("Hashes of same content should be equal", func() {
		ctx := context.Background()
		h1 := New()
		Expect(h1.AddString(ctx, "label1", "string1")).Should(Succeed())
		sum1 := h1.Sum(ctx)
		h2 := New()
		Expect(h2.AddString(ctx, "label1", "string1")).Should(Succeed())
		sum2 := h2.Sum(ctx)
		Expect(sum2).Should(Equal(sum1))
	})

	It("Hashes with different strings should be different", func() {
		ctx := context.Background()
		h1 := New()
		Expect(h1.AddString(ctx, "label1", "string1")).Should(Succeed())
		sum1 := h1.Sum(ctx)
		h2 := New()
		Expect(h2.AddString(ctx, "label1", "string2")).Should(Succeed())
		sum2 := h2.Sum(ctx)
		Expect(sum2).ShouldNot(Equal(sum1))
	})

	It("Hashes with different labels should be different", func() {
		ctx := context.Background()
		h1 := New()
		Expect(h1.AddString(ctx, "label1", "string1")).Should(Succeed())
		sum1 := h1.Sum(ctx)
		h2 := New()
		Expect(h2.AddString(ctx, "label2", "string1")).Should(Succeed())
		sum2 := h2.Sum(ctx)
		Expect(sum2).ShouldNot(Equal(sum1))
	})
})
