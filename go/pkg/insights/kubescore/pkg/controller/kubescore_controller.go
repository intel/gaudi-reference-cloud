// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	shared "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/actions/support"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/config"
	kubescoreConfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/ghclient"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/kubescore/pkg/provider"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type KubeScoreScheduler struct {
	syncTicker     *time.Ticker
	InsightsClient *InsightsClient
	GithubClient   *ghclient.GHClient
	Provider       *provider.OSSProvider
	Cfg            *kubescoreConfig.Config
}

func NewKubeScoreScheduler(insightsClient *InsightsClient, ghclient *ghclient.GHClient,
	provider *provider.OSSProvider,
	cfg *kubescoreConfig.Config) (*KubeScoreScheduler, error) {
	if insightsClient == nil {
		return nil, fmt.Errorf("insights client is requied")
	}

	return &KubeScoreScheduler{
		syncTicker:     time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
		InsightsClient: insightsClient,
		GithubClient:   ghclient,
		Provider:       provider,
		Cfg:            cfg,
	}, nil
}

func (ksSchd *KubeScoreScheduler) StartKubeScoreScheduler(ctx context.Context) {
	log := log.FromContext(ctx).WithName("KubeScoreScheduler.StartKubeScoreScheduler")
	log.Info("start kube score scheduler")
	ksSchd.ReleaseDiscoveryLoop(ctx)
}

func (ksSchd *KubeScoreScheduler) ReleaseDiscoveryLoop(ctx context.Context) {
	log := log.FromContext(ctx).WithName("KubeScoreScheduler.ReleaseDiscoveryLoop")
	log.Info("kubernetes release discovery")
	for {
		ksSchd.DiscoverKubernetesReleases(ctx)
		ksSchd.DiscoverCalicoReleases(ctx)
		tm := <-ksSchd.syncTicker.C
		if tm.IsZero() {
			return
		}
	}
}

func (ksSchd *KubeScoreScheduler) DiscoverCalicoReleases(ctx context.Context) {
	log := log.FromContext(ctx).WithName("KubeScoreScheduler.DiscoverCalicoReleases")
	log.Info("entering a new calico release discovery", "config", ksSchd.Cfg)

	releases, err := ksSchd.Provider.GetCalicoReleases(ctx, ksSchd.GithubClient, nil, 20)
	if err != nil {
		log.Error(err, "error listing calico releases")
		return
	}

	for _, r := range releases {
		report := common.ReleaseReport{}
		report.ReleaseTag = r.Tag

		rmd, err := ksSchd.Provider.GetCalicoReleaseMeta(ctx, r.Tag, ksSchd.GithubClient)
		if err != nil {
			log.Error(err, "version validation failed ", r.Tag)
			return
		}
		log.Info("calico metadata", "Calico version", r.Tag, "Release time", rmd.CreatedAt.String())

		rimgs, err := ksSchd.Provider.GetCalicoReleaseImages(ctx, r.Tag, ksSchd.GithubClient)
		if err != nil {
			log.Error(err, "failed to discover calico images for", r.Tag)
			return
		}

		supportMd, err := support.GetSupportMD(ctx, r.Tag)
		if err != nil {
			log.Info("", "error querying release support", err)
			//continue
		}

		report.SupportMD = supportMd
		report.Images = rimgs
		report.ReleaseMD = rmd

		if err := ksSchd.InsightsClient.StoreReleaseMetadata(ctx, report); err != nil {
			log.Error(err, "error storing calico release metadata into insights")
			return
		}
		log.Info("calico", "release metadata stored successfully to insights ", r.Tag)
	}

	// discover third party component releases
	//TODO: fixme with generic deps loop
	for _, tpComp := range ksSchd.Cfg.ThirdPartyComponentPolicy {
		rcmeta, err := ksSchd.Provider.GetReleaseComponentMeta(ctx, ksSchd.GithubClient, tpComp.GitHubSource, 10)
		if err != nil {
			log.Error(err, "version discovery failed for component ", tpComp)
			return
		}
		log.Info("third-party component discovery", "component name", tpComp.ComponentName, "# releases ", len(rcmeta))
		compatibleVersion := findCompabilitySupport(ctx, releases, rcmeta, tpComp)
		log.Info("third-party component filtering", "#compatible versions", len(compatibleVersion), "compatible versions", compatibleVersion)
		if err := ksSchd.InsightsClient.StoreReleaseComponent(ctx, tpComp.ComponentName, findCompabilitySupport(ctx, releases, rcmeta, tpComp)); err != nil {
			log.Error(err, "error storing calico release metadata into insights")
			return
		}
	}

	defer log.Info("returning from new calico release discovery")
}

