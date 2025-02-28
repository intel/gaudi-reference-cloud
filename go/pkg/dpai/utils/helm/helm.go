// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package helm

import (
	// "encoding/json"

	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"

	// "helm.sh/helm/v3/pkg/chartutil"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/config"
	"github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/go/pkg/dpai/utils"
	helmclient "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/kube"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	HelmUserName = ""
)

type ImageReference struct {
	Repository string
	Tag        string
}

type HelmChartReference struct {
	RepoName  string
	RepoUrl   string
	ChartName string
	Version   string
	Username  string
	SecretKey string
}

func LoginRegistry(conf config.OCIRegistryConfig) error {
	configDir := "/tmp/.helm"

	// the config dir to fix the nonexistent error used internally by Helm client to dump post registry login
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	registryConfig := filepath.Join(configDir, "registry.json")
	if err := os.Setenv("HELM_REGISTRY_CONFIG", registryConfig); err != nil {
		log.Printf("failed to set HELM_REGISTRY_CONFIG: %+v", err)
	}

	client, err := registry.NewClient(
		registry.ClientOptCredentialsFile(registryConfig),
	)

	if err != nil {
		return err
	}

	password, err := utils.ReadSecretFile(conf.PasswordFile)
	if err != nil {
		return err
	}

	err = client.Login(conf.Host, registry.LoginOptBasicAuth(conf.Username, *password))
	if err != nil {
		return fmt.Errorf("error: Failed to login to helm registry : Error message: %+v", err)
	}
	return nil
}

func GetHelmClient(restconfig *rest.Config, namespace string) (helmclient.Client, error) {
	opt := &helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        namespace, // Change this to the namespace you wish the client to operate in.
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true, // Change this to false if you don't want linting.
			DebugLog: func(format string, v ...interface{}) {
				// Change this to your own logger. Default is 'log.Printf(format, v...)'.
			},
		},
		RestConfig: restconfig,
	}

	helmClient, err := helmclient.NewClientFromRestConf(opt)
	if err != nil {
		return nil, fmt.Errorf("error: Failed to get helm client: Error message: %+v", err)
	}
	log.Printf("Helm Client %+v ", helmClient)

	return helmClient, nil
}

func LoadValuesYaml(valuesYamlPath string) (string, error) {
	valuesYamlBytes, err := os.ReadFile(valuesYamlPath)
	if err != nil {
		return "", err
	}
	return string(valuesYamlBytes), nil
}

func ListInstalls(kubeconfig *string) ([]string, error) {

	var releaseNames []string

	actionConfig := new(action.Configuration)
	err := actionConfig.Init(kube.GetConfig(*kubeconfig, "", ""), "", "", log.Printf)
	if err != nil {
		return nil, fmt.Errorf("error initalizing %+v", err)
	}

	releases, err := action.NewList(actionConfig).Run()
	if err != nil {
		return nil, fmt.Errorf("error listing %+v", err)
	}

	for _, release := range releases {
		log.Printf("The release object looks like this: %+v", release)
		log.Println(release.Name)
		releaseNames = append(releaseNames, release.Name)
	}

	return releaseNames, nil
}

// func InstallChart( kubeconfig *string, chartName string, chartVersion string, namespace string) {

//     actionConfig := new(action.Configuration)
//     err := actionConfig.Init(kube.GetConfig(*kubeconfig, "", ""), "", "", log.Printf)
//     if err != nil {
//         log.Fatal(err)
//     }

//     // Load the chart
//     chart, err := chartutil.Load(chartName)
//     if err != nil {
//         log.Fatal(err)
//     }

//     // Set the chart version
//     chart.Version = chartVersion

//     // Install the chart
//     err = action.NewInstall(actionConfig).Run(chart, namespace)
//     if err != nil {
//         log.Fatal(err)
//     }
// }

//
// package helm

// import (
//     "context"
//     "fmt"

//     "helm.sh/helm/v3/chart"
//     "helm.sh/helm/v3/chart/loader"
//     "helm.sh/helm/v3/cmd/helm"
//     "helm.sh/helm/v3/pkg/action"
//     "helm.sh/helm/v3/pkg/chartutil"
//     "helm.sh/helm/v3/pkg/storage/remote"
// 	"github.com/mittwald/go-helm-client"
// )

// func getHelmClient(namespace string) (helmclient.Client){

// 	opt := &helmclient.KubeConfClientOptions{
// 		Options: &helmclient.Options{
// 			Namespace:        "default", // Change this to the namespace you wish to install the chart in.
// 			RepositoryCache:  "/tmp/.helmcache",
// 			RepositoryConfig: "/tmp/.helmrepo",
// 			Debug:            true,
// 			Linting:          true, // Change this to false if you don't want linting.
// 			DebugLog: func(format string, v ...interface{}) {
// 				// Change this to your own logger. Default is 'log.Printf(format, v...)'.
// 			},
// 		},
// 		KubeContext: "",
// 		KubeConfig:  []byte{},
// 	}

