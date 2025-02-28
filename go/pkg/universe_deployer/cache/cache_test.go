// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cache

import (
	"context"
	"errors"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache Tests", func() {
	It("Basic", func() {
		ctx := context.Background()
		tempDir, err := os.MkdirTemp("", "universe_deployer_cache_test_")
		Expect(err).Should(Succeed())
		defer os.RemoveAll(tempDir)

		By("Create cache")
		c, err := New(ctx, tempDir)
		Expect(err).Should(Succeed())

		By("GetFile of non-existent key should return ErrNotFound")
		key1 := "key1"
		_, err = c.GetFile(ctx, key1)
		Expect(err).ShouldNot(Succeed())
		Expect(errors.Is(err, &ErrNotFound{})).Should(BeTrue())

		By("IsCached of non-existent key should return false")
		key2 := "key2"
		cached, err := c.IsCached(ctx, key2)
		Expect(err).Should(Succeed())
		Expect(cached).Should(BeFalse())

		By("Adding file to cache")
		tempFilePath, err := c.GetTempFilePath(ctx, key1)
		Expect(err).Should(Succeed())
		By("Writing file " + tempFilePath)
		Expect(os.WriteFile(tempFilePath, []byte("value for "+key1), 0640)).Should(Succeed())
		fileResult, err := c.MoveFileToCache(ctx, key1, tempFilePath)
		Expect(err).Should(Succeed())
		Expect(fileResult.Path).ShouldNot(BeEmpty())

		By("GetFile of added file should succeed")
		fileResultGet, err := c.GetFile(ctx, key1)
		Expect(err).Should(Succeed())
		Expect(fileResultGet.Path).Should(Equal(fileResult.Path))

		By("IsCached of added file should return true")
		cached, err = c.IsCached(ctx, key1)
		Expect(err).Should(Succeed())
		Expect(cached).Should(BeTrue())
	})
})

var _ = Describe("ensureValidKey Tests", func() {
	ctx := context.Background()
	It("ensureValidKey should accept a valid key", func() {
		Expect(ensureValidKey(ctx, "key-1_1.tar")).Should(Succeed())
	})
	It("ensureValidKey should reject an empty key", func() {
		Expect(ensureValidKey(ctx, "")).ShouldNot(Succeed())
	})
	It("ensureValidKey should reject a key with a '/'", func() {
		Expect(ensureValidKey(ctx, "path/key1")).ShouldNot(Succeed())
	})
	It("ensureValidKey should reject the key '.'", func() {
		Expect(ensureValidKey(ctx, ".")).ShouldNot(Succeed())
	})
	It("ensureValidKey should reject the key '..'", func() {
		Expect(ensureValidKey(ctx, "..")).ShouldNot(Succeed())
	})
})
