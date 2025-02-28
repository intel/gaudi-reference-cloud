// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/artifactory"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/builder"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/types"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/k8s_provisioner/kind"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/manifests_generator"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/multi_version_manifests_generator"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/pusher"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	restclient "k8s.io/client-go/rest"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceDeploymentMethod string

const (
	ServiceDeploymentMethodArgoCD = "argocd"
)

type DeployerOptions struct {
	ArgoApplicationsToDeleteRegex    string
	ArgoApplicationsToNotDeleteRegex string
	ArtifactRepositoryUrl            url.URL
	BazelBinary                      string
	BazelBuildOpts                   []string
	BazelStartupOpts                 []string
	BuildArtifactsDir                string
	CacheDir                         string
	ClusterPrefix                    string
	Commit                           string
	DeleteArgoCd                     bool
	DeleteGitea                      bool
	DeleteVault                      bool
	EnvConfig                        types.EnvConfigWithUnparsed
	GitPusherDryRun                  bool
	// The helmfile environment (IDC_ENV) to deploy.
	IdcEnv                      string
	IdcServicesDeploymentMethod ServiceDeploymentMethod
	IncludePush                 bool
	InitializeGitRepo           bool
	// Optional. If not provided, it will be built.
	HeadDeploymentArtifactsTar string
	HelmDisableForceUpdate     bool
	// If empty, use working dir.
	HomeDir                      string
	IncludeDeployK8sTlsSecrets   bool
	IncludeVaultConfigure        bool
	IncludeVaultLoadSecrets      bool
	KubeConfig                   string
	OverrideDefaultChartRegistry string
	K8sProvisioner               *kind.KindProvisioner
	// Directory containing all run and test dependencies (except deployment artifacts). If empty, use working dir.
	// For example: /srv/claudiof/.cache/bazel/_bazel_claudiof/5899c1825136666a3ddb04834a591e09/execroot/com_intel_devcloud/bazel-out/k8-fastbuild/bin/go/pkg/universe_deployer/cmd/deploy_all_in_kind/deploy_all_in_kind_/deploy_all_in_kind.runfiles/com_intel_devcloud
	RunfilesDir     string
	SecretsDir      string
	SemanticVersion string
	// This is a unique ID for this test environment.
	TestEnvironmentId string
	TempDir           string
	UniverseConfig    universe_config.UniverseConfig
	// Source directory containing the WORKSPACE.bazel file.
	// When run in a Bazel test sandbox, this should have the same value as RunfilesDir.
	WorkspaceDir string
}

// Deployer deploys IDC services to a Kubernetes cluster.
//
// This currently works for single-cluster development environments.
// In the future, this can be extended to multi-cluster production environments.
//
// For development environments, this deploys:
//   - Vault
//   - Vault Server
//   - Vault Agent Injector
//   - Configure Vault
//   - Load secrets
//   - CoreDNS
//   - Gitea (for Argo CD)
//   - Argo CD
//   - IDC applications (including waiting for the deployment to complete)
type Deployer struct {
	BackgroundProcessesCancel         func()
	BackgroundProcessesContext        context.Context
	BackgroundProcessesWaitGroup      sync.WaitGroup
	CacheDir                          string
	ClusterPrefix                     string
	ContainerAndChartPusher           *pusher.Pusher
	DefaultChartRegistry              string
	DeployArgoCdEnabled               bool
	DeployVaultEnabled                bool
	DeploymentArtifactsBuilder        *builder.Builder
	GenerateManifestsComplete         bool
	GitPassword                       string
	HasHeadCommit                     bool
	HelmBinary                        string
	HelmChartVersionsDir              string
	HelmDisableForceUpdate            bool
	HelmfileBinary                    string
	HelmfileConfigDir                 string
	HelmfileDumpYamlFile              string
	HomeDir                           string
	IdcArgoCdInitialDataDir           string
	IdcArgoCdLocalRepoDir             string
	IdcEnv                            string
	IdcServicesDeploymentMethod       ServiceDeploymentMethod
	InitializeComplete                bool
	InitializeGitRepo                 bool
	InitializeK8sClientsComplete      bool
	JqBinary                          string
	JwkSourceDir                      string
	K8sApiEnabled                     bool
	K8sClients                        map[string]k8sclient.Client
	K8sProvisioner                    *kind.KindProvisioner
	KubeConfig                        string
	RestConfigs                       map[string]*restclient.Config
	KubectlBinary                     string
	Manifests                         *manifests_generator.Manifests
	ManifestsGenerator                *multi_version_manifests_generator.MultiVersionManifestsGenerator
	ManifestsGitBranch                string
	ManifestsGitRemote                string
	ManifestsGitRemoteWithCredentials string
	ManifestsTar                      string
	Options                           DeployerOptions
	PatchCommand                      string
	PortFileDirectory                 string
	RunfilesDir                       string
	SecretsDir                        string
	TestEnvironmentId                 string
	UpgradeBaseDeploymentData         string
	VaultAddr                         string
	VaultBinary                       string
	VaultToken                        string
	WaitForDeployments                []types.NamespacedName
	WaitForJobs                       []types.NamespacedName
	// Source directory containing the WORKSPACE.bazel file.
	// When run in a Bazel test sandbox, this should have the same value as RunfilesDir.
	WorkspaceDir string
	YqBinary     string
}