// 	helmClient, err := NewClientFromKubeConf(opt, Burst(100), Timeout(10e9))
// 	if err != nil {
// 		panic(err)
// 	}
// 	_ = helmClient
// }

// func main() {
//     // List all versions available for the given chart name
//     listChartVersions(context.TODO(), "nginx")

//     // Install the chart on the given namespace
//     installChart(context.TODO(), "nginx", "latest", "default")

//     // List all the Helm installations on the given namespace
//     listReleases(context.TODO(), "default")
// }

// func listChartVersions(ctx context.Context, chartName string) {
//     repoClient := remote.New(&remote.Options{
//         RepositoryURL: "https://charts.helm.sh",
//     })

//     charts, err := repoClient.ListCharts(ctx, chartName)
//     if err != nil {
//         fmt.Printf("Failed to list chart versions: %v\n", err)
//         return
//     }

//     for _, chart := range charts {
//         fmt.Printf("Chart: %s, Version: %s\n", chart.Name, chart.Version)
//     }
// }

// func installChart(ctx context.Context, chartName, chartVersion, namespace string) {
//     actionConfig := new(action.Configuration)
//     actionConfig.Namespace = namespace

//     // Load the chart
//     chart, err := loader.LoadFile(chartName)
//     if err != nil {
//         fmt.Printf("Failed to load chart: %v\n", err)
//         return
//     }

//     // Set the chart version
//     chart.Version = chartVersion

//     // Install the chart
//     err = helm.Install(actionConfig, chart)
//     if err != nil {
//         fmt.Printf("Failed to install chart: %v\n", err)
//         return
//     }

//     fmt.Printf("Chart installed successfully: %s\n", chartName)
// }

// func listReleases(ctx context.Context, namespace string) {
//     actionConfig := new(action.Configuration)
//     actionConfig.Namespace = namespace

//     // Get the release client
//     releaseClient := action.NewRelease(actionConfig)

//     // List the releases
//     releases, err := releaseClient.List(ctx)
//     if err != nil {
//         fmt.Printf("Failed to list releases: %v\n", err)
//         return
//     }

//     for _, release := range releases {
//         fmt.Printf("Release: %s, Name: %s, Namespace: %s\n", release.Name, release.Chart.Name, release.Namespace)
//     }
// }

// ***Rajesh added code: Start

var credpath string = "/tmp/config.json"

//var dirPath string = "/tmp"

const (
	HelmUserName_HD = ""
)

// Can be deleted
type ImageReference_HD struct {
	Repository string
	Tag        string
}

// Can be deleted
type HelmChartReference_HD struct {
	RepoName  string
	RepoUrl   string
	ChartName string
	Version   string
	Username  string
	SecretKey string
}

// Can be deleted
func LoginRegistry_HD() error {
	client, err := registry.NewClient(registry.ClientOptCredentialsFile(credpath))
	if err != nil {
		return err
	}

	password, err := utils.ReadSecretFile("/vault/secrets/default_registry_password")
	if err != nil {
		fmt.Println("password file does not exists...")
	}

	err = client.Login("icir.cps.intel.com/dpai", registry.LoginOptBasicAuth("robot$dpai+robot_dpai_backend_stage", *password))
	if err != nil {
		fmt.Printf("Error: Failed to login to helm registry : Error message: %+v", err)
	}
	return nil
}

