// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/trade_scanner/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/trade_scanner/pkg/server"
)

func main() {
	ctx := context.Background()
	// Initialize logger.
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("gts trade Compliance scanner")
	logger.Info("trade compliance scanner scheduler starts")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	tradeScanSvc := server.TradeScannerService{}
	cfg := config.NewDefaultConfig()
	if err := loadConfig(ctx, cfg); err != nil {
		logger.Error(err, "error loading config", "err", err)
		// Keep going. The caller may decide to provide reasonable defaults.
	}
	logger.Info("Run", "cfg", cfg)

	if err := tradeScanSvc.Init(ctx, cfg); err != nil {
		logger.Error(err, "error initializing trade scanner scheduler")
	}
}

func loadConfig(ctx context.Context, cfg *config.Config) error {
	log.BindFlags()
	configFile := ""
	flag.StringVar(&configFile, "config", "", "config file")
	flag.Parse()
	if configFile == "" {
		return fmt.Errorf("config flag can't be an empty string")
	}
	if err := conf.LoadConfigFile(ctx, configFile, cfg); err != nil {
		return fmt.Errorf("error loading config file (%s): %w", configFile, err)
	}
	return nil
}
