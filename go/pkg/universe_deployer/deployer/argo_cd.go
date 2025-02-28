// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
)

func (e *Deployer) ListArgoApplications(ctx context.Context) (*unstructured.UnstructuredList, error) {
	k8sRestConfig, err := e.GetGlobalK8sRestConfig()
	if err != nil {
		return nil, err
	}
	k8sClient, err := dynamic.NewForConfig(k8sRestConfig)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}
	applications, err := k8sClient.Resource(gvr).Namespace("argocd").List(ctx, metav1.ListOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			// This will occur if Argo CD is not installed. Return an empty list of applications.
			return &unstructured.UnstructuredList{}, nil
		}
		return nil, fmt.Errorf("listing Argo applications: %w", err)
	}
	return applications, nil
}

func (e *Deployer) GetMatchingArgoApplicationNames(
	ctx context.Context,
	matchRegex string,
	excludeRegex string,
) ([]string, error) {
	if matchRegex == "" {
		return nil, nil
	}
	reMatch := regexp.MustCompile(matchRegex)
	reExclude := regexp.MustCompile(excludeRegex)
	applications, err := e.ListArgoApplications(ctx)
	if err != nil {
		return nil, err
	}
	applicationNames := []string{}
	for _, application := range applications.Items {
		name := application.GetName()
		if !reExclude.MatchString(name) && reMatch.MatchString(name) {
			applicationNames = append(applicationNames, name)
		}
	}
	return applicationNames, nil
}

func (e *Deployer) WaitForArgoApplicationsToBeDeleted(
	ctx context.Context,
	applicationsToDeleteRegex string,
	applicationsToNotDeleteRegex string,
) error {
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("WaitForArgoApplicationsToBeDeleted"))
	log.Info("Waiting for Argo CD Applications to be deleted")
	backoff := LinearBackoff(15*time.Minute, 2*time.Second)
	if err := retry.OnError(backoff, func(error) bool { return true }, func() error {
		applicationNames, err := e.GetMatchingArgoApplicationNames(ctx, applicationsToDeleteRegex, applicationsToNotDeleteRegex)
		if err != nil {
			return err
		}
		if len(applicationNames) != 0 {
			log.Info("Found applications", "applicationNames", applicationNames)
			return fmt.Errorf("application exists: %v", applicationNames)
		}
		return nil
	}); err != nil {
		return err
	}
	log.Info("Argo CD Applications have been deleted")
	return nil
}

func (e *Deployer) DeleteArgoApplications(
	ctx context.Context,
	applicationsToDeleteRegex string,
	applicationsToNotDeleteRegex string,
) error {
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("DeleteArgoApplications"))
	log.Info("BEGIN")
	defer log.Info("END")

	if applicationsToDeleteRegex == "" {
		return nil
	}

	if err := e.StartGiteaPortForward(ctx); err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp("", "universe_deployer_deployer_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	localGitDir := tempDir

	// Clone Git repo with Argo CD manifests.
	cmd := exec.CommandContext(ctx,
		"git",
		"-c", "color.ui=always",
		"clone",
		"--branch", e.ManifestsGitBranch,
		"--depth", "1",
		"--single-branch",
		"--no-tags",
		e.ManifestsGitRemoteWithCredentials,
		".",
	)
	cmd.Dir = localGitDir
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	reDelete := regexp.MustCompile(applicationsToDeleteRegex)
	reNotDelete := regexp.MustCompile(applicationsToNotDeleteRegex)
	foundDirs := []string{}
	searchDir := filepath.Join(localGitDir, "applications")
	err = filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if !reNotDelete.MatchString(base) && reDelete.MatchString(base) {
			foundDirs = append(foundDirs, path)
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	log.Info("Deleting Argo Application", "applicationsToDeleteRegex", applicationsToDeleteRegex, "foundDirs", foundDirs)

	if len(foundDirs) == 0 {
		return nil
	}

	for _, path := range foundDirs {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	util.Find(ctx, localGitDir)

	cmd = exec.CommandContext(ctx, "git", "add", "--all", "--verbose")
	cmd.Dir = localGitDir
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	commitMessage := "DeleteArgoApplications " + applicationsToDeleteRegex
	cmd = exec.CommandContext(ctx,
		"git",
		"commit",
		"-m", commitMessage,
	)
	cmd.Dir = localGitDir
	cmd.Env = e.GitEnv(ctx)
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "git", "push")
	cmd.Dir = localGitDir
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	return err
}

func (e *Deployer) DeleteArgoCd(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("DeleteArgoCd")

	if err := e.RunHelmfile(ctx,
		"destroy",
		"--file", "helmfile-argocd.yaml",
		"--selector", "name=argo-cd-resources",
	); err != nil {
		return err
	}

	if err := e.RunHelmfile(ctx,
		"destroy",
		"--file", "helmfile-argocd.yaml",
		"--selector", "name=argocd",
	); err != nil {
		return err
	}

	// TODO: only delete if it exists
	cmd := exec.CommandContext(ctx, e.KubectlBinary, "delete", "namespace/argocd", "--wait")
	if err := util.RunCmd(ctx, cmd); err != nil {
		log.Error(err, "delete namespace/argocd")
	}

	return nil
}

func (e *Deployer) DeployArgoCd(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("DeployArgoCd")
	log.Info("BEGIN")
	defer log.Info("END")

	if err := e.RunHelmfile(ctx,
		"apply",
		"--allow-no-matching-release",
		"--file", "helmfile-argocd.yaml",
		"--selector", "name=argocd",
		"--skip-diff-on-install",
		"--wait",
	); err != nil {
		return err
	}

	if err := e.RunHelmfile(ctx,
		"apply",
		"--allow-no-matching-release",
		"--file", "helmfile-argocd.yaml",
		"--selector", "name=argo-cd-resources",
		"--skip-diff-on-install",
	); err != nil {
		return err
	}

	return nil
}
