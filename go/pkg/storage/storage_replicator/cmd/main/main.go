// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_replicator/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/storage_replicator/pkg/server"
)

func main() {
	ctx := context.Background()
	configFile := ""
	flag.StringVar(&configFile, "config", "", "config file")
	log.BindFlags()
	flag.Parse()
	// Initialize logger.
	log.SetDefaultLogger()
	logger := log.FromContext(ctx).WithName("main")
	logger.Info("storage replicator service starts")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	filesystemReplicator := server.StorageReplicatorService{}
	cfg := config.NewDefaultConfig()
	if configFile == "" {
		logger.Error(fmt.Errorf("unable to read configuration"), "configuration file is not provided")
		os.Exit(1)
	}
	if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
		logger.Error(err, "error loading config file")
		os.Exit(1)
	}
	logger.Info("Configuration", logkeys.Configuration, cfg)

	err := filesystemReplicator.Init(ctx, cfg)
	if err != nil {
		logger.Error(err, "error intializing replicator")
	}
}
