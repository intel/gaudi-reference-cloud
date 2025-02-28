// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cloud_init

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndent(t *testing.T) {
	// things run through this indent filter are individual templated values
	// these are most likely to be multiple lines long but no newline at the end
	t.Run("templated directory", func(t *testing.T) {
		contents := "item: one\nitem: two"
		expected := "  item: one\n  item: two"
		assert.Equal(t, expected, Indent(contents, 2))
	})

	t.Run("templated list", func(t *testing.T) {
		contents := "listheader:\n- item: one\n- item: two"
		expected := "  listheader:\n  - item: one\n  - item: two"
		assert.Equal(t, expected, Indent(contents, 2))
	})

	t.Run("templated string", func(t *testing.T) {
		contents := "string"
		expected := "  string"
		assert.Equal(t, expected, Indent(contents, 2))
	})
}
