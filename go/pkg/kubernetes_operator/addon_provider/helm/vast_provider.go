package helm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/rest"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart/loader"
)

type VastAddonProvider struct {
	RestConfig *rest.Config
}

func NewVastAddonProvider(restConfig *rest.Config) (*VastAddonProvider, error) {
	return &VastAddonProvider{
		RestConfig: restConfig,
	}, nil
}

func (p *VastAddonProvider) Put(ctx context.Context, addon *privatecloudv1alpha1.Addon) error {
	log := log.FromContext(ctx).WithName("VastAddonProvider.Put")

	// Validate required information to install helm chart.
	namespace := "default"
	if v, found := addon.Spec.Args["namespace"]; found {
		namespace = v
	}

	repoUrl, found := addon.Spec.Args["repoUrl"]
	if !found {
		return fmt.Errorf("repoURL is required")
	}

	name, found := addon.Spec.Args["name"]
	if !found {
		return fmt.Errorf("name is required")
	}

	endpoint, found := addon.Spec.Args["endpoint"]
	if !found {
		return fmt.Errorf("endpoint is required")
	}

	storagePath, found := addon.Spec.Args["storagePath"]
	if !found {
		return fmt.Errorf("storagePath is required")
	}

	viewPolicy, found := addon.Spec.Args["viewPolicy"]
	if !found {
		return fmt.Errorf("viewPolicy is required")
	}

	artifactSplit := strings.Split(addon.Spec.Artifact, "/")
	if len(artifactSplit) < 2 {
		return fmt.Errorf("artifact is not valid. Must be chart/version")
	}
	chartName := artifactSplit[0]
	version := artifactSplit[1]

	log.V(0).Info("Installing helm chart", logkeys.Name, name, logkeys.ChartName, chartName, logkeys.Version, version, logkeys.Namespace, namespace, logkeys.RepoURL, repoUrl, logkeys.Endpoint, endpoint, logkeys.StoragePath, storagePath, logkeys.ViewPolicy, viewPolicy)

	// Download chart repository index.yaml file.
	parsedURL, err := url.Parse(repoUrl)
	if err != nil {
		return errors.Wrapf(err, "Parse repoURL %s", repoUrl)
	}
	parsedURL.RawPath = path.Join(parsedURL.RawPath, "index.yaml")
	parsedURL.Path = path.Join(parsedURL.Path, "index.yaml")

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return errors.Wrapf(err, "Download index.yaml file")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download index.yaml file from %s, status code: %s", parsedURL.String(), resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrapf(err, "Read response body from %s", parsedURL.String())
	}

	indexFile := &repo.IndexFile{}
	if err := yaml.UnmarshalStrict(body, indexFile); err != nil {
		return err
	}

	// Download helm chart .tgz file.
	chartVersion, err := indexFile.Get(chartName, version)
	if err != nil {
		return errors.Wrapf(err, "chart %s with version %s not found", name, version)
	}

	if len(chartVersion.URLs) == 0 {
		return errors.Errorf("chart %s with version %s has no downloadable URLs", name, version)
	}
	log.V(0).Info("Chart to be installed", logkeys.ChartVersionURL, chartVersion.URLs[0])

	chartDownloader := downloader.ChartDownloader{
		Out:     os.Stderr,
		Verify:  downloader.VerifyNever,
		Getters: getter.All(&cli.EnvSettings{}),
	}

	chartFilename, _, err := chartDownloader.DownloadTo(chartVersion.URLs[0], version, "/tmp")
	if err != nil {
		return errors.Wrapf(err, "Download chart")
	}
	defer func() {
		if err := os.Remove(chartFilename); err != nil {
			log.Error(err, "Delete chart")
		}
	}()
	log.V(0).Info("Chart path", logkeys.FileName, chartFilename)

	chart, err := loader.Load(chartFilename)
	if err != nil {
		return errors.Wrapf(err, "Load chart")
	}

	if len(chart.Metadata.Type) != 0 && chart.Metadata.Type != "application" {
		return fmt.Errorf("chart type must be application")
	}

	if chart.Metadata.Dependencies != nil {
		return fmt.Errorf("chart with dependencies is not supported")
	}

	// Create helm chart configuration.
	actionConfig, err := initActionConfig(p.RestConfig, namespace)
	if err != nil {
		return errors.Wrapf(err, "Create action configuration")
	}

	// Check if helm chart exists.
	historyClient := action.NewHistory(actionConfig)
	historyClient.Max = 1

	// overide values for VAST CSI

	values := map[string]interface{}{

		"verifySsl":          false,
		"useLocalIpForMount": "127.0.0.1",
		"storageClasses": map[string]interface{}{
			"itac-filestorage-gp1": map[string]interface{}{
				"secretName":                "vast-mgmt",
				"endpoint":                  endpoint,
				"storagePath":               storagePath,
				"viewPolicy":                viewPolicy,
				"volumeNameFormat":          "csi:{namespace}:{name}:{id}",
				"ephemeralVolumeNameFormat": "csi:{namespace}:{name}:{id}",
				"reclaimPolicy":             "Delete",
				"mountOptions": []string{
					"nfsvers=4.1",
					"noresvport",
					"nconnect=16",
				},
			},
		},
		"image": map[string]interface{}{
			"csiVastPlugin": map[string]interface{}{
				"repository":      "vastdataorg/csi",
				"tag":             "v2.5.0",
				"imagePullPolicy": "IfNotPresent",
			},
		},
	}

	if _, err = historyClient.Run(name); err == driver.ErrReleaseNotFound {
		log.V(0).Info("Installing new helm chart")
		client := action.NewInstall(actionConfig)
		client.ReleaseName = name
		client.Namespace = namespace
		client.CreateNamespace = true

		client.ChartPathOptions.RepoURL = repoUrl
		client.ChartPathOptions.Version = version

		// Install helm chart.
		release, err := client.Run(chart, values)
		if err != nil {
			return errors.Wrapf(err, "Run client")
		}
		log.Info("Helm chart installed", logkeys.Release, release.Name)

		return nil
	}

	log.V(0).Info("Upgrading helm chart")
	client := action.NewUpgrade(actionConfig)
	client.Namespace = namespace

	client.ChartPathOptions.RepoURL = repoUrl
	client.ChartPathOptions.Version = version

	release, err := client.Run(name, chart, values)
	if err != nil {
		return errors.Wrapf(err, "Run upgrade")
	}
	log.Info("Helm chart upgraded", logkeys.Release, release.Name)

	return nil
}

func (p *VastAddonProvider) Get(ctx context.Context, name string, namespace string) (*privatecloudv1alpha1.AddonStatus, error) {
	return nil, nil
}

func (p *VastAddonProvider) Delete(ctx context.Context, addon *privatecloudv1alpha1.Addon) error {
	log := log.FromContext(ctx).WithName("VastAddonProvider.Delete")

	// Validate required information to delete helm chart.
	namespace := "default"
	if v, found := addon.Spec.Args["namespace"]; found {
		namespace = v
	}

	name, found := addon.Spec.Args["name"]
	if !found {
		return fmt.Errorf("name is required")
	}

	log.V(0).Info("Deleting helm chart", logkeys.Name, name, logkeys.Namespace, namespace)

	// Create helm chart client.
	actionConfig, err := initActionConfig(p.RestConfig, namespace)
	if err != nil {
		return errors.Wrapf(err, "Create action configuration")
	}

	client := action.NewUninstall(actionConfig)
	client.KeepHistory = false

	// Uninstall helm chart.
	_, err = client.Run(name)
	if err != nil {
		return errors.Wrapf(err, "Run client")
	}
	log.Info("Helm chart deleted", logkeys.Name, name)

	return nil
}
