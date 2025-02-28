// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package multi_version_manifests_generator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/store_forward_logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/artifactory"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/cache"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/manifests_generator"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

// MultiVersionManifestsGenerator generates Argo CD manifests for all components in a Universe Config.
// It supports components with any Git hash for commit and configCommit.
type MultiVersionManifestsGenerator struct {
	// Deployment artifacts for non-HEAD commits will be downloaded from Artifactory.
	Artifactory           *artifactory.Artifactory
	ArtifactRepositoryUrl url.URL
	// Downloaded and generated files will be cached in this directory.
	CacheDir      string
	ClusterPrefix string
	// Config commits will be retrieved from this Git remote.
	ConfigGitRemote string
	// Config commits will be retrieved using this local Git repository.
	ConfigGitRepositoryDir string
	DefaultChartRegistry   string
	// If provided, a commit or configCommit set to HeadCommit will use HeadDeploymentArtifactsTar.
	HeadCommit string
	// If provided, deployment artifacts for HeadCommit will come from this tar file.
	HeadDeploymentArtifactsTar string
	// Generated combined manifests will be written to this tar file.
	CombinedManifestsTar string
	SecretsDir           string
	TestEnvironmentId    string
	UniverseConfig       *universe_config.UniverseConfig

	cache            *cache.Cache
	componentCommits []universe_config.ComponentCommit
	tempDir          string
}

type DownloadDeploymentArtifactsResult struct {
	// The list of deployment artifacts that were not found in the artifact repository.
	NotFoundErrors []ComponentCommitError
}

type ComponentCommitError struct {
	Component string
	Commit    string
	Error     error
}

// Try to download deployment artifacts for all components and commits referenced in the Universe Config from Artifactory.
// The list of artifacts which were not found will be returned in DownloadDeploymentArtifactsResult.
func (m MultiVersionManifestsGenerator) TryDownloadDeploymentArtifacts(ctx context.Context) (DownloadDeploymentArtifactsResult, error) {
	ctx, logger := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("TryDownloadDeploymentArtifacts"))
	logger.Info("BEGIN")
	defer logger.Info("END")

	timeBegin := time.Now()
	defer func() { logger.Info("Total duration", "duration", time.Since(timeBegin)) }()

	result := DownloadDeploymentArtifactsResult{}

	if m.CacheDir == "" {
		return result, fmt.Errorf("CacheDir is required")
	}

	logger.Info("Configuration",
		"ArtifactRepositoryUrl", m.ArtifactRepositoryUrl,
		"CacheDir", m.CacheDir,
		"HeadCommit", m.HeadCommit,
	)

	// For each (component, commit) tuple, download deployment artifacts.
	componentCommits, err := m.UniverseConfig.ComponentCommits(ctx, universe_config.ComponentCommitsModeIncludeComponentCommit)
	if err != nil {
		return result, err
	}

	for _, cc := range componentCommits {
		if cc.Commit != m.HeadCommit {
			_, err := m.downloadArtifactsForCommit(ctx, cc.Component, cc.Commit)
			if err != nil {
				logger.Error(err, "Error downloading")
				if !errors.Is(err, &artifactory.ErrNotFound{}) {
					return result, err
				}
				componentCommitError := ComponentCommitError{
					Component: cc.Component,
					Commit:    cc.Commit,
					Error:     err,
				}
				result.NotFoundErrors = append(result.NotFoundErrors, componentCommitError)
			}
		}
	}

	return result, nil
}