func (ksSchd *KubeScoreScheduler) DiscoverKubernetesReleases(ctx context.Context) {
	log := log.FromContext(ctx).WithName("KubeScoreScheduler.DiscoverKubernetesReleases")
	log.Info("entering a new kubernetes release discovery", "config", ksSchd.Cfg)

	releases, err := ksSchd.Provider.GetReleases(ctx, ksSchd.GithubClient, nil, 50)
	if err != nil {
		log.Error(err, "error listing releases")
		return
	}

	log.Info("number of new releases discovered since previous scan", "# new releases", len(releases))

	for _, r := range releases {
		report := common.ReleaseReport{}
		report.ReleaseTag = r.Tag

		rmd, err := ksSchd.Provider.GetReleaseMeta(ctx, r.Tag, ksSchd.GithubClient)
		if err != nil {
			log.Error(err, "version validation failed ", r.Tag)
			return
		}
		log.Info("kube metadata", "Kubernetes version", r.Tag, "Release time", rmd.CreatedAt.String())

		rimgs, err := ksSchd.Provider.GetReleaseImages(ctx, r.Tag, ksSchd.GithubClient)
		if err != nil {
			log.Error(err, "failed to discover images for", r.Tag)
			return
		}

		// scanner := vulns.SnykScanner{}
		// scanner.Init(ksSchd.Cfg)

		// for idx, img := range rimgs {
		// 	vData, err := scanner.ScanImage(ctx, img.URL)
		// 	if err != nil {
		// 		fmt.Printf("error scanning image: %s\n", img.URL)
		// 		continue
		// 	}
		// 	rimgs[idx].Vulnerabilities = vData
		// }

		supportMd, err := support.GetSupportMD(ctx, r.Tag)
		if err != nil {
			log.Info("", "error querying release support", err)
			//continue
		}

		report.SupportMD = supportMd
		report.Images = rimgs
		report.ReleaseMD = rmd

		if err := ksSchd.InsightsClient.StoreReleaseMetadata(ctx, report); err != nil {
			log.Error(err, "error storing release metadata into insights")
			return
		}
		log.Info("kubescore", "release metadata stored successfully to insights ", r.Tag)

		// get release sbom
		sbomfp, err := os.CreateTemp(os.TempDir(), r.Tag)
		if err != nil {
			log.Error(err, "error creating file")
			continue
		}

		err = ksSchd.Provider.GetReleaseSBOM(ctx, r.Tag, sbomfp.Name())
		if err != nil {
			log.Error(err, "sbom retrieval failed for release ", "releaseId", r.Tag)
			continue
		}
		sbomBuf, err := io.ReadAll(sbomfp)
		if err != nil {
			log.Error(err, "error reading sbom into buf")
		}
		if sbomBuf != nil {
			if err := ksSchd.InsightsClient.StoreReleaseSBOM(ctx, r.Tag, "spdx", time.Now(), sbomBuf); err != nil {
				log.Error(err, "error storing release sbom into insights")
				continue
			}
		} else {
			log.Info("empty sbom ", "releaseId", r.Tag)
		}
	}

	// discover third party component releases
	//TODO: fixme with generic deps loop
	for _, tpComp := range ksSchd.Cfg.ThirdPartyComponentPolicy {
		rcmeta, err := ksSchd.Provider.GetReleaseComponentMeta(ctx, ksSchd.GithubClient, tpComp.GitHubSource, 10)
		if err != nil {
			log.Error(err, "version discovery failed for component ", tpComp)
			return
		}
		log.Info("third-party component discovery", "component name", tpComp.ComponentName, "# releases ", len(rcmeta))
		compatibleVersion := findCompabilitySupport(ctx, releases, rcmeta, tpComp)
		log.Info("third-party component filtering", "#compatible versions", len(compatibleVersion), "compatible versions", compatibleVersion)
		if err := ksSchd.InsightsClient.StoreReleaseComponent(ctx, tpComp.ComponentName, findCompabilitySupport(ctx, releases, rcmeta, tpComp)); err != nil {
			log.Error(err, "error storing k8s release metadata into insights")
			return
		}
	}

	defer log.Info("returning from new kubernetes release discovery")
}

func findCompabilitySupport(ctx context.Context, k8sReleases []common.ReleaseMD, compReleases []common.ReleaseMD, policy config.ThirdPartyComponentPolicy) []common.ReleaseComponentMD {
	log := log.FromContext(ctx).WithName("findCompabilitySupport")
	comps := []common.ReleaseComponentMD{}

	for _, kr := range k8sReleases {
		isValidK8s := false
		compVersionPolicy := ""
		for _, p := range policy.Policies {
			if constraintsOk(kr.Tag, p.K8sVersions) {
				isValidK8s = true
				compVersionPolicy = p.MinimumVersion
				break
			}
		}
		if !isValidK8s {
			log.Info("k8s version constraint failed", "k8s version", kr.Tag, "policy constraints", policy.Policies)
			continue
		}
		compCnt := 0
		for _, cr := range compReleases {
			if !common.IsGreater(cr.Tag, compVersionPolicy) {
				log.Info("component version constraint failed", "component version", cr.Tag, "policy constraint", compVersionPolicy)
				continue
			}
			c := common.ReleaseComponentMD{
				ReleaseId:        kr.Tag,
				ComponentName:    cr.Name,
				License:          cr.License,
				ComponentVersion: cr.Tag,
				ReleaseTime:      cr.CreatedAt,
				Purl:             cr.URL,
				Type:             shared.ComponentTypeGitrepo,
			}
			comps = append(comps, c)
			compCnt++
			if compCnt >= policy.TopK {
				break
			}
		}
	}
	return comps
}

func constraintsOk(version, constraint string) bool {
	//TODO: Add some fail checks
	return common.CompareWithConstraints(version, constraint)
}
