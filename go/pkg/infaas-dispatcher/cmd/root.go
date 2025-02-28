// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package cmd

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "infaas-dispatcher",
		Short: "A load balancer for inference services",
		Long:  `A load balancer for inference services`,
		Run: func(cmd *cobra.Command, args []string) {
			// not really used now, It's just a preparation for a maybe future set of commands
			// the future common (root) flags will reside in this file
		},
	}

	rootCmd.AddCommand(NewServerCommand(), NewGenCommand(), NewRemoteGenCommand(), NewHealthCommand())
	return rootCmd
}
