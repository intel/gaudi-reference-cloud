// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package universe_config

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

var ErrGitBadObject = errors.New("bad object")

// Get Git commit info, fetching from gitRemote if needed.
func getGitCommitInfoWithFetch(ctx context.Context, commit string, gitRepositoryDir string, gitRemote string) (UniverseComponent, error) {
	log := log.FromContext(ctx).WithName("getGitCommitInfoWithFetch")
	universeComponent, err := getGitCommitInfo(ctx, commit, gitRepositoryDir)
	if errors.Is(err, ErrGitBadObject) {
		log.Info("Commit not found locally. Attempting to fetch from remote.", "commit", commit, "gitRemote", gitRemote)
		if err := fetchCommit(ctx, commit, gitRepositoryDir, gitRemote); err != nil {
			return UniverseComponent{}, err
		}
		universeComponent, err = getGitCommitInfo(ctx, commit, gitRepositoryDir)
	}
	return universeComponent, err
}

func getGitCommitInfo(ctx context.Context, commit string, gitRepositoryDir string) (UniverseComponent, error) {
	log := log.FromContext(ctx).WithName("getGitCommitInfo")
	// See "man git show" for the list of field placeholders
	fields := []string{
		"%at", // author date, UNIX timestamp
		"%ae", // author email
		"%an", // author name
		"%ct", // committer date, UNIX timestamp
		"%ce", // committer email
		"%cn", // committer name
		"%s",  // subject (title)
	}
	// Create format that will separate fields by new line.
	format := strings.Join(fields, "%n")
	cmd := exec.CommandContext(ctx,
		"git",
		"show",
		"--no-patch",
		"--format="+format,
		commit,
	)
	cmd.Dir = gitRepositoryDir
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)
	log.V(2).Info("output", "output", output)
	if err != nil {
		if matched, matchErr := regexp.MatchString("bad object", output); matchErr == nil && matched {
			// Commit not found in local git repo.
			return UniverseComponent{}, ErrGitBadObject
		}
		return UniverseComponent{}, err
	}

	values := strings.Split(output, "\n")
	log.V(2).Info("values", "values", values)
	authorDate, err := parseUnixTimestamp(values[0])
	if err != nil {
		return UniverseComponent{}, err
	}
	committerDate, err := parseUnixTimestamp(values[3])
	if err != nil {
		return UniverseComponent{}, err
	}
	universeComponent := UniverseComponent{
		Commit:         commit,
		AuthorDate:     &authorDate,
		AuthorEmail:    values[1],
		AuthorName:     values[2],
		CommitterDate:  &committerDate,
		CommitterEmail: values[4],
		CommitterName:  values[5],
		Subject:        values[6],
	}
	log.V(2).Info("universeComponent", "universeComponent", universeComponent)
	return universeComponent, nil
}

func fetchCommit(ctx context.Context, commit string, gitRepositoryDir string, gitRemote string) error {
	log := log.FromContext(ctx).WithName("fetchCommit")
	cmd := exec.CommandContext(ctx,
		"git",
		"fetch",
		gitRemote,
		commit,
	)
	cmd.Dir = gitRepositoryDir
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)
	log.V(2).Info("output", "output", output)
	if err != nil {
		return err
	}
	return nil
}

func parseUnixTimestamp(timestamp string) (time.Time, error) {
	timestampInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(timestampInt, 0).UTC(), nil
}
