// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package kindtestenv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/deployer"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/reader"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/k8s_provisioner/kind"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

type KindTestEnvOptions struct {
	// Any existing Kind clusters started with the same ClusterPrefix will be deleted when this environment starts.
	ClusterPrefix string
	// The helmfile environment (IDC_ENV) to deploy.
	IdcEnv                      string
	IdcServicesDeploymentMethod deployer.ServiceDeploymentMethod
	// If empty, use working dir.
	HomeDir           string
	LocalRegistryName string
	LocalRegistryPort int
	// Directory containing all run and test dependencies (except deployment artifacts). If empty, use working dir.
	// For example: /srv/claudiof/.cache/bazel/_bazel_claudiof/5899c1825136666a3ddb04834a591e09/execroot/com_intel_devcloud/bazel-out/k8-fastbuild/bin/go/pkg/universe_deployer/cmd/deploy_all_in_kind/deploy_all_in_kind_/deploy_all_in_kind.runfiles/com_intel_devcloud
	RunfilesDir string
	SecretsDir  string
	// This is a unique ID for this test environment.
	// Any existing Kind clusters started with the same TestEnvironmentId will be deleted when this environment starts.
	TestEnvironmentId string
	TempDir           string
	// Source directory containing the WORKSPACE.bazel file.
	// When run in a Bazel test sandbox, this should have the same value as RunfilesDir.
	WorkspaceDir string
}

// KindTestEnv deploys a kind cluster with IDC services.
// It is intended to be used by tests.
type KindTestEnv struct {
	ClusterPrefix     string
	Deployer          *deployer.Deployer
	HomeDir           string
	IdcEnv            string
	KubeConfig        string
	KubectlBinary     string
	K8sProvisioner    *kind.KindProvisioner
	LocalRegistryName string
	LocalRegistryPort int
	RunfilesDir       string
	SecretsDir        string
	SkipTearDown      bool
	TestEnvironmentId string
	// Source directory containing the WORKSPACE.bazel file.
	// When run in a Bazel test sandbox, this should have the same value as RunfilesDir.
	WorkspaceDir string
}

