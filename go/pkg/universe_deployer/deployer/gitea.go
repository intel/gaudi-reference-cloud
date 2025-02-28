// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package deployer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
	"k8s.io/client-go/util/retry"
)

func (e *Deployer) DeleteGitea(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("DeleteGitea")

	if err := e.RunHelmfile(ctx,
		"destroy",
		"--file", "helmfile-argocd.yaml",
		"--selector", "name=gitea",
	); err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, e.KubectlBinary, "delete", "-n", "gitea", "pvc", "--all")
	if err := util.RunCmd(ctx, cmd); err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, e.KubectlBinary, "delete", "namespace/gitea", "--wait")
	if err := util.RunCmd(ctx, cmd); err != nil {
		log.Error(err, "delete namespace/gitea")
	}

	return nil
}

func (e *Deployer) DeployGitea(ctx context.Context) error {
	if err := e.RunHelmfile(ctx,
		"apply",
		"--file", "helmfile-argocd.yaml",
		"--selector", "name=gitea",
		"--skip-diff-on-install",
		"--wait",
	); err != nil {
		return err
	}

	if err := e.StartGiteaPortForward(ctx); err != nil {
		return err
	}

	return nil
}

func (e *Deployer) StartGiteaPortForward(ctx context.Context) error {
	log := log.FromContext(ctx).WithName("StartGiteaPortForward")
	localPort, err := e.StartPortForward(ctx, "gitea", "service/gitea-http", "http")
	if err != nil {
		return err
	}

	backoff := LinearBackoff(3*time.Minute, 1*time.Second)
	if err := retry.OnError(backoff, func(error) bool { return true }, func() error {
		_, err := http.Get(fmt.Sprintf("http://localhost:%d", localPort))
		return err
	}); err != nil {
		return fmt.Errorf("checking status of Gitea through port forward: %w", err)
	}

	gitRemoteUrl, err := url.Parse(e.ManifestsGitRemote)
	if err != nil {
		return err
	}
	gitRemoteUrl.Host = fmt.Sprintf("localhost:%d", localPort)
	e.ManifestsGitRemote = gitRemoteUrl.String()
	log.Info("ManifestsGitRemote", "ManifestsGitRemote", e.ManifestsGitRemote)

	if err := e.SetManifestsGitRemoteWithCredentials(ctx); err != nil {
		return err
	}

	return nil
}
