// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package pusher

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/deployer_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/types"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

type Pusher struct {
	Commit            string
	Components        []string
	SemanticVersion   string
	UniverseConfig    *universe_config.UniverseConfig
	WorkspaceDir      string
	SecretsDir        string
	BazelBinary       string
	BazelStartupOpts  []string
	BazelRunOpts      []string
	HelmBinary        string
	MultipleEnvConfig *types.MultipleEnvConfig
}

// Push container images and Helm charts to Docker and Helm registries.
func (p Pusher) Push(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Push")
	log.Info("BEGIN")
	defer log.Info("END")

	if len(p.UniverseConfig.Environments) == 0 {
		log.Info("No environments to process")
		return nil
	}

	if false {
		cmd := exec.CommandContext(ctx, "/bin/find", "-L", p.WorkspaceDir)
		cmd.Env = os.Environ()
		if err := util.RunCmd(ctx, cmd); err != nil {
			return err
		}
	}

	// Get deployer config from commit.
	deployerConfigFile := filepath.Join(p.WorkspaceDir, "deployment/universe_deployer/deployment_artifacts/config.yaml")
	deployerConfig, err := deployer_config.NewConfigFromFile(ctx, deployerConfigFile)
	if err != nil {
		return err
	}
	log.Info("deployerConfig", "deployerConfig", deployerConfig)

	for idcEnv := range p.UniverseConfig.Environments {
		log.Info("Processing environment", "idcEnv", idcEnv)

		var dockerRegistry string
		var dockerImagePrefix string
		var helmRegistry string
		var helmProject string

		envConfig, ok := p.MultipleEnvConfig.Environments[idcEnv]
		if !ok {
			return fmt.Errorf("environment %s not found in MultipleEnvConfig", idcEnv)
		}

		dockerRegistry = envConfig.Values.Image.Registry
		dockerImagePrefix = envConfig.Values.Image.RepositoryPrefix
		helmUrl, err := url.Parse("https://" + envConfig.Values.IdcHelmRepository.Url)
		if err != nil {
			return err
		}
		helmRegistry = helmUrl.Host
		helmProject = strings.TrimPrefix(helmUrl.Path, "/")
		log.Info("Using registry information from EnvConfig",
			"dockerRegistry", dockerRegistry,
			"dockerImagePrefix", dockerImagePrefix,
			"helmRegistry", helmRegistry,
			"helmProject", helmProject,
		)

		envSecretsDir := filepath.Join(p.SecretsDir, idcEnv)

		log.Info("Logging into Docker registry", "dockerRegistry", dockerRegistry)
		if err := dockerLogin(ctx, envSecretsDir, dockerRegistry); err != nil {
			return err
		}

		log.Info("Logging into Helm registry", "helmRegistry", helmRegistry)
		if err := helmLogin(ctx, envSecretsDir, helmRegistry, p.HelmBinary); err != nil {
			return err
		}

		if err := PushAllContainersAndChartsOnce(
			ctx,
			p.Commit,
			p.Components,
			p.SemanticVersion,
			dockerRegistry,
			dockerImagePrefix,
			helmRegistry,
			helmProject,
			p.WorkspaceDir,
			p.BazelBinary,
			p.BazelStartupOpts,
			p.BazelRunOpts,
		); err != nil {
			return err
		}
	}

	return nil
}

// Push all containers and charts once.
// Containers are pushed to dockerRegistry.
// Charts are pushed to helmRegistry.
func PushAllContainersAndChartsOnce(
	ctx context.Context,
	commit string,
	components []string,
	semanticVersion string,
	dockerRegistry string,
	dockerImagePrefix string,
	helmRegistry string,
	helmProject string,
	workspaceDir string,
	bazelBinary string,
	bazelStartupOpts []string,
	bazelRunOpts []string,
) error {
	log := log.FromContext(ctx).WithName("PushAllContainersAndChartsOnce")
	log.Info("BEGIN")
	defer log.Info("END")

	push_target := ""
	if len(components) == 0 {
		push_target = "//deployment/push:all_container_and_chart_push"
	} else if len(components) == 1 {
		component := components[0]
		push_target = "//deployment/push:" + component + "_component_container_and_chart_push"
	} else {
		return fmt.Errorf("more than one component is unsupported")
	}

	// Write .bzl files with registry info.
	dockerRegistryBzl := fmt.Sprintf("DOCKER_REGISTRY = \"%s\"\nDOCKER_IMAGE_PREFIX = \"%s\"\n", dockerRegistry, dockerImagePrefix)
	if err := os.WriteFile(filepath.Join(workspaceDir, "build/dynamic/docker_registry.bzl"), []byte(dockerRegistryBzl), 0640); err != nil {
		return err
	}
	helmRegistryBzl := fmt.Sprintf("HELM_REGISTRY = \"%s\"\nHELM_PROJECT = \"%s\"\n", helmRegistry, helmProject)
	if err := os.WriteFile(filepath.Join(workspaceDir, "build/dynamic/helm_registry.bzl"), []byte(helmRegistryBzl), 0640); err != nil {
		return err
	}

	args := []string{}
	args = append(args, bazelStartupOpts...)
	args = append(args, "run")
	args = append(args,
		"--verbose_failures",
	)
	args = append(args, bazelRunOpts...)
	args = append(args, push_target)
	cmd := exec.CommandContext(ctx, bazelBinary, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PATH=/bin")
	cmd.Dir = workspaceDir
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	return nil
}

// Docker login.
// This writes the credentials in plaintext to homeDir/.docker/config.json.
func dockerLogin(ctx context.Context, envSecretsDir string, dockerRegistry string) error {
	log := log.FromContext(ctx).WithName("dockerLogin")
	harborUsernameBytes, err := os.ReadFile(filepath.Join(envSecretsDir, "HARBOR_USERNAME"))
	if err != nil {
		log.Info("Unable to read file; assuming that login is not needed", "err", err)
		return nil
	}
	harborUsername := string(harborUsernameBytes)
	harborPasswordFile, err := os.Open(filepath.Join(envSecretsDir, "HARBOR_PASSWORD"))
	if err != nil {
		return err
	}
	defer harborPasswordFile.Close()

	cmd := exec.CommandContext(ctx, "/usr/bin/docker",
		"login",
		"--username", harborUsername,
		"--password-stdin",
		dockerRegistry,
	)
	cmd.Stdin = harborPasswordFile
	cmd.Env = os.Environ()
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}
	return nil
}

func helmLogin(ctx context.Context, envSecretsDir string, helmRegistry string, helmBinary string) error {
	log := log.FromContext(ctx).WithName("helmLogin")
	harborUsernameBytes, err := os.ReadFile(filepath.Join(envSecretsDir, "HARBOR_USERNAME"))
	if err != nil {
		log.Info("Unable to read file; assuming that login is not needed", "err", err)
		return nil
	}
	harborUsername := string(harborUsernameBytes)
	harborPasswordFile, err := os.Open(filepath.Join(envSecretsDir, "HARBOR_PASSWORD"))
	if err != nil {
		return err
	}
	defer harborPasswordFile.Close()

	cmd := exec.CommandContext(ctx, helmBinary,
		"registry",
		"login",
		"--username", harborUsername,
		"--password-stdin",
		helmRegistry,
	)
	cmd.Stdin = harborPasswordFile
	cmd.Env = os.Environ()
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}
	return nil
}
