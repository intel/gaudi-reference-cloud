// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Deploy All In Kind deploys a kind cluster with IDC services.
// Execute with: make deploy-all-in-kind-v2 |& ts | ts -i | ts -s | tee local/deploy-all-in-kind-v2.log
//           or: make upgrade-all-in-kind-v2 |& ts | ts -i | ts -s | tee local/upgrade-all-in-kind-v2.log

package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/deployer"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/reader"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/k8s_provisioner/kind"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	flag "github.com/spf13/pflag"
)

type Arguments struct {
	ApplicationsToDeleteRegex   string
	BazelBinary                 string
	BazelBuildOpts              []string
	BazelStartupOpts            []string
	BuildArtifactsDir           string
	CacheDir                    string
	ClusterPrefix               string
	Commit                      string
	DeleteAllArgoApplications   bool
	DeleteArgoCd                bool
	DeleteGitea                 bool
	DeleteVault                 bool
	DockerImagePrefix           string
	HelmProject                 string
	HomeDir                     string
	IdcEnv                      string
	IdcServicesDeploymentMethod string
	IncludeDeploy               bool
	IncludeGenerateManifests    bool
	IncludePush                 bool
	IncludeVaultConfigure       bool
	IncludeVaultLoadSecrets     bool
	InitializeGitRepo           bool
	LocalRegistryName           string
	LocalRegistryPort           int
	RunfilesDir                 string
	SecretsDir                  string
	SemanticVersion             string
	TempDir                     string
	TestEnvironmentId           string
	UniverseConfig              string
	Upgrade                     bool
	WorkspaceDir                string
}