// Can be deleted
func GetHelmClient_HD(namespace string) (helmclient.Client, error) {

	kubeConfig := `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZEVENDQXZXZ0F3SUJBZ0lDQitNd0RRWUpLb1pJaHZjTkFRRUxCUUF3R0RFV01CUUdBMVVFQXhNTmEzVmkKWlhKdVpYUmxjeTFqWVRBZUZ3MHlOREE0TWpZeE1qUXlNREZhRncwek5EQTRNalF4TWpReU1ERmFNQmd4RmpBVQpCZ05WQkFNVERXdDFZbVZ5Ym1WMFpYTXRZMkV3Z2dJaU1BMEdDU3FHU0liM0RRRUJBUVVBQTRJQ0R3QXdnZ0lLCkFvSUNBUUM4dmdHeExJYUZPWUZ2VTdYNkVqY003bHl1eVFzYzhKSG1DZnE5aktzUk1BMmZ4R051aGZJeVhzbXEKY2dVMFNQUjkydSt4UUdkWlpwTzhsZEtkUm0vNis0QXdlQkJmZFR1TVRrTEF3cEE0cDhPVTNHaytVUmN3SmE4NApOWk83bkdMUmlETFhuTGpxSG9tRjBQYlVNbnkwYldKV3FnSUVJZVJMZjhXUE9oeldqWGpKTlliNmZsUUs1bXFjCms3b2ZNeGk2T2dZanhVck9zSEh0cnFKbjAvSFo0ZTZSbkxUTW1DVWJteWlZWEl5bFRlcTEvWVV1MVVOQklJWFQKQWNTQ3hpeVhHK09pZFdlL1NYN1l1RWMweWVrSldYa1JCckJYNW5DdnlZbDc4VGZiVzN6dy80MmtwQTBzK3RpZwo5MUMwOGtpTG1JeVZSSVVEWUJTbFJOaERvMzFDZDVNWTFXOWlmSko2MmNkejhTRUlwa25PV3BUQkRRTFRDbERlCnhBUTY2cGJCSENWcEVibUxMWDZ0bi9nb1pBZXVBdGROb1hhdmxVQW03bVBoY0hJemlrYmMxeFR6TTEyb2VLZ2gKUytZdFJNU1VsdmRPcHNZa1ZEUXlLQXNVL2ZWMjkyQndyb3Y3RGJnZ1FzelZ6MUNpeElGN1lxOGZwV1R6YzE2QgpycUo2bSt6K3lJRXhkZCtjQ0dURUNQT1IwWU16N3dXanViNWxzNWJZVG9xRzNRUEpLVEFaSkNhcEJ5c2VMRWYvClg4UlpjNVkzSUp2UUlGQVF2T2pnUzEzNkQ0TktSd0E5ZTlUTGVXbFl3UlNtcmpYYW9iWEZ6S0VsdEhDbzYzbEgKQVdSV1ZPSWluajZMNHNKdFRxc1JIVGZ2UUhxTmduakFYRlZxQVdNVG15V3M4MndnaHdJREFRQUJvMkV3WHpBTwpCZ05WSFE4QkFmOEVCQU1DQW9Rd0hRWURWUjBsQkJZd0ZBWUlLd1lCQlFVSEF3SUdDQ3NHQVFVRkJ3TUJNQThHCkExVWRFd0VCL3dRRk1BTUJBZjh3SFFZRFZSME9CQllFRkl4OUZsMFZmK0FnVHdXd2RQbDYvN0lGNEIxaU1BMEcKQ1NxR1NJYjNEUUVCQ3dVQUE0SUNBUUN4VDhLaGF5T0xGYzlSVjFsbmNPdW9tUUN1K3dQalg2VzRSb0tpVmtTWgo0dGxReUZZamtaRTliRFA1eUcxN3dyVjlVRUZSeW5aaGhxanRVNjBPMG1QOW83MVVjTFpwTHJDMW0xUG51TnJuCmEyS3M4OGxGTURGUlRSeENKRytRYmdPeC80c0N6OStTVm03d00vaW1rN2NqanFjNGdsamp3THlOWVpULy9YR3UKeFJ6ZzRhanVTcDJQOTdtVFNLMzREZkI0TmVVaFl2clc4azU3U3ZubnBYRGpFS2l3S2gxSlVaemhPRTUvbWlHcwoycWgyYUJ1OEJRWE9ycTljQUN6TzBLNG9vZGIxdFlTa2Z5dDdnUm5kb2dEZVBLNVJIL25ZOFJkV1lucGh4b0RwCmFFcEVqR2kxc3FZUTJvK2REcW9HcHhHdEFkZ0FQQ2ZidWYyWllmTnhRL2dlUUdETVhqYzNSc1lROFpMOHRiMU4KL2NUYjF2MElpbzBnTVl5N3lMcHdSUUNyUCtaWlk3aU5vV1FVZzhmQ0RCY0VFYWoyaHpUSUVLT1paTHlnZTlYYQpHb2dKMjRvczVUbWpYRll4ckJHdWdYVHpKNmdzQk8vZkZxSXRhWld1UGxiWWtsUWEyeStRcjhUVVIwaVA4QlBECllNMnltOStucFNlUXZER2k4VkVCa2FKTWlIUEJzY3hkTVJIWVhmMGhyYmIxR0ZkaDAxK1RSZlRJK2xTcHNlSXoKTEJpMTJ2ZUtHWWJneGNjSk9SQzZiYUJ6M2JZZS9zRFFmekFZYW5xcDhkOFQ0dVRVeURXYWw0TmNLcWpudHU1TQpFbW8wTDlJUGgwSDZsT1FpTTJqVVd1Q0FyWStXWmJtbmloVHZGcUxTUzlFMkh0cHJiaTc5Y25XRGVMRDVlNEdUCjNBPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://146.152.227.55:443
  name: cl-6zobp3cd2i
contexts:
- context:
    cluster: cl-6zobp3cd2i
    user: default-admin
  name: default
current-context: "default"
kind: Config
preferences: {}
users:
- name: default-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUZEekNDQXZlZ0F3SUJBZ0lDQm5vd0RRWUpLb1pJaHZjTkFRRUxCUUF3R0RFV01CUUdBMVVFQXhNTmEzVmkKWlhKdVpYUmxjeTFqWVRBZUZ3MHlOREE1TWpNeE1qRTFORGxhRncweU5ERXdNak14TWpFMU5EbGFNQ2t4RnpBVgpCZ05WQkFvVERuTjVjM1JsYlRwdFlYTjBaWEp6TVE0d0RBWURWUVFERXdWaFpHMXBiakNDQWlJd0RRWUpLb1pJCmh2Y05BUUVCQlFBRGdnSVBBRENDQWdvQ2dnSUJBS3IvdjEvY1QxRkM5em1mY1VmV2h1dnAyNytYUi9wV1NYdngKVkQ1NkpjYkVYRFN1UzhsRDJ0RUxnaXJ3SzZTZXlwUGtaZlJaWnVaNlhZQjVWTW1YSjFTWjk1bHVjM1BSQUNvMwozRFpaNUpqVjBkcXN6RUE3MGN0cDI2RHlla2Y0ZXM1OFUyNlVlV21DeEZLMy9VVG5lRUdxZlR0NUZ6akw5QVpzCnF6dFkvNE5uSjljTHArTE96WldpcUxDK3Y0UUlJc2xBUk1Kb1Y4anpQQUpvdDhFSWhRelI1dUExWVo2OFIrL3YKNVA5SmZGK3pvQWIyL3d2RlR5emEyVVFZdHJ6UkpjNCtIQ1krS0JsL1cxekhibFBCTFl3Tmk1YmJzZS9WenVTQQpCYzR6cVF6ZlQ2c0hWS1ltc0VMb2ZNb1d5TnNqcXJ0N0lQc3JaNUtaRFJYUUtnZTV2OWlEQi90RFVHaElRT0ZwClBNSzhrMFdJT2hZWkxscXlvR3NoMkk4N1lFVEYwbW5NVGs2UENXcmE5UnhwWnhtQVRxSjdpRkVlWjdkRzFxNXkKb3RxcjBUYlNpMzZMdm1EZTcyYjNiV1JBUWp5QkNySDJUQU14YmRNa2dlVFNWTmRBd1E1M3RWb1J0dkliRFJtcgpsdzF2VWhWaUNMZmtid2cyaHZYU2ZYc3BUdkpYUmxJcW9seXpVSFVodlcxdGI2K1ZlRHk5alRReGx5MklDcnJ0ClFnZHpaTHlKc1p6TkVRYi9mVGYyclpMRmtMRGU1NENKWkIwdUx3WllEbjVNNk1WVmNqQ1VFZ0VCeWtmOFRDSGUKank0b2FuQXk0Nmx0cEZPbzVmMURnNVU5dXJGY0Eyd2FhWHljYTFJTGpaNkZ4M1czdDNVSDIwV0RsREVySUY4OQp1UHJxcWR1eEFnTUJBQUdqVWpCUU1BNEdBMVVkRHdFQi93UUVBd0lIZ0RBZEJnTlZIU1VFRmpBVUJnZ3JCZ0VGCkJRY0RBZ1lJS3dZQkJRVUhBd0V3SHdZRFZSMGpCQmd3Rm9BVWpIMFdYUlYvNENCUEJiQjArWHIvc2dYZ0hXSXcKRFFZSktvWklodmNOQVFFTEJRQURnZ0lCQUxXeFpBeXhPZEhxam1SM0NVUXkxa1JUYTluOVpFKy9BZktvdXdPKwp6UHFkRjlYMVd2RFEvazZDcmxzSFVmV0RqZTIrc3hheVdpWlZFOGRsSU5XaDNVaHhyckYyTTBybktWalJUZHZ6ClZCZC96dHphSThPbk9WMzVidW5jaDVMZnJVcWJ1czJRc2g4Y1FTV01SVnRNUXpxUDF5bUwyV1VKY0k0ZWxLQU0KNlNoWmZ3SUJLZG9DaURGUVFWV1d5aERpb2JoR3RQVnFBVDV4UGU4RElwaC9LNzRpeW1mbmdHK212WFdFUUhHWgpQWHVEcDFMM1M0NVEyUlhuSjVYbWxHd0RlSXpRUDhvKzJ1eC90VUMwcis5N0NHekZ3S0l4Ym5kVThNcjY5dThOCjlPaG9WTVRwUGFyTTljanZQd0xvZWQ4cXRPcm1GWDcyM1BGRkhMYmZZaklUSno2cjlBeUxZdVUzMFhnTVRGUTkKOCtURE5xeVc5Zy9UUTJmWEhhaitXR2EwNXBEYTJ5bkpjS1ZhckJkUzJCUUNrS1dTNHVNeW1vVFBMdVh4eWkzQwplaHkraE1kci9nQmE3Vk5EbkVwZEZDb3VLdnRLcDYxSjFDZExTTjlrL0NOaERwL01VVjBrOGZjYzJYVVpObEt1Cmk5Z2NCMjlFWW1OcjVDc3dXc1E2SkU4bW15bDRqWkJCMWRCdDh1TDdPd3B5dFpydDZKekVjVk1Sb3BGelpnRVkKZkwxTHVXeFhvUWpKbmpMdkRMcXJHVFRRU1BuTE9ad1FmbXRhdGw0TThwV2FXTGtERk8zS3R6TSs1bXpndTFxcApsYTdHTExRUFdmeWlTVXFYc0I1QTYzdEMyMWRjQUlxM2tzZTl4dnNzcHVqM1lqenEvWnJKRXRFajVHYjg5d1RpCm1KaVEKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlKSndJQkFBS0NBZ0VBcXYrL1g5eFBVVUwzT1o5eFI5YUc2K25idjVkSCtsWkplL0ZVUG5vbHhzUmNOSzVMCnlVUGEwUXVDS3ZBcnBKN0trK1JsOUZsbTVucGRnSGxVeVpjblZKbjNtVzV6YzlFQUtqZmNObG5rbU5YUjJxek0KUUR2UnkybmJvUEo2Ui9oNnpueFRicFI1YVlMRVVyZjlST2Q0UWFwOU8za1hPTXYwQm15ck8xai9nMmNuMXd1bgo0czdObGFLb3NMNi9oQWdpeVVCRXdtaFh5UE04QW1pM3dRaUZETkhtNERWaG5yeEg3Ky9rLzBsOFg3T2dCdmIvCkM4VlBMTnJaUkJpMnZORWx6ajRjSmo0b0dYOWJYTWR1VThFdGpBMkxsdHV4NzlYTzVJQUZ6ak9wRE45UHF3ZFUKcGlhd1F1aDh5aGJJMnlPcXUzc2creXRua3BrTkZkQXFCN20vMklNSCswTlFhRWhBNFdrOHdyeVRSWWc2RmhrdQpXcktnYXlIWWp6dGdSTVhTYWN4T1RvOEphdHIxSEdsbkdZQk9vbnVJVVI1bnQwYldybktpMnF2Uk50S0xmb3UrCllON3ZadmR0WkVCQ1BJRUtzZlpNQXpGdDB5U0I1TkpVMTBEQkRuZTFXaEcyOGhzTkdhdVhEVzlTRldJSXQrUnYKQ0RhRzlkSjlleWxPOGxkR1VpcWlYTE5RZFNHOWJXMXZyNVY0UEwyTk5ER1hMWWdLdXUxQ0IzTmt2SW14bk0wUgpCdjk5Ti9hdGtzV1FzTjduZ0lsa0hTNHZCbGdPZmt6b3hWVnlNSlFTQVFIS1IveE1JZDZQTGlocWNETGpxVzJrClU2amwvVU9EbFQyNnNWd0RiQnBwZkp4clVndU5ub1hIZGJlM2RRZmJSWU9VTVNzZ1h6MjQrdXFwMjdFQ0F3RUEKQVFLQ0FnQTlCTW5ya1JnVXJVcS9HekEzTEV3MC90eFZmOHhGZm1qMmUyVk9iaFB3Mjd6elo0YlBxUkQ2SzVzbAphMUtIaWNwTC8rS0owU1V3OVZWTU5QK1dlQU9tNHRKQncvSWF6K2U1S1BuQncwNFpZNk5nM3V4N3QxempzMENXCkxEQ0tZaGFnZkNqaGVzWGdhck5YdVNQOVpJTzdHdlZaTlpxZHY5bXlPVERaR3FjQzR0cUttRFF1Y1JGWFpoWEEKRERFWEVqZ25qSEY4MWZNTldBNS81Wkk4cGFla3JYb3ZZNTBVSWFlaDdQN1FRZzdKcjdWWkJ1WjM4czZQK1FBeQpsb2NPMWFzaDczUG9DYUlSaHlxNDdzbGx4YmRWRkxoTSt0U0Ircys3Smh2c091OEdFdUhBNi9xKzEydHFWTC9DCjlXSnpJRVVhWlpPZFRSM0dhQ3NOTDV6djZNNldCRnlNWndZd2xJVGt1TFNTdzI1ZkV4ODViY0p2TjVnWXRCcDcKMnlYMlhNLzVCMCs2cEZORTJTTDEybENzSXVmQzJNNWxVNmE2TFhaZnJQaXhJaVFjTDF5OUNUL2tRejVXYm0zagp0Z2V1eTh1NThVMlBFd29hTFVmUGVRYUlqckdhNmxRNjlVOXJJT2RpdklRaEtHUUphMXcwZUtueitndVFZaFJSCkJsSGFXeTNPaVJRdFZZN1pzaFhRbkY4RHJyYzgrbmNKY1FPRUg1U0NNVU4yUklhS3NwVnlnMjlkZmJPOHB0M24KL0MrdTgxRmpEV0wyWHplUVJRZ092b1lDTmExTC9KN3hFQmJQbkpyWExCRjBiVk8veVl6eE1GV1JwNkt3M3J4NQovUGVKajlpbXRpR09GTWltWk43eTdoVjNKbm9uL2FHdmJYbUhqTWRXTGZTdXNpY0h5UUtDQVFFQXdnTVhnb1VoCkFwa01NOEl3am5vdXAvcFJtazRzMlBNMmMwU3JGeU50cWM5N3p2ck1sNGcwVVpUMlczckQ0eUo3ZURlNVNHaUsKbUtvMGpTMEE0QUIvQkpBQ0loajZidk1ETHFjcjd2TVEyWWpiN2Z3dzdxZHQwaTZUWlJONGphQVNMQ1hYMHY5dwpaVC9ZQU5FNWdTTU5MSVk5aGNZTzR3cWdEeTZ0WFJNRlNJVzZ6NVZWaUlZV2JxNnZaeDlxN282UElXMWlrT2Q1Ci8yZzUrSTgxREprdHVoMW9qbzRrU1FGUmlPMnlJS2x4aENEVlJEVS9oUXlHdjhGbVJWWnJuWVpwcTFJK1V1WWgKb2xwUGQ3aSs0ZU5uSGV2V0RUYTFZN1EwS2JRVnpUVW9SVG5YUFhaa1piN3o5UVVTcUVtNCtnRGVzYlJ0WjVMWQpzbjU1cXgyQnIvNm1Vd0tDQVFFQTRhSld3NWNGNlpQYU9oV0o0ZVY3SWE5d29nbStlbUhIbGJHWjNITjRRakNmCmZNMnNjSUxCTHUrVmFaQTVTZGpmV1hsS281NlBodlRnVWkzZExqWjBCejdaQmFZS1YxbGxqSnVNVXRrekdwTEkKMGNmYjR5empOek1heFllV05VTDloWndFaWJLZW9Md1MrQzJ4bzd1aVFnUFRpRURCRDF1N0lpb1pNUndGRmRnUQphcXpOUXd2aUhaTVVuSForV0J1MVFRRHRvUHQ3NC9lZ3YyUkxsUFdsUXQ1QjhKZ0dSeW9PL1pydkY4S0txRFdlCjlTbVIvcmVJWGlCdkJ4cUtvYitUNjVYRnR3OFFjVUdWQmtKQkl6Q0Z0UUJLVk9kcFhOQlBkNVpaRnB4RzZYR2IKaURKNVNRNUl0QkJhZ1k1b0ZHczVrNHRqbjZQWmNTM09xSEtSeU5OdGF3S0NBUUE3eXdRbDM2M0t4U3h6anplegowWWdya1FReFVFS1dJbTczbTRRM1AxMys1Y2s4Z3lNbTJIMTNYemVGL2hIOUlKVjQrWU9MQTEwanErRkNXVXBaCnZ3MW1kSk9UdXFzRUlyVXFYYTgybDRicjVEZ1Q0cE9hR2RQSTRUM2YrdDQrbDhUQ0FtKy93YVg4TG03OTRYMmQKaFJYOFVPc0pIWDlkRGR0Q2twb3ZnenN2bkxkMFhvdmI1YWRvT1VJcHdBOE9zclQrRWw0OFZuck04bXhiWkpkdQo2STZsTzRjTDJGYnFnUk9GNWV2dUVRckJNL1ZHYmpyRFlKYnU1a1lFdkp1eUVzamlXaGlIS0JIWm5ZZThXQjNNCk5HK0ZVemZISHNOTWxTODJZeUFNL0lNS3dzYkpWSUdnc2ZjeDNueGZqVWtMRTlXT1l4TU14cjh1VTdoZnVscEwKeVdtdEFvSUJBRzNVbE1GZVdSMXF2L1k0RjhiaTZuM3FKVHhxMjlJOG1HZFFiU0cyLzFuUXkwRjM2REZZSkdzUgphanhaWE5tS2ZLWFQrYllOYjdYMHF1QzF2STFMS2syQWxTay91cGJzU0JjYWZFS3p3VUYxSTlXaG9ISkRubEozClNOZlArUmp5QS9BdWtyTG9SSGpmTTZpa3JXeVM0QmVjUHpKNnVyOGNHc28xamMrdTRQYlNGcU9tZTcvZ0grL2YKY1UvOENlSWZrcW9TcHBrTzhTNTFra1Mzc0diUTcrSE55SEV4dnhTUzczc1pHTHNMRW0xd3RIQ0lESzNkYnI3SwpzT0RYVlpZSWFCdHI2ek1CWFRLVUJ0Tm1Hc2pqVEtKZmdzOHpXY0U1RVFXWUpNTnh6TDdEY3o3bnVzd0o3Ty9oCmFmelljZWRHamw4NG9ZVEt3QVJzbE1UQUNDQ21jZ01DZ2dFQUtmY3R1Q09MNHptY3hHRU9wSEdLNzQ3dDQ4eSsKQWxpbStUVTRlSDJWSFpxV0ZpcjIwbVR5b1pJcEFaNVhTcStkVUEwWTJMK0RaNnNJRVdHeGErTVlyK2twbzZzeApWTXdJb3VTUUd2YmM2SkJIMWZIeUlxRVRUYzgwOENLTFFKUkJlRlhlOGQydFl1ZWlyRDFjdlRqMkVIeTNxTjdOClQzY25HZlZFWUJSQ3QrWFFjTHJzclZoVzVZM3JZV0pSclJJTnZBZUpiOHBmVnJuUFduam5mZTl0UWpsemR4SkMKNXJKUFNPV1V6VXg4Skt3dFBtSjBtTllYV0NicitXdVhWWjBxL3I1SzhnWDFEL3BIU0trdEVhRnlRZi9tbmRyQQpGRW5KTHNEV1hlbEtQQ2cvbzM4K1E1ZnVlbkFsc1RsTUV4MHJEZEtEaDVTREFmUVVFZTJ4VUJBNGJBPT0KLS0tLS1FTkQgUlNBIFBSSVZBVEUgS0VZLS0tLS0K
`
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		// Handle error appropriately
		log.Println("Error parsing kubeconfig:", err)
	}
	opt := &helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace:        namespace, // Change this to the namespace you wish the client to operate in.
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true, // Change this to false if you don't want linting.
			RegistryConfig:   credpath,
			DebugLog: func(format string, v ...interface{}) {
				// Change this to your own logger. Default is 'log.Printf(format, v...)'.
			},
		},
		RestConfig: config,
	}

	helmClient, err := helmclient.NewClientFromRestConf(opt)
	if err != nil {
		return nil, fmt.Errorf("error: Failed to get helm client: Error message: %+v", err)
	}
	return helmClient, nil
}

