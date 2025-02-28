// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package multi_version_manifests_generator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

// Obtain the files for a configCommit.
// Returns the path to the tar file.
func (m MultiVersionManifestsGenerator) getConfigForCommit(ctx context.Context, component string, commit string) (string, error) {
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("getConfigForCommit"))
	log.Info("BEGIN")
	defer log.Info("END")

	timeBegin := time.Now()
	defer func() { log.Info("Total duration", "duration", time.Since(timeBegin)) }()

	cacheKey := fmt.Sprintf("universe_deployer_config_%s_%s.tar", component, commit)

	cached, err := m.cache.IsCached(ctx, cacheKey)
	if err != nil {
		return "", err
	}

	if cached {
		log.Info("Found in cache", "cacheKey", cacheKey)
	} else {
		log.Info("Not found in cache", "cacheKey", cacheKey)

		worktreeDir := filepath.Join(m.tempDir, commit)
		branchName := fmt.Sprintf("config-%s-%s", commit, uuid.NewString())

		tempFileName, err := m.cache.GetTempFilePath(ctx, cacheKey)
		if err != nil {
			return "", err
		}

		if err := util.CreateGitWorktree(
			ctx,
			m.ConfigGitRepositoryDir,
			worktreeDir,
			m.ConfigGitRemote,
			commit,
			branchName,
		); err != nil {
			return "", err
		}

		// Create tar file.
		args := []string{
			"-C", worktreeDir,
			"--sort=name",
			"--owner=root:0",
			"--group=root:0",
			"--mtime=@0",
			"-f", tempFileName,
			"-c",
		}
		args = append(args, filepaths.ConfigDirs()...)
		cmd := exec.CommandContext(ctx, "/bin/tar", args...)
		cmd.Env = os.Environ()
		if err := util.RunCmd(ctx, cmd); err != nil {
			return "", err
		}

		_, err = m.cache.MoveFileToCache(ctx, cacheKey, tempFileName)
		if err != nil {
			return "", err
		}

		if err := util.DeleteGitWorktree(ctx, m.ConfigGitRepositoryDir, worktreeDir); err != nil {
			return "", err
		}
	}

	fileResult, err := m.cache.GetFile(ctx, cacheKey)
	if err != nil {
		return "", err
	}

	return fileResult.Path, nil
}