func NewKindTestEnv(ctx context.Context, opts KindTestEnvOptions) (*KindTestEnv, error) {
	log := log.FromContext(ctx).WithName("NewKindTestEnv")

	if opts.IdcEnv == "" {
		return nil, fmt.Errorf("required parameter IdcEnv is empty")
	}
	idcEnv := opts.IdcEnv

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if opts.TempDir == "" {
		return nil, fmt.Errorf("required parameter TempDir is empty")
	}

	runfilesDir := opts.RunfilesDir
	if runfilesDir == "" {
		runfilesDir = workingDir
	}
	deploymentArtifactsDir := runfilesDir

	homeDir := opts.HomeDir
	if homeDir == "" {
		homeDir = workingDir
	}

	workspaceDir := opts.WorkspaceDir
	if workspaceDir == "" {
		workspaceDir = runfilesDir
	}

	secretsDir := opts.SecretsDir
	if secretsDir == "" {
		secretsDir = filepath.Join(workspaceDir, "local/secrets/"+idcEnv)
	}

	buildArtifactsDir := filepath.Join(workspaceDir, "local/build-artifacts")

	log.Info("directories",
		"homeDir", homeDir,
		"runfilesDir", runfilesDir,
		"deploymentArtifactsDir", deploymentArtifactsDir,
		"secretsDir", secretsDir,
		"workspaceDir", workspaceDir)

	// Use a fixed Git commit hash because we don't want KindTestEnv to depend on the actual Git commit hash,
	// as this would invalidate the Bazel test cache even if no relevant files changed but the Git commit hash changed.
	// This test always builds and deploys the currently checked out code.
	// This commit hash is used as the application version in Helm releases.
	commit := "01cd8985e2113e31fc102129d51150c8cc115b5b"
	idcServicesDeploymentMethod := opts.IdcServicesDeploymentMethod
	if idcServicesDeploymentMethod == "" {
		idcServicesDeploymentMethod = deployer.ServiceDeploymentMethodArgoCD
	}
	kubeConfig := filepath.Join(homeDir, ".kube/config")
	skipAll := os.Getenv("SKIP_ALL") == "1" // TODO
	skipTearDown := os.Getenv("SKIP_TEAR_DOWN") == "1"
	helmBinary := filepath.Join(runfilesDir, filepaths.HelmBinary)
	helmfileBinary := filepath.Join(runfilesDir, filepaths.HelmfileBinary)
	helmfileConfigDir := filepath.Join(deploymentArtifactsDir, filepaths.HelmfileConfigDir)
	helmfileDumpYamlFile := filepath.Join(secretsDir, "helmfile-dump.yaml")
	kubectlBinary := filepath.Join(runfilesDir, filepaths.KubectlBinary)

	// Note that local registry will not be used when run in Jenkins.
	localRegistryName := opts.LocalRegistryName
	if localRegistryName == "" {
		localRegistryName = os.Getenv("LOCAL_REGISTRY_NAME")
	}
	localRegistryPort := opts.LocalRegistryPort
	if localRegistryPort == 0 {
		localRegistryPort, err = strconv.Atoi(os.Getenv("LOCAL_REGISTRY_PORT"))
		if err != nil {
			localRegistryPort = 5001
		}
	}

	testEnvironmentId := opts.TestEnvironmentId
	if testEnvironmentId == "" {
		if skipTearDown {
			// Use a deterministic ID.
			testEnvironmentId = opts.IdcEnv
		} else {
			// Use a random ID to allow concurrent tests on the same host.
			testEnvironmentId, err = util.GenerateRandomAlphaNumericString(6)
			if err != nil {
				return nil, err
			}
		}
	}
	log.Info("testEnvironmentId", "testEnvironmentId", testEnvironmentId)

	clusterPrefix := opts.ClusterPrefix
	if clusterPrefix == "" {
		clusterPrefix = testEnvironmentId
	}
	log.Info("clusterPrefix", "clusterPrefix", clusterPrefix)

	if skipAll {
		kubeConfig = "local/secrets/test-e2e-compute-vm/kubeconfig/config"
	}

	// Load environment configuration from deployment artifacts.
	// This is used by Universe Deployer only.
	// It will not be used when generating manifests for components.
	envConfigReader := reader.Reader{
		ClusterPrefix:     clusterPrefix,
		HelmBinary:        helmBinary,
		HelmfileBinary:    helmfileBinary,
		HelmfileConfigDir: helmfileConfigDir,
		TestEnvironmentId: testEnvironmentId,
	}
	multipleEnvConfig, err := envConfigReader.Read(ctx, idcEnv)
	if err != nil {
		return nil, err
	}
	envConfig := multipleEnvConfig.Environments[idcEnv]

	if err := os.WriteFile(helmfileDumpYamlFile, envConfig.HelmfileDumpYamlBytes, 0640); err != nil {
		return nil, err
	}

	universeConfig, err := universe_config.NewUniverseConfigFromEnvConfig(ctx, envConfig.EnvConfig)
	if err != nil {
		return nil, err
	}

	if err := universeConfig.Normalize(ctx); err != nil {
		return nil, err
	}

	// Replace HEAD commits to commit hash.
	if _, err = universeConfig.ReplaceCommits(ctx, map[string]string{util.HEAD: commit}); err != nil {
		return nil, err
	}
	if _, err = universeConfig.ReplaceConfigCommits(ctx, map[string]string{util.HEAD: commit}); err != nil {
		return nil, err
	}

	log.Info("Effective Universe Config", "universeConfig", universeConfig)

	// Ensure that all non-empty commits are in the form of a Git commit hash (40 hex digits).
	if err := universeConfig.ValidateCommits(ctx); err != nil {
		return nil, err
	}

	cacheDir := filepath.Join(opts.TempDir, "cache")
	headDeploymentArtifactsTar := filepath.Join(opts.TempDir, "deployment_artifacts.tar")

	// Create deployment artifacts tar file.
	cmd := exec.CommandContext(ctx, "/bin/tar",
		"-C", workspaceDir,
		"--sort=name",
		"--owner=root:0",
		"--group=root:0",
		"--mtime=@0",
		"-f", headDeploymentArtifactsTar,
		"-c",
		".",
	)
	cmd.Env = os.Environ()
	if err := util.RunCmd(ctx, cmd); err != nil {
		return nil, err
	}

	// chartRegistryInsideKind will be used by Argo CD to download charts.
	// In Jenkins, this should be empty to use Harbor.
	// If set, it must be an HTTPS (TLS) server.
	// If running in a development VM, this should be set to "nginx-local-registry"
	// to use the NGINX reverse proxy started by /deployment/registry/start_registry.sh.
	chartRegistryInsideKind := ""

	kindProvisionerOpts := kind.KindProvisionerOptions{
		ClusterPrefix:     clusterPrefix,
		EnvConfig:         *envConfig,
		KubeConfig:        kubeConfig,
		LocalRegistryName: localRegistryName,
		LocalRegistryPort: localRegistryPort,
		RunfilesDir:       runfilesDir,
		SecretsDir:        secretsDir,
		TestEnvironmentId: testEnvironmentId,
		TempDir:           opts.TempDir,
		WorkspaceDir:      workspaceDir,
	}
	k8sProvisioner, err := kind.NewKindProvisioner(ctx, kindProvisionerOpts)
	if err != nil {
		return nil, err
	}

	deployerOptions := deployer.DeployerOptions{
		BuildArtifactsDir:            buildArtifactsDir,
		CacheDir:                     cacheDir,
		ClusterPrefix:                clusterPrefix,
		Commit:                       commit,
		DeleteArgoCd:                 true,
		DeleteGitea:                  true,
		DeleteVault:                  true,
		EnvConfig:                    *envConfig,
		IdcEnv:                       idcEnv,
		IdcServicesDeploymentMethod:  idcServicesDeploymentMethod,
		IncludePush:                  false,
		HeadDeploymentArtifactsTar:   headDeploymentArtifactsTar,
		HomeDir:                      homeDir,
		IncludeDeployK8sTlsSecrets:   false,
		IncludeVaultConfigure:        true,
		IncludeVaultLoadSecrets:      true,
		K8sProvisioner:               k8sProvisioner,
		KubeConfig:                   kubeConfig,
		OverrideDefaultChartRegistry: chartRegistryInsideKind,
		RunfilesDir:                  runfilesDir,
		SecretsDir:                   secretsDir,
		SemanticVersion:              "0.0.0",
		TestEnvironmentId:            testEnvironmentId,
		TempDir:                      opts.TempDir,
		UniverseConfig:               *universeConfig,
		WorkspaceDir:                 workspaceDir,
	}
	deployer, err := deployer.NewDeployer(ctx, deployerOptions)
	if err != nil {
		return nil, err
	}

	testEnv := &KindTestEnv{
		ClusterPrefix:     clusterPrefix,
		Deployer:          deployer,
		HomeDir:           homeDir,
		IdcEnv:            idcEnv,
		KubeConfig:        kubeConfig,
		KubectlBinary:     kubectlBinary,
		K8sProvisioner:    k8sProvisioner,
		LocalRegistryName: localRegistryName,
		LocalRegistryPort: localRegistryPort,
		RunfilesDir:       runfilesDir,
		SecretsDir:        secretsDir,
		SkipTearDown:      skipTearDown,
		TestEnvironmentId: testEnvironmentId,
		WorkspaceDir:      workspaceDir,
	}
	log.Info("testEnv", "testEnv", fmt.Sprintf("%+v", testEnv))
	return testEnv, nil
}

