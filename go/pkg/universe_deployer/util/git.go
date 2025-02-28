// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
)

var reGitCommit = regexp.MustCompile("^[a-f0-9]{40}$")

const HEAD = "HEAD"

func IsGitCommit(commit string) bool {
	return reGitCommit.MatchString(commit)
}

func CreateGitWorktree(
	ctx context.Context,
	gitRepositoryDir string,
	worktreeDir string,
	gitRemote string,
	ref string,
	branchName string,
) error {
	cmd := exec.CommandContext(ctx,
		"git",
		"fetch",
		gitRemote,
		ref,
	)
	cmd.Dir = gitRepositoryDir
	if err := RunCmd(ctx, cmd); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx,
		"git",
		"worktree",
		"add",
		"-B", branchName,
		worktreeDir,
		ref,
	)
	cmd.Dir = gitRepositoryDir
	if err := RunCmd(ctx, cmd); err != nil {
		return err
	}

	return nil
}

func DeleteGitWorktree(
	ctx context.Context,
	gitRepositoryDir string,
	worktreeDir string,
) error {
	cmd := exec.CommandContext(ctx,
		"git",
		"worktree",
		"remove",
		worktreeDir,
	)
	cmd.Dir = gitRepositoryDir
	if err := RunCmd(ctx, cmd); err != nil {
		return err
	}

	return nil
}

// List references in a remote repository.
// Returns a map from reference to commit hash.
// References will be in a format like refs/heads/add-training-coupon or refs/tags/v1.0.0.
func GitLsRemote(ctx context.Context, gitRepositoryDir string, gitRemote string) (map[string]string, error) {
	refToCommitMap := map[string]string{}
	cmd := exec.CommandContext(ctx, "git", "ls-remote", gitRemote)
	cmd.Dir = gitRepositoryDir
	output, err := RunCmdOutput(ctx, cmd)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(output, "\n")
	// line is formatted like (tab separates the words):
	// 620f7645e4be998f1f3662590865b86bf24ebdb4        refs/heads/add-training-coupon
	// 81f6d6929974a77d4fce29ac2b124d2162bcf07f        refs/tags/v1.0.0
	re := regexp.MustCompile(`^([a-f0-9]{40})\t(\S+)$`)
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			commit := matches[1]
			ref := matches[2]
			refToCommitMap[ref] = commit
		}
	}
	return refToCommitMap, nil
}
