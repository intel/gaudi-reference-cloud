// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package types

import (
	"context"
	"fmt"
)

type MultipleEnvConfig struct {
	Environments map[string]*EnvConfigWithUnparsed
}

// Type for parsing the output of HelmfileEnvConfig.CreateHelmfileDump() which writes helmfile-dump.yaml.
// Currently, this type only represents a subset of the contents of this file.
// Over time, this type (and dependent types) can be extended to make more of this information
// available to Go applications.
type EnvConfig struct {
	Environment EnvConfigEnvironment `yaml:"Environment"`
	Values      Values               `yaml:"Values"`
}

// Extends EnvConfig with the raw (unparsed) output of HelmfileEnvConfig.CreateHelmfileDump().
type EnvConfigWithUnparsed struct {
	EnvConfig
	// Contents of file created by HelmfileEnvConfig.CreateHelmfileDump().
	HelmfileDumpYamlBytes UnparsedBytes
}

func NewEnvConfigWithUnparsed(envConfig EnvConfig, helmfileDumpYamlBytes []byte) *EnvConfigWithUnparsed {
	return &EnvConfigWithUnparsed{
		EnvConfig:             envConfig,
		HelmfileDumpYamlBytes: helmfileDumpYamlBytes,
	}
}

type EnvConfigEnvironment struct {
	// Equal to IDC_ENV.
	Name string `yaml:"Name"`
}

type Values struct {
	ArgoCd            ArgoCd            `yaml:"argocd"`
	IdcHelmRepository IdcHelmRepository `yaml:"idcHelmRepository"`
	Image             Image             `yaml:"image"`
	Global            Global            `yaml:"global"`
	Regions           map[string]Region `yaml:"regions"`
	UniverseDeployer  UniverseDeployer  `yaml:"universeDeployer"`
	Vault             Vault             `yaml:"vault"`
}

type ArgoCd struct {
	Enabled bool `yaml:"enabled"`
}

type Component struct {
	Commit string `yaml:"commit"`
	// If provided, environment-specific configuration files will come from this commit.
	// Otherwise, environment-specific configuration files will come from Commit.
	ConfigCommit string `yaml:"configCommit,omitempty"`
	Enabled      bool   `yaml:"enabled"`
}

type IdcHelmRepository struct {
	Url string `yaml:"url"`
}

type Image struct {
	RepositoryPrefix string `yaml:"repositoryPrefix"`
	Registry         string `yaml:"registry"`
}

type Global struct {
	Components  map[string]*Component `yaml:"components"`
	KubeContext string                `yaml:"kubeContext"`
	Vault       GlobalVault           `yaml:"vault"`
}

type UniverseDeployer struct {
	ArtifactRepositoryUrl string `yaml:"artifactRepositoryUrl"`
	// Component.Commit values reference this Git remote.
	// This is normally "origin".
	ComponentCommitGitRemote string `yaml:"componentCommitGitRemote"`
	// Component.ConfigCommit values reference this Git remote.
	// This is normally "origin".
	ComponentConfigCommitGitRemote string `yaml:"componentConfigCommitGitRemote"`
	ForceCreateReleases            bool   `yaml:"forceCreateReleases"`
	// NamespacedName.Name is a regex that matches the Helm release name.
	// The map key is unused except for merging map items from multiple yaml files.
	IgnoreHealthCheckFor       map[string]NamespacedName `yaml:"ignoreHealthCheckFor"`
	IncludeDeployK8sTlsSecrets bool                      `yaml:"includeDeployK8sTlsSecrets"`
	IncludeVaultConfigure      bool                      `yaml:"includeVaultConfigure"`
	IncludeVaultLoadSecrets    bool                      `yaml:"includeVaultLoadSecrets"`
	// If false, all deployment options that require access to the K8s API servers will be disabled.
	K8sApiEnabled      bool   `yaml:"k8sApiEnabled"`
	ManifestsGitBranch string `yaml:"manifestsGitBranch"`
	ManifestsGitRemote string `yaml:"manifestsGitRemote"`
	PatchCommand       string `yaml:"patchCommand"`
}

type Vault struct {
	Enabled bool `yaml:"enabled"`
}

type GlobalVault struct {
	Server GlobalVaultServer `yaml:"server"`
}

type GlobalVaultServer struct {
	Enabled bool `yaml:"enabled"`
}

type Region struct {
	Components        map[string]*Component       `yaml:"components"`
	KubeContext       string                      `yaml:"kubeContext"`
	Region            string                      `yaml:"region"`
	AvailabilityZones map[string]AvailabilityZone `yaml:"availabilityZones"`
}

type AvailabilityZone struct {
	AvailabilityZone string                `yaml:"availabilityZone"`
	Components       map[string]*Component `yaml:"components"`
	KubeContext      string                `yaml:"kubeContext"`
	NetworkCluster   NetworkCluster        `yaml:"networkCluster"`
	QuickConnect     QuickConnect          `yaml:"quickConnect"`
}

type NetworkCluster struct {
	KubeContext string `yaml:"kubeContext"`
}

type QuickConnect struct {
	KubeContext string `yaml:"kubeContext"`
}

type NamespacedName struct {
	Namespace string `yaml:"namespace"`
	Name      string `yaml:"name"`
}

func (e EnvConfigWithUnparsed) ToMultipleEnvConfig(ctx context.Context) (MultipleEnvConfig, error) {
	return MultipleEnvConfig{
		Environments: map[string]*EnvConfigWithUnparsed{
			e.Environment.Name: {
				EnvConfig:             e.EnvConfig,
				HelmfileDumpYamlBytes: e.HelmfileDumpYamlBytes,
			},
		},
	}, nil
}

// A type to store unparsed bytes.
type UnparsedBytes []byte

// When logging fields of this type, only show the size.
func (b UnparsedBytes) String() string {
	return fmt.Sprintf("(%d bytes)", len(b))
}