func parseArgs() Arguments {
	var args Arguments

	flag.StringVar(&args.ApplicationsToDeleteRegex, "applications-to-delete", "", "When upgrading, any Argo Applications with names matching this regex will be deleted")
	flag.StringVar(&args.BazelBinary, "bazel-binary", "bazel", "Path to bazel binary")
	flag.StringArrayVar(&args.BazelBuildOpts, "bazel-build-opt", nil, "Bazel build options")
	flag.StringArrayVar(&args.BazelStartupOpts, "bazel-startup-opt", nil, "Bazel startup options")
	flag.StringVar(&args.BuildArtifactsDir, "build-artifacts-dir", "", "If provided, generated build and deployment artifacts (including logs) will be stored here.")
	flag.StringVar(&args.CacheDir, "cache-dir", "", "If provided, downloaded and generated artifacts will be cached here")
	flag.StringVar(&args.ClusterPrefix, "cluster-prefix", "idc", "An ID used to name the kind cluster")
	flag.StringVar(&args.Commit, "commit", "", "Git commit hash")
	flag.BoolVar(&args.DeleteAllArgoApplications, "delete-all-argo-applications", false, "If true, delete all Argo CD Applications")
	flag.BoolVar(&args.DeleteArgoCd, "delete-argo-cd", false, "If true, delete Argo CD")
	flag.BoolVar(&args.DeleteGitea, "delete-gitea", false, "If true, delete Gitea")
	flag.BoolVar(&args.DeleteVault, "delete-vault", false, "If true, delete Vault")
	flag.StringVar(&args.DockerImagePrefix, "docker-image-prefix", "", "Container images will have this prefix")
	flag.StringVar(&args.HelmProject, "helm-project", "", "Helm project")
	flag.StringVar(&args.IdcEnv, "idc-env", "", "IDC environment")
	flag.StringVar(&args.IdcServicesDeploymentMethod, "idc-services-deployment-method",
		deployer.ServiceDeploymentMethodArgoCD, "must be 'argocd'")
	flag.BoolVar(&args.IncludeDeploy, "include-deploy", true, "To only uninstall, set delete flags to true and this flag to false.")
	flag.BoolVar(&args.IncludeGenerateManifests, "include-generate-manifests", true, "If true, generate manifests")
	flag.BoolVar(&args.IncludePush, "include-push", true, "If true, push containers and Helm charts")
	flag.BoolVar(&args.IncludeVaultConfigure, "include-vault-configure", false, "If true, configure Vault")
	flag.BoolVar(&args.IncludeVaultLoadSecrets, "include-vault-load-secrets", false, "If true, load secrets into Vault")
	flag.StringVar(&args.LocalRegistryName, "local-registry-name", "", "Local Docker registry host name")
	flag.IntVar(&args.LocalRegistryPort, "local-registry-port", 0, "Local Docker registry port number")
	flag.StringVar(&args.SecretsDir, "secrets-dir", "",
		"Path to the directory containing any secrets required to generate Argo CD manifests")
	flag.StringVar(&args.SemanticVersion, "semantic-version", "", "IDC semantic version in format 1.2.3")
	flag.StringVar(&args.TestEnvironmentId, "test-environment-id", "idc", "An ID used to identify this test environment in telemetry.")
	flag.StringVar(&args.TempDir, "temp-dir", "", "If provided, temporary files will be stored here and not deleted")
	flag.BoolVar(&args.Upgrade, "upgrade", false, "upgrade")
	flag.StringVar(&args.UniverseConfig, "universe-config", "", "Path to the Universe Config file")

	flag.Parse()

	args.HomeDir = os.Getenv("HOME")
	args.WorkspaceDir = util.WorkspaceDir()
	args.BuildArtifactsDir = util.AbsFromWorkspace(args.BuildArtifactsDir)
	args.CacheDir = util.AbsFromWorkspace(args.CacheDir)
	args.TempDir = util.AbsFromWorkspace(args.TempDir)
	args.UniverseConfig = util.AbsFromWorkspace(args.UniverseConfig)

	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	args.RunfilesDir = workingDir

	if args.IncludeDeploy {
		args.IncludeGenerateManifests = true
	}

	if !args.Upgrade {
		args.InitializeGitRepo = true
		args.IncludeVaultConfigure = true
		args.IncludeVaultLoadSecrets = true
	}

	if args.DeleteAllArgoApplications {
		args.ApplicationsToDeleteRegex = ".*"
	}

	util.EnsureRequiredStringFlag("bazel-binary")
	util.EnsureRequiredStringFlag("commit")
	util.EnsureRequiredStringFlag("idc-env")
	util.EnsureRequiredStringFlag("commit")
	util.EnsureRequiredStringFlag("local-registry-name")
	util.EnsureRequiredStringFlag("secrets-dir")
	util.EnsureRequiredStringFlag("semantic-version")
	util.EnsureRequiredStringFlag("test-environment-id")

	return args
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("deploy_all_in_kind"))
	log.Info("BEGIN")
	defer log.Info("END")

	err := func() error {
		var args = parseArgs()
		log.Info("args", "args", args)
		log.Info("Environment", "env", os.Environ())

		if args.IdcEnv == "prod" || args.IdcEnv == "staging" {
			return fmt.Errorf("deployment to environment '%s' is not allowed with deploy_all_in_kind", args.IdcEnv)
		}

		tempDir := args.TempDir
		if tempDir == "" {
			var err error
			tempDir, err = os.MkdirTemp("", "deploy_all_in_kind_")
			if err != nil {
				return err
			}
			defer os.RemoveAll(tempDir)
			log.Info("Using random temporary directory", "tempDir", tempDir)
		} else {
			log.Info("Using fixed temporary directory", "tempDir", tempDir)
			if err := os.RemoveAll(tempDir); err != nil {
				return err
			}
			if err := os.MkdirAll(tempDir, 0750); err != nil {
				return err
			}
		}

		cacheDir := args.CacheDir
		if cacheDir == "" {
			// By default, don't use the cache because the working directory may have uncommitted changes,
			// making the Git commit hash an unreliable indicator of the inputs.
			// It is effectively disabled by setting it to the empty temporary directory that was created above.
			cacheDir = filepath.Join(tempDir, "cache")
		}

		// Copy runfiles directory so that it doesn't get changed by "bazel run" commands executed by this application.
		if args.RunfilesDir != "" {
			oldRunfilesDir := args.RunfilesDir
			newRunfilesDir := filepath.Join(tempDir, "runfiles")
			log.Info("Copying runfiles", "oldRunfilesDir", oldRunfilesDir, "newRunFilesDir", newRunfilesDir)
			if err := util.CopyBazelRunfiles(ctx, oldRunfilesDir, newRunfilesDir); err != nil {
				return err
			}
			if err := os.Chdir(newRunfilesDir); err != nil {
				return err
			}
			args.RunfilesDir = newRunfilesDir
		}

		helmBinary := filepath.Join(args.RunfilesDir, filepaths.HelmBinary)
		helmfileBinary := filepath.Join(args.RunfilesDir, filepaths.HelmfileBinary)
		helmfileConfigDir := filepath.Join(args.WorkspaceDir, filepaths.HelmfileConfigDir)

		localRegistryName := args.LocalRegistryName
		if localRegistryName == "" {
			localRegistryName = os.Getenv("LOCAL_REGISTRY_NAME")
		}
		localRegistryPort := args.LocalRegistryPort
		if localRegistryPort == 0 {
			var err error
			localRegistryPort, err = strconv.Atoi(os.Getenv("LOCAL_REGISTRY_PORT"))
			if err != nil {
				localRegistryPort = 5001
			}
		}

		kubeConfig := filepath.Join(args.HomeDir, ".kube/config")

		// Load environment configuration from deployment artifacts.
		// This is used by Universe Deployer only.
		// It will not be used when generating manifests for components.
		envConfigReader := reader.Reader{
			HelmBinary:        helmBinary,
			HelmfileBinary:    helmfileBinary,
			HelmfileConfigDir: helmfileConfigDir,
		}
		multipleEnvConfig, err := envConfigReader.Read(ctx, args.IdcEnv)
		if err != nil {
			return err
		}
		envConfig := multipleEnvConfig.Environments[args.IdcEnv]

		if args.BuildArtifactsDir != "" {
			if err := os.MkdirAll(args.BuildArtifactsDir, 0750); err != nil {
				return err
			}
			// Write to BuildArtifactsDir to make it available in Jenkins.
			dest := filepath.Join(args.BuildArtifactsDir, "helmfile-dump.yaml")
			if err := os.WriteFile(dest, envConfig.HelmfileDumpYamlBytes, 0640); err != nil {
				return err
			}
		}

		universeConfig, err := universe_config.NewUniverseConfigFromFile(ctx, args.UniverseConfig)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
			// Universe Config json file does not exist.
			// Use Universe Config from envConfig.
			universeConfig, err = universe_config.NewUniverseConfigFromEnvConfig(ctx, envConfig.EnvConfig)
			if err != nil {
				return err
			}
		}

		if err := universeConfig.Normalize(ctx); err != nil {
			return err
		}
		// Resolve branches and tags to commit hashes.
		if envConfig.Values.UniverseDeployer.ComponentConfigCommitGitRemote != envConfig.Values.UniverseDeployer.ComponentCommitGitRemote {
			return fmt.Errorf("ComponentCommitGitRemote and ComponentConfigCommitGitRemote must have the same value")
		}
		if err := universeConfig.ResolveReferences(ctx, args.WorkspaceDir, envConfig.Values.UniverseDeployer.ComponentCommitGitRemote); err != nil {
			return err
		}

		// Replace HEAD commits to commit hash.
		if _, err = universeConfig.ReplaceCommits(ctx, map[string]string{util.HEAD: args.Commit}); err != nil {
			return err
		}
		if _, err = universeConfig.ReplaceConfigCommits(ctx, map[string]string{util.HEAD: args.Commit}); err != nil {
			return err
		}

		log.Info("Effective Universe Config", "universeConfig", universeConfig)

		// Ensure that all non-empty commits are in the form of a Git commit hash (40 hex digits).
		if err := universeConfig.ValidateCommits(ctx); err != nil {
			return err
		}

		kindProvisionerOpts := kind.KindProvisionerOptions{
			ClusterPrefix:     args.ClusterPrefix,
			EnvConfig:         *envConfig,
			KubeConfig:        kubeConfig,
			LocalRegistryName: localRegistryName,
			LocalRegistryPort: localRegistryPort,
			RunfilesDir:       args.RunfilesDir,
			SecretsDir:        args.SecretsDir,
			TempDir:           tempDir,
			TestEnvironmentId: args.TestEnvironmentId,
			Upgrade:           args.Upgrade,
			WorkspaceDir:      args.WorkspaceDir,
		}
		k8sProvisioner, err := kind.NewKindProvisioner(ctx, kindProvisionerOpts)
		if err != nil {
			return err
		}

		// helmRegistryInsideKind will be used by Argo CD to download charts.
		// It must be an HTTPS (TLS) server.
		// This points to the NGINX reverse proxy started by /deployment/registry/start_registry.sh.
		helmRegistryInsideKind := "nginx-local-registry"

		deployerOptions := deployer.DeployerOptions{
			ArgoApplicationsToDeleteRegex:    args.ApplicationsToDeleteRegex,
			ArgoApplicationsToNotDeleteRegex: "^app-of-apps$",
			BazelBinary:                      args.BazelBinary,
			BazelStartupOpts:                 args.BazelStartupOpts,
			BazelBuildOpts:                   args.BazelBuildOpts,
			BuildArtifactsDir:                args.BuildArtifactsDir,
			CacheDir:                         cacheDir,
			ClusterPrefix:                    args.ClusterPrefix,
			Commit:                           args.Commit,
			DeleteArgoCd:                     args.DeleteArgoCd,
			DeleteGitea:                      args.DeleteGitea,
			DeleteVault:                      args.DeleteVault,
			HelmDisableForceUpdate:           true,
			EnvConfig:                        *envConfig,
			HomeDir:                          args.HomeDir,
			IdcEnv:                           args.IdcEnv,
			IdcServicesDeploymentMethod:      deployer.ServiceDeploymentMethod(args.IdcServicesDeploymentMethod),
			IncludePush:                      args.IncludePush,
			InitializeGitRepo:                args.InitializeGitRepo,
			KubeConfig:                       kubeConfig,
			K8sProvisioner:                   k8sProvisioner,
			IncludeDeployK8sTlsSecrets:       false,
			IncludeVaultConfigure:            args.IncludeVaultConfigure,
			IncludeVaultLoadSecrets:          args.IncludeVaultLoadSecrets,
			OverrideDefaultChartRegistry:     helmRegistryInsideKind,
			SecretsDir:                       args.SecretsDir,
			SemanticVersion:                  args.SemanticVersion,
			TempDir:                          tempDir,
			TestEnvironmentId:                args.TestEnvironmentId,
			UniverseConfig:                   *universeConfig,
			WorkspaceDir:                     args.WorkspaceDir,
		}

		log.Info("Deployer options", "deployerOptions", fmt.Sprintf("%+v", deployerOptions))

		deployer, err := deployer.NewDeployer(ctx, deployerOptions)
		if err != nil {
			return err
		}
		defer deployer.TerminateBackgroundProcesses(ctx)

		if err := deployer.BuilderPusher(ctx); err != nil {
			return err
		}

		// Undeploy must come after pushing containers because it relies on some Helm charts.
		if err := deployer.Undeploy(ctx); err != nil {
			return err
		}

		if args.IncludeDeploy {
			if err := deployer.Deploy(ctx); err != nil {
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
