// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
// Git Pusher pushes Argo CD manifests to a Git repository.
// The binary that is executed is built from the working tree in which "make universe-deployer" runs.

package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/git_pusher"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	flag "github.com/spf13/pflag"
)

type Arguments struct {
	AuthoritativeGitBranch  string
	AuthoritativeGitRemote  string
	LocalGitDir             string
	ManifestsGitBranch      string
	ManifestsGitRemote      string
	ManifestsTar            string
	PatchCommand            string
	PushStateFileName       string
	PushToNewBranch         bool
	ReplaceOwnedDirectories bool
	SourceGitBranch         string
	SourceGitRemote         string
	SourceSequenceNumber    int64
	WorkspaceDir            string
	YqBinary                string
}

func parseArgs() Arguments {
	var args Arguments

	flag.StringVar(&args.AuthoritativeGitBranch, "authoritative-git-branch", "",
		"The monorepo Git branch that is authoritative for manifests-tar")
	flag.StringVar(&args.AuthoritativeGitRemote, "authoritative-git-remote", "",
		"The monorepo Git remote that is authoritative for manifests-tar")
	flag.StringVar(&args.LocalGitDir, "local-git-dir", "",
		"If provided, this directory will be created as the local Git repository that will be pushed."+
			" Otherwise, a temporary directory will be used.")
	flag.StringVar(&args.ManifestsGitBranch, "manifests-git-branch", "",
		"The Argo CD manifests Git branch that this will clone")
	flag.StringVar(&args.ManifestsGitRemote, "manifests-git-remote", "",
		"The Argo CD manifests Git remote that this will clone and create a new branch in")
	flag.StringVar(&args.ManifestsTar, "manifests-tar", "", "Path to the manifests tar file")
	flag.StringVar(&args.PatchCommand, "patch-command", "",
		"If set, this command will executed to apply custom patches to the extracted manifests.")
	flag.StringVar(&args.PushStateFileName, "push-state-file-name", "",
		"If set, this file in the manifests-git-branch will be used so that an older commit of the monorepo is not pushed to this repository.")
	flag.BoolVar(&args.PushToNewBranch, "push-to-new-branch", true,
		"If true, a new branch will be created in the remote. Otherwise, manifests-git-branch will be updated.")
	flag.BoolVar(&args.ReplaceOwnedDirectories, "replace-owned-directories", true,
		"If true, owned directories will be completely replaced")
	flag.StringVar(&args.SourceGitBranch, "source-git-branch", "",
		"The monorepo Git branch used to build manifests-tar")
	flag.StringVar(&args.SourceGitRemote, "source-git-remote", "",
		"The monorepo Git remote used to build manifests-tar")
	flag.Int64Var(&args.SourceSequenceNumber, "source-sequence-number", 0,
		"If push-state-file-name is set, this should be the sequence number (count of commits) of source-git-branch.")
	flag.StringVar(&args.YqBinary, "yq-binary", "yq", "Path to yq binary")

	flag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	flag.Parse()

	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	args.WorkspaceDir = util.WorkspaceDir()
	args.PatchCommand = util.AbsFromWorkspace(args.PatchCommand)

	if !filepath.IsAbs(args.YqBinary) {
		args.YqBinary = filepath.Join(workingDir, args.YqBinary)
	}

	util.EnsureRequiredStringFlag("manifests-git-branch")
	util.EnsureRequiredStringFlag("manifests-git-remote")
	util.EnsureRequiredStringFlag("yq-binary")

	return args
}

func main() {
	ctx := context.Background()
	log.SetDefaultLogger()
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("git_pusher"))
	log.Info("BEGIN")
	defer log.Info("END")

	err := func() error {
		var args = parseArgs()
		log.Info("args", "args", args)

		authoritative := args.SourceGitRemote == args.AuthoritativeGitRemote &&
			args.SourceGitBranch == args.AuthoritativeGitBranch
		log.Info("authoritative", "authoritative", authoritative)
		dryRun := !authoritative

		gitPusher := git_pusher.GitPusher{
			LocalGitDir:             args.LocalGitDir,
			ManifestsGitBranch:      args.ManifestsGitBranch,
			ManifestsGitRemote:      args.ManifestsGitRemote,
			ManifestsTar:            args.ManifestsTar,
			ReplaceOwnedDirectories: args.ReplaceOwnedDirectories,
			SourceSequenceNumber:    args.SourceSequenceNumber,
			PushStateFileName:       args.PushStateFileName,
			PushToNewBranch:         args.PushToNewBranch,
			DryRun:                  dryRun,
			PatchCommand:            args.PatchCommand,
			YqBinary:                args.YqBinary,
		}
		if err := gitPusher.Push(ctx); err != nil {
			return err
		}

		if !authoritative {
			log.Info("DRY RUN: Git push skipped because the source Git remote/branch is not authorative for this universe.")
			log.Info("source git remote       ", "remote", args.SourceGitRemote)
			log.Info("authoritative git remote", "remote", args.AuthoritativeGitRemote)
			log.Info("source git branch       ", "branch", args.SourceGitBranch)
			log.Info("authoritative git branch", "branch", args.AuthoritativeGitBranch)
		}

		return nil
	}()
	if err != nil {
		log.Error(err, "error")
		os.Exit(1)
	}
}