func NewDeployer(ctx context.Context, opts DeployerOptions) (*Deployer, error) {
	log := log.FromContext(ctx).WithName("NewDeployer")

	if opts.CacheDir == "" {
		return nil, fmt.Errorf("required parameter CacheDir is empty")
	}
	if opts.Commit == "" {
		return nil, fmt.Errorf("required parameter Commit is empty")
	}
	if opts.IdcEnv == "" {
		return nil, fmt.Errorf("required parameter IdcEnv is empty")
	}
	if opts.TempDir == "" {
		return nil, fmt.Errorf("required parameter TempDir is empty")
	}

	idcEnv := opts.IdcEnv

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	runfilesDir := opts.RunfilesDir
	if runfilesDir == "" {
		runfilesDir = workingDir
	}

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

	allEnvironmentsSecretsDir := filepath.Dir(opts.SecretsDir)

	log.Info("directories",
		"homeDir", homeDir,
		"runfilesDir", runfilesDir,
		"secretsDir", secretsDir,
		"workspaceDir", workspaceDir,
	)

	defaultChartRegistry := opts.OverrideDefaultChartRegistry
	if defaultChartRegistry == "" && opts.EnvConfig.Values.IdcHelmRepository.Url != "" {
		helmUrl, err := url.Parse("https://" + opts.EnvConfig.Values.IdcHelmRepository.Url)
		if err != nil {
			return nil, err
		}
		defaultChartRegistry = helmUrl.Host
	}

	deployArgoCdEnabled := opts.EnvConfig.Values.ArgoCd.Enabled
	deployVaultEnabled := opts.EnvConfig.Values.Vault.Enabled && opts.EnvConfig.Values.Global.Vault.Server.Enabled
	initializeGitRepo := opts.DeleteGitea || opts.InitializeGitRepo

	idcArgoCdInitialDataDir := filepath.Join(workspaceDir, filepaths.IdcArgoCdInitialData)
	idcArgoCdLocalRepoPath := filepath.Join(opts.TempDir, "idc-argocd-local-repo")
	idcServicesDeploymentMethod := opts.IdcServicesDeploymentMethod
	if idcServicesDeploymentMethod == "" {
		idcServicesDeploymentMethod = ServiceDeploymentMethodArgoCD
	}
	portFileDirectory := filepath.Join(workspaceDir, "local")
	helmBinary := filepath.Join(runfilesDir, filepaths.HelmBinary)
	helmChartVersionsDir := filepath.Join(runfilesDir, filepaths.HelmChartVersionsDir)
	helmfileBinary := filepath.Join(runfilesDir, filepaths.HelmfileBinary)
	helmfileConfigDir := filepath.Join(runfilesDir, filepaths.HelmfileConfigDir)
	jqBinary := filepath.Join(runfilesDir, filepaths.JqBinary)
	jwkSourceDir := filepath.Join(runfilesDir, filepaths.JwkSourceBaseDir, idcEnv, "vault-jwk-validation-public-keys")
	kubectlBinary := filepath.Join(runfilesDir, filepaths.KubectlBinary)
	vaultBinary := filepath.Join(runfilesDir, filepaths.VaultBinary)
	yqBinary := filepath.Join(runfilesDir, filepaths.YqBinary)

	patchCommand := ""
	if opts.EnvConfig.Values.UniverseDeployer.PatchCommand != "" {
		patchCommand = filepath.Join(workspaceDir, opts.EnvConfig.Values.UniverseDeployer.PatchCommand)
	}

	multipleEnvConfig, err := opts.EnvConfig.ToMultipleEnvConfig(ctx)
	if err != nil {
		return nil, err
	}

	helmfileDumpYamlFile := filepath.Join(opts.TempDir, "helmfile-dump.yaml")
	if err := os.WriteFile(helmfileDumpYamlFile, opts.EnvConfig.HelmfileDumpYamlBytes, 0640); err != nil {
		return nil, err
	}

	hasHeadCommit, err := opts.UniverseConfig.HasCommit(ctx, opts.Commit)
	if err != nil {
		return nil, err
	}

	headDeploymentArtifactsTar := opts.HeadDeploymentArtifactsTar
	requireHeadBuild := false
	requireHeadPush := false
	if hasHeadCommit && opts.HeadDeploymentArtifactsTar == "" {
		headDeploymentArtifactsTar = filepath.Join(opts.TempDir, "deployment_artifacts.tar")
		requireHeadBuild = true
		requireHeadPush = true
	}

	// Builder to build deployment artifacts for any components with commit "HEAD".
	var deploymentArtifactsBuilder *builder.Builder
	if requireHeadBuild {
		deploymentArtifactsBuilder = &builder.Builder{
			Commit:                    opts.Commit,
			SemanticVersion:           opts.SemanticVersion,
			HomeDir:                   opts.HomeDir,
			WorkspaceDir:              opts.WorkspaceDir,
			BazelBinary:               opts.BazelBinary,
			BazelStartupOpts:          opts.BazelStartupOpts,
			BazelBuildOpts:            opts.BazelBuildOpts,
			LegacyDefines:             false,
			LegacyDeploymentArtifacts: false,
			Output:                    headDeploymentArtifactsTar,
		}
	}

	var containerAndChartPusher *pusher.Pusher
	if requireHeadPush {
		containerAndChartPusher = &pusher.Pusher{
			Commit:            opts.Commit,
			SemanticVersion:   opts.SemanticVersion,
			UniverseConfig:    &opts.UniverseConfig,
			WorkspaceDir:      opts.WorkspaceDir,
			SecretsDir:        allEnvironmentsSecretsDir,
			BazelBinary:       opts.BazelBinary,
			BazelStartupOpts:  opts.BazelStartupOpts,
			BazelRunOpts:      opts.BazelBuildOpts,
			HelmBinary:        helmBinary,
			MultipleEnvConfig: &multipleEnvConfig,
		}
	}

	artifactoryObj, err := artifactory.NewFromSecretsDir(ctx, allEnvironmentsSecretsDir)
	if err != nil {
		return nil, err
	}

	manifestsTar := filepath.Join(opts.TempDir, "manifests.tar")

	manifestsGenerator := &multi_version_manifests_generator.MultiVersionManifestsGenerator{
		Artifactory:                artifactoryObj,
		ArtifactRepositoryUrl:      opts.ArtifactRepositoryUrl,
		CacheDir:                   opts.CacheDir,
		ClusterPrefix:              opts.ClusterPrefix,
		ConfigGitRemote:            opts.EnvConfig.Values.UniverseDeployer.ComponentConfigCommitGitRemote,
		ConfigGitRepositoryDir:     opts.WorkspaceDir,
		DefaultChartRegistry:       defaultChartRegistry,
		HeadCommit:                 opts.Commit,
		HeadDeploymentArtifactsTar: headDeploymentArtifactsTar,
		CombinedManifestsTar:       manifestsTar,
		SecretsDir:                 secretsDir,
		TestEnvironmentId:          opts.TestEnvironmentId,
		UniverseConfig:             &opts.UniverseConfig,
	}

	backgroundProcessContext, backgroundProcessCancel := context.WithCancel(ctx)

	deployer := &Deployer{
		BackgroundProcessesCancel:   backgroundProcessCancel,
		BackgroundProcessesContext:  backgroundProcessContext,
		CacheDir:                    opts.CacheDir,
		ClusterPrefix:               opts.ClusterPrefix,
		ContainerAndChartPusher:     containerAndChartPusher,
		DefaultChartRegistry:        defaultChartRegistry,
		DeployArgoCdEnabled:         deployArgoCdEnabled,
		DeployVaultEnabled:          deployVaultEnabled,
		DeploymentArtifactsBuilder:  deploymentArtifactsBuilder,
		HasHeadCommit:               hasHeadCommit,
		HelmBinary:                  helmBinary,
		HelmChartVersionsDir:        helmChartVersionsDir,
		HelmfileBinary:              helmfileBinary,
		HelmfileConfigDir:           helmfileConfigDir,
		HelmDisableForceUpdate:      opts.HelmDisableForceUpdate,
		HelmfileDumpYamlFile:        helmfileDumpYamlFile,
		JqBinary:                    jqBinary,
		JwkSourceDir:                jwkSourceDir,
		HomeDir:                     homeDir,
		IdcArgoCdInitialDataDir:     idcArgoCdInitialDataDir,
		IdcArgoCdLocalRepoDir:       idcArgoCdLocalRepoPath,
		IdcEnv:                      idcEnv,
		IdcServicesDeploymentMethod: idcServicesDeploymentMethod,
		InitializeGitRepo:           initializeGitRepo,
		K8sApiEnabled:               opts.EnvConfig.Values.UniverseDeployer.K8sApiEnabled,
		K8sProvisioner:              opts.K8sProvisioner,
		KubectlBinary:               kubectlBinary,
		KubeConfig:                  opts.KubeConfig,
		ManifestsGenerator:          manifestsGenerator,
		ManifestsGitBranch:          opts.EnvConfig.Values.UniverseDeployer.ManifestsGitBranch,
		ManifestsGitRemote:          opts.EnvConfig.Values.UniverseDeployer.ManifestsGitRemote,
		ManifestsTar:                manifestsTar,
		Options:                     opts,
		PatchCommand:                patchCommand,
		PortFileDirectory:           portFileDirectory,
		RunfilesDir:                 runfilesDir,
		SecretsDir:                  secretsDir,
		TestEnvironmentId:           opts.TestEnvironmentId,
		VaultBinary:                 vaultBinary,
		WorkspaceDir:                workspaceDir,
		YqBinary:                    yqBinary,
	}

	log.Info("deployer", "deployer", fmt.Sprintf("%+v", deployer))
	return deployer, nil
}

