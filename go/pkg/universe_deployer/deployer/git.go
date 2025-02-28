// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/git_pusher"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

func (e *Deployer) SetManifestsGitRemoteWithCredentials(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("SetManifestsGitRemoteWithCredentials")
	if e.DeployArgoCdEnabled {
		gitRemoteUrl, err := url.Parse(e.ManifestsGitRemote)
		if err != nil {
			return err
		}
		gitUsername := gitRemoteUrl.User.Username()
		gitRemoteUrl.User = url.UserPassword(gitUsername, e.GitPassword)
		e.ManifestsGitRemoteWithCredentials = gitRemoteUrl.String()
	} else {
		e.ManifestsGitRemoteWithCredentials = e.ManifestsGitRemote
	}
	log.V(9).Info("SetManifestsGitRemoteWithCredentials", "ManifestsGitRemoteWithCredentials", e.ManifestsGitRemoteWithCredentials)
	return nil
}

func (e *Deployer) GitEnv(ctx context.Context) []string {
	env := os.Environ()
	gitCommitterName := "Universe Deployer"
	gitCommitterEmail := "universe.deployer@intel.com"
	env = append(env, "GIT_AUTHOR_EMAIL="+gitCommitterEmail)
	env = append(env, "GIT_AUTHOR_NAME="+gitCommitterName)
	env = append(env, "GIT_COMMITTER_EMAIL="+gitCommitterEmail)
	env = append(env, "GIT_COMMITTER_NAME="+gitCommitterName)
	return env
}

func (e *Deployer) DeleteLocalManifestsGitRepo(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("DeleteLocalManifestsGitRepo")
	log.Info("BEGIN")
	defer log.Info("END")
	if err := os.RemoveAll(e.IdcArgoCdLocalRepoDir); err != nil {
		return err
	}
	return nil
}

func (e *Deployer) InitManifestsGitRepo(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("InitManifestsGitRepo")
	log.Info("BEGIN")
	defer log.Info("END")

	if err := os.RemoveAll(e.IdcArgoCdLocalRepoDir); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(e.IdcArgoCdLocalRepoDir), 0750); err != nil {
		return err
	}

	// Copy initial Argo CD files including the ApplicationSet idc-applications.
	// These files come from /deployment/argocd/idc-argocd-initial-data.
	cmd := exec.CommandContext(ctx, "cp", "-rv", "--dereference", e.IdcArgoCdInitialDataDir, e.IdcArgoCdLocalRepoDir)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "init", "-b", e.ManifestsGitBranch, ".")
	cmd.Dir = e.IdcArgoCdLocalRepoDir
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	// Add an initial commit with no files.
	commitMessage := "Empty commit generated by go/pkg/universe_deployer/deployer/git.go InitGitRepo"
	cmd = exec.CommandContext(ctx,
		"git",
		"commit",
		"--allow-empty",
		"-m", commitMessage,
	)
	cmd.Dir = e.IdcArgoCdLocalRepoDir
	cmd.Env = e.GitEnv(ctx)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	manifestsGitRemote := e.ManifestsGitRemoteWithCredentials
	cmd = exec.CommandContext(ctx, "git", "remote", "add", "origin", manifestsGitRemote)
	cmd.Dir = e.IdcArgoCdLocalRepoDir
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	// Push an empty repo.
	// If this happens to be read by Argo CD, it will not find the ApplicationSet idc-applications and it will not delete any apps.
	cmd = exec.CommandContext(ctx, "git", "push", "-u", "origin", e.ManifestsGitBranch)
	cmd.Dir = e.IdcArgoCdLocalRepoDir
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	// Do not push any files to Git.
	// The initial Argo CD files will remain uncommitted in IdcArgoCdLocalRepoDir.
	// These will be pushed later by [PushManifestsToGitRepo] along with all manifests.
	// This avoids having an empty ApplicationSet, which could result in Argo CD deleting existing Applications.

	return nil
}

func (e *Deployer) PushManifestsToGitRepo(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("PushManifestsToGitRepo")

	shortCommit := e.Options.Commit[0:7]
	commitMessages := []string{
		fmt.Sprintf("Generated by Universe Deployer for %s from commit %s", e.IdcEnv, shortCommit),
		fmt.Sprintf("Environment: %s", e.IdcEnv),
		fmt.Sprintf("Commit: %s", e.Options.Commit),
	}
	buildUrl := os.Getenv("BUILD_URL")
	if buildUrl != "" {
		commitMessages = append(commitMessages, fmt.Sprintf("Build URL: %s", buildUrl))
	}

	gitPusher := git_pusher.GitPusher{
		CommitMessages:          commitMessages,
		LocalGitDir:             e.IdcArgoCdLocalRepoDir,
		ManifestsGitBranch:      e.ManifestsGitBranch,
		ManifestsGitRemote:      e.ManifestsGitRemoteWithCredentials,
		ManifestsTar:            e.ManifestsTar,
		ReplaceOwnedDirectories: true,
		SkipGitDiff:             e.InitializeGitRepo,
		PushToNewBranch:         false,
		DryRun:                  e.Options.GitPusherDryRun,
		PatchCommand:            e.PatchCommand,
		UseExistingLocalGitDir:  e.InitializeGitRepo,
		YqBinary:                e.YqBinary,
	}
	if err := gitPusher.Push(ctx); err != nil {
		log.Error(err, "An error occurred pushing manifests to Git."+
			" If you are deploying to a development environment with Jenkins, you can usually fix this by checking the parameter 'DELETE_GITEA'.")
		return fmt.Errorf("pushing manifests to Git repo: %w: ", err)
	}

	// Configure git repo to not use http proxy.
	// This must be performed after git_pusher.Push because that will delete the local git repo.
	cmd := exec.CommandContext(ctx, "git", "config", "--local", "http.proxy", "")
	cmd.Dir = e.IdcArgoCdLocalRepoDir
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	return nil
}
