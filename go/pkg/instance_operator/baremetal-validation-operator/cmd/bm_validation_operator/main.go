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
	controllers "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/instance_operator/baremetal-validation-operator/controllers/metal3.io"
	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/k8s/apis/private.cloud/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/observability"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/pb"
	baremetalv1alpha1 "github.com/metal3-io/baremetal-operator/apis/metal3.io/v1alpha1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	//+kubebuilder:scaffold:imports
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
		if err := baremetalv1alpha1.AddToScheme(scheme); err != nil {
			return err
		}
		if err := privatecloudv1alpha1.AddToScheme(scheme); err != nil {
			return err
		}

		cfg := &privatecloudv1alpha1.BmInstanceOperatorConfig{}
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
		if err := validateConfiguration(cfg); err != nil {
			return fmt.Errorf("invalid Configuration for validation operator: %w", err)
		}

		// Initialize tracing.
		obs := observability.New(ctx)
		tracerProvider := obs.InitTracer(ctx)
		defer tracerProvider.Shutdown(ctx)

		// Initialize connection to compute API server
		dialOptions := []grpc.DialOption{
			grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		}

		// Connect to compute api server.
		computeApiServerClientConn, err := grpcutil.NewClient(ctx, cfg.InstanceOperator.ComputeApiServerAddr, dialOptions...)
		if err != nil {
			return err
		}

		// SSH public key service
		sshKeyClient, err := createSSHKeyClient(ctx, computeApiServerClientConn)
		if err != nil {
			return err
		}

		// read public ssh key from the configured file.
		publicKey, err := controllers.GetSSHKeys(ctx, cfg.SshConfig.PublicKeyFilePath)
		if err != nil {
			return fmt.Errorf("unable to read public SSH Key file: %w", err)
		}
		_, err = sshKeyClient.Create(ctx, &pb.SshPublicKeyCreateRequest{
			Metadata: &pb.ResourceMetadataCreate{
				CloudAccountId: cfg.CloudAccountID,
				Name:           controllers.NAME,
			},
			Spec: &pb.SshPublicKeySpec{
				SshPublicKey: publicKey,
			},
		})
		if err != nil {
			if status.Code(err) == codes.AlreadyExists {
				log.Info("SSH public key already exists for this account", logkeys.CloudAccountId, cfg.CloudAccountID)
			} else {
				return fmt.Errorf("failed to create public Key: %w", err)
			}
		} else {
			log.Info("SSH public key created for account", logkeys.CloudAccountId, cfg.CloudAccountID)
		}

		// Ensure that we can ping the VNet service before starting the manager.
		// TODO: Allow the service to start without this ping. Use a health check to monitor this.
		vNetClient, err := createVnetClient(ctx, computeApiServerClientConn)
		if err != nil {
			return err
		}
		// Peform a GET and PUT due to bug in VNET API ref: TWC4727-823
		_, err = vNetClient.Get(ctx, &pb.VNetGetRequest{
			Metadata: &pb.VNetGetRequest_Metadata{
				CloudAccountId: cfg.CloudAccountID,
				NameOrId: &pb.VNetGetRequest_Metadata_Name{
					Name: fmt.Sprintf("%s-%s", cfg.EnvConfiguration.AvailabilityZone, controllers.NAME),
				},
			},
		})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				log.Info("Vnet does not exist for this account, create one", logkeys.CloudAccountId, cfg.CloudAccountID)
				vnet, err := vNetClient.Put(ctx, createVNetRequest(cfg))
				if err != nil {
					return fmt.Errorf("failed to create vNet: %w", err)
				} else {
					log.Info("VNet created", logkeys.VNetMetadata, vnet.Metadata, logkeys.CloudAccountId, cfg.CloudAccountID)
				}
			} else { // failed to get vnet
				return err
			}
		}

		// Ensure that we can ping the Compute private service before starting the manager.
		computeClient, err := createComputeApiClient(ctx, computeApiServerClientConn)
		if err != nil {
			return err
		}

		// Create a manager for creating k8s controllers
		mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
			Scheme:                 scheme,
			Metrics:                metricsserver.Options{BindAddress: metricsAddr},
			HealthProbeBindAddress: probeAddr,
			LeaderElection:         *cfg.LeaderElection.LeaderElect,
			LeaderElectionID:       cfg.LeaderElection.ResourceName,
		})
		if err != nil {
			return fmt.Errorf("unable to start manager: %w", err)
		}

		// Read private ssh key from the configured file.
		sshSigner, err := parseSshPrivateKey(ctx, cfg)
		if err != nil {
			return err
		}

		imageClient, err := createMachineImageClient(ctx, computeApiServerClientConn)
		if err != nil {
			return err
		}

		netBoxClient, err := createNetboxClient(ctx, cfg.EnvConfiguration.NetboxAddress, cfg.EnvConfiguration.NetboxKeyFilePath)
		if err != nil {
			return err
		}
		// Create Reconciler
		if err = (&controllers.BaremetalhostsReconciler{
			Client:               mgr.GetClient(),
			Scheme:               mgr.GetScheme(),
			Cfg:                  cfg,
			Signer:               &sshSigner,
			ComputePrivateClient: computeClient,
			ImageFinder:          controllers.NewImageFinder(imageClient, mgr.GetClient()),
			EventRecorder:        mgr.GetEventRecorderFor("validation-operator"),
			NetBoxClient:         netBoxClient,
		}).SetupWithManager(ctx, mgr); err != nil {
			return fmt.Errorf("unable to create controller for Baremetals hosts: %w", err)
		}

		if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up health check: %w", err)
		}
		if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			return fmt.Errorf("unable to set up ready check: %w", err)
		}

		log.Info("Starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			return fmt.Errorf("problem running manager: %w", err)
		}
		return nil
	}()
	if err != nil {
		log.Error(err, logkeys.Error)
		os.Exit(1)
	}

}

