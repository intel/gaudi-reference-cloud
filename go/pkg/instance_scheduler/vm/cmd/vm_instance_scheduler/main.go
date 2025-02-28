// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_scheduler/vm/server"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

func main() {
	ctx := context.Background()

	var configFile string
	var metricsAddr string
	var probeAddr string
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8082", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	log.BindFlags()
	klog.InitFlags(nil)
	flag.Parse()

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("main")

	err := func() error {
		log.Info("Configuration file", logkeys.ConfigFile, configFile)

		scheme := runtime.NewScheme()
		if err := clientgoscheme.AddToScheme(scheme); err != nil {
			return err
		}
		if err := privatecloudv1alpha1.AddToScheme(scheme); err != nil {
			return err
		}

		cfg := &privatecloudv1alpha1.VmInstanceSchedulerConfig{}
		options := ctrl.Options{
			Scheme: scheme,
		}
		if configFile != "" {
			var err error
			options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(cfg))
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

		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		computeApiServerClientConn, err := grpcutil.NewClient(ctx, cfg.ComputeApiServerAddr, dialOptions...)
		if err != nil {
			return err
		}
		// Connect to InstanceType Service client to Request InstanceTypes.
		// TODO: update the service to fail if it can't talk to compute API server
		instanceTypeServiceClient := pb.NewInstanceTypeServiceClient(computeApiServerClientConn)

		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
		if err != nil {
			return fmt.Errorf("unable to listen on port %d: %w", cfg.ListenPort, err)
		}

		_, err = server.NewSchedulingServer(ctx, cfg, k8sManager, listener, instanceTypeServiceClient)
		if err != nil {
			return fmt.Errorf("error creating scheduling server: %w", err)
		}

		if err := k8sManager.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up health check: %w", err)
		}
		if err := k8sManager.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up ready check: %w", err)
		}

		log.Info("Starting manager")
		// Manager will start scheduler and GRPC service.
		if err := k8sManager.Start(ctrl.SetupSignalHandler()); err != nil {
			return fmt.Errorf("problem running manager: %w", err)
		}
		return nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error)
		os.Exit(1)
	}
}
