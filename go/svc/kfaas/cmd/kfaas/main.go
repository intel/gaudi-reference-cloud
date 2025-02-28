// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	// _ "github.com/amacneil/dbmate/pkg/driver/postgres"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kfaas/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kfaas/server"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()
	var configFile string
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	log.BindFlags()
	flag.Parse()
	// Parse command line.

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	err := func() error {
		var cfg config.Config
		if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
			return err
		}
		log.Info("main", "cfg", cfg)

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)
		// Load database configuration.
		managedDb, err := manageddb.New(ctx, &cfg.Database)

		if err != nil {
			return err
		}
		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		iksClientConn, err := grpcutil.NewClient(ctx, cfg.IKSServerAddr, dialOptions...)
		if err != nil {
			log.Error(err, "error", "Not able to connect to gRPC service using grpcutil.NewClient")
			return err
		}
		defer iksClientConn.Close()

		iksClient := pb.NewIksClient(iksClientConn)
		// Start GRPC server.
		grpcService, err := server.New(ctx, &cfg, managedDb, iksClient)
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