// Generate Argo CD manifests. Store in a tar file.
func (m MultiVersionManifestsGenerator) GenerateManifests(ctx context.Context) (*manifests_generator.Manifests, error) {
	ctx, logger := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("MultiVersionManifestsGenerator.GenerateManifests"))
	logger.Info("BEGIN")
	defer logger.Info("END")

	timeBegin := time.Now()
	defer func() { logger.Info("Total duration", "duration", time.Since(timeBegin)) }()

	tempDir, err := os.MkdirTemp("", "universe_deployer_multi_version_manifests_generator_")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)
	m.tempDir = tempDir

	if m.CacheDir == "" {
		m.CacheDir = tempDir
	}

	logger.Info("Configuration",
		"ArtifactRepositoryUrl", m.ArtifactRepositoryUrl,
		"CacheDir", m.CacheDir,
		"DefaultChartRegistry", m.DefaultChartRegistry,
		"HeadDeploymentArtifactsTar", m.HeadDeploymentArtifactsTar,
		"HeadCommit", m.HeadCommit,
		"ManifestsTar", m.CombinedManifestsTar,
		"SecretsDir", m.SecretsDir,
		"tempDir", m.tempDir,
	)
	logger.Info("Configuration", "universeConfig", m.UniverseConfig)

	m.cache, err = cache.New(ctx, filepath.Join(m.CacheDir, "multi_version_manifests_generator_v1"))
	if err != nil {
		return nil, err
	}

	if err := m.UniverseConfig.Normalize(ctx); err != nil {
		return nil, err
	}

	hasHead, err := m.UniverseConfig.HasCommit(ctx, m.HeadCommit)
	if err != nil {
		return nil, err
	}
	if hasHead {
		if m.HeadDeploymentArtifactsTar == "" {
			return nil, fmt.Errorf("commit 'HEAD' specified but HeadDeploymentArtifactsTar '%s' is invalid", m.HeadDeploymentArtifactsTar)
		}
	}

	if err := m.UniverseConfig.ValidateCommits(ctx); err != nil {
		return nil, err
	}

	// Get distinct (component, commit, configCommit) tuples.
	componentCommits, err := m.UniverseConfig.ComponentCommits(ctx, universe_config.ComponentCommitsModeIncludeAll)
	if err != nil {
		return nil, err
	}
	m.componentCommits = componentCommits
	logger.Info("Configuration", "componentCommits", m.componentCommits)

	commitManifestsChan := make(chan manifests_generator.Manifests)

	// To improve the speed, each ComponentCommit will be processed in a separate Go routine.
	// Use a StoreForwardErrGroup so that logs for each Go routine are kept together (not interleaved).
	g := store_forward_logger.NewStoreForwardErrGroup(ctx)

	// For each ComponentCommit, download deployment artifacts and generate manifests.
	for _, componentCommit := range m.componentCommits {
		component := componentCommit.Component
		commit := componentCommit.Commit
		configCommit := componentCommit.ConfigCommit
		desc := fmt.Sprintf("multi-version manifests generator, component %s, commit %s, configCommit %s", component, commit, configCommit)
		g.Go(desc, func(ctx context.Context) error {
			ctx = log.IntoContext(ctx, log.FromContext(ctx).
				WithValues("component", component, "commit", commit, "configCommit", configCommit))
			if err := m.generateManifestsForCommitOrHead(ctx, component, commit, configCommit, commitManifestsChan); err != nil {
				return err
			}
			return nil
		})
	}

	combinedManifestsChan := make(chan manifests_generator.Manifests)

	// Go routine to combine the manifests for each ComponentCommit.
	go func() {
		combinedManifests := manifests_generator.Manifests{}
		for commitManifests := range commitManifestsChan {
			combinedManifests.Manifests = append(combinedManifests.Manifests, commitManifests.Manifests...)
		}
		combinedManifests.Sort()
		combinedManifestsChan <- combinedManifests
		close(combinedManifestsChan)
	}()

	if err := g.Wait(); err != nil {
		return nil, err
	}
	close(commitManifestsChan)
	manifests := <-combinedManifestsChan

	if err := m.combineManifestsTars(ctx); err != nil {
		return nil, fmt.Errorf("combining manifests tars: %w", err)
	}

	logger.Info("Manifests generated", "manifestsTar", m.CombinedManifestsTar, "count", len(manifests.Manifests))
	for _, manifest := range manifests.Manifests {
		logger.Info("Manifest", "manifest", manifest)
	}

	return &manifests, nil
}

