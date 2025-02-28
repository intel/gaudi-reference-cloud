// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"fmt"
	"log"

	root "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/idccli/cmd/root"
	"github.com/spf13/cobra"
)

func main() {
	err := mainRun()
	if err != nil {
		log.Fatalf("Error occurred: %v\n", err)
	}
}

func mainRun() error {
	ctx := context.Background()

	rootCmd, err := root.NewCmdRoot()
	if err != nil {
		return fmt.Errorf("failed to create root command: %w", err)
	}

	rootCmd.SetHelpCommand(&cobra.Command{
		Use:   "help-all",
		Short: "Get help for all subcommands",
		Long:  "Display help for all subcommands of the root command.",
		Run: func(cmd *cobra.Command, args []string) {
			for _, subCmd := range rootCmd.Commands() {
				subCmd.Help()
				fmt.Println()
			}
		},
	})

	if _, err := rootCmd.ExecuteContextC(ctx); err != nil {
		return err
	}
	return nil
}
