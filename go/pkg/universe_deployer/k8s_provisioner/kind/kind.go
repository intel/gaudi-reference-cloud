// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package kind

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/types"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

type KindProvisionerOptions struct {
	// Any existing Kind clusters started with the same ClusterPrefix will be deleted when this environment starts.
	ClusterPrefix     string
	EnvConfig         types.EnvConfigWithUnparsed
	KubeConfig        string
	LocalRegistryName string
	LocalRegistryPort int
	// Directory containing all run and test dependencies (except deployment artifacts). If empty, use working dir.
	// For example: /srv/claudiof/.cache/bazel/_bazel_claudiof/5899c1825136666a3ddb04834a591e09/execroot/com_intel_devcloud/bazel-out/k8-fastbuild/bin/go/pkg/universe_deployer/cmd/deploy_all_in_kind/deploy_all_in_kind_/deploy_all_in_kind.runfiles/com_intel_devcloud
	RunfilesDir string
	SecretsDir  string
	// This is a unique ID for this test environment.
	TestEnvironmentId string
	TempDir           string
	Upgrade           bool
	// Source directory containing the WORKSPACE.bazel file.
	// When run in a Bazel test sandbox, this should have the same value as RunfilesDir.
	WorkspaceDir string
}

// KindProvisioner deploys a kind cluster with IDC services.
type KindProvisioner struct {
	ClusterPrefix        string
	HelmfileDumpYamlFile string
	IdcEnv               string
	IngressHttpPort      int
	IngressHttpsPort     int
	KindBinary           string
	KubeConfig           string
	KubectlBinary        string
	LocalRegistryName    string
	LocalRegistryPort    int
	PortFileDirectory    string
	RunfilesDir          string
	SecretsDir           string
	TestEnvironmentId    string
	TempDir              string
	Upgrade              bool
	YqBinary             string
}

func NewKindProvisioner(ctx context.Context, opts KindProvisionerOptions) (*KindProvisioner, error) {
	log := log.FromContext(ctx).WithName("NewKindProvisioner")

	if opts.ClusterPrefix == "" {
		return nil, fmt.Errorf("required parameter ClusterPrefix is empty")
	}
	if opts.EnvConfig.Environment.Name == "" {
		return nil, fmt.Errorf("required parameter opts.EnvConfig.Environment.Name is empty")
	}
	if opts.KubeConfig == "" {
		return nil, fmt.Errorf("required parameter opts.KubeConfig is empty")
	}
	if opts.RunfilesDir == "" {
		return nil, fmt.Errorf("required parameter RunfilesDir is empty")
	}
	if opts.SecretsDir == "" {
		return nil, fmt.Errorf("required parameter SecretsDir is empty")
	}
	if opts.TempDir == "" {
		return nil, fmt.Errorf("required parameter TempDir is empty")
	}
	if opts.TestEnvironmentId == "" {
		return nil, fmt.Errorf("required parameter TestEnvironmentId is empty")
	}
	if opts.WorkspaceDir == "" {
		return nil, fmt.Errorf("required parameter WorkspaceDir is empty")
	}

	idcEnv := opts.EnvConfig.Environment.Name

	tempDir, err := os.MkdirTemp(opts.TempDir, "KindProvisioner_")
	if err != nil {
		return nil, err
	}

	helmfileDumpYamlFile := filepath.Join(tempDir, "helmfile-dump.yaml")
	kindBinary := filepath.Join(opts.RunfilesDir, filepaths.KindBinary)
	kubectlBinary := filepath.Join(opts.RunfilesDir, filepaths.KubectlBinary)
	portFileDirectory := filepath.Join(opts.WorkspaceDir, "local")
	yqBinary := filepath.Join(opts.RunfilesDir, filepaths.YqBinary)

	if err := os.WriteFile(helmfileDumpYamlFile, opts.EnvConfig.HelmfileDumpYamlBytes, 0640); err != nil {
		return nil, err
	}

	provisioner := &KindProvisioner{
		ClusterPrefix:        opts.ClusterPrefix,
		HelmfileDumpYamlFile: helmfileDumpYamlFile,
		IdcEnv:               idcEnv,
		KindBinary:           kindBinary,
		KubeConfig:           opts.KubeConfig,
		KubectlBinary:        kubectlBinary,
		LocalRegistryName:    opts.LocalRegistryName,
		LocalRegistryPort:    opts.LocalRegistryPort,
		PortFileDirectory:    portFileDirectory,
		RunfilesDir:          opts.RunfilesDir,
		SecretsDir:           opts.SecretsDir,
		TempDir:              tempDir,
		TestEnvironmentId:    opts.TestEnvironmentId,
		Upgrade:              opts.Upgrade,
		YqBinary:             yqBinary,
	}
	log.Info("provisioner", "provisioner", provisioner)
	return provisioner, nil
}

