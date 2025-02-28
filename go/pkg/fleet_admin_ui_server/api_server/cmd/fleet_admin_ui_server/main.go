// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	_ "github.com/amacneil/dbmate/pkg/driver/postgres"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin_ui_server/api_server/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_admin_ui_server/api_server/server"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()

	// Parse command line.
	var configFile string
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	log.BindFlags()
	flag.Parse()

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	err := func() error {
		// Load configuration from file.
		log.Info("main", logkeys.ConfigFile, configFile)
		var cfg config.Config
		if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
			return err
		}
		log.Info("main", logkeys.Configuration, cfg)

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		// Load database configuration.
		managedDb, err := manageddb.New(ctx, &cfg.Database)
		if err != nil {
			return err
		}

		// Start GRPC server.
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
		if err != nil {
			return fmt.Errorf("unable to listen on port %d: %w", cfg.ListenPort, err)
		}
		grpcService, err := server.New(ctx, &cfg, managedDb, listener)
		if err != nil {
			return err
		}
		return grpcService.Run(ctx)
	}()
	if err != nil {
		log.Error(err, "fatal error")
		os.Exit(1)
	}
}
