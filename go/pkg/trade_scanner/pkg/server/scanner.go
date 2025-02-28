// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package server

import (
	"context"
	"fmt"
	"os"

	idcCommon "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/billing_common"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	obs "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/trade_scanner/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/trade_scanner/pkg/controller"
	tradecheck "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/tradecheck/tradecheckintel"
)

type TradeScannerService struct {
}

func (svc *TradeScannerService) Init(ctx context.Context, cfg *config.Config) error {
	ctx, logger, span := obs.LogAndSpanFromContextOrGlobal(ctx).WithName("TradeScannerService.Init").Start()
	defer span.End()
	logger.V(9).Info("BEGIN")
	defer logger.V(9).Info("END")

	cloudaccountClient, err := idcCommon.NewCloudAccountClient(ctx, &grpcutil.DnsResolver{})
	if err != nil {
		logger.Error(err, "failed to initialize cloudaccount client")
		return err
	}

	usernamefile, exists := os.LookupEnv("usernameFile")
	if !exists {
		err := fmt.Errorf("usernamefile name not found")
		logger.Error(err, "environment variable not found", "variable", "usernameFile")
		return err
	}
	passwordfile, exists := os.LookupEnv("passwordFile")
	if !exists {
		err := fmt.Errorf("passwordfile name not found")
		logger.Error(err, "environment variable not found", "variable", "passwordFile")
		return err
	}
	token_url, exists := os.LookupEnv("gts_get_token_url")
	if !exists {
		err := fmt.Errorf("gts get token URL not found")
		logger.Error(err, "environment variable not found", "variable", "gts_get_token_url")
		return err
	}
	screen_url, exists := os.LookupEnv("gts_business_screen_url")
	if !exists {
		err := fmt.Errorf("gts business screening url not found")
		logger.Error(err, "environment variable not found", "variable", "gts_business_screen_url")
		return err
	}

	config, err := tradecheck.CreateConfig(usernamefile, passwordfile, token_url, "", "", screen_url)
	if err != nil {
		logger.Error(err, "failed to create gts client config")
		return err
	}
	gtsClient, err := tradecheck.NewClient(config)
	if err != nil {
		logger.Error(err, "failed to initialize gts client")
		return err
	}

	tradescanSchd, err := controller.NewTradeComplianceScheduler(cloudaccountClient, gtsClient, cfg)
	if err != nil {
		logger.Error(err, "error starting trade compliance scan scheduler")
		return err
	}
	tradescanSchd.StartComlianceScanScheduler(ctx)
	return nil
}

func (svc *TradeScannerService) Name() string {
	return "idc-trade-scanner"
}
