// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"os"

	kubescoreConfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/controller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/ghclient"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/provider"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type KubeScoreService struct {
}

func (svc *KubeScoreService) Init(ctx context.Context, cfg *kubescoreConfig.Config) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	log.Info("initializing IDC kube score scheduler...")

	insightsClient, err := controller.NewInsightClient(ctx, cfg.SecurityInsights.URL)
	if err != nil {
		log.Error(err, "failed to initialize insights client")
		return err
	}

	token, err := os.ReadFile(cfg.GitHubAPI.Key)
	if err != nil {
		log.Info("unable to read GithubKey file %s: %v", cfg.GitHubAPI.Key, err)
	}

	log.Info("Run", "cfg", cfg)
	if err := os.Setenv("GITHUB_TOKEN", string(token)); err != nil {
		log.Error(err, "failed to set env GITHUB_TOKEN")
	}
	ghcli := ghclient.GHClient{}
	if err := ghcli.Setup(ctx, string(token)); err != nil {
		log.Error(err, "failed to initialize github client")
		return fmt.Errorf("error connecting to github")
	}

	provider := provider.OSSProvider{
		KubeRepoURL:       "https://github.com/kubernetes/kubernetes",
		CalicoRegistryURL: cfg.ThirdPartyComponentPolicy[4].GitHubSource,
	}

	kubescoreSched, err := controller.NewKubeScoreScheduler(insightsClient, &ghcli, &provider, cfg)
	if err != nil {
		log.Error(err, "error starting kubescore scheduler")
		return err
	}
	kubescoreSched.StartKubeScoreScheduler(ctx)

	return nil
}

func (svc *KubeScoreService) Name() string {
	return "iks-kube-score-scheduler"
}
