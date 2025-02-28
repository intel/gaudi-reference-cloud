package helm

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	diskcached "k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	privatecloudv1alpha1 "github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/kubernetes_operator/api/v1alpha1"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/log/logkeys"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/chart/loader"
)

const (
	defaultStorageDriver = "secret"

	WekaFsPluginHelmChartName = "weka-fs-plugin"

	WekaFsPluginNamespaceKey = "namespace"
	WekaFsPluginRepoUrlKey   = "repoUrl"
	WekaFsPluginNameKey      = "name"
	WekaFsPluginPrefix       = "prefix"
)

type HelmAddonProvider struct {
	RestConfig *rest.Config
}

func NewHelmAddonProvider(restConfig *rest.Config) (*HelmAddonProvider, error) {
	return &HelmAddonProvider{
		RestConfig: restConfig,
	}, nil
}

func (p *HelmAddonProvider) Put(ctx context.Context, addon *privatecloudv1alpha1.Addon) error {
	log := log.FromContext(ctx).WithName("HelmAddonProvider.Put")

	// Validate required information to install helm chart.
	namespace := "default"
	if v, found := addon.Spec.Args[WekaFsPluginNamespaceKey]; found {
		namespace = v
	}

	repoUrl, found := addon.Spec.Args[WekaFsPluginRepoUrlKey]
	if !found {
		return fmt.Errorf("repoURL is required")
	}

	name, found := addon.Spec.Args[WekaFsPluginNameKey]
	if !found {
		return fmt.Errorf("name is required")
	}

	artifactSplit := strings.Split(addon.Spec.Artifact, "/")
	if len(artifactSplit) < 2 {
		return fmt.Errorf("artifact is not valid. Must be chart/version")
	}
	chartName := artifactSplit[0]
	version := artifactSplit[1]

	log.V(0).Info("Installing helm chart", logkeys.Name, name, logkeys.ChartName, chartName, logkeys.Version, version, logkeys.Namespace, namespace, logkeys.RepoURL, repoUrl)

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
	values := map[string]interface{}{
		"pluginConfig": map[string]interface{}{
			"objectNaming": map[string]interface{}{
				"volumePrefix": addon.Spec.Args[WekaFsPluginPrefix],
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

func (p *HelmAddonProvider) Get(ctx context.Context, name string, namespace string) (*privatecloudv1alpha1.AddonStatus, error) {
	return nil, nil
}

func (p *HelmAddonProvider) Delete(ctx context.Context, addon *privatecloudv1alpha1.Addon) error {
	log := log.FromContext(ctx).WithName("HelmAddonProvider.Delete")

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

// initActionConfig initializes the action configuration for helm. The namespace is used to create the k8s secret
// that will serve as the storage driver to store information of the installed chart (releases).
func initActionConfig(restConfig *rest.Config, namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(
		NewRESTClientGetterFromRestConfig(restConfig, namespace),
		namespace,
		defaultStorageDriver,
		stdlog.Printf); err != nil {
		return nil, err
	}

	opts := []registry.ClientOption{
		registry.ClientOptDebug(false),
		registry.ClientOptEnableCache(true),
		registry.ClientOptCredentialsFile(""),
	}

	registryClient, err := registry.NewClient(opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "create registry client")
	}

	actionConfig.RegistryClient = registryClient

	return actionConfig, nil
}

type RESTClientGetterFromRestConfig struct {
	RestConfig *rest.Config
	Namespace  string
}

// NewRESTClientGetterFromRestConfig returns a new RESTClientGetterFromRestConfig that implements the RESTClientGetter interface.
// This is done because currently the Helm library only supports the RESTClientGetter interface to create the REST client.
func NewRESTClientGetterFromRestConfig(restConfig *rest.Config, namespace string) *RESTClientGetterFromRestConfig {
	return &RESTClientGetterFromRestConfig{
		RestConfig: restConfig,
		Namespace:  namespace,
	}
}

func (r *RESTClientGetterFromRestConfig) ToRESTConfig() (*rest.Config, error) {
	return r.RestConfig, nil
}

func (r *RESTClientGetterFromRestConfig) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	cacheDir := path.Join("/tmp", "kube", "cache")

	discoveryCacheDir := filepath.Join(cacheDir, "discovery")

	httpCacheDir := filepath.Join(cacheDir, "http")

	return diskcached.NewCachedDiscoveryClientForConfig(r.RestConfig, discoveryCacheDir, httpCacheDir, time.Duration(6*time.Hour))
}

func (r *RESTClientGetterFromRestConfig) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := r.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient, nil)

	return expander, nil
}

func (r *RESTClientGetterFromRestConfig) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	// We set the namespace here so that the client is created with the correct namespace. This is required because
	// if any of the helmchart templates does not have a variable to set the namespace, default namespace is used.
	return clientcmd.NewDefaultClientConfig(*clientcmdapi.NewConfig(), &clientcmd.ConfigOverrides{
		Context: clientcmdapi.Context{
			Namespace: r.Namespace,
		},
	})
}
