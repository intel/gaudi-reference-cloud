// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true

// BmInstanceOperatorConfig is the Schema for the bminstanceoperatorconfigs API.
// It stores the configuration for the Instance Operator.
type BmInstanceOperatorConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ControllerManagerConfigurationSpec returns the configurations for controllers
	cfg.ControllerManagerConfigurationSpec `json:",inline"`

	// Configuration shared with BmInstanceOperatorConfig and VmInstanceOperatorConfig.
	InstanceOperator InstanceOperatorConfig `json:"instanceOperator"`

	// OS http server url for BM machine images
	OsHttpServerUrl string `json:"osHttpServerUrl"`

	// Network Configuration
	NetworkConfig NetworkConfiguration `json:"networkConfig"`

	// SSH configuration for this operator to connect to instances to check instance health.
	SshConfig SshConfiguration `json:"sshConfig"`

	// Faceless Cloud Account ID that will be used by operator.
	CloudAccountID string `json:"cloudAccountID"`

	//Max concurrent reconcile value of BM instance operator
	MaxConcurrentReconciles int `json:"maxConcurrentReconciles"`

	// Validation task repository base url
	ValidationTaskRepositoryURL string `json:"validationTaskRepositoryURL"`
	// Version of the validation task to be used for every InstanceType
	ValidationTaskVersion ValidationTaskVersion `json:"validationTaskVersion"`

	// Instances types enabled for validation
	EnabledInstanceTypes []string `json:"enabledInstanceTypes"`

	// Configuration details about the enviornment where the operator is running
	EnvConfiguration EnvironmentConfiguration `json:"envConfiguration"`

	// Enable SDN support for accelerator network in BM instance operator when set to true
	AcceleratorNetworkSdnEnabled bool `json:"acceleratorNetworkSdnEnabled"`

	// S3 Configuration where the validation logs/reports are stored
	ValidationReportS3Config ValidationReportS3Configuration `json:"validationReportS3Config"`

	// Feature flags used by Validation operator
	FeatureFlags FeatureFlags `json:"featureFlags"`

	// Storage server addresses
	StorageServerSubnets []string `json:"storageServerSubnets"`
}
type ValidationTaskVersion struct {
	// Validation task artifact version for clusters. It is a map of instanceType to the version to be used.
	ClusterVersionMap map[string]string `json:"clusterVersionMap"`
	// Validation task artifact version for plain instances. It is a map of instanceType to the version to be used.
	InstanceVersionMap map[string]string `json:"instanceVersionMap"`
}

type EnvironmentConfiguration struct {
	// Region details.
	Region string `json:"region"`
	// Availability Zone
	AvailabilityZone string `json:"availabilityZone"`
	// Subnet Prefix Length for the current environment
	SubnetPrefixLength int `json:"subnetPrefixLength"`
	// Netbox address
	NetboxAddress string `json:"netboxAddress"`
	// Netbox key file path
	NetboxKeyFilePath string `json:"netboxKeyFilePath"`
	// Huggingface token file path
	HuggingFaceTokenFilePath string `json:"huggingFaceTokenFilePath"`
}

type ValidationReportS3Configuration struct {
	// S3 bucket where the validation reports are stored
	BucketName string `json:"bucketName"`
	// S3 access key file path
	S3AccessKeyFilePath string `json:"accessKeyFilePath"`
	// S3 secret access key file path
	S3SecretAccessKeyFilePath string `json:"secretAccessKeyFilePath"`
	// Proxy for pushing reports
	HttpsProxy string `json:"httpsProxy"`
	// Cloudfront prefix
	CloudfrontPrefix string `json:"cloudfrontPrefix"`
}

type FeatureFlags struct {
	// Enable deletion of instance if validation fails.
	DeProvisionPostValidationFailure bool `json:"deProvisionPostValidationFailure"`
	// Enable of disable group validation.
	GroupValidation bool `json:"groupValidation"`
	// Instances types enabled for group-validation
	EnabledGroupInstanceTypes []string `json:"enabledGroupInstanceTypes"`
	// Enable firmware upgrade on BMHs
	EnableFirmwareUpgrade bool `json:"enableFirmwareUpgrade"`
}

type NetworkBackEnd string

const (
	// NetworkBackEndNone indicates no network backend will be used. Mostly used for testing purpose.
	NetworkBackEndNone NetworkBackEnd = "none"
	// NetworkBackEndRaven specifies Raven will be used as network backend (i.e, for Vlan update)
	NetworkBackEndRaven NetworkBackEnd = "raven"
	// NetworkBackEndSDN specifies SDN-Controller will be used as network backend.
	NetworkBackEndSDN NetworkBackEnd = "sdn"
	// NetworkBackEndTransition performs vlan update against BOTH Raven and SDN (which should be in readonly mode to avoid updating the switch from 2 places)
	NetworkBackEndTransition NetworkBackEnd = "transition"
)

type NetworkConfiguration struct {
	NetworkBackEndType              NetworkBackEnd `json:"networkBackEndType"`
	NetworkKubeConfig               string         `json:"networkKubeConfig"`
	ProvisioningVlan                int            `json:"provisioningVlan"`
	AcceleratorNetworkDefaultVlan   int            `json:"acceleratorNetworkDefaultVlan"`
	StorageDefaultVlan              int            `json:"storageDefaultVlan"`
	AccBGPNetworkDefaultCommunityID int            `json:"accBGPNetworkDefaultCommunityID"`
}

type SshConfiguration struct {
	// Whether to wait for the SSH access of the BM instance to be ready
	WaitForSSHAccess bool `json:"waitForSSHAccess"`

	// Path to the Private key used for connecting to the SSH proxy server
	PrivateKeyFilePath string `json:"privateKeyFilePath"`

	// Path to the Public key corresponding to the private key
	PublicKeyFilePath string `json:"publicKeyFilePath"`

	// The username for authenticating with the SSH proxy server.
	SshProxyUser string `json:"sshProxyUser"`
	// If blank, use address in instance.Status.SshProxy.ProxyAddress.
	SshProxyAddress string `json:"sshProxyAddress"`
	// If 0, use address in instance.Status.SshProxy.ProxyPort.
	SshProxyPort int `json:"sshProxyPort"`

	// Path to the ssh proxy server host public key used for verifying the SSH proxy server
	HostPublicKeyFilePath string `json:"hostPublicKeyFilePath"`
}

func init() {
	SchemeBuilder.Register(&BmInstanceOperatorConfig{})
}
