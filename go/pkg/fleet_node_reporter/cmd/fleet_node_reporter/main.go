// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/conf"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_node_reporter/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/fleet_node_reporter/fleet_node_reporter"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	// Create a context that can be cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse command line arguments.
	var configFile string
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	log.BindFlags()
	flag.Parse()

	// Initialize the default logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx)

	// Function to handle the main logic with error handling.
	err := func() error {
		// Load configuration from the specified file.
		log.Info("main", "configFile", configFile)
		var cfg config.Config
		if err := conf.LoadConfigFile(ctx, configFile, &cfg); err != nil {
			return err
		}
		log.Info("main", "cfg", cfg, "PollingIntervalSeconds", cfg.SchedulerStatisticsPollingInterval.Seconds())

		// Initialize observability (tracing).
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		// Get gRPC client credentials.
		creds, err := grpcutil.GetClientCredentials(ctx)
		if err != nil {
			// logging due to Coverity issue
			log.Error(err, "unable to create gRPC credentials")
			return fmt.Errorf("unable to create gRPC credentials: %w", err)
		}

		// Create gRPC client connection options with tracing interceptors.
		clientOptions := []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		// Establish a new gRPC client connection for instanceScheduler.
		instanceSchedulerClientConn, err := grpcutil.NewClient(ctx, cfg.InstanceSchedulerAddr, clientOptions...)
		if err != nil {
			return err
		}
		defer instanceSchedulerClientConn.Close()

		// Create an instance of the scheduling service client.
		instanceSchedulingServiceClient := pb.NewInstanceSchedulingServiceClient(instanceSchedulerClientConn)

		// Ping the VM Instance Scheduling service to ensure it's reachable.
		// If this fails, the service will fail.
		pingCtxInstanceScheduler, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := instanceSchedulingServiceClient.Ping(pingCtxInstanceScheduler, &emptypb.Empty{}); err != nil {
			log.Error(err, "unable to ping VM Instance Scheduling service")
			return err
		}

		// Establish a new gRPC client connection for fleetAdminService.
		fleetAdminClientConn, err := grpcutil.NewClient(ctx, cfg.FleetAdminServerAddr, clientOptions...)
		if err != nil {
			return err
		}
		defer fleetAdminClientConn.Close()

		// Create an instance of the fleetAdminService client.
		fleetAdminServiceClient := pb.NewFleetAdminServiceClient(fleetAdminClientConn)

		// Ping the FleetAdminServer to ensure it's reachable.
		// If this fails, the service will fail.
		pingCtxFleetAdminServer, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := fleetAdminServiceClient.Ping(pingCtxFleetAdminServer, &emptypb.Empty{}); err != nil {
			log.Error(err, "unable to ping FleetAdmin Server")
			return err
		}

		// Create a new FleetNodeReporter instance.
		fleetNodeReporter, err := fleet_node_reporter.NewFleetNodeReporter(ctx, instanceSchedulingServiceClient, fleetAdminServiceClient, cfg)
		if err != nil {
			return err
		}

		// Start the FleetNodeReporter.
		fleetNodeReporter.Start(ctx)

		// Wait for the context to be cancelled.
		<-ctx.Done()
		return nil
	}()

	// Handle any errors that occurred during the main logic.
	if err != nil {
		log.Error(err, "error")
		os.Exit(1)
	}
}
