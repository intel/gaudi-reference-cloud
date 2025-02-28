// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/infaas-dispatcher/cmd"
)

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	if err := cmd.NewRootCommand().ExecuteContext(ctx); err != nil {
		_, err = fmt.Fprintf(os.Stderr, "[ERROR] command failed with error: %s\n", err)
		if err != nil {
			fmt.Printf("Fprintf error: %s\n", err)
		}
		os.Exit(1)
	}
}
