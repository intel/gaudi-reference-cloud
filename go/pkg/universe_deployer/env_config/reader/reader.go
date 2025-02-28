// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package reader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/types"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	"gopkg.in/yaml.v3"
)

// Reader calculates and reads the configuration for IDC environments.
// This uses Helmfile to evaluate the configuration files in /deployment/helmfile/environment.
// This configuration is used for Helm releases (Argo Applications).
// It is also used for other things such as the location of the Docker registry (for pushing)
// and the list of regions (for configuring Vault).
type Reader struct {
	ClusterPrefix string
	// Optional
	DeploymentArtifactsDir string
	// Optional if DeploymentArtifactsDir is provided.
	HelmBinary string
	// Optional if DeploymentArtifactsDir is provided.
	HelmfileBinary string
	// Optional if DeploymentArtifactsDir is provided.
	HelmfileConfigDir string
	TestEnvironmentId string
}

// Read the EnvConfig for a single named IDC environment.
// This can be used when a Universe Config file should be generated based on EnvConfig.
func (r *Reader) Read(ctx context.Context, idcEnv string) (*types.MultipleEnvConfig, error) {
	if err := r.initialize(ctx); err != nil {
		return nil, err
	}
	envConfig, err := r.read(ctx, idcEnv)
	if err != nil {
		return nil, err
	}
	multipleEnvConfig := &types.MultipleEnvConfig{
		Environments: map[string]*types.EnvConfigWithUnparsed{
			idcEnv: envConfig},
	}
	return multipleEnvConfig, nil
}

// Read the EnvConfig for each IDC environment in the Universe Config.
func (r *Reader) ReadUniverse(ctx context.Context, universeConfig *universe_config.UniverseConfig) (*types.MultipleEnvConfig, error) {
	if err := r.initialize(ctx); err != nil {
		return nil, err
	}
	multipleEnvConfig := &types.MultipleEnvConfig{
		Environments: map[string]*types.EnvConfigWithUnparsed{},
	}
	for idcEnv := range universeConfig.Environments {
		envConfig, err := r.read(ctx, idcEnv)
		if err != nil {
			return nil, fmt.Errorf("environment %s: %w", idcEnv, err)
		}
		multipleEnvConfig.Environments[idcEnv] = envConfig
	}
	return multipleEnvConfig, nil
}

func (r *Reader) initialize(context.Context) error {
	if r.HelmBinary == "" {
		r.HelmBinary = filepath.Join(r.DeploymentArtifactsDir, filepaths.HelmBinary)
	}
	if r.HelmfileBinary == "" {
		r.HelmfileBinary = filepath.Join(r.DeploymentArtifactsDir, filepaths.HelmfileBinary)
	}
	if r.HelmfileConfigDir == "" {
		r.HelmfileConfigDir = filepath.Join(r.DeploymentArtifactsDir, filepaths.HelmfileConfigDir)
	}
	return nil
}

func (r *Reader) read(ctx context.Context, idcEnv string) (*types.EnvConfigWithUnparsed, error) {
	log := log.FromContext(ctx).WithName("Read")
	tempDir, err := os.MkdirTemp("", "universe_deployer_env_config_reader_")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)
	helmfileDumpYamlFile := filepath.Join(tempDir, "helmfile-dump.yaml")
	if err := r.createHelmfileDump(ctx, idcEnv, helmfileDumpYamlFile); err != nil {
		return nil, err
	}
	helmfileDumpYamlBytes, err := os.ReadFile(helmfileDumpYamlFile)
	if err != nil {
		return nil, err
	}
	envConfig := types.EnvConfig{}
	if err := yaml.Unmarshal(helmfileDumpYamlBytes, &envConfig); err != nil {
		return nil, err
	}
	log.Info("envConfig", "envConfig", envConfig)
	envConfigWithUnparsed := types.NewEnvConfigWithUnparsed(envConfig, helmfileDumpYamlBytes)
	return envConfigWithUnparsed, nil
}

func (e *Reader) createHelmfileDump(ctx context.Context, idcEnv string, helmfileDumpYamlFile string) error {
	env := os.Environ()
	env = append(env, "CLUSTER_PREFIX="+e.ClusterPrefix)
	env = append(env, "TEST_ENVIRONMENT_ID="+e.TestEnvironmentId)
	cmd := exec.CommandContext(ctx, e.HelmfileBinary,
		"write-values",
		"--helm-binary", e.HelmBinary,
		"--file", "helmfile-dump.yaml",
		"--environment", idcEnv,
		"--output-file-template", helmfileDumpYamlFile,
	)
	cmd.Dir = e.HelmfileConfigDir
	cmd.Env = env
	return util.RunCmd(ctx, cmd)
}