// Helper methods

func validateConfiguration(cfg *privatecloudv1alpha1.BmInstanceOperatorConfig) error {
	if cfg.CloudAccountID == "" {
		return fmt.Errorf("faceless CloudAccountID is not configured")
	}

	if cfg.InstanceOperator.ComputeApiServerAddr == "" {
		return fmt.Errorf("compute API service address is not configured")
	}

	if cfg.ValidationTaskRepositoryURL == "" {
		return fmt.Errorf("validation Task Repository URL is not configured")
	}

	if len(cfg.EnabledInstanceTypes) == 0 {
		return fmt.Errorf("validation is not enabled for any InstanceTypes")
	}

	if cfg.EnvConfiguration.AvailabilityZone == "" || cfg.EnvConfiguration.Region == "" ||
		cfg.EnvConfiguration.SubnetPrefixLength == 0 {
		return fmt.Errorf("environment configuration is not set")
	}

	if cfg.EnvConfiguration.NetboxAddress == "" {
		return fmt.Errorf("netbox configuration is not set")
	}
	return nil
}

func createComputeApiClient(ctx context.Context, computeApiServerClientConn *grpc.ClientConn) (pb.InstancePrivateServiceClient, error) {
	computeClient := pb.NewInstancePrivateServiceClient(computeApiServerClientConn)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := computeClient.PingPrivate(ctx, &emptypb.Empty{}); err != nil {
		return nil, fmt.Errorf("unable to ping Compute InstancePrivate service: %w", err)
	}
	return computeClient, nil
}
func createMachineImageClient(ctx context.Context, computeApiServerClientConn *grpc.ClientConn) (pb.MachineImageServiceClient, error) {
	imageClient := pb.NewMachineImageServiceClient(computeApiServerClientConn)
	return imageClient, nil
}

func createVnetClient(ctx context.Context, computeApiServerClientConn *grpc.ClientConn) (pb.VNetServiceClient, error) {
	vNetClient := pb.NewVNetServiceClient(computeApiServerClientConn)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := vNetClient.Ping(ctx, &emptypb.Empty{}); err != nil {
		return nil, fmt.Errorf("unable to ping VNet service: %w", err)
	}
	return vNetClient, nil
}

func createSSHKeyClient(ctx context.Context, computeApiServerClientConn *grpc.ClientConn) (pb.SshPublicKeyServiceClient, error) {
	sshKeyClient := pb.NewSshPublicKeyServiceClient(computeApiServerClientConn)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if _, err := sshKeyClient.Ping(ctx, &emptypb.Empty{}); err != nil {
		return nil, fmt.Errorf("unable to ping SSH public key service: %w", err)
	}
	return sshKeyClient, nil
}

func createNetboxClient(ctx context.Context, netboxAddress, tokenPath string) (dcim.DCIM, error) {
	os.Setenv("NETBOX_HOST", netboxAddress)
	// read netbox key
	netboxTokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Netbox key file %v", err)
	}
	netBox, err := dcim.NewNetBoxClient(ctx, string(netboxTokenBytes), true)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize NetBox client: %v", err)
	}
	return netBox, err
}

func createVNetRequest(cfg *privatecloudv1alpha1.BmInstanceOperatorConfig) *pb.VNetPutRequest {

	req := &pb.VNetPutRequest{
		Metadata: &pb.VNetPutRequest_Metadata{
			CloudAccountId: cfg.CloudAccountID,
			Name:           fmt.Sprintf("%s-%s", cfg.EnvConfiguration.AvailabilityZone, controllers.NAME),
		},
		Spec: &pb.VNetSpec{
			Region:           cfg.EnvConfiguration.Region,
			AvailabilityZone: cfg.EnvConfiguration.AvailabilityZone,
			PrefixLength:     int32(cfg.EnvConfiguration.SubnetPrefixLength),
		},
	}
	return req
}

func parseSshPrivateKey(ctx context.Context, cfg *privatecloudv1alpha1.BmInstanceOperatorConfig) (ssh.Signer, error) {
	log := log.FromContext(ctx).WithName("main.parseSshPrivateKey")
	privateKey, err := os.ReadFile(cfg.SshConfig.PrivateKeyFilePath)
	if err != nil {
		log.Error(err, "Failed to read Private key", logkeys.KeyPath, cfg.SshConfig.PrivateKeyFilePath)
		return nil, err
	}

	sshSigner, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		log.Error(err, "Failed to parse private key", logkeys.KeyPath, cfg.SshConfig.PrivateKeyFilePath)
		return nil, err
	}
	return sshSigner, nil
}
