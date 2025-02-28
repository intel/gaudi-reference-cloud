// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package filepaths

const (
	BuildEnvironments            = "build/environments"
	CreateReleasesUniverseConfig = "build/dynamic/create-releases-universe-config.json"
	DeploymentArtifactsTar       = "deployment/universe_deployer/deployment_artifacts/deployment_artifacts_tar.tar"
	HelmBinary                   = "external/helm3_linux_amd64/linux-amd64/helm"
	HelmChartVersionsDir         = "deployment/chart_versions"
	HelmfileBinary               = "external/helmfile_linux_amd64/helmfile"
	HelmfileConfigDir            = "deployment/helmfile"
	HelmfileEnvironments         = "deployment/helmfile/environments"
	IdcArgoCdInitialData         = "deployment/argocd/idc-argocd-initial-data"
	JqBinary                     = "external/jq_linux_amd64_file/file/jq"
	JwkSourceBaseDir             = "build/environments"
	KindBinary                   = "external/kind_linux_amd64/file/kind"
	KubectlBinary                = "external/kubectl_linux_amd64/file/kubectl"
	VaultBinary                  = "external/vault_linux_amd64/vault"
	YqBinary                     = "external/yq_linux_amd64_file/file/yq"
)

// Return the list of directories that contain environment-specific configuration files.
func ConfigDirs() []string {
	return []string{
		BuildEnvironments,
		HelmfileEnvironments,
	}
}
