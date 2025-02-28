// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Builder Pusher builds various artifacts from a specific commit, then pushes container images and Helm charts to Harbor.
// The binary that is executed is built from the working tree in which "make universe-deployer" runs.
// However, it executes Pusher (pushes containers and Helm charts) and Manifests Generator built from the component-specific commit referenced in the Universe Config File.

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/builder"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	"github.com/sethvargo/go-retry"
	flag "github.com/spf13/pflag"
)

type Arguments struct {
	BazelBinary               string
	BazelBuildOpts            []string
	BazelStartupOpts          []string
	Commit                    string
	Components                []string
	GitArchive                string
	LegacyDeploymentArtifacts bool
	MaxPoolDirs               int
	Output                    string
	PatchTar                  string
	PoolDir                   string
	PoolTimeout               time.Duration
	SemanticVersion           string
	SkipPush                  bool
	Snapshot                  bool
	UniverseConfig            string
}

func parseArgs(ctx context.Context) Arguments {
	var args Arguments

	flag.StringVar(&args.BazelBinary, "bazel-binary", "bazel", "Path to bazel binary")
	flag.StringArrayVar(&args.BazelBuildOpts, "bazel-build-opt", nil, "Bazel build options")
	flag.StringArrayVar(&args.BazelStartupOpts, "bazel-startup-opt", nil, "Bazel startup options")
	flag.StringVar(&args.Commit, "commit", "", "Git commit hash")
	flag.StringArrayVar(&args.Components, "component", nil, "Component")
	flag.StringVar(&args.GitArchive, "git-archive", "", "Path to Git archive tar file for the commit")
	flag.BoolVar(&args.LegacyDeploymentArtifacts, "legacy-deployment-artifacts", false, "If true, use the legacy Bazel target to build deployment artifacts")
	flag.IntVar(&args.MaxPoolDirs, "max-pool-dirs", 1,
		"The number of directories within pool-dir. Each directory supports a single concurrent Bazel invocation.")
	flag.StringVar(&args.Output, "output", "", "Path to output tar file")
	flag.StringVar(&args.PatchTar, "patch-tar", "", "Optional path to a tar file that will be extracted, replacing any files extracted from git-archive")
	flag.StringVar(&args.PoolDir, "pool-dir", "/tmp",
		"All intermediate files including Bazel local cache files will be stored within this directory. "+
			"This directory should be persistent to allow the cached data to be reused.")
	flag.DurationVar(&args.PoolTimeout, "pool-timeout", 60*time.Minute,
		"Wait for an available directory in pool-dir for this long.")
	flag.StringVar(&args.SemanticVersion, "semantic-version", "0.0.1", "IDC semantic version in format 1.2.3")
	flag.BoolVar(&args.SkipPush, "skip-push", false, "If true, do not push containers and Helm charts")
	flag.BoolVar(&args.Snapshot, "snapshot", false, "If true, use the current working directory instead of git-archive")
	flag.StringVar(&args.UniverseConfig, "universe-config", "", "Path to the Universe Config file")

	flag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	flag.Parse()

	util.EnsureRequiredStringFlag("bazel-binary")
	util.EnsureRequiredStringFlag("commit")
	if !args.Snapshot {
		util.EnsureRequiredStringFlag("git-archive")
	}
	util.EnsureRequiredStringFlag("semantic-version")
	util.EnsureRequiredStringFlag("universe-config")

	return args
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	args := parseArgs(ctx)
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("builder_pusher").WithValues("commit", args.Commit).WithValues("component", args.Components))
	log.Info("BEGIN")
	defer log.Info("END")

	err := func() error {
		log.Info("args", "args", args)
		log.Info("Environment", "env", os.Environ())

		workingDir, err := os.Getwd()
		if err != nil {
			return err
		}
		log.Info("workingDir", "workingDir", workingDir)

		if !filepath.IsAbs(args.UniverseConfig) {
			args.UniverseConfig = filepath.Join(workingDir, args.UniverseConfig)
		}

		workspaceDir := ""
		homeDir := ""
		bazelBinary := ""
		runfilesDir := workingDir

		tempDir, err := os.MkdirTemp("", "universe_deployer_builder_pusher_")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		if args.Snapshot {
			workspaceDir = workingDir
			homeDir = os.Getenv("HOME")
			bazelBinary = args.BazelBinary
		} else {
			bazelPool, err := util.NewBazelPool(args.MaxPoolDirs, args.PoolDir, 0700)
			if err != nil {
				return err
			}
			ctxPool, cancelPool := context.WithTimeout(ctx, args.PoolTimeout)
			defer cancelPool()
			poolDir, poolLock, err := bazelPool.GetDirectoryFromPool(ctxPool)
			if err != nil {
				return err
			}
			defer poolLock.Unlock()
			log.Info("Obtained directory from pool", "poolDir", poolDir)

			workspaceDir, homeDir, err = bazelPool.PrepareWorkspace(poolDir)
			if err != nil {
				return err
			}

			// Delete contents of workspace directory so we can extract on top of it without extra files being left.
			timeDeleteStart := time.Now()
			if err := util.DeleteDirectoryContents(workspaceDir); err != nil {
				return err
			}
			log.Info("Delete duration", "duration", time.Since(timeDeleteStart))

			// Extract Git archive.
			cmd := exec.CommandContext(ctx, "/bin/tar",
				"-C", workspaceDir,
				"-x",
				"-f", args.GitArchive,
			)
			timeExtractStart := time.Now()
			if err := util.RunCmd(ctx, cmd); err != nil {
				return err
			}
			log.Info("Extract duration", "duration", time.Since(timeExtractStart))

			bazelBinary = filepath.Join(runfilesDir, "./external/bazel_binaries_bazelisk/bazelisk")
		}

		log.Info("Paths", "runFilesDir", runfilesDir, "workspaceDir", workspaceDir, "homeDir", homeDir, "bazelBinary", bazelBinary)

		if err := applyPatchesToCommit(ctx, workspaceDir, args.PatchTar); err != nil {
			return err
		}

		bazelInfoInput, err := util.BazelInfo(ctx, workspaceDir, homeDir, bazelBinary)
		log.Info("Bazel Info", "bazelInfoInput", bazelInfoInput, "err", err)
		if err != nil {
			return err
		}

		// Build deployment artifacts and pusher binary.
		b := builder.Builder{
			Commit:                    args.Commit,
			Components:                args.Components,
			SemanticVersion:           args.SemanticVersion,
			HomeDir:                   homeDir,
			WorkspaceDir:              workspaceDir,
			BazelBinary:               bazelBinary,
			BazelStartupOpts:          args.BazelStartupOpts,
			BazelBuildOpts:            args.BazelBuildOpts,
			LegacyDefines:             true,
			LegacyDeploymentArtifacts: args.LegacyDeploymentArtifacts,
			Output:                    args.Output,
		}
		if err := b.Build(ctx); err != nil {
			return err
		}

		builtHelmBinary := filepath.Join(bazelInfoInput["execution_root"], filepaths.HelmBinary)
		builtHelmfileBinary := filepath.Join(bazelInfoInput["execution_root"], filepaths.HelmfileBinary)

		// Create and use a copy of the built binaries because they may be deleted by "bazel run" executed by pusher.
		helmBinary := filepath.Join(tempDir, filepath.Base(filepaths.HelmBinary))
		helmfileBinary := filepath.Join(tempDir, filepath.Base(filepaths.HelmfileBinary))
		if err := util.CopyFile(builtHelmBinary, helmBinary); err != nil {
			return err
		}
		if err := util.CopyFile(builtHelmfileBinary, helmfileBinary); err != nil {
			return err
		}

		// Run pusher binary created by above Build function.
		if !args.SkipPush {
			pusherBinary := filepath.Join(workspaceDir, "bazel-bin/go/pkg/universe_deployer/cmd/pusher/pusher_/pusher")
			secretsDir := filepath.Join(runfilesDir, "local/secrets")
			pusherArgs := []string{
				"--bazel-binary", bazelBinary,
				"--commit", args.Commit,
				"--helm-binary", helmBinary,
				"--helmfile-binary", helmfileBinary,
				"--secrets", secretsDir,
				fmt.Sprintf("--snapshot=%v", args.Snapshot),
				"--universe-config", args.UniverseConfig,
				"--workspace-dir", workspaceDir,
			}
			for _, opt := range args.BazelBuildOpts {
				pusherArgs = append(pusherArgs, "--bazel-run-opt", opt)
			}
			for _, opt := range args.BazelStartupOpts {
				pusherArgs = append(pusherArgs, "--bazel-startup-opt", opt)
			}
			for _, opt := range args.Components {
				pusherArgs = append(pusherArgs, "--component", opt)
			}

			// Errors often occur when pushing to Harbor.
			// Retry several times.
			backoff := retry.WithMaxRetries(6, retry.NewExponential(1*time.Second))
			err := retry.Do(ctx, backoff, func(ctx context.Context) error {
				cmd := exec.CommandContext(ctx, pusherBinary, pusherArgs...)
				cmd.Dir = workspaceDir
				cmd.Env = os.Environ()
				cmd.Env = append(cmd.Env, "HOME="+homeDir)
				timePusherStart := time.Now()
				err := util.RunCmd(ctx, cmd)
				log.Info("Pusher duration", "duration", time.Since(timePusherStart))
				if err != nil {
					log.Error(err, "retryable error running pusher")
					return retry.RetryableError(err)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	}()
	if err != nil {
		log.Error(err, "error")
		os.Exit(1)
	}
}

// Extract a tar file on top of the extracted commit.
// This can be used to update .bazelrclocal with the Bazel remote cache options.
// Such a patch must not cause the built deployment artifacts to change for the same commit.
func applyPatchesToCommit(ctx context.Context, workspaceDir string, patchTar string) error {
	log := log.FromContext(ctx).WithName("applyPatchesToCommit")
	log.Info("BEGIN")
	defer log.Info("END")
	if patchTar == "" {
		return nil
	}
	cmd := exec.CommandContext(ctx, "/bin/tar",
		"-C", workspaceDir,
		"-xv",
		"-f", patchTar,
	)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}
	return nil
}
