// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package agent

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestResolveDispatcher(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resolvedAddr := resolveDispatcherAddress(ctx, logr.Logger{}, "localhost")

	assert.Equal(t, "localhost", resolvedAddr)
}