func (m MultiVersionManifestsGenerator) generateManifestsForCommitOrHead(
	ctx context.Context, component string, commit string, configCommit string, manifestsChan chan<- manifests_generator.Manifests) error {

	logger := log.FromContext(ctx).WithName("MultiVersionManifestsGenerator.generateManifestsForCommitOrHead")

	timeBegin := time.Now()
	defer func() { logger.Info("Total duration", "duration", time.Since(timeBegin)) }()

	// Trim Universe Config by component, commit, and configCommit.
	pred := func(comp string, universeComponent universe_config.UniverseComponent) bool {
		return comp == component && universeComponent.Commit == commit && universeComponent.ConfigCommit == configCommit
	}
	universeConfig := m.UniverseConfig.DeepCopy().Filtered(ctx, pred)
	logger.Info("universeConfig", "universeConfig", universeConfig)
	universeConfigFileName := filepath.Join(m.tempDir, fmt.Sprintf("universe_config_%s_%s_%s.json", component, commit, configCommit))
	if err := universeConfig.WriteFile(ctx, universeConfigFileName); err != nil {
		return err
	}

	// Create a cache key that hashes all inputs of generateManifestsForCommit.
	hasher := m.cache.NewHasher(ctx)
	if err := hasher.AddString(ctx, "commit", commit); err != nil {
		return err
	}
	if err := hasher.AddString(ctx, "component", component); err != nil {
		return err
	}
	if err := hasher.AddString(ctx, "configCommit", configCommit); err != nil {
		return err
	}
	if err := hasher.AddString(ctx, "DefaultChartRegistry", m.DefaultChartRegistry); err != nil {
		return err
	}
	if err := hasher.AddFile(ctx, "universeConfigFile", universeConfigFileName); err != nil {
		return err
	}
	cacheKey := fmt.Sprintf("manifests_%s.tar", hasher.Sum(ctx))

	cached, err := m.cache.IsCached(ctx, cacheKey)
	if err != nil {
		return err
	}

	if cached {
		logger.Info("Found in cache", "cacheKey", cacheKey)
	} else {
		logger.Info("Not found in cache", "cacheKey", cacheKey)

		// Download deployment artifacts tar or use from HEAD.
		var deploymentArtifactsTar string
		if commit == m.HeadCommit {
			deploymentArtifactsTar = m.HeadDeploymentArtifactsTar
		} else {
			var err error
			deploymentArtifactsTar, err = m.downloadArtifactsForCommit(ctx, component, commit)
			if err != nil {
				return fmt.Errorf("downloading artifacts: %w", err)
			}
		}

		deploymentArtifactsDir, err := m.extractDeploymentArtifacts(ctx, component, commit, configCommit, deploymentArtifactsTar)
		if err != nil {
			return fmt.Errorf("extracting deployment artifacts: %w", err)
		}

		if configCommit != "" && configCommit != commit {
			if err := m.replaceConfig(ctx, component, commit, configCommit, deploymentArtifactsDir); err != nil {
				return fmt.Errorf("replacing configuration: %w", err)
			}
		}

		tempManifestsTarFileName, err := m.cache.GetTempFilePath(ctx, cacheKey)
		if err != nil {
			return err
		}

		if err := m.generateManifestsForCommit(ctx, component, commit, configCommit, universeConfigFileName, deploymentArtifactsDir, tempManifestsTarFileName); err != nil {
			return fmt.Errorf("generating manifests: %w", err)
		}

		_, err = m.cache.MoveFileToCache(ctx, cacheKey, tempManifestsTarFileName)
		if err != nil {
			return err
		}
	}

	fileResult, err := m.cache.GetFile(ctx, cacheKey)
	if err != nil {
		return err
	}
	manifestsTarFileName := m.manifestsTarFileName(component, commit, configCommit)
	if err := util.CopyFile(fileResult.Path, manifestsTarFileName); err != nil {
		return err
	}

	// Determine Manifests from manifests tar.
	manifests, err := m.manifestsTarToManifests(ctx, component, commit, configCommit)
	if err != nil {
		return fmt.Errorf("manifestsTarToManifests: %w", err)
	}
	logger.Info("manifests", "count", len(manifests.Manifests))
	logger.Info("manifests", "manifests", manifests)
	manifestsChan <- *manifests

	return nil
}

// Download deployment artifacts tar file from Artifactory.
// Use cached file if previously downloaded.
func (m MultiVersionManifestsGenerator) downloadArtifactsForCommit(ctx context.Context, component string, commit string) (string, error) {
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("downloadArtifactsForCommit"))
	log.Info("BEGIN")
	defer log.Info("END")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	cacheKey := fmt.Sprintf("universe_deployer_deployment_artifacts_%s_%s.tar", component, commit)

	cached, err := m.cache.IsCached(ctx, cacheKey)
	if err != nil {
		return "", err
	}

	if cached {
		log.Info("Found in cache", "cacheKey", cacheKey)
	} else {
		log.Info("Not found in cache", "cacheKey", cacheKey)
		if m.ArtifactRepositoryUrl.Host == "" {
			return "", fmt.Errorf("deployment artifacts for component %s, commit %s must be downloaded but ArtifactRepositoryUrl is empty",
				component, commit)
		}
		artifactUrl, err := util.DeploymentArtifactsTarUrl(m.ArtifactRepositoryUrl, component, commit)
		if err != nil {
			return "", err
		}
		tempFileName, err := m.cache.GetTempFilePath(ctx, cacheKey)
		if err != nil {
			return "", err
		}
		if err := m.Artifactory.Download(ctx, artifactUrl, tempFileName); err != nil {
			return "", fmt.Errorf("unable to get deployment artifacts for this commit: %w", err)
		}
		_, err = m.cache.MoveFileToCache(ctx, cacheKey, tempFileName)
		if err != nil {
			return "", err
		}
	}

	fileResult, err := m.cache.GetFile(ctx, cacheKey)
	if err != nil {
		return "", err
	}

	return fileResult.Path, nil
}

