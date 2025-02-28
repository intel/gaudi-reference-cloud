// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"os"
	"os/exec"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

// Initialize helmfile (install helm plugins).
func (e *Deployer) InitializeHelmfile(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, e.HelmfileBinary,
		"init",
		"--force",
		"--debug",
		"--helm-binary", e.HelmBinary,
	)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HOME="+e.HomeDir)
	return util.RunCmd(ctx, cmd)
}

func (e *Deployer) HelmfileEnv() []string {
	helmfileEnv := os.Environ()
	helmfileEnv = append(helmfileEnv, "CLUSTER_PREFIX="+e.ClusterPrefix)
	helmfileEnv = append(helmfileEnv, "DOCKER_REGISTRY="+e.DockerRegistry())
	helmfileEnv = append(helmfileEnv, "HELM_CHART_VERSIONS_DIR="+e.HelmChartVersionsDir)
	helmfileEnv = append(helmfileEnv, "HOME="+e.HomeDir)
	helmfileEnv = append(helmfileEnv, "KUBECONFIG="+e.KubeConfig)
	helmfileEnv = append(helmfileEnv, "SECRETS_DIR="+e.SecretsDir)
	helmfileEnv = append(helmfileEnv, "TEST_ENVIRONMENT_ID="+e.TestEnvironmentId)
	return helmfileEnv
}

func (e *Deployer) RunHelmfile(ctx context.Context, command string, args ...string) error {
	allArgs := append([]string{
		command,
		"--environment", e.IdcEnv,
		"--helm-binary", e.HelmBinary,
	}, args...)
	if e.HelmDisableForceUpdate {
		allArgs = append(allArgs, "--disable-force-update")
	}
	cmd := exec.CommandContext(ctx, e.HelmfileBinary, allArgs...)
	cmd.Dir = e.HelmfileConfigDir
	cmd.Env = e.HelmfileEnv()
	return util.RunCmd(ctx, cmd)
}
