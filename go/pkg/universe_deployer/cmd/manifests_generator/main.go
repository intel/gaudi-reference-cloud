// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Manifest Generator generates Argo CD manifests.
// The binary that is executed is built from the component-specific commit referenced in the Universe Config File.

package main

import (
	"context"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/reader"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/manifests_generator"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	flag "github.com/spf13/pflag"
)

type Arguments struct {
	Commit                 string
	Components             []string
	ConfigCommit           string
	DeploymentArtifactsDir string
	DefaultChartRegistry   string
	Output                 string
	SecretsDir             string
	Snapshot               bool
	UniverseConfig         string
}

func parseArgs() Arguments {
	var args Arguments

	flag.StringVar(&args.Commit, "commit", "", "Git commit hash of the application")
	flag.StringVar(&args.ConfigCommit, "config-commit", "", "Git commit hash of the configuration")
	flag.StringVar(&args.DeploymentArtifactsDir, "commit-dir", "", "Path to the directory containing the deployment artifacts for the commit")
	flag.StringArrayVar(&args.Components, "component", nil, "Component")
	flag.StringVar(&args.DefaultChartRegistry, "default-chart-registry", "", "Helm chart registry for IDC applications")
	flag.StringVar(&args.Output, "output", "", "Path to the output manifests tar file")
	flag.StringVar(&args.SecretsDir, "secrets-dir", "", "Path to the directory containing any secrets required to generate Argo CD manifests")
	flag.BoolVar(&args.Snapshot, "snapshot", false, "If false, the Universe Config will be filtered by the commit")
	flag.StringVar(&args.UniverseConfig, "universe-config", "", "Path to the Universe Config file")

	flag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	flag.Parse()

	util.EnsureRequiredStringFlag("commit")
	util.EnsureRequiredStringFlag("commit-dir")
	util.EnsureRequiredStringFlag("output")
	util.EnsureRequiredStringFlag("universe-config")

	return args
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("manifests_generator"))
	log.Info("BEGIN")
	defer log.Info("END")
	err := func() error {
		var args = parseArgs()
		log.Info("args", "args", args)

		tempDir, err := os.MkdirTemp("", "universe_deployer_manifests_generator_")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		universeConfig, err := universe_config.NewUniverseConfigFromFile(ctx, args.UniverseConfig)
		if err != nil {
			return err
		}

		envConfigReader := reader.Reader{
			DeploymentArtifactsDir: args.DeploymentArtifactsDir,
		}
		multipleEnvConfig, err := envConfigReader.ReadUniverse(ctx, universeConfig)
		if err != nil {
			return err
		}

		manifestsGenerator := &manifests_generator.ManifestsGenerator{
			DeploymentArtifactsDir:       args.DeploymentArtifactsDir,
			SecretsDir:                   args.SecretsDir,
			UniverseConfig:               universeConfig,
			MultipleEnvConfig:            multipleEnvConfig,
			Commit:                       args.Commit,
			Components:                   args.Components,
			ConfigCommit:                 args.ConfigCommit,
			Snapshot:                     args.Snapshot,
			ManifestsTar:                 args.Output,
			OverrideDefaultChartRegistry: args.DefaultChartRegistry,
		}
		if _, err := manifestsGenerator.GenerateManifests(ctx); err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		log.Error(err, "error")
		os.Exit(1)
	}
}