// Can be deleted
func PrintFileCont() {

	// Read the file
	data, err := ioutil.ReadFile(credpath)
	if err != nil {
		fmt.Printf("Error reading file: %+v", err)
	}

	// Print the contents of the file
	fmt.Println(string(data))
}

func logDebug(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func pullChart(reg string, chart string, version string, username *string, password *string) (*chart.Chart, error) {
	// registry client
	registryClient, err := registry.NewClient(
		registry.ClientOptDebug(false),
		registry.ClientOptCredentialsFile(credpath),
		registry.ClientOptEnableCache(false),
	)
	if err != nil {
		return nil, err
	}

	// init helm action config
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(nil, "", "secret", logDebug); err != nil {
		return nil, err
	}

	actionConfig.RegistryClient = registryClient

	// login to registry
	if username != nil && password != nil {
		auth := registry.LoginOptBasicAuth(*username, *password)
		if err := actionConfig.RegistryClient.Login(reg, auth); err != nil {
			return nil, err
		}

	}

	// pull the chart
	pull := action.NewPullWithOpts(action.WithConfig(actionConfig))
	pull.Settings = cli.New() // didn't want to do this but otherwise it goes nil pointer
	pull.ChartPathOptions.Version = version
	pull.DestDir = "./"
	pull.VerifyLater = true
	if _, err := pull.Run(fmt.Sprintf("oci://%s/%s", reg, chart)); err != nil {
		return nil, err
	}
	// load chart from file
	return loader.LoadFile(fmt.Sprintf("%s-%s.tgz", chart, version))

}

// Can be deleted
func findAirflowFile(folder string) (string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %+v", err)
	}

	prefix := "airflow-1.13.1.tgz"
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			return filepath.Join(folder, entry.Name()), nil
		}
	}
	fmt.Printf("no file starting with %s found in the directory %s", prefix, folder)
	return "", fmt.Errorf("no file starting with %s found in the directory %s", prefix, folder)
}

