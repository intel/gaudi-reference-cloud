// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Pusher pushes container images and Helm charts to Docker and Helm registries.
// The binary that is executed is built from the component-specific commit referenced in the Universe Config File.

package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/reader"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/pusher"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	flag "github.com/spf13/pflag"
)

type Arguments struct {
	BazelBinary      string
	BazelStartupOpts []string
	BazelRunOpts     []string
	Commit           string
	Components       []string
	HelmBinary       string
	HelmfileBinary   string
	Output           string
	SecretsDir       string
	SemanticVersion  string
	Snapshot         bool
	UniverseConfig   string
	WorkspaceDir     string
}

func parseArgs() Arguments {
	var args Arguments

	flag.StringVar(&args.BazelBinary, "bazel-binary", "", "Path to bazel binary")
	flag.StringArrayVar(&args.BazelRunOpts, "bazel-run-opt", nil, "Bazel run options")
	flag.StringArrayVar(&args.BazelStartupOpts, "bazel-startup-opt", nil, "Bazel startup options")
	flag.StringVar(&args.Commit, "commit", "", "Git commit hash")
	flag.StringArrayVar(&args.Components, "component", nil, "Component")
	flag.StringVar(&args.HelmBinary, "helm-binary", "helm", "Path to helm binary")
	flag.StringVar(&args.HelmfileBinary, "helmfile-binary", "helmfile", "Path to helmfile binary")
	flag.StringVar(&args.Output, "output", "", "If provided, this empty file will be created when complete")
	flag.StringVar(&args.SecretsDir, "secrets", "", "Path to the directory containing Docker and Helm registry secrets")
	flag.StringVar(&args.SemanticVersion, "semantic-version", "0.0.1", "IDC semantic version")
	flag.BoolVar(&args.Snapshot, "snapshot", false, "If false, the Universe Config will be filtered by the commit")
	flag.StringVar(&args.UniverseConfig, "universe-config", "", "Path to the Universe Config file")
	flag.StringVar(&args.WorkspaceDir, "workspace-dir", "", "Path to the Bazel workspace directory containing the IDC monorepo")

	flag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	flag.Parse()

	util.EnsureRequiredStringFlag("bazel-binary")
	util.EnsureRequiredStringFlag("commit")
	util.EnsureRequiredStringFlag("helm-binary")
	util.EnsureRequiredStringFlag("workspace-dir")
	util.EnsureRequiredStringFlag("secrets")
	util.EnsureRequiredStringFlag("universe-config")

	return args
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("pusher"))
	log.Info("BEGIN")
	defer log.Info("END")

	err := func() error {
		var args = parseArgs()
		log.Info("args", "args", args)

		tempDir, err := os.MkdirTemp("", "universe_deployer_pusher_")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		universeConfig, err := universe_config.NewUniverseConfigFromFile(ctx, args.UniverseConfig)
		if err != nil {
			return err
		}
		if !args.Snapshot {
			universeConfig = universeConfig.Trimmed(ctx, args.Commit)
		}
		log.Info("universeConfig", "universeConfig", universeConfig)

		envConfigReader := reader.Reader{
			HelmBinary:        args.HelmBinary,
			HelmfileBinary:    args.HelmfileBinary,
			HelmfileConfigDir: filepath.Join(args.WorkspaceDir, filepaths.HelmfileConfigDir),
		}
		multipleEnvConfig, err := envConfigReader.ReadUniverse(ctx, universeConfig)
		if err != nil {
			return err
		}

		p := pusher.Pusher{
			Commit:            args.Commit,
			Components:        args.Components,
			SemanticVersion:   args.SemanticVersion,
			UniverseConfig:    universeConfig,
			WorkspaceDir:      args.WorkspaceDir,
			SecretsDir:        args.SecretsDir,
			BazelBinary:       args.BazelBinary,
			BazelStartupOpts:  args.BazelStartupOpts,
			BazelRunOpts:      args.BazelRunOpts,
			HelmBinary:        args.HelmBinary,
			MultipleEnvConfig: multipleEnvConfig,
		}
		if err := p.Push(ctx); err != nil {
			return err
		}

		// Write empty output file.
		// This will be used as a dependency of manifests_generator since it requires Helm charts to be pushed.
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
