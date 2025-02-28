// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/security-scanner/pkg/actions/vulns"
	scannerConfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/security-scanner/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type SecurityScanScheduler struct {
	syncTicker     *time.Ticker
	InsightsClient *InsightsClient
	Cfg            *scannerConfig.Config
}

func NewSecurityScanScheduler(insightsClient *InsightsClient,
	cfg *scannerConfig.Config) (*SecurityScanScheduler, error) {
	if insightsClient == nil {
		return nil, fmt.Errorf("insights client is requied")
	}

	return &SecurityScanScheduler{
		syncTicker:     time.NewTicker(time.Duration(cfg.SchedulerInterval) * time.Second),
		InsightsClient: insightsClient,
		Cfg:            cfg,
	}, nil
}

func (secSchd *SecurityScanScheduler) StartSecurityScanScheduler(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("SecurityScanScheduler.StartSecurityScanScheduler")
	logger.Info("start security scan scheduler")
	secSchd.ReleaseDiscoveryLoop(ctx)
}

func (secSchd *SecurityScanScheduler) ReleaseDiscoveryLoop(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("SecurityScanScheduler.ReleaseDiscoveryLoop")
	logger.Info("kubernetes release discovery")
	for {
		secSchd.DiscoverReleases(ctx)
		tm := <-secSchd.syncTicker.C
		if tm.IsZero() {
			return
		}
	}
}

func (secSchd *SecurityScanScheduler) DiscoverReleases(ctx context.Context) {
	logger := log.FromContext(ctx).WithName("SecurityScanScheduler.DiscoverReleases")
	logger.Info("entering a new kubernetes release discovery", "config", secSchd.Cfg)

	images, err := secSchd.InsightsClient.GetAllImages(ctx)
	if err != nil {
		logger.Info("error getting k8s release images, skipping scans")
		return
	}

	for _, img := range images {
		imgUrl := fmt.Sprintf("%s@%s", img.Name, img.Sha256)
		if img.Sha256 == "" {
			imgUrl = fmt.Sprintf("%s@%s", img.Name, img.Version)
		}
		vulnerabilityReport, err := vulns.ScanImage(ctx, imgUrl)
		if err != nil {
			logger.Info("vulnerability scan failed", "imageurl", imgUrl)
			continue
		}
		logger.Info("vulnerability scan completed", "imageurl", imgUrl)

		if vulnerabilityReport == nil || len(vulnerabilityReport) != 1 {
			logger.Info("invalid vulnerability report ", "imageurl", imgUrl)
			continue
		}

		logger.Info("store vulnerability report", "imageurl", imgUrl)
		if err := secSchd.InsightsClient.StoreVulnerabilityReport(ctx, img, vulnerabilityReport[0]); err != nil {
			logger.Info("error storing vulnerability report", "imageurl", imgUrl)
			continue
		}
		logger.Info("vulnerability report stored successfully", "imageurl", imgUrl)
	}
	logger.Info("returning from new kubernetes release discovery", "config", secSchd.Cfg)
}
