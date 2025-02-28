// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/amacneil/dbmate/pkg/driver/postgres"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/productcatalog/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	"k8s.io/client-go/rest"
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
		log.Info("main", "configFile", configFile)
		cfg := config.Config{}
		err := conf.LoadConfigFile(ctx, configFile, &cfg)
		if err != nil {
			return err
		}
		log.Info("main", "cfg", cfg)

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		// Load Kubernetes client configuration
		var defaultKubeRestConfig *rest.Config
		if defaultKubeRestConfig, err = conf.GetKubeRestConfig(); err != nil {
			defaultKubeRestConfig = &rest.Config{}
		}
		log.V(9).Info("main", "defaultKubeRestConfig", defaultKubeRestConfig)

		// Load cloudaccount database configuration.
		cloudAccountDb, err := manageddb.New(ctx, &cfg.CloudAccountDatabase)
		if err != nil {
			return err
		}

		// Load productcatalog database configuration - iff ProductCatalogDatabase.URL defined.
		var productCatalogDb *manageddb.ManagedDb
		if cfg.ProductCatalogDatabase.URL != "" {
			productCatalogDb, err = manageddb.New(ctx, &cfg.ProductCatalogDatabase)
			if err != nil {
				return err
			}
		}

		// Start GRPC server.
		grpcService := productcatalog.GrpcService{
			ListenAddr:       fmt.Sprintf(":%d", cfg.ListenPort),
			Config:           &cfg,
			CloudAccountDb:   cloudAccountDb,
			ProductCatalogDb: productCatalogDb,
		}
		if err := grpcService.Start(ctx, defaultKubeRestConfig); err != nil {
			return err
		}

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan

		return nil
	}()
	if err != nil {
		log.Error(err, "fatal error")
		os.Exit(1)
	}
}
