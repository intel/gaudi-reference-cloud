// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	"github.com/sethvargo/go-retry"
)

// Build and push containers, charts, and artifacts.
// Components with HEAD commit are built and pushed from the current source tree using the Bazel deployment_artifacts_COMPONENT_tar rule.
// Components with non-HEAD commits are built and pushed using the Bazel create_releases rule.
func (e *Deployer) BuilderPusher(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BuilderPusher")
	log.Info("BEGIN")
	defer log.Info("END")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	if err := e.DownloadOrCreateNonHeadReleases(ctx); err != nil {
		return err
	}
	if e.HasHeadCommit {
		log.Info("Building deployment artifacts from HEAD")
		if err := e.BuildHead(ctx); err != nil {
			return err
		}
		if e.Options.IncludePush {
			log.Info("Pushing containers and charts from HEAD")
			if err := e.PushHead(ctx); err != nil {
				return err
			}
		}
	} else {
		log.Info("Building and pushing from HEAD is not required.")
	}
	return nil
}

// Download or create releases for each non-HEAD commit.
func (e *Deployer) DownloadOrCreateNonHeadReleases(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("DownloadOrCreateNonHeadReleases")
	log.Info("BEGIN")
	defer log.Info("END")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	var createReleasesUniverseConfig *universe_config.UniverseConfig
	createReleasesRequired := false
	forceCreateReleases := e.Options.EnvConfig.EnvConfig.Values.UniverseDeployer.ForceCreateReleases

	if forceCreateReleases {
		// Create trimmed Universe Config that does not include the HEAD commit.
		hasComponents := false
		createReleasesUniverseConfig = e.Options.UniverseConfig.DeepCopy()
		pred := func(component string, universeComponent universe_config.UniverseComponent) bool {
			if universeComponent.Commit != e.Options.Commit {
				hasComponents = true
				return true
			}
			return false
		}
		createReleasesUniverseConfig = createReleasesUniverseConfig.Filtered(ctx, pred)
		if !hasComponents {
			log.Info("ForceCreateReleases is true but there are no non-HEAD commits. Multi-version build is not required.")
		} else {
			log.Info("ForceCreateReleases is true and there are non-HEAD commits. Multi-version build is required.")
			createReleasesRequired = true
		}
	} else {
		// For each non-HEAD commit:
		//  Attempt cached download of deployment artifacts.
		//  If failed, add to build list.
		downloadResult, err := e.ManifestsGenerator.TryDownloadDeploymentArtifacts(ctx)
		if err != nil {
			return err
		}
		log.Info("Download result", "downloadResult", downloadResult)

		if len(downloadResult.NotFoundErrors) == 0 {
			log.Info("All required non-HEAD deployment artifacts have been downloaded. Multi-version build is not required.")
		} else {
			log.Info("Some required non-HEAD deployment artifacts could not be downloaded. Multi-version build is required.")

			// Create trimmed Universe Config that only includes the component/commits that could not be downloaded from Artifactory.
			// The HEAD commit will never be included.
			createReleasesUniverseConfig = e.Options.UniverseConfig.DeepCopy()
			pred := func(componentName string, componentDetails universe_config.UniverseComponent) bool {
				for _, notFoundError := range downloadResult.NotFoundErrors {
					if componentName == notFoundError.Component && componentDetails.Commit == notFoundError.Commit {
						return true
					}
				}
				return false
			}
			createReleasesUniverseConfig = createReleasesUniverseConfig.Filtered(ctx, pred)
			createReleasesRequired = true
		}
	}

	if createReleasesRequired {
		log.Info("createReleasesUniverseConfig", "createReleasesUniverseConfig", createReleasesUniverseConfig)
		if err := e.CreateNonHeadReleases(ctx, *createReleasesUniverseConfig); err != nil {
			return err
		}
	}

	return nil
}

// Run Bazel to create releases for all commits in Universe Config.
// This only runs with non-HEAD commits.
//   - Build and push containers and charts to Docker registry.
//   - Build and upload deployment artifacts to Artifactory.
func (e *Deployer) CreateNonHeadReleases(ctx context.Context, universeConfig universe_config.UniverseConfig) error {
	log := log.FromContext(ctx).WithName("CreateNonHeadReleases")
	log.Info("BEGIN")
	defer log.Info("END")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	// Create Universe Config file that will be used during Bazel build.
	createReleasesUniverseConfigFile := filepath.Join(e.WorkspaceDir, filepaths.CreateReleasesUniverseConfig)
	log.Info("Writing Universe Config file", "createReleasesUniverseConfigFile", createReleasesUniverseConfigFile)
	if err := universeConfig.WriteFile(ctx, createReleasesUniverseConfigFile); err != nil {
		return err
	}

	// Write /build/dynamic/ARTIFACT_REPOSITORY_URL.
	// This will be during the Bazel build by /go/pkg/universe_deployer/cmd/create_release/main.go.
	artifactRepositoryUrl := e.Options.ArtifactRepositoryUrl.String()
	if err := os.WriteFile(filepath.Join(e.WorkspaceDir, "build/dynamic/ARTIFACT_REPOSITORY_URL"), []byte(artifactRepositoryUrl), 0640); err != nil {
		return err
	}

	args := []string{}
	args = append(args, e.Options.BazelStartupOpts...)
	args = append(args, "build")
	args = append(args, "--jobs", os.Getenv("UNIVERSE_DEPLOYER_JOBS_PER_PIPELINE"))
	args = append(args, "--sandbox_writable_path="+os.Getenv("UNIVERSE_DEPLOYER_POOL_DIR"))
	args = append(args, "--verbose_failures")
	args = append(args, e.Options.BazelBuildOpts...)
	args = append(args, "//deployment/universe_deployer/create_releases:create_releases")

	backoff := retry.WithMaxRetries(6, retry.NewExponential(1*time.Second))
	err := retry.Do(ctx, backoff, func(ctx context.Context) error {
		cmd := exec.CommandContext(ctx, e.Options.BazelBinary, args...)
		cmd.Dir = e.WorkspaceDir
		timeBuildStart := time.Now()
		err := util.RunCmd(ctx, cmd)
		log.Info("Build duration", "duration", time.Since(timeBuildStart))
		if err != nil {
			log.Error(err, "retryable error running bazel build")
			return retry.RetryableError(err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (e *Deployer) BuildHead(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("BuildHead")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	// TODO: filter components
	if err := e.DeploymentArtifactsBuilder.Build(ctx); err != nil {
		return err
	}
	return nil
}

func (e *Deployer) PushHead(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("PushHead")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	// TODO: filter components
	if err := e.ContainerAndChartPusher.Push(ctx); err != nil {
		return err
	}
	return nil
}