func (e *Deployer) DeployCoreDns(ctx context.Context) error {
	return e.RunHelmfile(
		ctx,
		"apply",
		"--allow-no-matching-release",
		"--file", "helmfile-coredns.yaml",
		"--selector", "chart=coredns",
	)
}

func (e *Deployer) DeployDebugTools(ctx context.Context) error {
	return e.RunHelmfile(
		ctx,
		"apply",
		"--allow-no-matching-release",
		"--file", "helmfile-debugTools.yaml",
		"--selector", "component=debugTools",
	)
}

func (e *Deployer) DockerRegistry() string {
	return os.Getenv("DOCKER_REGISTRY")
}

func (e *Deployer) Initialize(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Initialize")
	if e.InitializeComplete {
		return nil
	}
	log.Info("BEGIN")
	defer log.Info("END")

	if err := e.MakeSecrets(ctx); err != nil {
		return err
	}
	if err := e.ReadSecrets(ctx); err != nil {
		return err
	}
	if err := e.InitializeHelmfile(ctx); err != nil {
		return err
	}
	if err := e.SetManifestsGitRemoteWithCredentials(ctx); err != nil {
		return err
	}
	e.InitializeComplete = true
	return nil
}

func (e *Deployer) Undeploy(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Undeploy")
	log.Info("BEGIN")
	defer log.Info("END")

	if e.K8sApiEnabled {
		applicationNames, err := e.GetMatchingArgoApplicationNames(ctx, e.Options.ArgoApplicationsToDeleteRegex, e.Options.ArgoApplicationsToNotDeleteRegex)
		if err != nil {
			return err
		}
		log.Info("applicationsNames", "applicationsNames", applicationNames)

		if len(applicationNames) > 0 {
			// Delete selected Argo CD Applications from git repo.
			if err := e.DeleteArgoApplications(ctx, e.Options.ArgoApplicationsToDeleteRegex, e.Options.ArgoApplicationsToNotDeleteRegex); err != nil {
				log.Error(err, "unable to delete Argo Applications")
				// If Gitea is not running, this will fail. Continue to wait for Argo CD Applications to be deleted.
			}

			// Wait for selected Argo CD Applications to be deleted.
			if err := e.WaitForArgoApplicationsToBeDeleted(ctx, e.Options.ArgoApplicationsToDeleteRegex, e.Options.ArgoApplicationsToNotDeleteRegex); err != nil {
				return err
			}
		}

		if e.DeployArgoCdEnabled {
			if e.Options.DeleteArgoCd {
				if err := e.DeleteArgoCd(ctx); err != nil {
					return err
				}
			}
			if e.Options.DeleteGitea {
				if err := e.DeleteGitea(ctx); err != nil {
					return err
				}
				if err := e.DeleteLocalManifestsGitRepo(ctx); err != nil {
					return err
				}
			}
		}

		if e.DeployVaultEnabled {
			if e.Options.DeleteVault {
				if err := e.DeleteVault(ctx); err != nil {
					return err
				}
			}
		}

		// TODO: Delete namespaces
		// TODO: Delete CRDs
	}
	return nil
}

