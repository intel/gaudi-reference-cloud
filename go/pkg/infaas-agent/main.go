// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-agent/cmd"
	"os"
	"os/signal"
)

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	if err := cmd.NewRootCommand().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
