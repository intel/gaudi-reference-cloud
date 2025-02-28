// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "maas_gw",
		Short: "Model as a Service Gateway",
	}

	rootCmd.AddCommand(NewStartCommand())
	rootCmd.AddCommand(NewDevCommand())
	return rootCmd
}