// Can be deleted
func RenameAirflowFile(folder string) error {
	// Define the pattern to look for
	pattern := `^airflow-1\.13\.1\.tgz\d+$`
	// Compile the regular expression
	re := regexp.MustCompile(pattern)

	// Read the files in the specified directory
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	// Iterate over the files
	for _, file := range files {
		// Check if the file name matches the pattern
		if re.MatchString(file.Name()) {
			// Define the new file name
			newName := "airflow-1.13.1.tgz"
			// Rename the file
			err := os.Rename(filepath.Join(folder, file.Name()), filepath.Join(folder, newName))
			if err != nil {
				return fmt.Errorf("error renaming file %s to %s: %w", file.Name(), newName, err)
			} else {
				fmt.Printf("Renamed %s to %s\n", file.Name(), newName)
			}
			// Assuming there should only be one such file, we can break the loop after renaming
			break
		}
	}

	return nil
}

func DownloadChart(folder, reg, chart, version string, username, password *string) (string, error) {
	aiflowfilepath, err := findChartFile(folder, chart, version)
	if err != nil {
		return "", fmt.Errorf("unable to find helm chart file . Error message: %+v", err)
	}

	if aiflowfilepath != "" {
		fmt.Printf("Temp file path first loop %s \n", aiflowfilepath)
		if err := RenameChartFile(folder, chart, version); err != nil {
			return "", fmt.Errorf("unable to rename helm chart file. Error message: %+v", err)
		}
		time.Sleep(5 * time.Second)
		//fmt.Println("Done waiting!")
		fileInfo, err := os.Stat(aiflowfilepath)
		if err != nil {
			fmt.Println("Error getting file stats:", err)
			return "error", err
		}
		fmt.Println("Done waiting!")
		fmt.Printf("File Name: %s\n", fileInfo.Name())
		fmt.Printf("Size: %d bytes\n", fileInfo.Size())
		fmt.Printf("Permissions: %s\n", fileInfo.Mode())
		fmt.Printf("Last Modified: %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))
		return aiflowfilepath, nil

	} else {
		_, err := pullChart(reg, chart, version, username, password)
		if err != nil {
			fmt.Printf("Error pulling from registry: %s\n", err)
			fmt.Println("checking the file insde pull if condition")
			aiflowfilepath, err = findChartFile(folder, chart, version)
			if err != nil {
				return "", fmt.Errorf("unable to find helm chart file . Error message: %+v", err)
			}
			fmt.Printf("Temp file path %s\n.. printing stats..\n", aiflowfilepath)
			if err := RenameChartFile(folder, chart, version); err != nil {
				return "", fmt.Errorf("unable to rename helm chart file. Error message: %+v", err)
			}
			fmt.Println("Waiting for 5 seconds...")
			time.Sleep(5 * time.Second)
			aiflowfilepathnew := folder + "/" + chart + "-" + version + ".tgz"
			fmt.Printf("New file path %s\n.. printing stats..\n", aiflowfilepathnew)
			fileInfo, err := os.Stat(aiflowfilepathnew)
			if err != nil {
				fmt.Println("Error getting file stats:", err)
				return "error", err
			}
			fmt.Println("Done waiting!")
			fmt.Printf("File Name: %s\n", fileInfo.Name())
			fmt.Printf("Size: %d bytes\n", fileInfo.Size())
			fmt.Printf("Permissions: %s\n", fileInfo.Mode())
			fmt.Printf("Last Modified: %s\n", fileInfo.ModTime().Format("2006-01-02 15:04:05"))
			return aiflowfilepathnew, nil
		} else {
			return folder + "/" + chart + "-" + version + ".tgz", nil
		}
	}
}
func findChartFile(folder string, chart string, version string) (string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return "", fmt.Errorf("error reading directory: %+v", err)
	}

	prefix := chart + "-" + version + ".tgz"
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			return filepath.Join(folder, entry.Name()), nil
		}
	}
	fmt.Printf("no file starting with %s found in the directory %s", prefix, folder)
	return "", nil
}

func RenameChartFile(folder string, chart string, version string) error {
	prefix := chart + "-" + version + ".tgz"
	// Define the pattern to look for
	pattern := `^` + prefix + `\d+$`
	// Compile the regular expression
	re := regexp.MustCompile(pattern)

	// Read the files in the specified directory
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	// Iterate over the files
	for _, file := range files {
		// Check if the file name matches the pattern
		if re.MatchString(file.Name()) {
			// Define the new file name
			newName := prefix
			// Rename the file
			err := os.Rename(filepath.Join(folder, file.Name()), filepath.Join(folder, newName))
			if err != nil {
				return fmt.Errorf("error renaming file %s to %s: %w", file.Name(), newName, err)
			} else {
				fmt.Printf("Renamed %s to %s\n", file.Name(), newName)
			}
			// Assuming there should only be one such file, we can break the loop after renaming
			break
		}
	}

	return nil
}

// ***Rajesh Added code:End
