// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Build and deploy all components to an existing Kubernetes cluster.
// Execute with: make deploy-all-in-k8s

package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/deployer"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/reader"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	flag "github.com/spf13/pflag"
)

type Arguments struct {
	BazelBinary                 string
	BazelBuildOpts              []string
	BazelStartupOpts            []string
	BuildArtifactsDir           string
	CacheDir                    string
	Commit                      string
	DeleteAllArgoApplications   bool
	DeleteArgoCd                bool
	DeleteGitea                 bool
	DeleteVault                 bool
	HomeDir                     string
	GitPusherDryRun             bool
	IdcEnv                      string
	IdcServicesDeploymentMethod string
	IncludeDeploy               bool
	IncludeGenerateManifests    bool
	IncludePush                 bool
	IncludeVaultConfigure       bool
	IncludeVaultLoadSecrets     bool
	RunfilesDir                 string
	SecretsDir                  string
	SemanticVersion             string
	TempDir                     string
	TestEnvironmentId           string
	UniverseConfig              string
	WorkspaceDir                string
}

func parseArgs() Arguments {
	var args Arguments

	flag.StringVar(&args.BazelBinary, "bazel-binary", "bazel", "Path to bazel binary")
	flag.StringArrayVar(&args.BazelBuildOpts, "bazel-build-opt", nil, "Bazel build options")
	flag.StringArrayVar(&args.BazelStartupOpts, "bazel-startup-opt", nil, "Bazel startup options")
	flag.StringVar(&args.BuildArtifactsDir, "build-artifacts-dir", "", "If provided, generated build and deployment artifacts (including logs) will be stored here.")
	flag.StringVar(&args.CacheDir, "cache-dir", "", "If provided, downloaded and generated artifacts will be cached here")
	flag.StringVar(&args.Commit, "commit", "", "Git commit hash")
	flag.BoolVar(&args.DeleteAllArgoApplications, "delete-all-argo-applications", false, "If true, delete all Argo CD Applications")
	flag.BoolVar(&args.DeleteArgoCd, "delete-argo-cd", false, "If true, delete Argo CD")
	flag.BoolVar(&args.DeleteGitea, "delete-gitea", false, "If true, delete Gitea")
	flag.BoolVar(&args.DeleteVault, "delete-vault", false, "If true, delete Vault")
	flag.BoolVar(&args.GitPusherDryRun, "git-pusher-dry-run", false, "If true, do not push changes to the manifests Git remote")
	flag.StringVar(&args.IdcEnv, "idc-env", "", "IDC environment")
	flag.StringVar(&args.IdcServicesDeploymentMethod, "idc-services-deployment-method",
		deployer.ServiceDeploymentMethodArgoCD, "must be 'argocd'")
	flag.BoolVar(&args.IncludeDeploy, "include-deploy", true, "To only uninstall, set delete flags to true and this flag to false.")
	flag.BoolVar(&args.IncludeGenerateManifests, "include-generate-manifests", true, "If true, generate manifests")
	flag.BoolVar(&args.IncludePush, "include-push", true, "If true, push containers and Helm charts")
	flag.BoolVar(&args.IncludeVaultConfigure, "include-vault-configure", true, "If true, configure Vault")
	flag.BoolVar(&args.IncludeVaultLoadSecrets, "include-vault-load-secrets", true, "If true, load secrets into Vault")
	flag.StringVar(&args.SecretsDir, "secrets-dir", "",
		"Path to the directory containing any secrets required to generate Argo CD manifests")
	flag.StringVar(&args.SemanticVersion, "semantic-version", "", "IDC semantic version in format 1.2.3")
	flag.StringVar(&args.TestEnvironmentId, "test-environment-id", "idc", "An ID used to name the kind cluster")
	flag.StringVar(&args.TempDir, "temp-dir", "", "If provided, temporary files will be stored here and not deleted")
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

	if args.IncludeDeploy {
		args.IncludeGenerateManifests = true
	}

	util.EnsureRequiredStringFlag("bazel-binary")
	util.EnsureRequiredStringFlag("commit")
	util.EnsureRequiredStringFlag("idc-env")
	util.EnsureRequiredStringFlag("commit")
	util.EnsureRequiredStringFlag("secrets-dir")
	util.EnsureRequiredStringFlag("semantic-version")
	util.EnsureRequiredStringFlag("test-environment-id")

	return args
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("deploy_all_in_k8s"))
	log.Info("BEGIN")
	defer log.Info("END")

	err := func() error {
		var args = parseArgs()
		log.Info("args", "args", args)
		log.Info("Environment", "env", os.Environ())

		if args.IdcEnv == "prod" || args.IdcEnv == "staging" {
			return fmt.Errorf("deployment to environment '%s' is not allowed with deploy_all_in_k8s", args.IdcEnv)
		}

		if args.DeleteArgoCd && !args.DeleteAllArgoApplications {
			// Prevent this scenario because any existing Argo Applications will get stuck in the deleting phase.
			return fmt.Errorf("when choosing to delete Argo CD (DELETE_ARGO_CD), you must also choose to delete all Argo Applications (DELETE_ALL_ARGO_APPLICATIONS)")
		}

		tempDir := args.TempDir
		if tempDir == "" {
			var err error
			tempDir, err = os.MkdirTemp("", "deploy_all_in_k8s_")
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

		argoApplicationsToDeleteRegex := ""
		if args.DeleteAllArgoApplications {
			argoApplicationsToDeleteRegex = ".*"
		}

		artifactRepositoryUrl, err := url.Parse(envConfig.Values.UniverseDeployer.ArtifactRepositoryUrl)
		if err != nil {
			return err
		}

		deployerOptions := deployer.DeployerOptions{
			ArgoApplicationsToDeleteRegex:    argoApplicationsToDeleteRegex,
			ArgoApplicationsToNotDeleteRegex: "^app-of-apps$",
			ArtifactRepositoryUrl:            *artifactRepositoryUrl,
			BazelBinary:                      args.BazelBinary,
			BazelStartupOpts:                 args.BazelStartupOpts,
			BazelBuildOpts:                   args.BazelBuildOpts,
			BuildArtifactsDir:                args.BuildArtifactsDir,
			CacheDir:                         args.CacheDir,
			Commit:                           args.Commit,
			DeleteArgoCd:                     args.DeleteArgoCd,
			DeleteGitea:                      args.DeleteGitea,
			DeleteVault:                      args.DeleteVault,
			GitPusherDryRun:                  args.GitPusherDryRun,
			HelmDisableForceUpdate:           true,
			EnvConfig:                        *envConfig,
			HomeDir:                          args.HomeDir,
			IdcEnv:                           args.IdcEnv,
			IdcServicesDeploymentMethod:      deployer.ServiceDeploymentMethod(args.IdcServicesDeploymentMethod),
			IncludePush:                      args.IncludePush,
			KubeConfig:                       os.Getenv("KUBECONFIG"),
			IncludeDeployK8sTlsSecrets:       envConfig.Values.UniverseDeployer.IncludeDeployK8sTlsSecrets,
			IncludeVaultConfigure:            args.IncludeVaultConfigure && envConfig.Values.UniverseDeployer.IncludeVaultConfigure,
			IncludeVaultLoadSecrets:          args.IncludeVaultLoadSecrets && envConfig.Values.UniverseDeployer.IncludeVaultLoadSecrets,
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

		if err := deployer.Initialize(ctx); err != nil {
			return err
		}

		if err := deployer.InitializeK8sClients(ctx); err != nil {
			return err
		}

		// Generate Argo CD manifests. Store in a tar file.
		if args.IncludeGenerateManifests {
			if err := deployer.GenerateManifests(ctx); err != nil {
				return fmt.Errorf("generating manifests: %w", err)
			}
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