func (e *KindTestEnv) RunExperiments(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "go/test/kindtestenv/experiments.sh"))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HELMFILE_ENVIRONMENT="+e.IdcEnv)
	cmd.Env = append(cmd.Env, "HOME="+e.HomeDir)
	cmd.Env = append(cmd.Env, "IDC_ENV="+e.IdcEnv)
	return util.RunCmd(ctx, cmd)
}

func (e *KindTestEnv) Start(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Start")
	log.Info("BEGIN")
	defer log.Info("END")

	if err := e.RunExperiments(ctx); err != nil {
		return err
	}

	if err := e.Deployer.Deploy(ctx); err != nil {
		return err
	}

	commonNames := []string{"client1"}
	for _, commonName := range commonNames {
		if err := e.Deployer.CreateVaultPkiCert(ctx, commonName); err != nil {
			return err
		}
	}

	return nil
}

func (e *KindTestEnv) Stop(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Stop")
	log.Info("BEGIN")
	defer log.Info("END")

	// Run before_stop.sh.
	cmd := exec.CommandContext(ctx, filepath.Join(e.RunfilesDir, "go/test/kindtestenv/before_stop.sh"))
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HOME="+e.HomeDir)
	cmd.Env = append(cmd.Env, "KUBECONFIG="+e.KubeConfig)
	cmd.Env = append(cmd.Env, "KUBECTL="+e.KubectlBinary)
	if err := util.RunCmd(ctx, cmd); err != nil {
		log.Error(err, "before_stop.sh failed")
		// continue trying to stop environment
	}

	if err := e.Deployer.TerminateBackgroundProcesses(ctx); err != nil {
		log.Error(err, "unable to terminate background processes")
		// continue trying to stop environment
	}

	if !e.SkipTearDown {
		if err := e.K8sProvisioner.Deprovision(ctx); err != nil {
			log.Error(err, "unable to undeploy kind")
			// continue trying to stop environment
		}
	}

	return nil
}

func (e *KindTestEnv) VaultAddr() string {
	return e.Deployer.VaultAddr
}

func (e *KindTestEnv) VaultToken() string {
	return e.Deployer.VaultToken
}

func (e *KindTestEnv) IngressHttpPort() int {
	return e.Deployer.K8sProvisioner.IngressHttpPort
}

func (e *KindTestEnv) IngressHttpsPort() int {
	return e.Deployer.K8sProvisioner.IngressHttpsPort
}
