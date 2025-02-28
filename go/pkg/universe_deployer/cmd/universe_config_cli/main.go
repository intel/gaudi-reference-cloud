// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// CLI for working with Universe Config files for Universe Deployer.

package main

import (
	"context"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	"github.com/spf13/cobra"
	"github.com/thediveo/enumflag/v2"
)

type RootArguments struct {
	GitRepositoryDir string
	GitRemote        string
}

var rootArguments RootArguments

type PrintArguments struct {
	RenderMode universe_config.RenderMode
	SortBy     universe_config.SortBy
}

var printArguments = PrintArguments{
	RenderMode: universe_config.RenderModePrettyTable,
	SortBy:     universe_config.SortByHierarchy,
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("universe_config_cli"))

	var rootCmd = &cobra.Command{
		Use:           "universe_config_cli",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	rootCmd.PersistentFlags().StringVar(&rootArguments.GitRepositoryDir, "git-repository-dir", "", "")
	rootCmd.PersistentFlags().StringVar(&rootArguments.GitRemote, "git-remote", "origin", "")

	// annotate
	cmdAnnotate := &cobra.Command{
		Use:   "annotate file...",
		Short: "Annotate Universe Config files",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runAnnotate,
	}
	rootCmd.AddCommand(cmdAnnotate)

	// test-annotate
	cmdTestAnnotate := &cobra.Command{
		Use:   "test-annotate file...",
		Short: "Test that Universe Config file annotations are up-to-date",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runTestAnnotate,
	}
	rootCmd.AddCommand(cmdTestAnnotate)

	// print
	cmdPrint := &cobra.Command{
		Use:   "print file...",
		Short: "Print Universe Config files in a pretty table",
		Args:  cobra.MinimumNArgs(1),
		RunE:  runPrint,
	}
	cmdPrint.Flags().Var(
		enumflag.New(&printArguments.RenderMode, "render-mode", universe_config.RenderModeIds, enumflag.EnumCaseSensitive),
		"render-mode", "'pretty' or 'csv'")
	cmdPrint.Flags().Var(
		enumflag.New(&printArguments.SortBy, "sort", universe_config.SortByIds, enumflag.EnumCaseSensitive),
		"sort", "'hierarchy', 'authorDate', or 'component'")
	rootCmd.AddCommand(cmdPrint)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Error(err, "Error")
		os.Exit(1)
	}
}

func runAnnotate(cmd *cobra.Command, args []string) error {
	return universe_config.AnnotateFiles(
		cmd.Context(),
		util.AbsFromWorkspaceList(args),
		util.AbsFromWorkspace(rootArguments.GitRepositoryDir),
		rootArguments.GitRemote,
		false,
	)
}

func runTestAnnotate(cmd *cobra.Command, args []string) error {
	return universe_config.AnnotateFiles(
		cmd.Context(),
		util.AbsFromWorkspaceList(args),
		util.AbsFromWorkspace(rootArguments.GitRepositoryDir),
		rootArguments.GitRemote,
		true,
	)
}

func runPrint(cmd *cobra.Command, args []string) error {
	return universe_config.PrintFiles(
		cmd.Context(),
		util.AbsFromWorkspaceList(args),
		util.AbsFromWorkspace(rootArguments.GitRepositoryDir),
		rootArguments.GitRemote,
		printArguments.RenderMode,
		printArguments.SortBy,
	)
}
