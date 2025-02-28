// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package config

import (
	"time"

	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/compute_api_server/vnet"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/manageddb"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Application configuration
type Config struct {
	ListenPort uint16           `koanf:"listenPort"`
	Database   manageddb.Config `koanf:"database"`
	Region     string           `koanf:"region"`
	// Format should be "vm-instance-scheduler:80"
	VmInstanceSchedulerAddr string                 `koanf:"vmInstanceSchedulerAddr"`
	BillingServerAddr       string                 `koanf:"billingServerAddr"`
	Nameservers             []string               `koanf:"nameservers"`
	VNetService             vnet.VNetServiceConfig `koanf:"vNetService"`
	// Interval of time between attempts to purge all instances.
	PurgeInstanceInterval time.Duration `koanf:"purgeInstanceInterval"`
	// Purge (permanently delete from database) instances marked for deletion more than this duration in the past.
	PurgeInstanceAge time.Duration `koanf:"purgeInstanceAge"`
	DomainSuffix     string        `koanf:"domainSuffix"`
	// Interval of time between attempts to obtain the instance types that needs to be deleted for each CloudAccountId.
	GetDeactivateInstancesInterval time.Duration        `koanf:"getDeactivateInstancesInterval"`
	DbMaxIdleConnectionCount       uint16               `koanf:"dbMaxIdleConnectionCount"`
	CloudaccountServerAddr         string               `koanf:"cloudaccountServerAddr"`
	CloudAccountQuota              CloudAccountQuota    `koanf:"cloudAccountQuota"`
	FeatureFlags                   FeatureFlags         `koanf:"featureFlags"`
	AcceleratorInterface           AcceleratorInterface `koanf:"acceleratorInterface"`
	StorageInterface               StorageInterface     `koanf:"storageInterface"`
	ObjectStoragePrivateServerAddr string               `koanf:"objectStoragePrivateServerAddr"`
	FleetAdminServerAddr           string               `koanf:"fleetAdminServerAddr"`
	QuotaManagementServerAddr      string               `koanf:"quotaManagementServerAddr"`
}

type FeatureFlags struct {
	EnableComputeNodePoolsForScheduling bool `koanf:"enableComputeNodePoolsForScheduling"`
	EnableMultipleFirmwareSupport       bool `koanf:"enableMultipleFirmwareSupport"`
	EnableQMSForQuotaProcessing         bool `koanf:"enableQMSForQuotaProcessing"`
}

type AcceleratorInterface struct {
	EnabledInstanceTypes []string `koanf:"enabledInstanceTypes"`
	EnableStaticBGP      bool     `koanf:"enableStaticBGP"`
}

type StorageInterface struct {
	Enabled              bool     `koanf:"enabled"`
	EnabledInstanceTypes []string `koanf:"enabledInstanceTypes"`
	EnabledOnSingles     bool     `koanf:"enabledOnSingles"`
}

type CloudAccountQuota struct {
	// CloudAccounts is keyed by pb.AccountType
	CloudAccounts map[string]LaunchQuota `koanf:"cloudAccounts"`

	// CloudAccountIDQuotas is keyed by the unique identifier cloudaccountid of the user.
	// If the CloudAccountID is not found in the configured map, then the default quota will apply.
	CloudAccountIDQuotas map[string]CloudAccountIDQuota `koanf:"cloudAccountIDQuotas"`

	// DefaultLoadBalancerQuota defines the default number of load balancers that can be created for a cloud account.
	// This value can be overridden by adding the cloud account id to the CloudAccountIDQuotas map and providing a different value.
	DefaultLoadBalancerQuota int `koanf:"defaultLoadbalancerQuota"`

	// DefaultLoadBalancerListenerQuota defines the default number of listeners that can be created for a load balancer.
	// This value can be overridden by adding the cloud account id to the CloudAccountIDQuotas map and providing a different value.
	DefaultLoadBalancerListenerQuota int `koanf:"defaultLoadBalancerListenerQuota"`

	// DefaultLoadBalancerSourceIPQuota defines the default number of source IPs that can be configured for a load balancer.
	// This value can be overridden by adding the cloud account id to the CloudAccountIDQuotas map and providing a different value.
	DefaultLoadBalancerSourceIPQuota int `koanf:"defaultLoadBalancerSourceIPQuota"`
}

type CloudAccountIDQuota struct {
	// Defines the max number of load balancers that this Cloud Account can create.
	LoadbalancerQuota int `yaml:"loadbalancerQuota"`
	// Defines the max number of listeners that this Cloud Account can create per load balancer.
	LoadbalancerListenerQuota int `yaml:"loadbalancerListenerQuota"`
	// Defines the max number of source IPs that this Cloud Account can configure per load balancer.
	LoadbalancerSourceIPQuota int `yaml:"loadbalancerSourceIPQuota"`
}

// TODO: Update structs based on the final design
type LaunchQuota struct {
	// InstanceQuota is keyed from the InstanceType
	InstanceQuota map[string]int `yaml:"instanceQuota"`
}

func GetKubeRestConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}

func (i *AcceleratorInterface) IsEnabledForInstanceType(instanceType string) bool {
	for _, t := range i.EnabledInstanceTypes {
		if t == instanceType {
			return true
		}
	}
	return false
}