func (m MultiVersionManifestsGenerator) extractDeploymentArtifacts(ctx context.Context, component string, commit string, configCommit string, deploymentArtifactsTar string) (string, error) {
	logger := log.FromContext(ctx).WithName("extractDeploymentArtifacts")

	timeBegin := time.Now()
	defer func() { logger.Info("Total duration", "duration", time.Since(timeBegin)) }()

	deploymentArtifactsDir := filepath.Join(m.tempDir, component, commit, configCommit, "deployment_artifacts")
	if err := os.RemoveAll(deploymentArtifactsDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(deploymentArtifactsDir, 0750); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "/bin/tar",
		"-C", deploymentArtifactsDir,
		"-x",
		"-f", deploymentArtifactsTar,
	)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return "", err
	}

	return deploymentArtifactsDir, nil
}

// If needed, replace the contents of environment-specific configuration directories with the same directories from configCommit.
func (m MultiVersionManifestsGenerator) replaceConfig(ctx context.Context, component string, commit string, configCommit string, deploymentArtifactsDir string) error {
	log := log.FromContext(ctx).WithName("replaceConfig")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	// Download deployment artifacts tar or use from HEAD.
	var configTar string
	if configCommit == m.HeadCommit {
		configTar = m.HeadDeploymentArtifactsTar
	} else {
		var err error
		configTar, err = m.getConfigForCommit(ctx, component, configCommit)
		if err != nil {
			return fmt.Errorf("downloading artifacts: %w", err)
		}
	}

	configDir, err := m.extractConfig(ctx, component, commit, configCommit, configTar)
	if err != nil {
		return fmt.Errorf("extracting config: %w", err)
	}

	for _, subdir := range filepaths.ConfigDirs() {
		source := filepath.Join(configDir, subdir)
		dest := filepath.Join(deploymentArtifactsDir, subdir)
		log.Info("Replacing configuration directory", "source", source, "dest", dest)
		if err := util.ChmodAll(dest, 0700); err != nil {
			return err
		}
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
		if err := util.CopyDir(source, dest); err != nil {
			return err
		}
	}

	return nil
}

func (m MultiVersionManifestsGenerator) extractConfig(ctx context.Context, component string, commit string, configCommit string, envConfigTar string) (string, error) {
	envConfigDir := filepath.Join(m.tempDir, component, commit, configCommit, "env_config")
	if err := os.RemoveAll(envConfigDir); err != nil {
		return "", err
	}
	if err := os.MkdirAll(envConfigDir, 0750); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "/bin/tar",
		"-C", envConfigDir,
		"-x",
		"-f", envConfigTar,
	)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return "", err
	}

	return envConfigDir, nil
}

var showDirectoryContentsOnError bool = false

