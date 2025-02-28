// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_replicator/replicator"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
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
	log := log.FromContext(ctx).WithName("main")

	err := func() error {
		// Load configuration from file.
		log.Info("Configuration file", logkeys.ConfigFile, configFile)

		scheme := runtime.NewScheme()
		if err := clientgoscheme.AddToScheme(scheme); err != nil {
			return err
		}
		if err := privatecloudv1alpha1.AddToScheme(scheme); err != nil {
			return err
		}

		cfg := privatecloudv1alpha1.InstanceReplicatorConfig{}
		options := ctrl.Options{
			Scheme: scheme,
		}
		if configFile != "" {
			var err error
			options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&cfg))
			if err != nil {
				return fmt.Errorf("unable to load the config file: %w", err)
			}
		}

		log.Info("Configuration", logkeys.Configuration, cfg)

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		k8sManager, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
		if err != nil {
			return fmt.Errorf("unable to start manager: %w", err)
		}

		if err := k8sManager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up ready check: %w", err)
		}

		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		clientConn, err := grpcutil.NewClient(ctx, cfg.ComputeApiServerAddr, dialOptions...)
		if err != nil {
			return err
		}
		instanceClient := pb.NewInstancePrivateServiceClient(clientConn)

		// Ensure that we can ping the instance service before starting the manager.
		pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
		defer cancel()
		if _, err := instanceClient.PingPrivate(pingCtx, &emptypb.Empty{}); err != nil {
			return fmt.Errorf("unable to ping instance service: %w", err)
		}

		_, err = replicator.NewReplicator(ctx, k8sManager, instanceClient)
		if err != nil {
			return fmt.Errorf("error creating replicator: %w", err)
		}

		log.Info("Starting manager")
		if err := k8sManager.Start(ctrl.SetupSignalHandler()); err != nil {
			return fmt.Errorf("problem running manager: %w", err)
		}
		return nil
	}()
	if err != nil {
		log.Error(err, "error")
		os.Exit(1)
	}
}
