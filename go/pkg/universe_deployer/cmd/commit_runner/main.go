// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Commit Runner extracts a tar file to a temporary directory, then runs a binary from the temporary directory.
// The binary that is executed is built from the working tree in which "make universe-deployer" runs.

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

type Arguments struct {
	Input      string
	RunCmdName string
	RunCmdArgs []string
}

func parseArgs() Arguments {
	var args Arguments
	args.Input = os.Args[1]
	args.RunCmdName = os.Args[2]
	args.RunCmdArgs = os.Args[3:]
	return args
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("commit_runner"))
	log.Info("BEGIN")
	defer log.Info("END")

	err := func() error {
		var args = parseArgs()
		log.Info("args", "args", args)

		workspaceDir, err := os.Getwd()
		if err != nil {
			return err
		}
		log.Info("workspaceDir", "workspaceDir", workspaceDir)

		tempDir, err := os.MkdirTemp("", "universe_deployer_commit_runner_")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		commitDir := filepath.Join(tempDir, "commit")
		if err := os.Mkdir(commitDir, 0750); err != nil {
			return err
		}

		homeDir := filepath.Join(tempDir, "home")
		if err := os.Mkdir(homeDir, 0750); err != nil {
			return err
		}

		cmd := exec.CommandContext(ctx, "/bin/tar",
			"-C", commitDir,
			"-x",
			"-f", args.Input,
		)
		timeExtractStart := time.Now()
		if err := util.RunCmd(ctx, cmd); err != nil {
			return err
		}
		log.Info("Extract duration", "duration", time.Since(timeExtractStart))

		// This runs manifests_generator.
		runCmdName := filepath.Join(commitDir, args.RunCmdName)
		runCmdArgs := append(append([]string{}, args.RunCmdArgs...), "--commit-dir", commitDir)
		cmd = exec.CommandContext(ctx, runCmdName, runCmdArgs...)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "HOME="+homeDir)
		timeRunStart := time.Now()
		if err := util.RunCmd(ctx, cmd); err != nil {
			return err
		}
		log.Info("Run duration", "duration", time.Since(timeRunStart))

		return nil
	}()
	if err != nil {
		log.Error(err, "error")
		os.Exit(1)
	}
}
