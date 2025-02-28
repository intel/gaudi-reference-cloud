// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package api

import (
	"context"
	"fmt"
	"os"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/apiservice/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/dcim"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/baremetal_enrollment/secrets"
)

type Client struct {
	vault  secrets.SecretManager
	netBox dcim.DCIM
}

func (c Client) GetK8SClientset(ctx context.Context, region string, availabilityZone string, objects ...runtime.Object) (
	kubernetes.Interface, error) {

	// Get AZ kubeconfig from Vault
	secretsPath := fmt.Sprintf("%s/baremetal/enrollment/%s", region, availabilityZone)
	secret, err := c.vault.GetControlPlaneSecrets(ctx, secretsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve secrets from vault: %v", err)
	}

	// get kubeconfig from vault
	kubeconfig, ok := secret.Data["kubeconfig"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"value type assertion failed: %T %#v", secret.Data["kubeconfig"], secret.Data["kubeconfig"])
	}

	// create RESTconfig from kubeconfig
	clusterConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, err
	}

	// creates the K8SClientset
	clientSet, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

type Region struct {
	netBox dcim.DCIM
}

func (r Region) GetClusterRegion(deviceInfo config.BMaaSEnrollmentData) (string, error) {
	region, err := r.netBox.GetDeviceRegionName(context.TODO(), deviceInfo.AvailabilityZone)
	if err != nil {
		return "", fmt.Errorf("failed to get region from netbox with error: %v", err)
	}
	return region, nil
}

func NewRouter(ctx context.Context, jobConfig config.EnrollmentJobConfig) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("unable to initialize logger: %v", err)
	}
	utc := true
	router.Use(ginzap.Ginzap(logger, time.RFC3339, utc))

	stack := true
	router.Use(ginzap.RecoveryWithZap(logger, stack))

	// set VAULT_ADDR
	if jobConfig.VaultAddress != "" {
		if err = os.Setenv(secrets.VaultAddressEnvVar, jobConfig.VaultAddress); err != nil {
			return nil, fmt.Errorf("unable to set vault addr env variable: %v", err)
		}

	}
	// set NETBOX_HOST
	if jobConfig.NetboxAddress != "" {
		if err = os.Setenv(dcim.NetBoxAddressEnvVar, jobConfig.NetboxAddress); err != nil {
			return nil, fmt.Errorf("unable to set netbox  addr env variable: %v", err)
		}

	}
	// connect to Vault to get the AZ kubeconfig
	vault, err := secrets.NewVaultClient(ctx,
		secrets.VaultOptionValidateClient(true),
		secrets.VaultOptionRenewToken(true),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize Vault client: %v", err)
	}

	secretPath := fmt.Sprintf("%s/baremetal/enrollment/netbox", jobConfig.Region)
	token, err := vault.GetNetBoxAPIToken(ctx, secretPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get NetBox API token: %v", err)
	}
	if !jobConfig.NetboxSkipTlsVerify {
		if err = os.Setenv(dcim.InsecureSkipVerifyEnvVar, "false"); err != nil {
			return nil, fmt.Errorf("unable to set InsecureSkipVerify env variable: %v", err)
		}
	}
	netbox, err := dcim.NewNetBoxClient(ctx, token, false)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize NetBox client: %v", err)
	}

	client := Client{vault: vault, netBox: netbox}
	region := Region{netBox: netbox}
	enrollment, err := NewBMaaSEnrollment(client, jobConfig, region)
	if err != nil {
		return nil, err
	}
	// get enrollment basic auth secrets from vault
	apiAuthSecretPath := fmt.Sprintf("%s/baremetal/enrollment/apiservice", jobConfig.Region)
	username, password, err := vault.GetEnrollBasicAuth(ctx, apiAuthSecretPath, true)
	if err != nil {
		return nil, err
	}
	if err := enrollment.AddRoutes(router, username, password); err != nil {
		return nil, err
	}

	return router, nil
}
