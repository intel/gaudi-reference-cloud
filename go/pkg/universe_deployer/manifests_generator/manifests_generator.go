// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package manifests_generator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/store_forward_logger"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/deployer_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/env_config/types"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/filepaths"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/universe_config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/universe_deployer/util"
)

// A set of application manifests (Argo CD Applications, Helm releases).
// Created by:
//   - [helmReleasesToManifests]
//   - [multi_version_manifests_generator.MultiVersionManifestsGenerator.manifestsTarToManifests].
type Manifests struct {
	Manifests []Manifest
}

// Sort manifests for easier troubleshooting.
func (m *Manifests) Sort() {
	sort.Slice(m.Manifests, func(i, j int) bool {
		return m.Manifests[i].ConfigFileName < m.Manifests[j].ConfigFileName
	})
}

type Manifest struct {
	// Configuration Git commit.
	ConfigCommit string
	ConfigFileData
	// Relative path such as applications/idc-global-services/idc-staging/idc02-k01-ekcp/cloudaccount/config.json.
	ConfigFileName string
	// Application Git commit.
	GitCommit   string
	KubeContext string
}

func (m Manifest) ReleaseName() string {
	return m.Envconfig.ReleaseName
}

func (m Manifest) Namespace() string {
	return m.Envconfig.Namespace
}

