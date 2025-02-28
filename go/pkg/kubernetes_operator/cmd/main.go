// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/internal/controller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/spf13/viper"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(privatecloudv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	ctx := context.Background()

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var operatorConfig string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&operatorConfig, "operator-config", "/etc/kubernetes-operator", "This is the path to the operator config.")
	opts := zap.Options{
		Development:     false,
		StacktraceLevel: zapcore.DPanicLevel,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Initialize tracing.
	obs := observability.New(ctx)
	tracerProvider := obs.InitTracer(ctx)
	defer tracerProvider.Shutdown(ctx)

	// Read operator config.
	viper.AddConfigPath(operatorConfig)
	if err := viper.ReadInConfig(); err != nil {
		setupLog.Error(err, "read operator config")
		os.Exit(1)
	}

	var config controller.Config
	if err := viper.Unmarshal(&config); err != nil {
		setupLog.Error(err, "unmarshal operator config")
		os.Exit(1)
	}

	config.ConvertMonitoringToSystemMetrics()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "4496faa7.intel.com",
		// We currently rely on the status of the nodegroup to store the list of nodes
		// available in the nodegroup. So we disable the cache for nodegroups to avoid
		// a stale cache and miss a node created or deleted from the list.
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	idcGrpcClient, err := newIdcGrpcClient(ctx, config)
	if err != nil {
		setupLog.Error(err, "unable to create idc grpc client")
		os.Exit(1)
	}

	computeGrpcClient, err := newIdcGrpcClient(ctx, config)
	if err != nil {
		setupLog.Error(err, "unable to create compute grpc client")
		os.Exit(1)
	}

	if err = (&controller.ClusterReconciler{
		Client:                  mgr.GetClient(),
		Scheme:                  mgr.GetScheme(),
		Config:                  &config,
		FsOrgPrivateClient:      pb.NewFilesystemOrgPrivateServiceClient(idcGrpcClient),
		FsPrivateClient:         pb.NewFilesystemPrivateServiceClient(idcGrpcClient),
		StorageKMSPrivateClient: pb.NewStorageKMSPrivateServiceClient(idcGrpcClient),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster")
		os.Exit(1)
	}
	if err = (&controller.NodegroupReconciler{
		Client:                                mgr.GetClient(),
		Scheme:                                mgr.GetScheme(),
		Config:                                &config,
		NoCacheClientReader:                   mgr.GetAPIReader(),
		FilesystemStorageClusterPrivateClient: pb.NewFilesystemStorageClusterPrivateServiceClient(idcGrpcClient),
		InstanceTypeClient:                    pb.NewInstanceTypeServiceClient(computeGrpcClient),
		MachineImageClient:                    pb.NewMachineImageServiceClient(computeGrpcClient),
		InstanceServiceClient:                 pb.NewInstanceServiceClient(computeGrpcClient),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Nodegroup")
		os.Exit(1)
	}
	if err = (&controller.AddonReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Config: &config,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Addon")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func newComputeGrpcClient(ctx context.Context, config controller.Config) (*grpc.ClientConn, error) {
	creds, err := grpcutil.GetClientCredentials(ctx)
	if err != nil {
		return nil, err
	}

	clientOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

	client, err := grpcutil.NewClient(ctx, config.NodeProviders.Compute.URL,
		clientOptions...)
	if err != nil {
		setupLog.Error(err, "unable to create compute grpc client")
		os.Exit(1)
	}

	return client, nil
}

func newIdcGrpcClient(ctx context.Context, config controller.Config) (*grpc.ClientConn, error) {
	creds, err := grpcutil.GetClientCredentials(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "Get grpc client credentials")
	}

	clientOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	}

	client, err := grpcutil.NewClient(ctx, config.IdcGrpcUrl, clientOptions...)
	if err != nil {
		return nil, errors.Wrapf(err, "Create idc grpc client")
	}

	return client, nil
}
