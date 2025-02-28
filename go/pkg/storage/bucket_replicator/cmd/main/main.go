package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/bucket_replicator/pkg/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/storage/bucket_replicator/pkg/server"
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
	logger.Info("storage bucket replicator service starts")

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	bucketReplicator := server.BucketReplicatorService{}
	cfg := config.NewDefaultConfig()
	if configFile == "" {
		logger.Error(fmt.Errorf("config file not provided in config"), logkeys.Error)
		// Keep going. The caller may decide to provide reasonable defaults.
	}
	if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
		logger.Error(err, "error loading config")
		// Keep going. The caller may decide to provide reasonable defaults.
	}
	logger.Info("Configuration", logkeys.Configuration, cfg)

	err := bucketReplicator.Init(ctx, cfg)
	if err != nil {
		logger.Error(err, "error intializing replicator")
	}
}
