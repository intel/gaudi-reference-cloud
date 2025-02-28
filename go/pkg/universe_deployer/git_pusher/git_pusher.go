// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package git_pusher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

// A type for persisting the state of git_pusher in the Argo CD manifests Git repository.
type PushState struct {
	Doc            string `json:"_doc,omitempty"`
	SequenceNumber int64  `json:"sequenceNumber"`
}

type GitPusher struct {
	CommitMessages          []string
	LocalGitDir             string
	ManifestsGitBranch      string
	ManifestsGitRemote      string
	ManifestsTar            string
	ReplaceOwnedDirectories bool
	SkipGitDiff             bool
	SourceSequenceNumber    int64
	PushStateFileName       string
	PushToNewBranch         bool
	DryRun                  bool
	PatchCommand            string
	// If false, the directory LocalGitDir (if provided) will be deleted and the repo will be cloned into it from ManifestsGitRemote.
	// Set to true to skip cloning and push any uncommitted files to ManifestsGitRemote in addition to generated manifests.
	UseExistingLocalGitDir bool
	YqBinary               string
}

// Push Argo CD manifests to a Git repository.
func (p GitPusher) Push(ctx context.Context) error {
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("git_pusher.Push"))
	log.Info("BEGIN")
	defer log.Info("END")

	localGitDir := p.LocalGitDir
	if localGitDir == "" {
		tempDir, err := os.MkdirTemp("", "universe_deployer_git_pusher_")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)
		localGitDir = tempDir
	} else if !p.UseExistingLocalGitDir {
		if err := os.RemoveAll(localGitDir); err != nil {
			return err
		}
		if err := os.MkdirAll(localGitDir, 0750); err != nil {
			return err
		}
	}

	// Clone Git repo with Argo CD manifests.
	if !p.UseExistingLocalGitDir {
		cmd := exec.CommandContext(ctx,
			"git",
			"-c", "color.ui=always",
			"clone",
			"--branch", p.ManifestsGitBranch,
			"--depth", "1",
			"--single-branch",
			"--no-tags",
			p.ManifestsGitRemote,
			".",
		)
		cmd.Dir = localGitDir
		if err := util.RunCmd(ctx, cmd); err != nil {
			return err
		}
	}

	// Git checkout new branch.
	pushToBranch := p.ManifestsGitBranch
	if p.PushToNewBranch {
		timestamp := time.Now().UTC()
		pushToBranch = fmt.Sprintf("universe-deployer-%s", timestamp.Format("2006-01-02T15-04-05"))
		cmd := exec.CommandContext(ctx,
			"git",
			"-c", "color.ui=always",
			"checkout",
			"-b", pushToBranch,
		)
		cmd.Dir = localGitDir
		if err := util.RunCmd(ctx, cmd); err != nil {
			return err
		}
	}

	// Read push state file from Git repo.
	// If file does not exist, this will assume sequenceNumber of 0 which will always allow updates.
	var pushState PushState
	var pushStatePath string
	if p.PushStateFileName != "" {
		pushStatePath = filepath.Join(localGitDir, p.PushStateFileName)
		pushStateBytes, err := os.ReadFile(pushStatePath)
		if err != nil {
			// Continue if file does not exist.
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		} else {
			if err := json.Unmarshal(pushStateBytes, &pushState); err != nil {
				return err
			}
		}
		log.Info("Original push state in target Git repo", "pushState", pushState)
	}

	// Determine owned directories by searching for marker files in the manifests tar file.
	ownedDirectories, err := FindDirectoriesWithOwnedMarkerInTar(ctx, p.ManifestsTar)
	if err != nil {
		return err
	}
	log.Info("ownedDirectories", "ownedDirectories", ownedDirectories)

	// Delete owned directories from the new Git branch.
	if p.ReplaceOwnedDirectories {
		for _, ownedDirectory := range ownedDirectories {
			ownedDirectoryPath := filepath.Join(localGitDir, ownedDirectory)
			log.Info("Deleting owned directory", "ownedDirectoryPath", ownedDirectoryPath)
			if err := os.RemoveAll(ownedDirectoryPath); err != nil {
				return err
			}
		}
	}

	// Extract manifests tar on top of new Git branch.
	cmd := exec.CommandContext(ctx, "/bin/tar",
		"-C", localGitDir,
		"-xv",
		"-f", p.ManifestsTar,
	)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	// Apply patches.
	if p.PatchCommand != "" {
		cmd = exec.CommandContext(ctx, p.PatchCommand)
		cmd.Dir = localGitDir
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "YQ="+p.YqBinary)
		if err := util.RunCmd(ctx, cmd); err != nil {
			return err
		}
	}

	// Log git status.
	cmd = exec.CommandContext(ctx,
		"git",
		"-c", "color.ui=always",
		"status",
	)
	cmd.Dir = localGitDir
	if err := util.RunCmdWithoutCapture(ctx, cmd); err != nil {
		return err
	}

	// Git add.
	cmd = exec.CommandContext(ctx,
		"git",
		"-c", "color.ui=always",
		"add",
		"--all",
		"--verbose",
	)
	cmd.Dir = localGitDir
	if err := util.RunCmdWithoutCapture(ctx, cmd); err != nil {
		return err
	}

	// Check whether any changes were made.
	// git diff exits with 1 if there were differences and 0 means no differences.
	changed := false
	cmd = exec.CommandContext(ctx,
		"git",
		"-c", "color.ui=always",
		"diff",
		"--staged",
		"--exit-code",
		"--quiet",
	)
	cmd.Dir = localGitDir
	if err := util.RunCmd(ctx, cmd); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				changed = true
			} else {
				return err
			}
		} else {
			return err
		}
	}

	// Run git diff again to log all changed lines.
	if !p.SkipGitDiff {
		log.Info("\n" +
			`~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~` + "\n" +
			`~~~ git diff BEGIN ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~` + "\n" +
			`~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~`)
		cmd = exec.CommandContext(ctx,
			"git",
			"-c", "color.ui=always",
			"diff",
			"--staged",
		)
		cmd.Dir = localGitDir
		if err := util.RunCmdWithoutCapture(ctx, cmd); err != nil {
			return err
		}
		log.Info("\n" +
			`~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~` + "\n" +
			`~~~ git diff END ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~` + "\n" +
			`~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~`)
	}

	gitCommitterName := "Universe Deployer"
	gitCommitterEmail := "universe.deployer@intel.com"

	if changed {
		if !p.DryRun && p.PushStateFileName != "" {
			// Ensure that sequence number does not decrease.
			if p.SourceSequenceNumber < pushState.SequenceNumber {
				return fmt.Errorf("git_pusher aborted because a newer commit of the monorepo has already been pushed. "+
					"The push state file %s in branch %s of remote %s has sequence number %d "+
					"but the monorepo source sequence number (the count of commits) "+
					"for this execution of Git Pusher is %d. "+
					"Usually this error indicates that Jenkins ran Universe Deployer out of order, in which case there is no need to rerun this. "+
					"If this error persists, you may delete the push state file.",
					p.PushStateFileName, p.ManifestsGitBranch, p.ManifestsGitRemote, pushState.SequenceNumber, p.SourceSequenceNumber)
			}

			log.Info("Previous push state", "pushState", pushState)

			// Write new push state file.
			pushState.Doc = fmt.Sprintf("This file is used by Universe Deployer Git Pusher to ensure that an older commit of the monorepo is not "+
				"pushed to this repository. "+
				"The sequenceNumber field equals the count of commits in the monorepo branch. "+
				"To disable this check one time and allow Git Pusher to update this repository, delete this file or set sequenceNumber to 0. "+
				"To prevent Git Pusher from updating this repository, set sequenceNumber to %d.", math.MaxInt64)
			pushState.SequenceNumber = p.SourceSequenceNumber
			log.Info("Writing new push state", "pushState", pushState)
			pushStateBytes, err := json.MarshalIndent(pushState, "", "  ")
			if err != nil {
				return nil
			}
			if err := os.WriteFile(pushStatePath, pushStateBytes, 0644); err != nil {
				return err
			}

			// Git add push state file.
			cmd = exec.CommandContext(ctx,
				"git",
				"-c", "color.ui=always",
				"add",
				p.PushStateFileName,
				"--verbose",
			)
			cmd.Dir = localGitDir
			if err := util.RunCmd(ctx, cmd); err != nil {
				return err
			}
		}

		// Git commit.
		commitMessages := p.CommitMessages
		if len(commitMessages) == 0 {
			commitMessages = []string{"Generated by Universe Deployer"}
		}
		commitArgs := []string{
			"-c", "color.ui=always",
			"commit",
		}
		for _, commitMessage := range commitMessages {
			commitArgs = append(commitArgs, "-m", commitMessage)
		}
		cmd = exec.CommandContext(ctx, "git", commitArgs...)
		cmd.Dir = localGitDir
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "GIT_AUTHOR_EMAIL="+gitCommitterEmail)
		cmd.Env = append(cmd.Env, "GIT_AUTHOR_NAME="+gitCommitterName)
		cmd.Env = append(cmd.Env, "GIT_COMMITTER_EMAIL="+gitCommitterEmail)
		cmd.Env = append(cmd.Env, "GIT_COMMITTER_NAME="+gitCommitterName)
		if err := util.RunCmd(ctx, cmd); err != nil {
			return err
		}

		// Git push.
		if !p.DryRun {
			cmd = exec.CommandContext(ctx,
				"git",
				"-c", "color.ui=always",
				"push",
				"--set-upstream", "origin",
				pushToBranch,
			)
			cmd.Dir = localGitDir
			if err := util.RunCmd(ctx, cmd); err != nil {
				return err
			}
		}
	} else {
		log.Info("No changes made")
	}

	return nil
}

// Returned directories will be relative to dir.
func FindDirectoriesWithOwnedMarker(ctx context.Context, dir string) ([]string, error) {
	foundDirs := []string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == util.GetOwnedDirectoryMarkerFileName() {
			relative, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			ownedDirectory := filepath.Dir(relative)
			foundDirs = append(foundDirs, ownedDirectory)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return foundDirs, nil
}

func FindDirectoriesWithOwnedMarkerInTar(ctx context.Context, tarFile string) ([]string, error) {
	tempDir, err := os.MkdirTemp("", "universe_deployer_git_pusher_")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	cmd := exec.CommandContext(ctx, "/bin/tar",
		"-C", tempDir,
		"-x",
		"-f", tarFile,
	)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return nil, err
	}
	return FindDirectoriesWithOwnedMarker(ctx, tempDir)
}