// Run Manifests Generator binary from extracted deployment artifacts.
// This uses [manifests_generator.ManifestsGenerator.GenerateManifests].
func (m MultiVersionManifestsGenerator) generateManifestsForCommit(
	ctx context.Context,
	component string,
	commit string,
	configCommit string,
	universeConfigFileName string,
	deploymentArtifactsDir string,
	manifestsTar string,
) error {
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("generateManifestsForCommit"))
	log.Info("BEGIN")
	defer log.Info("END")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	manifestsGeneratorBinary := filepath.Join(deploymentArtifactsDir, "go/pkg/universe_deployer/cmd/manifests_generator/manifests_generator_/manifests_generator")
	args := []string{
		"--commit", commit,
		"--commit-dir", deploymentArtifactsDir,
		"--config-commit", configCommit,
		"--component", component,
		"--default-chart-registry", m.DefaultChartRegistry,
		"--output", manifestsTar,
		"--secrets-dir", m.SecretsDir,
		"--universe-config", universeConfigFileName,
	}
	cmd := exec.CommandContext(ctx, manifestsGeneratorBinary, args...)
	cmd.Dir = deploymentArtifactsDir

	// Set environment variables used by helmfile "env" function.
	env := os.Environ()
	env = append(env, "CLUSTER_PREFIX="+m.ClusterPrefix)
	env = append(env, "TEST_ENVIRONMENT_ID="+m.TestEnvironmentId)
	cmd.Env = env

	if err := util.RunCmd(ctx, cmd); err != nil {
		if showDirectoryContentsOnError {
			log.Error(err, "Manifests generator returned an error. Directory listing follows.")
			// For each file, show size and path.
			util.Find(ctx, deploymentArtifactsDir, m.SecretsDir, universeConfigFileName, "-type", "f", "-exec", "stat", `--format="%10s %n"`, "{}", ";")
		}
		return err
	}
	return nil
}

// Read the contents of the manifests tar to determine the Manifests defined in it.
func (m MultiVersionManifestsGenerator) manifestsTarToManifests(ctx context.Context, component string, commit string, configCommit string) (*manifests_generator.Manifests, error) {
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("manifestsTarToManifests"))
	log.Info("BEGIN")
	defer log.Info("END")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	manifests := manifests_generator.Manifests{}
	manifestsTar := m.manifestsTarFileName(component, commit, configCommit)
	manifestsDir := filepath.Join(m.tempDir, fmt.Sprintf("manifests_%s_%s_%s", component, commit, configCommit))
	if err := os.MkdirAll(manifestsDir, 0750); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, "/bin/tar",
		"-C", manifestsDir,
		"-x",
		"-f", manifestsTar,
	)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return nil, err
	}

	// Search extracted tar file for config.json files.
	err := filepath.Walk(manifestsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if base == manifests_generator.ArgoCdConfigFileName {
			log.V(2).Info("found", "path", path)
			relative, err := filepath.Rel(manifestsDir, path)
			if err != nil {
				return err
			}

			// Read config.json file.
			configFileBytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			configFileData := manifests_generator.ConfigFileData{}
			if err := json.Unmarshal(configFileBytes, &configFileData); err != nil {
				return err
			}

			// kubeContext is not stored in config.json but can be derived from the path.
			// This must be consistent with manifests_generator.helmReleasesToManifests().
			kubeContext := filepath.Base(filepath.Dir(filepath.Dir(relative)))

			manifest := manifests_generator.Manifest{
				ConfigCommit:   configCommit,
				ConfigFileData: configFileData,
				ConfigFileName: relative,
				KubeContext:    kubeContext,
				GitCommit:      commit,
			}
			manifests.Manifests = append(manifests.Manifests, manifest)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &manifests, nil
}

func (m MultiVersionManifestsGenerator) manifestsTarFileName(component string, commit string, configCommit string) string {
	return filepath.Join(m.tempDir, fmt.Sprintf("manifests_%s_%s_%s.tar", component, commit, configCommit))
}

// Combine all manifests tars into a single tar file.
// Any duplicate files will be overwritten in a unspecified but deterministic order.
func (m MultiVersionManifestsGenerator) combineManifestsTars(ctx context.Context) error {
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("combineManifestsTars"))
	log.Info("BEGIN")
	defer log.Info("END")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	manifestsDir := filepath.Join(m.tempDir, "manifests")
	if err := os.MkdirAll(manifestsDir, 0750); err != nil {
		return err
	}
	// Extract each tar file in a deterministic order.
	// This ensures that any file replacements are consistent with the same set of commits.
	for _, cc := range m.componentCommits {
		commitManifestsTar := m.manifestsTarFileName(cc.Component, cc.Commit, cc.ConfigCommit)
		cmd := exec.CommandContext(ctx, "/bin/tar",
			"-C", manifestsDir,
			"-x",
			"-f", commitManifestsTar,
		)
		if err := util.RunCmd(ctx, cmd); err != nil {
			return err
		}
	}

	// Create combined manifests tar file.
	cmd := exec.CommandContext(ctx, "/bin/tar",
		"-C", manifestsDir,
		"--sort=name",
		"--owner=root:0",
		"--group=root:0",
		"--mtime=@0",
		"-f", m.CombinedManifestsTar,
		"-cv",
		".",
	)
	cmd.Env = os.Environ()
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	return nil
}
