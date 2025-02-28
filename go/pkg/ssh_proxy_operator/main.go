// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	// _ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	_ "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil/grpclog"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	cloudclient "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/clientset/versioned"
	idcinformerfactory "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/generated/cloudintelclient/informers/externalversions"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	privatecloudcontrollers "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/ssh_proxy_operator/controllers/private.cloud"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(privatecloudv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	ctx := context.Background()

	var configFile string
	var metricsAddr string
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8082", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	log.BindFlags()
	flag.Parse()

	// Initialize logger.
	log.SetDefaultLogger()
	setupLog := log.FromContext(ctx).WithName("main")

	err := func() error {
		setupLog.Info("main", logkeys.ConfigFile, configFile)

		ctrlConfig := privatecloudv1alpha1.SshProxyOperatorConfig{}
		options := ctrl.Options{Scheme: scheme}
		if configFile != "" {
			var err error
			options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
			if err != nil {
				return fmt.Errorf("unable to load the config file: %w", err)
			}
		}

		setupLog.Info("main", logkeys.Configuration, ctrlConfig)
		setupLog.Info("main", logkeys.KeyPath, ctrlConfig.AuthorizedKeysFilePath)

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
		if err != nil {
			return fmt.Errorf("unable to start manager: %w", err)
		}

		publicKey, err := os.ReadFile(ctrlConfig.PublicKeyFilePath)
		if err != nil {
			return fmt.Errorf("missing public key file %v: %w", ctrlConfig.PublicKeyFilePath, err)
		}

		// Convert publcKey []byte to string
		publicKeyStr := string(publicKey)

		// Check if the public Key ends with a new line
		if !strings.HasSuffix(publicKeyStr, "\n") {
			// Append a new line if no new line in public key
			publicKeyStr += "\n"
		}

		privateKey, err := os.ReadFile(ctrlConfig.PrivateKeyFilePath)
		if err != nil {
			return fmt.Errorf("missing private key file %v: %w", ctrlConfig.PrivateKeyFilePath, err)
		}

		hostPublicKey, err := os.ReadFile(ctrlConfig.HostPublicKeyFilePath)
		if err != nil {
			return fmt.Errorf("missing host key file %v: %w", ctrlConfig.HostPublicKeyFilePath, err)
		}

		sshProxyConfig := privatecloudcontrollers.SshProxyTunnelConfig{
			AuthorizedKeysFilePath:   ctrlConfig.AuthorizedKeysFilePath,
			ProxyUser:                ctrlConfig.ProxyUser,
			ProxyAddress:             ctrlConfig.ProxyAddress,
			ProxyPort:                ctrlConfig.ProxyPort,
			AuthorizedKeysScpTargets: ctrlConfig.AuthorizedKeysScpTargets,
			PublicKey:                publicKeyStr,
			PrivateKey:               string(privateKey),
			HostPublicKey:            string(hostPublicKey),
		}

		// creating kubernetes clientset
		kubeClientSet, err := kubernetes.NewForConfig(mgr.GetConfig())
		if err != nil {
			return fmt.Errorf("error building kubernetes clientset: %w", err)
		}

		// create IDC clientset
		idcClientSet, err := cloudclient.NewForConfig(mgr.GetConfig())
		if err != nil {
			return fmt.Errorf("unable to build idc clientset: %w", err)
		}

		defaultResync := 10 * time.Minute

		// creating SharedInformers
		informerFactory := idcinformerfactory.NewSharedInformerFactory(idcClientSet, defaultResync)

		// Creating sshProxy controller
		sshProxyController, err := privatecloudcontrollers.NewSshProxyController(ctx, kubeClientSet, idcClientSet, informerFactory.Private().V1alpha1().SshProxyTunnels(), sshProxyConfig)
		if err != nil {
			return fmt.Errorf("error creating sshProxyController: %w", err)
		}

		stopCh := make(chan struct{})

		// Intializing all the requested informers
		informerFactory.Start(stopCh)

		// Adding the sshProxy controller to the manager
		err = mgr.Add(manager.RunnableFunc(func(context.Context) error {
			return sshProxyController.Run(ctx, stopCh)
		}))
		if err != nil {
			return fmt.Errorf("unable to add SSH proxy controller to manager: %w", err)
		}

		//+kubebuilder:scaffold:builder

		if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up health check: %w", err)
		}
		if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up ready check: %w", err)
		}

		setupLog.Info("Starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			return fmt.Errorf("problem running manager: %w", err)
		}
		return nil
	}()
	if err != nil {
		setupLog.Error(err, logkeys.Error)
		os.Exit(1)
	}
}