// Create Kind cluster.
func (e *KindProvisioner) Provision(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("KindProvisioner.Provision")
	logger.Info("BEGIN")
	defer logger.Info("END")

	if e.Upgrade {
		logger.Info("Skipping provisioning because we are upgrading.")
		return nil
	}

	timeBegin := time.Now()
	defer func() { logger.Info("Total duration", "duration", time.Since(timeBegin)) }()

	cmd := exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "deployment/kind/deploy-kind.sh"))
	cmd.Env = e.deployKindEnv()
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}
	if err := e.readDynamicPorts(ctx); err != nil {
		return err
	}
	return nil
}

// Delete Kind cluster.
func (e *KindProvisioner) Deprovision(ctx context.Context) error {
	logger := log.FromContext(ctx).WithName("KindProvisioner.Deprovision")
	logger.Info("BEGIN")
	defer logger.Info("END")

	timeBegin := time.Now()
	defer func() { logger.Info("Total duration", "duration", time.Since(timeBegin)) }()

	cmd := exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "deployment/kind/deploy-kind.sh"))
	cmd.Env = e.deployKindEnv()
	cmd.Env = append(cmd.Env, "CREATE_ENABLED=false")
	return util.RunCmd(ctx, cmd)
}

func (e *KindProvisioner) deployKindEnv() []string {
	env := os.Environ()
	env = append(env, "CLUSTER_PREFIX="+e.ClusterPrefix)
	env = append(env, "HELMFILE_DUMP="+e.HelmfileDumpYamlFile)
	env = append(env, "KIND="+e.KindBinary)
	env = append(env, "KUBECONFIG="+e.KubeConfig)
	env = append(env, "KUBECTL="+e.KubectlBinary)
	env = append(env, "LOCAL_REGISTRY_NAME="+e.LocalRegistryName)
	env = append(env, fmt.Sprintf("LOCAL_REGISTRY_PORT=%d", e.LocalRegistryPort))
	env = append(env, "PORT_FILE_DIRECTORY="+e.PortFileDirectory)
	env = append(env, "SECRETS_DIR="+e.SecretsDir)
	env = append(env, "TEST_ENVIRONMENT_ID="+e.TestEnvironmentId)
	env = append(env, "YQ="+e.YqBinary)
	return env
}

func (e *KindProvisioner) readDynamicPorts(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("ReadDynamicPorts")

	ingressHttpPort, err := readDynamicPortFile(ctx, filepath.Join(e.PortFileDirectory, e.ClusterPrefix+"-global_host_port_80"))
	if err != nil {
		return err
	}
	e.IngressHttpPort = ingressHttpPort
	log.Info("Port", "IngressHttpPort", e.IngressHttpPort)

	ingressHttpsPort, err := readDynamicPortFile(ctx, filepath.Join(e.PortFileDirectory, e.ClusterPrefix+"-global_host_port_443"))
	if err != nil {
		return err
	}
	e.IngressHttpsPort = ingressHttpsPort
	log.Info("Port", "IngressHttpsPort", e.IngressHttpsPort)

	return nil
}

func readDynamicPortFile(_ context.Context, filename string) (int, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimRight(string(fileBytes), "\n"))
}
