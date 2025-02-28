// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"

	worker "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/cloudcredits_worker/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
)

func main() {
	ctx := context.Background()

	// Initialize logger
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("CloudCreditsWorker main")
	logger.Info("cloudcreditworker starts")

	// Initialize tracing
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	cfg := config.NewDefaultConfig()
	if err := loadConfig(ctx, cfg); err != nil {
		logger.Error(err, "error loading config", "err", err)
	}
	logger.Info("Run", "cfg", cfg)

	workerIns := &worker.Worker{}
	if err := workerIns.Init(ctx, cfg, &grpcutil.DnsResolver{}); err != nil {
		logger.Error(err, "init err")
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
