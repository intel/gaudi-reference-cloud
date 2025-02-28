// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
/*
Copyright 2024.

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
	"strconv"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/hashicorp/go-multierror"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/ddi"
	controllers "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/operator/controllers"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/tasks"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/grpcutil"
	baremetalv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/metal3.io/v1alpha1"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(privatecloudv1alpha1.AddToScheme(scheme))
	utilruntime.Must(baremetalv1alpha1.AddToScheme(scheme))
}

func main() {
	ctx := context.Background()

	var configFile string
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&configFile, "config", "", "The application will load its configuration from this file.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	log.BindFlags()
	flag.Parse()

	// Initialize logger.
	log.SetDefaultLogger()
	log := log.FromContext(ctx).WithName("main")

	log.Info("Configuration file", "configFile", configFile)
	cfg := &privatecloudv1alpha1.BMEnrollmentOperatorConfig{}
	options := ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "81c21bc7.intel.com",
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
	}

	if configFile != "" {
		var err error
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(cfg))
		if err != nil {
			log.Error(err, "unable to load the config file")
			os.Exit(1)
		}
	}
	log.Info("Configuration", "cfg", cfg)

	log.Info("set bm enrollment operator environment variables")
	if err := setEnvVars(cfg); err != nil {
		log.Error(err, "failed to set bm enrollment operator environment variables")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	var vault secrets.SecretManager
	var netBox dcim.DCIM
	var menAndMice ddi.DDI
	var instancetypeClient pb.InstanceTypeServiceClient

	// retry for 5 minutes if a client is not available
	isRetryable := func(err error) bool {
		return true
	}
	backoff := wait.Backoff{
		Steps:    10,
		Duration: 10 * time.Millisecond,
		Factor:   4.0,
		Jitter:   0.1,
	}

	err = retry.OnError(backoff, isRetryable, func() error {
		// vault client
		vault, err = getVaultClient(ctx)
		if err != nil {
			return fmt.Errorf("unable to initialize Vault client: %v", err)
		}
		// netbox client
		netBox, err = getNetBoxClient(ctx, vault, cfg.Region)
		if err != nil {
			return fmt.Errorf("unable to initialize NetBox client: %v", err)
		}
		// ddi client
		menAndMice, err = getMenAndMiceClient(ctx, cfg.Region, vault)
		if err != nil {
			return fmt.Errorf("unable to initialize MenAndMice client: %v", err)
		}
		// instance type service client
		instancetypeClient, err = getInstanceTypeServiceClient(ctx)
		if err != nil {
			return fmt.Errorf("unable to initialize InstanceTypeServiceClient: %v", err)
		}
		return nil
	})

	if err != nil {
		log.Error(err, "failed to initialize BM enrollment clients")
		os.Exit(1)
	}

	if err = (&controllers.BMEnrollmentReconciler{
		Client:                    mgr.GetClient(),
		Scheme:                    mgr.GetScheme(),
		Cfg:                       cfg,
		DDI:                       menAndMice,
		InstanceTypeServiceClient: instancetypeClient,
		NetBox:                    netBox,
		Vault:                     vault,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create bm enrollment controller")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// set required environmental variables
func setEnvVars(cfg *privatecloudv1alpha1.BMEnrollmentOperatorConfig) error {
	var result *multierror.Error
	// Vault address
	err := os.Setenv(secrets.VaultAddressEnvVar, cfg.VaultAddress)
	if err != nil {
		result = multierror.Append(result, err)
	}
	// Netbox address
	err = os.Setenv(dcim.NetBoxAddressEnvVar, cfg.Netbox.Address)
	if err != nil {
		result = multierror.Append(result, err)
	}
	// ComputeApi server address
	err = os.Setenv(tasks.ComputeApiServerAddrEnvVar, cfg.ComputeApiServerAddress)
	if err != nil {
		result = multierror.Append(result, err)
	}
	// check skip tls verify for Netbox
	if !cfg.Netbox.SkipTlsVerify {
		err = os.Setenv(dcim.InsecureSkipVerifyEnvVar, strconv.FormatBool(cfg.Netbox.SkipTlsVerify))
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	// set dhcp proxy variable if present
	if cfg.DhcpProxy.Enabled {
		err = os.Setenv(tasks.DhcpProxyUrlEnvVar, cfg.DhcpProxy.URL)
		if err != nil {
			result = multierror.Append(result, err)
		}
	}
	// set men and mice variable if present
	if cfg.MenAndMice.Enabled {
		err = os.Setenv(tasks.MenAndMiceUrlEnvVar, cfg.MenAndMice.URL)
		if err != nil {
			result = multierror.Append(result, err)
		}
		err = os.Setenv(tasks.MenAndMiceServerAddressEnvVar, cfg.MenAndMice.ServerAddress)
		if err != nil {
			result = multierror.Append(result, err)
		}
		err = os.Setenv(tasks.TftpServerIPEnvVar, cfg.MenAndMice.TftpServerIp)
		if err != nil {
			result = multierror.Append(result, err)
		}

		// check skip tls verify for MenAndMice
		if !cfg.MenAndMice.InsecureSkipVerify {
			err = os.Setenv(ddi.InsecureSkipVerifyEnvVar, strconv.FormatBool(cfg.MenAndMice.InsecureSkipVerify))
			if err != nil {
				result = multierror.Append(result, err)
			}
		}
	}
	// enabled Bios master password
	if cfg.SetBiosPassword {
		err = os.Setenv(tasks.SetBiosPasswordEnvVar, strconv.FormatBool(cfg.SetBiosPassword))
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

// vault client
func getVaultClient(ctx context.Context) (secrets.SecretManager, error) {
	log := log.FromContext(ctx).WithName("getVaultClient")
	log.Info("Initializing Vault client")

	vault, err := secrets.NewVaultClient(ctx,
		secrets.VaultOptionRenewToken(true),
		secrets.VaultOptionValidateClient(true),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	log.Info("Successfully initialized Vault client")
	return vault, nil
}

// netbox client
func getNetBoxClient(ctx context.Context, accessor secrets.NetBoxSecretAccessor, region string) (dcim.DCIM, error) {
	log := log.FromContext(ctx).WithName("getNetBoxClient")
	log.Info("Initializing NetBox client")

	if accessor == nil {
		return nil, fmt.Errorf("NetBox secret accessor has been initialized")
	}

	secretPath := fmt.Sprintf("%s/baremetal/enrollment/netbox", region)
	token, err := accessor.GetNetBoxAPIToken(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get NetBox API token: %v", err)
	}

	netBox, err := dcim.NewNetBoxClient(ctx, token, false)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize NetBox client: %v", err)
	}

	log.Info("Successfully initialized NetBox client")
	return netBox, nil
}

// ddi client
func getMenAndMiceClient(ctx context.Context, region string, vault secrets.DDISecretAccessor) (menAndMice ddi.DDI, err error) {
	log := log.FromContext(ctx).WithName("getMenAndMiceClient")
	log.Info("Initializing MenAndMice client")

	menAndMiceUrl, exists := os.LookupEnv(controllers.MenAndMiceUrlEnvVar)
	if !exists {
		log.Info("MenAndMice URL is not found in environment")
		return nil, nil
	}

	menAndMiceServerAddress, exists := os.LookupEnv(controllers.MenAndMiceServerAddressEnvVar)
	if !exists {
		log.Info("MenAndMice server address is not found in environment")
		return nil, nil
	}

	secretPath := fmt.Sprintf("%s/baremetal/enrollment/menandmice", region)
	ddiUsername, ddiPassword, err := vault.GetDDICredentials(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get Men and Mice Credentails:  %v", err)
	}

	menAndMice, err = ddi.NewMenAndMice(ctx, ddiUsername, ddiPassword, menAndMiceUrl, menAndMiceServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create men and mice object:  %v", err)
	}

	log.Info("Successfully initialized MenAndMice client")
	return menAndMice, nil
}

// instance type service client
func getInstanceTypeServiceClient(ctx context.Context) (pb.InstanceTypeServiceClient, error) {
	log := log.FromContext(ctx).WithName("getInstanceTypeServiceClient")
	log.Info("Initializing InstanceTypeService Client")

	computeApiServerAddr := os.Getenv(controllers.ComputeApiServerAddrEnvVar)
	if computeApiServerAddr == "" {
		return nil, fmt.Errorf("failed to get the compute api server Address")
	}

	computeApiServerClientConn, err := grpcutil.NewClient(ctx, computeApiServerAddr)
	if err != nil {
		return nil, fmt.Errorf("computeApiServerClientConn is not getting init %v", err)
	}

	return pb.NewInstanceTypeServiceClient(computeApiServerClientConn), nil
}
