// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Builder builds various artifacts from a specific commit.
// When run with "make universe-deployer", Builder runs from the HEAD commit.
// However, it executes "bazel build" in the workspace for the specific commit.
package builder

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	retry "github.com/sethvargo/go-retry"
)

type Builder struct {
	Commit                    string
	Components                []string
	SemanticVersion           string
	HomeDir                   string
	WorkspaceDir              string
	BazelBinary               string
	BazelStartupOpts          []string
	BazelBuildOpts            []string
	LegacyDefines             bool
	LegacyDeploymentArtifacts bool
	// Path to the deployment artifacts tar that will be written by Build.
	Output string
}

// Run Bazel to build deployment artifacts.
func (b Builder) Build(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("builder")
	log.Info("BEGIN")
	defer log.Info("END")

	deployment_artifacts_target := ""
	deployment_artifacts_output := ""
	if len(b.Components) == 0 {
		if b.LegacyDeploymentArtifacts {
			deployment_artifacts_target = "//deployment/universe_deployer:deployment_artifacts_tar"
			deployment_artifacts_output = "bazel-bin/deployment/universe_deployer/deployment_artifacts_tar.tar"
		} else {
			deployment_artifacts_target = "//deployment/universe_deployer/deployment_artifacts:deployment_artifacts_tar"
			deployment_artifacts_output = "bazel-bin/deployment/universe_deployer/deployment_artifacts/deployment_artifacts_tar.tar"
		}
	} else if len(b.Components) == 1 {
		component := b.Components[0]
		deployment_artifacts_target = "//deployment/universe_deployer/deployment_artifacts:deployment_artifacts_" + component + "_tar"
		deployment_artifacts_output = "bazel-bin/deployment/universe_deployer/deployment_artifacts/deployment_artifacts_" + component + "_tar.tar"
	} else {
		return fmt.Errorf("more than one component is unsupported")
	}

	if err := os.MkdirAll(filepath.Join(b.WorkspaceDir, "build/dynamic"), 0750); err != nil {
		return err
	}
	if err := writeDynamicFile(ctx, b.WorkspaceDir, "build/dynamic/DOCKER_TAG", b.SemanticVersion+"-"+b.Commit); err != nil {
		return err
	}
	if err := writeDynamicFile(ctx, b.WorkspaceDir, "build/dynamic/HELM_CHART_VERSION", b.SemanticVersion); err != nil {
		return err
	}
	if err := writeDynamicFile(ctx, b.WorkspaceDir, "build/dynamic/IDC_FULL_VERSION", b.SemanticVersion+"-"+b.Commit); err != nil {
		return err
	}

	// The file build/dynamic/universe_deployer.bzl must define these values but the actual values are irrelevant.
	dynamicUniverseDeployerBzl := "MAX_POOL_DIRS = 1\nPOOL_DIR = \"/tmp\"\n"
	if err := writeDynamicFile(ctx, b.WorkspaceDir, "build/dynamic/universe_deployer.bzl", dynamicUniverseDeployerBzl); err != nil {
		return err
	}

	// When executed with "make universe-deployer", this will result in a nested execution of Bazel (Bazel in Bazel).
	args := []string{}
	args = append(args, b.BazelStartupOpts...)
	args = append(args, "build")
	args = append(args,
		"--verbose_failures",
	)
	if b.LegacyDefines {
		// These defines are required for older commits.
		// This section should be removed once all components have been upgraded.
		args = append(args,
			"--define", "APPEND_CHART_HASH_TO_HELM_CHART_VERSION=True",
			"--define", "DOCKER_TAG=latest",
			"--define", "HELM_CHART_VERSION=0.0.1",
		)
	}
	args = append(args, b.BazelBuildOpts...)
	args = append(args,
		deployment_artifacts_target,
		"//go/pkg/universe_deployer/cmd/pusher",
		"@helm3_linux_amd64//:helm",
	)

	// Errors often occur when getting Go dependencies from the Internet.
	// Retry several times.
	backoff := retry.WithMaxRetries(6, retry.NewExponential(1*time.Second))
	err := retry.Do(ctx, backoff, func(ctx context.Context) error {
		cmd := exec.CommandContext(ctx, b.BazelBinary, args...)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "HOME="+b.HomeDir)
		cmd.Env = append(cmd.Env, "PATH=/bin")
		cmd.Dir = b.WorkspaceDir
		timeBuildStart := time.Now()
		err := util.RunCmd(ctx, cmd)
		log.Info("Build duration", "duration", time.Since(timeBuildStart))
		if err != nil {
			log.Error(err, "retryable error running bazel build")
			return retry.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	builtTar := filepath.Join(b.WorkspaceDir, deployment_artifacts_output)
	if err := util.CopyFile(builtTar, b.Output); err != nil {
		return err
	}
	return nil
}

func writeDynamicFile(_ context.Context, workspaceDir string, file string, contents string) error {
	return os.WriteFile(filepath.Join(workspaceDir, file), []byte(contents), 0640)
}
