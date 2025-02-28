// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Create a release:
//   - Upload deployment artifacts to Artifactory.
// This is used by create_releases in deployment/universe_deployer/universe_deployer.bzl.

package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/artifactory"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	flag "github.com/spf13/pflag"
)

type Arguments struct {
	ArtifactRepositoryUrl     string
	ArtifactRepositoryUrlFile string
	Commit                    string
	Component                 string
	DeploymentArtifactsTar    string
	Output                    string
	RetentionDays             int
	RunfilesDir               string
	SecretsDir                string
}

func parseArgs() Arguments {
	var args Arguments

	flag.StringVar(&args.ArtifactRepositoryUrl, "artifact-repository-url", "", "Artifactory URL")
	flag.StringVar(&args.ArtifactRepositoryUrlFile, "artifact-repository-url-file", "", "File that contains the Artifactory URL")
	flag.StringVar(&args.Commit, "commit", "", "Git commit hash")
	flag.StringVar(&args.Component, "component", "", "Component")
	flag.StringVar(&args.DeploymentArtifactsTar, "deployment-artifacts-tar", "", "File to upload")
	flag.StringVar(&args.Output, "output", "", "If provided, this empty file will be created when complete")
	flag.IntVar(&args.RetentionDays, "retention-days", util.DeploymentArtifactsRetentionDays(), "Retention days")

	flag.Parse()

	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	args.RunfilesDir = workingDir

	args.DeploymentArtifactsTar = util.AbsFromWorkspace(args.DeploymentArtifactsTar)
	args.Output = util.AbsFromWorkspace(args.Output)
	args.SecretsDir = filepath.Join(args.RunfilesDir, "local/secrets")

	util.EnsureRequiredStringFlag("commit")
	util.EnsureRequiredStringFlag("component")
	util.EnsureRequiredStringFlag("deployment-artifacts-tar")

	return args
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("create_release"))
	log.Info("BEGIN")
	defer log.Info("END")

	err := func() error {
		var args = parseArgs()
		log.Info("args", "args", args)
		log.Info("Environment", "env", os.Environ())

		artifactRepositoryUrl := args.ArtifactRepositoryUrl
		if artifactRepositoryUrl == "" {
			if args.ArtifactRepositoryUrlFile == "" {
				return fmt.Errorf("one of artifact-repository-url or artifact-repository-url-file is required")
			}
			fileBytes, err := os.ReadFile(args.ArtifactRepositoryUrlFile)
			if err != nil {
				return err
			}
			artifactRepositoryUrl = string(fileBytes)
		}

		if err := uploadDeploymentArtifacts(
			ctx,
			args.DeploymentArtifactsTar,
			args.Component,
			args.Commit,
			artifactRepositoryUrl,
			args.SecretsDir,
			args.RetentionDays,
		); err != nil {
			return err
		}

		// Write empty output file.
		// This will be used as the Bazel build target to ensure that the release is created.
		// Bazel caching will minimize duplicate executions.
		if args.Output != "" {
			if err := os.WriteFile(args.Output, nil, 0640); err != nil {
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

func uploadDeploymentArtifacts(ctx context.Context, filename string, component string, commit string, artifactRepositoryUrl string, secretsDir string, retentionDays int) error {
	parsedUrl, err := url.Parse(artifactRepositoryUrl)
	if err != nil {
		return err
	}
	artifactory, err := artifactory.NewFromSecretsDir(ctx, secretsDir)
	if err != nil {
		return err
	}
	artifactory.RetentionDays = retentionDays
	artifactUrl, err := util.DeploymentArtifactsTarUrl(*parsedUrl, component, commit)
	if err != nil {
		return err
	}
	if err := artifactory.Upload(ctx, filename, artifactUrl); err != nil {
		return err
	}
	return nil
}
