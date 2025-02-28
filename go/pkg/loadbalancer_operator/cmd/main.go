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
	"fmt"
	"os"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/processor"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/provider"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	firewallv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/firewall_operator/api/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	loadbalancerv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/loadbalancer_operator/internal/controller"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	k8scontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(loadbalancerv1alpha1.AddToScheme(scheme))
	utilruntime.Must(privatecloudv1alpha1.AddToScheme(scheme))
	utilruntime.Must(firewallv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	ctx := context.Background()

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var operatorConfig string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8078", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8079", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&operatorConfig, "operator-config", "/etc/loadbalancer-operator", "This is the path to the operator config.")
	opts := zap.Options{
		Development: true,
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

	var inputConfig controller.LoadbalancerProviderConfig
	if err := viper.Unmarshal(&inputConfig); err != nil {
		setupLog.Error(err, "unmarshal operator config")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "34a547c03.intel.com",
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

	// Create a new config object to allow reading the username
	config, err := config.NewConfiguration(inputConfig.UsernameFile, inputConfig.PasswordFile)
	if err != nil {
		setupLog.Error(err, "unable to initialize configuration")
		os.Exit(1)
	}

	// Initialize the Loadbalancer Provider API which is responsible for managing actual LBs via an API.
	providerAPI, err := provider.NewLoadbalancerProvider(inputConfig.ProviderType, &provider.Config{
		BaseURL:       inputConfig.BaseURL,
		Domain:        inputConfig.Domain,
		Configuration: config,
		Environment:   inputConfig.Environment,
		UserGroup:     inputConfig.UserGroup,
	})
	if err != nil {
		setupLog.Error(err, "unable to initialize load balancer provider")
		os.Exit(1)
	}
	setupLog.Info("The Provider is initialized", "Provider Name",
		inputConfig.ProviderType, "Provider URL", inputConfig.BaseURL)

	if inputConfig.RegionId == "" {
		setupLog.Error(fmt.Errorf("regionId is required"), "invalid config")
		os.Exit(1)
	}

	if inputConfig.AvailabilityZoneId == "" {
		setupLog.Error(fmt.Errorf("availabilityZoneId is required"), "invalid config")
		os.Exit(1)
	}
	appProcessor := processor.NewProcessor(mgr.GetClient(), providerAPI, mgr.GetScheme(), inputConfig.RegionId, inputConfig.AvailabilityZoneId)

	lbr := &controller.LoadbalancerReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		LBProvider: providerAPI,
		Processor:  appProcessor,
	}

	setupLog.Info("initializing controller", "MaxConcurrentReconciles", inputConfig.LoadbalancerMaxConcurrentReconciles)

	// Setup the Loadbalancer Reconciler
	err = ctrl.NewControllerManagedBy(mgr).
		For(&loadbalancerv1alpha1.Loadbalancer{},
			builder.WithPredicates(
				// Reconcile Loadbalancer if Loadbalancer spec changes or annotation changes.
				predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{}),
			),
		).
		Owns(&firewallv1alpha1.FirewallRule{}).
		Watches(
			&privatecloudv1alpha1.Instance{},
			handler.EnqueueRequestsFromMapFunc(lbr.MapInstanceToLoadbalancer)).
		WithOptions(k8scontroller.Options{
			MaxConcurrentReconciles: inputConfig.LoadbalancerMaxConcurrentReconciles,
		}).
		Complete(lbr)
	if err != nil {
		setupLog.Error(err, "unable to create load balancer controller")
		os.Exit(1)
	}

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
