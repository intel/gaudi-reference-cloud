// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"

	scannerConfig "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/security-scanner/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/insights/security-scanner/pkg/controller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
)

type SecurityScanService struct {
}

func (svc *SecurityScanService) Init(ctx context.Context, cfg *scannerConfig.Config) error {
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	log.Info("initializing IDC kube security scan service...")

	insightsClient, err := controller.NewInsightClient(ctx, cfg.SecurityInsights.URL)
	if err != nil {
		log.Error(err, "failed to initialize insights client")
		return err
	}

	scanSched, err := controller.NewSecurityScanScheduler(insightsClient, cfg)
	if err != nil {
		log.Error(err, "error starting kube security scan scheduler")
		return err
	}
	scanSched.StartSecurityScanScheduler(ctx)
	return nil
}

func (svc *SecurityScanService) Name() string {
	return "kube-security-scan-scheduler"
}