// A type for parsing the output of "helmfile list".
type HelmRelease struct {
	Chart     string `json:"chart"`
	Enabled   bool   `json:"enabled"`
	Installed bool   `json:"installed"`
	Labels    string `json:"labels"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
}

const ArgoCdConfigFileName = "config.json"

// A type for writing an Argo CD config.json file.
type ConfigFileData struct {
	Envconfig Envconfig `json:"envconfig"`
}

type Envconfig struct {
	ReleaseName   string `json:"releaseName"`
	ChartName     string `json:"chartName"`
	ChartVersion  string `json:"chartVersion"`
	ChartRegistry string `json:"chartRegistry"`
	Namespace     string `json:"namespace"`
	GitCommit     string `json:"gitCommit,omitempty"`
	ConfigCommit  string `json:"configCommit,omitempty"`
}

type HelmfileEnv struct {
	Commit               string
	ConfigCommit         string
	HelmBinary           string
	HelmChartVersionsDir string
	HelmfileBinary       string
	HelmfileConfigDir    string
	HelmfileConfigFile   string
	SecretsDir           string
}

func (e *HelmfileEnv) HelmfileEnv() []string {
	helmfileEnv := os.Environ()
	helmfileEnv = append(helmfileEnv, "CONFIG_COMMIT="+e.ConfigCommit)
	helmfileEnv = append(helmfileEnv, "GIT_COMMIT="+e.Commit)
	helmfileEnv = append(helmfileEnv, "HELM_CHART_VERSIONS_DIR="+e.HelmChartVersionsDir)
	helmfileEnv = append(helmfileEnv, "SECRETS_DIR="+e.SecretsDir)
	return helmfileEnv
}

func (e *HelmfileEnv) HelmfileVersion(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, e.HelmfileBinary, "--version")
	return util.RunCmd(ctx, cmd)
}

func (e *HelmfileEnv) WriteHelmValues(ctx context.Context, idcEnv string, selectors []string, ouputFileTemplate string) error {
	args := []string{
		"write-values",
		"--allow-no-matching-release",
		"--helm-binary", e.HelmBinary,
		"--file", e.HelmfileConfigFile,
		"--environment", idcEnv,
		"--output-file-template", ouputFileTemplate,
		"--skip-deps",
	}
	for _, selector := range selectors {
		args = append(args, "--selector", selector)
	}
	cmd := exec.CommandContext(ctx, e.HelmfileBinary, args...)
	cmd.Dir = e.HelmfileConfigDir
	cmd.Env = e.HelmfileEnv()
	return util.RunCmd(ctx, cmd)
}

func (e *HelmfileEnv) GetReleases(ctx context.Context, idcEnv string, selectors []string) ([]HelmRelease, error) {
	log := log.FromContext(ctx).WithName("GetReleases")
	log.Info("BEGIN")
	defer log.Info("END")
	args := []string{
		"list",
		"--allow-no-matching-release",
		"--helm-binary", e.HelmBinary,
		"--file", e.HelmfileConfigFile,
		"--environment", idcEnv,
		"--output", "json",
		"--skip-charts",
	}
	for _, selector := range selectors {
		args = append(args, "--selector", selector)
	}
	cmd := exec.CommandContext(ctx, e.HelmfileBinary, args...)
	cmd.Dir = e.HelmfileConfigDir
	cmd.Env = e.HelmfileEnv()
	log.Info("Running", "Args", cmd.Args, "Dir", cmd.Dir, "Env", cmd.Env)
	output, err := cmd.Output()
	log.Info("Completed", "err", err, "outputLen", len(output))
	// Logging complete output when run in Jenkins will cause the process to hang.
	const maxLengthToLog = 1024
	if len(output) > maxLengthToLog {
		log.V(2).Info("Output", "outputTruncated", string(output[:maxLengthToLog]))
	} else {
		log.V(2).Info("Output", "output", string(output))
	}
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			log.Info("Stderr", "Stderr", string(exitError.Stderr))
		}
		return nil, err
	}

	var releaseList []HelmRelease
	if len(output) > 0 {
		if err := json.Unmarshal(output, &releaseList); err != nil {
			return nil, err
		}
	}

	return releaseList, nil
}

// Convert HelmReleases (produced by helmfile) to a Manifests object which contains
// all required information for each application.
func helmReleasesToManifests(
	ctx context.Context,
	releases []HelmRelease,
	repoToRegistryMap map[string]string,
	defaultChartRegistry string,
	idcEnv string,
	deployerConfigFile string,
	commit string,
	configCommit string,
) (*Manifests, error) {
	log := log.FromContext(ctx).WithName("helmReleasesToManifests")
	manifests := Manifests{}
	for _, release := range releases {
		labels := map[string]string{}
		for _, pair := range strings.Split(release.Labels, ",") {
			tokens := strings.SplitN(pair, ":", 2)
			if len(tokens) == 2 {
				labels[tokens[0]] = tokens[1]
			}
		}
		log.V(2).Info("labels", "labels", labels)
		environmentName := labels["environmentName"]
		geographicScope := labels["geographicScope"]
		kubeContext := labels["kubeContext"]
		if kubeContext == "" {
			return nil, fmt.Errorf("kubeContext is not set for release %s", release.Name)
		}
		region := labels["region"]

		// split the release.Chart at the first "/", so if the input is "idc-networking/amd64/ovn-central", then we take repo name "idc-networking" and chart name "amd64/ovn-central"
		chartTokens := strings.SplitN(release.Chart, "/", 2)
		if len(chartTokens) != 2 {
			return nil, fmt.Errorf("unable to parse chart '%s'", release.Chart)
		}
		repo := chartTokens[0]
		chartName := chartTokens[1]

		chartRegistry := repoToRegistryMap[repo]
		if chartRegistry == "" {
			if defaultChartRegistry == "" {
				return nil, fmt.Errorf("'helmRepositories.%s.registry' and 'environments.%s.helm.registry' are empty in %s",
					repo, idcEnv, deployerConfigFile)
			}
			chartRegistry = defaultChartRegistry
			chartName = "intelcloud/" + chartName
		}

		var filename string
		if geographicScope == "global" {
			if environmentName == "" {
				return nil, fmt.Errorf("environmentName is not set for release %s", release.Name)
			}
			filename = filepath.Join("applications", "idc-global-services", environmentName, kubeContext, release.Name, ArgoCdConfigFileName)
		} else if geographicScope == "regional" || geographicScope == "az" {
			if region == "" {
				return nil, fmt.Errorf("region is not set for release %s", release.Name)
			}
			filename = filepath.Join("applications", "idc-regional", region, kubeContext, release.Name, ArgoCdConfigFileName)
		} else if geographicScope == "az-network" {
			if region == "" {
				return nil, fmt.Errorf("region is not set for release %s", release.Name)
			}
			filename = filepath.Join("applications", "idc-network", region, kubeContext, release.Name, ArgoCdConfigFileName)
		} else {
			return nil, fmt.Errorf("unknown geographicScope '%s'", geographicScope)
		}

		configFileData := ConfigFileData{
			Envconfig: Envconfig{
				ReleaseName:   release.Name,
				ChartName:     chartName,
				ChartVersion:  release.Version,
				ChartRegistry: chartRegistry,
				Namespace:     release.Namespace,
				GitCommit:     commit,
				ConfigCommit:  configCommit,
			},
		}

		manifest := Manifest{
			ConfigFileData: configFileData,
			ConfigFileName: filename,
			GitCommit:      commit,
			KubeContext:    kubeContext,
		}
		manifests.Manifests = append(manifests.Manifests, manifest)
	}

	manifests.Sort()
	return &manifests, nil
}

// Write config.json files for each Argo application.
// This is used by the Argo CD ApplicationSet (/deployment/argocd/idc-argocd-initial-data/manifests/base/applications/base-appset.yaml).
// This also writes README.md and OWNED_BY_UNIVERSE_DEPLOYER.md.
// Based on /deployment/helmfile/scripts/generate_config_jsons.sh.
func writeConfigJsonFiles(
	ctx context.Context,
	manifests *Manifests,
	manifestDir string,
) error {
	log := log.FromContext(ctx).WithName("WriteConfigJsonFiles")
	ownedDirectories := map[string]bool{}
	for _, manifest := range manifests.Manifests {
		filename := filepath.Join(manifestDir, manifest.ConfigFileName)
		configFileBytes, err := json.MarshalIndent(&manifest.ConfigFileData, "", "  ")
		if err != nil {
			return err
		}
		configFileBytes = append(configFileBytes, "\n"...)
		log.V(3).Info(string(configFileBytes))
		if err := os.WriteFile(filename, configFileBytes, 0644); err != nil {
			return err
		}
		ownedDirectories[filepath.Dir(filename)] = true
		ownedDirectories[filepath.Dir(filepath.Dir(filename))] = true
	}

	for ownedDirectory := range ownedDirectories {
		readmePath := filepath.Join(ownedDirectory, "README.md")
		if err := os.WriteFile(readmePath, []byte(util.GetOwnedDirectoryReadmeContent()), 0644); err != nil {
			return err
		}
		ownedDirectoryMarkerPath := filepath.Join(ownedDirectory, util.GetOwnedDirectoryMarkerFileName())
		if err := os.WriteFile(ownedDirectoryMarkerPath, []byte(util.GetOwnedDirectoryReadmeContent()), 0644); err != nil {
			return err
		}
	}
	return nil
}

func CreateUuidSecretFileIfMissing(ctx context.Context, filename string) error {
	log := log.FromContext(ctx).WithName("CreateUuidSecretFileIfMissing").WithValues("filename", filename)
	_, err := os.Stat(filename)
	if err == nil {
		log.Info("Skipping existing secret file")
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	log.Info("Creating new secret file")
	return os.WriteFile(filename, []byte(uuid.NewString()), 0600)
}

// Secrets should not be required to generate manifests!
// However, some manifests currently require secrets so we generate random secrets here.
// Warning: These secrets will be stored in plaintext in the idc-argocd Git repository.
// This function is a partial reimplementation of deployment/common/vault/make-secrets.sh.
func MakeSecrets(ctx context.Context, envConfig *types.EnvConfig, secretsDir string) error {
	if err := os.MkdirAll(secretsDir, 0700); err != nil {
		return err
	}
	if err := CreateUuidSecretFileIfMissing(ctx, filepath.Join(secretsDir, "gitea_admin_password")); err != nil {
		return err
	}
	for _, region := range envConfig.Values.Regions {
		if err := CreateUuidSecretFileIfMissing(ctx, filepath.Join(secretsDir, fmt.Sprintf("%s-netbox_secretKey", region.Region))); err != nil {
			return err
		}
		for _, availabilityZone := range region.AvailabilityZones {
			if err := CreateUuidSecretFileIfMissing(ctx, filepath.Join(secretsDir, fmt.Sprintf("%s-inspector_password", availabilityZone.AvailabilityZone))); err != nil {
				return err
			}
			if err := CreateUuidSecretFileIfMissing(ctx, filepath.Join(secretsDir, fmt.Sprintf("%s-ironic_password", availabilityZone.AvailabilityZone))); err != nil {
				return err
			}
		}
	}
	// Create databases secrets. These are used for development environments only.
	databases := []string{
		"authz",
		"billing",
		"catalog",
		"cloudaccount",
		"cloudcredits",
		"cloudmonitor",
		"insights",
		"metering",
		"notification",
		"productcatalog",
		"training",
		"usage",
	}
	regionalDatabases := []string{
		"cloudmonitor-logs",
		"compute",
		"dpai",
		"fleet-admin",
		"insights",
		"kfaas",
		"netbox-postgres",
		"netbox-redis",
		"network",
		"quota-management-service",
		"sdn-vn-controller",
		"storage",
		"training",
	}
	for _, region := range envConfig.Values.Regions {
		for _, regionalDatabase := range regionalDatabases {
			databases = append(databases, region.Region+"-"+regionalDatabase)
		}
	}
	for _, database := range databases {
		if err := CreateUuidSecretFileIfMissing(ctx, filepath.Join(secretsDir, fmt.Sprintf("%s_db_admin_password", database))); err != nil {
			return err
		}
		if err := CreateUuidSecretFileIfMissing(ctx, filepath.Join(secretsDir, fmt.Sprintf("%s_db_user_password", database))); err != nil {
			return err
		}
	}
	return nil
}

type ManifestsGenerator struct {
	DeploymentArtifactsDir       string
	SecretsDir                   string
	UniverseConfig               *universe_config.UniverseConfig
	MultipleEnvConfig            *types.MultipleEnvConfig
	Commit                       string
	Components                   []string
	ConfigCommit                 string
	Snapshot                     bool
	ManifestsTar                 string
	OverrideDefaultChartRegistry string
}

// Generate Argo CD manifests. Store in a tar file.
func (m ManifestsGenerator) GenerateManifests(ctx context.Context) (*Manifests, error) {
	ctx, log := log.IntoContextWithLogger(ctx, log.FromContext(ctx).WithName("GenerateManifests"))
	log.Info("BEGIN")
	defer log.Info("END")

	if m.DeploymentArtifactsDir == "" {
		return nil, fmt.Errorf("required parameter DeploymentArtifactsDir is empty")
	}

	universeConfig := m.UniverseConfig
	if !m.Snapshot {
		universeConfig = universeConfig.Trimmed(ctx, m.Commit)
	}
	log.Info("universeConfig", "universeConfig", universeConfig)

	tempDir, err := os.MkdirTemp("", "universe_deployer_manifests_generator_")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// Get deployer config from commit.
	deployerConfigFile := filepath.Join(m.DeploymentArtifactsDir, "deployment/universe_deployer/deployment_artifacts/config.yaml")
	deployerConfig, err := deployer_config.NewConfigFromFile(ctx, deployerConfigFile)
	if err != nil {
		return nil, err
	}
	log.Info("deployerConfig", "deployerConfig", deployerConfig)

	helmBinary := filepath.Join(m.DeploymentArtifactsDir, filepaths.HelmBinary)
	helmfileBinary := filepath.Join(m.DeploymentArtifactsDir, filepaths.HelmfileBinary)
	helmChartVersionsDir := filepath.Join(m.DeploymentArtifactsDir, filepaths.HelmChartVersionsDir)
	helmfileConfigDir := filepath.Join(m.DeploymentArtifactsDir, filepaths.HelmfileConfigDir)

	if m.SecretsDir == "" {
		// Use a non-existing directory.
		m.SecretsDir = filepath.Join(tempDir, "local/secrets")
	}

	helmfileConfigFile := ""
	if len(m.Components) == 0 {
		helmfileConfigFile = "helmfile.yaml"
	} else if len(m.Components) == 1 {
		component := m.Components[0]
		helmfileConfigFile = "helmfile-" + component + ".yaml"
	} else {
		return nil, fmt.Errorf("more than one component is unsupported")
	}

	helmfileEnv := HelmfileEnv{
		Commit:               m.Commit,
		ConfigCommit:         m.ConfigCommit,
		HelmBinary:           helmBinary,
		HelmChartVersionsDir: helmChartVersionsDir,
		HelmfileBinary:       helmfileBinary,
		HelmfileConfigDir:    helmfileConfigDir,
		HelmfileConfigFile:   helmfileConfigFile,
		SecretsDir:           m.SecretsDir,
	}
	manifestDir := filepath.Join(tempDir, "manifests")
	if err := os.MkdirAll(manifestDir, 0750); err != nil {
		return nil, err
	}
	manifests := Manifests{}

	for idcEnv, universeEnvironment := range universeConfig.Environments {
		log.Info("Processing environment", "idcEnv", idcEnv)

		envConfig, ok := m.MultipleEnvConfig.Environments[idcEnv]
		if !ok {
			return nil, fmt.Errorf("environment %s not found in MultipleEnvConfig", idcEnv)
		}

		forceAllComponents := universeEnvironment.ForceAllComponents
		if forceAllComponents {
			log.Info("All components enabled", "idcEnv", idcEnv)
		}

		var selectorsGlobal []string
		if forceAllComponents {
			selectorsGlobal = append(selectorsGlobal, "geographicScope=global")
		} else {
			for component := range universeEnvironment.Components {
				log.Info("Component enabled", "component", component, "idcEnv", idcEnv)
				selectorsGlobal = append(selectorsGlobal,
					fmt.Sprintf("geographicScope=global,component=%s", component))
			}
		}

		var selectorsRegional []string
		var selectorsNetwork []string
		if forceAllComponents {
			selectorsRegional = append(selectorsRegional,
				"geographicScope=regional",
				"geographicScope=az",
			)
			selectorsNetwork = append(selectorsNetwork, "geographicScope=az-network")
		} else {
			for region, universeRegion := range universeEnvironment.Regions {
				for component := range universeRegion.Components {
					log.Info("Component enabled", "component", component, "idcEnv", idcEnv, "region", region)
					selectorsRegional = append(selectorsRegional,
						fmt.Sprintf("geographicScope=regional,region=%s,component=%s",
							region, component))
				}
				for availabilityZone, universeAvailabilityZone := range universeRegion.AvailabilityZones {
					for component := range universeAvailabilityZone.Components {
						log.Info("Component enabled", "component", component, "idcEnv", idcEnv, "region", region, "availabilityZone", availabilityZone)
						selectorsRegional = append(selectorsRegional,
							fmt.Sprintf("geographicScope=az,region=%s,availabilityZone=%s,component=%s",
								region, availabilityZone, component))
						selectorsNetwork = append(selectorsNetwork,
							fmt.Sprintf("geographicScope=az-network,region=%s,availabilityZone=%s,component=%s",
								region, availabilityZone, component))
					}
				}
			}
		}

		if err := helmfileEnv.HelmfileVersion(ctx); err != nil {
			return nil, err
		}

		if err := MakeSecrets(ctx, &envConfig.EnvConfig, m.SecretsDir); err != nil {
			return nil, err
		}

		g := store_forward_logger.NewStoreForwardErrGroup(ctx)

		if len(selectorsGlobal) > 0 {
			outputFileTemplate := manifestDir + "/applications/idc-global-services/{{ .Release.Labels.environmentName }}/{{ .Release.Labels.kubeContext }}/{{ .Release.Name }}/values.yaml"
			desc := "manifests generator, global"
			g.Go(desc, func(ctx context.Context) error {
				return helmfileEnv.WriteHelmValues(ctx, idcEnv, selectorsGlobal, outputFileTemplate)
			})
		}

		if len(selectorsRegional) > 0 {
			outputFileTemplate := manifestDir + "/applications/idc-regional/{{ .Release.Labels.region }}/{{ .Release.Labels.kubeContext }}/{{ .Release.Name }}/values.yaml"
			desc := "manifests generator, regional"
			g.Go(desc, func(ctx context.Context) error {
				return helmfileEnv.WriteHelmValues(ctx, idcEnv, selectorsRegional, outputFileTemplate)
			})
		}

		if len(selectorsNetwork) > 0 {
			outputFileTemplate := manifestDir + "/applications/idc-network/{{ .Release.Labels.region }}/{{ .Release.Labels.kubeContext }}/{{ .Release.Name }}/values.yaml"
			desc := "manifests generator, network"
			g.Go(desc, func(ctx context.Context) error {
				return helmfileEnv.WriteHelmValues(ctx, idcEnv, selectorsNetwork, outputFileTemplate)
			})
		}

		releasesChan := make(chan []HelmRelease, 1)
		g.Go("GetReleases", func(ctx context.Context) error {
			selectorsAll := append(append(append([]string{}, selectorsGlobal...), selectorsRegional...), selectorsNetwork...)
			releases, err := helmfileEnv.GetReleases(ctx, idcEnv, selectorsAll)
			releasesChan <- releases
			close(releasesChan)
			return err
		})

		if err := g.Wait(); err != nil {
			return nil, err
		}

		releases := <-releasesChan
		log.V(2).Info("releases", "releases", releases)

		repoToRegistryMap := map[string]string{}
		for k, v := range deployerConfig.HelmRepositories {
			repoToRegistryMap[k] = v.Registry
		}
		defaultChartRegistry := m.OverrideDefaultChartRegistry
		if defaultChartRegistry == "" {
			helmUrl, err := url.Parse("https://" + envConfig.EnvConfig.Values.IdcHelmRepository.Url)
			if err != nil {
				return nil, err
			}
			defaultChartRegistry = helmUrl.Host
		}
		log.Info("defaultChartRegistry", "defaultChartRegistry", defaultChartRegistry)

		envManifests, err := helmReleasesToManifests(
			ctx,
			releases,
			repoToRegistryMap,
			defaultChartRegistry,
			idcEnv,
			deployerConfigFile,
			m.Commit,
			m.ConfigCommit,
		)
		if err != nil {
			return nil, err
		}

		if err := writeConfigJsonFiles(ctx, envManifests, manifestDir); err != nil {
			return nil, err
		}

		manifests.Manifests = append(manifests.Manifests, envManifests.Manifests...)
	}

	// Create manifests tar file.
	// Tar file must be deterministic.
	cmd := exec.CommandContext(ctx, "/bin/tar",
		"-C", manifestDir,
		"--sort=name",
		"--owner=root:0",
		"--group=root:0",
		"--mtime=@0",
		"-f", m.ManifestsTar,
		"-cv",
		".",
	)
	cmd.Env = os.Environ()
	if err := util.RunCmd(ctx, cmd); err != nil {
		return nil, err
	}

	log.Info("Manifests generated", "manifestsTar", m.ManifestsTar, "manifestDir", manifestDir)

	return &manifests, nil
}