func (e *Deployer) Deploy(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("Deploy")
	log.Info("BEGIN")
	defer log.Info("END")

	if err := e.Initialize(ctx); err != nil {
		return err
	}

	if err := e.GenerateManifests(ctx); err != nil {
		return err
	}

	if e.Options.K8sProvisioner != nil {
		if err := e.Options.K8sProvisioner.Provision(ctx); err != nil {
			return err
		}
	}

	if e.K8sApiEnabled {
		if err := e.InitializeK8sClients(ctx); err != nil {
			return err
		}

		if e.Options.IncludeDeployK8sTlsSecrets {
			if err := e.DeployK8sTlsSecrets(ctx); err != nil {
				return err
			}
		}

		// Debug tools deployment is controlled by per-cluster Helmfile flags.
		if err := e.DeployDebugTools(ctx); err != nil {
			return err
		}

		// Core DNS deployment is controlled by per-cluster Helmfile flags.
		if err := e.DeployCoreDns(ctx); err != nil {
			return err
		}

		if e.DeployVaultEnabled {
			if err := e.DeployVault(ctx); err != nil {
				return err
			}
		}

		if e.DeployArgoCdEnabled {
			if err := e.DeployGitea(ctx); err != nil {
				return err
			}
			if e.InitializeGitRepo {
				if err := e.InitManifestsGitRepo(ctx); err != nil {
					return err
				}
			}
		}
	}

	if e.IdcServicesDeploymentMethod == ServiceDeploymentMethodArgoCD {
		if err := e.PushManifestsToGitRepo(ctx); err != nil {
			return err
		}
	}

	if e.K8sApiEnabled {
		if e.DeployArgoCdEnabled {
			if err := e.DeployArgoCd(ctx); err != nil {
				return err
			}
		}

		if err := e.WaitForK8sResources(ctx); err != nil {
			return err
		}

	}
	return nil
}

func (e *Deployer) TerminateBackgroundProcesses(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("TerminateBackgroundProcesses")
	log.V(2).Info("BEGIN")
	defer log.V(2).Info("END")
	e.BackgroundProcessesCancel()
	e.BackgroundProcessesWaitGroup.Wait()
	return nil
}
